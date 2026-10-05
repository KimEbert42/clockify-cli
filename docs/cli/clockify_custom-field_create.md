## clockify custom-field create

Create a custom field

```
clockify custom-field create [flags]
```

### Options

```
      --description string   field description
      --file string          JSON file with custom field data
  -h, --help                 help for create
      --name string          custom field name (required)
      --options string       comma-separated allowed values (required for DROPDOWN)
      --status string        VISIBLE, INVISIBLE, or INACTIVE
      --stdin                read custom field data from stdin
      --type string          field type: TEXT, NUMBER, DROPDOWN, DROPDOWN_SINGLE, DROPDOWN_MULTIPLE, CHECKBOX, LINK (required)
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

* [clockify custom-field](clockify_custom-field.md)	 - Manage custom fields

