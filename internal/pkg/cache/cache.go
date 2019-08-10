package cache

import (
	"fmt"
	"github.com/chirino/uc/internal/pkg/archive"
	"github.com/chirino/uc/internal/pkg/files"
	"github.com/chirino/uc/internal/pkg/signature"
	"github.com/chirino/uc/internal/pkg/user"
	"golang.org/x/crypto/openpgp"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Request struct {
	URL           string             `json:"url,omitempty"`
	Signature     string             `json:"signature,omitempty"`
	Size          int64              `json:"size,omitempty"`
	ExtractZip    string             `json:"extract-zip,omitempty"`
	ExtractTgz    string             `json:"extract-tgz,omitempty"`
	Uncompress    string             `json:"uncompress,omitempty"`
	Platform      string             `json:"-"`
	Version       string             `json:"-"`
	CommandName   string             `json:"-"`
	ForceDownload bool               `json:"-"`
	InfoLog       io.Writer          `json:"-"`
	DebugLog      io.Writer          `json:"-"`
	Keyring       openpgp.EntityList `json:"-"`
}

func Get(r *Request) (string, error) {
	commandsDir, err := CacheCommandPath()
	if err != nil {
		return "", err
	}

	targetExe := filepath.Join(commandsDir, r.CommandName, r.Version, r.Platform, ExeSuffix(r.Platform, r.CommandName))

	if !r.ForceDownload {
		exists, err := Exists(targetExe)
		if err != nil {
			return "", err
		}
		if exists {
			err := Verify(r, targetExe)
			if err != nil {
				return "", err
			}
			return targetExe, nil
		}
	}

	fmt.Fprintln(r.InfoLog, "downloading:", r.URL)
	if r.ExtractZip != "" {
		// we can stream right to the target file..
		err = WithHttpGetReader(r.URL, func(src io.Reader) error {
			return files.WithCreateThenReplace(targetExe, 0775, func(dst *os.File) error {
				_, err := io.Copy(dst, src)
				return err
			})
		}, archive.ZipReaderMiddleware(&r.ExtractZip))
		if err != nil {
			return "", err
		}
	} else if r.ExtractTgz != "" {
		// we can stream right to the target file..
		err = WithHttpGetReader(r.URL, func(src io.Reader) error {
			return files.WithCreateThenReplace(targetExe, 0775, func(dst *os.File) error {
				_, err := io.Copy(dst, src)
				return err
			})
		}, archive.GzipReaderMiddleware, archive.TarReaderMiddleware(&r.ExtractTgz))
		if err != nil {
			return "", err
		}

	} else if r.Uncompress == "gz" {
		// we can stream right to the target file..
		err = WithHttpGetReader(r.URL, func(src io.Reader) error {
			return files.WithCreateThenReplace(targetExe, 0775, func(dst *os.File) error {
				_, err := io.Copy(dst, src)
				return err
			})
		}, archive.GzipReaderMiddleware)
		if err != nil {
			return "", err
		}
	} else {
		// we can stream right to the target file..
		err = WithHttpGetReader(r.URL, func(src io.Reader) error {
			return files.WithCreateThenReplace(targetExe, 0775, func(dst *os.File) error {
				_, err := io.Copy(dst, src)
				return err
			})
		})
		if err != nil {
			return "", err
		}
	}
	fmt.Fprintln(r.InfoLog, "wrote:", targetExe)

	err = Verify(r, targetExe)
	if err != nil {
		return "", err
	}
	return targetExe, nil
}

// Returns the path on the local file system for the requested exe
func DownloadToTempFile(r *Request) (string, error) {

	// Create the download dir...
	downloadPath, err := CacheDownloadPath()
	err = os.MkdirAll(downloadPath, 0755)
	if err != nil {
		return "", err
	}

	to, err := ioutil.TempFile(downloadPath, "download")
	if err != nil {
		log.Fatal(err)
	}
	defer to.Close()

	fmt.Fprintln(r.InfoLog, "downloading:", r.URL)
	err = HttpGet(r.URL, to)
	if err != nil {
		to.Close()
		os.Remove(to.Name())
		return "", err
	}
	fmt.Fprintln(r.InfoLog, "wrote:", r.URL)
	return to.Name(), nil
}

func Exists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, nil
	}
	return true, nil
}

func HttpGet(url string, to io.Writer) error {
	return WithHttpGetReader(url, func(reader io.Reader) error {
		_, err := io.Copy(to, reader)
		return err
	})
}

func WithHttpGetReader(url string, action func(io.Reader) error, filters ...func(r io.Reader) (io.Reader, error)) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("get '%s' status code: %d", url, resp.StatusCode)
	}
	return files.WithReader(resp.Body, action, filters...)
}

func Verify(r *Request, file string) error {
	if r.Keyring == nil {
		return nil
	}

	i, err := os.Stat(file)
	if err != nil {
		return err
	}

	if i.Size() != r.Size {
		return fmt.Errorf("downloaded file is %d bytes, expected %d bytes", i.Size(), r.Size)
	}

	return signature.CheckSignature(r.Keyring, r.Signature, file)
}

func CacheDownloadPath() (string, error) {
	home := user.HomeDir()
	if home == "" {
		return "", fmt.Errorf("Cannot determine the user home directory")
	}
	return filepath.Join(home, ".uc", "cache", "downloads"), nil
}

func CacheCommandPath() (string, error) {
	home := user.HomeDir()
	if home == "" {
		return "", fmt.Errorf("Cannot determine the user home directory")
	}
	return filepath.Join(home, ".uc", "cache", "commands"), nil
}

func ExeSuffix(platform string, name string) string {
	if strings.HasPrefix(platform, "windows-") {
		return name + ".exe"
	}
	return name
}
