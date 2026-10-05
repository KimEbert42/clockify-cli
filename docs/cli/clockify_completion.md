## clockify completion

Generate or install shell completion scripts

### Synopsis

Generate shell completion scripts for bash, zsh, fish, or powershell.

Output a script to stdout:
  clockify completion bash

Install to the standard location for your shell:
  clockify completion install --shell zsh

Refresh cached resource names used by dynamic tab completion:
  clockify completion refresh

```
clockify completion [flags]
```

### Options

```
  -h, --help   help for completion
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
* [clockify completion bash](clockify_completion_bash.md)	 - Generate bash completion script
* [clockify completion fish](clockify_completion_fish.md)	 - Generate fish completion script
* [clockify completion install](clockify_completion_install.md)	 - Install completion script to a standard shell path
* [clockify completion powershell](clockify_completion_powershell.md)	 - Generate PowerShell completion script
* [clockify completion refresh](clockify_completion_refresh.md)	 - Refresh cached workspace/project/client/tag/user names for completions
* [clockify completion zsh](clockify_completion_zsh.md)	 - Generate zsh completion script

