## clockify group create

Create a user group

```
clockify group create [flags]
```

### Options

```
      --file string      JSON file with group data
  -h, --help             help for create
      --if-not-exists    return existing group if one with the same name exists
      --members string   comma-separated user IDs, emails, or names to add
      --name string      group name (required)
      --stdin            read group data from stdin
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

* [clockify group](clockify_group.md)	 - Manage user groups

