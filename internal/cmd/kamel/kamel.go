// This package implements a sub command plugin for camel-k.
package kamel

import (
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/cmd/utils"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

func init() {
	cmd.SubCmdFactories = append(cmd.SubCmdFactories, NewCmd)
}

func NewCmd(options *cmd.Options, api *kubernetes.Clientset) (*cobra.Command, error) {
	// Todo: figure out how to pick the best client version for the server we are connected against.
	clientVersion := "latest"
	return utils.GetCobraCommand(options, "kamel", clientVersion)
}
