## clockify task create

Create a task in a project

```
clockify task create [flags]
```

### Options

```
      --assignee string   assignee user ID, email, or name
      --billable          task is billable
      --estimate string   task estimate (seconds, or e.g. 1h, 1.5h, 30m)
      --file string       JSON file with task data
  -h, --help              help for create
      --if-not-exists     return existing task if one with the same name exists in the project
      --name string       task name (required)
      --project string    project ID or name (required)
      --status string     task status: TODO, IN_PROGRESS, DONE
      --stdin             read task data from stdin
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

* [clockify task](clockify_task.md)	 - Manage tasks

