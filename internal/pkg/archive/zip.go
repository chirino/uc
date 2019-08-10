package archive

import (
	"archive/zip"
	"fmt"
	"github.com/chirino/uc/internal/pkg/files"
	"github.com/krolaw/zipstream"
	"io"
	"os"
	"strings"
)

func ZipReaderMiddleware(name *string) func(r io.Reader) (closer io.Reader, e error) {
	return func(r io.Reader) (closer io.Reader, e error) {
		names := map[string]bool{}
		for _, value := range strings.Split(*name, "|") {
			names[value] = true
		}

		archive := zipstream.NewReader(r)
		for {
			entry, err := archive.Next()

			// if no more files are found return
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			if names[entry.Name] {
				*name = entry.Name
				return archive, nil
			}
		}
		return nil, fmt.Errorf("File not found in zip: " + *name)
	}
}

func UnzipCommand(zipFile string, pathInZip string, to string) error {
	found := false
	err := Unzip(zipFile, func(dest *zip.File) (string, os.FileMode) {
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
		return fmt.Errorf("File not found in zip: " + pathInZip)
	}
	return nil
}

func Unzip(zipFile string, filter func(dest *zip.File) (string, os.FileMode)) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, zipEntry := range r.File {
		target, targetMode := filter(zipEntry)
		if target == "" {
			continue
		}

		zippedFile, err := zipEntry.Open()
		if err != nil {
			return err
		}

		if err := files.WithOpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, targetMode, func(file *os.File) error {
			_, err = io.Copy(file, zippedFile)
			return err
		}); err != nil {
			return err
		}
		zippedFile.Close()
	}
	return nil
}
