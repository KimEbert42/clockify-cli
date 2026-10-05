## clockify report

Generate time reports

### Synopsis

Generate Clockify summary, detailed, and weekly reports.

Reports use the Clockify Reports API (separate host/path from the main API).
Date shortcuts: today, yesterday, this-week, last-week, this-month, last-month.

Presets (summary with matching date range):
  report today|yesterday|this-week|last-week|this-month|last-month

### Options

```
  -h, --help   help for report
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
* [clockify report detailed](clockify_report_detailed.md)	 - Detailed time entry report
* [clockify report last-month](clockify_report_last-month.md)	 - Summary report for last-month
* [clockify report last-week](clockify_report_last-week.md)	 - Summary report for last-week
* [clockify report summary](clockify_report_summary.md)	 - Summary report by project/user
* [clockify report this-month](clockify_report_this-month.md)	 - Summary report for this-month
* [clockify report this-week](clockify_report_this-week.md)	 - Summary report for this-week
* [clockify report today](clockify_report_today.md)	 - Summary report for today
* [clockify report weekly](clockify_report_weekly.md)	 - Weekly report starting from a given date
* [clockify report yesterday](clockify_report_yesterday.md)	 - Summary report for yesterday

