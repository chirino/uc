package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type Options struct {
	Context    context.Context
	Kubeconfig string
	Master     string
	Printf     func(format string, a ...interface{})
}

type SubCommandFactory func(options *Options) (*cobra.Command, error)

var SubCommandFactories = []SubCommandFactory{}

func StdErrPrintf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func (o *Options) LoadBuildConfig() (*restclient.Config, error) {
	return clientcmd.BuildConfigFromFlags(o.Master, o.Kubeconfig)
}

func (o *Options) NewApiClient() (*kubernetes.Clientset, error) {
	config, err := o.LoadBuildConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
