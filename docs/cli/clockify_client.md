## clockify client

Manage clients

### Synopsis

List, get, create, update, and delete Clockify clients.

### Options

```
  -h, --help   help for client
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
* [clockify client create](clockify_client_create.md)	 - Create a client
* [clockify client delete](clockify_client_delete.md)	 - Delete a client
* [clockify client get](clockify_client_get.md)	 - Get client by ID
* [clockify client list](clockify_client_list.md)	 - List clients in current workspace
* [clockify client update](clockify_client_update.md)	 - Update a client

