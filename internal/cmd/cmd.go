package cmd

import (
    "context"
    "fmt"
    "github.com/chirino/uc/internal/pkg/user"
    "github.com/spf13/cobra"
    "k8s.io/apimachinery/pkg/version"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "os"
    "path/filepath"
    "runtime"
)

type CmdFactory func(options *Options, api *kubernetes.Clientset, serverVersion *version.Info) (*cobra.Command, error)

var SubCmdFactories = []CmdFactory{}

type Options struct {
    Context    context.Context
    kubeconfig string
    master     string
    cmdsAdded  bool
    Printf     func(format string, a ...interface{})
}

func New(ctx context.Context) (*cobra.Command, error) {
    o := Options{
        Context: ctx,
        Printf: StdErrPrintf,
    }
    var cmd = cobra.Command{
        // BashCompletionFunction: bashCompletionFunction,
        Use:               `uc`,
        Short:             `uber client`,
        Long:              `uber client runs sub commands using clients at the version that are compatible with the cluster your logged into.`,
        PersistentPreRunE: o.Run,
        RunE: func(cmd *cobra.Command, args []string) error {
            if o.cmdsAdded {
                return fmt.Errorf("invalid usage")
            }
            return nil
        },
        PersistentPostRun: func(cmd *cobra.Command, args []string) {
            o.cmdsAdded = true
        },
    }

    cmd.Flags().StringVarP(&o.kubeconfig, "kubeconfig", "", filepath.Join(user.HomeDir(), ".kube", "config"), "path to the kubeconfig file")
    cmd.Flags().StringVarP(&o.master, "master", "", "", "master url")

    return &cmd, nil
}

func (o *Options) Run(cmd *cobra.Command, args []string) error {
    if !o.cmdsAdded {

        config, err := clientcmd.BuildConfigFromFlags(o.master, o.kubeconfig)
        if err != nil {
            return err
        }

        api, err := kubernetes.NewForConfig(config)
        if err != nil {
            return err
        }

        serverVersion, err := api.ServerVersion()
        if err != nil {
            return err
        }

        for _, cmdFactory := range SubCmdFactories {
            subCommand, err := cmdFactory(o, api, serverVersion)
            if err != nil {
                return err
            }
            if subCommand != nil {
                cmd.AddCommand(subCommand)
            }
        }
    }
    return nil
}

func ExeSuffix(s string) string {
    if runtime.GOOS == "windows" {
        return s + ".exe"
    }
    return s
}

func StdErrPrintf(format string, a ...interface{}) {
    fmt.Fprintf(os.Stderr, format, a...)
}
