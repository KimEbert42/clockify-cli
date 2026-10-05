# E010 — Internal

**Exit code:** 11  
**Meaning:** Unexpected CLI bug (marshal/unmarshal failure, invalid request construction).

## Causes

- Bug in clockify-cli
- Unexpected API response shape

## Fix

Re-run with `-v -v` (debug). File an issue with:

- `clockify version --output json`
- command line (redact API key)
- debug output (API keys are masked)

## Example

```text
ERROR [E010]: failed to unmarshal response

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E010-internal.md
```
