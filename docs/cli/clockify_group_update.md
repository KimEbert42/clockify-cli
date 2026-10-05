## clockify group update

Update a user group and manage members

```
clockify group update [flags]
```

### Options

```
      --add-member string      user ID, email, or name to add
      --file string            JSON file with group update data
  -h, --help                   help for update
      --id string              group ID
      --name string            group name (fuzzy match)
      --new-name string        new group name
      --remove-member string   user ID, email, or name to remove
      --stdin                  read group update data from stdin
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

