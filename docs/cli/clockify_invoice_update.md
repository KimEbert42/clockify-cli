## clockify invoice update

Update an invoice or change its status

```
clockify invoice update [flags]
```

### Options

```
      --client string     new client ID or name
      --currency string   new currency code
      --end string        new due date
  -h, --help              help for update
      --id string         invoice ID
      --note string       new note
      --number string     new invoice number
      --start string      new issued date
      --status string     new invoice status
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

