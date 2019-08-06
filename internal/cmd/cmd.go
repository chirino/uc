package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"os"
)

type Options struct {
	Context    context.Context
	Kubeconfig string
	Master     string
	CmdsAdded  bool
	Printf     func(format string, a ...interface{})
}

type CmdFactory func(options *Options, api *kubernetes.Clientset) (*cobra.Command, error)

var SubCmdFactories = []CmdFactory{}

func StdErrPrintf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}
