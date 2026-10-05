# E009 — Server

**Exit code:** 9  
**Meaning:** Clockify API returned a 5xx error.

## Causes

- Transient Clockify incident
- Undocumented edge on regional hosts

## Fix

Retry later. If persistent, check [status.clockify.me](https://status.clockify.me) and report the command + region.

## Example

```text
ERROR [E009]: API request failed with status 500

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E009-server.md
```
