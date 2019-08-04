// +build dev

package updatecat

import (
    "encoding/base64"
    "fmt"
    "github.com/chirino/hawtgo/sh"
    "github.com/chirino/uc/internal/cmd"
    "github.com/chirino/uc/internal/pkg/cache"
    "github.com/chirino/uc/internal/pkg/dev"
    "github.com/spf13/cobra"
    "io/ioutil"
    "k8s.io/apimachinery/pkg/version"
    "k8s.io/client-go/kubernetes"
    "os"
    "path/filepath"
    "sigs.k8s.io/yaml"
)

func init() {
    cmd.SubCmdFactories = append(cmd.SubCmdFactories, NewCmd)
}

func NewCmd(options *cmd.Options, api *kubernetes.Clientset, info *version.Info) (*cobra.Command, error) {
    var forceDownload = false
    command := &cobra.Command{
        Use: "update-catalog",
        RunE: func(c *cobra.Command, args []string) error {
            return run(forceDownload)
        },
    }
    command.Flags().BoolVarP(&forceDownload, "force", "", false, "force download all the commands")
    return command, nil
}

func run(forceDownload bool) error {

    fmt.Println("loading catalog")
    config, err := cmd.LoadConfig()
    if err != nil {
        return err
    }

    updates := 0
    for command, value := range config.Commands {
        for version, platforms := range value.Versions {
            for platform, request := range platforms {
                fmt.Printf("Checking %s, %s, %s\n", command, version, platform)
                request.SkipVerification = true
                request.ForceDownload = forceDownload
                request.CommandName = command
                request.Version = version
                request.Printf = cmd.StdErrPrintf
                file, err := cache.Get(request)
                if err != nil {
                    return err
                }
                request.SkipVerification = false
                if request.Size != 0 && request.Signature != "" {
                    if err := cache.Verify(request, file); err != nil {
                        fmt.Println("verification failed: ", err)
                        if err := updateVerification(request, file); err != nil {
                            return err
                        }
                        updates += 1
                    }
                } else {
                    fmt.Println("please sign: ", file)
                    if err := updateVerification(request, file); err != nil {
                        return err
                    }
                    updates += 1
                }
            }
        }
    }

    if updates > 0 {

        bytes, err := yaml.Marshal(config)
        if err != nil {
            return err
        }

        fmt.Println("storing catalog")
        path := filepath.Join(dev.GO_MOD_DIRECTORY, "docs", "catalog.yaml")
        ioutil.WriteFile(path, bytes, 0755)
        if err != nil {
            return err
        }

        sigEncoded, err := gpgSign(path)
        if err != nil {
            return err
        }

        err = ioutil.WriteFile(path+".sig", []byte(sigEncoded), 0755)
        if err != nil {
            return err
        }
    } else {
        fmt.Println("No catalog updates needed.")
    }

    return nil
}

func gpgSign(path string) (string, error) {
    sigRaw, _, err := sh.New().LineArgs(`gpg`, `--output`, `-`, `--detach-sig`, path).Output(sh.OutputOptions{NoTrim: true})
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString([]byte(sigRaw)), nil
}

func updateVerification(request *cache.Request, file string) error {
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
