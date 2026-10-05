## clockify invoice create

Create an invoice

```
clockify invoice create [flags]
```

### Options

```
      --client string     client ID or name (required)
      --currency string   currency code e.g. USD (required)
      --end string        due date YYYY-MM-DD or relative (required)
      --file string       JSON file with invoice data
  -h, --help              help for create
      --note string       invoice note
      --number string     invoice number (auto-generated if omitted)
      --start string      issued date YYYY-MM-DD or relative (required)
      --status string     invoice status
      --stdin             read invoice data from stdin
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

* [clockify invoice](clockify_invoice.md)	 - Manage invoices

