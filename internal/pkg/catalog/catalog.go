package catalog

import (
	"fmt"
	"github.com/chirino/uc/internal/pkg/signature"
	"io/ioutil"
)

type CatalogConfig struct {
	Update   string                     `json:"update,omitempty"`
	Commands map[string]*CatalogCommand `json:"commands,omitempty"`
}

type CatalogCommand struct {
	Short         string `json:"short-description,omitempty"`
	Long          string `json:"long-description,omitempty"`
	LatestVersion string `json:"latest-version,omitempty"`
}

func CheckSigatureAgainstSigFile(path string) error {
	sigpath := path + ".sig"
	sig, err := ioutil.ReadFile(sigpath)
	if err != nil {
		return err
	}
	if err := signature.CheckSignature(string(sig), path); err != nil {
		return fmt.Errorf("validating %s: %v", path, err)
	}
	return nil
}
