# clockify-cli

A production-ready command-line interface for [Clockify](https://clockify.me) time tracking.

Track time by **name** (not just UUIDs), manage projects/clients/tags/tasks, run reports, and script everything with stable JSON output and exit codes.

```bash
clockify in --project "Website Redesign" --task QA --tags billable,frontend
clockify out --description "standup + planning"
clockify report today
```

---

## Install

### Go

```bash
go install github.com/kim/clockify-cli/cmd/clockify@latest
```

### Binary (GitHub Releases)

Download from [Releases](https://github.com/kim/clockify-cli/releases), extract, put `clockify` on your `PATH`.

Then:

```bash
clockify update          # self-update later (requires a published GitHub Release)
clockify update --check  # check only
```

> Live `clockify update` needs a public GitHub repo + first Release tag. Until then install via `go install` or a local `make build`.

### Homebrew

```bash
brew tap kim/tap && brew install clockify
```

> Tap publish is operational: requires the `kim/homebrew-tap` repo + first GitHub Release. Until then use `go install` or binaries. GoReleaser skips brew publish when `HOMEBREW_TAP_GITHUB_TOKEN` is unset.

### npm

```bash
npm install -g clockify-cli
# or scoped, if the unscoped name is taken:
npm install -g @kim/clockify-cli
```

> npm postinstall downloads a release binary and verifies `checksums.txt` (sha256). It needs a published GitHub Release; until then use `go install`.

### Build from source

```bash
git clone https://github.com/kim/clockify-cli.git
cd clockify-cli
make build
./clockify version
```

---

## Quickstart

### 1. Configure

```bash
clockify configure
```

Stores your API key in the OS keychain (Secret Service on Linux), validates it against the API, and lets you pick a workspace/region.

```bash
clockify status          # profile, masked key, workspace, region
clockify config list     # all config values
```

API key without config file:

```bash
export CLOCKIFY_API_KEY=ck_live_...
clockify status
```

### 2. See your data

```bash
clockify project list
clockify project get --name "Website"
clockify tag list
clockify task list --project "Website"
```

### 3. Track time (daily driver)

```bash
# Start a timer by names — no UUIDs required
clockify in --project "Website Redesign" --task "QA" --tags "billable,frontend"

# Or human times
clockify time-entry create --start 9:00 --end 10:30 --project Website --description "standup"

# Stop the running timer
clockify out
clockify time-entry list --start today --end today
```

Interactive pickers (TTY only — never hang in pipes/CI):

```bash
clockify in -i
# or
clockify config set interactive true
clockify in
```

### 4. Fix up entries

```bash
clockify time-entry get current
clockify time-entry clone last
clockify time-entry split <id> --at 12:00
clockify time-entry update current --description "corrected"
clockify time-entry delete last          # confirms; use --confirm in scripts
```

---

## Top workflows

| # | Goal | Command |
|---|------|---------|
| 1 | Start timer by name | `clockify in --project X --task Y` |
| 2 | Stop timer | `clockify out` |
| 3 | Log a past block | `clockify time-entry create --start 9:00 --end 10:30 --project X` |
| 4 | Today’s entries | `clockify report today` or `clockify time-entry list --start today --end today` |
| 5 | This week | `clockify report this-week` |
| 6 | Export CSV | `clockify report detailed --start this-week --export csv --file week.csv` |
| 7 | List projects | `clockify lp` (alias for `project list`) |
| 8 | Create project idempotently | `clockify project create --name X --if-not-exists` |
| 9 | Mark task done | `clockify task done QA --project X` |
| 10 | Machine-readable IDs | `clockify project list --output ids` |
| 11 | Template output | `clockify project list --output template --template '{{.name}}'` |
| 12 | CI status check | `clockify status --output json --quiet --no-color`; check exit code |

### Relative dates & human times

Accepted on list/report/create/update date flags:

- `today`, `yesterday`, `this-week`, `last-week`, `this-month`, `last-month`
- `YYYY-MM-DD`
- `9:00`, `9:00am`, `yesterday 09:00`, `now`, `-15m`, `+2h`
- Durations: `3600`, `1h`, `1.5h`, `30m`, `1h30m`

Bare dates and `HH:MM` parse in **local** timezone (or config `timezone`).

### Aliases

| Alias | Command |
|-------|---------|
| `in` | `time-entry start` |
| `out` | `time-entry stop` |
| `le` | `time-entry list` |
| `ge` | `time-entry get` |
| `ce` | `time-entry create` |
| `lp` | `project list` |
| `lw` | `workspace list` |
| `cfg` | `config set` |

---

## Configuration

Config file: `clockify config path` (XDG-aware, mode `0600`).

**Precedence:** flags > environment > config file > defaults

| Key | Env | Notes |
|-----|-----|-------|
| `api_key` | `CLOCKIFY_API_KEY` | Prefer keychain via `configure` |
| `workspace_id` | `CLOCKIFY_WORKSPACE_ID` | |
| `region` | `CLOCKIFY_REGION` | `global`, `eu-de`, `us`, `uk`, `au` |
| `output` | `CLOCKIFY_OUTPUT` | `table`, `json`, `csv`, `yaml`, `template`, `ids` |
| `auto_update` | — | `false` disables passive update checks |
| `interactive` | — | `true` enables opt-in pickers on TTY |
| `timezone` | — | IANA name; empty = OS local |
| `week_starts` | — | used for week-relative dates |
| cache TTL | `CLOCKIFY_CACHE_TTL` | default `20m` |

### Profiles

```bash
clockify configure --profile work
clockify --profile work status
clockify config list
clockify config delete --profile work
```

### Output formats

```bash
clockify project list --output json
clockify project list --output csv
clockify project list --output yaml
clockify project list --output ids
clockify project list --output template --template '{{.name}}: {{.id}}'
```

When stdout is not a TTY and `--output` is unset, output defaults to **JSON** (pipe-friendly).

---

## CI / scripting

```bash
#!/usr/bin/env bash
set -euo pipefail

# Non-interactive: never prompt
export CI=1
export CLOCKIFY_NO_UPDATE_CHECK=1

out=$(clockify time-entry list --start today --end today --output json --quiet --no-color)
echo "$out" | jq length
```

### Exit codes

| Code | Category | Typical cause |
|------|----------|---------------|
| 0 | Success | |
| 2 | E001 Auth | bad/missing API key |
| 3 | E002 Not found | unknown resource |
| 4 | E003 Permission | insufficient workspace role (HTTP 403) |
| 5 | E005 Rate limit | 429 |
| 6 | E006 Network | offline/timeout |
| 7 | E007 Config | bad config values |
| 8 | E008 Conflict | timer running / duplicate |
| 9 | E009 Server | API 5xx |
| 10 | E004 Validation | bad flags/input |
| 11 | E010 Internal | CLI bug |

There is no exit code 12. 401 → E001 (exit 2); 403 → E003 (exit 4).

Full error reference: [docs/errors/](docs/errors/). Errors print as:

```text
ERROR [E001]: API request failed with status 401

  <detail>

  Suggested fix: <action>

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E001-auth.md
```

Destroy operations prompt unless `--confirm` is set. Prefer `--confirm` in automation.

---

## Shell completion

```bash
clockify completion install --shell bash   # or zsh|fish|powershell
clockify completion refresh                # warm resource-name cache
```

Tab completion is **cache-only** (never blocks on the network). Run list commands or `completion refresh` to populate names. TTL default 20 minutes (`CLOCKIFY_CACHE_TTL`).

```bash
clockify config clear-cache
```

---

## Performance

Measured on linux/amd64, stripped build (budget: &lt;20MB, &lt;50ms):

| Metric | Measured (2026-10-05) |
|--------|------------------------|
| Binary size | **~11.5MB** (`make build`) |
| Startup (`clockify version`) | **~0–50ms** warm |
| Release archives (GoReleaser) | ~4MB compressed per platform |

```bash
make perf
```

---

## Manual pages

```bash
make man
# pages under man/ ; markdown under docs/cli/
```

Install `man/clockify.1` (and children) onto your `MANPATH`, e.g.:

```bash
mkdir -p ~/.local/share/man/man1
cp man/clockify.1 ~/.local/share/man/man1/
man clockify
```

---

## Development

```bash
make test          # unit tests (no network)
make cover         # coverage summary
make lint          # golangci-lint or go vet
make build         # ldflags-injected binary
make man           # regenerate man + docs/cli
make check         # lint + test
```

### Coverage (core packages)

| Package | Role |
|---------|------|
| `internal/config`, `internal/output`, `internal/errors`, `internal/update`, `internal/resolve`, `internal/api`, `pkg/duration`, `pkg/clockify` | Core — target ≥80% |
| `internal/cli` | Command layer — measured separately |

### Integration tests (optional, live API)

Requires credentials + a Clockify workspace you can write to:

```bash
export CLOCKIFY_API_KEY=ck_live_...
export CLOCKIFY_WORKSPACE_ID=...   # optional
go test -tags integration ./integration/... -v
```

Without env vars the suite **skips**. Default `go test ./...` never hits the network.

### Security notes

- API keys are stored in the OS keychain when available; config file is written `0600`
- `clockify status` masks keys; `--debug` prints `api-key=****` only
- Prefer `CLOCKIFY_API_KEY` only in ephemeral CI secrets, not shell history
- Never pass `--api-key` on shared machines; use profiles/keychain

---

## Full command surface

```text
clockify
├── configure / status / config / version / update / completion
├── workspace, user, project, client, tag, task
├── time-entry (list get create update delete start stop bulk clone split)
├── report (summary detailed weekly + presets today|yesterday|this-week|…)
├── invoice, schedule, time-off, expense, custom-field, webhook, approval, group
└── aliases: in out le ge ce lp lw cfg
```

```bash
clockify --help
clockify time-entry --help
```

Project plan and milestone status: [PLAN.md](PLAN.md).

---

## License

MIT
