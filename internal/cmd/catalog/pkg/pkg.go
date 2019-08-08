package pkg

import (
    "encoding/base64"
    "fmt"
    "github.com/chirino/hawtgo/sh"
    "github.com/chirino/uc/internal/cmd"
    "github.com/chirino/uc/internal/pkg/cache"
    "github.com/chirino/uc/internal/pkg/catalog"
    "github.com/chirino/uc/internal/pkg/dev"
    "github.com/chirino/uc/internal/pkg/signature"
    "io/ioutil"
    "os"
    "path/filepath"
    "sigs.k8s.io/yaml"
    "strings"
)

func LoadCatalogIndex() (string, *cmd.CatalogIndex, error) {
    fmt.Println("loading catalog")
    docsDir := filepath.Join(dev.GO_MOD_DIRECTORY, "docs")
    catalogFileName := filepath.Join(docsDir, "catalog", "index.yaml")
    cat := &cmd.CatalogIndex{}
    err := LoadYaml(catalogFileName, cat)
    if err != nil {
        return "", nil, err
    }
    return catalogFileName, cat, nil
}

func SignIfNeeded(fn string) error {
    err := catalog.CheckSigatureAgainstSigFile(signature.DefaultPublicKeyring, fn)
    if err != nil {
        err = Sign(fn)
        if err != nil {
            return err
        }
    }
    return nil
}

func LoadYaml(fileName string, cat interface{}) error {
    file, err := os.Open(fileName)
    if err != nil {
        return err
    }
    defer file.Close()
    bytes, err := ioutil.ReadAll(file)
    if err != nil {
        return err
    }
    err = yaml.Unmarshal(bytes, cat)
    if err != nil {
        return err
    }
    return nil
}

func Sign(file string) error {
    fmt.Println("writing: ", file+".sig")
    sigEncoded, err := gpgSign(file)
    if err != nil {
        return err
    }
    err = ioutil.WriteFile(file+".sig", []byte(sigEncoded), 0755)
    if err != nil {
        return err
    }
    return nil
}

func CheckDownload(request *cache.Request) (bool, error) {
    fmt.Printf("Checking %s, %s, %s\n", request.CommandName, request.Version, request.Platform)
    request.Keyring = nil
    request.InfoLog = os.Stderr
    file, err := cache.Get(request)
    if err != nil {
        return false, err
    }
    request.Keyring = signature.DefaultPublicKeyring
    if request.Size != 0 && request.Signature != "" {
        if err := cache.Verify(request, file); err != nil {
            fmt.Println("verification failed: ", err)
            if err := UpdateVerification(request, file); err != nil {
                return false, err
            }
            return true, nil
        }
    } else {
        fmt.Println("please sign: ", file)
        if err := UpdateVerification(request, file); err != nil {
            return false, err
        }
        return true, nil
    }
    return false, nil
}

func UpdateVerification(request *cache.Request, file string) error {
    sig, err := gpgSign(file)
    if err != nil {
        return err
    }
    request.Signature = sig

    info, err := os.Stat(file)
    if err != nil {
        return err
    }
    request.Size = info.Size()
    return nil
}

func StoreYaml(cat string, config interface{}) error {
    bytes, err := yaml.Marshal(config)
    if err != nil {
        return err
    }
    fmt.Println("storing catalog")
    ioutil.WriteFile(cat, bytes, 0755)
    if err != nil {
        return err
    }
    return nil
}

func Store(path string) (string, error) {
    sigRaw, _, err := sh.New().LineArgs(`gpg`, `--output`, `-`, `--detach-sig`, path).Output(sh.OutputOptions{NoTrim: true})
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString([]byte(sigRaw)), nil
}

func gpgSign(path string) (string, error) {
    sigRaw, _, err := sh.New().LineArgs(`gpg`, `--output`, `-`, `--detach-sig`, path).Output(sh.OutputOptions{NoTrim: true})
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString([]byte(sigRaw)), nil
}

func ForCommandPlatforms(command string, action func(version string, file string, platforms map[string]*cache.Request) error) error {
    docsDir := filepath.Join(dev.GO_MOD_DIRECTORY, "docs")
    commandDir := filepath.Join(docsDir, "catalog", command)
    files, err := ioutil.ReadDir(commandDir)
    if err != nil {
        return err
    }

    for _, f := range files {
        if strings.HasSuffix(f.Name(), ".yaml") {
            platforms := map[string]*cache.Request{}
            fileName := filepath.Join(commandDir, f.Name())
            err = LoadYaml(fileName, &platforms)
            if err != nil {
                return err
            }
            err = action(strings.TrimSuffix(f.Name(), ".yaml"), fileName, platforms)
            if err != nil {
                return err
            }
        }
    }
    return nil
}
