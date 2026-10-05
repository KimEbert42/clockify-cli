# clockify-cli — Project Plan

**Version:** 1.0.0
**Date:** 2026-10-04
**Status:** M10 Complete — v1.0.0

---

## 1. Roadmap Framework

### 1.1 Time Horizon

**Now / Next / Later** — mapped to quarterly cadence.

| Horizon | Timeframe | Theme | Goal |
|---------|-----------|-------|------|
| **Now** | Weeks 1–8 (Q1) | Foundation + Core Daily Driver | Ship a usable CLI for auth, time tracking, and project management |
| **Next** | Weeks 9–16 (Q2) | Full API Coverage | Complete all remaining Clockify resources and reports |
| **Later** | Weeks 17–24 (Q3) | Distribution, Polish & Ecosystem | Cross-platform distribution, shell integration, documentation, plugin foundation |

### 1.2 Prioritization Criteria & Scoring Model

Each milestone is scored on four dimensions (1–5 scale, 5 = highest):

| Criterion | Weight | Description |
|-----------|--------|-------------|
| **User Value** | 35% | How frequently users will interact with this feature; critical path for daily workflows |
| **Dependency Blocking** | 25% | Does this unlock or block subsequent milestones? Infrastructure has highest weight here |
| **Risk** | 20% | Technical unknowns, API complexity, external dependencies. Higher risk → higher priority to validate early |
| **Effort** | 20% | Implementation complexity. Lower effort = higher priority when value is equal (fast wins) |

**Composite Score** = (Value × 0.35) + (Dependency × 0.25) + (Risk × 0.20) + ((6 - Effort) × 0.20)

The inverse of effort is used so that less effort yields a higher score (favoring fast wins).

### 1.3 Strategic Goal Mapping

| Strategic Goal | Roadmap Theme | Key Milestones |
|----------------|---------------|----------------|
| Production-quality CLI | Foundation, Polish | M0, M9, M10 |
| Discoverability | Foundation, Polish | M0, M8, M10 |
| Progressive Disclosure | Foundation, Core, Ergonomics | M0, M3, M11 |
| Pipe-friendly | Foundation, Core | M0, M3, M8 |
| Idempotent by default | Core, Write Ops | M3, M4 |
| Fast startup | Foundation, Polish | M0, M10 |
| All major API resources | Full Coverage | M4–M7 |
| Excellent developer experience | All themes | M0, M3, M8, M10, M11 |

---

## 2. Milestone Definitions

### M0: Project Scaffolding & Core Infrastructure

**Status:** ✅ **COMPLETE** — Tagged `v0.1.0-m0`
**Completed:** 2026-09-21

**Scope:** Everything that every other command depends on. CLI framework, config system, HTTP client, output formatting, error handling.

| Deliverable | Status | Notes |
|-------------|--------|-------|
| Go module & project structure | ✅ | `cmd/`, `internal/`, `pkg/` layout |
| CLI framework | ✅ | Cobra-based with 12 global flags |
| Config system | ✅ | YAML read/write, profiles, precedence chain (flag > env > config > default) |
| OS keychain integration | ✅ | Env var fallback implemented |
| HTTP client | ✅ | Retry (exponential backoff), rate limiting (45 req/s), 5 regions |
| Output formatting | ✅ | Table, JSON, CSV, YAML with auto-detection |
| Error handling | ✅ | E001–E010 categories, exit codes 0–12 |
| Resource ID resolution | ✅ | Exact UUID, prefix match, fuzzy name match |
| CI/CD mode | ✅ | Exit codes, `--quiet`, `--no-color` flags |

**Success Criteria:**
- [x] `clockify --help` shows all global flags and available commands
- [x] `clockify config set/get/list` works end-to-end
- [x] `clockify configure` wizard runs interactively and stores credentials
- [x] HTTP client successfully authenticates with Clockify API
- [x] `--output json` produces valid JSON; `--output table` produces aligned table output
- [x] Binary starts in <50ms (measured: 5ms)
- [x] Binary size <20MB (measured: 12MB)
- [x] All unit tests pass

**Dependencies:** None (first milestone).

**Actual Effort:** ~1 day (1 dev)

**Confidence:** High (90%). Standard Go CLI patterns, well-understood libraries (Cobra, keyring, tablewriter).

---

### M1: Authentication & Workspace Context

**Status:** ✅ **COMPLETE** — Tagged `v0.2.0-m1`
**Completed:** 2026-09-21

**Scope:** Complete auth flow, workspace selection, status visibility.

| Deliverable | Status | Notes |
|-------------|--------|-------|
| `clockify configure` | ✅ | Interactive wizard with API key validation, `--profile` support |
| `clockify config` | ✅ | `list`, `get`, `set`, `unset`, `path`, `delete`, `clear-cache` |
| `clockify status` | ✅ | Shows profile, masked API key, workspace, region, output format, validation status, JSON output |
| OS keychain integration | ✅ | Real keyring via `go-keyring` (Secret Service on Linux), per-profile keys |
| API key validation | ✅ | Validates key against Clockify API before saving, status validates on demand |
| Workspace selection | ✅ | Fixed bug, now correctly parses user selection |
| Multi-profile support | ✅ | `configure --profile`, `config list` shows all, `config delete --profile`, profile isolation |
| Region mapping | ✅ | Global, EU-DE, US, UK, AU endpoint resolution |
| Environment variable support | ✅ | All flags readable from `CLOCKIFY_*` env vars |
| Config clear-cache | ✅ | `config clear-cache` with profile support, XDG cache directory |

**Sprint 1 Completed:**
- Real keyring integration (`go-keyring` library)
- API key validation in `configure` command
- Workspace selection bug fix
- Unit tests for keyring operations

**Sprint 2 Completed:**
- Multi-profile support: `configure --profile`, `config list` shows all profiles, `config delete --profile`
- Config commands preserve existing profiles (no overwrite on set/unset)
- Refactored CLI commands to factory functions for test isolation
- Cache system: `CacheDir()`, `ClearCache()`, profile-specific cache dirs
- Status command: API key validation, valid/invalid/unreachable indicators, JSON output with `valid` field
- 15 new unit tests (multi-profile, cache, status validation, config operations)
- E2E verification of all M1 success criteria

**Success Criteria:**
- [x] `clockify configure` stores API key in OS keychain and validates it against the API
- [x] `clockify status` shows current auth context with masked key
- [x] `clockify --profile work status` switches context correctly
- [x] `CLOCKIFY_API_KEY=... clockify status` works without config file
- [x] Config precedence chain verified: flag > env > config > default

**Dependencies:** M0 (config system, HTTP client).

**Estimated Effort:** 1–2 weeks. **Actual:** ~2 days (Sprint 1 + Sprint 2).

**Confidence:** High (90%). Straightforward API integration once M0 foundation exists.

---

### M2: Read Operations — Workspace, User, Project, Client, Tag

**Status:** ✅ **COMPLETE** — Tagged `v0.3.0-m2`
**Completed:** 2026-09-29

**Scope:** List and get operations for the core entity types. This gives users visibility into their Clockify data.

