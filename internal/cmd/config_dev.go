// +build dev

package cmd

import (
    "fmt"
    "github.com/chirino/uc/internal/pkg/cache"
    "github.com/chirino/uc/internal/pkg/dev"
    "github.com/chirino/uc/internal/pkg/signature"
    "io/ioutil"
    "os"
    "path/filepath"
    "sigs.k8s.io/yaml"
)

func LoadCatalogConfig() (*CatalogConfig, error) {
    path := filepath.Join(dev.GO_MOD_DIRECTORY, "docs", "catalog.yaml")
    result := &CatalogConfig{}
    err := loadConfig(path, result)
    return result, err
}

func LoadCommandPlatforms(command string, version string) (map[string]*cache.Request, error) {
    path := filepath.Join(dev.GO_MOD_DIRECTORY, "docs", "catalog", command, version, "platforms.yaml")
    result := map[string]*cache.Request{}
    err := loadConfig(path, &result)
    return result, err
}

func loadConfig(path string, config interface{}) (error) {
    err := CheckSigatureAgainstSigFile(path)
    if err != nil {
        return err
    }
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()
    bytes, err := ioutil.ReadAll(file)
    if err != nil {
        return err
    }
    return yaml.Unmarshal(bytes, config)
}

func CheckSigatureAgainstSigFile(path string) (error) {
    sigpath := path + ".sig"
    sig, err := ioutil.ReadFile(sigpath)
    if err != nil {
        return err
    }
    if err := signature.CheckSignature(string(sig), path); err != nil {
        return fmt.Errorf("validating %s: %v", path, err)
    }
    return  nil
}
