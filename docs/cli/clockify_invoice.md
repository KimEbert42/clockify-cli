## clockify invoice

Manage invoices

### Synopsis

List, get, create, update, and delete Clockify invoices.

Status mapping (CLI → API):
  DRAFT, FINALIZED → UNSENT; SENT → SENT; ACCEPTED, PAID → PAID; VOID → VOID
  Official statuses also accepted: UNSENT, PARTIALLY_PAID, OVERDUE

### Options

```
  -h, --help   help for invoice
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
* [clockify invoice create](clockify_invoice_create.md)	 - Create an invoice
* [clockify invoice delete](clockify_invoice_delete.md)	 - Delete an invoice
* [clockify invoice get](clockify_invoice_get.md)	 - Get an invoice by ID
* [clockify invoice list](clockify_invoice_list.md)	 - List invoices
* [clockify invoice update](clockify_invoice_update.md)	 - Update an invoice or change its status

