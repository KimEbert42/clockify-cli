## clockify webhook create

Create a webhook

### Synopsis

Create a Clockify webhook.

Supported events (official API enum):
NEW_PROJECT, PROJECT_UPDATED, PROJECT_DELETED, NEW_TASK, TASK_UPDATED, TASK_DELETED, NEW_CLIENT, CLIENT_UPDATED, CLIENT_DELETED, NEW_TIMER_STARTED, TIMER_STOPPED, NEW_TIME_ENTRY, TIME_ENTRY_UPDATED, TIME_ENTRY_DELETED, TIME_ENTRY_SPLIT, TIME_ENTRY_RESTORED, NEW_TAG, TAG_UPDATED, TAG_DELETED, USER_JOINED_WORKSPACE, USER_DELETED_FROM_WORKSPACE, USER_DEACTIVATED_ON_WORKSPACE, USER_ACTIVATED_ON_WORKSPACE, USER_EMAIL_CHANGED, USER_UPDATED, NEW_INVOICE, INVOICE_UPDATED, NEW_APPROVAL_REQUEST, APPROVAL_REQUEST_STATUS_UPDATED, TIME_OFF_REQUESTED, TIME_OFF_REQUEST_UPDATED, TIME_OFF_REQUEST_APPROVED, TIME_OFF_REQUEST_REJECTED, TIME_OFF_REQUEST_STARTED, TIME_OFF_REQUEST_WITHDRAWN, BALANCE_UPDATED, EXPENSE_CREATED, EXPENSE_UPDATED, EXPENSE_DELETED, EXPENSE_RESTORED, ASSIGNMENT_CREATED, ASSIGNMENT_UPDATED, ASSIGNMENT_DELETED, ASSIGNMENT_PUBLISHED, USER_GROUP_CREATED, USER_GROUP_UPDATED, USER_GROUP_DELETED, USERS_INVITED_TO_WORKSPACE, LIMITED_USERS_ADDED_TO_WORKSPACE, COST_RATE_UPDATED, BILLABLE_RATE_UPDATED

Aliases accepted: TIME_ENTRY_CREATED→NEW_TIME_ENTRY, PROJECT_CREATED→NEW_PROJECT, etc.

```
clockify webhook create [flags]
```

### Options

```
      --active          webhook enabled (default true)
      --events string   comma-separated webhook events (required)
      --file string     JSON file with webhook data
  -h, --help            help for create
      --name string     webhook name
      --stdin           read webhook data from stdin
      --url string      webhook URL (required)
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

* [clockify webhook](clockify_webhook.md)	 - Manage webhooks

