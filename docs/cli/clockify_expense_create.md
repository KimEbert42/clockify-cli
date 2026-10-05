## clockify expense create

Create an expense

```
clockify expense create [flags]
```

### Options

```
      --amount string        expense amount (required)
      --billable             expense is billable
      --category string      expense category ID or name (overrides --type)
      --date string          expense date YYYY-MM-DD or relative (required)
      --description string   expense notes
      --file string          JSON file with expense data
  -h, --help                 help for create
      --project string       project ID or name (required)
      --stdin                read expense data from stdin
      --type string          expense type/category name (e.g. TRAVEL, MEAL)
      --user string          user ID, email, or name (defaults to current user)
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

* [clockify expense](clockify_expense.md)	 - Manage expenses

