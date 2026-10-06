package main

import (
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
		// E00X format + Learn more on stderr; stdout stays pipe-clean.
		fmt.Fprintln(os.Stderr, clierrors.Format(err))
		var cliErr *clierrors.CLIError
		if clierrors.As(err, &cliErr) {
			os.Exit(cliErr.ExitCode())
		}
		os.Exit(1)
	}
}
