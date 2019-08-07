package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/chirino/uc/internal/cmd"
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

	cmd, err := uc.New(&cmd.Options{
		Context: ctx,
	})
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	err = cmd.Execute()

	switch err {
	case flag.ErrHelp:
		fallthrough
	case pflag.ErrHelp:
		cmd.Help()
		os.Exit(0)
	default:
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	}
}
