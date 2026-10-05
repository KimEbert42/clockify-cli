## clockify

A CLI tool for Clockify time tracking

### Synopsis

Clockify CLI - A command-line interface for Clockify time tracking.

Manage workspaces, projects, time entries, and reports directly from your terminal.
Supports multiple output formats (table, JSON, CSV, YAML) and shell completion.

### Options

```
      --api-key string     Clockify API key (overrides config)
      --confirm            skip confirmation prompts
  -h, --help               help for clockify
      --no-cache           bypass local cache and fetch fresh data
      --no-color           disable colored output
      --no-pagination      fetch all pages automatically
      --output string      output format: table, json, csv, yaml, template, ids (default: auto-detect)
      --page int           page number for paginated results (default 1)
      --page-size int      number of items per page (default 50)
      --profile string     config profile to use
      --quiet              suppress non-essential output
      --region string      API region: global, eu-de, us, uk, au
      --template string    template string for --output template (Go text/template)
  -v, --verbose count      verbose output (can be repeated for debug)
      --workspace string   workspace ID or name
```

### SEE ALSO

* [clockify approval](clockify_approval.md)	 - Manage approval requests
* [clockify ce](clockify_ce.md)	 - Create a time entry
* [clockify cfg](clockify_cfg.md)	 - Set a configuration value
* [clockify client](clockify_client.md)	 - Manage clients
* [clockify completion](clockify_completion.md)	 - Generate or install shell completion scripts
* [clockify config](clockify_config.md)	 - Manage CLI configuration
* [clockify configure](clockify_configure.md)	 - Interactive setup wizard for Clockify CLI
* [clockify custom-field](clockify_custom-field.md)	 - Manage custom fields
* [clockify expense](clockify_expense.md)	 - Manage expenses
* [clockify ge](clockify_ge.md)	 - Get time entry by ID, current, or last
* [clockify group](clockify_group.md)	 - Manage user groups
* [clockify in](clockify_in.md)	 - Start a timer
* [clockify invoice](clockify_invoice.md)	 - Manage invoices
* [clockify le](clockify_le.md)	 - List time entries
* [clockify lp](clockify_lp.md)	 - List projects in current workspace
* [clockify lw](clockify_lw.md)	 - List all workspaces
* [clockify out](clockify_out.md)	 - Stop the running timer
* [clockify project](clockify_project.md)	 - Manage projects
* [clockify report](clockify_report.md)	 - Generate time reports
* [clockify schedule](clockify_schedule.md)	 - Manage schedule assignments
* [clockify status](clockify_status.md)	 - Show current authentication and configuration status
* [clockify tag](clockify_tag.md)	 - Manage tags
* [clockify task](clockify_task.md)	 - Manage tasks
* [clockify time-entry](clockify_time-entry.md)	 - Manage time entries
* [clockify time-off](clockify_time-off.md)	 - Manage time-off requests
* [clockify update](clockify_update.md)	 - Update clockify to a newer release
* [clockify user](clockify_user.md)	 - Manage users
* [clockify version](clockify_version.md)	 - Show version information
* [clockify webhook](clockify_webhook.md)	 - Manage webhooks
* [clockify workspace](clockify_workspace.md)	 - Manage workspaces

