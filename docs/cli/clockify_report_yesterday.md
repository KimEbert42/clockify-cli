## clockify report yesterday

Summary report for yesterday

```
clockify report yesterday [flags]
```

### Options

```
      --billable             show only billable time
      --client string        filter by client (ID or name)
      --description string   filter by description
      --detailed             use detailed report instead of summary
      --end string           end date (YYYY-MM-DD or relative)
      --export string        export format: csv or pdf
      --file string          output path for --export, or JSON file with report parameters
      --group string         filter by user group ID
  -h, --help                 help for yesterday
      --project string       filter by project (ID or name)
      --start string         start date (YYYY-MM-DD or relative)
      --status string        filter by approval state: IN_PROGRESS, NEW, APPROVED, REJECTED
      --stdin                read report parameters JSON from stdin
      --task string          filter by task (ID or name; requires --project)
      --user string          filter by user (ID, email, or name)
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

* [clockify report](clockify_report.md)	 - Generate time reports

