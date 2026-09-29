package cli

import (
	"context"
	"fmt"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/spf13/cobra"
)

func newClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Manage clients",
		Long:  `List and get Clockify clients.`,
	}
	cmd.AddCommand(newClientListCmd())
	cmd.AddCommand(newClientGetCmd())
	return cmd
}

func newClientListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List clients in current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			searchFilter, _ := cmd.Flags().GetString("search")
			page, _ := cmd.Flags().GetInt("page")
			pageSize, _ := cmd.Flags().GetInt("page-size")
			noPagination, _ := cmd.Flags().GetBool("no-pagination")

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			clients, pageInfo, err := fetchPaginatedClients(context.Background(), client, api.ListClientsOpts{
				Search: searchFilter,
			}, page, pageSize, noPagination)
			if err != nil {
				return err
			}

			if len(clients) == 0 && !quiet {
				cmd.Println("No clients found.")
				return nil
			}

			if err := formatOutput(cmd, clients, outputFormat, []string{"ID", "NAME", "EMAIL"}); err != nil {
				return err
			}

			if !quiet && outputFormat != "json" && outputFormat != "csv" && outputFormat != "yaml" && pageInfo != nil && pageInfo.TotalPages > 1 {
				cmd.Printf("\nPage %d of %d (%d results)\n", pageInfo.Page, pageInfo.TotalPages, pageInfo.TotalResults)
			}
			return nil
		},
	}
	cmd.Flags().String("search", "", "search by client name")
	return cmd
}

func newClientGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get client by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")

			if id == "" {
				return fmt.Errorf("must specify --id")
			}

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			c, err := client.GetClient(context.Background(), id)
			if err != nil {
				return err
			}

			return formatOutput(cmd, c, outputFormat, []string{"ID", "NAME", "EMAIL", "ADDRESS", "PHONE", "NOTE"})
		},
	}
	cmd.Flags().String("id", "", "client ID")
	return cmd
}
