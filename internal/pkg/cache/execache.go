package cache

import (
    "compress/gzip"
    "fmt"
    "github.com/chirino/uc/internal/pkg/archive"
    "github.com/chirino/uc/internal/pkg/pkgsign"
    "github.com/chirino/uc/internal/pkg/user"
    "io"
    "net/http"
    "os"
    "path"
    "path/filepath"
    "runtime"
)

type Request struct {
    Printf       func(format string, a ...interface{})
    URL          string
    Signature    string
    Size         int64
    DownloadPath string `json:"download-path,omitempty"`
    CommandName  string `json:"command-name,omitempty"`
    Version      string
    ExtractZip   string `json:"extract-zip,omitempty"`
    ExtractTgz   string `json:"extract-tgz,omitempty"`
    Uncompress   string `json:"extract-gz,omitempty"`
}

func Get(r *Request) (string, error) {
    dir, err := CacheCommandPath()
    if err != nil {
        return "", err
    }

    targetExe := filepath.Join(dir, r.CommandName, r.Version, ExeSuffix(r.CommandName))
    exists, err := Exists(targetExe)
    if err != nil {
        return "", err
    }
    if exists {
        return targetExe, nil
    }
    downloadPath, err := Download(r)
    if err != nil {
        return "", err
    }

    dir = filepath.Dir(targetExe)
    err = os.MkdirAll(dir, 0755)
    if err != nil {
        return "", err
    }

    if r.ExtractZip != "" {
        err = archive.UnzipCommand(downloadPath, r.ExtractZip, targetExe)
        if err != nil {
            return "", err
        }
    } else if r.ExtractTgz != "" {
        err = archive.UntgzCommand(downloadPath, r.ExtractTgz, targetExe)
        if err != nil {
            return "", err
        }
    } else if r.Uncompress != "gz" {
        _, err := CopyExectuable(downloadPath, targetExe, func(r io.Reader) (closer io.ReadCloser, e error) {
            return gzip.NewReader(r)
        })
        if err != nil {
            return "", err
        }
    } else {
        _, err := CopyExectuable(downloadPath, targetExe)
        if err != nil {
            return "", err
        }
    }
    return targetExe, nil
}

// Returns the path on the local file system for the requested exe
func Download(r *Request) (string, error) {

    downloadPath, err := CacheDownloadPath()
    if r.DownloadPath == "" {
        r.DownloadPath = path.Base(r.URL)
    }
    to := filepath.Join(downloadPath, r.DownloadPath)

    err = download(r, to)
    if err != nil {
        return "", err
    }

    return to, nil
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

func download(r *Request, to string) error {

    // Create the directories...
    dir := filepath.Dir(to)
    err := os.MkdirAll(dir, 0755)
    if err != nil {
        return err
    }

    exists, err := Exists(to)
    if err != nil {
        return err
    }
    if exists {
        err := verifyDownload(r, to)
        if err == nil {
            return nil
        }
    }

    // Do the download in a new block so that at the end the files
    // and resources are closed.
    {

        r.Printf("downloading: %s\n", r.URL)
        err := HttpGet(r.URL, to)
        if err != nil {
            return err
        }
        r.Printf("done: %s\n", r.URL)

    }

    return verifyDownload(r, to)
}

func HttpGet(url string, to string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode < 200  || resp.StatusCode >= 300 {
        return fmt.Errorf("get '%s' status code: %d", url, resp.StatusCode)
    }
    out, err := os.Create(to)
    if err != nil {
        return err
    }
    defer out.Close()
    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    return nil
}

func verifyDownload(r *Request, downloadPath string) error {
    i, err := os.Stat(downloadPath)
    if err != nil {
        return err
    }

    if i.Size() != r.Size {
        return fmt.Errorf("downloaded file is %d bytes, expected %d bytes", i.Size(), r.Size)
    }

    r.Printf("checking digital signature of: %s\n", downloadPath)
    return pkgsign.CheckSignature(r.Signature, downloadPath)
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
