## clockify time-entry list

List time entries

```
clockify time-entry list [flags]
```

### Options

```
      --description string   search by description
      --end string           end date (same formats as --start)
  -h, --help                 help for list
      --hydrated             include full project/task/user objects in output
      --in-progress          show only running timers
      --project string       filter by project ID or name
      --start string         start date (YYYY-MM-DD, HH:MM, relative: today, yesterday, this-week, last-week, this-month, last-month, or RFC3339)
      --tags string          filter by tag IDs or names (comma-separated)
      --task string          filter by task ID or name (requires --project)
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

* [clockify time-entry](clockify_time-entry.md)	 - Manage time entries

