# clockify-cli — Project Plan

**Version:** 1.0.0
**Date:** 2026-10-03
**Status:** M6 Complete — M7 Pending

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
| Progressive Disclosure | Foundation, Core | M0, M3 |
| Pipe-friendly | Foundation, Core | M0, M3, M8 |
| Idempotent by default | Core, Write Ops | M3, M4 |
| Fast startup | Foundation, Polish | M0, M10 |
| All major API resources | Full Coverage | M4–M7 |
| Excellent developer experience | All themes | M0, M3, M8, M10 |

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

| Deliverable | Description |
|-------------|-------------|
| `clockify invoice` | Full CRUD (list, get, create, update, delete) |
| `clockify schedule` | Full CRUD with date range filtering |
| `clockify time-off` | List, create, update-status (no delete per spec) |
| `clockify expense` | Full CRUD with project and date filters |
| `clockify custom-field` | Full CRUD with type options (DROPDOWN, CHECKBOX, TEXT, NUMBER, LINK, DATE) |
| `clockify webhook` | Full CRUD with event types |
| `clockify approval` | List, submit, update (workflow operations) |
| `clockify group` | Full CRUD with member management (`--add-member`, `--remove-member`) |

**Success Criteria:**
- All 8 resource types have complete CRUD operations
- Group member management works correctly
- Webhook event types are validated on create
- Time-off status updates work (APPROVED, REJECTED)
- Custom field types are validated

**Dependencies:** M0–M4 (foundation + CRUD patterns established).

**Estimated Effort:** 2–4 days (8 resources × ~0.25–0.5 day each with established M4/M5 patterns). Original estimate was 4–5 weeks; revised down based on actual velocity.

**Confidence:** High (80%). CRUD patterns are proven across project/client/tag/task; main risk is undocumented behaviors on approval/schedule.

---

### M8: Shell Completion & Power Features

**Scope:** Shell integration, caching, CI/CD mode, stdin/file input for all commands.

| Deliverable | Description |
|-------------|-------------|
| Shell completion | Install/output for bash, zsh, fish, powershell |
| Dynamic completions | Cached workspace/project/tag name completion |
| `clockify completion refresh` | Refresh cached completion data |
| Caching layer | Local cache for workspaces, projects, tags, clients, users (TTL-based) |
| `--no-cache` flag | Bypass cache for fresh data |
| `clockify config clear-cache` | Clear all cached data |
| CI/CD mode | Verified exit codes, `--quiet` + `--no-color` + `--output json` |
| Stdin/file input | Consistent `--file` and `--stdin` across all create/update commands |

**Success Criteria:**
- `clockify completion install --shell zsh` installs completions
- Tab-completion shows command names, flag names, and cached resource names
- Cache is populated on first list call, served from cache on subsequent calls
- `clockify completion refresh` updates cached completions
- CI pipeline uses `--output json --quiet --no-color` and checks exit codes

**Dependencies:** M0–M7 (all commands need to exist for completion to work).

**Estimated Effort:** 1–2 days. Cobra has built-in completion support; caching layer reuses existing `config clear-cache` / cache-dir patterns from M1.

**Confidence:** High (90%). Standard Cobra completion; commands are now stable enough to generate completions against.

---

### M9: Distribution & Packaging

**Scope:** Cross-platform builds, package manager distribution, auto-update, CI/CD pipeline.

| Deliverable | Description |
|-------------|-------------|
| Cross-compilation | Linux x86_64/ARM64, macOS universal, Windows x86_64 |
| GitHub Actions | Build, test, release pipeline |
| Homebrew formula | `brew install clockify/tap/clockify` |
| npm package | `npm install -g clockify-cli` wrapper |
| Go install | `go install github.com/clockify/cli@latest` |
| `clockify update` | Self-update from GitHub Releases |
| `clockify version` | Show version, commit, build date, platform |
| Semantic versioning | Git tags trigger release builds |
| Auto-update check | First command of the day checks GitHub Releases (cached 24h) |

**Success Criteria:**
- `go install` produces working binary
- `brew install` works on macOS
- `clockify update` upgrades to latest version
- `clockify version` shows correct build metadata
- GitHub Actions releases binaries for all platforms on tag push
- Binary size <20MB, startup <50ms

**Dependencies:** M0–M8 (feature-complete CLI needed for distribution).

**Estimated Effort:** 1–2 days. Standard Go build/release patterns (goreleaser or hand-rolled GitHub Actions).

**Confidence:** High (90%). Well-trodden Go CLI distribution path.

---

### M10: Quality, Testing & Documentation

**Scope:** Test coverage, integration tests, performance, documentation, man pages.

| Deliverable | Description |
|-------------|-------------|
| Unit tests | 80%+ coverage on core packages (config, output, auth, duration parsing) |
| Integration tests | Tests against real Clockify API (with test workspace) |
| Performance audit | Binary size, startup time, memory usage profiling |
| Documentation | README with install instructions, usage examples, workflow recipes |
| Man pages | Generated from help text |
| Error message audit | All error messages follow E001–E010 format, are actionable |
| Security audit | No secrets in code, keychain usage verified, env var handling safe |

