package main

import (
	"context"
	"fmt"
	_ "github.com/chirino/uc/internal/cmd/kamel"
	_ "github.com/chirino/uc/internal/cmd/kubectl"
	"github.com/chirino/uc/internal/cmd/uc"
	_ "github.com/chirino/uc/internal/cmd/updatecat"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Cancel ctx as soon as main returns

	cmd, err := uc.New(ctx)
	ExitOnError(err)

	// in case our binary gets renamed, use that in
	// our help/usage screens.
	// unless it's because it's being run by `go run ...`
	exeName := filepath.Base(os.Args[0])
	if !strings.HasPrefix(exeName, "___go_build_") {
		cmd.Use = exeName
	}
	// the first time we execute the root command, it discovers / dynamically sets up the sub commands
	err = cmd.Execute()
	ExitOnError(err)

	// the second time we execute, control gets passed to the sub commands.
	err = cmd.Execute()
	ExitOnError(err)
}

func ExitOnError(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
