## clockify custom-field

Manage custom fields

### Synopsis

List, create, update, and delete Clockify workspace custom fields.

Supported types (CLI names → API):
  TEXT→TXT, NUMBER, DROPDOWN→DROPDOWN_SINGLE, DROPDOWN_SINGLE, DROPDOWN_MULTIPLE, CHECKBOX, LINK
  DATE is not supported by the Clockify API.

### Options

```
  -h, --help   help for custom-field
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
* [clockify custom-field create](clockify_custom-field_create.md)	 - Create a custom field
* [clockify custom-field delete](clockify_custom-field_delete.md)	 - Delete a custom field
* [clockify custom-field list](clockify_custom-field_list.md)	 - List custom fields
* [clockify custom-field update](clockify_custom-field_update.md)	 - Update a custom field

