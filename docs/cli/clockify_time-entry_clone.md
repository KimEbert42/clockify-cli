## clockify time-entry clone

Clone a time entry (start timer or create completed entry)

### Synopsis

Copy project, task, tags, description, and billable from an existing entry.

Without --start/--end/--duration: starts a new timer with copied metadata.
With --start and --end/--duration: creates a completed entry.
Any flag overrides the copied field.

```
clockify time-entry clone [id|last] [flags]
```

### Options

```
      --billable             override billable flag
      --description string   override description
      --duration string      duration for the new entry
      --end string           end time for the new entry
  -h, --help                 help for clone
      --id string            source time entry ID or last
      --project string       override project (ID or name)
      --start string         start time for the new entry (enables completed-entry mode with --end/--duration)
      --tags string          override tags (comma-separated IDs or names)
      --task string          override task (ID or name)
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

