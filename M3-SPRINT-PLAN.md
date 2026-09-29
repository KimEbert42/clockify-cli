# M3: Time Entry — Sprint Plan

**Milestone:** M3
**Sprint Duration:** 3 × 1-week sprints (5 working days each)
**Developer:** 1 (solo)
**Start Date:** TBD
**End Date:** TBD (3 weeks from start)

---

## 1. Sprint Structure

| Sprint | Theme | Days | Goal |
|--------|-------|------|------|
| **S1** | Duration Parsing, API Client & Read Commands | D1–D5 | Duration parser, time entry API methods, list/get commands, in-progress filter |
| **S2** | Write Commands & Timer Operations | D6–D10 | Create, update, delete, start, stop commands |
| **S3** | Bulk Operations, Hydrated & E2E | D11–D15 | Bulk create, --hydrated flag, stdin/file input, relative dates, E2E verification |

---

## 2. Current State Assessment

### Already Implemented (from M0 + M1 + M2)
| Feature | Status | Notes |
|---------|--------|-------|
| `TimeEntry` struct | ✅ | `pkg/clockify/types.go` — ID, Description, ProjectID, TaskID, UserID, Billable, TimeInterval, TagIDs |
| `TimeInterval` struct | ✅ | Start, End (*time.Time), Duration (int64) |
| HTTP client with retry/rate-limit | ✅ | `api.NewClient()`, `Do()`, `DoWithPagination()` |
| Output formatters | ✅ | Table, JSON, CSV, YAML via `formatOutput()` |
| Resource ID resolution | ✅ | `resolve.ResolveInput()` — exact, prefix, fuzzy |
| Pagination infrastructure | ✅ | `fetchPaginated*()` helpers, `--page`, `--page-size`, `--no-pagination` |
| Config loading with profiles | ✅ | `loadAuthWithConfig()` returns apiKey, region, cfg |
| Aliases defined | ✅ | `"le": "time-entry list"`, `"ge": "get time-entry"`, `"ce": "create time-entry"` |

### What M3 Needs to Build
| Component | Status | Notes |
|-----------|--------|-------|
| Duration parser | ❌ | `pkg/duration/duration.go` is empty |
| Time entry API methods | ❌ | No methods in `internal/api/resources.go` |
| Time entry CLI commands | ❌ | No `time_entry.go` in `internal/cli/` |
| Relative date parser | ❌ | `today`, `yesterday`, `this-week`, etc. |
| Confirmation prompt for delete | ❌ | Uses `--confirm` global flag |
| Stdin/file input | ❌ | `--file` and `--stdin` for create/update |
| TimeEntry hydration | ❌ | API supports `?hydrated=true` for full objects |

### Clockify API Endpoints (M3 Scope)

| Operation | Method | Path | Notes |
|-----------|--------|------|-------|
| List time entries | GET | `/workspaces/{wid}/user/{uid}/time-entries` | Filters: start, end, project, task, tags, description, in-progress |
| Get time entry | GET | `/workspaces/{wid}/time-entries/{id}` | Optional `?hydrated=true` |
| Create time entry | POST | `/workspaces/{wid}/time-entries` | Body: start, end (optional for timer), projectId, description, etc. |
| Update time entry | PUT | `/workspaces/{wid}/time-entries/{id}` | Full replace semantics |
| Delete time entry | DELETE | `/workspaces/{wid}/time-entries/{id}` | Returns 204 No Content |
| Stop running timer | PATCH | `/workspaces/{wid}/user/{uid}/time-entries` | Body: `{"end": "ISO8601"}` |
| Get in-progress | GET | `/workspaces/{wid}/time-entries/status/in-progress` | Returns all running timers |
| Bulk create/update | PUT | `/workspaces/{wid}/user/{uid}/time-entries` | Array of time entry objects |
| Batch get | POST | `/workspaces/{wid}/time-entries/batch` | Body: `{hydrated, timeEntryIds}` |

---

## 3. Sprint 1: Duration Parsing, API Client & Read Commands (Days 1–5)

