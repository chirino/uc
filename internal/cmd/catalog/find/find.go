package find

import (
    "github.com/chirino/uc/internal/cmd/catalog/find/kubectl"
    "github.com/spf13/cobra"
)

func New() *cobra.Command {
    command := &cobra.Command{
        Use:   "find",
        Short: "finds new releases of a sub command",
    }
    command.AddCommand(kubectl.New())
    return command
}
