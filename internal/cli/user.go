package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/internal/config"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
		Long:  `Get current user, list workspace users, and search users.`,
	}
	cmd.AddCommand(newUserGetCmd())
	cmd.AddCommand(newUserListCmd())
	cmd.AddCommand(newUserInfoCmd())
	return cmd
}

func newUserGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get current user profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")

			apiKey, region, err := loadAuth(profile)
			if err != nil {
				return err
			}

			client := api.NewClient(apiKey, api.BaseURL(region), "")
			user, err := client.GetCurrentUser(context.Background())
			if err != nil {
				return err
			}

			return formatOutput(cmd, user, outputFormat, []string{"ID", "NAME", "EMAIL", "STATUS"})
		},
	}
}

func newUserListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users in current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			statusFilter, _ := cmd.Flags().GetString("status")

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			users, err := client.ListUsers(context.Background(), api.ListUsersOpts{
				Status: statusFilter,
			})
			if err != nil {
				return err
			}

			if len(users) == 0 && !quiet {
				cmd.Println("No users found.")
				return nil
			}

			return formatOutput(cmd, users, outputFormat, []string{"ID", "NAME", "EMAIL", "STATUS"})
		},
	}
	cmd.Flags().String("status", "", "filter by status: ACTIVE, INACTIVE")
	return cmd
}

func newUserInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Search users by email or name",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			email, _ := cmd.Flags().GetString("email")
			name, _ := cmd.Flags().GetString("name")

			if email == "" && name == "" {
				return fmt.Errorf("must specify --email or --name")
			}

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			users, err := client.ListUsers(context.Background(), api.ListUsersOpts{})
			if err != nil {
				return err
			}

			var matches []userWithMatch
			for _, u := range users {
				if email != "" && strings.EqualFold(u.Email, email) {
					matches = append(matches, userWithMatch{User: u, MatchType: "email"})
				} else if name != "" && strings.Contains(strings.ToLower(u.Name), strings.ToLower(name)) {
					matches = append(matches, userWithMatch{User: u, MatchType: "name"})
				}
			}

			if len(matches) == 0 && !quiet {
				cmd.Println("No users found matching criteria.")
				return nil
			}

			if len(matches) == 1 {
				return formatOutput(cmd, matches[0].User, outputFormat, []string{"ID", "NAME", "EMAIL", "STATUS"})
			}

			return formatOutput(cmd, matches, outputFormat, []string{"ID", "NAME", "EMAIL", "STATUS", "MATCH"})
		},
	}
	cmd.Flags().String("email", "", "search by email")
	cmd.Flags().String("name", "", "search by name")
	return cmd
}

type userWithMatch struct {
	clockify.User
	MatchType string `json:"match" yaml:"match"`
}

func loadAuthWithConfig(profile string) (apiKey, region string, cfg *config.Config, err error) {
	apiKey, err = config.GetAPIKey(profile)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get API key: %w", err)
	}

	cfg, err = config.Load(profile)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to load config: %w", err)
	}

	return apiKey, cfg.Region, cfg, nil
}
