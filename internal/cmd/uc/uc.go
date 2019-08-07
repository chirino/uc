package uc

import (
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/cmd/utils"
	"github.com/chirino/uc/internal/pkg/catalog"
	"github.com/chirino/uc/internal/pkg/user"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func New(o *cmd.Options) *cobra.Command {
	if o.Printf == nil {
		o.Printf = cmd.StdErrPrintf
	}

	// in case our binary gets renamed, use that in
	// our help/usage screens.
	use := filepath.Base(os.Args[0])
	if strings.HasPrefix(use, "___") { // looks like an execution from idea.
		use = `uc`
	}

	var cmd = &cobra.Command{
		// BashCompletionFunction: bashCompletionFunction,
		Use:   use,
		Short: `The Kubernetes/OpenShift uber client`,
		//        10        20        30        40        50        60        70        80
		Long: use + ` is an uber client that automatically installs keeps updated Kubernetes and 
OpenShift related command line tools at versions that are best suited to operate 
against the cluster that you are connected to.`,
		Example: `
  ` + use + ` kubectl get pods
  ` + use + ` oc new-project sandbox1
  ` + use + ` kamel run examples/dns.js`,
		TraverseChildren:  true,
		DisableAutoGenTag: true,
		SilenceErrors:     true,
	}
	cmd.Flags().SetInterspersed(false)
	cmd.Flags().StringVarP(&o.Kubeconfig, "kubeconfig", "", filepath.Join(user.HomeDir(), ".kube", "config"), "path to the Kubeconfig file")
	cmd.Flags().StringVarP(&o.Master, "master", "", "", "Master url")

	addSubcommands(o, cmd)
	return cmd
}

func addSubcommands(o *cmd.Options, parent *cobra.Command) error {

	subcommands := map[string]*cobra.Command{}
	for _, cmdFactory := range cmd.SubCommandFactories {
		subCommand, err := cmdFactory(o)
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
				subCommand, err = utils.GetCobraCommand(o, command, "latest")
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
	return nil
}
