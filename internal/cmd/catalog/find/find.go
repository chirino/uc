// +build dev

package find

import (
    "github.com/chirino/uc/internal/cmd"
    "github.com/chirino/uc/internal/cmd/catalog/find/kamel"
    "github.com/chirino/uc/internal/cmd/catalog/find/kn"
    "github.com/chirino/uc/internal/cmd/catalog/find/kubectl"
    "github.com/chirino/uc/internal/cmd/catalog/find/oc"
    "github.com/chirino/uc/internal/cmd/catalog/find/odo"
    "github.com/chirino/uc/internal/cmd/catalog/find/uc"
    "github.com/spf13/cobra"
)

func New(o *cmd.Options) *cobra.Command {
    command := &cobra.Command{
        Use:   "find",
        Short: "finds new releases of a sub command",
    }
    command.AddCommand(kubectl.New(o))
    command.AddCommand(oc.New(o))
    command.AddCommand(odo.New(o))
    command.AddCommand(kamel.New(o))
    command.AddCommand(kn.New(o))
    command.AddCommand(uc.New(o))
    return command
}
