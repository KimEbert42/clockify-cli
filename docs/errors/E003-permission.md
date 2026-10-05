# E003 — Permission

**Exit code:** 4  
**Meaning:** Authenticated, but the user lacks workspace role/permission for this operation.

## Causes

- Member without admin rights trying to create/delete projects, manage users, etc.
- Plan-gated features (invoice, approval, webhook) not enabled

## Fix

Ask a workspace admin for the required role, or perform the action in a workspace where you have permission.

## Example

```text
ERROR [E003]: API request failed with status 403

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E003-permission.md
```
