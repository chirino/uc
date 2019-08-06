package cmd

import (
	"context"
	"fmt"
	"github.com/chirino/hawtgo/sh"
	"github.com/chirino/uc/internal/pkg/cache"
	"github.com/chirino/uc/internal/pkg/user"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"runtime"
)

type CmdFactory func(options *Options, api *kubernetes.Clientset) (*cobra.Command, error)

var SubCmdFactories = []CmdFactory{}

type CatalogConfig struct {
	Update   string                     `json:"update,omitempty"`
	Commands map[string]*CatalogCommand `json:"commands,omitempty"`
}

type CatalogCommand struct {
	Short         string `json:"short-description,omitempty"`
	Long          string `json:"long-description,omitempty"`
	LatestVersion string `json:"latest-version,omitempty"`
}

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
		Printf:  StdErrPrintf,
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
	cmd.DisableAutoGenTag = true

	return &cmd, nil
}

func (o *Options) Run(cmd *cobra.Command, args []string) error {
	if !o.cmdsAdded {

		var api *kubernetes.Clientset = nil
		config, err := clientcmd.BuildConfigFromFlags(o.master, o.kubeconfig)
		if err == nil {
			api, err = kubernetes.NewForConfig(config)
			if err != nil {
				return err
			}
		}

		subcommands := map[string]*cobra.Command{}
		for _, cmdFactory := range SubCmdFactories {
			subCommand, err := cmdFactory(o, api)
			if err != nil {
				return err
			}
			if subCommand != nil {
				subcommands[subCommand.Use] = subCommand
				cmd.AddCommand(subCommand)
			}
		}

		// Catalog sig might be invalid when it's being updated manually, and
		// we need to run the update-catalog command..
		catalog, err := LoadCatalogConfig()
		if err == nil {
			for command, c := range catalog.Commands {
				subCommand := subcommands[command]
				if subCommand == nil {

					// We dont require a CmdFactory to be created for every command we support.  Only needed if
					// we need more customization than we can do using the data contained in the catalog data.
					// here we setup a command that only exists in the catalog
					subCommand, err = GetCobraCommand(o, command, "latest")
					if err != nil {
						return err
					}
					subcommands[command] = subCommand
					cmd.AddCommand(subCommand)
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

func StdErrPrintf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func GetExecutable(options *Options, command string, version string) (string, error) {
	catalog, err := LoadCatalogConfig()
	if err != nil {
		return "", err
	}

	config := catalog.Commands[command]
	if config == nil {
		return "", fmt.Errorf("command not found in catalog: %s", command)
	}

	if version == "latest" {
		version = config.LatestVersion
	}

	platforms, err := LoadCommandPlatforms(command, version)
	if err != nil {
		return "", err
	}
	if platforms == nil {
		return "", fmt.Errorf("%s version not found in catalog: %s", command, version)
	}

	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	request := platforms[platform]
	if request == nil {
		return "", fmt.Errorf("%s %s platform not found in catalog: %s", command, version, platform)
	}
	request.CommandName = command
	request.Platform = platform
	request.Version = version
	request.Printf = options.Printf
	return cache.Get(request)
}

func GetCobraCommand(options *Options, command string, clientVersion string) (*cobra.Command, error) {
	return &cobra.Command{
		Use: command,
		RunE: func(c *cobra.Command, args []string) error {

			// Get the executable for that client version...
			executable, err := GetExecutable(options, command, clientVersion)
			if err != nil {
				return err
			}

			// call it pass along any args....
			return sh.New().LineArgs(append([]string{executable}, args...)...).Exec()
		},
	}, nil
}
