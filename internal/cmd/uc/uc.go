package uc

import (
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/cmd/utils"
	"github.com/chirino/uc/internal/pkg/catalog"
	"github.com/chirino/uc/internal/pkg/user"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	cmd.Options
	subcommands []*cobra.Command
}

func DiscoverPhase(o *Options) *cobra.Command {
	o.subcommands = []*cobra.Command{}
	cmd := basicUcCommand(o)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return o.createSubcommands()
	}
	cmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		o.createSubcommands()
	})
	return cmd
}

func ExecutePhase(o *Options) *cobra.Command {
	if o.subcommands == nil {
		panic("DiscoverPhase(o) must be run before ExecutePhase(o)")
	}
	cmd := basicUcCommand(o)
	for _, command := range o.subcommands {
		cmd.AddCommand(command)
	}
	return cmd
}

func basicUcCommand(o *Options) *cobra.Command {
	if o.Printf == nil {
		o.Printf = cmd.StdErrPrintf
	}

	// in case our binary gets renamed, use that in
	// our help/usage screens.
	// unless it's because it's being run by `go run ...`
	use := filepath.Base(os.Args[0])
	if strings.HasPrefix(use, "___go_build_") {
		use = `uc`
	}

	var cmd = cobra.Command{
		// BashCompletionFunction: bashCompletionFunction,
		Use:   use,
		Short: `The Kubernetes/OpenShift uber client`,
		//        10        20        30        40        50        60        70        80
		Long: use + ` is an uber client that automatically installs keeps updated Kubernetes and 
OpenShift related command line tools at versions that are best suited to operate 
against the cluster that you are connected to.`,
		Example: `
  $ uc kubectl get pods
  $ uc oc new-project sandbox1
  $ uc kamel run examples/dns.js`,
		TraverseChildren:  true,
		DisableAutoGenTag: true,
		SilenceErrors:     true,
	}
	cmd.Flags().SetInterspersed(false)
	cmd.Flags().StringVarP(&o.Kubeconfig, "kubeconfig", "", filepath.Join(user.HomeDir(), ".kube", "config"), "path to the Kubeconfig file")
	cmd.Flags().StringVarP(&o.Master, "master", "", "", "Master url")

	return &cmd
}

func (o *Options) createSubcommands() error {
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
			o.subcommands = append(o.subcommands, subCommand)
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
				o.subcommands = append(o.subcommands, subCommand)
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