### User Stories

**US-1.1: Duration Parser**
> As a user, I want to specify durations in human-readable formats like `1h`, `1.5h`, `30m`, `1h30m` so that I don't have to calculate seconds.

**Acceptance Criteria:**
- `pkg/duration/duration.go` implements `Parse(s string) (int64, error)` returning seconds
- Supports: raw seconds (`3600`), hours (`1h`, `1.5h`), minutes (`30m`), combined (`1h30m`, `2h15m`)
- Returns clear error messages for invalid formats
- `Format(seconds int64) string` converts seconds back to human-readable (`1h30m`)

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.1 | Implement `Parse()` — regex-based parser for all duration formats | 2h | `pkg/duration/duration.go` | Unit tests pass |
| T1.2 | Implement `Format()` — seconds to human-readable string | 1h | `pkg/duration/duration.go` | Unit tests pass |
| T1.3 | Write comprehensive unit tests (valid formats, edge cases, errors) | 1.5h | `pkg/duration/duration_test.go` | `go test ./pkg/duration/` passes |

**US-1.2: Relative Date Parser**
> As a user, I want to use relative dates like `today`, `yesterday`, `this-week` in `--start` and `--end` flags so I don't have to type exact dates.

**Acceptance Criteria:**
- `pkg/duration/relative.go` implements `ParseRelativeDate(s string, weekStart string) (time.Time, error)`
- Supports: `today`, `yesterday`, `this-week`, `last-week`, `this-month`, `last-month`
- `weekStart` parameter handles different week start days (monday, sunday)
- Returns start-of-day for day values, start-of-week for week values, start-of-month for month values

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.4 | Implement `ParseRelativeDate()` for day/week/month values | 2h | `pkg/duration/relative.go` | Unit tests pass |
| T1.5 | Write unit tests for relative date parsing | 1h | `pkg/duration/relative_test.go` | `go test ./pkg/duration/` passes |

**US-1.3: Time Entry API Methods**
> As a developer, I want typed API methods for time entry operations so that CLI commands are clean and testable.

**Acceptance Criteria:**
- `client.ListTimeEntries(ctx, opts)` — list with filters (start, end, project, task, tags, description, in-progress)
- `client.GetTimeEntry(ctx, id)` — get by ID
- `client.GetTimeEntryInProgress(ctx)` — get running timer
- `client.BatchGetTimeEntries(ctx, ids, hydrated)` — fetch multiple entries
- All methods use `Do()` and unmarshal into typed structs
- `ListTimeEntriesOpts` struct with all filter fields

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.6 | Define `ListTimeEntriesOpts` struct with all filter fields | 0.5h | `internal/api/resources.go` | `go build ./internal/api/` |
| T1.7 | Implement `ListTimeEntries()` with query params for all filters | 2h | `internal/api/resources.go` | Unit test with httptest |
| T1.8 | Implement `GetTimeEntry()` by ID | 0.5h | `internal/api/resources.go` | Unit test with httptest |
| T1.9 | Implement `GetTimeEntryInProgress()` | 0.5h | `internal/api/resources.go` | Unit test with httptest |
| T1.10 | Implement `BatchGetTimeEntries()` with hydrated flag | 1h | `internal/api/resources.go` | Unit test with httptest |
| T1.11 | Write integration tests for all time entry API methods | 2h | `internal/api/resources_test.go` | `go test ./internal/api/` passes |

**US-1.4: Time Entry List Command**
> As a user, I want `clockify time-entry list` with filters so I can find and review my time entries.

