# E001 — Auth

**Exit code:** 2  
**Meaning:** Invalid or missing API key, or the key was rejected by Clockify.

## Causes

- API key not configured
- Key revoked / expired in Clockify
- Wrong `--region` (key belongs to another Clockify region)
- Key passed via env/config does not match the workspace

## Fix

```bash
clockify configure
# or
clockify config set api_key ck_live_...
clockify status
```

Verify region:

```bash
clockify config get region
clockify --region global status
```

## Example

```text
ERROR [E001]: API request failed with status 401

  Run 'clockify configure' to set up your API key

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E001-auth.md
```
