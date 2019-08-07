// +build dev

package updatecat

import (
	"encoding/base64"
	"fmt"
	"github.com/chirino/hawtgo/sh"
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/pkg/cache"
	"github.com/chirino/uc/internal/pkg/catalog"
	"github.com/chirino/uc/internal/pkg/dev"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

func init() {
	cmd.SubCommandFactories = append(cmd.SubCommandFactories, NewCmd)
}

func NewCmd(options *cmd.Options) (*cobra.Command, error) {
	var forceDownload = false
	command := &cobra.Command{
		Use:   "update-catalog",
		Short: "Updates and GPG signs the local uc catalog (only available when built with --tags dev)",
		RunE: func(c *cobra.Command, args []string) error {
			return run(forceDownload)
		},
	}
	command.Flags().BoolVarP(&forceDownload, "force", "", false, "force download all the commands")
	return command, nil
}

func run(forceDownload bool) error {

	fmt.Println("loading catalog")
	docsDir := filepath.Join(dev.GO_MOD_DIRECTORY, "docs")
	catalogFileName := filepath.Join(docsDir, "catalog.yaml")

	cat := &catalog.CatalogConfig{}
	err := loadYaml(catalogFileName, cat)
	if err != nil {
		return err
	}

	// findNewKubectlReleases(catalog)
	for command, _ := range cat.Commands {
		err := forCommandPlatforms(command, func(version string, fn string, platforms map[string]*cache.Request) error {
			for platform, request := range platforms {
				request.ForceDownload = forceDownload
				request.CommandName = command
				request.Version = version
				request.Platform = platform
				updated, err := checkDownload(request)
				if err != nil {
					return err
				}
				if updated {
					err := storeYaml(fn, platforms)
					if err != nil {
						return err
					}
					err = signIfNeeded(fn)
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

	err = signIfNeeded(catalogFileName)
	if err != nil {
		return err
	}

	return nil
}

func forCommandPlatforms(command string, action func(version string, file string, platforms map[string]*cache.Request) error) error {
	docsDir := filepath.Join(dev.GO_MOD_DIRECTORY, "docs")
	commandDir := filepath.Join(docsDir, "catalog", command)
	files, err := ioutil.ReadDir(commandDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		version := f.Name()
		fn := filepath.Join(commandDir, version, "platforms.yaml")
		_, err := os.Stat(fn)
		if err != nil {
			continue
		}

		platforms := map[string]*cache.Request{}
		err = loadYaml(fn, platforms)
		if err != nil {
			return err
		}
		action(version, fn, platforms)
	}
	return nil
}

func signIfNeeded(fn string) error {
	err := catalog.CheckSigatureAgainstSigFile(fn)
	if err != nil {
		err = sign(fn)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadYaml(fileName string, cat interface{}) error {
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

func sign(file string) error {
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

func checkDownload(request *cache.Request) (bool, error) {
	fmt.Printf("Checking %s, %s, %s\n", request.CommandName, request.Version, request.Platform)
	request.SkipVerification = true
	request.Printf = cmd.StdErrPrintf
	file, err := cache.Get(request)
	if err != nil {
		return false, err
	}
	request.SkipVerification = false
	if request.Size != 0 && request.Signature != "" {
		if err := cache.Verify(request, file); err != nil {
			fmt.Println("verification failed: ", err)
			if err := updateVerification(request, file); err != nil {
				return false, err
			}
			return true, nil
		}
	} else {
		fmt.Println("please sign: ", file)
		if err := updateVerification(request, file); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func findNewKubectlReleases(cat *catalog.CatalogConfig, forceDownload bool) error {
	command := "kubectl"
minor:
	for minor := 15; ; minor++ {
	micro:
		for micro := 0; ; micro++ {
			version := fmt.Sprintf("1.%d.%d", minor, micro)

			platforms := map[string]*cache.Request{}
			fn := filepath.Join(dev.GO_MOD_DIRECTORY, "docs", "catalog", command, version, "platforms.yaml")
			_, err := os.Stat(fn)
			if err == nil {
				platforms := map[string]*cache.Request{}
				err = loadYaml(fn, platforms)
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
					updated, err := checkDownload(platforms[platform])
					if err != nil {
						return err
					}
					if updated {
						err := storeYaml(fn, platforms)
						if err != nil {
							return err
						}
						err = signIfNeeded(fn)
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

func storeYaml(cat string, config interface{}) error {
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

func store(path string) (string, error) {
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
