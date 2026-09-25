# clockify-cli — Project Plan

**Version:** 1.0.0
**Date:** 2026-09-21
**Status:** M2 Active — Active Development

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

**Scope:** List and get operations for the core entity types. This gives users visibility into their Clockify data.

| Deliverable | Description |
|-------------|-------------|
| `clockify workspace list` | List all workspaces for authenticated user |
| `clockify workspace get` | Get workspace by ID or name |
| `clockify user get` | Current user profile |
| `clockify user list` | List workspace users with status filter |
| `clockify user info` | Search by email/name |
| `clockify project list` | List projects with filters (status, client, search) |
| `clockify project get` | Get project by ID or name |
| `clockify client list` | List clients with search |
| `clockify client get` | Get client by ID or name |
| `clockify tag list` | List tags with search |
| Pagination support | `--page`, `--page-size`, `--no-pagination` flags working across all list commands |

**Success Criteria:**
- All list commands return correctly formatted output in all formats
- Filters work correctly (e.g., `--status ACTIVE`, `--search "web"`)
- Pagination auto-fetches all pages with `--no-pagination`
- Name resolution works: `clockify project get --name "Website"` finds the right project
- Fuzzy matching warns on multiple matches and suggests `--id`

**Dependencies:** M0 (HTTP client, output formatting), M1 (auth context).

**Estimated Effort:** 2–3 weeks.

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

---

### M4: Write Operations — Project, Client, Tag CRUD

**Scope:** Create, update, delete for projects, clients, and tags. Includes destructive operation confirmation and JSON input.

| Deliverable | Description |
|-------------|-------------|
| `clockify project create` | With `--name`, `--client`, `--billable`, `--color`, `--estimate` flags |
| `clockify project update` | PATCH semantics, partial updates |
| `clockify project delete` | With confirmation, `--confirm` to skip |
| `clockify client create` | With `--name`, `--email`, `--address`, `--phone`, `--note` |
| `clockify client update` | Partial updates |
| `clockify client delete` | With confirmation |
| `clockify tag create` | With `--name`, `--color` |
| `clockify tag update` | Rename, recolor |
| `clockify tag delete` | With confirmation |
| JSON file/stdin input | `--file` and `--stdin` for all create/update commands |
| Idempotent create | `--if-not-exists` flag for create commands |

**Success Criteria:**
- Create project with client reference works end-to-end
- Delete shows confirmation prompt; `--confirm` bypasses it
- `echo '{"name":"X"}' | clockify project create --stdin` works
- `--if-not-exists` returns existing resource without error if it already exists
- All outputs are machine-readable in JSON mode

**Dependencies:** M2 (read operations exist), M0 (error handling, confirmation prompts).

**Estimated Effort:** 2–3 weeks.

**Confidence:** High (85%). CRUD patterns are well-established after M2.

---

### M5: Task Management

**Scope:** Task CRUD nested under projects.

| Deliverable | Description |
|-------------|-------------|
| `clockify task list` | List tasks for a project, `--is-active` filter |
| `clockify task create` | With `--project`, `--name`, `--estimate`, `--status`, `--assignee` |
| `clockify task update` | Update name, status, estimate |
| `clockify task delete` | With confirmation |

**Success Criteria:**
- `clockify task list --project X` returns tasks for that project
- Create task with estimate and assignee works
- Task status can be set to TODO, IN_PROGRESS, DONE

**Dependencies:** M4 (project CRUD), M2 (user list for assignee).

**Estimated Effort:** 1 week.

**Confidence:** High (90%). Straightforward CRUD following established patterns.

---

### M6: Reports & Analytics

**Scope:** Summary, detailed, and weekly reports with export capabilities.

| Deliverable | Description |
|-------------|-------------|
| `clockify report summary` | Summary by project/user with date range, billable filter |
| `clockify report detailed` | Detailed entries with export to CSV |
| `clockify report weekly` | Weekly summary starting from a given date |
| Report filters | `--project`, `--task`, `--user`, `--group`, `--client`, `--status` |
| Export formats | `--export csv`, `--export pdf` (if API supports) |
| Date range shortcuts | `this-week`, `last-week`, `this-month`, `last-month` |

**Success Criteria:**
- `clockify report summary --start this-month --end today` returns correct totals
- `clockify report detailed --export csv --file report.csv` produces valid CSV
- Weekly report shows daily breakdown
- Billable-only filter works correctly

**Dependencies:** M3 (time entries), M2 (project/user data).

**Estimated Effort:** 2–3 weeks.

**Confidence:** Medium (75%). Report API endpoints may have different patterns than standard CRUD; needs API exploration.

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

**Estimated Effort:** 4–5 weeks (8 resources × ~0.5–1 week each).

**Confidence:** Medium (70%). Some resources (approval, schedule) may have undocumented API behaviors.

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

**Estimated Effort:** 2–3 weeks.

**Confidence:** High (85%). Cobra has built-in completion support; caching is straightforward.

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

