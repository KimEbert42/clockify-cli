# E006 — Network

**Exit code:** 6  
**Meaning:** Connection failed (DNS, timeout, TLS, or host unreachable) after retries.

## Causes

- Offline / captive portal
- Corporate proxy
- Clockify outage
- Wrong region endpoint

## Fix

```bash
curl -I https://api.clockify.me/api/v1/user
clockify --region global status
```

If `status` shows `unreachable`, check network/proxy.

## Example

```text
ERROR [E006]: request failed

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E006-network.md
```
