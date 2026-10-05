# E005 — Rate Limit

**Exit code:** 5  
**Meaning:** Clockify API rate limit exceeded (HTTP 429).

## Causes

- Burst of scripted requests
- Shared IP / many concurrent CLI processes

## Fix

Wait and retry. The CLI already retries with exponential backoff. Reduce parallelism in scripts; prefer `--no-cache` sparingly.

## Example

```text
ERROR [E005]: API request failed with status 429

  Wait a moment and try again

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E005-rate-limit.md
```
