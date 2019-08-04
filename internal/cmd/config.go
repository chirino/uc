// +build !dev

package cmd

import (
    "fmt"
    "github.com/chirino/uc/internal/pkg/cache"
    "github.com/chirino/uc/internal/pkg/files"
    "github.com/chirino/uc/internal/pkg/signature"
    "github.com/chirino/uc/internal/pkg/user"
    "io/ioutil"
    "os"
    "path/filepath"
    "sigs.k8s.io/yaml"
    "time"
)

func LoadConfig() (*CatalogConfig, error) {

    path, err := CatalogPath()
    if err != nil {
        return nil, err
    }

    downloadCatalog := false
    if info, err := os.Stat(path); err != nil || time.Now().After(info.ModTime().Add(24*time.Hour)){
        downloadCatalog = true
    }

    sigpath := path + ".sig"
    if info, err := os.Stat(sigpath); err != nil || time.Now().After(info.ModTime().Add(24*time.Hour)){
        downloadCatalog = true
    }

    if downloadCatalog {
        if err := files.WithCreate(sigpath, func(file *os.File) error {
            return cache.HttpGet("https://chirino.github.io/uc/catalog.yaml.sig", file)
        }); err != nil {
            return nil, err
        }
        if err := files.WithCreate(path, func(file *os.File) error {
            return cache.HttpGet("https://chirino.github.io/uc/catalog.yaml", file)
        }); err != nil {
            return nil, err
        }
    }

    // Validate the catalog signature..
    sig, err := ioutil.ReadFile(sigpath)
    if err != nil {
        return nil, err
    }
    if err := signature.CheckSignature(string(sig), path); err != nil {
        os.Remove(path) // this will trigger a re-download of the catalog..
        return nil, fmt.Errorf("validating %s: %v", path, err)
    }

    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    bytes, err := ioutil.ReadAll(file)
    if err != nil {
        return nil, err
    }

    config := &CatalogConfig{}
    err = yaml.Unmarshal(bytes, config)
    return config, err
}

func CatalogPath() (string, error) {
    home := user.HomeDir()
    if home == "" {
        return "", fmt.Errorf("Cannot determine the user home directory")
    }
    return filepath.Join(home, ".uc", "catalog.yaml"), nil
}