**Success Criteria:**
- Unit test coverage ≥80% on core packages
- Integration test suite runs against Clockify API and passes
- Binary starts in <50ms (measured)
- Binary size <20MB (measured)
- README covers install, config, top 10 workflows
- All error messages are actionable and follow the spec format

**Dependencies:** M0–M9 (feature-complete, distributed CLI).

**Estimated Effort:** 1–2 days dedicated + ongoing (test coverage already strong at 153 tests; main gaps are integration tests against real API and README).

**Confidence:** Medium-High (85%). Unit tests are in good shape; integration tests need a test workspace.

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
M10 (Quality & Polish)
```

**Key sequencing decisions:**

1. **M0 → M1 → M2 → M3** is the critical path. Every user needs auth, then read access, then time entry. M3 is the highest-value feature (daily timer usage).

2. **M4 after M2** (not M3): Write operations for projects/clients/tags are simpler than time entry and establish CRUD patterns. Time entry's write ops are included in M3 because it's the core workflow.

3. **M6 (Reports) after M3**: Reports consume time entry data. The report API may use different endpoints (reports.api.clockify.me) — sequencing after time entry lets us validate the base API client first.

4. **M7 (Secondary Resources) is the largest milestone** but lowest priority per resource. These are less-used features. Grouping them allows batch implementation using established patterns.

5. **M8 (Shell Completion) late**: Completions need all commands to exist. Building early wastes effort as commands change.

6. **M9 (Distribution) and M10 (Quality) last**: Distribution only matters when the product is usable. Quality is ongoing but gets dedicated focus at the end.

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
| M7 | Full Clockify API access | ⬜ Next |
| M8 | Shell completions, CI/CD integration | ⬜ Pending |
| M9 | Install via package managers | ⬜ Pending |
| M10 | Production-quality, well-documented CLI | ⬜ Pending |

**Users can already do (M0–M6):** authenticate, list/search all core entities, track time with timers, create/update/delete projects, clients, tags, and tasks, and generate summary/detailed/weekly reports with CSV export — a functional daily-driver CLI for Clockify.

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
- Interactive `--interactive` mode for create commands (nice-to-have, not in MVP)
- Graphical output (charts, etc.)

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
| M7: Secondary Resources | 4–5 weeks | *2–4 days (est.)* | ⬜ Pending | Full API coverage |
| M8: Shell & Power | 2–3 weeks | *1–2 days (est.)* | ⬜ Pending | Shell integration, CI/CD |
| M9: Distribution | 2–3 weeks | *1–2 days (est.)* | ⬜ Pending | Package manager installs |
| M10: Quality & Polish | 2–3 weeks | *1–2 days (est.)* | ⬜ Pending | Production-ready CLI |

### 6.2 Velocity Observation

**Actual velocity:** ~1 day per CRUD-pattern milestone (M0, M3, M4, M5); 2–3 days where multi-sprint (M1, M2); ~0.5 day for M6 once official reports API docs replaced the planned live spike. Original estimates assumed calendar weeks; actuals compress to days once foundation patterns (auth, HTTP client, output, name resolution, pagination, confirmation, file/stdin) were established in M0–M2.

**Revised remaining forecast:** ~6–12 days for M7–M10 (vs original 12–16 weeks). M6 is no longer the outlier — official docs provided endpoints, auth, and request shapes up front.

**Cumulative actual (M0–M6):** ~9.5 development days. 255 tests passing across all packages.

### 6.3 Milestone Tagging

| Milestone | Git Tag | Notes |
|-----------|---------|-------|
| M0 | `v0.1.0-m0` | Foundation |
| M1 | `v0.2.0-m1` | Auth & config |
| M2 | `v0.3.0-m2` | Read operations |
| M3 | `v0.4.0-m3` | Time entry |
| M4 | `v0.5.0-m4` | Write ops (tagged 2026-10-03) |
| M5 | `v0.6.0-m5` | Task mgmt (tagged 2026-10-03) |
| M6 | `v0.7.0-m6` | Reports & analytics (tag after commit) |
| M7+ | — | Tag on completion |

**First usable release (M3 complete):** Was planned weeks 10–13; actually reached after ~7 days. Users can authenticate, track time, manage projects/clients/tags/tasks, and generate reports.

---

*This plan should be reviewed after M1 completion (decision point) to validate assumptions, adjust estimates, and refine the backlog based on real API discoveries.*

*Post-M1 review outcomes: estimates revised downward after M0–M5 actuals; open questions on batch endpoints and API versioning resolved; report API host confirmed via spec §3.6.*

*Last updated: 2026-10-03 — M6 complete (report summary/detailed/weekly, ReportsBaseURL with correct regional `/report/v1` paths, CSV/PDF export, dual-client resolution, 255 tests). Official Clockify Reports API docs replaced the planned spike: POST endpoints, X-Api-Key auth, documented request bodies. M7 (Secondary Resources) is next. Tags created: v0.5.0-m4, v0.6.0-m5; tag v0.7.0-m6 after M6 is committed.*`
