// +build dev

package cmd

import (
    "fmt"
    "github.com/chirino/uc/internal/dev"
    "github.com/chirino/uc/internal/pkg/pkgsign"
    "io/ioutil"
    "os"
    "path/filepath"
    "sigs.k8s.io/yaml"
)

func loadConfig() (*CatalogConfig, error) {
    path := filepath.Join(dev.GO_MOD_DIRECTORY, "docs", "catalog.yaml")
    sigpath := path + ".sig"

    // Validate the catalog signature..
    sig, err := ioutil.ReadFile(sigpath)
    if err != nil {
        return nil, err
    }
    if err := pkgsign.CheckSignature(string(sig), path); err != nil {
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