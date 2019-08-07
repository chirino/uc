package kubectl

import (
    "fmt"
    "github.com/chirino/uc/internal/cmd"
    "github.com/chirino/uc/internal/cmd/catalog/pkg"
    "github.com/chirino/uc/internal/cmd/utils"
    "github.com/chirino/uc/internal/pkg/cache"
    "github.com/chirino/uc/internal/pkg/dev"
    "github.com/spf13/cobra"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

func New() *cobra.Command {
    var forceDownload = false
    var startingMinor = 15
    command := &cobra.Command{
        Use:   "kubectl",
        Short: "finds new releases of kubectl",
        Run: func(c *cobra.Command, args []string) {
            _, index, err := pkg.LoadCatalogIndex()
            utils.ExitOnError(err)
            err = findNewKubectlReleases(index, forceDownload, startingMinor)
            utils.ExitOnError(err)
        },
    }
    command.Flags().BoolVarP(&forceDownload, "force", "", false, "force download all the commands")
    command.Flags().IntVarP(&startingMinor, "force", "", 15, "Kubernetes minor to start searching at.")
    return command
}

func findNewKubectlReleases(cat *cmd.CatalogIndex, forceDownload bool, startingMinor int) error {
    command := "kubectl"
minor:
    for minor := startingMinor; ; minor++ {
    micro:
        for micro := 0; ; micro++ {
            version := fmt.Sprintf("1.%d.%d", minor, micro)

            platforms := map[string]*cache.Request{}
            fn := filepath.Join(dev.GO_MOD_DIRECTORY, "docs", "catalog", command, version, "platforms.yaml")
            _, err := os.Stat(fn)
            if err == nil {
                platforms := map[string]*cache.Request{}
                err = pkg.LoadYaml(fn, platforms)
                if err != nil {
                    return err
                }
            }

            for _, platform := range []string{"windows-amd64", "darwin-amd64", "linux-amd64"} {

                if platforms[platform] != nil {
                    continue
                }
                url := fmt.Sprintf("https://dl.k8s.io/v%s/kubernetes-client-%s.tar.gz", version, platform)
                req, _ := http.NewRequest("INFO", url, nil)
                client := &http.Client{}
                resp, err := client.Do(req)
                if err != nil {
                    continue
                }
                resp.Body.Close()
                if resp.StatusCode == 200 {
                    file := "kubernetes/client/bin/kubectl"
                    if strings.HasPrefix(platform, "windows-") {
                        file += ".exe"
                    }
                    platforms[platform] = &cache.Request{
                        URL:           url,
                        ExtractTgz:    file,
                        ForceDownload: forceDownload,
                        CommandName:   command,
                        Version:       version,
                        Platform:      platform,
                    }
                    updated, err := pkg.CheckDownload(platforms[platform])
                    if err != nil {
                        return err
                    }
                    if updated {
                        err := pkg.StoreYaml(fn, platforms)
                        if err != nil {
                            return err
                        }
                        err = pkg.SignIfNeeded(fn)
                        if err != nil {
                            return err
                        }
                    }
                } else {
                    if micro == 0 {
                        break minor
                    } else {
                        break micro
                    }
                }
            }
        }
    }
    return nil
}
