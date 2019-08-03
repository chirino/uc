package archive

import (
    "archive/tar"
    "compress/gzip"
    "fmt"
    "io"
    "os"
)

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
        if err!=nil {
            return err
        }

        target, targetMode := filter(tgzEntry)
        if target == "" {
            continue
        }

        if err := WithOpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, targetMode, func(file *os.File) error {
            _, err = io.Copy(file, tarReader)
            return err
        }); err != nil {
            return err
        }

    }
    return nil
}

func WithOpenFile(name string, flag int, perm os.FileMode, action func(*os.File)error) error {
    targetFile, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
    if err != nil {
        return err
    }
    defer targetFile.Close()
    return action(targetFile)
}