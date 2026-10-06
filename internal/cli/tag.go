package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	clierrors "github.com/kim/clockify-cli/internal/errors"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
		Long:  `List, get, create, update, and delete Clockify tags.`,
	}
	cmd.AddCommand(newTagListCmd())
	cmd.AddCommand(newTagCreateCmd())
	cmd.AddCommand(newTagUpdateCmd())
	cmd.AddCommand(newTagDeleteCmd())
	return cmd
}

func newTagListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tags in current workspace",
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
				var cached []clockify.Tag
				if cachedList(cmd, wsID, compResourceTags, &cached) && len(cached) > 0 {
					return formatOutput(cmd, cached, outputFormat, []string{"ID", "NAME", "COLOR"})
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
			tags, pageInfo, err := fetchPaginatedTags(context.Background(), client, api.ListTagsOpts{
				Search: searchFilter,
			}, page, pageSize, noPagination)
			if err != nil {
				return err
			}

			if unfiltered {
				rememberCompletionNames(cmd, wsID, compResourceTags, namesFromTags(tags))
				rememberList(cmd, wsID, compResourceTags, tags)
			}

			if len(tags) == 0 && !quiet {
				cmd.Println("No tags found.")
				return nil
			}

			if err := formatOutput(cmd, tags, outputFormat, []string{"ID", "NAME", "COLOR"}); err != nil {
				return err
			}

			if !quiet && outputFormat != "json" && outputFormat != "csv" && outputFormat != "yaml" && pageInfo != nil && pageInfo.TotalPages > 1 {
				cmd.Printf("\nPage %d of %d (%d results)\n", pageInfo.Page, pageInfo.TotalPages, pageInfo.TotalResults)
			}
			return nil
		},
	}
	cmd.Flags().String("search", "", "search by tag name")
	return cmd
}

func newTagCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			name, _ := cmd.Flags().GetString("name")
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

				var req clockify.CreateTagRequest
				if err := json.Unmarshal(inputData, &req); err != nil {
					return fmt.Errorf("invalid JSON input: %w", err)
				}
				if req.Name == "" {
					return fmt.Errorf("tag name is required")
				}
				name = req.Name
			}

			if name == "" {
				return fmt.Errorf("must specify --name")
			}

			// Handle if-not-exists — only swallow not-found.
			if ifNotExists {
				existing, err := resolveTagByName(ctx, client, name)
				if err == nil {
					cmd.Printf("Tag %s already exists (ID: %s).\n", existing.Name, existing.ID)
					return formatOutput(cmd, existing, outputFormat, []string{"ID", "NAME", "COLOR"})
				}
				if cliErr, ok := err.(*clierrors.CLIError); !ok || cliErr.Code != clierrors.E002NotFound {
					return err
				}
			}

			req := clockify.CreateTagRequest{
				Name: name,
			}

			tag, err := client.CreateTag(ctx, req)
			if err != nil {
				return err
			}

			return formatOutput(cmd, tag, outputFormat, []string{"ID", "NAME", "COLOR"})
		},
	}
	cmd.Flags().String("name", "", "tag name (required)")
	cmd.Flags().String("file", "", "JSON file with tag data")
	cmd.Flags().Bool("stdin", false, "read tag data from stdin")
	cmd.Flags().Bool("if-not-exists", false, "return existing tag if one with the same name exists")
	return cmd
}

func newTagUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")
			newName, _ := cmd.Flags().GetString("new-name")
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

			// Resolve tag ID
			if id == "" {
				t, err := resolveTagByName(ctx, client, name)
				if err != nil {
					return err
				}
				id = t.ID
			}

			// Get existing tag to merge updates
			existing, err := client.GetTag(ctx, id)
			if err != nil {
				return err
			}

			req := clockify.UpdateTagRequest{
				Name: existing.Name,
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

				var jsonReq clockify.UpdateTagRequest
				if err := json.Unmarshal(inputData, &jsonReq); err != nil {
					return fmt.Errorf("invalid JSON input: %w", err)
				}
				// Merge JSON input with existing
				if jsonReq.Name != "" {
					req.Name = jsonReq.Name
				}
				if jsonReq.Archived != nil {
					req.Archived = jsonReq.Archived
				}
			} else {
				// Apply flag updates
				if newName != "" {
					req.Name = newName
				}
				if archive {
					req.Archived = &archive
				}
			}

			tag, err := client.UpdateTag(ctx, id, req)
			if err != nil {
				return err
			}

			return formatOutput(cmd, tag, outputFormat, []string{"ID", "NAME", "COLOR"})
		},
	}
	cmd.Flags().String("id", "", "tag ID")
	cmd.Flags().String("name", "", "tag name (fuzzy match)")
	cmd.Flags().String("new-name", "", "new tag name")
	cmd.Flags().Bool("archive", false, "archive the tag")
	cmd.Flags().String("file", "", "JSON file with tag data")
	cmd.Flags().Bool("stdin", false, "read tag data from stdin")
	_ = cmd.RegisterFlagCompletionFunc("id", completeTagRefC)
	_ = cmd.RegisterFlagCompletionFunc("name", completeTagRefC)
	_ = cmd.RegisterFlagCompletionFunc("new-name", completeTagRefC)
	return cmd
}

func newTagDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a tag",
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

			// Resolve tag ID if needed
			if id == "" {
				t, err := resolveTagByName(ctx, client, name)
				if err != nil {
					return err
				}
				id = t.ID
				name = t.Name
			} else {
				// Get tag info for confirmation message
				t, err := client.GetTag(ctx, id)
				if err != nil {
					return err
				}
				name = t.Name
			}

			if !confirm {
				cmd.Printf("Delete tag %s (%s)?\n", id, name)
				cmd.Print("Are you sure? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					cmd.Println("Aborted.")
					return nil
				}
			}

			if err := client.DeleteTag(ctx, id); err != nil {
				return err
			}

			cmd.Printf("Tag %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "tag ID")
	cmd.Flags().String("name", "", "tag name (fuzzy match)")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	_ = cmd.RegisterFlagCompletionFunc("id", completeTagRefC)
	_ = cmd.RegisterFlagCompletionFunc("name", completeTagRefC)
	return cmd
}

func resolveTagByName(ctx context.Context, client *api.ClockifyClient, name string) (*clockify.Tag, error) {
	tags, err := client.ListTags(ctx, api.ListTagsOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	for _, t := range tags {
		if t.Name == name {
			return &t, nil
		}
	}

	var fuzzyMatches []clockify.Tag
	for _, t := range tags {
		if containsIgnoreCase(t.Name, name) {
			fuzzyMatches = append(fuzzyMatches, t)
		}
	}

	if len(fuzzyMatches) == 1 {
		return &fuzzyMatches[0], nil
	}
	if len(fuzzyMatches) > 1 {
		cands := make([]nameCandidate, 0, len(fuzzyMatches))
		for _, t := range fuzzyMatches {
			cands = append(cands, nameCandidate{Name: t.Name, ID: t.ID})
		}
		return nil, ambiguousRefError("tags", name, cands)
	}

	return nil, notFoundRefError("tag", name)
}
