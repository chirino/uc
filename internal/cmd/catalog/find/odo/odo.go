// +build dev

package odo

import (
    "fmt"
    "github.com/chirino/uc/internal/cmd"
    "github.com/chirino/uc/internal/cmd/catalog/find/findutil"
    "github.com/chirino/uc/internal/cmd/catalog/pkg"
    "github.com/chirino/uc/internal/pkg/cache"
    "github.com/chirino/uc/internal/pkg/catalog"
    "github.com/chirino/uc/internal/pkg/utils"
    "github.com/google/go-github/v27/github"
    "github.com/spf13/cobra"
    "io"
    "path"
    "strings"
)

type Options struct {
    *findutil.GithubUtils
    ForceDownload bool
}

func New(global *cmd.Options) *cobra.Command {
    o := &Options{
        GithubUtils: &findutil.GithubUtils{},
    }
    o.Context = global.Context
    command := &cobra.Command{
        Use:   "odo",
        Short: "finds new releases of odo",
        Run: func(c *cobra.Command, args []string) {
            utils.ExitOnError(o.findGithubReleases())
        },
    }
    command.Flags().BoolVarP(&o.ForceDownload, "force", "", false, "force download all the commands")
    command.Flags().StringVar(&o.GithubUser, "ghuser", "", "your github username (only needed if you hit api limits)")
    command.Flags().StringVar(&o.GithubPersonalAccessToken, "ghtoken", "", "your personal access token username (only needed if you hit api limits)")
    return command
}

func (o *Options) findGithubReleases() error {
    command := "odo"
    platformMap := map[string]string{}
    for _, value := range utils.Platforms {
        platformMap[value]=value
    }

    return o.GithubUtils.ListGithubReleases("openshift", "odo", func(release *github.RepositoryRelease) error {
        if len(release.Assets) == 0 || release.Name == nil {
            return nil
        }

        version := *release.Name
        platforms := map[string]*cache.Request{}
        fn := catalog.CatalogPathJoin(command, version+".yaml")
        _ = pkg.LoadYaml(fn, &platforms)

        platformsFound := 0
        updates := 0
        for _, a := range release.Assets {

            downloadURL := o.GithubUtils.AuthURL(*a.BrowserDownloadURL)
            file := path.Base(downloadURL)

            filePrefix := "odo-"
            fileSuffix := ".gz"
            if !strings.HasPrefix(file, filePrefix) || !strings.HasSuffix(file, fileSuffix) {
                continue
            }

            filePlatform := strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(file, filePrefix), fileSuffix), ".exe")
            platform := platformMap[filePlatform]
            if platform == "" {
                return fmt.Errorf("Unsupported platform part: %s", filePlatform)
            }

            platformsFound += 1
            request := platforms[platform]
            if request == nil {
                request = &cache.Request{}
                platforms[platform] = request
            }

            request.URL = downloadURL
            request.ForceDownload = o.ForceDownload
            request.CommandName = command
            request.Version = version
            request.Platform = platform
            request.Uncompress = "gz"

            updated, err := pkg.CheckDownload(request)
            if err != nil {
                return err
            }
            if updated {
                updates += 1
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

        if platformsFound == 0 {
            return fmt.Errorf("No client tools found in release: ", version)
        }

        if updates == 0 && !o.ForceDownload {
            fmt.Printf("Previously had %s in the catalog.  Usee --force option to reprocess past this release.\n", version)
            return io.EOF
        }
        return nil
    })
}
