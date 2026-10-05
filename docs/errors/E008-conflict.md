# E008 — Conflict

**Exit code:** 8  
**Meaning:** State conflict — resource already exists, timer already running, or update raced.

## Causes

- Create without `--if-not-exists` when the name exists
- `time-entry start` while another timer runs
- Concurrent edits from web UI + CLI

## Fix

- Prefer `--if-not-exists` for idempotent creates
- `clockify time-entry stop` then `in` again
- Re-fetch with `clockify time-entry get current` before update

## Example

```text
ERROR [E008]: timer already running

  Suggested fix: clockify time-entry stop

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E008-conflict.md
```
