// This package implements a sub command plugin for oc.
package oc

import (
    "github.com/chirino/uc/internal/cmd"
    "github.com/spf13/cobra"
    "k8s.io/apimachinery/pkg/version"
    "k8s.io/client-go/kubernetes"
)

func init() {
    cmd.SubCmdFactories = append(cmd.SubCmdFactories, NewCmd)
}

func NewCmd(options *cmd.Options, api *kubernetes.Clientset, info *version.Info) (*cobra.Command, error) {

    // Todo: figure out best way select the best client version for the server we are connected to.
    serverToClientVersionMap := map[string]string{
        "3.11.0": "3.11.0",
        "1.12+":  "3.11.0",
    }
    serverVersion := info.Major + "." + info.Minor
    clientVersion := serverToClientVersionMap[serverVersion]
    if clientVersion == "" {
        clientVersion = "3.11.0"
    }

    return cmd.GetCobraCommand(options, "oc", clientVersion)
}

