package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	clierrors "github.com/kim/clockify-cli/internal/errors"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage user groups",
		Long:  `List, get, create, update, and delete Clockify user groups (memberships).`,
	}
	cmd.AddCommand(newGroupListCmd())
	cmd.AddCommand(newGroupGetCmd())
	cmd.AddCommand(newGroupCreateCmd())
	cmd.AddCommand(newGroupUpdateCmd())
	cmd.AddCommand(newGroupDeleteCmd())
	return cmd
}

func newGroupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			search, _ := cmd.Flags().GetString("search")
			page, _ := cmd.Flags().GetInt("page")
			pageSize, _ := cmd.Flags().GetInt("page-size")
			noPagination, _ := cmd.Flags().GetBool("no-pagination")

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			groups, pageInfo, err := fetchPaginatedGroups(context.Background(), client, api.ListUserGroupsOpts{Name: search}, page, pageSize, noPagination)
			if err != nil {
				return err
			}
			if len(groups) == 0 && !quiet {
				cmd.Println("No groups found.")
				return nil
			}
			type groupDisplay struct {
				ID          string `json:"id" yaml:"id"`
				Name        string `json:"name" yaml:"name"`
				MemberCount int    `json:"memberCount" yaml:"memberCount"`
				Members     string `json:"members" yaml:"members"`
			}
			displays := make([]groupDisplay, 0, len(groups))
			for _, g := range groups {
				displays = append(displays, groupDisplay{
					ID:          g.ID,
					Name:        g.Name,
					MemberCount: len(g.UserIDs),
					Members:     strings.Join(g.UserIDs, ","),
				})
			}
			if err := formatOutput(cmd, displays, outputFormat, []string{"ID", "NAME", "MEMBER_COUNT", "MEMBERS"}); err != nil {
				return err
			}
			printPaginationSummary(cmd, quiet, outputFormat, pageInfo)
			return nil
		},
	}
	cmd.Flags().String("search", "", "filter by group name")
	return cmd
}

func newGroupGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a user group",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")

			if id == "" && name == "" {
				return fmt.Errorf("must specify --id or --name")
			}
			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()
			if id == "" {
				g, err := resolveGroupByName(ctx, client, name)
				if err != nil {
					return err
				}
				id = g.ID
			}
			g, err := client.GetUserGroup(ctx, id)
			if err != nil {
				return err
			}
			return formatOutput(cmd, g, outputFormat, []string{"ID", "NAME", "USER_IDS"})
		},
	}
	cmd.Flags().String("id", "", "group ID")
	cmd.Flags().String("name", "", "group name (fuzzy match)")
	return cmd
}

func newGroupCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user group",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			name, _ := cmd.Flags().GetString("name")
			members, _ := cmd.Flags().GetString("members")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")
			ifNotExists, _ := cmd.Flags().GetBool("if-not-exists")

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			if filePath != "" || stdin {
				var req clockify.CreateUserGroupRequest
				if err := readJSONInput(filePath, stdin, &req); err != nil {
					return err
				}
				if req.Name != "" {
					name = req.Name
				}
			}

			if name == "" {
				return fmt.Errorf("must specify --name")
			}

			if ifNotExists {
				if existing, err := resolveGroupByName(ctx, client, name); err == nil {
					cmd.Printf("Group %s already exists (ID: %s).\n", existing.Name, existing.ID)
					return formatOutput(cmd, existing, outputFormat, []string{"ID", "NAME", "USER_IDS"})
				} else if cliErr, ok := err.(*clierrors.CLIError); !ok || cliErr.Code != clierrors.E002NotFound {
					return err
				}
			}

			g, err := client.CreateUserGroup(ctx, clockify.CreateUserGroupRequest{Name: name})
			if err != nil {
				return err
			}

			if members != "" {
				for _, ref := range splitCSV(members) {
					uid, err := resolveUserRef(ctx, client, ref)
					if err != nil {
						return err
					}
					g, err = client.AddUserToGroup(ctx, g.ID, uid)
					if err != nil {
						return fmt.Errorf("failed to add member %s: %w", ref, err)
					}
				}
			}

			return formatOutput(cmd, g, outputFormat, []string{"ID", "NAME", "USER_IDS"})
		},
	}
	cmd.Flags().String("name", "", "group name (required)")
	cmd.Flags().String("members", "", "comma-separated user IDs, emails, or names to add")
	cmd.Flags().String("file", "", "JSON file with group data")
	cmd.Flags().Bool("stdin", false, "read group data from stdin")
	cmd.Flags().Bool("if-not-exists", false, "return existing group if one with the same name exists")
	return cmd
}

func newGroupUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a user group and manage members",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")
			newName, _ := cmd.Flags().GetString("new-name")
			addMember, _ := cmd.Flags().GetString("add-member")
			removeMember, _ := cmd.Flags().GetString("remove-member")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			if id == "" && name == "" {
				return fmt.Errorf("must specify --id or --name")
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			if id == "" {
				g, err := resolveGroupByName(ctx, client, name)
				if err != nil {
					return err
				}
				id = g.ID
				name = g.Name
			}

			if newName != "" || filePath != "" || stdin {
				req := clockify.UpdateUserGroupRequest{Name: newName}
				if filePath != "" || stdin {
					var jsonReq clockify.UpdateUserGroupRequest
					if err := readJSONInput(filePath, stdin, &jsonReq); err != nil {
						return err
					}
					if jsonReq.Name != "" {
						req.Name = jsonReq.Name
					}
				}
				if req.Name != "" {
					g, err := client.UpdateUserGroup(ctx, id, req)
					if err != nil {
						return err
					}
					name = g.Name
				}
			}

			var g *clockify.UserGroup
			if addMember != "" {
				uid, err := resolveUserRef(ctx, client, addMember)
				if err != nil {
					return err
				}
				g, err = client.AddUserToGroup(ctx, id, uid)
				if err != nil {
					return err
				}
			} else if removeMember != "" {
				uid, err := resolveUserRef(ctx, client, removeMember)
				if err != nil {
					return err
				}
				if err := client.RemoveUserFromGroup(ctx, id, uid); err != nil {
					return err
				}
				g, err = client.GetUserGroup(ctx, id)
				if err != nil {
					return err
				}
			} else {
				g, err = client.GetUserGroup(ctx, id)
				if err != nil {
					return err
				}
			}

			return formatOutput(cmd, g, outputFormat, []string{"ID", "NAME", "USER_IDS"})
		},
	}
	cmd.Flags().String("id", "", "group ID")
	cmd.Flags().String("name", "", "group name (fuzzy match)")
	cmd.Flags().String("new-name", "", "new group name")
	cmd.Flags().String("add-member", "", "user ID, email, or name to add")
	cmd.Flags().String("remove-member", "", "user ID, email, or name to remove")
	cmd.Flags().String("file", "", "JSON file with group update data")
	cmd.Flags().Bool("stdin", false, "read group update data from stdin")
	return cmd
}

func newGroupDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a user group",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")
			confirm, _ := cmd.Flags().GetBool("confirm")

			if id == "" && name == "" {
				return fmt.Errorf("must specify --id or --name")
			}
			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()
			if id == "" {
				g, err := resolveGroupByName(ctx, client, name)
				if err != nil {
					return err
				}
				id = g.ID
				name = g.Name
			} else if g, err := client.GetUserGroup(ctx, id); err == nil {
				name = g.Name
			}

			if !confirmDestructive(cmd, confirm, "group", id, name) {
				return nil
			}
			if err := client.DeleteUserGroup(ctx, id); err != nil {
				return err
			}
			cmd.Printf("Group %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "group ID")
	cmd.Flags().String("name", "", "group name (fuzzy match)")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	return cmd
}
