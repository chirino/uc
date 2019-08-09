// +build mage

package main

import (
	"fmt"
	"github.com/chirino/hawtgo/sh"
	"github.com/chirino/uc/internal/pkg/dev"
	"github.com/magefile/mage/mg"

	"os"
	"runtime"
)

/////////////////////////////////////////////////////////////////////////
// A little setup to make defining the build targets easier
/////////////////////////////////////////////////////////////////////////
var (
	// d is for dependencies
	d   = mg.Deps
	cli = sh.New().
		CommandLog(os.Stdout).
		CommandLogPrefix("running > ").
		Dir(dev.GO_MOD_DIRECTORY)
)

/////////////////////////////////////////////////////////////////////////
// Build Targets:
/////////////////////////////////////////////////////////////////////////
var Default = All

func All() {
	d(Build, Test, Format)
}

type Platform struct {
	GOOS   string
	GOARCH string
}

func Build() {
	platforms := []Platform{
		Platform{"linux", "amd64"},
		Platform{"linux", "arm64"},
		Platform{"windows", "amd64"},
		Platform{"darwin", "amd64"},
	}
	for _, p := range platforms {
		cli.
			Env(map[string]string{
				"GOOS":   p.GOOS,
				"GOARCH": p.GOARCH,
			}).
			Line(fmt.Sprintf(`go build -o dist/%s-%s/uc%s`, p.GOOS, p.GOARCH, exeSuffix(p.GOOS))).
			MustZeroExit()
	}
	cli.
		Line(fmt.Sprintf(`go build --tags dev -o build/uc-dev%s`, exeSuffix(runtime.GOOS))).
		MustZeroExit()
}

func Test() {
	cli.Line(`go test ./... `).MustZeroExit()
}

func Format() {
	cli.Line(`go fmt ./... `).MustZeroExit()
}

type Catalog mg.Namespace

func (Catalog) Sign() {
	d(Build)
	cli.
		Line(fmt.Sprintf(`./build/uc-dev%s catalog sign`, exeSuffix(runtime.GOOS))).
		MustZeroExit()
}

func exeSuffix(s string) string {
	if s == "windows" {
		return ".exe"
	}
	return ""
}
