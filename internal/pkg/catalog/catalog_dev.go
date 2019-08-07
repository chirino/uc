// +build dev

package catalog

import (
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/pkg/cache"
	"github.com/chirino/uc/internal/pkg/dev"
	"io/ioutil"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

func LoadCatalogConfig(o *cmd.Options) (*CatalogConfig, error) {
	path := filepath.Join(dev.GO_MOD_DIRECTORY, "docs", "catalog.yaml")
	result := &CatalogConfig{}
	err := loadConfig(o, path, result)
	return result, err
}

func LoadCommandPlatforms(o *cmd.Options, command string, version string) (map[string]*cache.Request, error) {
	path := filepath.Join(dev.GO_MOD_DIRECTORY, "docs", "catalog", command, version, "platforms.yaml")
	result := map[string]*cache.Request{}
	err := loadConfig(o, path, &result)
	return result, err
}

func loadConfig(o *cmd.Options, path string, config interface{}) error {
	o.DebugF("dev mode: checking file signature: %s\n", path)
	err := CheckSigatureAgainstSigFile(path)
	if err != nil {
		return err
	}
	o.DebugF("dev mode: loading: %s\n", path)
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
