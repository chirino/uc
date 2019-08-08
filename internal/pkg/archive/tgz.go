package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/chirino/uc/internal/pkg/files"
	"io"
	"os"
)

func GzipReaderMiddleware(r io.Reader) (closer io.Reader, e error) {
	return gzip.NewReader(r)
}

func TarReaderMiddleware(name string) func(r io.Reader) (closer io.Reader, e error) {
	return func(r io.Reader) (closer io.Reader, e error) {
		tarReader := tar.NewReader(r)
		for {
			tgzEntry, err := tarReader.Next()

			// if no more files are found return
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			if tgzEntry.Name == name {
				return tarReader, nil
			}
		}
		return nil, fmt.Errorf("File not found in tgz: " + name)
	}
}

func UntgzCommand(tgzFile string, pathInZip string, to string) error {
	found := false
	err := Untgz(tgzFile, func(dest *tar.Header) (string, os.FileMode) {
		if dest.Name == pathInZip {
			found = true
			return to, 0755
		}
		return "", 0755
	})
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("File not found in tgz: " + pathInZip)
	}
	return nil
}

func Untgz(tgzFile string, filter func(dest *tar.Header) (string, os.FileMode)) error {
	file, err := os.Open(tgzFile)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		tgzEntry, err := tarReader.Next()

		// if no more files are found return
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target, targetMode := filter(tgzEntry)
		if target == "" {
			continue
		}

		if err := files.WithOpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, targetMode, func(file *os.File) error {
			_, err = io.Copy(file, tarReader)
			return err
		}); err != nil {
			return err
		}

	}
	return nil
}