**Acceptance Criteria:**
- `clockify time-entry list` lists recent entries (default: today)
- `--start <date>` and `--end <date>` filter by date range (supports relative dates)
- `--project <id>` filters by project
- `--task <id>` filters by task
- `--tags <tag1,tag2>` filters by tags (AND logic)
- `--description <text>` searches by description
- `--in-progress` shows only running timers
- `--page`, `--page-size`, `--no-pagination` work for pagination
- Table columns: ID, DESCRIPTION, PROJECT, DURATION, START, END, BILLABLE
- Duration displayed in human-readable format (via `duration.Format()`)

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.12 | Create `internal/cli/time_entry.go` with `newTimeEntryCmd()` factory | 0.5h | `internal/cli/time_entry.go` | `go build ./cmd/clockify/` |
| T1.13 | Implement `newTimeEntryListCmd()` with all filter flags | 3h | `internal/cli/time_entry.go` | `clockify time-entry list --start today` works |
| T1.14 | Add `fetchPaginatedTimeEntries()` helper | 1.5h | `internal/cli/pagination.go` | Auto-fetches all pages |
| T1.15 | Register time-entry command in `root.go` | 0.5h | `internal/cli/root.go` | `clockify --help` shows time-entry |
| T1.16 | Write unit tests for time-entry list command | 2h | `internal/cli/time_entry_test.go` | `go test ./internal/cli/` passes |

**US-1.5: Time Entry Get Command**
> As a user, I want `clockify time-entry get` so I can see details of a specific entry.

**Acceptance Criteria:**
- `clockify time-entry get --id <id>` gets entry by ID
- Table output shows all fields: ID, DESCRIPTION, PROJECT, TASK, USER, DURATION, START, END, BILLABLE, TAGS
- `--hydrated` flag includes full project/task/user objects in output

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.17 | Implement `newTimeEntryGetCmd()` with `--id` and `--hydrated` flags | 1.5h | `internal/cli/time_entry.go` | `clockify time-entry get --id X` works |
| T1.18 | Write unit tests for time-entry get command | 1h | `internal/cli/time_entry_test.go` | `go test ./internal/cli/` passes |

### Sprint 1 Definition of Done

- [ ] Duration parser handles all formats: seconds, `1h`, `1.5h`, `30m`, `1h30m`
- [ ] Relative date parser handles: today, yesterday, this-week, last-week, this-month, last-month
- [ ] API methods: ListTimeEntries, GetTimeEntry, GetTimeEntryInProgress, BatchGetTimeEntries
- [ ] `clockify time-entry list` works with all filters
- [ ] `clockify time-entry get --id` works with `--hydrated`
- [ ] Pagination works with `--no-pagination`
- [ ] All unit tests pass: `go test ./...`
- [ ] No lint errors: `golangci-lint run`

### Daily Checkpoints — Sprint 1

| Day | End-of-Day Verification |
|-----|------------------------|
| D1 | Duration parser + relative date parser implemented and tested |
| D2 | All time entry API methods implemented with httptest mocks |
| D3 | `clockify time-entry list` command works with basic filters |
| D4 | All list filters working; pagination helper implemented |
| D5 | `clockify time-entry get` works; all S1 tests pass |

---

## 4. Sprint 2: Write Commands & Timer Operations (Days 6–10)

### User Stories

**US-2.1: Time Entry Create Command**
> As a user, I want `clockify time-entry create` so I can log time with start+duration or start+end.

**Acceptance Criteria:**
- `clockify time-entry create --start <time> --duration <dur>` creates entry with duration
- `clockify time-entry create --start <time> --end <time>` creates entry with end time
- `--project <id>` sets the project
- `--task <id>` sets the task
- `--description <text>` sets description
- `--tags <tag1,tag2>` sets tags
- `--billable` / `--not-billable` sets billable status
- `--file <path>` reads entry data from JSON file
- `--stdin` reads entry data from stdin
- Output shows created entry in requested format

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.1 | Implement `newTimeEntryCreateCmd()` with all flags | 3h | `internal/cli/time_entry.go` | `clockify time-entry create --start X --duration 1h` works |
| T2.2 | Add `CreateTimeEntry()` API method | 1h | `internal/api/resources.go` | Unit test with httptest |
| T2.3 | Implement stdin/file input parsing (`--file`, `--stdin`) | 1.5h | `internal/cli/time_entry.go` | `echo '...' \| clockify time-entry create --stdin` works |
| T2.4 | Write unit tests for create command | 1.5h | `internal/cli/time_entry_test.go` | `go test ./internal/cli/` passes |

