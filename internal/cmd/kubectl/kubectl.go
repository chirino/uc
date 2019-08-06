// This package implements a sub command plugin for oc.
package kubectl

import (
	"github.com/chirino/uc/internal/cmd"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"strings"
)

func init() {
	cmd.SubCmdFactories = append(cmd.SubCmdFactories, NewCmd)
}

func NewCmd(options *cmd.Options, api *kubernetes.Clientset) (*cobra.Command, error) {
	clientVersion := "latest"
	if api != nil {
		info, err := api.ServerVersion()
		if err == nil {
			clientVersion = strings.Split(strings.TrimPrefix(info.GitVersion, "v"), "-")[0]
		}
	}
	return cmd.GetCobraCommand(options, "kubectl", clientVersion)
}
