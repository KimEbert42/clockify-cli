## clockify config

Manage CLI configuration

### Synopsis

View and modify Clockify CLI configuration settings.

### Options

```
  -h, --help   help for config
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
* [clockify config clear-cache](clockify_config_clear-cache.md)	 - Clear cached data
* [clockify config delete](clockify_config_delete.md)	 - Delete a configuration profile
* [clockify config get](clockify_config_get.md)	 - Get a configuration value
* [clockify config list](clockify_config_list.md)	 - List configuration values
* [clockify config path](clockify_config_path.md)	 - Show configuration file path
* [clockify config set](clockify_config_set.md)	 - Set a configuration value
* [clockify config unset](clockify_config_unset.md)	 - Remove a configuration value

