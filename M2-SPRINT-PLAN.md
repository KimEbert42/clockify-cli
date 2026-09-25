# M2: Read Operations — Sprint Plan

**Milestone:** M2
**Sprint Duration:** 2 × 1-week sprints (5 working days each)
**Developer:** 1 (solo)
**Start Date:** TBD
**End Date:** TBD (2 weeks from start)

---

## 1. Sprint Structure

| Sprint | Theme | Days | Goal |
|--------|-------|------|------|
| **S1** | Types, API Client & Workspace/User Commands | D1–D5 | Entity types, API methods, workspace list/get, user get/list/info |
| **S2** | Project/Client/Tag Commands & Pagination | D6–D10 | Project list/get, client list/get, tag list, pagination, E2E verification |

---

## 2. Current State Assessment

### Already Implemented (from M0 + M1)
| Feature | Status | Notes |
|---------|--------|-------|
| HTTP client (`api.ClockifyClient.Do()`) | ✅ | Auth headers, retry, rate limiting all working |
| Output formatters (table/JSON/CSV/YAML) | ✅ | `Formatter.Format(w, data, headers)` interface |
| Resource ID resolution | ✅ | `resolve.ResolveInput()` — exact, prefix, fuzzy |
| Global flags (`--page`, `--page-size`, `--no-pagination`) | ✅ | Registered on root command |
| Config loading with profiles | ✅ | `config.Load()` returns `*Config` with workspace ID |
| Error types with codes | ✅ | `CLIError` with E001–E010 categories |
| Aliases system | ✅ | `registerAliases()` resolves at startup |

### What M2 Needs to Build
| Component | Status | Notes |
|-----------|--------|-------|
| Entity types (`pkg/clockify/types.go`) | ❌ Empty | All API responses currently `map[string]any` |
| API methods for list/get | ❌ | Need typed methods on `ClockifyClient` |
| Workspace commands | ❌ | `workspace list`, `workspace get` |
| User commands | ❌ | `user get`, `user list`, `user info` |
| Project commands | ❌ | `project list`, `project get` |
| Client commands | ❌ | `client list`, `client get` |
| Tag commands | ❌ | `tag list` |
| Pagination logic | ❌ | Auto-fetch all pages with `--no-pagination` |

### Clockify API Endpoints (M2 Scope)

| Resource | Method | Path | Query Params |
|----------|--------|------|-------------|
| Workspaces | GET | `/workspaces` | — |
| Workspace | GET | `/workspaces/{id}` | — |
| Users (workspace) | GET | `/workspaces/{wid}/users` | `status`, `page`, `pageSize` |
| User (current) | GET | `/user` | — |
| Projects | GET | `/workspaces/{wid}/projects` | `status`, `clientId`, `search`, `page`, `pageSize` |
| Project | GET | `/workspaces/{wid}/projects/{id}` | — |
| Clients | GET | `/workspaces/{wid}/clients` | `search`, `page`, `pageSize` |
| Client | GET | `/workspaces/{wid}/clients/{id}` | — |
| Tags | GET | `/workspaces/{wid}/tags` | `search`, `page`, `pageSize` |

**Pagination:** `Last-Page` response header indicates total pages. Default `pageSize=100`, max=5000.

---

## 3. Sprint 1: Types, API Client & Workspace/User Commands (Days 1–5)

### User Stories

**US-1.1: Entity Types**
> As a developer, I want typed structs for all Clockify entities so that API responses are strongly typed and code is maintainable.

**Acceptance Criteria:**
- `pkg/clockify/types.go` defines: `Workspace`, `User`, `Project`, `Client`, `Tag`
- Each struct has `json` and `yaml` tags matching Clockify API field names
- Pagination response type: `PageInfo` with `Page`, `PageSize`, `TotalPages`, `TotalResults`
- Types are used by all subsequent API methods

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.1 | Define `Workspace` struct (ID, Name, HourlyRate, Currency, etc.) | 0.5h | `pkg/clockify/types.go` | `go build ./pkg/clockify/` |
| T1.2 | Define `User` struct (ID, Email, Name, Status, Profile) | 0.5h | `pkg/clockify/types.go` | `go build ./pkg/clockify/` |
| T1.3 | Define `Project` struct (ID, Name, ClientID, Billable, Status, Color, Estimate) | 0.5h | `pkg/clockify/types.go` | `go build ./pkg/clockify/` |
| T1.4 | Define `Client` struct (ID, Name, Email, Address, Phone, Note) | 0.5h | `pkg/clockify/types.go` | `go build ./pkg/clockify/` |
| T1.5 | Define `Tag` struct (ID, Name, Color) | 0.25h | `pkg/clockify/types.go` | `go build ./pkg/clockify/` |
| T1.6 | Define `PageInfo` struct and `LastPage` header parser | 0.5h | `pkg/clockify/types.go` | `go build ./pkg/clockify/` |
| T1.7 | Write unit tests for JSON deserialization of each type | 1.5h | `pkg/clockify/types_test.go` | `go test ./pkg/clockify/` passes |

