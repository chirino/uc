// +build !dev

package catalog

import (
	"fmt"
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/pkg/cache"
	"github.com/chirino/uc/internal/pkg/files"
	"github.com/chirino/uc/internal/pkg/signature"
	"github.com/chirino/uc/internal/pkg/user"
	"io/ioutil"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

var CatalogBaseURL = "https://chirino.github.io/uc/catalog"

func LoadCatalogConfig(o *cmd.Options) (*CatalogConfig, error) {
	home := user.HomeDir()
	if home == "" {
		return nil, fmt.Errorf("Cannot determine the user home directory")
	}
	path := filepath.Join(home, ".uc", "cache", "catalog", "index.yaml")
	result := &CatalogConfig{}
	err := downloadYamlWithSig(o, CatalogBaseURL+"/index.yaml", path, result)
	return result, err
}

func LoadCommandPlatforms(o *cmd.Options, command string, version string) (map[string]*cache.Request, error) {
	home := user.HomeDir()
	if home == "" {
		return nil, fmt.Errorf("Cannot determine the user home directory")
	}

	path := filepath.Join(home, ".uc", "cache", "catalog", command, version, "platforms.yaml")
	result := map[string]*cache.Request{}
	url := fmt.Sprintf(CatalogBaseURL+"/%s/%s/platforms.yaml", command, version)
	err := downloadYamlWithSig(o, url, path, &result)
	return result, err
}

func downloadYamlWithSig(o *cmd.Options, url string, path string, config interface{}) error {
	err := downloadFileWithSig(o, url, path)
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

func downloadFileWithSig(o *cmd.Options, url string, path string) error {
	sigpath := path + ".sig"
	sigurl := url + ".sig"

	skip := true
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(o.DebugLog, "missing: %s\n", path)
		skip = false
	} else if info.ModTime().Before(o.CacheExpires) {
		fmt.Fprintln(o.DebugLog, "expired from the cache (stale by %s): %s\n", info.ModTime().Sub(o.CacheExpires), path)
		skip = false
	}

	info, err = os.Stat(sigpath)
	if err != nil {
		fmt.Fprintln(o.DebugLog, "missing: %s\n", sigpath)
		skip = false
	} else if info.ModTime().Before(o.CacheExpires) {
		fmt.Fprintln(o.DebugLog, "expired from the cache (stale by %s): %s\n", info.ModTime().Sub(o.CacheExpires), path)
		skip = false
	}

	if skip {
		fmt.Fprintln(o.DebugLog, "download skipped (cache expires in %s): %s\n", info.ModTime().Sub(o.CacheExpires), path)
	} else {
		fmt.Fprintln(o.DebugLog, "downloading: %s\n", sigurl)
		if err := files.WithCreate(sigpath, func(file *os.File) error {
			return cache.HttpGet(sigurl, file)
		}); err != nil {
			return err
		}
		fmt.Fprintln(o.DebugLog, "stored: %s\n", sigpath)

		fmt.Fprintln(o.DebugLog, "downloading: %s\n", url)
		if err := files.WithCreate(path, func(file *os.File) error {
			return cache.HttpGet(url, file)
		}); err != nil {
			return err
		}
		fmt.Fprintln(o.DebugLog, "stored: %s\n", path)
	}
	return nil
}
