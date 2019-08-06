package uc

import (
	"context"
	"fmt"
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/cmd/utils"
	"github.com/chirino/uc/internal/pkg/catalog"
	"github.com/chirino/uc/internal/pkg/user"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

type Options struct {
	cmd.Options
}

func New(ctx context.Context) (*cobra.Command, error) {

	o := Options{}
	o.Context = ctx
	o.Printf = cmd.StdErrPrintf

	var cmd = cobra.Command{
		// BashCompletionFunction: bashCompletionFunction,
		Use:               `uc`,
		Short:             `uber client`,
		Long:              `uber client runs sub commands using clients at the version that are compatible with the cluster your logged into.`,
		PersistentPreRunE: o.Run,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.CmdsAdded {
				return fmt.Errorf("invalid usage")
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			o.CmdsAdded = true
		},
	}

	cmd.Flags().StringVarP(&o.Kubeconfig, "Kubeconfig", "", filepath.Join(user.HomeDir(), ".kube", "config"), "path to the Kubeconfig file")
	cmd.Flags().StringVarP(&o.Master, "Master", "", "", "Master url")
	cmd.DisableAutoGenTag = true

	return &cmd, nil
}

func (o *Options) Run(parent *cobra.Command, args []string) error {
	if !o.CmdsAdded {

		var api *kubernetes.Clientset = nil
		config, err := clientcmd.BuildConfigFromFlags(o.Master, o.Kubeconfig)
		if err == nil {
			api, err = kubernetes.NewForConfig(config)
			if err != nil {
				return err
			}
		}

		subcommands := map[string]*cobra.Command{}
		for _, cmdFactory := range cmd.SubCmdFactories {
			subCommand, err := cmdFactory(&o.Options, api)
			if err != nil {
				return err
			}
			if subCommand != nil {
				subcommands[subCommand.Use] = subCommand
				parent.AddCommand(subCommand)
			}
		}

		// Catalog sig might be invalid when it's being updated manually, and
		// we need to run the update-catalog command..
		catalog, err := catalog.LoadCatalogConfig()
		if err == nil {
			for command, c := range catalog.Commands {
				subCommand := subcommands[command]
				if subCommand == nil {

					// We dont require a CmdFactory to be created for every command we support.  Only needed if
					// we need more customization than we can do using the data contained in the catalog data.
					// here we setup a command that only exists in the catalog
					subCommand, err = utils.GetCobraCommand(&o.Options, command, "latest")
					if err != nil {
						return err
					}
					subcommands[command] = subCommand
					parent.AddCommand(subCommand)
				}
				if c.Short != "" {
					subCommand.Short = c.Short
				}
				if c.Long != "" {
					subCommand.Long = c.Long
				}
				subCommand.Use = command
				subCommand.DisableFlagParsing = true
				subCommand.DisableAutoGenTag = true
			}
		}

	}
	return nil
}