**US-1.2: Typed API Methods**
> As a developer, I want typed methods on `ClockifyClient` for each resource so that commands don't deal with raw maps.

**Acceptance Criteria:**
- `client.ListWorkspaces(ctx) ([]Workspace, error)`
- `client.GetWorkspace(ctx, id) (*Workspace, error)`
- `client.ListUsers(ctx, status string) ([]User, error)`
- `client.GetCurrentUser(ctx) (*User, error)`
- `client.ListProjects(ctx, opts ListProjectsOpts) ([]Project, error)`
- `client.GetProject(ctx, id) (*Project, error)`
- `client.ListClients(ctx, opts ListClientsOpts) ([]Client, error)`
- `client.GetClient(ctx, id) (*Client, error)`
- `client.ListTags(ctx, opts ListTagsOpts) ([]Tag, error)`
- All methods use `Do()` internally and unmarshal into typed structs

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.8 | Add `ListWorkspaces()` and `GetWorkspace()` to client | 1h | `internal/api/client.go` | Unit test with httptest |
| T1.9 | Add `ListUsers()` (with status filter) and `GetCurrentUser()` | 1h | `internal/api/client.go` | Unit test with httptest |
| T1.10 | Add `ListProjects()` (with status/client/search filters) and `GetProject()` | 1h | `internal/api/client.go` | Unit test with httptest |
| T1.11 | Add `ListClients()` (with search filter) and `GetClient()` | 1h | `internal/api/client.go` | Unit test with httptest |
| T1.12 | Add `ListTags()` (with search filter) | 0.5h | `internal/api/client.go` | Unit test with httptest |
| T1.13 | Write integration tests for all API methods | 2h | `internal/api/client_test.go` | `go test ./internal/api/` passes |

**US-1.3: Workspace Commands**
> As a user, I want `clockify workspace list` and `clockify workspace get` so that I can see and select my workspaces.

**Acceptance Criteria:**
- `clockify workspace list` lists all workspaces in table format
- `clockify workspace get --id <id>` gets workspace by ID
- `clockify workspace get --name <name>` gets workspace by name (fuzzy match)
- `clockify workspace get` (no args) gets current workspace from config
- All output formats work (table, JSON, CSV, YAML)
- Table columns: ID, NAME, HOURLY_RATE, CURRENCY

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.14 | Create `internal/cli/workspace.go` with `newWorkspaceCmd()` factory | 1h | `internal/cli/workspace.go` | `go build ./cmd/clockify/` |
| T1.15 | Implement `newWorkspaceListCmd()` — calls `client.ListWorkspaces()` | 1.5h | `internal/cli/workspace.go` | `clockify workspace list` shows workspaces |
| T1.16 | Implement `newWorkspaceGetCmd()` — with `--id` and `--name` flags | 1.5h | `internal/cli/workspace.go` | `clockify workspace get --name "X"` works |
| T1.17 | Register workspace command in `root.go` | 0.5h | `internal/cli/root.go` | `clockify --help` shows workspace |
| T1.18 | Write unit tests for workspace commands (httptest mock) | 1.5h | `internal/cli/workspace_test.go` | `go test ./internal/cli/` passes |

**US-1.4: User Commands**
> As a user, I want `clockify user get`, `clockify user list`, and `clockify user info` so that I can see user information.

**Acceptance Criteria:**
- `clockify user get` shows current user profile
- `clockify user list` lists workspace users
- `clockify user list --status ACTIVE` filters by status
- `clockify user info --email user@acme.com` searches by email
- `clockify user info --name "John"` searches by name
- Table columns: ID, NAME, EMAIL, STATUS

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.19 | Create `internal/cli/user.go` with `newUserCmd()` factory | 0.5h | `internal/cli/user.go` | `go build ./cmd/clockify/` |
| T1.20 | Implement `newUserGetCmd()` — calls `client.GetCurrentUser()` | 1h | `internal/cli/user.go` | `clockify user get` shows current user |
| T1.21 | Implement `newUserListCmd()` — with `--status` filter | 1.5h | `internal/cli/user.go` | `clockify user list --status ACTIVE` works |
| T1.22 | Implement `newUserInfoCmd()` — with `--email` and `--name` flags | 1.5h | `internal/cli/user.go` | `clockify user info --email X` works |
| T1.23 | Register user command in `root.go` | 0.5h | `internal/cli/root.go` | `clockify --help` shows user |
| T1.24 | Write unit tests for user commands | 1.5h | `internal/cli/user_test.go` | `go test ./internal/cli/` passes |

