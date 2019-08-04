package files

import "os"

func WithCreate(name string, action func(*os.File) error) error {
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