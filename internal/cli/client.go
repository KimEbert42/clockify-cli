package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Manage clients",
		Long:  `List, get, create, update, and delete Clockify clients.`,
	}
	cmd.AddCommand(newClientListCmd())
	cmd.AddCommand(newClientGetCmd())
	cmd.AddCommand(newClientCreateCmd())
	cmd.AddCommand(newClientUpdateCmd())
	cmd.AddCommand(newClientDeleteCmd())
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

			unfiltered := searchFilter == ""
			wsID := workspaceIDFromCmd(cmd)
			if unfiltered && wsID != "" {
				var cached []clockify.Client
				if cachedList(cmd, wsID, compResourceClients, &cached) && len(cached) > 0 {
					return formatOutput(cmd, cached, outputFormat, []string{"ID", "NAME", "EMAIL"})
				}
			}

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}
			wsID = cfg.WorkspaceID

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			clients, pageInfo, err := fetchPaginatedClients(context.Background(), client, api.ListClientsOpts{
				Search: searchFilter,
			}, page, pageSize, noPagination)
			if err != nil {
				return err
			}

			if unfiltered {
				rememberCompletionNames(cmd, wsID, compResourceClients, namesFromClients(clients))
				rememberList(cmd, wsID, compResourceClients, clients)
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
	_ = cmd.RegisterFlagCompletionFunc("id", completeClientRefC)
	return cmd
}

func newClientCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a client",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			name, _ := cmd.Flags().GetString("name")
			email, _ := cmd.Flags().GetString("email")
			address, _ := cmd.Flags().GetString("address")
			note, _ := cmd.Flags().GetString("note")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")
			ifNotExists, _ := cmd.Flags().GetBool("if-not-exists")

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			ctx := context.Background()

			// Handle JSON input from file or stdin
			if filePath != "" || stdin {
				var inputData []byte
				if filePath != "" {
					inputData, err = os.ReadFile(filePath)
					if err != nil {
						return fmt.Errorf("failed to read file %s: %w", filePath, err)
					}
				} else {
					inputData, err = io.ReadAll(os.Stdin)
					if err != nil {
						return fmt.Errorf("failed to read from stdin: %w", err)
					}
				}

				var req clockify.CreateClientRequest
				if err := json.Unmarshal(inputData, &req); err != nil {
					return fmt.Errorf("invalid JSON input: %w", err)
				}
				if req.Name == "" {
					return fmt.Errorf("client name is required")
				}
				name = req.Name
				email = req.Email
				address = req.Address
				note = req.Note
			}

			if name == "" {
				return fmt.Errorf("must specify --name")
			}

			// Handle if-not-exists
			if ifNotExists {
				existing, err := resolveClientByName(ctx, client, name)
				if err == nil {
					cmd.Printf("Client %s already exists (ID: %s).\n", existing.Name, existing.ID)
					return formatOutput(cmd, existing, outputFormat, []string{"ID", "NAME", "EMAIL", "ADDRESS", "PHONE", "NOTE"})
				}
			}

			req := clockify.CreateClientRequest{
				Name:    name,
				Email:   email,
				Address: address,
				Note:    note,
			}

			c, err := client.CreateClient(ctx, req)
			if err != nil {
				return err
			}

			return formatOutput(cmd, c, outputFormat, []string{"ID", "NAME", "EMAIL", "ADDRESS", "PHONE", "NOTE"})
		},
	}
	cmd.Flags().String("name", "", "client name (required)")
	cmd.Flags().String("email", "", "client email")
	cmd.Flags().String("address", "", "client address")
	cmd.Flags().String("note", "", "client note")
	cmd.Flags().String("file", "", "JSON file with client data")
	cmd.Flags().Bool("stdin", false, "read client data from stdin")
	cmd.Flags().Bool("if-not-exists", false, "return existing client if one with the same name exists")
	return cmd
}

func newClientUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a client",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")
			newName, _ := cmd.Flags().GetString("new-name")
			email, _ := cmd.Flags().GetString("email")
			address, _ := cmd.Flags().GetString("address")
			note, _ := cmd.Flags().GetString("note")
			archive, _ := cmd.Flags().GetBool("archive")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			if id == "" && name == "" {
				return fmt.Errorf("must specify --id or --name")
			}

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			ctx := context.Background()

			// Resolve client ID
			if id == "" {
				c, err := resolveClientByName(ctx, client, name)
				if err != nil {
					return err
				}
				id = c.ID
			}

			// Get existing client to merge updates
			existing, err := client.GetClient(ctx, id)
			if err != nil {
				return err
			}

			req := clockify.UpdateClientRequest{
				Name:    existing.Name,
				Email:   existing.Email,
				Address: existing.Address,
				Note:    existing.Note,
			}

			// Handle JSON input from file or stdin
			if filePath != "" || stdin {
				var inputData []byte
				if filePath != "" {
					inputData, err = os.ReadFile(filePath)
					if err != nil {
						return fmt.Errorf("failed to read file %s: %w", filePath, err)
					}
				} else {
					inputData, err = io.ReadAll(os.Stdin)
					if err != nil {
						return fmt.Errorf("failed to read from stdin: %w", err)
					}
				}

				var jsonReq clockify.UpdateClientRequest
				if err := json.Unmarshal(inputData, &jsonReq); err != nil {
					return fmt.Errorf("invalid JSON input: %w", err)
				}
				// Merge JSON input with existing
				if jsonReq.Name != "" {
					req.Name = jsonReq.Name
				}
				if jsonReq.Email != "" {
					req.Email = jsonReq.Email
				}
				if jsonReq.Address != "" {
					req.Address = jsonReq.Address
				}
				if jsonReq.Note != "" {
					req.Note = jsonReq.Note
				}
				if jsonReq.Archived != nil {
					req.Archived = jsonReq.Archived
				}
			} else {
				// Apply flag updates
				if newName != "" {
					req.Name = newName
				}
				if email != "" {
					req.Email = email
				}
				if address != "" {
					req.Address = address
				}
				if note != "" {
					req.Note = note
				}
				if archive {
					req.Archived = &archive
				}
			}

			c, err := client.UpdateClient(ctx, id, req)
			if err != nil {
				return err
			}

			return formatOutput(cmd, c, outputFormat, []string{"ID", "NAME", "EMAIL", "ADDRESS", "PHONE", "NOTE"})
		},
	}
	cmd.Flags().String("id", "", "client ID")
	cmd.Flags().String("name", "", "client name (fuzzy match)")
	cmd.Flags().String("new-name", "", "new client name")
	cmd.Flags().String("email", "", "client email")
	cmd.Flags().String("address", "", "client address")
	cmd.Flags().String("note", "", "client note")
	cmd.Flags().Bool("archive", false, "archive the client")
	cmd.Flags().String("file", "", "JSON file with client data")
	cmd.Flags().Bool("stdin", false, "read client data from stdin")
	_ = cmd.RegisterFlagCompletionFunc("id", completeClientRefC)
	_ = cmd.RegisterFlagCompletionFunc("name", completeClientRefC)
	_ = cmd.RegisterFlagCompletionFunc("new-name", completeClientRefC)
	return cmd
}

func newClientDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a client",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")
			confirm, _ := cmd.Flags().GetBool("confirm")

			if id == "" && name == "" {
				return fmt.Errorf("must specify --id or --name")
			}

			apiKey, region, cfg, err := loadAuthWithConfig(profile)
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace configured; run 'clockify configure' or set --workspace")
			}

			client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
			ctx := context.Background()

			// Resolve client ID if needed
			if id == "" {
				c, err := resolveClientByName(ctx, client, name)
				if err != nil {
					return err
				}
				id = c.ID
				name = c.Name
			} else {
				// Get client info for confirmation message
				c, err := client.GetClient(ctx, id)
				if err != nil {
					return err
				}
				name = c.Name
			}

			if !confirm {
				cmd.Printf("Delete client %s (%s)?\n", id, name)
				cmd.Print("Are you sure? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					cmd.Println("Aborted.")
					return nil
				}
			}

			if err := client.DeleteClient(ctx, id); err != nil {
				return err
			}

			cmd.Printf("Client %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "client ID")
	cmd.Flags().String("name", "", "client name (fuzzy match)")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	_ = cmd.RegisterFlagCompletionFunc("id", completeClientRefC)
	_ = cmd.RegisterFlagCompletionFunc("name", completeClientRefC)
	return cmd
}

func resolveClientByName(ctx context.Context, client *api.ClockifyClient, name string) (*clockify.Client, error) {
	clients, err := client.ListClients(ctx, api.ListClientsOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}

	for _, c := range clients {
		if c.Name == name {
			return &c, nil
		}
	}

	var fuzzyMatches []clockify.Client
	for _, c := range clients {
		if containsIgnoreCase(c.Name, name) {
			fuzzyMatches = append(fuzzyMatches, c)
		}
	}

	if len(fuzzyMatches) == 1 {
		return &fuzzyMatches[0], nil
	}
	if len(fuzzyMatches) > 1 {
		return nil, fmt.Errorf("multiple clients match %q; use --id to disambiguate", name)
	}

	return nil, fmt.Errorf("no client found matching %q", name)
}
