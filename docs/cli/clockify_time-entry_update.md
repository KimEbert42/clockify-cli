## clockify time-entry update

Update one or more time entries

```
clockify time-entry update [id|current|last] [id...] [flags]
```

### Options

```
      --billable             mark as billable
      --description string   description
      --duration string      duration (e.g., 1h, 30m, 1h30m)
      --end string           end time (same formats as --start)
      --file string          JSON file with time entry data
  -h, --help                 help for update
      --id string            time entry ID, current, or last
      --project string       project ID or name
      --start string         start time (RFC3339, YYYY-MM-DD, HH:MM, now, ±offsets, or relative date)
      --stdin                read time entry data from stdin
      --stdin-ids            read time entry IDs from stdin (one per line or JSON array)
      --tags string          tag IDs or names (comma-separated)
      --task string          task ID or name
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