### Sprint 1 Definition of Done

- [ ] Entity types defined in `pkg/clockify/types.go` with json/yaml tags
- [ ] Typed API methods for workspace, user, project, client, tag
- [ ] `clockify workspace list` works in all output formats
- [ ] `clockify workspace get --id/--name` works
- [ ] `clockify user get` shows current user
- [ ] `clockify user list --status` filters correctly
- [ ] `clockify user info --email/--name` searches correctly
- [ ] All unit tests pass: `go test ./...`
- [ ] No lint errors: `golangci-lint run`

### Daily Checkpoints — Sprint 1

| Day | End-of-Day Verification |
|-----|------------------------|
| D1 | All entity types defined; JSON deserialization tests pass |
| D2 | Typed API methods for workspace + user; httptest mocks work |
| D3 | `clockify workspace list/get` commands work end-to-end |
| D4 | `clockify user get/list/info` commands work end-to-end |
| D5 | All S1 tests pass; `golangci-lint run` clean |

---

## 4. Sprint 2: Project/Client/Tag Commands & Pagination (Days 6–10)

### User Stories

**US-2.1: Project Commands**
> As a user, I want `clockify project list` and `clockify project get` so that I can browse and find projects.

**Acceptance Criteria:**
- `clockify project list` lists all projects in current workspace
- `clockify project list --status ACTIVE` filters by status
- `clockify project list --client <id>` filters by client
- `clockify project list --search "web"` searches by name
- `clockify project get --id <id>` gets project by ID
- `clockify project get --name "Website"` gets project by name (fuzzy match)
- Table columns: ID, NAME, CLIENT, BILLABLE, STATUS

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.1 | Create `internal/cli/project.go` with `newProjectCmd()` factory | 0.5h | `internal/cli/project.go` | `go build ./cmd/clockify/` |
| T2.2 | Implement `newProjectListCmd()` — with `--status`, `--client`, `--search` filters | 2h | `internal/cli/project.go` | `clockify project list --status ACTIVE` works |
| T2.3 | Implement `newProjectGetCmd()` — with `--id` and `--name` flags, client name resolution | 1.5h | `internal/cli/project.go` | `clockify project get --name "X"` works |
| T2.4 | Register project command in `root.go` | 0.5h | `internal/cli/root.go` | `clockify --help` shows project |
| T2.5 | Write unit tests for project commands | 1.5h | `internal/cli/project_test.go` | `go test ./internal/cli/` passes |

**US-2.2: Client Commands**
> As a user, I want `clockify client list` and `clockify client get` so that I can browse clients.

**Acceptance Criteria:**
- `clockify client list` lists all clients
- `clockify client list --search "Acme"` searches by name
- `clockify client get --id <id>` gets client by ID
- Table columns: ID, NAME, EMAIL, PHONE

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.6 | Create `internal/cli/client.go` with `newClientCmd()` factory | 0.5h | `internal/cli/client.go` | `go build ./cmd/clockify/` |
| T2.7 | Implement `newClientListCmd()` — with `--search` filter | 1h | `internal/cli/client.go` | `clockify client list --search "X"` works |
| T2.8 | Implement `newClientGetCmd()` — with `--id` flag | 1h | `internal/cli/client.go` | `clockify client get --id X` works |
| T2.9 | Register client command in `root.go` | 0.5h | `internal/cli/root.go` | `clockify --help` shows client |
| T2.10 | Write unit tests for client commands | 1h | `internal/cli/client_test.go` | `go test ./internal/cli/` passes |

**US-2.3: Tag Commands**
> As a user, I want `clockify tag list` so that I can browse tags.

