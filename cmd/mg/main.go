package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/taigrr/mg/cmd/mg/cmd"
)

// version and commit are overridable via -ldflags at build time (GoReleaser).
// When empty, fang falls back to the module version embedded by the toolchain
// via runtime/debug.ReadBuildInfo (populated for `go install ...@version`).
var (
	version = ""
	commit  = ""
)

func main() {
	opts := []fang.Option{}
	if version != "" {
		opts = append(opts, fang.WithVersion(version))
	}
	if commit != "" {
		opts = append(opts, fang.WithCommit(commit))
	}
	if err := fang.Execute(context.Background(), cmd.RootCmd, opts...); err != nil {
		os.Exit(1)
	}
}