**US-2.2: Time Entry Update Command**
> As a user, I want `clockify time-entry update` so I can modify existing entries.

**Acceptance Criteria:**
- `clockify time-entry update --id <id>` with any of: `--description`, `--project`, `--start`, `--duration`, `--end`, `--tags`, `--billable`
- Partial updates — only provided fields are changed
- `--file <path>` and `--stdin` for JSON input
- Output shows updated entry

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.5 | Implement `newTimeEntryUpdateCmd()` with update flags | 2h | `internal/cli/time_entry.go` | `clockify time-entry update --id X --description "new"` works |
| T2.6 | Add `UpdateTimeEntry()` API method | 1h | `internal/api/resources.go` | Unit test with httptest |
| T2.7 | Write unit tests for update command | 1h | `internal/cli/time_entry_test.go` | `go test ./internal/cli/` passes |

**US-2.3: Time Entry Delete Command**
> As a user, I want `clockify time-entry delete` so I can remove entries with confirmation.

**Acceptance Criteria:**
- `clockify time-entry delete --id <id>` prompts for confirmation
- `--confirm` flag skips confirmation prompt
- Shows deleted entry info before confirming
- `--quiet` suppresses output

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.8 | Implement `newTimeEntryDeleteCmd()` with confirmation | 1.5h | `internal/cli/time_entry.go` | `clockify time-entry delete --id X` prompts |
| T2.9 | Add `DeleteTimeEntry()` API method | 0.5h | `internal/api/resources.go` | Unit test with httptest |
| T2.10 | Write unit tests for delete command (with/without --confirm) | 1h | `internal/cli/time_entry_test.go` | `go test ./internal/cli/` passes |

**US-2.4: Timer Start Command**
> As a user, I want `clockify time-entry start` so I can start a running timer.

**Acceptance Criteria:**
- `clockify time-entry start --project <id>` starts timer with project
- `--task <id>`, `--description <text>`, `--tags <tag1,tag2>`, `--billable` supported
- Reports error if timer already running (use `--confirm` to override)
- Output shows started timer entry

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.11 | Implement `newTimeEntryStartCmd()` — creates entry without end time | 2h | `internal/cli/time_entry.go` | `clockify time-entry start --project X` works |
| T2.12 | Add `StartTimeEntry()` API method (alias for create without end) | 0.5h | `internal/api/resources.go` | Unit test |
| T2.13 | Add running timer detection and conflict handling | 1h | `internal/cli/time_entry.go` | Error on double-start |
| T2.14 | Write unit tests for start command | 1h | `internal/cli/time_entry_test.go` | `go test ./internal/cli/` passes |

**US-2.5: Timer Stop Command**
> As a user, I want `clockify time-entry stop` so I can stop the running timer.

**Acceptance Criteria:**
- `clockify time-entry stop` stops the running timer
- `--description <text>` updates description when stopping
- Reports error if no timer running
- Output shows stopped entry with duration

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.15 | Implement `newTimeEntryStopCmd()` — PATCH with end time | 1.5h | `internal/cli/time_entry.go` | `clockify time-entry stop` works |
| T2.16 | Add `StopTimeEntry()` API method (PATCH with end time) | 0.5h | `internal/api/resources.go` | Unit test |
| T2.17 | Add no-timer-running detection | 0.5h | `internal/cli/time_entry.go` | Error when no timer |
| T2.18 | Write unit tests for stop command | 1h | `internal/cli/time_entry_test.go` | `go test ./internal/cli/` passes |

### Sprint 2 Definition of Done

- [ ] `clockify time-entry create` works with --start, --duration/--end, all options
- [ ] `clockify time-entry update` works with partial updates
- [ ] `clockify time-entry delete` prompts for confirmation; --confirm bypasses
- [ ] `clockify time-entry start` starts timer; detects conflicts
- [ ] `clockify time-entry stop` stops timer; detects no-timer
- [ ] Stdin/file input works for create and update
- [ ] All unit tests pass: `go test ./...`
- [ ] No lint errors: `golangci-lint run`