**Acceptance Criteria:**
- `clockify tag list` lists all tags
- `clockify tag list --search "meeting"` searches by name
- Table columns: ID, NAME, COLOR

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.11 | Create `internal/cli/tag.go` with `newTagCmd()` factory | 0.5h | `internal/cli/tag.go` | `go build ./cmd/clockify/` |
| T2.12 | Implement `newTagListCmd()` — with `--search` filter | 1h | `internal/cli/tag.go` | `clockify tag list --search "X"` works |
| T2.13 | Register tag command in `root.go` | 0.5h | `internal/cli/root.go` | `clockify --help` shows tag |
| T2.14 | Write unit tests for tag commands | 1h | `internal/cli/tag_test.go` | `go test ./internal/cli/` passes |

**US-2.4: Pagination**
> As a user, I want `--no-pagination` to auto-fetch all pages so that I can get complete results without manual paging.

**Acceptance Criteria:**
- `--page <n>` and `--page-size <n>` flags send correct query params
- `--no-pagination` fetches all pages automatically
- Progress indicator shows page count (unless `--quiet`)
- `Last-Page` header parsed to determine total pages
- Default: single page (page=1, pageSize=100)

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.15 | Implement `fetchPaginated()` helper in `internal/api/client.go` | 2h | `internal/api/client.go` | Auto-fetches all pages |
| T2.16 | Wire `--page`, `--page-size`, `--no-pagination` flags to API calls | 1h | `internal/api/client.go` | Query params sent correctly |
| T2.17 | Add progress indicator for multi-page fetches | 1h | `internal/api/client.go` | Shows "Fetching page X/Y..." |
| T2.18 | Write unit tests for pagination (mock multi-page responses) | 1.5h | `internal/api/client_test.go` | `go test ./internal/api/` passes |

**US-2.5: End-to-End Verification**
> As a team lead, I want all M2 success criteria verified so that the milestone is complete.

**Acceptance Criteria:**
- [ ] All list commands return correctly formatted output in all formats
- [ ] Filters work correctly (e.g., `--status ACTIVE`, `--search "web"`)
- [ ] Pagination auto-fetches all pages with `--no-pagination`
- [ ] Name resolution works: `clockify project get --name "Website"` finds the right project
- [ ] Fuzzy matching warns on multiple matches and suggests `--id`

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.19 | Run full E2E verification against success criteria | 2h | — | All 5 criteria pass |
| T2.20 | Fix any issues found during E2E verification | 1.5h | Various | Re-verify all criteria |
| T2.21 | Create M2 completion commit and tag | 0.5h | — | `git tag v0.3.0-m2` |

### Sprint 2 Definition of Done

- [ ] `clockify project list/get` works with all filters
- [ ] `clockify client list/get` works with search
- [ ] `clockify tag list` works with search
- [ ] Pagination works with `--page`, `--page-size`, `--no-pagination`
- [ ] All 5 M2 success criteria verified
- [ ] All unit tests pass: `go test ./...`
- [ ] No lint errors: `golangci-lint run`
- [ ] M2 completion tag created

### Daily Checkpoints — Sprint 2

| Day | End-of-Day Verification |
|-----|------------------------|
| D6 | `clockify project list/get` commands work |
| D7 | `clockify client list/get` commands work |
| D8 | `clockify tag list` works; pagination helper implemented |
| D9 | All E2E criteria pass; issues fixed |
| D10 | All S2 tests pass; M2 tag created |

---

## 5. Task Dependency Map

```
S1 (Types, API Client & Workspace/User)
├── T1.1–T1.6 (entity types) ──► T1.7 (type tests)
│   └──► T1.8–T1.12 (API methods) ──► T1.13 (API tests)
│       ├──► T1.14–T1.18 (workspace commands)
│       └──► T1.19–T1.24 (user commands)

S2 (Project/Client/Tag & Pagination)
├── T2.1–T2.5 (project commands)
├── T2.6–T2.10 (client commands)
├── T2.11–T2.14 (tag commands)
├── T2.15–T2.18 (pagination)
└── T2.19–T2.21 (E2E + tag)
```

### Critical Path

```
T1.1–T1.6 (types) → T1.8–T1.12 (API methods) → T1.14–T1.16 (workspace cmds)
  → T2.1–T2.3 (project cmds) → T2.15–T2.17 (pagination) → T2.19–T2.21 (E2E)
```

**Estimated critical path duration:** ~30 hours (6 days of focused work)
**Buffer for non-critical tasks:** ~10 hours (2 days)
**Total estimated:** ~40 hours (8 days) = 1.5 weeks + buffer → fits in 2-week sprint

---

## 6. Risk Mitigation

