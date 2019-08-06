// +build !dev

package catalog

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

func LoadCatalogConfig() (*CatalogConfig, error) {
	home := user.HomeDir()
	if home == "" {
		return nil, fmt.Errorf("Cannot determine the user home directory")
	}
	path := filepath.Join(home, ".uc", "catalog.yaml")
	result := &CatalogConfig{}
	err := downloadYamlWithSig("https://chirino.github.io/uc/catalog.yaml", path, result)
	return result, err
}

func LoadCommandPlatforms(command string, version string) (map[string]*cache.Request, error) {
	home := user.HomeDir()
	if home == "" {
		return nil, fmt.Errorf("Cannot determine the user home directory")
	}

	path := filepath.Join(home, ".uc", "catalog", command, version, "platforms.yaml")
	result := map[string]*cache.Request{}
	err := downloadYamlWithSig(fmt.Sprintf("https://chirino.github.io/uc/catalog/%s/%s/platforms.yaml", command, version), path, &result)
	return result, err
}

func downloadYamlWithSig(url string, path string, config interface{}) error {
	err := downloadFileWithSig(url, path)
	if err != nil {
		return err
	}
	// Validate the catalog signature..
	sig, err := ioutil.ReadFile(path + ".sig")
	if err != nil {
		return err
	}
	if err := signature.CheckSignature(string(sig), path); err != nil {
		os.Remove(path) // this will trigger a re-download of the catalog..
		return fmt.Errorf("validating %s: %v", path, err)
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

func downloadFileWithSig(url string, path string) error {
	downloadCatalog := false
	if info, err := os.Stat(path); err != nil || time.Now().After(info.ModTime().Add(24*time.Hour)) {
		downloadCatalog = true
	}
	if info, err := os.Stat(path + ".sig"); err != nil || time.Now().After(info.ModTime().Add(24*time.Hour)) {
		downloadCatalog = true
	}
	if downloadCatalog {
		if err := files.WithCreate(path+".sig", func(file *os.File) error {
			return cache.HttpGet(url+".sig", file)
		}); err != nil {
			return err
		}
		if err := files.WithCreate(path, func(file *os.File) error {
			return cache.HttpGet(url, file)
		}); err != nil {
			return err
		}
	}
	return nil
}
