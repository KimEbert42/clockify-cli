package cli

import "github.com/spf13/cobra"

var aliases = map[string]string{
	"lw":  "workspace list",
	"le":  "time-entry list",
	"lp":  "project list",
	"ge":  "time-entry get",
	"ce":  "time-entry create",
	"cfg": "config set",
	"in":  "time-entry start",
	"out": "time-entry stop",
}

func registerAliases(root *cobra.Command) {
	for alias, target := range aliases {
		parts := splitCommand(target)
		cmd := findCommand(root, parts)
		if cmd == nil {
			continue
		}
		aliasCmd := &cobra.Command{
			Use:   alias,
			Short: cmd.Short,
			RunE:  cmd.RunE,
			Run:   cmd.Run,
		}
		// Inherit flags so README examples like `clockify in --project X` parse.
		aliasCmd.Flags().AddFlagSet(cmd.Flags())
		aliasCmd.PersistentFlags().AddFlagSet(cmd.PersistentFlags())
		root.AddCommand(aliasCmd)
	}
}

func splitCommand(s string) []string {
	var parts []string
	for _, p := range splitFields(s) {
		parts = append(parts, p)
	}
	return parts
}

func splitFields(s string) []string {
	var result []string
	for _, f := range splitBySpace(s) {
		if f != "" {
			result = append(result, f)
		}
	}
	return result
}

func splitBySpace(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func findCommand(root *cobra.Command, parts []string) *cobra.Command {
	if len(parts) == 0 {
		return nil
	}
	for _, cmd := range root.Commands() {
		if cmd.Name() == parts[0] || cmd.HasAlias(parts[0]) {
			if len(parts) == 1 {
				return cmd
			}
			return findCommand(cmd, parts[1:])
		}
	}
	return nil
}
