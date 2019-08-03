// This package implements a sub command plugin for odo.
package odo

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
        "3.11.0": "1.0.0-beta3",
        "1.12+":  "1.0.0-beta3",
    }
    serverVersion := info.Major + "." + info.Minor
    clientVersion := serverToClientVersionMap[serverVersion]
    if clientVersion == "" {
        return nil, nil
    }

    return cmd.GetCobraCommand(options, "odo", clientVersion)
}

