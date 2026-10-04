# M7: Secondary Resources — Sprint Plan

**Milestone:** M7
**Sprint Duration:** 2 × 2-day sprints (compressed from original 2–4 days plan)
**Developer:** 1 (solo)
**Start Date:** 2026-10-04
**End Date:** 2026-10-04
**Status:** ✅ **COMPLETE**
**Prerequisite:** Tag `v0.7.0-m6` on M6 commit `817c831` — **done 2026-10-04**
**Spec reference:** `clockify-cli-spec.md` §8.12–§8.19
**API reference:** docs.clockify.me — official v1 endpoints (verified 2026-10-04)

**Implementation summary:** All 8 resource command families shipped in one day using official API docs + M4/M5 patterns. Unit tests green across packages. Live E2E against a paid workspace still recommended for invoice/approval/webhook plan-gated features.

---

## 1. Sprint Structure

| Sprint | Theme | Days | Goal |
|--------|-------|------|------|
| **S1** | Standard CRUD Batch | D1–D2 | Types, API methods, and full CLI for group, webhook, custom-field, expense, invoice |
| **S2** | Workflow Resources + E2E | D3–D4 | time-off, schedule, approval + tests, PLAN.md update, tag |

**Why this split:** S1 resources follow proven M4/M5 patterns (flat workspace-scoped CRUD). S2 resources have policy/workflow semantics, non-obvious paths, or spec↔API mismatches that need focused design.

---

## 2. Current State Assessment

### Already Implemented (M0–M6)

| Feature | Status | Notes |
|---------|--------|-------|
| CLI framework + global flags | ✅ | `--profile`, `--output`, `--confirm`, `--file`/`--stdin` patterns |
| HTTP client + retry + pagination | ✅ | `Do()`, `DoWithPagination()`, `fetchPaginated*()` helpers |
| Output formatters | ✅ | table / JSON / CSV / YAML via `formatOutput()` |
| Name resolution | ✅ | `resolveProjectRef`, `resolveUserRef`, `resolveClientByName`, etc. |
| Confirmation on delete | ✅ | `--confirm` global flag pattern from M4 |
| JSON file/stdin input | ✅ | `--file` / `--stdin` on create/update (M4) |
| Idempotent create | ✅ | `--if-not-exists` pattern (M4) |
| Relative dates + duration parse | ✅ | `pkg/duration.ParseRelativeDate`, `Parse` |
| CRUD command shape | ✅ | `internal/cli/{project,client,tag,task}.go` — factory funcs, display structs |
| Root command registration | ⬜ | Need `internal/cli/root.go` additions for 8 new parent commands |
| Aliases | ⬜ | Optional: no M7 aliases required by spec |

### What M7 Needs to Build

| Component | Status | Notes |
|-----------|--------|-------|
| Entity types (8 resources) | ❌ | `pkg/clockify/m7_types.go` or extend `types.go` |
| Request types | ❌ | Create/Update per resource |
| API methods | ❌ | `internal/api/m7_resources.go` (keep `resources.go` from bloating) |
| CLI commands | ❌ | One file per resource under `internal/cli/` |
| Pagination helpers | ❌ | `fetchPaginated*` for list endpoints that support page/pageSize |
| Enum validation maps | ❌ | Statuses, types, webhook events, custom-field types |
| Spec↔API type mapping | ❌ | Friendly CLI names → official API enums |
| Unit tests | ❌ | API httptest + CLI flag/validation tests |

---

## 3. Official API Endpoints (docs.clockify.me)

All paths are relative to `{base}/workspaces/{workspaceId}` unless noted. Auth: `X-Api-Key`.

### 3.1 Invoice — §8.12

| Op | Method | Path | Notes |
|----|--------|------|-------|
| List | GET | `/invoices` | Response: `{invoices, total}` — **not a bare array** |
| Get | GET | `/invoices/{invoiceId}` | |
| Create | POST | `/invoices` | Required: `clientId`, `currency`, `issuedDate`, `dueDate`, `number` |
| Update | PUT | `/invoices/{invoiceId}` | Full replace |
| Delete | DELETE | `/invoices/{invoiceId}` | |
| Status change | PATCH | `/invoices/{invoiceId}/status` | Body: `{invoiceStatus}` |

**Invoice status (official PATCH enum):** `UNSENT`, `SENT`, `PAID`, `PARTIALLY_PAID`, `VOID`, `OVERDUE`

**Spec status list:** `DRAFT`, `FINALIZED`, `SENT`, `ACCEPTED`, `PAID`, `VOID` — **partially wrong**. CLI accepts both; maps spec names → API names where possible; rejects unknown with actionable error.

**Required create fields vs spec:** Spec says `--client --start --end`. Official API also requires `currency`, `number`, and uses `issuedDate`/`dueDate` (not start/end). CLI maps:
- `--start` → `issuedDate` (ISO8601)
- `--end` → `dueDate` (ISO8601)
- `--currency` required (default from workspace if available, else error)
- `--number` required (or auto-generate timestamp-based if omitted — document choice)

**Out of M7 scope (defer):** items, payments, duplicate, export, invoice settings, filter endpoint.

### 3.2 Schedule — §8.13