### Daily Checkpoints — Sprint 2

| Day | End-of-Day Verification |
|-----|------------------------|
| D6 | `clockify time-entry create` works with all flags |
| D7 | `clockify time-entry update` and `delete` work |
| D8 | `clockify time-entry start` and `stop` work with timer detection |
| D9 | Stdin/file input working; all S2 commands tested |
| D10 | All S2 tests pass; lint clean |

---

## 5. Sprint 3: Bulk Operations, Hydrated & E2E (Days 11–15)

### User Stories

**US-3.1: Bulk Create/Update Command**
> As a user, I want `clockify time-entry bulk` so I can create/update multiple entries from a JSON file.

**Acceptance Criteria:**
- `clockify time-entry bulk --file entries.json` processes all entries
- `echo '[...]' | clockify time-entry bulk --stdin` reads from stdin
- Each entry can specify: start, end/duration, project, description, tags
- Reports success/failure count
- `--output json` returns all created/updated entries

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.1 | Implement `newTimeEntryBulkCmd()` with file/stdin flags | 2h | `internal/cli/time_entry.go` | `clockify time-entry bulk --file X` works |
| T3.2 | Add `BulkUpdateTimeEntries()` API method (PUT with array) | 1h | `internal/api/resources.go` | Unit test with httptest |
| T3.3 | Implement JSON array parsing and validation | 1h | `internal/cli/time_entry.go` | Invalid JSON shows error |
| T3.4 | Write unit tests for bulk command | 1.5h | `internal/cli/time_entry_test.go` | `go test ./internal/cli/` passes |

**US-3.2: Hydrated Output Enhancement**
> As a user, I want `--hydrated` on list and get commands so I can see full project/task/user names inline.

**Acceptance Criteria:**
- `clockify time-entry list --hydrated` includes full objects
- `clockify time-entry get --hydrated` includes full objects
- Table output shows PROJECT_NAME instead of PROJECT_ID
- JSON output includes nested objects

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.5 | Add `--hydrated` flag to list and get commands | 1h | `internal/cli/time_entry.go` | `--hydrated` flag works |
| T3.6 | Update table headers to show hydrated names when available | 1h | `internal/cli/time_entry.go` | Table shows PROJECT_NAME |
| T3.7 | Write unit tests for hydrated output | 1h | `internal/cli/time_entry_test.go` | `go test ./internal/cli/` passes |

**US-3.3: End-to-End Verification**
> As a team lead, I want all M3 success criteria verified so that the milestone is complete.

**Acceptance Criteria:**
- [ ] `clockify time-entry start --project X` starts a timer; `clockify time-entry stop` stops it
- [ ] Duration parsing handles all formats: seconds, `1h`, `1.5h`, `30m`, `1h30m`
- [ ] `clockify time-entry list --start today --end today` returns today's entries
- [ ] `--output json | jq` pipeline works for scripting
- [ ] Bulk create from JSON file processes all entries
- [ ] Confirmation required for delete; `--confirm` skips it

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.8 | Run full E2E verification against success criteria | 3h | — | All 6 criteria pass |
| T3.9 | Fix any issues found during E2E verification | 2h | Various | Re-verify all criteria |
| T3.10 | Create M3 completion commit and tag | 0.5h | — | `git tag v0.4.0-m3` |

### Sprint 3 Definition of Done

- [ ] `clockify time-entry bulk` works with --file and --stdin
- [ ] `--hydrated` flag works on list and get commands
- [ ] All 6 M3 success criteria verified
- [ ] All unit tests pass: `go test ./...`
- [ ] No lint errors: `golangci-lint run`
- [ ] M3 completion tag created

### Daily Checkpoints — Sprint 3

| Day | End-of-Day Verification |
|-----|------------------------|
| D11 | `clockify time-entry bulk` works with JSON input |
| D12 | `--hydrated` flag working on list and get |
| D13 | All E2E criteria pass; issues fixed |
| D14 | Remaining edge cases handled; all tests pass |
| D15 | M3 tag created; documentation updated |

