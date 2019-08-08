package utils

import (
	"fmt"
	"github.com/chirino/hawtgo/sh"
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/pkg/cache"
	"github.com/chirino/uc/internal/pkg/catalog"
	"github.com/chirino/uc/internal/pkg/signature"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/openpgp"
	"os"
	"runtime"
	"strings"
)

func GetExecutable(options *cmd.Options, command string, version string) (string, error) {
	index, err := catalog.LoadCatalogIndex(options)
	if err != nil {
		return "", err
	}

	commandCatalog := index.Commands[command]
	if commandCatalog == nil {
		return "", fmt.Errorf("command not found in catalog: %s", command)
	}

	if version == "latest" {
		version = commandCatalog.LatestVersion
	}

	keyring := signature.DefaultPublicKeyring
	if commandCatalog.CatalogPublicKey != "" {
		k, err := openpgp.ReadArmoredKeyRing(strings.NewReader(commandCatalog.CatalogPublicKey))
		if err != nil {
			fmt.Errorf("invalid catalog-public-key configured for command %s: %s", command, err)
		}
		keyring = k
	}

	baseurl := catalog.DefaultCatalogBaseURL
	if commandCatalog.CatalogBaseURL != "" {
		baseurl = commandCatalog.CatalogBaseURL
	}

	platforms, err := catalog.LoadCommandPlatforms(options, keyring, baseurl, command, version)
	if err != nil {
		return "", err
	}
	if platforms == nil {
		return "", fmt.Errorf("%s version not found in catalog: %s", command, version)
	}

	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	request := platforms[platform]
	if request == nil {
		supported := []string{}
		for p, _ := range platforms {
			supported = append(supported, p)
		}
		return "", fmt.Errorf("your platform %s is not supported by the %s command, it is available on: %s", platform, command, supported)
	}

	request.Keyring = keyring
	request.CommandName = command
	request.Platform = platform
	request.Version = version
	request.InfoLog = options.InfoLog
	request.DebugLog = options.DebugLog
	return cache.Get(request)
}

func GetCobraCommand(o *cmd.Options, command string, clientVersion string) (*cobra.Command, error) {
	return &cobra.Command{
		Use:                command,
		DisableFlagParsing: true,
		Run: func(c *cobra.Command, args []string) {

			// Get the executable for that client version...
			executable, err := GetExecutable(o, command, clientVersion)
			if err != nil {
				fmt.Fprintln(o.InfoLog, "could not get a suitable command executable:", err)
				os.Exit(2)
			}

			// call it pass along any args....
			err = sh.New().CommandLog(o.DebugLog).CommandLogPrefix("running > ").LineArgs(append([]string{executable}, args...)...).Exec()
			if err != nil {
				fmt.Fprintln(o.InfoLog, "error:", err)
				os.Exit(3)
			}

		},
	}, nil
}
