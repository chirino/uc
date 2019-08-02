package archive

import (
    "archive/zip"
    "fmt"
    "io"
    "os"
    "path/filepath"
)

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

        dir := filepath.Dir(target)
        err := os.MkdirAll(dir, 0755)
        if err != nil {
            return err
        }

        zippedFile, err := zipEntry.Open()
        if err != nil {
            return err
        }
        defer zippedFile.Close()

        targetFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, targetMode)
        if err != nil {
            return err
        }
        defer targetFile.Close()

        _, err = io.Copy(targetFile, zippedFile)
        if err != nil {
            return err
        }
    }
    return nil
}
