package main

import (
    "context"
    "fmt"
    "github.com/chirino/uc/internal/cmd"
    _ "github.com/chirino/uc/internal/cmd/kamel"
    _ "github.com/chirino/uc/internal/cmd/kn"
    _ "github.com/chirino/uc/internal/cmd/kubectl"
    _ "github.com/chirino/uc/internal/cmd/oc"
    _ "github.com/chirino/uc/internal/cmd/odo"
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

    cmd, err := cmd.New(ctx)
    ExitOnError(err)

    exeName := filepath.Base(os.Args[0])
    if !strings.Contains(exeName, "go_build_main_go") {
        cmd.Use = exeName
    }
    // First time discovers sub commands..
    err = cmd.Execute()
    ExitOnError(err)

    // Second time the sub-commands will be setup..
    err = cmd.Execute()
    ExitOnError(err)
}

func ExitOnError(err error) {
    if err != nil {
        fmt.Println("error:", err)
        os.Exit(1)
    }
}
