## clockify time-entry

Manage time entries

### Synopsis

List, get, create, update, and delete Clockify time entries.

Selectors for get/update/delete/stop/clone/split:
  <id>     time entry UUID
  current  the running timer
  last     most recent completed entry

### Options

```
  -h, --help   help for time-entry
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

* [clockify](clockify.md)	 - A CLI tool for Clockify time tracking
* [clockify time-entry bulk](clockify_time-entry_bulk.md)	 - Bulk create time entries from JSON
* [clockify time-entry clone](clockify_time-entry_clone.md)	 - Clone a time entry (start timer or create completed entry)
* [clockify time-entry create](clockify_time-entry_create.md)	 - Create a time entry
* [clockify time-entry delete](clockify_time-entry_delete.md)	 - Delete a time entry
* [clockify time-entry get](clockify_time-entry_get.md)	 - Get time entry by ID, current, or last
* [clockify time-entry list](clockify_time-entry_list.md)	 - List time entries
* [clockify time-entry split](clockify_time-entry_split.md)	 - Split a time entry at one or more times
* [clockify time-entry start](clockify_time-entry_start.md)	 - Start a timer
* [clockify time-entry stop](clockify_time-entry_stop.md)	 - Stop the running timer
* [clockify time-entry update](clockify_time-entry_update.md)	 - Update one or more time entries

