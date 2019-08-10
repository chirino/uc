// +build dev

package catalog

import (
	"fmt"
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/pkg/cache"
	"github.com/chirino/uc/internal/pkg/dev"
	"github.com/chirino/uc/internal/pkg/signature"
	"golang.org/x/crypto/openpgp"
	"io/ioutil"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

func CatalogPathJoin(elem ...string) string {
	return filepath.Join(append([]string{dev.GO_MOD_DIRECTORY, "docs", "catalog", "v1"}, elem...)...)
}

func LoadCatalogIndex(o *cmd.Options) (*cmd.CatalogIndex, error) {
	path := CatalogPathJoin("index.yaml")
	result := &cmd.CatalogIndex{}
	err := load(o, signature.DefaultPublicKeyring, path, result)
	return result, err
}

func LoadCommandIndex(o *cmd.Options, keyring openpgp.EntityList, catalogBaseURL string, command string) (*cmd.CatalogCommandIndex, error) {
	path := CatalogPathJoin(command + ".yaml")
	result := &cmd.CatalogCommandIndex{}
	err := load(o, signature.DefaultPublicKeyring, path, result)
	return result, err
}

func LoadCommandPlatforms(o *cmd.Options, keyring openpgp.EntityList, catalogBaseURL string, command string, version string) (map[string]*cache.Request, error) {
	path := CatalogPathJoin(command, version+".yaml")
	result := map[string]*cache.Request{}
	err := load(o, keyring, path, &result)
	return result, err
}

func load(o *cmd.Options, keyring openpgp.EntityList, path string, config interface{}) error {
	fmt.Fprintln(o.DebugLog, "dev mode: checking file signature:", path)
	err := CheckSigatureAgainstSigFile(keyring, path)
	if err != nil {
		return err
	}
	fmt.Fprintln(o.DebugLog, "dev mode: loading:", path)
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