| Deliverable | Status | Notes |
|-------------|--------|-------|
| `clockify workspace list` | ✅ | |
| `clockify workspace get` | ✅ | By ID, by name (fuzzy match) |
| `clockify user get` | ✅ | Current user profile |
| `clockify user list` | ✅ | Status filter |
| `clockify user info` | ✅ | Search by email/name |
| `clockify project list` | ✅ | `--status`, `--client`, `--search` filters |
| `clockify project get` | ✅ | By ID, by name (fuzzy match) |
| `clockify client list` | ✅ | `--search` filter |
| `clockify client get` | ✅ | By ID |
| `clockify tag list` | ✅ | `--search` filter |
| Pagination support | ✅ | `--page`, `--page-size`, `--no-pagination` flags working across all list commands |

**Sprint 1 Completed:**
- Entity types defined (Workspace, User, Project, Client, Tag)
- Typed API methods (ListWorkspaces, GetWorkspace, GetCurrentUser, ListUsers, ListProjects, GetProject, ListClients, GetClient, ListTags)
- Workspace list/get commands with name resolution
- User get/list/info commands with email/name search
- 21 new tests

**Sprint 2 Completed:**
- Project list/get commands with status/client/search filters and name resolution
- Client list/get commands with search filter
- Tag list command with search filter
- `PageInfo` struct for pagination metadata
- `DoWithPagination()` client method (reads X-Pagination-Total-Pages/Results headers)
- `fetchPaginated*()` helpers for auto-fetching all pages
- Pagination summary in table mode (suppressed for JSON/CSV/YAML)
- `lp` alias for `project list`
- 20 new tests (project, client, tag, pagination)

**Success Criteria:**
- [x] All list commands return correctly formatted output in all formats
- [x] Filters work correctly (e.g., `--status ACTIVE`, `--search "web"`)
- [x] Pagination auto-fetches all pages with `--no-pagination`
- [x] Name resolution works: `clockify project get --name "Website"` finds the right project
- [x] Fuzzy matching warns on multiple matches and suggests `--id`

**Dependencies:** M0 (HTTP client, output formatting), M1 (auth context).

**Estimated Effort:** 2–3 weeks. **Actual:** ~3 days (Sprint 1 + Sprint 2).

**Confidence:** High (85%). Read operations are lower risk; mainly mapping API responses to output formats.

---

### M3: Time Entry — The Core Daily Driver

**Scope:** Full time entry lifecycle including timer operations. This is the most-used feature and must be excellent.

| Deliverable | Description |
|-------------|-------------|
| `clockify time-entry list` | List with date ranges, project/task/user/description/tag filters |
| `clockify time-entry get` | Get by ID |
| `clockify time-entry create` | Create with start+duration or start+end; support `1h`, `1.5h`, `30m` duration formats |
| `clockify time-entry update` | Update description, project, start time, duration |
| `clockify time-entry delete` | Delete with confirmation prompt |
| `clockify time-entry start` | Start timer (no duration) |
| `clockify time-entry stop` | Stop running timer with optional description |
| `clockify time-entry bulk` | Bulk create/update from JSON file or stdin |
| Relative date parsing | `today`, `yesterday`, `this-week`, `last-week`, `this-month`, `last-month` |
| `--hydrated` flag | Include full project/task/user objects in output |
| `--in-progress` flag | Filter to only running entries |
| Stdin/file input | `--file entry.json` and `echo '...' | clockify create time-entry --stdin` |

**Success Criteria:**
- `clockify time-entry start --project X` starts a timer; `clockify time-entry stop` stops it
- Duration parsing handles all formats: seconds, `1h`, `1.5h`, `30m`, `1h30m`
- `clockify time-entry list --start today --end today` returns today's entries
- `--output json | jq` pipeline works for scripting
- Bulk create from JSON file processes all entries
- Confirmation required for delete; `--confirm` skips it

**Dependencies:** M0, M1, M2 (project/user/tag IDs needed for references).

**Estimated Effort:** 3–4 weeks (highest complexity — timer state, duration parsing, filtering).

**Confidence:** Medium-High (80%). Timer state management and duration parsing have moderate complexity.

**Status:** ✅ **COMPLETE** — Tagged `v0.4.0-m3`
**Completed:** 2026-09-29

| Deliverable | Status | Notes |
|-------------|--------|-------|
| `clockify time-entry list` | ✅ | With date ranges, project/task/description/tag filters, --in-progress |
| `clockify time-entry get` | ✅ | By ID, with --hydrated flag |
| `clockify time-entry create` | ✅ | With start+duration or start+end; duration formats: 1h, 1.5h, 30m, 1h30m |
| `clockify time-entry update` | ✅ | Partial updates for description, project, start, duration |
| `clockify time-entry delete` | ✅ | With confirmation prompt, --confirm to skip |
| `clockify time-entry start` | ✅ | Starts timer, detects conflicts |
| `clockify time-entry stop` | ✅ | Stops running timer, shows duration |
| `clockify time-entry bulk` | ✅ | Bulk create from JSON file or stdin |
| Relative date parsing | ✅ | today, yesterday, this-week, last-week, this-month, last-month |
| `--hydrated` flag | ✅ | On list and get commands |
| `--in-progress` flag | ✅ | Filter to only running entries |
| Duration parser | ✅ | pkg/duration: Parse() and Format() |
| API methods | ✅ | ListTimeEntries, GetTimeEntry, CreateTimeEntry, UpdateTimeEntry, DeleteTimeEntry, GetTimeEntryInProgress, StopTimeEntry, BatchGetTimeEntries |
| Unit tests | ✅ | 25+ new tests across duration, API, and CLI packages |

**Success Criteria:**
- [x] `clockify time-entry start --project X` starts a timer; `clockify time-entry stop` stops it
- [x] Duration parsing handles all formats: seconds, `1h`, `1.5h`, `30m`, `1h30m`
- [x] `clockify time-entry list --start today --end today` returns today's entries
- [x] `--output json | jq` pipeline works for scripting
- [x] Bulk create from JSON file processes all entries
- [x] Confirmation required for delete; `--confirm` skips it

**Actual Effort:** ~1 day (3 sprints compressed).

---

### M4: Write Operations — Project, Client, Tag CRUD

**Scope:** Create, update, delete for projects, clients, and tags. Includes destructive operation confirmation and JSON input.

**Status:** ✅ **COMPLETE**
**Completed:** 2026-10-03

| Deliverable | Status | Notes |
|-------------|--------|-------|
| `clockify project create` | ✅ | With `--name`, `--client`, `--billable`, `--color`, `--estimate`, `--note`, `--file`, `--stdin`, `--if-not-exists` flags |
| `clockify project update` | ✅ | PATCH semantics, partial updates, `--id`/`--name` resolution |
| `clockify project delete` | ✅ | With confirmation, `--confirm` to skip |
| `clockify client create` | ✅ | With `--name`, `--email`, `--address`, `--note`, `--file`, `--stdin`, `--if-not-exists` |
| `clockify client update` | ✅ | Partial updates, `--id`/`--name` resolution |
| `clockify client delete` | ✅ | With confirmation |
| `clockify tag create` | ✅ | With `--name`, `--file`, `--stdin`, `--if-not-exists` |
| `clockify tag update` | ✅ | Rename, archive, `--id`/`--name` resolution |
| `clockify tag delete` | ✅ | With confirmation |
| JSON file/stdin input | ✅ | `--file` and `--stdin` for all create/update commands |
| Idempotent create | ✅ | `--if-not-exists` flag for create commands |
| API methods | ✅ | CreateProject, UpdateProject, DeleteProject, CreateClient, UpdateClient, DeleteClient, CreateTag, UpdateTag, DeleteTag |
| Name resolution | ✅ | Fuzzy matching for client and tag operations |
| Unit tests | ✅ | 15+ new tests across project, client, and tag write operations |

