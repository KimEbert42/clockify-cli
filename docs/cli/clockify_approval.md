## clockify approval

Manage approval requests

### Synopsis

List, submit, and update Clockify approval requests.

Note: the official Clockify API uses period-based approval submission (WEEKLY/MONTHLY),
not per-time-entry IDs as shown in older spec examples.

### Options

```
  -h, --help   help for approval
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
* [clockify approval list](clockify_approval_list.md)	 - List approval requests
* [clockify approval submit](clockify_approval_submit.md)	 - Submit an approval request for a period
* [clockify approval update](clockify_approval_update.md)	 - Update an approval request status

