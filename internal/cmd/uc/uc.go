package uc

import (
	"fmt"
	"github.com/chirino/uc/internal/cmd"
	"github.com/chirino/uc/internal/cmd/utils"
	"github.com/chirino/uc/internal/pkg/catalog"
	"github.com/chirino/uc/internal/pkg/user"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func New(o *cmd.Options) (*cobra.Command, error) {
	if o.InfoF == nil {
		o.InfoF = cmd.StdErrPrintf
		o.DebugF = cmd.NoopPrintf
	}
	o.CacheExpires = time.Now().Add(10000 * time.Hour)

	// in case our binary gets renamed, use that in
	// our help/usage screens.
	use := filepath.Base(os.Args[0])
	if strings.HasPrefix(use, "___") { // looks like an execution from idea.
		use = `uc`
	}
	cacheExpires := "24h"
	verbosity := "info"
	var result = &cobra.Command{
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
	result.Flags().SetInterspersed(false)
	result.Flags().StringVarP(&o.Kubeconfig, "kubeconfig", "", filepath.Join(user.HomeDir(), ".kube", "config"), "path to the Kubeconfig file")
	result.Flags().StringVarP(&o.Master, "master", "", "", "Master url")
	result.Flags().StringVarP(&cacheExpires, "cache-expires", "", "24h", "Controls when the catalog and command caches expire. One of *duration*|never|now")
	result.Flags().StringVarP(&verbosity, "verbosity", "v", "info", "Sets the verbosity level: One of none|info|debug")

	result.PersistentPreRunE = func(_ *cobra.Command, args []string) error {
		switch strings.ToLower(cacheExpires) {
		case "never":
			o.CacheExpires = time.Unix(0, 0) // way in the past.
		case "now":
			o.CacheExpires = time.Now().Add(10000 * time.Hour) // way in the future..
		default:
			duration, err := time.ParseDuration(cacheExpires)
			if err != nil {
				return fmt.Errorf("invalid flag value --cache-expires '%s': %s", cacheExpires, err)
			}
			o.CacheExpires = time.Now().Add(-duration)
		}

		switch strings.ToLower(verbosity) {
		case "none":
			o.InfoF = cmd.NoopPrintf
			o.DebugF = cmd.NoopPrintf
		case "info":
			o.InfoF = cmd.StdErrPrintf
			o.DebugF = cmd.NoopPrintf
		case "debug":
			o.InfoF = cmd.StdErrPrintf
			o.DebugF = cmd.StdErrPrintf
		default:
			return fmt.Errorf("invalid flag value --verbosity '%s'", verbosity)
		}

		return nil
	}

	err := addSubcommands(o, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func addSubcommands(o *cmd.Options, result *cobra.Command) error {

	subcommands := map[string]*cobra.Command{}
	for _, cmdFactory := range cmd.SubCommandFactories {
		subCommand, err := cmdFactory(o)
		if err != nil {
			return err
		}
		if subCommand != nil {
			subcommands[subCommand.Use] = subCommand
			result.AddCommand(subCommand)
		}
	}

	// options are not yet parsed from the CLI flags, so this basically using
	// --cache-expires never to avoid a doing a network round trip. After parsing this will
	// called against when a sub command is invoked with the right --cache-expires config
	catalog, err := catalog.LoadCatalogConfig(o)
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
				result.AddCommand(subCommand)
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