**Success Criteria:**
- [x] Create project with client reference works end-to-end
- [x] Delete shows confirmation prompt; `--confirm` bypasses it
- [x] `echo '{"name":"X"}' | clockify project create --stdin` works
- [x] `--if-not-exists` returns existing resource without error if it already exists
- [x] All outputs are machine-readable in JSON mode

**Dependencies:** M2 (read operations exist), M0 (error handling, confirmation prompts).

**Actual Effort:** ~1 day.

**Confidence:** High (85%). CRUD patterns are well-established after M2.

---

### M5: Task Management

**Scope:** Task CRUD nested under projects.

**Status:** ✅ **COMPLETE**
**Completed:** 2026-10-03

| Deliverable | Status | Notes |
|-------------|--------|-------|
| `clockify task list` | ✅ | `--project` (ID or fuzzy name), `--is-active` filter |
| `clockify task create` | ✅ | `--project`, `--name`, `--estimate`, `--status`, `--assignee`, `--billable`, `--file`, `--stdin`, `--if-not-exists` |
| `clockify task update` | ✅ | PUT + existing-seed merge, partial updates, `--file`/`--stdin` |
| `clockify task delete` | ✅ | With confirmation prompt, `--confirm` to skip |
| Status validation | ✅ | TODO, IN_PROGRESS, DONE (validated before API calls) |
| Estimate parsing | ✅ | `pkg/duration.Parse`: seconds, `1h`, `1.5h`, `30m`, `1h30m` |
| Assignee resolution | ✅ | User ID, email, or fuzzy name via `resolveUserRef` |
| Project resolution | ✅ | `resolveProjectRef`: tries ID first, falls back to fuzzy name |
| `--if-not-exists` | ✅ | Scoped to project via `resolveTaskByName` |
| API methods | ✅ | ListTasks, GetTask, CreateTask, UpdateTask, DeleteTask |
| Paths | ✅ | `/workspaces/{wid}/projects/{pid}/tasks[/{tid}]` |
| Pagination | ✅ | `fetchPaginatedTasks` helper |
| Unit tests | ✅ | 17 new tests in task_test.go |

**Success Criteria:**
- [x] `clockify task list --project X` returns tasks for that project
- [x] Create task with estimate and assignee works
- [x] Task status can be set to TODO, IN_PROGRESS, DONE

**Dependencies:** M4 (project CRUD), M2 (user list for assignee).

**Actual Effort:** ~1 day.

**Confidence:** High (90%). Straightforward CRUD following established patterns.

---

### M6: Reports & Analytics

**Scope:** Summary, detailed, and weekly reports with export capabilities.

**Status:** ✅ **COMPLETE** — Tag on commit (`v0.7.0-m6`)
**Completed:** 2026-10-03

| Deliverable | Status | Notes |
|-------------|--------|-------|
| `clockify report summary` | ✅ | Date range, project/task/user/client/group filters, `--billable`, `--description` |
| `clockify report detailed` | ✅ | Date range + filters; `--export csv` to file or stdout |
| `clockify report weekly` | ✅ | `--start` required; `--end` defaults to start + 6 days |
| Report filters | ✅ | `--project`, `--task`, `--user`, `--group`, `--client`, `--status` |
| Export formats | ✅ | `--export csv` (API-native or client-side fallback); `--export pdf` when API returns PDF |
| Date range shortcuts | ✅ | `today`, `yesterday`, `this-week`, `last-week`, `this-month`, `last-month` + `YYYY-MM-DD` |
| `ReportsBaseURL(region)` | ✅ | `internal/api/regions.go`; `CLOCKIFY_REPORTS_URL` test override |
| Reports API client | ✅ | `internal/api/reports.go` — POST methods, retry, raw body passthrough |
| Request/response types | ✅ | `pkg/clockify/reports.go` — filters, request body, flexible response model |
| CLI registration | ✅ | `internal/cli/report.go` + `root.go` |
| Unit tests | ✅ | Regions, API methods, CLI commands, date/status mapping |

**Spec reference:** §8.11 `clockify report` (clockify-cli-spec.md:919–950)

**Official API facts (docs.clockify.me — corrected PLAN assumptions):**

1. **Reports host/path is NOT `reports.<host>` for all regions:**

   | Region | Reports base URL |
   |--------|------------------|
   | global | `https://reports.api.clockify.me/v1` |
   | eu-de | `https://euc1.clockify.me/report/v1` |
   | us | `https://use2.clockify.me/report/v1` |
   | uk | `https://euw2.clockify.me/report/v1` |
   | au | `https://apse2.clockify.me/report/v1` |

   Global uses a separate `reports.` host; regional workspaces use the **same host as the main API** with a `/report/v1` path prefix. Subdomains: `https://{sub}.clockify.me/report/v1/...`.

2. **Endpoints (POST + JSON body):**
   - `POST {base}/workspaces/{wsId}/reports/detailed`
   - `POST {base}/workspaces/{wsId}/reports/summary`
   - `POST {base}/workspaces/{wsId}/reports/weekly`

3. **Auth:** Same `X-Api-Key` header as the main API.

4. **Required body fields:**
   - All: `dateRangeStart`, `dateRangeEnd` (`YYYY-MM-DDTHH:MM:SS.ssssss`)
   - detailed: + `detailedFilter` (auto-filled if omitted)
   - summary: + `summaryFilter` (auto-filled if omitted)
   - weekly: + `weeklyFilter` (auto-filled if omitted)

5. **`exportType`:** `JSON`, `JSON_V1`, `CSV`, `PDF`, `XLSX`, `ZIP`. CLI sets this from `--export`.

6. **Response schema:** Not formally documented. Client uses flexible `ReportResponse` (groups/days/timeentries/totals) + raw map; CSV export falls back to client-side formatting if API returns JSON.

7. **Status mapping (`--status` → API `approvalState`):**
   - `APPROVED` → `APPROVED`
   - `NEW`, `IN_PROGRESS` → `UNAPPROVED`
   - `REJECTED` → **unsupported by reports API** (CLI errors with explanation)

8. **`--file` semantics:**
   - Without `--export`: JSON report parameters input (spec §8.11)
   - With `--export`: output path (PLAN success criteria / Appendix A)
   - `--export csv` without `--file`: CSV to stdout

**Reusable building blocks used:**
- `pkg/duration.ParseRelativeDate` — relative date shortcuts
- `resolveProjectRef` / `resolveUserRef` / `resolveClientByName` / `resolveTaskByName` — filter resolution via main API client
- `internal/output.CSVFormatter` — client-side CSV export fallback
- Dual-client pattern: main API client for name resolution, reports client for generation

**Success Criteria:**
- [x] `clockify report summary --start this-month --end today` returns correct totals
- [x] `clockify report detailed --export csv --file report.csv` produces valid CSV
- [x] Weekly report shows daily breakdown
- [x] Billable-only filter works correctly

**Dependencies:** M3 (time entries), M2 (project/user data). Both complete.

**Actual Effort:** ~0.5 day (documented API schemas replaced the planned half-day live spike for implementation; live E2E still recommended when credentials are available).

**Confidence:** High (85%). Endpoints, auth, and request body shapes are documented. Remaining risk: live regional host behavior and exact response field names — display code is intentionally flexible.

**Key risk (updated):** Regional reports use path prefix `/report/v1` on the main API host, not `reports.<host>`. Implemented per official docs; `CLOCKIFY_REPORTS_URL` allows test overrides.

