// This package implements a sub command plugin for oc.
package oc


import (
    "fmt"
    "github.com/chirino/hawtgo/sh"
    "github.com/chirino/uc/internal/cmd"
    "github.com/chirino/uc/internal/pkg/cache"
    "github.com/spf13/cobra"
    "k8s.io/apimachinery/pkg/version"
    "k8s.io/client-go/kubernetes"
    "os"
)

var serverToClientVersionMap = map[string]string{
    "3.11.0": "3.11.0",
    "1.12+": "3.11.0",
}

//
// We could also download this metadata... but for now lets just embed it..
//
var clientRequests = map[string]*cache.Request{
    "3.11.0": &cache.Request{
        URL:       `https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-mac.zip`,
        Signature: `iF0EABECAB0WIQTluCR6+KYZoo+Q/fyf8lmA9bp+TwUCXUSrYAAKCRCf8lmA9bp+TyRhAJ9LwDvN8paD4+OTxmCmGfQZ2f/pDgCgvj3efSc/+0KdnQr7DgtxY0u0gFY=`,
        Size:      55858072,
    },
}


func init() {
    cmd.SubCmdFactories = append(cmd.SubCmdFactories, NewCmd)
}

func NewCmd(options *cmd.Options, api *kubernetes.Clientset, serverVersion *version.Info) (command *cobra.Command, e error) {

    // Figure out what client version we should download based on the k8s server we are connected to...
    clientVersion := toClientVersion(serverVersion)

    // disable the sub command if we can't figure out what version to use
    if clientVersion == "" {
        return nil, nil
    }

    return &cobra.Command{
        Use: `oc`,
        RunE: func(cmd *cobra.Command, args []string) error {

            // Get the executable for that client version...
            executable, err := toExe(options, clientVersion)
            if err != nil {
                return err
            }

            // call it pass along any args....
            rc, err := sh.New().LineArgs(append([]string{executable}, args...)...).Exec()
            os.Exit(rc)
            return nil
        },
    }, nil
}

func toClientVersion(info *version.Info) string {
    // TODO: figure out..
    serverVersion := info.Major + "." + info.Minor
    clientVersion := serverToClientVersionMap[serverVersion]
    return clientVersion
}


func toExe(options *cmd.Options, clientVersion string) (string, error) {
    request := clientRequests[clientVersion]
    if request == nil {
        return "", fmt.Errorf("unknonwn version: %s", clientVersion)
    }
    request.PathInZip = "oc"
    request.CommandName = "oc"
    request.Version = clientVersion
    request.Printf = options.Printf
    return cache.GetCommandFromCache(request)
}


