package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

type Options struct {
	Context        context.Context
	Kubeconfig     string
	Master         string
	CacheExpires   time.Time
	InfoLog        io.Writer
	DebugLog       io.Writer
	CatalogIndex   *CatalogIndex
	CommandVersion string
}

type CatalogIndex struct {
	Commands map[string]*CatalogCommand `json:"commands,omitempty"`
}

type CatalogCommand struct {
	Short            string `json:"short-description,omitempty"`
	Long             string `json:"long-description,omitempty"`
	LatestVersion    string `json:"latest-version,omitempty"`
	CatalogBaseURL   string `json:"catalog-base-url,omitempty"`
	CatalogPublicKey string `json:"catalog-public-key,omitempty"`
}

type SubCommandFactory func(options *Options) *cobra.Command

var SubCommandFactories = []SubCommandFactory{}

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