**Estimated Effort:** 2–3 weeks.

**Confidence:** High (85%). Standard Go build and release patterns.

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

**Estimated Effort:** 2–3 weeks (ongoing throughout, but dedicated sprint at end).

**Confidence:** Medium-High (80%). Testing Clockify API integration requires test account setup.

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

| After Milestone | What Users Can Do |
|-----------------|-------------------|
| M0 | Build and run CLI, see help |
| M1 | Authenticate, configure workspace |
| M2 | List and search projects, clients, tags, users |
| M3 | **Daily driver**: track time, start/stop timers, list entries |
| M4 | Create and manage projects, clients, tags |
| M5 | Manage tasks within projects |
| M6 | Generate time reports and export |
| M7 | Full Clockify API access |
| M8 | Shell completions, CI/CD integration |
| M9 | Install via package managers |
| M10 | Production-quality, well-documented CLI |

---

## 4. Risk Assessment

### 4.1 Technical Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **Clockify API undocumented behaviors** | High — blocks feature implementation | Medium | Prototype API calls early in M0; maintain a test workspace; document discovered behaviors in ADRs |
| **Timer state management** | Medium — start/stop race conditions | Medium | Use `--in-progress` flag to check state before start/stop; handle 409 conflicts gracefully |
| **OS keychain availability** | Medium — headless/CI environments fail | Low | Env var fallback is built-in; test on all platforms in M1 |
| **Duration parsing ambiguity** | Low — user confusion | Low | Strict parsing with clear error messages; document supported formats |
| **Report API uses different base URL** | Medium — requires separate client | Medium | Detect in M6 planning; may need reports-specific HTTP client configuration |
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

### 5.2 Open Questions

| Question | Impact | Resolution Approach |
|----------|--------|---------------------|
| Does Clockify API support batch/bulk endpoints natively? | M3 (bulk time entries) | Test in M0 spike; if not, implement client-side batching |
| What are the exact report API endpoints and request formats? | M6 | Prototype in M2/M3; may need separate reports client |
| Is there a Clockify API versioning scheme? | All milestones | Check API docs; pin to v1 |
| Are there rate limit headers beyond 50/s? | M0 (retry logic) | Test empirically; document discovered limits |
| Does the plugin system have a design spec? | M8 (future) | Defer to post-M10; not in current scope |

### 5.3 Out of Scope (Current Roadmap)

- Plugin system (deferred to post-M10)
- PDF report export (API-dependent, may not be supported)
- Interactive `--interactive` mode for create commands (nice-to-have, not in MVP)
- Graphical output (charts, etc.)

---

## 6. Appendix: Estimated Timeline Summary

| Milestone | Weeks | Cumulative | Status | Deliverable to Users |
|-----------|-------|------------|--------|---------------------|
| M0: Foundation | 3–4 | 3–4 | ✅ **DONE** (1 day) | Buildable CLI skeleton |
| M1: Auth & Config | 1–2 | 5–6 | ✅ **DONE** (2 days) | Authenticatable CLI with multi-profile |
| M2: Read Operations | 2–3 | 7–9 | 🔵 **ACTIVE** | Can query Clockify data |
| M3: Time Entry | 3–4 | 10–13 | ⬜ Pending | **Daily time tracking** |
| M4: Write Operations | 2–3 | 12–16 | ⬜ Pending | Can create/manage resources |
| M5: Tasks | 1 | 13–17 | ⬜ Pending | Task management |
| M6: Reports | 2–3 | 15–20 | ⬜ Pending | Time reports & export |
| M7: Secondary Resources | 4–5 | 19–25 | ⬜ Pending | Full API coverage |
| M8: Shell & Power | 2–3 | 21–28 | ⬜ Pending | Shell integration, CI/CD |
| M9: Distribution | 2–3 | 23–31 | ⬜ Pending | Package manager installs |
| M10: Quality & Polish | 2–3 | 25–34 | ⬜ Pending | Production-ready CLI |

**Total estimated effort:** 25–34 weeks (6–8 months) for 1 developer; 18–24 weeks (4.5–6 months) for 2 developers.

**First usable release (M3 complete):** Weeks 10–13 — users can authenticate, track time, and generate basic reports. This is the minimum viable product.

**M0 Actual:** Completed in ~1 day vs 3–4 week estimate. 34 files, 2612 lines of Go. All success criteria met.

**M1 Actual:** Completed in ~2 days vs 1–2 week estimate. Multi-profile support, cache system, status validation, 37 unit tests. All success criteria met.

**M2 Sprint 1 Actual:** Completed in ~1 day. Entity types defined, typed API methods, workspace list/get commands, user get/list/info commands, 21 new tests. All success criteria met.

---

*This plan should be reviewed after M1 completion (decision point) to validate assumptions, adjust estimates, and refine the backlog based on real API discoveries.*

*Last updated: 2026-09-25 — M2 Sprint 1 complete (workspace/user commands, entity types, typed API methods). Sprint 2 pending (project/client/tag + pagination).*`
