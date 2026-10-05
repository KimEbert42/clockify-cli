## clockify approval list

List approval requests

```
clockify approval list [flags]
```

### Options

```
  -h, --help            help for list
      --status string   filter by status: PENDING, APPROVED, REJECTED, WITHDRAWN_APPROVAL
      --type string     filter by type: TIMESHEET, EXPENSE, TIMESHEET_AND_EXPENSE
      --user string     filter by user ID, email, or name (client-side)
```

### Options inherited from parent commands

```
      --api-key string     Clockify API key (overrides config)
      --confirm            skip confirmation prompts
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

