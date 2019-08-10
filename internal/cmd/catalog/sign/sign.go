// +build dev

package sign

import (
    "github.com/chirino/uc/internal/cmd"
    "github.com/chirino/uc/internal/cmd/catalog/pkg"
    "github.com/chirino/uc/internal/pkg/cache"
    "github.com/chirino/uc/internal/pkg/catalog"
    "github.com/chirino/uc/internal/pkg/utils"
    "github.com/spf13/cobra"
)

func New(*cmd.Options) *cobra.Command {
    var forceDownload = false
    command := &cobra.Command{
        Use:   "sign [command(s)]",
        Short: "signs the local development uc catalog in the docs directory",
        Run: func(c *cobra.Command, args []string) {
            utils.ExitOnError(run(forceDownload, args))
        },
    }
    command.Flags().BoolVarP(&forceDownload, "force", "", false, "force download all the commands")
    return command
}

func run(forceDownload bool, commands []string) error {

    catalogFileName, _, err := pkg.LoadCatalogIndex()
    if err != nil {
        return err
    }
    err = pkg.SignIfNeeded(catalogFileName)
    if err != nil {
        return err
    }

    for _, command := range commands {

        commandIndexFile := catalog.CatalogPathJoin(command+".yaml")
        err = pkg.SignIfNeeded(commandIndexFile)
        if err != nil {
            return err
        }

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

    return nil
}