| Op | Method | Path | Notes |
|----|--------|------|-------|
| List | GET | `/scheduling/assignments/all` | **Requires** `start` and `end` query params (ISO8601) |
| Create | POST | `/scheduling/assignments/recurring` | Required body: `projectId`, `userId`, `start`, `end`, `hoursPerDay` |
| Update | PATCH | `/scheduling/assignments/recurring/{assignmentId}` | Body: partial; optional `seriesUpdateOption` |
| Delete | DELETE | `/scheduling/assignments/recurring/{assignmentId}` | Query: `seriesUpdateOption` = `THIS_ONE` \| `THIS_AND_FOLLOWING` \| `ALL` |

**CLI mapping:**
- Spec `--duration 8h` → `hoursPerDay` (hours as float; parse via `pkg/duration`)
- Spec `--user`, `--project`, `--start`, `--end` map directly
- List always requires date range (default: `this-month` → start + end if omitted — document as CLI convenience)
- Response is hydrated (projectId/projectName/userName/clientId etc.) — display flattened structs

**Out of M7 scope:** publish, copy, recurring period change, capacity totals, project totals.

### 3.3 Time-off — §8.14

| Op | Method | Path | Notes |
|----|--------|------|-------|
| List | POST | `/time-off/requests` | **POST not GET.** Body: `start`, `end`, `statuses[]`, `users[]`, `page`, `pageSize` |
| Create | POST | `/time-off/policies/{policyId}/requests` | Body: `note`, `timeOffPeriod{start,end}` |
| Update status | PATCH | `/time-off/policies/{policyId}/requests/{requestId}` | Body: `{note, status}` — `APPROVED`\|`REJECTED` |
| Delete | DELETE | `/time-off/policies/{policyId}/requests/{requestId}` | Exists in official API; PLAN says "no delete per spec" |

**Spec vs API:**
- Spec create: `--type VACATION|SICK|...` — Official create is **policy-scoped**, not type-scoped. CLI: require `--policy` (ID or name via `resolvePolicyByName`) OR add `--type` that resolves to a policy by matching type/name if found; otherwise clear error explaining policy requirement.
- Spec list: `--status PENDING` → body `statuses: ["PENDING"]`
- Spec update-status: `--id` + `--status` — also need `--policy` for the path. CLI resolves request if possible; otherwise require `--policy`.
- Spec types (VACATION, SICK, PERSONAL, UNPAID, OTHER) kept as user-facing labels for display/filter; policy association is authoritative.

**Decision:** Implement delete behind `--confirm` even though spec omits it (official API supports it). Note in PLAN.md.

**Out of M7 scope:** policies CRUD, balances, balance assignments.

### 3.4 Expense — §8.15

| Op | Method | Path | Notes |
|----|--------|------|-------|
| List | GET | `/expenses` | Query: `project`, `start`, `end`, `user-id`, page params. Response: `{expenses, dailyTotals, weeklyTotals}` |
| Get | GET | `/expenses/{expenseId}` | |
| Create | POST | `/expenses` | Required: `projectId`, `userId`, `date`, `total`, `categoryId` |
| Update | PUT | `/expenses/{expenseId}` | Full replace |
| Delete | DELETE | `/expenses/{expenseId}` | |

**Expense categories (official):** separate endpoint `/expenses/categories`. Spec `--type TRAVEL|MEAL|...` maps to categories if present; otherwise `--category` accepts category ID or name via `resolveExpenseCategoryByName`.

**Response fields:** `total` (not `amount`), `quantity`, `billable`, `notes`, `categoryId`, `fileId`. Display struct flattens to human-friendly columns.

**Spec create:** `--project --date --amount --type` — CLI:
- `--amount` → body `total` (float)
- `--type` → category resolution; `--category` overrides
- `--user` optional (defaults to current user via `GetCurrentUser`)
- `--description` → `notes`

**Out of M7 scope:** categories CRUD, receipt file download, expense reports (M6 has report API).

### 3.5 Custom-field — §8.16

| Op | Method | Path | Notes |
|----|--------|------|-------|
| List | GET | `/custom-fields` | Query: `name`, `status` |
| Create | POST | `/custom-fields` | Required: `name`, `type`; optional `allowedValues`, `status`, `onlyAdminCanEdit`, `placeholder`, `workspaceDefaultValue` |
| Update | PUT | `/custom-fields/{customFieldId}` | Body: `name` required; other fields optional |
| Delete | DELETE | `/custom-fields/{customFieldId}` | |

**Official type enum (API):** `TXT`, `NUMBER`, `DROPDOWN_SINGLE`, `DROPDOWN_MULTIPLE`, `CHECKBOX`, `LINK`

**Spec type list:** `DROPDOWN`, `CHECKBOX`, `TEXT`, `NUMBER`, `LINK`, `DATE`

**CLI mapping (spec name → API type):**

| Spec / CLI flag | API type |
|-----------------|----------|
| `TEXT` | `TXT` |
| `NUMBER` | `NUMBER` |
| `DROPDOWN` | `DROPDOWN_SINGLE` (default) |
| `DROPDOWN_SINGLE` | `DROPDOWN_SINGLE` |
| `DROPDOWN_MULTIPLE` | `DROPDOWN_MULTIPLE` |
| `CHECKBOX` | `CHECKBOX` |
| `LINK` | `LINK` |
| `DATE` | **unsupported** — error with explanation (not in official enum) |

- `--options "a,b,c"` → `allowedValues` (required for DROPDOWN_*)
- `--status`: `VISIBLE` \| `INVISIBLE` \| `INACTIVE`
- Validate type before API call

