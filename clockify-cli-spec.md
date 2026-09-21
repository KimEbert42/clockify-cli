# Clockify CLI — Design Specification

**Version:** 0.1.0 (Draft)
**Date:** 2026-09-21
**Status:** Design Document

---

## Table of Contents

1. [Design Principles](#1-design-principles)
2. [Command Structure](#2-command-structure)
3. [Authentication & Configuration](#3-authentication--configuration)
4. [Output Formatting](#4-output-formatting)
5. [Error Handling](#5-error-handling)
6. [Advanced Features](#6-advanced-features)
7. [Distribution & Packaging](#7-distribution--packaging)
8. [Complete Command Reference](#8-complete-command-reference)

---

## 1. Design Principles

### 1.1 Core Tenets

| Principle | Description |
|-----------|-------------|
| **Discoverability** | Every command supports `--help`. Subcommands are self-documenting. Users should never need to leave the terminal to find out what's possible. |
| **Progressive Disclosure** | Simple operations are simple. Advanced operations are available but not in the way. Default flags produce human-friendly output; `--output json` unlocks scripting. |
| **Convention over Configuration** | Follow patterns established by `aws`, `gcloud`, and `gh`. Users familiar with any of these should feel at home immediately. |
| **Pipe-Friendly** | Every command that produces data supports `--output json`. Output is newline-delimited JSON (JSONL) when piped, table when terminal. |
| **Idempotent by Default** | Create/update commands accept `--if-not-exists` or use PATCH semantics where the API allows. Destructive commands require `--confirm` or `--force`. |
| **Fast Startup** | Binary should start in <50ms. No runtime dependencies. Single static binary per platform. |

### 1.2 Naming Conventions

- **Commands:** lowercase, hyphen-separated (`list-time-entries`, not `listTimeEntries`)
- **Flags:** lowercase, hyphen-separated with `--` prefix (`--workspace-id`, not `--workspaceId`)
- **Resource names:** singular in commands (`workspace` not `workspaces`), plural in list operations
- **Verb-first ordering:** `list`, `get`, `create`, `update`, `delete`, `stop`, `start`
- **Boolean flags:** `--flag` (presence = true), `--no-flag` (presence = false)

### 1.3 Alias System

Every resource gets a short alias:

| Full Command | Alias |
|-------------|-------|
| `clockify list workspaces` | `clockify lw` |
| `clockify list time-entries` | `clockify le` |
| `clockify list projects` | `clockify lp` |
| `clockify get time-entry` | `clockify ge` |
| `clockify create time-entry` | `clockify ce` |
| `clockify config set` | `clockify cfg` |

---

## 2. Command Structure

### 2.1 Top-Level Layout

```
clockify
├── configure          # Interactive setup wizard
├── config             # Direct config manipulation
├── status             # Show current auth & workspace context
├── workspace          # Workspace operations
│   ├── list
│   └── get
├── user               # User operations
│   ├── get            # Current user
│   ├── list           # List workspace users
│   └── info           # Filter/search users
├── project            # Project CRUD
│   ├── list
│   ├── get
│   ├── create
│   ├── update
│   └── delete
├── task               # Task CRUD (nested under project)
│   ├── list
│   ├── create
│   ├── update
│   └── delete
├── time-entry         # Time entry CRUD + timer
│   ├── list
│   ├── get
│   ├── create
│   ├── update
│   ├── delete
│   ├── start          # Start timer
│   ├── stop           # Stop running timer
│   └── bulk           # Bulk edit
├── client             # Client CRUD
│   ├── list
│   ├── get
│   ├── create
│   ├── update
│   └── delete
├── tag                # Tag CRUD
│   ├── list
│   ├── create
│   ├── update
│   └── delete
├── report             # Reports
│   ├── detailed
│   ├── summary
│   └── weekly
├── invoice            # Invoice CRUD
│   ├── list
│   ├── get
│   ├── create
│   ├── update
│   └── delete
├── schedule           # Scheduling assignments
│   ├── list
│   ├── create
│   ├── update
│   └── delete
├── time-off           # Time off requests
│   ├── list
│   ├── create
│   └── update-status
├── expense            # Expense CRUD
│   ├── list
│   ├── get
│   ├── create
│   ├── update
│   └── delete
├── custom-field       # Custom field CRUD
│   ├── list
│   ├── create
│   ├── update
│   └── delete
├── webhook            # Webhook CRUD
│   ├── list
│   ├── create
│   ├── update
│   └── delete
├── approval           # Approval workflow
│   ├── list
│   ├── submit
│   └── update
├── group              # User group CRUD
│   ├── list
│   ├── create
│   ├── update
│   └── delete
├── completion         # Shell completion scripts
├── update             # Self-update
└── version            # Version info
```

### 2.2 Global Flags

Every `clockify` command accepts these flags:

```
--profile <name>        Use a named configuration profile (default: "default")
--workspace <id>        Override the configured workspace ID
--output <format>       Output format: table (default), json, yaml, csv
--region <server>       API region: global, eu-de, us, uk, au (default: from config)
--api-key <key>         Override API key (env: CLOCKIFY_API_KEY)
--no-color              Disable colored output
--verbose / -v          Verbose output (show HTTP requests)
--debug                 Debug output (show headers, payloads, timing)
--quiet / -q            Suppress non-essential output
--confirm / -y          Skip confirmation prompts
--page <n>              Page number for paginated results
--page-size <n>         Results per page (default: 100, max: 5000)
--no-pagination         Fetch all pages automatically
```

### 2.3 Resource ID Resolution

Clockify uses UUIDs extensively. The CLI supports multiple ways to reference resources:

```bash
# By exact ID
clockify get project --id 5f8a1b2c3d4e5f6a7b8c9d0e

# By name (fuzzy match with warning)
clockify get project --name "My Project"

# By prefix (unique prefix match)
clockify get project --id 5f8a1b2c

# Current workspace context (omitting --workspace)
clockify list projects
```

When `--name` is used and multiple matches exist, the CLI errors with a list of matches and suggests using `--id`.

### 2.4 Argument Patterns

**CRUD Commands:**
```bash
clockify <resource> <action> [flags]
```

**Nested Resources:**
```bash
clockify task list --project <projectId>
clockify task create --project <projectId> --name "Task Name"
```

**ID-Agnostic Lists:**
```bash
clockify list projects                          # current workspace
clockify list projects --workspace <workspaceId> # specific workspace
clockify list projects --all-workspaces          # all workspaces (where supported)
```

---

## 3. Authentication & Configuration

### 3.1 Configuration File Location

| Platform | Path |
|----------|------|
| Linux/macOS | `~/.clockify/config.yaml` |
| Windows | `%USERPROFILE%\.clockify\config.yaml` |
| XDG (Linux) | `$XDG_CONFIG_HOME/clockify/config.yaml` |

### 3.2 Config File Format

```yaml
# ~/.clockify/config.yaml
version: 1

# Default profile
default:
  api_key: "ck_live_xxxxxxxxxxxxxxxxxxxx"  # Encrypted at rest (see 3.3)
  workspace_id: "5f8a1b2c3d4e5f6a7b8c9d0e"
  region: "global"                         # global | eu-de | us | uk | au
  output: "table"                          # table | json | yaml | csv
  date_format: "2006-01-02"
  time_format: "15:04"
  week_starts: "MONDAY"                    # MONDAY | SUNDAY

# Named profiles
work:
  api_key: "ck_live_yyyyyyyyyyyyyyyyyyyy"
  workspace_id: "6g9b2c3d4e5f6a7b8c9d0e1f"
  region: "eu-de"
  output: "json"

personal:
  api_key: "ck_live_zzzzzzzzzzzzzzzzzzzz"
  workspace_id: "7h0c3d4e5f6a7b8c9d0e1f2a"
  region: "global"
```

### 3.3 Credential Storage

**Primary method:** OS keychain via `keyring` library.

| Platform | Backend |
|----------|---------|
| macOS | Keychain |
| Linux | Secret Service (GNOME Keyring / KWallet) |
| Windows | Windows Credential Manager |

Fallback for headless/CI: `CLOCKIFY_API_KEY` environment variable.

```bash
# Store API key securely
clockify configure
# → Prompts for API key, stores in OS keychain

# Store in file (less secure, for CI)
clockify config set api-key ck_live_xxx --insecure-store
```

### 3.4 Authentication Flow

```bash
# First-time setup (interactive wizard)
$ clockify configure
? API Key: ck_live_xxxxxxxxxxxxxxxxxxxx
? Default workspace: (auto-detected) Acme Corp (5f8a1b...)
? Region: Global (api.clockify.me)
? Default output format: table
✓ Configuration saved to ~/.clockify/config.yaml

# Quick auth via flag
$ clockify --api-key ck_live_xxx status

# Quick auth via environment
$ export CLOCKIFY_API_KEY=ck_live_xxx
$ clockify status
```

### 3.5 Multi-Workspace Support

```bash
# Switch active workspace
clockify config set workspace_id 5f8a1b2c3d4e5f6a7b8c9d0e

# Or use a profile
clockify --profile work list projects

# Or list across all workspaces
clockify list projects --all-workspaces
```

### 3.6 Region Mapping

| Flag Value | API Host | Reports Host |
|------------|----------|--------------|
| `global` | `api.clockify.me` | `reports.api.clockify.me` |
| `eu-de` | `euc1.clockify.me` | `reports.euc1.clockify.me` |
| `us` | `use2.clockify.me` | `reports.use2.clockify.me` |
| `uk` | `euw2.clockify.me` | `reports.euw2.clockify.me` |
| `au` | `apse2.clockify.me` | `reports.apse2.clockify.me` |

---

## 4. Output Formatting

### 4.1 Table Format (Default)

Human-friendly, auto-width, with color:

```
$ clockify list projects
ID                     NAME              CLIENT        Billable  STATUS
─────────────────────  ────────────────  ────────────  ────────  ──────
5f8a1b2c3d4e5f6a7b8c  Website Redesign  Acme Corp     Yes       ACTIVE
6g9b2c3d4e5f6a7b8c9d  Mobile App        TechStart     No        ACTIVE
7h0c3d4e5f6a7b8c9d0e  Internal Ops      —             No        ARCHIVED

3 projects (page 1 of 1)
```

### 4.2 JSON Format

```bash
$ clockify list projects --output json
[
  {
    "id": "5f8a1b2c3d4e5f6a7b8c9d0e",
    "name": "Website Redesign",
    "clientId": "4e7f0a1b2c3d4e5f6a7b8c9d",
    "billable": true,
    "status": "ACTIVE",
    "duration": null,
    "estimate": { "type": "AUTO", "estimate": 7200 },
    "activeTaskCount": 5,
    "color": "#e74c3c"
  }
]
```

### 4.3 CSV Format

```bash
$ clockify list projects --output csv
id,name,client_id,billable,status
5f8a1b2c3d4e5f6a7b8c9d0e,Website Redesign,4e7f0a1b2c3d4e5f6a7b8c9d,Yes,ACTIVE
```

### 4.4 YAML Format

```bash
$ clockify list projects --output yaml
- id: 5f8a1b2c3d4e5f6a7b8c9d0e
  name: Website Redesign
  client_id: 4e7f0a1b2c3d4e5f6a7b8c9d
  billable: true
  status: ACTIVE
```

### 4.5 Auto-Detection

- Piped to another command → JSON (unless `--output` specified)
- TTY detected → table
- `--quiet` → suppress headers and footers

### 4.6 Progress & Spinners

```bash
$ clockify list time-entries --all-workspaces --no-pagination
⠋ Fetching page 1/5...
⠋ Fetching page 2/5...
⠋ Fetching page 3/5...
⠋ Fetching page 4/5...
⠋ Fetching page 5/5...
1,247 time entries fetched.
```

Disable with `--quiet` or `NO_COLOR=1`.

---

## 5. Error Handling

### 5.1 Error Message Format

```
ERROR: <concise description>

  <detailed context>

  Suggested fix: <actionable step>

  Learn more: https://clockify.github.io/cli/errors/E001
```

### 5.2 Error Categories

| Code | Category | Example |
|------|----------|---------|
| E001 | Auth | Invalid API key, expired token |
| E002 | Not Found | Resource doesn't exist |
| E003 | Permission | Insufficient workspace role |
| E004 | Validation | Invalid input (missing required fields) |
| E005 | Rate Limit | Too many requests (50/s limit) |
| E006 | Network | Connection timeout, DNS failure |
| E007 | Config | Missing or corrupt config file |
| E008 | Conflict | Resource already exists |
| E009 | Server | Clockify API 5xx error |
| E010 | Internal | CLI bug (include --debug output in report) |

### 5.3 Verbose & Debug Modes

```bash
# Verbose: show HTTP method + URL
$ clockify list projects -v
GET https://api.clockify.me/api/v1/workspaces/5f8a1b.../projects?page=1&pageSize=100

# Debug: show full request/response
$ clockify list projects --debug
> GET /api/v1/workspaces/5f8a1b.../projects?page=1&pageSize=100 HTTP/1.1
> Host: api.clockify.me
> X-Api-Key: ck_live_****xxxx
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< X-Rate-Limit-Remaining: 49
< Last-Page: 3
< [... response body ...]
```

### 5.4 Retry Logic

Automatic retries with exponential backoff:
- **429 Rate Limit:** Retry after `Retry-After` header or 2s, up to 5 retries
- **5xx Server Error:** Retry after 1s, 2s, 4s, 8s, 16s (max 5 retries)
- **Network Error:** Retry after 1s, 2s, 4s (max 3 retries)
- **4xx Client Error:** No retry (user must fix input)

### 5.5 Destructive Operation Confirmation

```bash
$ clockify project delete --id 5f8a1b2c3d4e5f6a7b8c9d0e
⚠ This will permanently delete project "Website Redesign" and all its tasks.
  This action cannot be undone.

? Type the project name to confirm: Website Redesign
✓ Project "Website Redesign" deleted.

# Skip confirmation in CI
$ clockify project delete --id 5f8a1b2c --confirm
```

---

## 6. Advanced Features

### 6.1 Shell Completion

```bash
# Install completions
clockify completion install --shell bash    # ~/.bash_completion
clockify completion install --shell zsh     # ~/.zsh/completions/_clockify
clockify completion install --shell fish    # ~/.config/fish/completions/clockify.fish
clockify completion install --shell powershell

# Output completion script to stdout
clockify completion bash > /etc/bash_completion.d/clockify
```

Completion includes:
- Command and subcommand names
- Flag names and value hints
- Dynamic workspace/project/tag IDs (cached locally, refresh with `clockify completion refresh`)
- Resource name completion for `--name` flags

### 6.2 Environment Variables

All flags can be set via environment variables with `CLOCKIFY_` prefix:

| Flag | Environment Variable |
|------|---------------------|
| `--api-key` | `CLOCKIFY_API_KEY` |
| `--workspace` | `CLOCKIFY_WORKSPACE_ID` |
| `--profile` | `CLOCKIFY_PROFILE` |
| `--output` | `CLOCKIFY_OUTPUT` |
| `--region` | `CLOCKIFY_REGION` |
| `--no-color` | `NO_COLOR` (standard) |
| `--page` | `CLOCKIFY_PAGE` |
| `--page-size` | `CLOCKIFY_PAGE_SIZE` |

Priority order: Flag > Env Var > Config File > Default

### 6.3 Piping & Chaining

```bash
# Pipe JSON to jq
clockify list time-entries --output json | jq '.[] | select(.description | contains("meeting"))'

# Get IDs for scripting
PROJECT_ID=$(clockify list projects --output json | jq -r '.[] | select(.name == "Website") | .id')

# Create entry using piped input
echo '{"description":"Lunch break","duration":3600}' | clockify create time-entry --stdin

# Chain with xargs
clockify list projects --output json | jq -r '.[].id' | xargs -I{} clockify list tasks --project {}
```

### 6.4 Input via Stdin / File

```bash
# From file
clockify create time-entry --file entry.json

# From stdin
cat entry.json | clockify create time-entry --stdin

# Interactive prompt for individual fields
clockify create time-entry --interactive
? Start time: 2026-09-21T09:00:00Z
? Duration (hours): 2.5
? Description: Sprint planning
? Project: (type to search) Website Redesign
? Tags: (comma-separated) meeting, planning
✓ Time entry created: 6g9b2c3d4e5f6a7b8c9d0e1f
```

### 6.5 Plugin System (Future)

```yaml
# ~/.clockify/config.yaml
plugins:
  - name: clockify-report-pdf
    source: github.com/clockify/cli-plugin-pdf
    version: ">=1.0.0"
```

```bash
clockify plugin install clockify-report-pdf
clockify plugin list
clockify plugin update --all
```

### 6.6 CI/CD Mode

```bash
# Non-interactive, machine-readable, exit codes
clockify list time-entries \
  --output json \
  --quiet \
  --no-color \
  --page-size 5000 \
  --no-pagination

# Exit codes:
# 0 = success
# 1 = general error
# 2 = auth error
# 3 = not found
# 4 = permission error
# 5 = rate limited
# 6 = network error
```

### 6.7 Caching

Local cache for frequently accessed metadata:

```
~/.clockify/cache/
├── workspaces.json      # TTL: 1h
├── projects.json        # TTL: 5m
├── tags.json            # TTL: 5m
├── clients.json         # TTL: 5m
└── users.json           # TTL: 5m
```

```bash
# Force refresh
clockify list projects --no-cache

# Clear all cache
clockify config clear-cache
```

---

## 7. Distribution & Packaging

### 7.1 Binary Distribution

| Platform | Format | Location |
|----------|--------|----------|
| Linux x86_64 | Static binary | GitHub Releases |
| Linux ARM64 | Static binary | GitHub Releases |
| macOS x86_64 | Universal binary | GitHub Releases |
| macOS ARM64 | Universal binary | GitHub Releases |
| Windows x86_64 | .exe + .zip | GitHub Releases |

### 7.2 Package Managers

| Manager | Install Command |
|---------|----------------|
| Homebrew (macOS/Linux) | `brew install clockify/tap/clockify` |
| npm (cross-platform) | `npm install -g clockify-cli` |
| pip (cross-platform) | `pip install clockify-cli` |
| AUR (Arch Linux) | `yay -S clockify-cli` |
| Snap | `snap install clockify` |
| Go install | `go install github.com/clockify/cli@latest` |

### 7.3 Auto-Update

```bash
# Check for updates
clockify update --check

# Update to latest
clockify update

# Update to specific version
clockify update --version 1.2.3

# Disable auto-update checks
clockify config set auto_update false
```

Update check hits GitHub Releases API on first command of the day (cached for 24h). No check if `CLOCKIFY_NO_UPDATE_CHECK=1` or in CI.

### 7.4 Versioning

Follows [Semantic Versioning](https://semver.org/):

```
clockify version
clockify-cli 1.2.3
commit:     abc1234
built:      2026-09-21
go:         1.22
platform:   linux/amd64
```

---

## 8. Complete Command Reference

### 8.1 `clockify configure`

Interactive setup wizard.

```bash
clockify configure                  # Run full wizard
clockify configure --profile work   # Configure a named profile
clockify configure --reset          # Reset all configuration
```

---

### 8.2 `clockify config`

Direct configuration manipulation.

```bash
clockify config list                           # Show all config values
clockify config get <key>                      # Get a single value
clockify config set <key> <value>              # Set a value
clockify config unset <key>                    # Remove a value
clockify config path                           # Show config file path
clockify config clear-cache                    # Clear local cache
```

**Examples:**
```bash
clockify config set workspace_id 5f8a1b2c3d4e5f6a7b8c9d0e
clockify config set output json
clockify config set region eu-de
clockify config get api_key                    # Shows masked: ck_live_****xxxx
```

---

### 8.3 `clockify status`

Show current authentication and context.

```bash
clockify status
clockify status --output json
```

**Output:**
```
Profile:      default
API Key:      ck_live_****xxxx ✓
Workspace:    Acme Corp (5f8a1b2c3d4e5f6a7b8c9d0e)
Region:       Global (api.clockify.me)
Output:       table
Workspace Users: 24
Projects:     12
```

---

### 8.4 `clockify workspace`

```bash
clockify workspace list                        # List all workspaces
clockify workspace get --id <workspaceId>      # Get workspace details
clockify workspace get --current               # Get current workspace (default)
```

**Flags:**
- `--id <id>` — Workspace ID
- `--name <name>` — Workspace name (fuzzy match)

---

### 8.5 `clockify user`

```bash
clockify user get                              # Get current user
clockify user list                             # List users in current workspace
clockify user list --status ACTIVE             # Filter by status
clockify user info --email user@acme.com       # Search users
```

**Flags (list):**
- `--status <status>` — Filter: ACTIVE, INACTIVE
- `--page-size <n>` — Results per page (default: 100)

**Flags (info):**
- `--email <email>` — Search by email
- `--name <name>` — Search by name

---

### 8.6 `clockify project`

```bash
clockify project list                          # List projects in current workspace
clockify project list --status ACTIVE          # Filter by status
clockify project list --client <clientId>      # Filter by client
clockify project list --search "web"           # Search by name

clockify project get --id <projectId>          # Get project details
clockify project get --name "Website"          # Get by name

clockify project create --name "New Project"   # Create project
clockify project create --name "Billable" --billable --client <clientId>

clockify project update --id <id> --name "Updated Name"
clockify project update --id <id> --status ARCHIVED

clockify project delete --id <projectId>       # Delete project (with confirmation)
clockify project delete --id <projectId> --confirm  # Skip confirmation
```

**Flags (create):**
- `--name <name>` — Project name (required)
- `--client <id>` — Client ID
- `--billable` / `--no-billable` — Billable status (default: false)
- `--color <hex>` — Project color (e.g., `#e74c3c`)
- `--estimate <type>` — Estimate type: AUTO, MANUAL, NO_ESTIMATE
- `--estimate-value <seconds>` — Estimate in seconds
- `--file <path>` — JSON file with project data
- `--stdin` — Read JSON from stdin

**Flags (update):**
- `--id <id>` — Project ID (required)
- `--name <name>` — New name
- `--client <id>` — New client
- `--billable` / `--no-billable`
- `--status <status>` — ACTIVE, ARCHIVED
- `--color <hex>`
- `--file <path>` — JSON file with update data

**Flags (delete):**
- `--id <id>` — Project ID (required)
- `--confirm` — Skip confirmation prompt

---

### 8.7 `clockify task`

```bash
clockify task list --project <projectId>               # List tasks for project
clockify task list --project <projectId> --is-active   # Active tasks only

clockify task create --project <projectId> --name "Task 1"
clockify task create --project <projectId> --name "Task 1" --estimate 3600

clockify task update --project <projectId> --id <taskId> --name "Renamed"
clockify task update --project <projectId> --id <taskId> --status DONE

clockify task delete --project <projectId> --id <taskId>
```

**Flags (create):**
- `--project <id>` — Parent project ID (required)
- `--name <name>` — Task name (required)
- `--estimate <seconds>` — Estimate in seconds
- `--status <status>` — TODO, IN_PROGRESS, DONE
- `--assignee <userId>` — Assignee user ID
- `--file <path>` / `--stdin` — JSON input

---

### 8.8 `clockify time-entry`

```bash
clockify time-entry list                                            # List current user's entries
clockify time-entry list --start 2026-09-01 --end 2026-09-21       # Date range
clockify time-entry list --project <projectId>                      # Filter by project
clockify time-entry list --user <userId>                            # Other user's entries
clockify time-entry list --description "meeting"                    # Filter by description
clockify time-entry list --tags meeting,planning                    # Filter by tags

clockify time-entry get --id <timeEntryId>                          # Get specific entry

clockify time-entry create --project <id> --start "2026-09-21T09:00:00Z" --duration 3600
clockify time-entry create --project <id> --start "2026-09-21T09:00:00Z" --end "2026-09-21T11:00:00Z"
clockify time-entry create --project <id> --start "2026-09-21T09:00:00Z" --duration 1.5h

clockify time-entry update --id <id> --description "Updated desc"
clockify time-entry update --id <id> --start "2026-09-21T10:00:00Z"

clockify time-entry delete --id <id>

clockify time-entry start --project <id>                            # Start timer (no duration)
clockify time-entry stop                                             # Stop running timer
clockify time-entry stop --description "Finished coding"            # Stop with description

clockify time-entry bulk --file bulk-entries.json                   # Bulk create/update
```

**Duration Formats:**
- Seconds: `3600`
- Human: `1h`, `1.5h`, `30m`, `1h30m`, `2h15m`

**Flags (list):**
- `--start <date>` — Start date (ISO 8601 or relative: `today`, `yesterday`, `this-week`, `last-week`, `this-month`, `last-month`)
- `--end <date>` — End date
- `--project <id>` — Filter by project
- `--task <id>` — Filter by task
- `--user <id>` — Filter by user
- `--description <text>` — Filter by description (partial match)
- `--tags <tag1,tag2>` — Filter by tags (AND logic)
- `--billable` / `--not-billable` — Filter by billable status
- `--in-progress` — Only running entries
- `--hydrated` — Include full project/task/user objects (slower)

**Flags (create):**
- `--project <id>` — Project ID (required unless using --stdin)
- `--task <id>` — Task ID
- `--start <datetime>` — Start time (required)
- `--duration <value>` — Duration (required unless --end)
- `--end <datetime>` — End time (alternative to --duration)
- `--description <text>` — Description
- `--tags <id1,id2>` — Tag IDs
- `--billable` / `--no-billable` — Billable flag
- `--custom-field <fieldId=value>` — Custom field values
- `--file <path>` / `--stdin` — JSON input

---

### 8.9 `clockify client`

```bash
clockify client list                                    # List clients
clockify client list --search "Acme"                    # Search by name
clockify client get --id <clientId>
clockify client create --name "New Client"
clockify client create --name "New Client" --email billing@newclient.com
clockify client update --id <id> --name "Renamed Client"
clockify client delete --id <id>
```

**Flags (create):**
- `--name <name>` — Client name (required)
- `--email <email>` — Client email
- `--address <address>` — Client address
- `--phone <phone>` — Client phone
- `--note <note>` — Client note
- `--file <path>` / `--stdin` — JSON input

---

### 8.10 `clockify tag`

```bash
clockify tag list                                       # List tags
clockify tag list --search "meeting"
clockify tag create --name "meeting"
clockify tag create --name "meeting" --color "#3498db"
clockify tag update --id <id> --name "team meeting"
clockify tag delete --id <id>
```

**Flags (create):**
- `--name <name>` — Tag name (required)
- `--color <hex>` — Tag color

---

### 8.11 `clockify report`

```bash
clockify report summary --start 2026-09-01 --end 2026-09-21
clockify report summary --start 2026-09-01 --end 2026-09-21 --project <id>
clockify report summary --start 2026-09-01 --end 2026-09-21 --user <id>
clockify report summary --start 2026-09-01 --end 2026-09-21 --output json

clockify report detailed --start 2026-09-01 --end 2026-09-21
clockify report detailed --start 2026-09-01 --end 2026-09-21 --project <id>
clockify report detailed --start 2026-09-01 --end 2026-09-21 --export csv

clockify report weekly --start 2026-09-15
clockify report weekly --start 2026-09-15 --user <id>
```

**Flags (all reports):**
- `--start <date>` — Start date (required)
- `--end <date>` — End date (required except weekly)
- `--project <id>` — Filter by project
- `--task <id>` — Filter by task
- `--user <id>` — Filter by user
- `--group <id>` — Filter by user group
- `--client <id>` — Filter by client
- `--status <status>` — IN_PROGRESS, NEW, APPROVED, REJECTED
- `--export <format>` — Export format: csv, pdf (if supported)
- `--file <path>` — JSON file with report parameters

**Report-specific flags (summary):**
- `--billable` — Show only billable time
- `--description <text>` — Filter by description

---

### 8.12 `clockify invoice`

```bash
clockify invoice list                                          # List invoices
clockify invoice get --id <invoiceId>
clockify invoice create --client <clientId> --start 2026-09-01 --end 2026-09-30
clockify invoice create --client <clientId> --start 2026-09-01 --end 2026-09-30 --status DRAFT
clockify invoice update --id <id> --status SENT
clockify invoice delete --id <id>
```

**Flags (create):**
- `--client <id>` — Client ID (required)
- `--start <date>` — Start date (required)
- `--end <date>` — End date (required)
- `--status <status>` — DRAFT, FINALIZED, SENT, ACCEPTED, PAID, VOID
- `--currency <code>` — Currency code (e.g., USD, EUR)
- `--note <text>` — Invoice note
- `--file <path>` / `--stdin` — JSON input

---

### 8.13 `clockify schedule`

```bash
clockify schedule list                                                    # List assignments
clockify schedule list --start 2026-09-01 --end 2026-09-30                # Date range
clockify schedule list --user <userId>                                    # User's schedule
clockify schedule list --project <projectId>                              # Project assignments

clockify schedule create --user <userId> --project <id> --start 2026-09-21 --end 2026-09-25
clockify schedule create --user <userId> --project <id> --start 2026-09-21 --duration 8h

clockify schedule update --id <id> --start 2026-09-22
clockify schedule delete --id <id>
```

---

### 8.14 `clockify time-off`

```bash
clockify time-off list                                                   # List time-off requests
clockify time-off list --status PENDING                                  # Filter by status
clockify time-off list --user <userId>                                   # User's requests

clockify time-off create --type VACATION --start 2026-12-20 --end 2026-12-31
clockify time-off create --type SICK --start 2026-09-21 --end 2026-09-21 --note "Feeling unwell"

clockify time-off update-status --id <id> --status APPROVED
clockify time-off update-status --id <id> --status REJECTED --note "Busy period"
```

**Time-off Types:** VACATION, SICK, PERSONAL, UNPAID, OTHER

---

### 8.15 `clockify expense`

```bash
clockify expense list                                            # List expenses
clockify expense list --project <projectId>                      # Filter by project
clockify expense list --start 2026-09-01 --end 2026-09-30       # Date range

clockify expense get --id <expenseId>
clockify expense create --project <id> --date 2026-09-21 --amount 45.50 --type TRAVEL
clockify expense create --project <id> --date 2026-09-21 --amount 45.50 --type TRAVEL --description "Uber to client"

clockify expense update --id <id> --amount 50.00
clockify expense delete --id <id>
```

**Expense Types:** TRAVEL, MEAL, SUPPLIES, SOFTWARE, OTHER

---

### 8.16 `clockify custom-field`

```bash
clockify custom-field list                                           # List custom fields
clockify custom-field create --name "Department" --type DROPDOWN --options "Engineering,Marketing,Sales"
clockify custom-field create --name "Priority" --type DROPDOWN --options "Low,Medium,High"
clockify custom-field create --name "Approved" --type CHECKBOX
clockify custom-field update --id <id> --name "Dept"
clockify custom-field delete --id <id>
```

**Field Types:** DROPDOWN, CHECKBOX, TEXT, NUMBER, LINK, DATE

---

### 8.17 `clockify webhook`

```bash
clockify webhook list                                            # List webhooks
clockify webhook create --url "https://example.com/hook" --events "TIME_ENTRY_CREATED,TIME_ENTRY_UPDATED"
clockify webhook update --id <id> --active
clockify webhook update --id <id> --inactive
clockify webhook delete --id <id>
```

**Webhook Events:**
- `TIME_ENTRY_CREATED`, `TIME_ENTRY_UPDATED`, `TIME_ENTRY_DELETED`
- `PROJECT_CREATED`, `PROJECT_UPDATED`, `PROJECT_DELETED`
- `USER_CREATED`, `USER_UPDATED`, `USER_DELETED`
- `CLIENT_CREATED`, `CLIENT_UPDATED`, `CLIENT_DELETED`

---

### 8.18 `clockify approval`

```bash
clockify approval list                                                      # List approval requests
clockify approval list --status PENDING                                      # Filter by status
clockify approval list --user <userId>

clockify approval submit --time-entry <entryId> --note "Please approve"
clockify approval submit --time-entry <id1> --time-entry <id2> --note "Weekly batch"

clockify approval update --id <id> --status APPROVED
clockify approval update --id <id> --status REJECTED --note "Missing details"
```

---

### 8.19 `clockify group`

```bash
clockify group list                                               # List groups
clockify group create --name "Engineering"
clockify group create --name "Engineering" --members <userId1>,<userId2>
clockify group update --id <id> --name "Platform Engineering"
clockify group update --id <id> --add-member <userId>
clockify group update --id <id> --remove-member <userId>
clockify group delete --id <id>
```

---

### 8.20 `clockify completion`

```bash
clockify completion install --shell bash
clockify completion install --shell zsh
clockify completion install --shell fish
clockify completion install --shell powershell
clockify completion bash              # Output to stdout
clockify completion zsh               # Output to stdout
clockify completion refresh           # Refresh cached completions
```

---

### 8.21 `clockify update`

```bash
clockify update                         # Update to latest version
clockify update --check                 # Check only, don't install
clockify update --version 1.2.3         # Install specific version
```

---

### 8.22 `clockify version`

```bash
clockify version                        # Show version info
clockify version --output json          # Machine-readable version
```

---

## Appendix A: Common Workflows

### Daily Standup Query

```bash
# What did I work on today?
clockify time-entry list --start today --output json | \
  jq -r '.[] | "\(.timeInterval.start[11:16]) - \(.timeInterval.end[11:16] // "now") | \(.description // "No description")"'
```

### Weekly Timesheet Export

```bash
# Export this week's time as CSV
clockify report detailed --start this-week --end today --export csv --file timesheet.csv
```

### Billable Hours This Month

```bash
# Quick billable summary
clockify report summary --start this-month --end today --billable --output json | \
  jq '{total_hours: ([.[].timeentries[].duration] | add / 3600), total_amount: ([.[].timeentries[] | select(.billable) | .amount] | add)}'
```

### Project Setup Automation

```bash
#!/bin/bash
set -euo pipefail

# Create project with client
PROJECT=$(clockify project create \
  --name "New Client Website" \
  --client "$(clockify client list --search "New Client" --output json | jq -r '.[0].id')" \
  --billable \
  --color "#3498db" \
  --output json)

PROJECT_ID=$(echo "$PROJECT" | jq -r '.id')

# Create tasks
clockify task create --project "$PROJECT_ID" --name "Design Phase"
clockify task create --project "$PROJECT_ID" --name "Development"
clockify task create --project "$PROJECT_ID" --name "Testing"
clockify task create --project "$PROJECT_ID" --name "Deployment"

echo "Project setup complete: $PROJECT_ID"
```

### Multi-Workspace Time Report

```bash
# Get time across all workspaces
clockify list time-entries --all-workspaces --start this-month --end today --output json | \
  jq 'group_by(.workspaceId) | map({workspace: .[0].workspaceId, total: (map(.duration) | add / 3600)})'
```

### Bulk Time Entry Creation

```bash
# Create entries from a CSV
while IFS=, read -r date project duration desc; do
  clockify time-entry create \
    --project "$project" \
    --start "${date}T09:00:00Z" \
    --duration "$duration" \
    --description "$desc" \
    --quiet
done < entries.csv
```

---

## Appendix B: Configuration Precedence

```
┌─────────────────────────────────────────┐
│  1. CLI Flags (--api-key, --workspace)  │  Highest priority
├─────────────────────────────────────────┤
│  2. Environment Variables               │
│     CLOCKIFY_API_KEY, CLOCKIFY_WORKSPACE│
├─────────────────────────────────────────┤
│  3. Profile Config                      │
│     ~/.clockify/config.yaml [profile]   │
├─────────────────────────────────────────┤
│  4. Default Config                      │
│     ~/.clockify/config.yaml [default]   │
├─────────────────────────────────────────┤
│  5. Built-in Defaults                   │  Lowest priority
│     Region: global, Output: table       │
└─────────────────────────────────────────┘
```

---

## Appendix C: Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication error (invalid/expired key) |
| 3 | Resource not found |
| 4 | Permission denied |
| 5 | Rate limited (retry later) |
| 6 | Network/connection error |
| 7 | Configuration error |
| 8 | Resource conflict (already exists) |
| 9 | Server error (Clockify 5xx) |
| 10 | Input validation error |
| 11 | CLI internal error (bug) |
| 12 | Update required (version incompatible) |

---

## Appendix D: File Structure Summary

```
~/.clockify/
├── config.yaml              # Configuration file
├── cache/
│   ├── workspaces.json      # Cached workspace list
│   ├── projects.json        # Cached project list
│   ├── tags.json            # Cached tag list
│   ├── clients.json         # Cached client list
│   └── users.json           # Cached user list
└── plugins/                 # Future: installed plugins
    └── ...
```
