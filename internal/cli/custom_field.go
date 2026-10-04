package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newCustomFieldCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "custom-field",
		Aliases: []string{"custom-fields", "customfield"},
		Short:   "Manage custom fields",
		Long: `List, create, update, and delete Clockify workspace custom fields.

Supported types (CLI names → API):
  TEXT→TXT, NUMBER, DROPDOWN→DROPDOWN_SINGLE, DROPDOWN_SINGLE, DROPDOWN_MULTIPLE, CHECKBOX, LINK
  DATE is not supported by the Clockify API.`,
	}
	cmd.AddCommand(newCustomFieldListCmd())
	cmd.AddCommand(newCustomFieldCreateCmd())
	cmd.AddCommand(newCustomFieldUpdateCmd())
	cmd.AddCommand(newCustomFieldDeleteCmd())
	return cmd
}

func newCustomFieldListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom fields",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			search, _ := cmd.Flags().GetString("search")
			status, _ := cmd.Flags().GetString("status")
			page, _ := cmd.Flags().GetInt("page")
			pageSize, _ := cmd.Flags().GetInt("page-size")
			noPagination, _ := cmd.Flags().GetBool("no-pagination")

			mappedStatus, err := clockify.MapCustomFieldStatus(status)
			if err != nil {
				return err
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			fields, pageInfo, err := fetchPaginatedCustomFields(context.Background(), client, api.ListCustomFieldsOpts{Name: search, Status: mappedStatus}, page, pageSize, noPagination)
			if err != nil {
				return err
			}
			if len(fields) == 0 && !quiet {
				cmd.Println("No custom fields found.")
				return nil
			}
			type cfDisplay struct {
				ID      string `json:"id" yaml:"id"`
				Name    string `json:"name" yaml:"name"`
				Type    string `json:"type" yaml:"type"`
				Status  string `json:"status" yaml:"status"`
				Options string `json:"options" yaml:"options"`
			}
			displays := make([]cfDisplay, 0, len(fields))
			for _, f := range fields {
				displays = append(displays, cfDisplay{
					ID: f.ID, Name: f.Name, Type: f.Type, Status: f.Status,
					Options: strings.Join(f.AllowedValues, ","),
				})
			}
			if err := formatOutput(cmd, displays, outputFormat, []string{"ID", "NAME", "TYPE", "STATUS", "OPTIONS"}); err != nil {
				return err
			}
			printPaginationSummary(cmd, quiet, outputFormat, pageInfo)
			return nil
		},
	}
	cmd.Flags().String("search", "", "filter by name")
	cmd.Flags().String("status", "", "filter by status: VISIBLE, INVISIBLE, INACTIVE")
	return cmd
}

func newCustomFieldCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom field",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			name, _ := cmd.Flags().GetString("name")
			fieldType, _ := cmd.Flags().GetString("type")
			options, _ := cmd.Flags().GetString("options")
			status, _ := cmd.Flags().GetString("status")
			description, _ := cmd.Flags().GetString("description")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			if filePath != "" || stdin {
				var raw map[string]any
				if err := readJSONInput(filePath, stdin, &raw); err != nil {
					return err
				}
				if v, ok := raw["name"].(string); ok && name == "" {
					name = v
				}
				if v, ok := raw["type"].(string); ok && fieldType == "" {
					fieldType = v
				}
				if v, ok := raw["allowedValues"].([]any); ok && options == "" {
					parts := make([]string, 0, len(v))
					for _, x := range v {
						parts = append(parts, fmt.Sprint(x))
					}
					options = strings.Join(parts, ",")
				}
				if v, ok := raw["status"].(string); ok && status == "" {
					status = v
				}
				if v, ok := raw["description"].(string); ok && description == "" {
					description = v
				}
			}

			if name == "" {
				return fmt.Errorf("must specify --name")
			}
			if fieldType == "" {
				return fmt.Errorf("must specify --type")
			}
			mappedType, err := clockify.MapCustomFieldType(fieldType)
			if err != nil {
				return err
			}
			mappedStatus, err := clockify.MapCustomFieldStatus(status)
			if err != nil {
				return err
			}

			allowed := splitCSV(options)
			if strings.HasPrefix(mappedType, "DROPDOWN") && len(allowed) == 0 {
				return fmt.Errorf("--options is required for DROPDOWN custom fields (comma-separated values)")
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			f, err := client.CreateCustomField(context.Background(), clockify.CreateCustomFieldRequest{
				Name:          name,
				Type:          mappedType,
				AllowedValues: allowed,
				Status:        mappedStatus,
				Description:   description,
			})
			if err != nil {
				return err
			}
			return formatOutput(cmd, f, outputFormat, []string{"ID", "NAME", "TYPE", "STATUS", "ALLOWED_VALUES"})
		},
	}
	cmd.Flags().String("name", "", "custom field name (required)")
	cmd.Flags().String("type", "", "field type: TEXT, NUMBER, DROPDOWN, DROPDOWN_SINGLE, DROPDOWN_MULTIPLE, CHECKBOX, LINK (required)")
	cmd.Flags().String("options", "", "comma-separated allowed values (required for DROPDOWN)")
	cmd.Flags().String("status", "", "VISIBLE, INVISIBLE, or INACTIVE")
	cmd.Flags().String("description", "", "field description")
	cmd.Flags().String("file", "", "JSON file with custom field data")
	cmd.Flags().Bool("stdin", false, "read custom field data from stdin")
	return cmd
}

func newCustomFieldUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a custom field",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			name, _ := cmd.Flags().GetString("name")
			fieldType, _ := cmd.Flags().GetString("type")
			options, _ := cmd.Flags().GetString("options")
			status, _ := cmd.Flags().GetString("status")
			description, _ := cmd.Flags().GetString("description")
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
				f, err := resolveCustomFieldByName(ctx, client, name)
				if err != nil {
					return err
				}
				id = f.ID
				name = f.Name
			}

			req := clockify.UpdateCustomFieldRequest{Name: name}
			if fieldType != "" {
				mapped, err := clockify.MapCustomFieldType(fieldType)
				if err != nil {
					return err
				}
				req.Type = mapped
			}
			if options != "" {
				req.AllowedValues = splitCSV(options)
			}
			if status != "" {
				mapped, err := clockify.MapCustomFieldStatus(status)
				if err != nil {
					return err
				}
				req.Status = mapped
			}
			if description != "" {
				req.Description = description
			}
			if filePath != "" || stdin {
				var jsonReq clockify.UpdateCustomFieldRequest
				if err := readJSONInput(filePath, stdin, &jsonReq); err != nil {
					return err
				}
				if jsonReq.Name != "" {
					req.Name = jsonReq.Name
				}
				if jsonReq.Type != "" {
					mapped, err := clockify.MapCustomFieldType(jsonReq.Type)
					if err != nil {
						return err
					}
					req.Type = mapped
				}
				if len(jsonReq.AllowedValues) > 0 {
					req.AllowedValues = jsonReq.AllowedValues
				}
				if jsonReq.Status != "" {
					req.Status = jsonReq.Status
				}
				if jsonReq.Description != "" {
					req.Description = jsonReq.Description
				}
			}

			f, err := client.UpdateCustomField(ctx, id, req)
			if err != nil {
				return err
			}
			return formatOutput(cmd, f, outputFormat, []string{"ID", "NAME", "TYPE", "STATUS", "ALLOWED_VALUES"})
		},
	}
	cmd.Flags().String("id", "", "custom field ID")
	cmd.Flags().String("name", "", "custom field name (fuzzy match)")
	cmd.Flags().String("type", "", "new field type")
	cmd.Flags().String("options", "", "comma-separated allowed values")
	cmd.Flags().String("status", "", "VISIBLE, INVISIBLE, or INACTIVE")
	cmd.Flags().String("description", "", "field description")
	cmd.Flags().String("file", "", "JSON file with update data")
	cmd.Flags().Bool("stdin", false, "read update data from stdin")
	return cmd
}

func newCustomFieldDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a custom field",
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
				f, err := resolveCustomFieldByName(ctx, client, name)
				if err != nil {
					return err
				}
				id = f.ID
				name = f.Name
			}
			if !confirmDestructive(cmd, confirm, "custom field", id, name) {
				return nil
			}
			if err := client.DeleteCustomField(ctx, id); err != nil {
				return err
			}
			cmd.Printf("Custom field %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "custom field ID")
	cmd.Flags().String("name", "", "custom field name (fuzzy match)")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	return cmd
}