**Out of M7 scope:** project-level custom field attach/remove/patch.

### 3.6 Webhook — §8.17

| Op | Method | Path | Notes |
|----|--------|------|-------|
| List | GET | `/webhooks` | |
| Get | GET | `/webhooks/{webhookId}` | |
| Create | POST | `/webhooks` | Body includes `url`, `webhookEvent`, `enabled`, optional `name`, `triggerSourceType`, `triggerSource[]` |
| Update | PUT | `/webhooks/{webhookId}` | Can toggle `enabled` |
| Delete | DELETE | `/webhooks/{webhookId}` | |

**Official event enum** (superset of spec — validate against this):

```
NEW_PROJECT, PROJECT_UPDATED, PROJECT_DELETED,
NEW_TASK, TASK_UPDATED, TASK_DELETED,
NEW_CLIENT, CLIENT_UPDATED, CLIENT_DELETED,
NEW_TIMER_STARTED, TIMER_STOPPED,
NEW_TIME_ENTRY, TIME_ENTRY_UPDATED, TIME_ENTRY_DELETED,
TIME_ENTRY_SPLIT, TIME_ENTRY_RESTORED,
NEW_TAG, TAG_UPDATED, TAG_DELETED,
USER_JOINED_WORKSPACE, USER_DELETED_FROM_WORKSPACE,
USER_DEACTIVATED_ON_WORKSPACE, USER_ACTIVATED_ON_WORKSPACE,
USER_EMAIL_CHANGED, USER_UPDATED,
NEW_INVOICE, INVOICE_UPDATED,
NEW_APPROVAL_REQUEST, APPROVAL_REQUEST_STATUS_UPDATED,
TIME_OFF_REQUESTED, TIME_OFF_REQUEST_UPDATED,
TIME_OFF_REQUEST_APPROVED, TIME_OFF_REQUEST_REJECTED,
TIME_OFF_REQUEST_STARTED, TIME_OFF_REQUEST_WITHDRAWN,
BALANCE_UPDATED,
EXPENSE_CREATED, EXPENSE_UPDATED, EXPENSE_DELETED, EXPENSE_RESTORED,
ASSIGNMENT_CREATED, ASSIGNMENT_UPDATED, ASSIGNMENT_DELETED, ASSIGNMENT_PUBLISHED,
USER_GROUP_CREATED, USER_GROUP_UPDATED, USER_GROUP_DELETED,
USERS_INVITED_TO_WORKSPACE, LIMITED_USERS_ADDED_TO_WORKSPACE,
COST_RATE_UPDATED, BILLABLE_RATE_UPDATED
```

- Spec `--events "TIME_ENTRY_CREATED,TIME_ENTRY_UPDATED"` — `TIME_ENTRY_CREATED` is **not** in official enum; closest is `NEW_TIME_ENTRY`. CLI validates; error lists valid events. Map spec aliases where sensible (`TIME_ENTRY_CREATED` → `NEW_TIME_ENTRY` with warning).
- Spec `--active` / `--inactive` → body `enabled: true/false`
- Spec `--url` required on create

**Out of M7 scope:** logs, statuses, token regenerate, addon webhooks.

### 3.7 Approval — §8.18

| Op | Method | Path | Notes |
|----|--------|------|-------|
| List | GET | `/approval-requests` | Query: `status`, `types`, page params. Status: `PENDING`, `APPROVED`, `WITHDRAWN_APPROVAL` |
| Submit (self) | POST | `/approval-requests/{type}` | `type` = `TIMESHEET` \| `EXPENSE`. Body: `period`, `periodStart` |
| Submit (other) | POST | `/approval-requests/users/{userId}/{type}` | Same body |
| Update | PATCH | `/approval-requests/{approvalRequestId}` | Body: `{state, note}`. State: `PENDING`, `APPROVED`, `WITHDRAWN_SUBMISSION`, `WITHDRAWN_APPROVAL`, `REJECTED` |

**Spec vs API mismatch (important):**
- Spec: `approval submit --time-entry <id1> --time-entry <id2>`
- Official: submit is **period-based**, not per-entry-ID. Body uses `period` (e.g. `WEEKLY`, `MONTHLY`) + `periodStart`.
- CLI implements official model:
  - `approval submit --type TIMESHEET --period WEEKLY --start 2026-09-29`
  - Optional `--user` for other-user submit
  - Optional `--note`
  - If user passes repeated `--time-entry`, return clear error pointing to period-based submit (or implement resubmit-entries path in a later milestone)
- Spec `approval update --status APPROVED|REJECTED` → body `state`
- Spec list `--status PENDING` works with official status enum; also allow `WITHDRAWN_APPROVAL`
- Spec `--user` filter: official list has no userId query — CLI filters client-side after fetch, or uses resubmit path if needed. Prefer: list all + client-side filter with note in help text.

**Out of M7 scope:** resubmit-entries-for-approval, deprecated approval endpoints.

### 3.8 Group — §8.19

| Op | Method | Path | Notes |
|----|--------|------|-------|
| List | GET | `/user-groups` | Query: `name`, `project-id`, page params. Response: array of `{id, name, userIds[], teamManagers[]}` |
| Get | GET | `/user-groups/{id}` | Path param name is `id` |
| Create | POST | `/user-groups` | Body: `{name}` (members added after) |
| Update | PUT | `/user-groups/{id}` | Body: `{name}` |
| Delete | DELETE | `/user-groups/{id}` | |
| Add member | POST | `/user-groups/{userGroupId}/users` | Body: `{userId}` |
| Remove member | DELETE | `/user-groups/{userGroupId}/users/{userId}` | |

