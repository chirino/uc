package cache

import (
    "fmt"
    "github.com/chirino/uc/internal/cmd"
    "github.com/chirino/uc/internal/pkg/archive"
    "github.com/chirino/uc/internal/pkg/pkgsign"
    "github.com/chirino/uc/internal/pkg/user"
    "io"
    "net/http"
    "os"
    "path"
    "path/filepath"
)

type Request struct {
    Printf       func(format string, a ...interface{})
    URL          string
    Signature    string
    Size         int64
    DownloadPath string
    CommandName  string
    Version      string
    PathInZip    string
}

func GetCommandFromCache(r *Request) (string, error) {
    dir, err := CacheCommandPath()
    if err != nil {
        return "", err
    }

    targetExe := filepath.Join(dir, r.CommandName, r.Version, cmd.ExeSuffix(r.CommandName))
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

    if r.PathInZip != "" {
        err = archive.UnzipCommand(downloadPath, r.PathInZip, targetExe)
        if err != nil {
            return "", err
        }
        // TODO: figure out..
        return targetExe, nil
    } else {
        return downloadPath, nil
    }
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

        resp, err := http.Get(r.URL)
        if err != nil {
            return err
        }
        defer resp.Body.Close()

        out, err := os.Create(to)
        if err != nil {
            return err
        }
        defer out.Close()

        // Write the body to file
        _, err = io.Copy(out, resp.Body)
        r.Printf("done: %s\n", r.URL)

    }

    return verifyDownload(r, to)
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
    return filepath.Join(home, ".uc", "cache", "downloads"), nil
}
