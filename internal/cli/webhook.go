package cli

import (
	"context"
	"fmt"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

func newWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webhook",
		Aliases: []string{"webhooks"},
		Short:   "Manage webhooks",
		Long:    `List, get, create, update, and delete Clockify webhooks.`,
	}
	cmd.AddCommand(newWebhookListCmd())
	cmd.AddCommand(newWebhookGetCmd())
	cmd.AddCommand(newWebhookCreateCmd())
	cmd.AddCommand(newWebhookUpdateCmd())
	cmd.AddCommand(newWebhookDeleteCmd())
	return cmd
}

func newWebhookListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			quiet, _ := cmd.Flags().GetBool("quiet")
			page, _ := cmd.Flags().GetInt("page")
			pageSize, _ := cmd.Flags().GetInt("page-size")
			noPagination, _ := cmd.Flags().GetBool("no-pagination")

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			hooks, pageInfo, err := fetchPaginatedWebhooks(context.Background(), client, api.ListWebhooksOpts{}, page, pageSize, noPagination)
			if err != nil {
				return err
			}
			if len(hooks) == 0 && !quiet {
				cmd.Println("No webhooks found.")
				return nil
			}
			type webhookDisplay struct {
				ID     string `json:"id" yaml:"id"`
				Name   string `json:"name" yaml:"name"`
				URL    string `json:"url" yaml:"url"`
				Event  string `json:"event" yaml:"event"`
				Active bool   `json:"active" yaml:"active"`
			}
			displays := make([]webhookDisplay, 0, len(hooks))
			for _, h := range hooks {
				displays = append(displays, webhookDisplay{
					ID: h.ID, Name: h.Name, URL: h.URL, Event: h.WebhookEvent, Active: h.Enabled,
				})
			}
			if err := formatOutput(cmd, displays, outputFormat, []string{"ID", "NAME", "URL", "EVENT", "ACTIVE"}); err != nil {
				return err
			}
			printPaginationSummary(cmd, quiet, outputFormat, pageInfo)
			return nil
		},
	}
	return cmd
}

func newWebhookGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a webhook by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			if id == "" {
				return fmt.Errorf("must specify --id")
			}
			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			h, err := client.GetWebhook(context.Background(), id)
			if err != nil {
				return err
			}
			return formatOutput(cmd, h, outputFormat, []string{"ID", "NAME", "URL", "EVENT", "ENABLED"})
		},
	}
	cmd.Flags().String("id", "", "webhook ID")
	return cmd
}

func newWebhookCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook",
		Long: `Create a Clockify webhook.

Supported events (official API enum):
` + joinEvents(clockify.ValidWebhookEvents()) + `

Aliases accepted: TIME_ENTRY_CREATED→NEW_TIME_ENTRY, PROJECT_CREATED→NEW_PROJECT, etc.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			url, _ := cmd.Flags().GetString("url")
			events, _ := cmd.Flags().GetString("events")
			name, _ := cmd.Flags().GetString("name")
			active, _ := cmd.Flags().GetBool("active")
			filePath, _ := cmd.Flags().GetString("file")
			stdin, _ := cmd.Flags().GetBool("stdin")

			if filePath != "" || stdin {
				var raw map[string]any
				if err := readJSONInput(filePath, stdin, &raw); err != nil {
					return err
				}
				if v, ok := raw["url"].(string); ok && url == "" {
					url = v
				}
				if v, ok := raw["name"].(string); ok && name == "" {
					name = v
				}
				if v, ok := raw["webhookEvent"].(string); ok && events == "" {
					events = v
				}
				if v, ok := raw["events"].(string); ok && events == "" {
					events = v
				}
			}

			if url == "" {
				return fmt.Errorf("must specify --url")
			}
			if events == "" {
				return fmt.Errorf("must specify --events")
			}

			mapped, err := clockify.MapWebhookEvents(events)
			if err != nil {
				return err
			}
			// API create accepts a single webhookEvent; use first (spec allows comma list —
			// we create one webhook per event when multiple are provided).
			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			ctx := context.Background()

			var last *clockify.Webhook
			for i, ev := range mapped {
				req := clockify.CreateWebhookRequest{
					Name:         name,
					URL:          url,
					WebhookEvent: ev,
					Enabled:      active,
				}
				if len(mapped) > 1 && name != "" {
					req.Name = fmt.Sprintf("%s (%s)", name, ev)
				} else if len(mapped) > 1 {
					req.Name = ev
				}
				h, err := client.CreateWebhook(ctx, req)
				if err != nil {
					if i == 0 {
						return err
					}
					return fmt.Errorf("created %d webhook(s) then failed on %s: %w", i, ev, err)
				}
				last = h
			}
			return formatOutput(cmd, last, outputFormat, []string{"ID", "NAME", "URL", "EVENT", "ENABLED"})
		},
	}
	cmd.Flags().String("url", "", "webhook URL (required)")
	cmd.Flags().String("events", "", "comma-separated webhook events (required)")
	cmd.Flags().String("name", "", "webhook name")
	cmd.Flags().Bool("active", true, "webhook enabled")
	cmd.Flags().String("file", "", "JSON file with webhook data")
	cmd.Flags().Bool("stdin", false, "read webhook data from stdin")
	return cmd
}

func newWebhookUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a webhook",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			outputFormat, _ := cmd.Flags().GetString("output")
			id, _ := cmd.Flags().GetString("id")
			url, _ := cmd.Flags().GetString("url")
			events, _ := cmd.Flags().GetString("events")
			name, _ := cmd.Flags().GetString("name")
			active, _ := cmd.Flags().GetBool("active")
			inactive, _ := cmd.Flags().GetBool("inactive")

			if id == "" {
				return fmt.Errorf("must specify --id")
			}

			req := clockify.UpdateWebhookRequest{}
			if url != "" {
				req.URL = url
			}
			if name != "" {
				req.Name = name
			}
			if events != "" {
				mapped, err := clockify.MapWebhookEvents(events)
				if err != nil {
					return err
				}
				req.WebhookEvent = mapped[0]
			}
			if active {
				t := true
				req.Enabled = &t
			}
			if inactive {
				t := false
				req.Enabled = &t
			}

			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			h, err := client.UpdateWebhook(context.Background(), id, req)
			if err != nil {
				return err
			}
			return formatOutput(cmd, h, outputFormat, []string{"ID", "NAME", "URL", "EVENT", "ENABLED"})
		},
	}
	cmd.Flags().String("id", "", "webhook ID")
	cmd.Flags().String("url", "", "webhook URL")
	cmd.Flags().String("events", "", "comma-separated webhook events")
	cmd.Flags().String("name", "", "webhook name")
	cmd.Flags().Bool("active", false, "enable the webhook")
	cmd.Flags().Bool("inactive", false, "disable the webhook")
	return cmd
}

func newWebhookDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a webhook",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			id, _ := cmd.Flags().GetString("id")
			confirm, _ := cmd.Flags().GetBool("confirm")
			if id == "" {
				return fmt.Errorf("must specify --id")
			}
			client, _, err := newM7Client(profile)
			if err != nil {
				return err
			}
			if !confirmDestructive(cmd, confirm, "webhook", id, "") {
				return nil
			}
			if err := client.DeleteWebhook(context.Background(), id); err != nil {
				return err
			}
			cmd.Printf("Webhook %s deleted.\n", id)
			return nil
		},
	}
	cmd.Flags().String("id", "", "webhook ID")
	cmd.Flags().Bool("confirm", false, "skip confirmation prompt")
	return cmd
}

func joinEvents(events []string) string {
	out := ""
	for i, e := range events {
		if i > 0 {
			out += ", "
		}
		out += e
	}
	return out
}
