package utils

import (
	"fmt"
	"github.com/chirino/hawtgo/sh"
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/pkg/cache"
	catalog2 "github.com/chirino/uc/internal/pkg/catalog"
	"github.com/spf13/cobra"
	"runtime"
)

func GetExecutable(options *cmd.Options, command string, version string) (string, error) {
	catalog, err := catalog2.LoadCatalogConfig()
	if err != nil {
		return "", err
	}

	config := catalog.Commands[command]
	if config == nil {
		return "", fmt.Errorf("command not found in catalog: %s", command)
	}

	if version == "latest" {
		version = config.LatestVersion
	}

	platforms, err := catalog2.LoadCommandPlatforms(command, version)
	if err != nil {
		return "", err
	}
	if platforms == nil {
		return "", fmt.Errorf("%s version not found in catalog: %s", command, version)
	}

	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	request := platforms[platform]
	if request == nil {
		return "", fmt.Errorf("%s %s platform not found in catalog: %s", command, version, platform)
	}
	request.CommandName = command
	request.Platform = platform
	request.Version = version
	request.Printf = options.Printf
	return cache.Get(request)
}

func GetCobraCommand(options *cmd.Options, command string, clientVersion string) (*cobra.Command, error) {
	return &cobra.Command{
		Use:                command,
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {

			// Get the executable for that client version...
			executable, err := GetExecutable(options, command, clientVersion)
			if err != nil {
				return err
			}

			// call it pass along any args....
			return sh.New().LineArgs(append([]string{executable}, args...)...).Exec()
		},
	}, nil
}
