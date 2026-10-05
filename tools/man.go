// Command man generates man pages and markdown docs from the Cobra command tree.
//
// Usage:
//
//	go run ./tools/man.go            # writes man/ and docs/cli/
//	go run ./tools/man.go -man-only
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kim/clockify-cli/internal/cli"
	"github.com/spf13/cobra/doc"
)

func main() {
	manOnly := flag.Bool("man-only", false, "generate only man pages")
	mdOnly := flag.Bool("md-only", false, "generate only markdown docs")
	out := flag.String("out", ".", "output root directory")
	flag.Parse()

	root := cli.NewRootCommand()
	root.DisableAutoGenTag = true
	root.InitDefaultHelpCmd()
	root.InitDefaultHelpFlag()

	manDir := filepath.Join(*out, "man")
	mdDir := filepath.Join(*out, "docs", "cli")
	if err := os.MkdirAll(manDir, 0o755); err != nil {
		fail(err)
	}
	if err := os.MkdirAll(mdDir, 0o755); err != nil {
		fail(err)
	}

	header := &doc.GenManHeader{
		Title:   "CLOCKIFY",
		Section: "1",
		Source:  "clockify-cli",
	}

	if !*mdOnly {
		if err := doc.GenManTree(root, header, manDir); err != nil {
			fail(err)
		}
		fmt.Printf("man pages → %s\n", manDir)
	}
	if !*manOnly {
		if err := doc.GenMarkdownTree(root, mdDir); err != nil {
			fail(err)
		}
		fmt.Printf("markdown docs → %s\n", mdDir)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "man generation failed:", err)
	os.Exit(1)
}
