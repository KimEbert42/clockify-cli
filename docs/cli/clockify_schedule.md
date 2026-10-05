## clockify schedule

Manage schedule assignments

### Synopsis

List, create, update, and delete Clockify schedule assignments.

List requires a date range (--start/--end). If omitted, defaults to this-month → today.

### Options

```
  -h, --help   help for schedule
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
* [clockify schedule create](clockify_schedule_create.md)	 - Create a schedule assignment
* [clockify schedule delete](clockify_schedule_delete.md)	 - Delete a schedule assignment
* [clockify schedule list](clockify_schedule_list.md)	 - List schedule assignments
* [clockify schedule update](clockify_schedule_update.md)	 - Update a schedule assignment

