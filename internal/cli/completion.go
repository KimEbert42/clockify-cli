package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate or install shell completion scripts",
		Long: `Generate shell completion scripts for bash, zsh, fish, or powershell.

Output a script to stdout:
  clockify completion bash

Install to the standard location for your shell:
  clockify completion install --shell zsh

Refresh cached resource names used by dynamic tab completion:
  clockify completion refresh`,
		Args: cobra.MaximumNArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Run: func(c *cobra.Command, args []string) {
			if len(args) == 1 {
				_ = runCompletionScript(c, args[0])
				return
			}
			_ = c.Help()
		},
	}

	cmd.AddCommand(newCompletionShellCmd("bash", "Generate bash completion script"))
	cmd.AddCommand(newCompletionShellCmd("zsh", "Generate zsh completion script"))
	cmd.AddCommand(newCompletionShellCmd("fish", "Generate fish completion script"))
	cmd.AddCommand(newCompletionShellCmd("powershell", "Generate PowerShell completion script"))

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install completion script to a standard shell path",
		Args:  cobra.NoArgs,
		RunE:  runCompletionInstall,
	}
	installCmd.Flags().String("shell", "", "shell to install for: bash, zsh, fish, powershell (required)")
	installCmd.Flags().Bool("force", false, "overwrite an existing completion file")
	_ = installCmd.MarkFlagRequired("shell")
	cmd.AddCommand(installCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "refresh",
		Short: "Refresh cached workspace/project/client/tag/user names for completions",
		Args:  cobra.NoArgs,
		RunE:  runCompletionRefresh,
	})

	return cmd
}

func newCompletionShellCmd(shell, short string) *cobra.Command {
	return &cobra.Command{
		Use:   shell,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompletionScript(cmd, shell)
		},
	}
}

func runCompletionScript(cmd *cobra.Command, shell string) error {
	root := cmd.Root()
	var buf bytes.Buffer
	var err error
	switch shell {
	case "bash":
		err = root.GenBashCompletionV2(&buf, true)
	case "zsh":
		err = root.GenZshCompletion(&buf)
	case "fish":
		err = root.GenFishCompletion(&buf, true)
	case "powershell":
		err = root.GenPowerShellCompletion(&buf)
	default:
		return fmt.Errorf("unsupported shell %q (use bash, zsh, fish, or powershell)", shell)
	}
	if err != nil {
		return fmt.Errorf("failed to generate %s completion: %w", shell, err)
	}
	_, err = cmd.OutOrStdout().Write(buf.Bytes())
	return err
}

func completionInstallPath(shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve home directory: %w", err)
	}
	switch shell {
	case "bash":
		if dir := os.Getenv("BASH_COMP_INSTALL_DIR"); dir != "" {
			return filepath.Join(dir, "clockify"), nil
		}
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "bash-completion", "completions", "clockify"), nil
		}
		return filepath.Join(home, ".local", "share", "bash-completion", "completions", "clockify"), nil
	case "zsh":
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "zsh", "site-functions", "_clockify"), nil
		}
		return filepath.Join(home, ".zsh", "completions", "_clockify"), nil
	case "fish":
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "fish", "completions", "clockify.fish"), nil
		}
		return filepath.Join(home, ".config", "fish", "completions", "clockify.fish"), nil
	case "powershell":
		docs := os.Getenv("USERPROFILE")
		if docs == "" {
			docs = home
		}
		return filepath.Join(docs, "Documents", "PowerShell", "clockify.ps1"), nil
	default:
		return "", fmt.Errorf("unsupported shell %q (use bash, zsh, fish, or powershell)", shell)
	}
}

func runCompletionInstall(cmd *cobra.Command, args []string) error {
	shell, _ := cmd.Flags().GetString("shell")
	force, _ := cmd.Flags().GetBool("force")
	quiet, _ := cmd.Flags().GetBool("quiet")

	path, err := completionInstallPath(shell)
	if err != nil {
		return err
	}

	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("completion file already exists at %s; use --force to overwrite", path)
		}
	}

	var buf bytes.Buffer
	root := cmd.Root()
	switch shell {
	case "bash":
		err = root.GenBashCompletionV2(&buf, true)
	case "zsh":
		err = root.GenZshCompletion(&buf)
	case "fish":
		err = root.GenFishCompletion(&buf, true)
	case "powershell":
		err = root.GenPowerShellCompletion(&buf)
	}
	if err != nil {
		return fmt.Errorf("failed to generate %s completion: %w", shell, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write completion file: %w", err)
	}

	if !quiet {
		cmd.Printf("Installed %s completions to %s\n", shell, path)
		if shell == "powershell" {
			cmd.Println("Add this line to your PowerShell profile:")
			cmd.Printf("  . %s\n", path)
		}
	}
	return nil
}

func runCompletionRefresh(cmd *cobra.Command, args []string) error {
	return refreshCompletionCache(cmd)
}
