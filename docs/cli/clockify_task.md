## clockify task

Manage tasks

### Synopsis

List, create, update, and delete Clockify tasks within projects.

### Options

```
  -h, --help   help for task
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
* [clockify task create](clockify_task_create.md)	 - Create a task in a project
* [clockify task delete](clockify_task_delete.md)	 - Delete a task
* [clockify task done](clockify_task_done.md)	 - Mark a task as DONE
* [clockify task list](clockify_task_list.md)	 - List tasks for a project
* [clockify task quick-add](clockify_task_quick-add.md)	 - Create a task with defaults (prints ID only under --quiet)
* [clockify task update](clockify_task_update.md)	 - Update a task

