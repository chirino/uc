// This package implements a sub command plugin for camel-k.
package kamel

import (
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/cmd/utils"
	"github.com/spf13/cobra"
)

func init() {
	cmd.SubCommandFactories = append(cmd.SubCommandFactories, New)
}

func New(o *cmd.Options) *cobra.Command {
	return utils.GetCobraCommand(o, "kamel", func() (version string) {
		// Todo: figure out how to pick the best client version
		//       for the server we are connected against.  We could look at the
		//       CRD versions installed to pick the best client version.
		return "latest"
	})
}
