// +build dev

package sign

import (
    "github.com/chirino/uc/internal/cmd/catalog/pkg"
    "github.com/chirino/uc/internal/pkg/cache"
    "github.com/spf13/cobra"
)

func New() *cobra.Command {
    var forceDownload = false
    command := &cobra.Command{
        Use:   "sign",
        Short: "signs the local development uc catalog in the docs directory",
        RunE: func(c *cobra.Command, args []string) error {
            return run(forceDownload)
        },
    }
    command.Flags().BoolVarP(&forceDownload, "force", "", false, "force download all the commands")
    return command
}

func run(forceDownload bool) error {

    catalogFileName, index, err := pkg.LoadCatalogIndex()
    if err != nil {
        return err
    }

    for command, _ := range index.Commands {
        err := pkg.ForCommandPlatforms(command, func(version string, fn string, platforms map[string]*cache.Request) error {
            for platform, request := range platforms {
                request.ForceDownload = forceDownload
                request.CommandName = command
                request.Version = version
                request.Platform = platform
                updated, err := pkg.CheckDownload(request)
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
            }
            return nil
        })
        if err != nil {
            return err
        }

    }

    err = pkg.SignIfNeeded(catalogFileName)
    if err != nil {
        return err
    }

    return nil
}

