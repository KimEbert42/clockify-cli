## clockify time-entry split

Split a time entry at one or more times

### Synopsis

Divide an entry at the given local times. The original entry ends at the
first split point; remaining segments are created with identical metadata.
Split points must fall strictly inside the entry.

```
clockify time-entry split <id|current|last> --at HH:MM [--at HH:MM...] [flags]
```

### Options

```
      --at stringArray   split point (HH:MM or full timestamp; repeatable)
      --confirm          skip confirmation prompt
  -h, --help             help for split
```

### Options inherited from parent commands

```
      --api-key string     Clockify API key (overrides config)
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

