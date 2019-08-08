// This package implements a sub command plugin for version.
package version

import (
	"fmt"
	"github.com/chirino/uc/internal/cmd"
	"github.com/spf13/cobra"
)

func init() {
	cmd.SubCommandFactories = append(cmd.SubCommandFactories, New)
}

var Version = "unknown"

func New(options *cmd.Options) (*cobra.Command, error) {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}, nil
}
