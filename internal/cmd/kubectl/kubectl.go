// This package implements a sub command plugin for kubectl.
package kubectl

import (
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/cmd/utils"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	cmd.SubCommandFactories = append(cmd.SubCommandFactories, New)
}

func New(o *cmd.Options) *cobra.Command {

	return utils.GetCobraCommand(o, "kamel", func() (version string) {
		api, err := o.NewApiClient()
		if err != nil {
			return "latest"
		}
		info, err := api.ServerVersion()
		if err != nil {
			return "latest"
		}
		return strings.Split(strings.TrimPrefix(info.GitVersion, "v"), "-")[0]
	})
}
