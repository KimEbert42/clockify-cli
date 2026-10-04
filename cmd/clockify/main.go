package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/kim/clockify-cli/internal/cli"
	clierrors "github.com/kim/clockify-cli/internal/errors"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetBuildInfo(version, commit, date)
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		var cliErr *clierrors.CLIError
		if errors.As(err, &cliErr) {
			os.Exit(cliErr.ExitCode())
		}
		os.Exit(1)
	}
}