---

## 6. Task Dependency Map

```
S1 (Duration, API Client & Read)
├── T1.1–T1.3 (duration parser) ──► T1.4–T1.5 (relative dates)
├── T1.6–T1.10 (API methods) ──► T1.11 (API tests)
│   ├──► T1.12–T1.16 (list command) ──► T1.14 (pagination helper)
│   └──► T1.17–T1.18 (get command)

S2 (Write & Timer)
├── T2.1–T2.4 (create command)
├── T2.5–T2.7 (update command)
├── T2.8–T2.10 (delete command)
├── T2.11–T2.14 (start timer)
└── T2.15–T2.18 (stop timer)

S3 (Bulk, Hydrated & E2E)
├── T3.1–T3.4 (bulk command)
├── T3.5–T3.7 (hydrated flag)
└── T3.8–T3.10 (E2E + tag)
```

### Critical Path

```
T1.1–T1.3 (duration) → T1.6–T1.10 (API methods) → T1.12–T1.16 (list command)
  → T2.1–T2.4 (create) → T2.11–T2.14 (start) → T2.15–T2.18 (stop)
  → T3.8–T3.10 (E2E)
```

**Estimated critical path duration:** ~35 hours (7 days of focused work)
**Buffer for non-critical tasks:** ~15 hours (3 days)
**Total estimated:** ~50 hours (10 days) = 2 weeks + buffer → fits in 3-week sprint

---

## 7. Risk Mitigation

| Risk | Mitigation | Owner | Sprint |
|------|------------|-------|--------|
| **Timer start/stop race conditions** | Check in-progress before start/stop; handle 409 conflicts gracefully with user-friendly message | Dev | S2 |
| **Duration parsing ambiguity** | Strict regex patterns; clear error messages showing supported formats | Dev | S1 |
| **Clockify API create requires start in ISO8601** | Parse user input with `time.Parse()` using multiple formats; default to UTC | Dev | S2 |
| **Bulk operation partial failures** | Process entries sequentially; report per-entry success/failure; continue on error | Dev | S3 |
| **Relative date edge cases** | Test with different week starts; handle timezone considerations | Dev | S1 |
| **Large time entry lists** | Pagination working from M2; test with `--no-pagination` on large datasets | Dev | S1 |

---

## 8. Trade-offs: Where to Cut Corners vs. Where to Hold Firm

### Hold Firm (Non-Negotiable)

| Area | Why |
|------|-----|
| **Duration parser** | Core UX; users must be able to type `1h30m` |
| **Timer start/stop** | Most-used feature; must be reliable |
| **Confirmation on delete** | Destructive operation; must prevent accidents |
| **All output formats per command** | Consistent UX from M2 |
| **Relative date parsing** | Essential for `--start today` workflow |

### Cut Corners (If Behind Schedule)

| Area | How to Cut | Impact |
|------|------------|--------|
| **Bulk command** | Defer to M4; users can create one at a time | Power users affected |
| **`--hydrated` on list** | Only on get command; list shows IDs | Less readable, but functional |
| **Task filter** | Defer `--task` filter to M4 | Users can filter with `jq` |
| **Billable flag** | Default to workspace default; skip `--billable` flag | Users set billable in UI |

### Cut Corners — Emergency Only

| Area | How to Cut | Impact |
|------|------------|--------|
| **File input** | Stdin only; defer `--file` flag | Minor UX loss |
| **Tag filtering** | Defer `--tags` filter | Users can filter locally |
| **CSV/YAML output** | JSON + table only for time-entry | Covers 90% of use cases |

---

## 9. File-Level Task Summary

