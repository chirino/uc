package catalog

import (
	"fmt"
	"github.com/chirino/uc/internal/pkg/signature"
	"golang.org/x/crypto/openpgp"
	"io/ioutil"
)

var DefaultCatalogBaseURL = "https://chirino.github.io/uc/catalog/v1"

func CheckSigatureAgainstSigFile(keyring openpgp.EntityList, path string) error {
	sigpath := path + ".sig"
	sig, err := ioutil.ReadFile(sigpath)
	if err != nil {
		return err
	}
	if err := signature.CheckSignature(keyring, string(sig), path); err != nil {
		return fmt.Errorf("validating %s: %v", path, err)
	}
	return nil
}
