package files

import (
    "os"
    "path/filepath"
)

func WithCreate(name string, action func(*os.File) error) error {
    dir := filepath.Dir(name)
    err := os.MkdirAll(dir, 0755)
    if err != nil {
        return err
    }

    targetFile, err := os.Create(name)
    if err != nil {
        return err
    }
    defer targetFile.Close()
    return action(targetFile)
}

func WithOpenFile(name string, flag int, perm os.FileMode, action func(*os.File) error) error {
    targetFile, err := os.OpenFile(name, flag, perm)
    if err != nil {
        return err
    }
    defer targetFile.Close()
    return action(targetFile)
}