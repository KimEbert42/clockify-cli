package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAliasesInheritFlags(t *testing.T) {
	root := newRootCmd()
	alias := findCommand(root, []string{"in"})
	if alias == nil {
		t.Fatal("alias in not registered")
	}
	if alias.Flags().Lookup("project") == nil {
		t.Error("alias in should inherit --project flag from time-entry start")
	}
	if alias.Flags().Lookup("task") == nil {
		t.Error("alias in should inherit --task flag")
	}
	if alias.Flags().Lookup("tags") == nil {
		t.Error("alias in should inherit --tags flag")
	}

	outAlias := findCommand(root, []string{"out"})
	if outAlias == nil {
		t.Fatal("alias out not registered")
	}
	// stop may not have many local flags; at least command exists with RunE
	if outAlias.RunE == nil && outAlias.Run == nil {
		t.Error("alias out should have RunE")
	}

	ce := findCommand(root, []string{"ce"})
	if ce == nil {
		t.Fatal("alias ce not registered")
	}
	if ce.Flags().Lookup("start") == nil && ce.Flags().Lookup("project") == nil {
		// create has --start/--project — one of them must exist
		t.Error("alias ce should inherit create flags")
	}
}

func TestAliasExecuteParsesTargetFlags(t *testing.T) {
	// Parse-level regression: README `clockify in --project X` must not fail at flag parse.
	root := newRootCmd()
	root.SetArgs([]string{"in", "--help"})
	// help should not error even without network
	if err := root.Execute(); err != nil {
		// cobra help path may return nil; if it errors on unknown flag that's a fail
		if stringsContains(err.Error(), "unknown flag") {
			t.Fatalf("alias flags not inherited: %v", err)
		}
	}
}

func stringsContains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestFindCommandWalksSubcommands(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	child := &cobra.Command{Use: "child"}
	sub := &cobra.Command{Use: "sub"}
	child.AddCommand(sub)
	root.AddCommand(child)
	if findCommand(root, []string{"child", "sub"}) != sub {
		t.Error("findCommand should walk nested commands")
	}
}
