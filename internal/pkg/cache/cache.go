package cache

import (
    "compress/gzip"
    "fmt"
    "github.com/chirino/uc/internal/pkg/archive"
    "github.com/chirino/uc/internal/pkg/signature"
    "github.com/chirino/uc/internal/pkg/user"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "runtime"
)

type Request struct {
    URL              string                                `json:"url,omitempty"`
    Signature        string                                `json:"signature,omitempty"`
    Size             int64                                 `json:"size,omitempty"`
    ExtractZip       string                                `json:"extract-zip,omitempty"`
    ExtractTgz       string                                `json:"extract-tgz,omitempty"`
    Uncompress       string                                `json:"uncompress,omitempty"`
    Platform         string                                `json:"-"`
    Version          string                                `json:"-"`
    CommandName      string                                `json:"-"`
    SkipVerification bool                                  `json:"-"`
    ForceDownload    bool                                  `json:"-"`
    Printf           func(format string, a ...interface{}) `json:"-"`
}

func Get(r *Request) (string, error) {
    dir, err := CacheCommandPath()
    if err != nil {
        return "", err
    }

    targetExe := filepath.Join(dir, r.CommandName, r.Version, r.Platform, ExeSuffix(r.CommandName))

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

    tempFile, err := DownloadToTempFile(r)
    if err != nil {
        return "", err
    }

    // Delete the downloaded file...
    defer os.Remove(tempFile)

    dir = filepath.Dir(targetExe)
    err = os.MkdirAll(dir, 0755)
    if err != nil {
        return "", err
    }

    if r.ExtractZip != "" {
        err = archive.UnzipCommand(tempFile, r.ExtractZip, targetExe)
        if err != nil {
            return "", err
        }
    } else if r.ExtractTgz != "" {
        err = archive.UntgzCommand(tempFile, r.ExtractTgz, targetExe)
        if err != nil {
            return "", err
        }
    } else if r.Uncompress == "gz" {
        _, err := CopyExectuable(tempFile, targetExe, func(r io.Reader) (closer io.ReadCloser, e error) {
            return gzip.NewReader(r)
        })
        if err != nil {
            return "", err
        }
    } else {
        _, err := CopyExectuable(tempFile, targetExe)
        if err != nil {
            return "", err
        }
    }

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

    r.Printf("downloading: %s\n", r.URL)
    err = HttpGet(r.URL, to)
    if err != nil {
        to.Close()
        os.Remove(to.Name())
        return "", err
    }
    r.Printf("done: %s\n", r.URL)
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
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("get '%s' status code: %d", url, resp.StatusCode)
    }
    // Write the body to file
    _, err = io.Copy(to, resp.Body)
    return nil
}

func Verify(r *Request, file string) error {
    if r.SkipVerification {
        return nil
    }

    i, err := os.Stat(file)
    if err != nil {
        return err
    }

    if i.Size() != r.Size {
        return fmt.Errorf("downloaded file is %d bytes, expected %d bytes", i.Size(), r.Size)
    }

    r.Printf("checking digital signature of: %s\n", file)
    return signature.CheckSignature(r.Signature, file)
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

func ExeSuffix(s string) string {
    if runtime.GOOS == "windows" {
        return s + ".exe"
    }
    return s
}

func CopyExectuable(from string, to string, filters ...func(r io.Reader) (io.ReadCloser, error)) (int64, error) {
    source, err := os.Open(from)
    if err != nil {
        return 0, err
    }
    defer source.Close()

    sourceReader := io.ReadCloser(source)
    for _, f := range filters {
        sourceReader, err = f(sourceReader)
        if err != nil {
            return 0, err
        }
        defer sourceReader.Close()
    }

    destination, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
    if err != nil {
        return 0, err
    }
    defer destination.Close()
    nBytes, err := io.Copy(destination, sourceReader)
    return nBytes, err
}
