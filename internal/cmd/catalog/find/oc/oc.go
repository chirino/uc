// +build dev

package oc

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
    "regexp"
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
        Use:   "oc",
        Short: "finds new releases of oc",
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
    command := "oc"
    platformMap := map[string]string{
        "windows":       "windows-amd64",
        "windows-amd64": "windows-amd64",
        "mac":           "darwin-amd64",
        "darwin-amd64":  "darwin-amd64",
        "linux-64bit":   "linux-amd64",
        "linux-amd64":   "linux-amd64",
        "linux-32bit":   "linux-386",
        "linux-386":     "linux-386",
    }

    return o.GithubUtils.ListGithubReleases("openshift", "origin", func(release *github.RepositoryRelease) error {
        if len(release.Assets) == 0 || release.Name == nil {
            return nil
        }

        version := *release.Name
        platforms := map[string]*cache.Request{}
        fn := catalog.CatalogPathJoin(command, version+".yaml")
        _ = pkg.LoadYaml(fn, &platforms)

        platformsFound := 0
        updates := 0
        for _, prefix := range []string{"openshift-origin-client-tools-", "openshift-origin-"} {
            for _, a := range release.Assets {
                downloadURL := o.GithubUtils.AuthURL(*a.BrowserDownloadURL)
                file := path.Base(downloadURL)
                if !strings.HasPrefix(file, prefix) {
                    continue
                }

                // matches: (openshift-origin-client-tools-v3.11.0)-(0cbc58b)-(mac).(zip)
                // matches: (openshift-origin-client-tools-v1.3.3)-(bc17c1527938fa03b719e1a117d584442e3727b8)-(linux-32bit).(tar.gz)
                re := regexp.MustCompile(`(.*)(-|\.)([\dabcdef]{7,40})-(\w+(-[\w\d]+)?)\.(tar\.gz|zip)`)
                firstPart := ""
                //sepPart := ""
                hashPart := ""
                ocPlatform := ""
                if subs := re.FindAllStringSubmatch(file, -1); len(subs) > 0 {
                    firstPart = subs[0][1];
                    //sepPart = subs[0][2];
                    hashPart = subs[0][3];
                    ocPlatform = subs[0][4];
                }

                platform := platformMap[ocPlatform]
                if platform == "" {
                    return fmt.Errorf("Unsupported platform part: %s", ocPlatform)
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

                if request.ExtractTgz == "" && request.ExtractZip == "" {
                    exeName := "oc"
                    if strings.HasPrefix(platform, "windows-") {
                        exeName = "oc.exe"
                    }
                    paths := exeName
                    paths += "|./" + exeName
                    paths += "|" + fmt.Sprintf("%s-%s-%s/%s", firstPart, hashPart, ocPlatform, exeName)
                    paths += "|" + fmt.Sprintf("%s+%s-%s/%s", firstPart, hashPart, ocPlatform, exeName)

                    if strings.HasSuffix(file, ".tar.gz") {
                        request.ExtractTgz = paths
                    }
                    if strings.HasSuffix(file, ".zip") {
                        request.ExtractZip = paths
                    }
                }

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
            if platformsFound != 0 {
                break
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
