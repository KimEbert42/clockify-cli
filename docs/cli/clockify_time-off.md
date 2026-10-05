## clockify time-off

Manage time-off requests

### Synopsis

List, create, update status, and delete Clockify time-off requests.

Time-off is policy-scoped in the official API. Use --policy with a policy ID or name.
Spec types (VACATION, SICK, PERSONAL, UNPAID, OTHER) can be used with --type to resolve a matching policy.

### Options

```
  -h, --help   help for time-off
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
* [clockify time-off create](clockify_time-off_create.md)	 - Create a time-off request
* [clockify time-off delete](clockify_time-off_delete.md)	 - Delete a time-off request
* [clockify time-off list](clockify_time-off_list.md)	 - List time-off requests
* [clockify time-off update-status](clockify_time-off_update-status.md)	 - Approve or reject a time-off request

