# E002 — Not Found

**Exit code:** 3  
**Meaning:** The requested resource does not exist (or is not visible to this workspace/user).

## Causes

- Wrong or truncated ID
- Resource deleted in Clockify
- Name resolution failed
- Wrong workspace selected

## Fix

```bash
clockify project list
clockify project get --name "Actual Name"
clockify --workspace <id> time-entry list
```

Ambiguous names print candidate IDs — use the full ID.

## Example

```text
ERROR [E002]: API request failed with status 404

  Check the resource ID or name

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E002-not-found.md
```
