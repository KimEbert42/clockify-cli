// Coverage gate: fails when any core package is below the threshold.
// Usage: go run ./tools/coveragegate cover.out
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var corePackages = []string{
	"internal/config",
	"internal/output",
	"internal/errors",
	"internal/update",
	"internal/resolve",
	"internal/api",
	"pkg/duration",
	"pkg/clockify",
}

const threshold = 80.0
const modulePrefix = "github.com/kim/clockify-cli/"

func matchCore(pkgPath string) (string, bool) {
	// cover profile paths look like github.com/kim/clockify-cli/internal/api/client.go
	if !strings.HasPrefix(pkgPath, modulePrefix) {
		return "", false
	}
	rel := strings.TrimPrefix(pkgPath, modulePrefix)
	if i := strings.LastIndex(rel, "/"); i >= 0 {
		rel = rel[:i]
	}
	for _, core := range corePackages {
		if rel == core || strings.HasPrefix(rel, core+"/") {
			return core, true
		}
	}
	return "", false
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: coverage_gate <coverprofile>")
		os.Exit(2)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer f.Close()

	coveredStmts := map[string]float64{}
	totalStmts := map[string]float64{}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		// format: path/file.go:line.col,line.col numstmts count
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		loc := parts[0]
		idx := strings.LastIndex(loc, ":")
		if idx < 0 {
			continue
		}
		pkgPath := loc[:idx]
		stmts, _ := strconv.Atoi(parts[1])
		cnt, _ := strconv.Atoi(parts[2])
		if core, ok := matchCore(pkgPath); ok {
			totalStmts[core] += float64(stmts)
			if cnt > 0 {
				coveredStmts[core] += float64(stmts)
			}
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	failed := false
	fmt.Println("core package coverage gate (threshold 80%):")
	for _, core := range corePackages {
		total := totalStmts[core]
		if total == 0 {
			fmt.Printf("  %s: NO DATA — fail\n", core)
			failed = true
			continue
		}
		pct := 100.0 * coveredStmts[core] / total
		status := "ok"
		if pct < threshold {
			status = "FAIL"
			failed = true
		}
		fmt.Printf("  %s: %.1f%% (%s)\n", core, pct, status)
	}
	if failed {
		os.Exit(1)
	}
}
