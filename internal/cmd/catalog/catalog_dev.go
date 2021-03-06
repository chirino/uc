// +build dev

package catalog

import (
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/cmd/catalog/find"
	"github.com/chirino/uc/internal/cmd/catalog/sign"
	"github.com/spf13/cobra"
)

func init() {
	cmd.SubCommandFactories = append(cmd.SubCommandFactories, New)
}

func New(options *cmd.Options) *cobra.Command {
	command := &cobra.Command{
		Use:   "catalog",
		Short: "Tools to manage the uc catalog",
	}
	command.AddCommand(sign.New(options))
	command.AddCommand(find.New(options))
	return command
}
