// +build dev

package kubectl

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
        Use:   "kubectl",
        Short: "finds new releases of kubectl",
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
    command := "kubectl"
    return o.GithubUtils.ListGithubReleases("openshift", "origin", func(release *github.RepositoryRelease) error {
        if len(release.Assets) == 0 || release.Name == nil {
            return nil
        }

        version := *release.Name
        platforms := map[string]*cache.Request{}
        fn := catalog.CatalogPathJoin(command, version+".yaml")
        _ = pkg.LoadYaml(fn, &platforms)
        updates := 0
        for _, platform := range []string{"windows-amd64", "darwin-amd64", "linux-amd64"} {

            request := platforms[platform]
            if request == nil {
                request = &cache.Request{}
                platforms[platform] = request
            }

            url := fmt.Sprintf("https://dl.k8s.io/%s/kubernetes-client-%s.tar.gz", version, platform)
            request.URL = url
            request.ForceDownload = o.ForceDownload
            request.CommandName = command
            request.Version = version
            request.Platform = platform
            file := "kubernetes/client/bin/kubectl"
            if strings.HasPrefix(platform, "windows-") {
                file += ".exe"
            }
            request.ExtractTgz = file

            updated, err := pkg.CheckDownload(platforms[platform])
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
        if updates == 0 && !o.ForceDownload {
            fmt.Printf("Previously had %s in the catalog.  Usee --force option to reprocess past this release.\n", version)
            return io.EOF
        }
        return nil
    })
}