**CLI mapping:**
- Spec uses `--add-member` / `--remove-member` on `group update` — implement as flags that call add/remove endpoints after (or instead of) name update
- Spec `group create --members id1,id2` — create group then loop add-member calls
- Path is `user-groups` not `groups` — command name stays `clockify group` per spec
- Resolve users via `resolveUserRef`
- Fuzzy name resolution via `resolveGroupByName` (list + match)

---

## 4. Sprint 1: Standard CRUD Batch (Days 1–2)

### User Stories

**US-1.1: Entity Types & Shared Validation**

> As a developer, I want typed structs and validation helpers for M7 resources so commands stay thin and consistent.

**Acceptance Criteria:**
- Types in `pkg/clockify/` (new file `m7.go` preferred over growing `types.go`)
- Request structs for create/update per resource
- Enum maps: custom-field types, webhook events, approval states, invoice status, time-off status
- Mapping helpers: `MapCustomFieldType(spec string) (string, error)`, `MapInvoiceStatus(spec string) (string, error)`, webhook event alias map

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.1 | Define types: Invoice, ScheduleAssignment, TimeOffRequest, Expense, CustomField, Webhook, ApprovalRequest, UserGroup | 2h | `pkg/clockify/m7.go` | `go build ./pkg/clockify/` |
| T1.2 | Define Create/Update request structs for all 8 | 1.5h | `pkg/clockify/m7.go` | `go build` |
| T1.3 | Enum validation maps + mapping helpers (spec→API) | 1.5h | `pkg/clockify/m7_enums.go` | Unit tests pass |
| T1.4 | Unit tests for enums/mappings | 1h | `pkg/clockify/m7_enums_test.go` | `go test ./pkg/clockify/` |

**US-1.2: API Methods — Low-Risk Resources**

> As a user, I want full CRUD for group, webhook, custom-field, expense, and invoice via typed API methods.

**Acceptance Criteria:**
- Methods in `internal/api/m7_resources.go`
- Handle non-array list responses (invoice `{invoices,total}`, expense `{expenses,...}`)
- Use `Do()` / `DoWithPagination()` where applicable

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.5 | Group: List, Get, Create, Update, Delete, AddUser, RemoveUser | 1.5h | `internal/api/m7_resources.go` | httptest |
| T1.6 | Webhook: List, Get, Create, Update, Delete | 1h | same | httptest |
| T1.7 | CustomField: List, Create, Update, Delete | 1h | same | httptest |
| T1.8 | Expense: List, Get, Create, Update, Delete + ListExpenseCategories | 1.5h | same | httptest |
| T1.9 | Invoice: List, Get, Create, Update, Delete, ChangeStatus | 1.5h | same | httptest |
| T1.10 | API unit tests for all S1 methods | 2h | `internal/api/m7_resources_test.go` | `go test ./internal/api/` |

**US-1.3: CLI Commands — Group**

> As a user, I want `clockify group` CRUD with member management.

**Acceptance Criteria:**
- `group list` (`--search`/`--name` filter), `group get --id/--name`
- `group create --name [--members u1,u2] [--file|--stdin] [--if-not-exists]`
- `group update --id/--name [--new-name] [--add-member u] [--remove-member u]`
- `group delete --id/--name` with confirmation
- Name resolution + `--confirm` + all output formats
- Display: ID, NAME, MEMBER_COUNT, MEMBERS (comma-joined emails/names when resolvable)

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.11 | `internal/cli/group.go` — all subcommands | 2h | `internal/cli/group.go` | Manual + unit |
| T1.12 | `resolveGroupByName` | 0.5h | same | Unit test |
| T1.13 | `group_test.go` — flags, validation, resolution | 1.5h | `internal/cli/group_test.go` | `go test ./internal/cli/` |
| T1.14 | Register `newGroupCmd()` in `root.go` | 0.25h | `internal/cli/root.go` | `clockify --help` |

**US-1.4: CLI Commands — Webhook**

**Acceptance Criteria:**
- `webhook list`, `webhook get`, `webhook create --url --events [--name]`, `webhook update --id [--active|--inactive|--url]`, `webhook delete`
- Events validated against official enum + aliases
- Create output includes ID/URL/enabled (mask authToken if present)

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.15 | `internal/cli/webhook.go` | 1.5h | same | Manual + unit |
| T1.16 | `webhook_test.go` | 1h | same | tests |
| T1.17 | Register in `root.go` | 0.25h | `root.go` | help |

**US-1.5: CLI Commands — Custom Field**

**Acceptance Criteria:**
- `custom-field list [--search] [--status]`
- `custom-field create --name --type [--options] [--status]`
- `custom-field update --id/--name --name [--type]...`
- `custom-field delete`
- Type mapping TEXT→TXT, DROPDOWN→DROPDOWN_SINGLE; DATE rejected with message

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.18 | `internal/cli/custom_field.go` | 1.5h | same | Manual + unit |
| T1.19 | `custom_field_test.go` | 1h | same | tests |
| T1.20 | Register in `root.go` | 0.25h | `root.go` | help |

**US-1.6: CLI Commands — Expense**

