package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"time"
)

type Options struct {
	Context      context.Context
	Kubeconfig   string
	Master       string
	CacheExpires time.Time
	InfoF        func(format string, a ...interface{})
	DebugF       func(format string, a ...interface{})
}

type SubCommandFactory func(options *Options) (*cobra.Command, error)

var SubCommandFactories = []SubCommandFactory{}

func StdErrPrintf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func NoopPrintf(format string, a ...interface{}) {
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
