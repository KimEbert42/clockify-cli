package cli

import (
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long:  `Generate shell completion scripts for bash, zsh, fish, or powershell.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCompletion,
}

func runCompletion(cmd *cobra.Command, args []string) error {
	root := cmd.Root()
	switch args[0] {
	case "bash":
		return root.GenBashCompletionV2(cmd.OutOrStdout(), true)
	case "zsh":
		return root.GenZshCompletion(cmd.OutOrStdout())
	case "fish":
		return root.GenFishCompletion(cmd.OutOrStdout(), true)
	case "powershell":
		return root.GenPowerShellCompletion(cmd.OutOrStdout())
	default:
		return cmd.Help()
	}
}