| File | Sprint | Tasks | Est. Hours |
|------|--------|-------|------------|
| `pkg/duration/duration.go` | S1 | T1.1–T1.2 | 3h |
| `pkg/duration/duration_test.go` | S1 | T1.3 | 1.5h |
| `pkg/duration/relative.go` | S1 | T1.4 | 2h |
| `pkg/duration/relative_test.go` | S1 | T1.5 | 1h |
| `internal/api/resources.go` | S1/S2/S3 | T1.6–T1.10, T2.2, T2.6, T2.9, T2.12, T2.16, T3.2 | 8h |
| `internal/api/resources_test.go` | S1/S2 | T1.11 | 2h |
| `internal/cli/time_entry.go` | S1/S2/S3 | T1.12–T1.13, T1.17, T2.1, T2.3, T2.5, T2.8, T2.11, T2.13, T2.15, T2.17, T3.1, T3.3, T3.5–T3.6 | 24h |
| `internal/cli/time_entry_test.go` | S1/S2/S3 | T1.16, T1.18, T2.4, T2.7, T2.10, T2.14, T2.18, T3.4, T3.7 | 12h |
| `internal/cli/pagination.go` | S1 | T1.14 | 1.5h |
| `internal/cli/root.go` | S1 | T1.15 | 0.5h |

---

## 10. M3 Overall Definition of Done

| Criterion | Verification | Status |
|-----------|-------------|--------|
| Duration parser handles all formats | Unit tests for `1h`, `1.5h`, `30m`, `1h30m`, `3600` | ☐ |
| Relative date parsing works | `--start today`, `--start this-week` resolve correctly | ☐ |
| `clockify time-entry list` with filters | `--start`, `--end`, `--project`, `--description`, `--in-progress` | ☐ |
| `clockify time-entry get` by ID | `--id <valid-id>` returns entry | ☐ |
| `clockify time-entry create` with duration | `--start X --duration 1h` creates entry | ☐ |
| `clockify time-entry create` with end time | `--start X --end Y` creates entry | ☐ |
| `clockify time-entry update` partial | `--id X --description "new"` updates only description | ☐ |
| `clockify time-entry delete` with confirmation | Prompts; `--confirm` bypasses | ☐ |
| `clockify time-entry start` starts timer | Creates entry without end time | ☐ |
| `clockify time-entry stop` stops timer | PATCHes running entry with end time | ☐ |
| Bulk create from JSON file | `--file entries.json` processes all entries | ☐ |
| `--hydrated` includes full objects | Project/task/user names shown instead of IDs | ☐ |
| Stdin/file input works | `echo '...' \| clockify time-entry create --stdin` | ☐ |
| `--output json \| jq` pipeline | Scripting works without errors | ☐ |
| All unit tests pass | `go test ./...` | ☐ |
| No lint errors | `golangci-lint run` | ☐ |
| M3 completion tag created | `git tag v0.4.0-m3` | ☐ |

---

## 11. TimeEntry Type Extensions

The existing `TimeEntry` struct may need additional fields to match the full API response:

```go
type TimeEntry struct {
    ID              string         `json:"id" yaml:"id"`
    Description     string         `json:"description" yaml:"description"`
    WorkspaceID     string         `json:"workspaceId" yaml:"workspaceId"`
    ProjectID       string         `json:"projectId" yaml:"projectId"`
    TaskID          string         `json:"taskId" yaml:"taskId"`
    UserID          string         `json:"userId" yaml:"userId"`
    Billable        bool           `json:"billable" yaml:"billable"`
    TimeInterval    TimeInterval   `json:"timeInterval" yaml:"timeInterval"`
    TagIDs          []string       `json:"tagIds" yaml:"tagIds"`
    Type            string         `json:"type" yaml:"type"`               // "REGULAR" or "BREAK"
    IsLocked        bool           `json:"isLocked" yaml:"isLocked"`
    CustomFieldValues interface{}  `json:"customFieldValues" yaml:"customFieldValues"`
}
```

---

## 12. Daily Standup Template

Use this at end of each day:

```markdown
## Day X Standup — Sprint Y

**Done today:**
- [List completed tasks]

**Blocked:**
- [Any blockers]

**Plan for tomorrow:**
- [Next tasks]

**Confidence for sprint:**
- 🟢 On track / 🟡 At risk / 🔴 Behind
```

---

*This plan is a living document. Update task estimates and status as work progresses.*