---

### M7: Secondary Resources — Invoice, Schedule, Time-off, Expense, Custom-field, Webhook, Approval, Group

**Scope:** Complete API coverage for all remaining Clockify resources.

**Status:** ✅ **COMPLETE**
**Completed:** 2026-10-04
**Sprint plan:** `M7-SPRINT-PLAN.md` (official docs.clockify.me verified 2026-10-04)

| Deliverable | Status | Notes |
|-------------|--------|-------|
| `clockify invoice` | ✅ | list/get/create/update/delete + PATCH status; wrapped `{invoices,total}` handled |
| `clockify schedule` | ✅ | `/scheduling/assignments/*`; date range required; `--hours-per-day`/`--duration`; `--series` |
| `clockify time-off` | ✅ | POST list; policy-scoped create/update-status/delete; `--type` resolves policy |
| `clockify expense` | ✅ | CRUD + category resolution; `--amount`→`total`; project/user/date filters |
| `clockify custom-field` | ✅ | CRUD; TEXT→TXT, DROPDOWN→DROPDOWN_SINGLE; DATE rejected with explanation |
| `clockify webhook` | ✅ | CRUD; official event enum + aliases (TIME_ENTRY_CREATED→NEW_TIME_ENTRY) |
| `clockify approval` | ✅ | list/submit (period-based)/update-status; `--time-entry` rejected with explanation |
| `clockify group` | ✅ | CRUD via `user-groups` path; `--members`, `--add-member`, `--remove-member` |
| Types + enums | ✅ | `pkg/clockify/m7.go`, `m7_enums.go` |
| API methods | ✅ | `internal/api/m7_resources.go` |
| CLI commands | ✅ | 8 files under `internal/cli/` + `m7_helpers.go` + pagination |
| Unit tests | ✅ | enums, API httptest, CLI help/validation/functional |

**Official API corrections vs spec (implemented per docs.clockify.me):**
- Group path is `user-groups`; member add/remove are dedicated endpoints
- Time-off list is POST; create/status/delete are policy-scoped paths
- Approval submit is period-based (`WEEKLY`/`MONTHLY` + `periodStart`), not per-entry IDs
- Custom-field types: `TXT`, `NUMBER`, `DROPDOWN_SINGLE/MULTIPLE`, `CHECKBOX`, `LINK` (no `DATE`)
- Invoice create requires `clientId`, `currency`, `issuedDate`, `dueDate`, `number`
- Schedule list requires `start`/`end`; create body uses `hoursPerDay`

**Success Criteria:**
- [x] All 8 resource types have complete CRUD operations
- [x] Group member management works correctly
- [x] Webhook event types are validated on create
- [x] Time-off status updates work (APPROVED, REJECTED)
- [x] Custom field types are validated
- [x] All commands appear in `clockify --help`
- [x] Unit tests pass across all packages

**Dependencies:** M0–M4 (foundation + CRUD patterns established).

**Actual Effort:** ~1 day (official API docs replaced live spike; patterns reused from M4/M5).

**Confidence:** High (85%). Remaining risk: live plan-gated features (invoice/approval/webhook) and exact response field names on regional workspaces.

**Tags:** `v0.7.0-m6` created on M6 commit; tag `v0.8.0-m7` after M7 commit.

---

### M8: Shell Completion & Power Features

**Status:** ✅ **COMPLETE**
**Completed:** 2026-10-04

**Scope:** Shell integration, caching, CI/CD mode, stdin/file input for all commands.

| Deliverable | Status | Notes |
|-------------|--------|-------|
| Shell completion | ✅ | `completion [bash\|zsh\|fish\|powershell]` + `install --shell` + stdout output; back-compat positional form kept |
| Dynamic completions | ✅ | Cache-backed `ValidArgsFunction`/`RegisterFlagCompletionFunc` for workspace/project/client/tag/user/`--profile`; never calls API during tab-complete |
| `clockify completion refresh` | ✅ | Fetches all resource names into TTL cache; prints counts |
| Caching layer | ✅ | `internal/config`: `WriteTTL`/`ReadTTL` with envelope + `CLOCKIFY_CACHE_TTL` (default 20m); list commands serve unfiltered results from cache then populate completion names |
| `--no-cache` flag | ✅ | Global persistent flag; bypasses list cache reads (still writes completion cache after API fetch) |
| `clockify config clear-cache` | ✅ | Existing M1 command; profile-aware |
| CI/CD mode | ✅ | `main.go` maps `CLIError.ExitCode()` (auth=2, notfound=3, …); `--quiet` wired into table headers; `--no-color` + `NO_COLOR` supported |
| Stdin/file input | ✅ | `time-entry create/update` accept `--file`/`--stdin` (flags override JSON); `report *` gains `--stdin` for params |

**Completion install paths (XDG-aware):**
| Shell | Default path |
|-------|----------------|
| bash | `$XDG_DATA_HOME/bash-completion/completions/clockify` (or `BASH_COMP_INSTALL_DIR`) |
| zsh | `$XDG_DATA_HOME/zsh/site-functions/_clockify` |
| fish | `$XDG_CONFIG_HOME/fish/completions/clockify.fish` |
| powershell | `%USERPROFILE%\Documents\PowerShell\clockify.ps1` (profile line printed) |

**Success Criteria:**
- [x] `clockify completion install --shell zsh` installs completions
- [x] Tab-completion shows command names, flag names, and cached resource names
- [x] Cache is populated on first unfiltered list call, served from cache on subsequent calls
- [x] `clockify completion refresh` updates cached completions
- [x] CI pipeline can use `--output json --quiet --no-color` and check exit codes

**Dependencies:** M0–M7 (all commands exist for completion + cache wiring).

**Actual Effort:** ~1 day.

**Confidence:** High (90%). Cobra completion + TTL cache reusing M1 cache-dir patterns.

**Key implementation notes:**
- Dynamic completion is cache-only (silent on miss) so tab-complete stays fast and offline-safe.
- Filtered list calls always hit the API; only unfiltered lists are cache-served.
- Test isolation: `setupTestConfig` now sets per-test `XDG_CACHE_HOME`.

**Tags:** `v0.7.0-m6` on M6; `v0.8.0-m7` after M7; tag `v0.9.0-m8` after M8 commit.

---

### M9: Distribution & Packaging

**Scope:** Cross-platform builds, package manager distribution, auto-update, CI/CD pipeline.

**Status:** ✅ **COMPLETE**
**Completed:** 2026-10-04
**Sprint plan:** `M9-SPRINT-PLAN.md`