**Acceptance Criteria:**
- `expense list [--project] [--start] [--end] [--user]`
- `expense get --id`
- `expense create --project --date --amount --type/--category [--description] [--user] [--file|--stdin]`
- `expense update --id [--amount|--category|--description...]`
- `expense delete` with confirmation
- Display uses `total`/`notes`/category ID

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.21 | `internal/cli/expense.go` | 2h | same | Manual + unit |
| T1.22 | `resolveExpenseCategoryByName` | 0.5h | same | Unit test |
| T1.23 | `expense_test.go` | 1.5h | same | tests |
| T1.24 | Register in `root.go` | 0.25h | `root.go` | help |

**US-1.7: CLI Commands — Invoice**

**Acceptance Criteria:**
- `invoice list`, `invoice get --id`
- `invoice create --client --start --end --currency [--number] [--status] [--note] [--file|--stdin]`
- `invoice update --id [fields...]`
- `invoice delete` with confirmation
- `invoice update --id --status SENT` uses PATCH status endpoint (or maps to update body if simpler — prefer PATCH)
- Status mapping documented in help text

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.25 | `internal/cli/invoice.go` | 2h | same | Manual + unit |
| T1.26 | Client resolution via `resolveClientByName` | 0.25h | same | Unit |
| T1.27 | `invoice_test.go` | 1.5h | same | tests |
| T1.28 | Register in `root.go` | 0.25h | `root.go` | help |

**US-1.8: Pagination Helpers for S1 Lists**

> As a user, I want list commands to auto-fetch all pages like M2.

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.29 | `fetchPaginatedGroups`, `fetchPaginatedWebhooks`, `fetchPaginatedCustomFields`, `fetchPaginatedExpenses`, `fetchPaginatedInvoices` | 1.5h | `internal/cli/pagination_m7.go` or extend `pagination.go` | Unit tests |
| T1.30 | Wire `--page/--page-size/--no-pagination` into S1 list commands | 0.5h | CLI files | Manual |

**Sprint 1 Exit Criteria:**
- [ ] `clockify group|webhook|custom-field|expense|invoice` all have list/get/create/update/delete (as applicable)
- [ ] Group member add/remove works
- [ ] Webhook events validated; custom-field types mapped
- [ ] Invoice status mapping documented and tested
- [ ] All S1 unit tests pass; `go test ./...` green
- [ ] Commands appear in `clockify --help`

**Sprint 1 Estimated:** ~24h (2 focused days)

---

## 5. Sprint 2: Workflow Resources + E2E (Days 3–4)

### User Stories

**US-2.1: Time-off**

> As a user, I want to list, create, and approve/reject time-off requests.

**Acceptance Criteria:**
- `time-off list [--start] [--end] [--status] [--user]` using POST body
- `time-off create --policy <id|name> --start --end [--note]`
- `time-off update-status --id --status APPROVED|REJECTED --policy <id> [--note]`
- Optional `time-off delete --id --policy` with confirmation
- `--type` flag attempts policy resolution; clear error if no matching policy
- Status enum validated before API calls

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.1 | API: ListTimeOffRequests (POST), CreateTimeOffRequest, UpdateTimeOffStatus, DeleteTimeOffRequest, ListTimeOffPolicies (for resolution) | 2h | `internal/api/m7_resources.go` | httptest |
| T2.2 | `resolvePolicyByName` / `resolvePolicyByType` | 1h | `internal/cli/time_off.go` | Unit |
| T2.3 | `internal/cli/time_off.go` — list/create/update-status/delete | 2h | same | Manual + unit |
| T2.4 | `time_off_test.go` | 1.5h | same | tests |
| T2.5 | Register in `root.go` | 0.25h | `root.go` | help |

**US-2.2: Schedule**

> As a user, I want to list and manage project assignments for a date range.

**Acceptance Criteria:**
- `schedule list --start --end [--user] [--project]` (date range required; default `this-month` if omitted — document)
- `schedule create --user --project --start --end --hours-per-day` (or `--duration 8h` → hoursPerDay)
- `schedule update --id --hours-per-day|--start|--end... [--series THIS_ONE|THIS_AND_FOLLOWING|ALL]`
- `schedule delete --id [--series ...]` with confirmation
- Resolve user/project refs; validate hours > 0
- Display flattened hydrated fields (projectName, userName, hoursPerDay, billable)

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.6 | API: ListAssignments, CreateAssignment, UpdateAssignment, DeleteAssignment | 2h | `internal/api/m7_resources.go` | httptest |
| T2.7 | `internal/cli/schedule.go` — all subcommands + series option flag | 2.5h | same | Manual + unit |
| T2.8 | `schedule_test.go` | 1.5h | same | tests |
| T2.9 | Register in `root.go` | 0.25h | `root.go` | help |

**US-2.3: Approval**

> As a user, I want to list approval requests, submit period approvals, and approve/reject them.

