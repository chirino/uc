// This package implements a sub command plugin for kubectl.
package kubectl

import (
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/cmd/utils"
	"github.com/spf13/cobra"
	"regexp"
)

func init() {
	cmd.SubCommandFactories = append(cmd.SubCommandFactories, New)
}

func New(o *cmd.Options) *cobra.Command {

	return utils.GetCobraCommand(o, "kubectl", func() (version string) {
		api, err := o.NewApiClient()
		if err != nil {
			return "latest"
		}
		info, err := api.ServerVersion()
		if err != nil {
			return "latest"
		}

		re := regexp.MustCompile(`^(v\d+.\d+.\d+)`)
		if matches := re.FindAllStringSubmatch(info.GitVersion, -1); len(matches) > 0 {
			result := matches[0][1]
			return result
		}
		return "latest"
	})
}
