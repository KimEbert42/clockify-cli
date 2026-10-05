## clockify schedule update

Update a schedule assignment

```
clockify schedule update [flags]
```

### Options

```
      --duration string        alias for --hours-per-day
      --end string             new end date
  -h, --help                   help for update
      --hours-per-day string   new hours per day
      --id string              assignment ID
      --note string            new note
      --series string          series option: THIS_ONE, THIS_AND_FOLLOWING, ALL (default THIS_ONE)
      --start string           new start date
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

