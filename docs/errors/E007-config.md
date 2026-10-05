# E007 — Config

**Exit code:** 7  
**Meaning:** Config file missing, corrupt, or invalid values (e.g. `interactive`/`auto_update` parse errors).

## Causes

- Hand-edited `config.yaml` with invalid YAML
- Invalid `timezone` / enum-ish keys written by scripts
- Unreadable config directory permissions

## Fix

```bash
clockify config path
clockify config list
clockify configure
```

Repair by re-running `configure`, or delete the broken file and re-create it.

## Example

```text
ERROR [E007]: interactive must be true or false, got "maybe"

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E007-config.md
```