| Risk | Mitigation | Owner | Sprint |
|------|------------|-------|--------|
| **Clockify API response shapes differ from spec** | Prototype first API call on D1; define types from actual response; use `json:"-"` for unknown fields | Dev | S1 |
| **Pagination `Last-Page` header missing or different** | Check response headers empirically; fall back to empty-page detection | Dev | S2 |
| **Fuzzy matching produces false positives** | Use `sahilm/fuzzy` with scoring threshold; show warning on multiple matches, suggest `--id` | Dev | S1 |
| **Client ID resolution for project list** | Projects reference clients by ID; table output shows client name via separate lookup or skip if slow | Dev | S2 |
| **Large workspace with many pages** | Test with `--no-pagination` on workspaces with 500+ items; verify memory usage | Dev | S2 |

---

## 7. Trade-offs: Where to Cut Corners vs. Where to Hold Firm

### Hold Firm (Non-Negotiable)

| Area | Why |
|------|-----|
| **Typed entity structs** | Foundation for all M2+ commands; `map[string]any` is unmaintainable |
| **All output formats per command** | Consistent UX; users expect `--output json` everywhere |
| **Name resolution on get commands** | Core UX: users shouldn't need to memorize UUIDs |
| **Pagination with `--no-pagination`** | Essential for scripting; users need all data, not just page 1 |

### Cut Corners (If Behind Schedule)

| Area | How to Cut | Impact |
|------|------------|--------|
| **Client name in project table** | Show client ID instead of name; skip the join lookup | Less readable, but functional |
| **`user info` search** | Implement email only; defer name search | Email is more reliable anyway |
| **Progress indicator** | Skip page progress indicator; use `--quiet` default | Minor UX loss |
| **Unit test coverage** | Target 60% instead of 80%; focus on API methods + commands | Critical paths mandatory |

### Cut Corners — Emergency Only

| Area | How to Cut | Impact |
|------|------------|--------|
| **Tag list search** | List all tags without `--search` filter only | Users can `jq` locally |
| **CSV/YAML output** | JSON + table only; defer CSV/YAML for M2 commands | Covers 90% of use cases |

---

## 8. File-Level Task Summary

| File | Sprint | Tasks | Est. Hours |
|------|--------|-------|------------|
| `pkg/clockify/types.go` | S1 | T1.1–T1.6 | 2.75h |
| `pkg/clockify/types_test.go` | S1 | T1.7 | 1.5h |
| `internal/api/client.go` | S1 | T1.8–T1.12 | 4.5h |
| `internal/api/client_test.go` | S1 | T1.13 | 2h |
| `internal/cli/workspace.go` | S1 | T1.14–T1.17 | 4.5h |
| `internal/cli/workspace_test.go` | S1 | T1.18 | 1.5h |
| `internal/cli/user.go` | S1 | T1.19–T1.23 | 5h |
| `internal/cli/user_test.go` | S1 | T1.24 | 1.5h |
| `internal/cli/project.go` | S2 | T2.1–T2.4 | 4.5h |
| `internal/cli/project_test.go` | S2 | T2.5 | 1.5h |
| `internal/cli/client.go` | S2 | T2.6–T2.9 | 3h |
| `internal/cli/client_test.go` | S2 | T2.10 | 1h |
| `internal/cli/tag.go` | S2 | T2.11–T2.13 | 2h |
| `internal/cli/tag_test.go` | S2 | T2.14 | 1h |
| `internal/api/client.go` (pagination) | S2 | T2.15–T2.17 | 4h |
| `internal/api/client_test.go` (pagination) | S2 | T2.18 | 1.5h |
| `internal/cli/root.go` | S1/S2 | T1.17, T1.23, T2.4, T2.9, T2.13 | 2h (registrations) |

---

## 9. M2 Overall Definition of Done

| Criterion | Verification | Status |
|-----------|-------------|--------|
| All list commands return correctly formatted output | Run each with `--output json/table/csv/yaml` | ☐ |
| Filters work correctly | `--status ACTIVE`, `--search "web"`, `--client <id>` | ☐ |
| Pagination auto-fetches all pages | `--no-pagination` on multi-page workspace | ☐ |
| Name resolution works | `clockify project get --name "Website"` | ☐ |
| Fuzzy matching warns on multiple matches | `clockify project get --name "app"` with 2+ matches | ☐ |
| All unit tests pass | `go test ./...` | ☐ |
| No lint errors | `golangci-lint run` | ☐ |
| M2 completion tag created | `git tag v0.3.0-m2` | ☐ |

---

## 10. Daily Standup Template

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