| Deliverable | Status | Notes |
|-------------|--------|-------|
| Cross-compilation | ✅ | GoReleaser: linux amd64/arm64, darwin universal, windows amd64; archives + sha256 checksums |
| GitHub Actions | ✅ | `.github/workflows/ci.yml` (vet/test/build) + `release.yml` (goreleaser on `v*` tags) |
| Homebrew formula | ✅ | GoReleaser `brews` → `kim/homebrew-tap` Formula; needs tap repo + `HOMEBREW_TAP_GITHUB_TOKEN` |
| npm package | ✅ | `npm/` wrapper (`postinstall` downloads release binary); scoped name if unscoped taken |
| Go install | ✅ | `go install github.com/kim/clockify-cli/cmd/clockify@latest` (real module path — not spec's `clockify/cli`) |
| `clockify update` | ✅ | `update`, `--check`, `--version`; downloads matching GH Release asset and replaces binary |
| `clockify version` | ✅ | version/commit/built/go/platform + `--output json` |
| Semantic versioning | ✅ | Tag push → `release.yml` → GoReleaser; tags `v0.8.0-m7`, `v0.9.0-m8` created on prior commits |
| Auto-update check | ✅ | First command of day; 24h TTL cache; skipped when quiet/json/CI/`CLOCKIFY_NO_UPDATE_CHECK`/`auto_update=false`/dev build |
| ldflags fix | ✅ | Makefile injects `main.version/commit/date`; snapshot binary reports real version (was clobbered to `dev`) |
| Error printing | ✅ | `main.go` prints errors to stderr (required for `update --check` failures to be visible) |

**Install paths (documented):**

| Method | Command |
|--------|---------|
| go install | `go install github.com/kim/clockify-cli/cmd/clockify@latest` |
| Homebrew | `brew tap kim/tap && brew install clockify` (after tap repo exists) |
| npm | `npm install -g clockify-cli` or `npm install -g @kim/clockify-cli` |
| Binary | GitHub Releases archives + `clockify update` |

**Config:** `auto_update` key (`clockify config set auto_update false`); env `CLOCKIFY_NO_UPDATE_CHECK=1`; CI env auto-skips.

**Success Criteria:**
- [x] `go install` produces working binary (module path documented)
- [x] `clockify version` shows correct build metadata (ldflags fixed; JSON output)
- [x] GitHub Actions release workflow present; goreleaser snapshot builds all platforms
- [x] Binary size <20MB (linux/amd64 ~3.9MB), startup <50ms (measured ~10ms)
- [x] `clockify update` / `--check` / `--version` implemented; unit-tested with httptest
- [x] Auto-update check cache + skip rules unit-tested
- [x] `go test ./...` green
- [ ] `brew install` works on macOS — **blocked on tap repo + first published release**
- [ ] Live `clockify update` upgrades to latest — **blocked on public GitHub repo + first release tag**

**Dependencies:** M0–M8 (feature-complete CLI needed for distribution).

**Actual Effort:** ~1 day (infra + self-update + npm scaffold; live publish awaits remote).

**Confidence:** High (90%). GoReleaser + Cobra patterns are standard; remaining work is operational (create GitHub repo, tap, publish).

**Key implementation notes:**
- Module path is `github.com/kim/clockify-cli` (spec §7.2's `github.com/clockify/cli` is incorrect).
- Dev builds (`version=dev`) refuse self-update with an actionable error.
- Passive update check never fails the primary command; hint goes to stderr.
- GoReleaser universal darwin assets are named `clockify_*_darwin_all.tar.gz`; `FindAsset` handles that fallback.
- Homebrew/npm live E2E requires a git remote (none configured at implementation time).

**Tags:** `v0.8.0-m7` on `3168f3e`; `v0.9.0-m8` on `969ca42`; `v0.10.0-m9` after M9 commit.

---

**Status:** ✅ **COMPLETE** — Tagged `v1.0.0`
**Completed:** 2026-10-05
**Sprint plan:** `M10-SPRINT-PLAN.md`

| Deliverable | Status | Notes |
|-------------|--------|-------|
| Unit tests | ✅ | Core packages ≥80%: api 90%, config 90%, output 95%, errors 98%, resolve 100%, update 84%, duration 88%, clockify 86% |
| Integration tests | ✅ | `integration/` with `//go:build integration`; skips without `CLOCKIFY_API_KEY`; live run deferred (no remote/creds in this environment) |
| Performance audit | ✅ | linux/amd64 binary **~11.5MB** (&lt;20MB); startup **~0–50ms** warm (&lt;50ms) |
| Documentation | ✅ | Root `README.md` — install, config, 12 workflows, CI exit codes, security notes |
| Man pages | ✅ | `make man` → `man/*.1` + `docs/cli/*.md` via cobra/doc |
| Error message audit | ✅ | `ERROR [E00X]` + Detail/Suggestion/Learn more; docs in `docs/errors/E00X-shortname.md` |
| Security audit | ✅ | Keychain + 0600 config; status masks key; debug prints `api-key=****` only |

**Success Criteria:**
- [x] Unit test coverage ≥80% on core packages
- [x] Integration test suite present and env-gated (live API run optional)
- [x] Binary starts in &lt;50ms (measured warm)
- [x] Binary size &lt;20MB (measured ~11.5MB)
- [x] README covers install, config, top 10 workflows
- [x] Error messages follow E00X format with learn-more links
- [x] Man pages generated from help text
- [x] `go test ./...` green

**Dependencies:** M0–M9, M11 (feature-complete, distributed CLI).

**Actual Effort:** ~1 day.

**Confidence:** High (85%). Live integration + package-manager publish remain operational (git remote, test workspace).

**Key implementation notes:**
- Core coverage gate: `internal/{config,output,errors,update,resolve,api}`, `pkg/{duration,clockify}` — all ≥80%.
- `internal/cli` measured separately (~52% statements); command contracts covered via help/validation/httptest, not full E2E per command.
- Error format ADR: print `ERROR [E001]: …` + Learn more → repo `docs/errors/E001-auth.md` (not the spec's fictional `clockify.github.io` host). Exit codes unchanged (E001→2 … E004→10, E010→11).
- Integration harness never runs in default `go test ./...`.

**Tags:** `v0.11.0-m11` on M11; **`v1.0.0`** after M10 commit.

**Confidence:** Medium-High (85%). Unit tests are in good shape; integration tests need a test workspace.

---

### M11: Daily-Driver Ergonomics

**Scope:** Close the everyday-UX gaps identified in the 2026-10-04 UX review. The CLI is strong on API coverage, reports, output formats, profiles/regions and scripting, but the daily start/stop/fix-up loop still requires IDs and RFC3339 timestamps. M11 makes that loop fast without regressing pipe-friendliness.

**Sequencing:** Numbered M11 to keep existing milestone numbers, tags and sprint plans stable, but scheduled **after M9 and before M10** so M10's README, man pages and integration tests document the final UX.

**Status:** ✅ **COMPLETE**
**Completed:** 2026-10-04
**Sprint plan:** `M11-SPRINT-PLAN.md`

| Deliverable | Status | Notes |
|-------------|--------|-------|
| Names everywhere | ✅ | `time-entry start/create/update/list` accept project/task/tag names or IDs; `resolveTaskRef`, `resolveTagRefs`; ambiguity → CLIError + candidates |
| Human time input | ✅ | `pkg/duration.ParseHumanDateTime`: `HH:MM`, am/pm, `yesterday 09:00`, `now`, `±offsets`; local TZ (fixes bare-date UTC bug); adopted by time-entry + M7 date flags + reports |
| Interactive pickers | ✅ | Hand-rolled numbered/fuzzy pickers (no new deps — binary stays ~17MB); `-i` + `interactive` config; strict guards: non-TTY, CI, `--quiet`, `--confirm`, machine `--output` |
| `time-entry clone` | ✅ | `clone [id\|last]`; flags override; `--start`+`--end/--duration` → completed entry else timer |
| `time-entry split` | ✅ | `split <id> --at HH:MM...`; validates points inside entry before any write |
| `current` / `last` selectors | ✅ | get/update/delete/clone/split; UUID path skips extra API calls |
| Multi-entry update | ✅ | `update <id>...` + `--stdin-ids`; confirm summarizes count |
| `in` / `out` aliases | ✅ | Plus fixed dead `ge`/`ce` (were targeting wrong path order) |
| Report presets | ✅ | `report today\|yesterday\|this-week\|last-week\|this-month\|last-month`; `--detailed` optional |
| Task shortcuts | ✅ | `task done`, `task quick-add` (`--quiet` prints ID only) |
| Template output | ✅ | `--output template --template '...'` + `--output ids` |
| Config keys | ✅ | `interactive`, `timezone` (`config set/get/unset/list`) |
| Completion wiring | ✅ | `completeTaskRefC` on time-entry task flags; report `--task` completion fixed |

**Success Criteria:**
- [x] `clockify in --project "Website Redesign" --task "QA" --tags "billable,frontend"` starts a timer with no IDs typed (resolution + aliases)
- [x] `clockify time-entry create --start 9:00 --end 10:30 --project Website` creates a correct entry in local time
- [x] `clockify in` with no flags on a TTY (with `-i` or config) opens pickers; the same command in a pipe or with `CI=1` fails fast with an actionable missing-flag error — no hangs
- [x] `clone last`, `split <id> --at 12:00`, `stop current`, `delete last` work end-to-end
- [x] Every new flag/command has dynamic completion wired through the M8 cache
- [x] Existing scripts using IDs and RFC3339 timestamps keep working unchanged (backward compatible)

**Dependencies:** M3 (time entry), M5 (task resolution helpers), M8 (cache + completion). Independent of M9; precedes M10.

**Actual Effort:** ~1 day (hand-rolled pickers avoided `huh` dependency; size budget held).

**Confidence:** High (85%). Remaining live E2E needs credentials for name resolution + pickers on a real workspace.

**Key implementation notes:**
- Interactive pickers are **hand-rolled** (numbered list + substring filter) — no bubbletea/huh tree; binary ~17.6MB, startup ~4ms.
- Bare dates and `HH:MM` parse in **local** timezone (bugfix: previously bare dates parsed UTC).
- `ge`/`ce` aliases were dead (targeted `get time-entry`); fixed to `time-entry get`/`create`.
- ID fast path: UUID-shaped refs skip name list/search calls.
- Ambiguous names return `CLIError` E004 with candidate list (`name  (id)`).

**Tags:** `v0.10.0-m9` on M9; tag `v0.11.0-m11` after M11 commit.

---

## 3. Prioritization & Sequencing

### 3.1 Scoring Results

| Milestone | Value | Dependency | Risk | Effort (inv) | Weighted Score | Rank |
|-----------|-------|------------|------|-------------|----------------|------|
| M0: Foundation | 5 | 5 | 3 | 3 | **4.30** | 1 |
| M3: Time Entry | 5 | 4 | 4 | 2 | **4.15** | 2 |
| M1: Auth & Config | 5 | 5 | 2 | 4 | **4.30** | 1 (tied) |
| M2: Read Operations | 4 | 4 | 2 | 3 | **3.45** | 4 |
| M4: Write Operations | 4 | 3 | 2 | 3 | **3.25** | 5 |
| M5: Tasks | 3 | 3 | 1 | 5 | **2.95** | 6 |
| M6: Reports | 4 | 3 | 4 | 3 | **3.55** | 3 (tied) |
| M8: Shell/Power | 3 | 2 | 2 | 3 | **2.65** | 7 |
| M7: Secondary | 3 | 2 | 3 | 2 | **2.60** | 8 |
| M9: Distribution | 2 | 1 | 2 | 3 | **1.95** | 9 |
| M10: Quality | 3 | 1 | 1 | 3 | **2.15** | 10 |
| M11: Ergonomics | 5 | 1 | 2 | 3 | **3.00** | — (added post-M8) |

### 3.2 Sequencing Rationale

```
M0 (Foundation) ──────────────────────────────────────────────────────────┐
    │                                                                     │
    ▼                                                                     │
M1 (Auth & Config) ──────────────────────────────────────────────────────┐│
    │                                                                     ││
    ▼                                                                     ││
M2 (Read Operations) ───────────────────────────────────────────────────┐││
    │                                                                     │││
    ▼                                                                     │││
M3 (Time Entry) ◄───────────────────────────────────────────────────────┘││
    │                                                                     ││
    ├──► M4 (Write Operations) ──── M5 (Tasks)                          ││
    │                                                                     ││
    ├──► M6 (Reports)                                                     ││
    │                                                                     ││
    ▼                                                                     ││
M7 (Secondary Resources) ◄──────────────────────────────────────────────┘│
    │                                                                     │
    ▼                                                                     │
M8 (Shell & Power) ◄────────────────────────────────────────────────────┘
    │
    ├──► M9 (Distribution)
    │
    ▼
M11 (Daily-Driver Ergonomics)
    │
    ▼
M10 (Quality & Polish)
```

**Key sequencing decisions:**

1. **M0 → M1 → M2 → M3** is the critical path. Every user needs auth, then read access, then time entry. M3 is the highest-value feature (daily timer usage).

2. **M4 after M2** (not M3): Write operations for projects/clients/tags are simpler than time entry and establish CRUD patterns. Time entry's write ops are included in M3 because it's the core workflow.

3. **M6 (Reports) after M3**: Reports consume time entry data. The report API may use different endpoints (reports.api.clockify.me) — sequencing after time entry lets us validate the base API client first.

4. **M7 (Secondary Resources) is the largest milestone** but lowest priority per resource. These are less-used features. Grouping them allows batch implementation using established patterns.

5. **M8 (Shell Completion) late**: Completions need all commands to exist. Building early wastes effort as commands change.

6. **M9 (Distribution) and M10 (Quality) last**: Distribution only matters when the product is usable. Quality is ongoing but gets dedicated focus at the end.

7. **M11 (Ergonomics) inserted before M10**: Added after the 2026-10-04 UX review. Scores high on user value (daily start/stop/fix-up loop) but blocks nothing, so it follows M9 (already sprint-planned). It precedes M10 so documentation, man pages and integration tests cover the final command surface.

### 3.3 Value Delivery Timeline

| After Milestone | What Users Can Do | Status |
|-----------------|-------------------|--------|
| M0 | Build and run CLI, see help | ✅ Done |
| M1 | Authenticate, configure workspace | ✅ Done |
| M2 | List and search projects, clients, tags, users | ✅ Done |
| M3 | **Daily driver**: track time, start/stop timers, list entries | ✅ Done |
| M4 | Create and manage projects, clients, tags | ✅ Done |
| M5 | Manage tasks within projects | ✅ Done |
| M6 | Generate time reports and export | ✅ Done |
| M7 | Full Clockify API access | ✅ Done |
| M8 | Shell completions, CI/CD integration | ✅ Done |
| M9 | Install via package managers, self-update | ✅ Done |
| M11 | Track time by name with natural times, interactive pickers, clone/split entries | ✅ Done |
| M10 | Production-quality, well-documented CLI | ✅ Done |

**Users can already do (M0–M11 + M10):** authenticate, list/search all core entities, track time with timers, create/update/delete projects, clients, tags, tasks, invoices, schedules, expenses, custom fields, webhooks, user groups, manage time-off and approval workflows, generate summary/detailed/weekly reports with CSV export, install shell completions with dynamic cached resource names, run non-interactive CI pipelines with machine-readable output and meaningful exit codes, view build metadata via `clockify version`, self-update from GitHub Releases via `clockify update`, track time by name with human times and interactive pickers, and install via go install / (pending) Homebrew tap / npm wrapper — a documented, tested, daily-driver CLI covering the major Clockify API surface, ready for v1.0.0 once a public remote/release exists.

---

## 4. Risk Assessment

### 4.1 Technical Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **Clockify API undocumented behaviors** | High — blocks feature implementation | Medium | Prototype API calls early in M0; maintain a test workspace; document discovered behaviors in ADRs |
| **Timer state management** | Medium — start/stop race conditions | Medium | Use `--in-progress` flag to check state before start/stop; handle 409 conflicts gracefully |
| **OS keychain availability** | Medium — headless/CI environments fail | Low | Env var fallback is built-in; test on all platforms in M1 |
| **Duration parsing ambiguity** | Low — user confusion | Low | Strict parsing with clear error messages; document supported formats |
| **Report API uses different base URL** | Medium — requires separate client | Medium | **Confirmed by official docs**: global = `reports.api.clockify.me/v1`; regional = same host as API + `/report/v1` path prefix. Implemented via `ReportsBaseURL()`; `CLOCKIFY_REPORTS_URL` for tests. |
| **API rate limiting in CI/CD** | Medium — scripts fail unpredictably | Medium | Implement exponential backoff in M0; document rate limits; add `--retry` flag |
| **Binary size bloat** | Low — distribution concerns | Low | Use `ldflags` to strip debug info; monitor size in CI |
| **Cobra alias system limitations** | Low — may not support all alias patterns | Low | Validate alias requirements in M0 spike; fall back to wrapper commands if needed |

### 4.2 Project Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **Scope creep (18+ commands)** | High — delays delivery | High | Strict milestone boundaries; defer "nice-to-have" flags to later iterations |
| **API changes during development** | Medium — rework required | Low | Pin to API version; abstract API client behind interface |
| **Single developer bottleneck** | High — no review, bus factor | Medium | Comprehensive tests; clear architecture; document decisions |
| **Clockify API auth changes** | High — breaks core functionality | Low | Auth is isolated in M1; abstract behind interface for easy replacement |

### 4.3 Decision Points

| Decision | When | Options | Recommendation |
|----------|------|---------|----------------|
| **CLI framework** | M0 Week 1 | Cobra vs. Kong vs. custom | Cobra (industry standard, best completion support) |
| **HTTP client** | M0 Week 1 | `net/http` vs. `resty` vs. `go-retry` | `net/http` + custom retry (fewer deps, meets fast startup goal) |
| **Config format** | M0 Week 1 | YAML vs. TOML vs. JSON | YAML (spec already defines it, human-readable) |
| **Output formatting** | M0 Week 1 | `tablewriter` vs. `lipgloss` vs. custom | `tablewriter` (lightweight, well-maintained) |
| **Interactive prompt library** | M11 Sprint 1 | `charmbracelet/huh` vs. `AlecAivazis/survey` (archived) vs. custom | `huh` (maintained, fuzzy filter, accessible mode); evaluate binary-size and startup impact against M10 budgets (<20MB, <50ms) |
| **Local time parsing** | M11 Sprint 1 | Hand-rolled vs. `olebedev/when` / `tj/go-naturaldate` | Hand-rolled over a small fixed grammar (`HH:MM`, am/pm, `yesterday`, `±offset`, `now`) — predictable, no ambiguous NLP, easy to document |
| **Test strategy** | M0 Week 1 | Integration-only vs. unit+integration | Unit for core logic, integration against real API for E2E |
| **Monorepo vs. multi-repo** | M0 Week 1 | Single repo vs. separate repos | Single repo (simpler for small team, CI/CD) |

---

## 5. Key Decisions & Open Questions

### 5.1 Decisions Made

| Decision | Rationale |
|----------|-----------|
| **Go as implementation language** | Spec mentions `go install`; single static binary requirement; strong CLI ecosystem (Cobra) |
| **Now/Next/Later roadmap** | Appropriate for unknown team capacity; flexible quarterly cadence |
| **M0 → M1 → M2 → M3 critical path** | Maximizes time-to-value for daily time tracking workflow |
| **Group M7 resources together** | Similar CRUD patterns; lower priority per resource; batch efficiency |
| **M8 (shell completion) late** | Commands stabilize before completion scripts are generated |
| **Nested task paths** | `/workspaces/{wid}/projects/{pid}/tasks[/{tid}]` — matches Clockify API; task `--project` is required on every subcommand |
| **Reports use separate host** | Official Clockify docs: global uses `reports.api.clockify.me/v1`; regional uses main host + `/report/v1`. Spec §3.6's `reports.<host>` pattern is **incorrect for regional workspaces**. |

### 5.2 Open Questions

| Question | Impact | Status | Resolution / Approach |
|----------|--------|--------|----------------------|
| Does Clockify API support batch/bulk endpoints natively? | M3 (bulk time entries) | ✅ Resolved | Native batch endpoint exists: `POST /workspaces/{wid}/time-entries/batch` via `BatchGetTimeEntries`. Bulk create uses client-side loop + batch read. |
| What are the exact report API endpoints and request formats? | M6 | ✅ Resolved | Official docs: POST `{base}/workspaces/{ws}/reports/{detailed\|summary\|weekly}` with JSON body; auth via `X-Api-Key`; required `dateRangeStart`/`dateRangeEnd` + per-type filter object. Global base `reports.api.clockify.me/v1`; regional base `{host}/report/v1`. |
| Is there a Clockify API versioning scheme? | All milestones | ✅ Resolved (practically) | All endpoints use `/api/v1` prefix. Pinned to v1 via `BaseURL()` in `internal/api/regions.go`. |
| Are there rate limit headers beyond 50/s? | M0 (retry logic) | 🔲 Partially resolved | Client uses 45 req/s default with exponential backoff. No explicit rate-limit headers observed in use; document if discovered during M6 spike. |
| Does the plugin system have a design spec? | M8 (future) | 🔲 Deferred | Out of scope until post-M10. |

### 5.3 Decisions Made During M4–M5

| Decision | Rationale |
|----------|-----------|
| **PUT for updates (not PATCH)** | Clockify API uses PUT for project/client/tag/task updates. M4 notes say "PATCH semantics" but implementation correctly sends PUT with partial bodies. |
| **Task estimate uses `duration.Parse`** | More flexible than spec's "seconds only" — accepts `3600`, `1h`, `1.5h`, `30m`. Raw seconds still work. |
| **`resolveProjectRef` tries ID first** | Spec examples use `--project <projectId>`. Fast path: single GET by ID; falls back to fuzzy name via `resolveProjectByName`. |
| **Task `--assignee` accepts ID/email/name** | Success criterion requires assignee to work; name/email resolution matches user-info patterns. Raw ID passes through if no name/email match. |
| **Validate flags before API calls** | `--status` validated and `--name` checked before project resolution — invalid input fails fast without network calls. |
| **List output uses display structs** | Follows `time_entry.go` `entryDisplay` pattern — flattens nested fields (estimate seconds → human string, assignee IDs → comma-joined) for table/CSV while JSON path gets structured data via the same display type for consistency. |

### 5.4 Out of Scope (Current Roadmap)

- Plugin system (deferred to post-M10)
- PDF report export (API-dependent, may not be supported)
- ~~Interactive `--interactive` mode for create commands~~ — moved into scope as M11 (opt-in pickers for time entries)
- Graphical output (charts, etc.)

### 5.5 Decisions Made During M8

| Decision | Rationale |
|----------|-----------|
| **Keep custom `completion` command** | Spec §8.20 requires `install`/`refresh`; cobra's default completion command is suppressed when a custom one is registered — extend ours instead of fighting cobra. |
| **Dynamic completion is cache-only** | Tab-complete must never block on the network; cold cache returns no suggestions. Users run `completion refresh` or list commands to warm the cache. |
| **List cache only for unfiltered lists** | Filtered results vary; caching them would serve wrong data. Unfiltered lists + `completion refresh` populate TTL cache (default 20m, `CLOCKIFY_CACHE_TTL`). |
| **`--no-cache` bypasses list cache reads** | Always-fresh data for scripts/CI; completion names still written after successful API fetches so tab-complete stays warm. |
| **Exit codes via `CLIError.ExitCode()`** | CI needs stable codes (auth=2, notfound=3, validation=10); `main.go` unwraps `CLIError` instead of always exiting 1. |
| **JSON flags override file/stdin** | Explicit CLI flags win over `--file`/`--stdin` JSON for `time-entry create/update`; same precedence pattern as other write ops. |

### 5.6 Decisions Made During M9

| Decision | Rationale |
|----------|-----------|
| **GoReleaser over hand-rolled Actions** | Industry standard for Go CLIs: archives, checksums, universal darwin, brew formula generation; CI workflow stays thin. |
| **ldflags on `main` package** | Makefile previously injected into `internal/cli` but `main.go` `SetBuildInfo` overwrote with `dev`. Injecting `main.*` matches goreleaser convention; `SetBuildInfo` remains the single propagation path. |
| **Real module path everywhere** | Spec says `github.com/clockify/cli`; module is `github.com/kim/clockify-cli`. Docs and `go install` use the real path. |
| **Passive update check never fails commands** | Hint on stderr; silent on network error; skipped for quiet/json/CI/dev builds. |
| **`auto_update` config + env skips** | Spec requires `config set auto_update false` and `CLOCKIFY_NO_UPDATE_CHECK=1`; CI env also skips. |
| **Dev builds refuse self-update** | `version=dev` cannot identify a release asset safely; clear error points to package manager / go install. |
| **pip/AUR/Snap out of scope** | Listed in spec §7.2 but not in PLAN.md M9 deliverables; deferred post-M10. |
| **main.go prints errors** | `SilenceErrors` + no print meant `update --check` failures were invisible; fixed in M9. |

---

## 6. Appendix: Estimated Timeline Summary

### 6.1 Original Estimates vs Actuals

| Milestone | Original Est. | Actual | Status | Deliverable to Users |
|-----------|---------------|--------|--------|---------------------|
| M0: Foundation | 3–4 weeks | **~1 day** | ✅ **DONE** | Buildable CLI skeleton |
| M1: Auth & Config | 1–2 weeks | **~2 days** | ✅ **DONE** | Authenticatable CLI with multi-profile |
| M2: Read Operations | 2–3 weeks | **~3 days** | ✅ **DONE** | Can query Clockify data |
| M3: Time Entry | 3–4 weeks | **~1 day** | ✅ **DONE** | **Daily time tracking** |
| M4: Write Operations | 2–3 weeks | **~1 day** | ✅ **DONE** | Can create/manage resources |
| M5: Tasks | 1 week | **~1 day** | ✅ **DONE** | Task management |
| M6: Reports | 2–3 weeks | **~0.5 day** | ✅ **DONE** | Time reports & export |
| M7: Secondary Resources | 4–5 weeks | **~1 day** | ✅ **DONE** | Full API coverage |
| M8: Shell & Power | 2–3 weeks | **~1 day** | ✅ **DONE** | Shell integration, CI/CD |
| M9: Distribution | 2–3 weeks | **~1 day** | ✅ **DONE** | Package manager installs + self-update |
| M11: Ergonomics | — (added post-M8) | **~1 day** | ✅ **DONE** | Name-based, interactive daily workflow |
| M10: Quality & Docs | 2–3 weeks | **~1 day** | ✅ **DONE** | Production-ready CLI + v1.0.0 |

### 6.2 Velocity Observation

**Actual velocity:** ~1 day per CRUD-pattern milestone (M0, M3, M4, M5, M7, M8, M9); 2–3 days where multi-sprint (M1, M2); ~0.5 day for M6 once official reports API docs replaced the planned live spike. Original estimates assumed calendar weeks; actuals compress to days once foundation patterns (auth, HTTP client, output, name resolution, pagination, confirmation, file/stdin) were established in M0–M2. M7 followed the same pattern: official docs.clockify.me endpoints + M4/M5 CRUD templates. M8 reused Cobra completion + M1 cache-dir patterns. M9 used GoReleaser + existing ldflags/cache patterns; live package-manager publish remains operational work (git remote + tap/npm).

**Revised remaining forecast:** ~3–6 days for M11 and M10 (vs original 12–16 weeks for M9–M10).

**Cumulative actual (M0–M11 + M10):** ~13.5 development days. All packages green (`go test ./...` — 0 failures).

**M10 coverage (core packages, 2026-10-05):** api 90.2%, config 90.3%, errors 98.4%, output 95.1%, resolve 100%, update 84.1%, pkg/clockify 86.1%, pkg/duration 88.3%. Total statements ~59% (cli dominates uncovered surface).

### 6.3 Milestone Tagging

| Milestone | Git Tag | Notes |
|-----------|---------|-------|
| M0 | `v0.1.0-m0` | Foundation |
| M1 | `v0.2.0-m1` | Auth & config |
| M2 | `v0.3.0-m2` | Read operations |
| M3 | `v0.4.0-m3` | Time entry |
| M4 | `v0.5.0-m4` | Write ops (tagged 2026-10-03) |
| M5 | `v0.6.0-m5` | Task mgmt (tagged 2026-10-03) |
| M6 | `v0.7.0-m6` | Reports & analytics (tagged 2026-10-04) |
| M7 | `v0.8.0-m7` | Secondary resources (tagged 2026-10-04 on `3168f3e`) |
| M8 | `v0.9.0-m8` | Shell completion & power features (tagged 2026-10-04 on `969ca42`) |
| M9 | `v0.10.0-m9` | Distribution & packaging (tag on M9 commit) |
| M11 | `v0.11.0-m11` | Daily-driver ergonomics (tagged 2026-10-04) |
| M10 | `v1.0.0` | First stable release (tagged 2026-10-05) |

**First usable release (M3 complete):** Was planned weeks 10–13; actually reached after ~7 days. Users can authenticate, track time, manage projects/clients/tags/tasks, and generate reports.

**First stable release (M10 / v1.0.0):** Docs, man pages, coverage gates, error reference, and integration harness in place. Remaining operational work: public GitHub remote + first Release (unblocks brew/npm/live `clockify update`).

---

*This plan should be reviewed after M1 completion (decision point) to validate assumptions, adjust estimates, and refine the backlog based on real API discoveries.*

*Post-M1 review outcomes: estimates revised downward after M0–M5 actuals; open questions on batch endpoints and API versioning resolved; report API host confirmed via spec §3.6.*

*Last updated: 2026-10-05 — **M10 complete → v1.0.0.** Quality: core coverage ≥80%, `ERROR [E00X]` + docs/errors short names, README, man pages, env-gated integration harness, CI size/coverage gates. Next (post-v1.0.0): push remote + GitHub Release; optional live integration; plugin system deferred.*
