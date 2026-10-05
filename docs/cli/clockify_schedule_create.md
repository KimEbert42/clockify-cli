## clockify schedule create

Create a schedule assignment

```
clockify schedule create [flags]
```

### Options

```
      --billable               assignment is billable
      --duration string        alias for --hours-per-day (e.g. 8h)
      --end string             assignment end date (required)
      --file string            JSON file with assignment data
  -h, --help                   help for create
      --hours-per-day string   hours per day (e.g. 8 or 8h)
      --note string            assignment note
      --project string         project ID or name (required)
      --start string           assignment start date (required)
      --stdin                  read assignment data from stdin
      --user string            user ID, email, or name (required)
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

* [clockify schedule](clockify_schedule.md)	 - Manage schedule assignments