**Acceptance Criteria:**
- `approval list [--status] [--type] [--user]` (user filter client-side)
- `approval submit --type TIMESHEET|EXPENSE --period WEEKLY|MONTHLY --start <date> [--user u] [--note]`
- `approval update --id --status APPROVED|REJECTED [--note]`
- Status enum: PENDING, APPROVED, REJECTED (+ WITHDRAWN_* accepted on list)
- Help text explains period-based model (spec's `--time-entry` is not supported by official API)
- If `--time-entry` flag is present, error with explanation

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.10 | API: ListApprovalRequests, SubmitApprovalRequest, SubmitApprovalRequestForUser, UpdateApprovalStatus | 2h | `internal/api/m7_resources.go` | httptest |
| T2.11 | `internal/cli/approval.go` — list/submit/update | 2h | same | Manual + unit |
| T2.12 | `approval_test.go` | 1.5h | same | tests |
| T2.13 | Register in `root.go` | 0.25h | `root.go` | help |

**US-2.4: E2E Verification & Docs**

> As a maintainer, I want M7 verified end-to-end and PLAN.md updated.

**Acceptance Criteria:**
- All success criteria from PLAN.md §M7 checked off (unit-level where no credentials; live E2E when available)
- `go test ./...` passes
- `go build` / lint clean
- PLAN.md M7 section marked complete; sprint outcomes recorded
- Git tag `v0.8.0-m7` after commit (also tag `v0.7.0-m6` if still missing — do this first)

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.14 | Housekeeping: `git tag v0.7.0-m6 817c831` | 0.1h | git | `git tag -l` |
| T2.15 | Full unit test suite + fix failures | 2h | all | `go test ./...` |
| T2.16 | CLI smoke: help output for all 8 parents; flag validation errors | 1h | binary | Manual |
| T2.17 | Update PLAN.md M7 status, deliverables, decisions, timeline | 1h | `PLAN.md` | Review |
| T2.18 | Live E2E (when credentials available): one happy path per resource | 2h | — | Checklist |
| T2.19 | Commit + tag `v0.8.0-m7` | 0.25h | git | Tag exists |

**Sprint 2 Exit Criteria:**
- [ ] All 8 resource command families complete per scope above
- [ ] Time-off status updates work (APPROVED, REJECTED)
- [ ] Group member management works
- [ ] Webhook events validated on create
- [ ] Custom field types validated
- [ ] Approval period submit + status update work
- [ ] Schedule list requires date range; create/update/delete work
- [ ] `go test ./...` green; PLAN.md updated; tags created

**Sprint 2 Estimated:** ~20h (2 focused days)

---

## 6. Spec ↔ API Decision Log

| Topic | Spec Says | Official API | CLI Decision |
|-------|-----------|--------------|--------------|
| Custom field types | TEXT, DATE, DROPDOWN | TXT, DROPDOWN_SINGLE/MULTIPLE, no DATE | Map TEXT→TXT; DROPDOWN→DROPDOWN_SINGLE; reject DATE with message |
| Webhook events | TIME_ENTRY_CREATED etc. | NEW_TIME_ENTRY etc. | Validate official enum; alias TIME_ENTRY_CREATED→NEW_TIME_ENTRY |
| Invoice status | DRAFT, FINALIZED, ACCEPTED | UNSENT, SENT, PAID, PARTIALLY_PAID, VOID, OVERDUE | Accept both sets; map where possible; document |
| Invoice create | --start --end | issuedDate, dueDate + currency + number | Map start→issuedDate, end→dueDate; require currency/number |
| Time-off create | --type VACATION | policyId path required | Require --policy; --type resolves policy if match |
| Time-off list | GET | POST body | Implement POST |
| Time-off delete | "no delete per spec" | DELETE exists | Implement delete with confirmation |
| Approval submit | --time-entry IDs | period + periodStart body | Implement official model; reject --time-entry with explanation |
| Schedule list | optional dates | start+end required | Require dates; default this-month |
| Group members | --add-member on update | dedicated POST/DELETE endpoints | Flags trigger member endpoints |
| Group path | `group` | `user-groups` | Command name `group`; path `user-groups` |
| Expense amount | --amount | total field | Map --amount→total |

---

## 7. Risk Mitigation

| Risk | Impact | Likelihood | Mitigation | Sprint |
|------|--------|------------|------------|--------|
| **Invoice create required fields incomplete** | Medium | High | Document required body; auto-number if omitted; fail fast on missing currency | S1 |
| **Time-off policy resolution UX** | Medium | High | Clear error message listing how to find policies; `time-off` could list policies as future command | S2 |
| **Approval model mismatch confuses users** | Medium | Medium | Help text + error message pointing to period-based submit | S2 |
| **Schedule series update options** | Low | Medium | Expose `--series` flag; default `THIS_ONE` | S2 |
| **List responses are wrapped objects not arrays** | Medium | Medium | Typed response wrappers (InvoiceList, ExpenseList) — don't assume bare arrays | S1 |
| **Plan-gated features (invoice, approval, webhook)** | Medium | Medium | Surface API errors clearly; unit tests don't require paid plan | S1/S2 |
| **Custom field DATE unsupported** | Low | High | Explicit error listing supported types | S1 |
| **Webhook event list drift** | Low | Low | Single enum map in `m7_enums.go`; unit test covers aliases | S1 |
| **Scope creep (items, payments, policies)** | High | Medium | Hard cut list below; defer to M10+/backlog | All |

---

## 8. Trade-offs: Where to Cut Corners vs. Where to Hold Firm

### Hold Firm (Non-Negotiable)

| Area | Why |
|------|-----|
| **Confirmation on delete** | Destructive ops must not surprise users |
| **Enum validation before API calls** | Fail fast; actionable errors |
| **All output formats per command** | Consistent UX from M2 |
| **`--file` / `--stdin` on create/update** | Spec + established pattern |
| **Pagination on list commands** | Consistency with M2 |
| **Name/ID resolution (`--id` or `--name`)** | Core UX from M4 |
| **Spec↔API mapping documented in help** | Prevents user confusion |

### Cut Corners (If Behind Schedule)

| Area | How to Cut | Impact |
|------|------------|--------|
| **Invoice items/payments** | Already out of scope | Fine |
| **Invoice auto-number** | Require `--number` instead of auto-gen | Slightly less convenient |
| **Expense categories CRUD** | Only resolve existing categories | Users manage categories in UI |
| **Schedule publish/copy** | Out of scope | Fine |
| **Approval resubmit-entries** | Out of scope | Users resubmit in UI |
| **Group `--if-not-exists`** | Defer if time-constrained | Minor |
| **Webhook triggerSource filters** | Create without filters first | Less precise hooks |

### Cut Corners — Emergency Only

| Area | How to Cut | Impact |
|------|------------|--------|
| **Invoice status PATCH** | Fold into PUT update body | Minor |
| **Time-off delete** | Defer entirely | Spec omitted it anyway |
| **Schedule default date range** | Require explicit --start/--end | Slightly more typing |
| **Live E2E** | Unit/httptest only | PLAN notes live E2E still recommended |
| **CSV/YAML on new commands** | JSON + table only for emergency | Power users use jq |

---

## 9. File-Level Task Summary

| File | Sprint | Tasks | Est. Hours |
|------|--------|-------|------------|
| `pkg/clockify/m7.go` | S1 | T1.1–T1.2 | 3.5h |
| `pkg/clockify/m7_enums.go` + test | S1 | T1.3–T1.4 | 2.5h |
| `internal/api/m7_resources.go` + test | S1/S2 | T1.5–T1.10, T2.1, T2.6, T2.10 | 12.5h |
| `internal/cli/group.go` + test | S1 | T1.11–T1.13 | 4h |
| `internal/cli/webhook.go` + test | S1 | T1.15–T1.16 | 2.5h |
| `internal/cli/custom_field.go` + test | S1 | T1.18–T1.19 | 2.5h |
| `internal/cli/expense.go` + test | S1 | T1.21–T1.23 | 4h |
| `internal/cli/invoice.go` + test | S1 | T1.25–T1.27 | 3.75h |
| `internal/cli/pagination_m7.go` | S1 | T1.29–T1.30 | 2h |
| `internal/cli/time_off.go` + test | S2 | T2.2–T2.4 | 4.5h |
| `internal/cli/schedule.go` + test | S2 | T2.7–T2.8 | 4h |
| `internal/cli/approval.go` + test | S2 | T2.11–T2.12 | 3.5h |
| `internal/cli/root.go` | S1/S2 | T1.14, T1.17, T1.20, T1.24, T1.28, T2.5, T2.9, T2.13 | 2h |
| `PLAN.md` | S2 | T2.17 | 1h |
| E2E + tag | S2 | T2.14–T2.16, T2.18–T2.19 | 3.35h |

**Total estimated:** ~56h ≈ **2 focused sprints (4 days)** + buffer for live E2E.

---

## 10. Type Extensions (Draft)

```go
// pkg/clockify/m7.go

type Invoice struct {
    ID           string `json:"id" yaml:"id"`
    WorkspaceID  string `json:"workspaceId" yaml:"workspaceId"`
    ClientID     string `json:"clientId" yaml:"clientId"`
    Number       string `json:"number" yaml:"number"`
    Currency     string `json:"currency" yaml:"currency"`
    IssuedDate   string `json:"issuedDate" yaml:"issuedDate"`
    DueDate      string `json:"dueDate" yaml:"dueDate"`
    InvoiceStatus string `json:"invoiceStatus" yaml:"invoiceStatus"` // or status
    Note         string `json:"note" yaml:"note"`
    Total        float64 `json:"total" yaml:"total"`
}

type ScheduleAssignment struct {
    ID            string  `json:"id" yaml:"id"`
    WorkspaceID   string  `json:"workspaceId" yaml:"workspaceId"`
    ProjectID     string  `json:"projectId" yaml:"projectId"`
    ProjectName   string  `json:"projectName" yaml:"projectName"`
    TaskID        string  `json:"taskId" yaml:"taskId"`
    TaskName      string  `json:"taskName" yaml:"taskName"`
    UserID        string  `json:"userId" yaml:"userId"`
    UserName      string  `json:"userName" yaml:"userName"`
    ClientID      string  `json:"clientId" yaml:"clientId"`
    ClientName    string  `json:"clientName" yaml:"clientName"`
    HoursPerDay   float64 `json:"hoursPerDay" yaml:"hoursPerDay"`
    StartTime     string  `json:"startTime" yaml:"startTime"`
    Billable      bool    `json:"billable" yaml:"billable"`
    Note          string  `json:"note" yaml:"note"`
    Published     bool    `json:"published" yaml:"published"`
}

type TimeOffRequest struct {
    ID            string `json:"id" yaml:"id"`
    WorkspaceID   string `json:"workspaceId" yaml:"workspaceId"`
    UserID        string `json:"userId" yaml:"userId"`
    PolicyID      string `json:"policyId" yaml:"policyId"`
    Start         string `json:"start" yaml:"start"`
    End           string `json:"end" yaml:"end"`
    Status        string `json:"status" yaml:"status"`
    Note          string `json:"note" yaml:"note"`
}

type Expense struct {
    ID          string  `json:"id" yaml:"id"`
    WorkspaceID string  `json:"workspaceId" yaml:"workspaceId"`
    ProjectID   string  `json:"projectId" yaml:"projectId"`
    TaskID      string  `json:"taskId" yaml:"taskId"`
    UserID      string  `json:"userId" yaml:"userId"`
    CategoryID  string  `json:"categoryId" yaml:"categoryId"`
    Date        string  `json:"date" yaml:"date"`
    Total       float64 `json:"total" yaml:"total"`
    Quantity    float64 `json:"quantity" yaml:"quantity"`
    Billable    bool    `json:"billable" yaml:"billable"`
    Notes       string  `json:"notes" yaml:"notes"`
    FileID      string  `json:"fileId" yaml:"fileId"`
}

type CustomField struct {
    ID            string   `json:"id" yaml:"id"`
    WorkspaceID   string   `json:"workspaceId" yaml:"workspaceId"`
    Name          string   `json:"name" yaml:"name"`
    Type          string   `json:"type" yaml:"type"` // TXT, NUMBER, DROPDOWN_SINGLE, ...
    Status        string   `json:"status" yaml:"status"`
    AllowedValues []string `json:"allowedValues" yaml:"allowedValues"`
    Description   string   `json:"description" yaml:"description"`
    OnlyAdminCanEdit bool  `json:"onlyAdminCanEdit" yaml:"onlyAdminCanEdit"`
}

type Webhook struct {
    ID             string   `json:"id" yaml:"id"`
    WorkspaceID    string   `json:"workspaceId" yaml:"workspaceId"`
    Name           string   `json:"name" yaml:"name"`
    URL            string   `json:"url" yaml:"url"`
    WebhookEvent   string   `json:"webhookEvent" yaml:"webhookEvent"`
    Enabled        bool     `json:"enabled" yaml:"enabled"`
    DeliveryEnabled bool    `json:"deliveryEnabled" yaml:"deliveryEnabled"`
    UserID         string   `json:"userId" yaml:"userId"`
    // AuthToken present on create response — never print in table output
}

type ApprovalRequest struct {
    ID          string   `json:"id" yaml:"id"`
    WorkspaceID string   `json:"workspaceId" yaml:"workspaceId"`
    UserID      string   `json:"userId" yaml:"userId"`
    Status      string   `json:"status" yaml:"status"`
    Type        string   `json:"type" yaml:"type"`
    Start       string   `json:"start" yaml:"start"`
    End         string   `json:"end" yaml:"end"`
    Note        string   `json:"note" yaml:"note"`
}

type UserGroup struct {
    ID           string   `json:"id" yaml:"id"`
    WorkspaceID  string   `json:"workspaceId" yaml:"workspaceId"`
    Name         string   `json:"name" yaml:"name"`
    UserIDs      []string `json:"userIds" yaml:"userIds"`
    TeamManagers []struct {
        ID   string `json:"id" yaml:"id"`
        Name string `json:"name" yaml:"name"`
    } `json:"teamManagers" yaml:"teamManagers"`
}
```

*Exact JSON field names must be confirmed against live API responses / docs samples during implementation; use flexible unmarshaling for list wrappers.*

---

## 11. M7 Overall Definition of Done

| Criterion | Verification | Status |
|-----------|-------------|--------|
| Tag `v0.7.0-m6` exists | `git tag -l` | ☐ |
| `clockify group` full CRUD + members | create → add-member → remove-member → delete | ☐ |
| `clockify webhook` CRUD + event validation | invalid event errors; valid create works | ☐ |
| `clockify custom-field` CRUD + type mapping | TEXT accepted; DATE rejected; DROPDOWN needs options | ☐ |
| `clockify expense` CRUD + filters | list --project/--start/--end; create --amount→total | ☐ |
| `clockify invoice` CRUD + status | create with client/dates/currency; update --status | ☐ |
| `clockify time-off` list/create/update-status | POST list; policy-scoped create; APPROVED/REJECTED | ☐ |
| `clockify schedule` list/create/update/delete | date range required; hoursPerDay parse; series flag | ☐ |
| `clockify approval` list/submit/update | period-based submit; state update | ☐ |
| All 8 appear in `clockify --help` | Visual check | ☐ |
| `--file` / `--stdin` on creates | Per-resource | ☐ |
| `--confirm` required for deletes | Per-resource | ☐ |
| `--output json` pipeline works | jq on list/get outputs | ☐ |
| Unit tests pass | `go test ./...` | ☐ |
| PLAN.md updated | M7 marked complete; decisions logged | ☐ |
| Tag `v0.8.0-m7` created | `git tag -l` | ☐ |

---

## 12. Out of Scope (M7 Hard Cuts)

- Invoice items, payments, duplicate, export, settings
- Expense categories CRUD (list/resolve only), receipt download
- Custom fields on projects/users/time entries
- Webhook logs, token regenerate, addon webhooks
- Time-off policies CRUD, balances, balance assignments
- Approval resubmit-entries, deprecated approval endpoints
- Schedule publish, copy, capacity totals
- Plugin system (post-M10)

---

## 13. Daily Standup Template

```markdown
## Day X Standup — M7 Sprint Y

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

*This plan is a living document. Update task estimates and status as work progresses. Live API responses may adjust display field names — keep display structs tolerant.*

*Official docs.clockify.me verified 2026-10-04 for endpoint paths, enums, and required body fields. Spec §8.12–§8.19 discrepancies are logged in §6.*
