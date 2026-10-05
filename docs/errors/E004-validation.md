# E004 — Validation

**Exit code:** 10  
**Meaning:** Invalid input — missing required flags, bad enums, unparseable dates/durations, or API rejected the payload.

## Causes

- Missing `--name` / `--project` / required create fields
- Invalid `--status` value
- Ambiguous resource name (candidates listed in error)
- Bad date or duration format

## Fix

Read the error detail and suggestion. Use `clockify <cmd> --help` for accepted values.

Common formats:

- Duration: `3600`, `1h`, `1.5h`, `30m`, `1h30m`
- Dates: `today`, `yesterday`, `this-week`, `YYYY-MM-DD`, `9:00`, `yesterday 09:00`

## Example

```text
ERROR [E004]: invalid --status "DONE" for project create

  Suggested fix: use ACTIVE or ARCHIVED

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E004-validation.md
```
