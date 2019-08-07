package main

import (
	"context"
	"flag"
	"fmt"
	_ "github.com/chirino/uc/internal/cmd/kamel"
	_ "github.com/chirino/uc/internal/cmd/kubectl"
	"github.com/chirino/uc/internal/cmd/uc"
	_ "github.com/chirino/uc/internal/cmd/updatecat"
	"github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"math/rand"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Cancel ctx as soon as main returns

	o := &uc.Options{}
	o.Context = ctx

	// This first phase just uses the cli flags to connect to the cluster
	// so we can figure out which commands can be used against the cluster.
	cmd := uc.DiscoverPhase(o)
	err := cmd.Execute()
	ExitOnError(err)

	// This second phase now has a fully configured command with sub commands.
	cmd = uc.ExecutePhase(o)
	err = cmd.Execute()
	switch err {
	case flag.ErrHelp:
		fallthrough
	case pflag.ErrHelp:
		cmd.Help()
	default:
		ExitOnError(err)
	}
}

func ExitOnError(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
