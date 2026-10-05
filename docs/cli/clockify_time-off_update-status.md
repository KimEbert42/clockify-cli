## clockify time-off update-status

Approve or reject a time-off request

```
clockify time-off update-status [flags]
```

### Options

```
  -h, --help            help for update-status
      --id string       time-off request ID
      --note string     note for the status change
      --policy string   policy ID or name (required)
      --status string   APPROVED or REJECTED
      --type string     time-off type label used to resolve a policy
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

* [clockify time-off](clockify_time-off.md)	 - Manage time-off requests

