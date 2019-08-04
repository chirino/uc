// +build dev

package dev

import (
    "os"
    "path/filepath"
)

var (
    GO_MOD_DIRECTORY string
)

func init() {
    // Save the original directory the process started in.
    wd, err := os.Getwd();
    if err != nil {
        panic(err)
    }

    wd, err = filepath.Abs(wd)
    if err != nil {
        panic(err)
    }

    // Find the module dir..
    current := ""
    for next := wd; current != next; next = filepath.Dir(current) {
        current = next
        if fileExists(filepath.Join(current, "go.mod")) && fileExists(filepath.Join(current, "go.sum")) {
            GO_MOD_DIRECTORY = current
            break
        }
    }

    if GO_MOD_DIRECTORY == "" {
        panic("could not find the root module directory")
    }
}

func fileExists(filename string) bool {
    s, err := os.Stat(filename)
    if err != nil {
        return false
    }
    return !s.IsDir()
}
