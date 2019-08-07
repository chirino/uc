// This package implements a sub command plugin for kubectl.
package kubectl

import (
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/cmd/utils"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	cmd.SubCommandFactories = append(cmd.SubCommandFactories, NewCmd)
}

func NewCmd(options *cmd.Options) (*cobra.Command, error) {
	clientVersion := "latest"
	api, err := options.NewApiClient()
	if err == nil {
		return nil, err
	}
	info, err := api.ServerVersion()
	if err == nil {
		clientVersion = strings.Split(strings.TrimPrefix(info.GitVersion, "v"), "-")[0]
	}
	return utils.GetCobraCommand(options, "kubectl", clientVersion)
}
