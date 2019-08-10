// +build !dev

package catalog

import (
    "fmt"
    "github.com/chirino/uc/internal/cmd"
    "github.com/chirino/uc/internal/pkg/cache"
    "github.com/chirino/uc/internal/pkg/files"
    "github.com/chirino/uc/internal/pkg/signature"
    "github.com/chirino/uc/internal/pkg/user"
    "golang.org/x/crypto/openpgp"
    "io"
    "io/ioutil"
    "os"
    "path"
    "path/filepath"
    "sigs.k8s.io/yaml"
)

func CatalogPath(elem []string) (string, error) {
    home := user.HomeDir()
    if home == "" {
        return "", fmt.Errorf("Cannot determine the user home directory")
    }
    return filepath.Join(append(join(home, ".uc", "cache", "catalog", "v1"), elem...)...), nil
}

func LoadCatalogIndex(o *cmd.Options) (*cmd.CatalogIndex, error) {
    if o.CatalogIndex != nil {
        return o.CatalogIndex, nil
    }
    result := &cmd.CatalogIndex{}
    err := fetch(o, signature.DefaultPublicKeyring, DefaultCatalogBaseURL, join("index.yaml"), result)
    if err != nil {
        return nil, err
    }
    o.CatalogIndex = result
    return result, nil
}

func LoadCommandIndex(o *cmd.Options, keyring openpgp.EntityList, catalogBaseURL string, command string) (*cmd.CatalogCommandIndex, error) {
    result := &cmd.CatalogCommandIndex{}
    err := fetch(o, keyring, catalogBaseURL, join(command+".yaml"), result)
    if err != nil {
        return nil, err
    }
    return result, err
}

func LoadCommandPlatforms(o *cmd.Options, keyring openpgp.EntityList, catalogBaseURL string, command string, version string) (map[string]*cache.Request, error) {
    result := map[string]*cache.Request{}
    err := fetch(o, keyring, catalogBaseURL, join(command, version+".yaml"), &result)
    if err != nil {
        return nil, err
    }
    return result, err
}

func join(first string, extra ...string) []string {
    return append([]string{first}, extra...)
}

func fetch(o *cmd.Options, keyring openpgp.EntityList, urlBase string, filePath []string, config interface{}) error {
    p, err := CatalogPath(filePath)
    if err != nil {
        return err
    }
    url := path.Join(join(urlBase, filePath...)...)

    err = downloadFileWithSig(o, url, p)
    if err != nil {
        return err
    }
    // Validate the catalog signature..
    sig, err := ioutil.ReadFile(p + ".sig")
    if err != nil {
        return err
    }
    if err := signature.CheckSignature(keyring, string(sig), p); err != nil {
        os.Remove(p) // this will trigger a re-download of the catalog..
        return fmt.Errorf("validating %s: %v", p, err)
    }
    file, err := os.Open(p)
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
        fmt.Fprintln(o.DebugLog, "missing:", path)
        skip = false
    } else if info.ModTime().Before(o.CacheExpires) {
        fmt.Fprintf(o.DebugLog, "expired from the cache (stale by %s): %s\n", info.ModTime().Sub(o.CacheExpires), path)
        skip = false
    }

    info, err = os.Stat(sigpath)
    if err != nil {
        fmt.Fprintln(o.DebugLog, "missing:", sigpath)
        skip = false
    } else if info.ModTime().Before(o.CacheExpires) {
        fmt.Fprintf(o.DebugLog, "expired from the cache (stale by %s): %s\n", info.ModTime().Sub(o.CacheExpires), path)
        skip = false
    }

    if skip {
        fmt.Fprintf(o.DebugLog, "download skipped (cache expires in %s): %s\n", info.ModTime().Sub(o.CacheExpires), path)
    } else {
        if err := download(o, sigurl, sigpath); err != nil {
            return err
        }
        if err := download(o, url, path); err != nil {
            return err
        }
    }
    return nil
}

func download(o *cmd.Options, url string, toFileName string) error {
    fmt.Fprintln(o.InfoLog, "downloading:", url)
    err := cache.WithHttpGetReader(url, func(src io.Reader) error {
        return files.WithCreateThenReplace(toFileName, 0644, func(dst *os.File) error {
            _, err := io.Copy(dst, src)
            return err
        })
    })
    if err != nil {
        return err
    }
    fmt.Fprintln(o.InfoLog, "wrote:", toFileName)
    return nil
}
