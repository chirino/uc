// +build dev

package cmd

import (
    "github.com/chirino/uc/internal/dev"
    "io/ioutil"
    "os"
    "path/filepath"
    "sigs.k8s.io/yaml"
)

func loadConfig() (*CatalogConfig, error) {
    path := filepath.Join(dev.GO_MOD_DIRECTORY, "docs", "catalog.yaml")
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