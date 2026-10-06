# Commit Review — 3168f3e

- **Commit:** 3168f3ebc2061ab6487c2378b33c5058320e179e
- **Subject:** feat: M7 - Secondary Resources (full API coverage)
- **Date:** 2026-10-04
- **Experts engaged:** Backend Engineer, Architect
- **Skill mode:** Auto-Assignment (kims-team)
- **Review date:** 2026-10-05
- **Related plan:** PLAN.md §5.7 Review Backlog

## Diff summary

Implements M7 Secondary Resources in one commit (+6043/−29 across 19 files). Adds:

- `pkg/clockify/m7.go` + `m7_enums.go`: types and enum/alias maps for invoice, schedule, time-off, expense, custom-field, webhook, approval, user-group (validation before network; DATE rejected; webhook official events + aliases; period-based approval; invoice status map)
- `internal/api/m7_resources.go`: official v1 methods for all 8 resources (policy-scoped time-off, `user-groups` paths, period-based approval submit, wrapped invoice/expense/time-off/approval list decoding)
- CLI: `invoice.go`, `schedule.go`, `time_off.go`, `expense.go`, `custom_field.go`, `webhook.go`, `approval.go`, `group.go` + `m7_helpers.go` + `pagination_m7.go`; root registration for the 8 commands
- Tests: `pkg/clockify/m7_enums_test.go`, `internal/api/m7_resources_test.go`, `internal/cli/m7_test.go` (help, validation, httptest path/CRUD)
- Docs: PLAN.md M7 completion + `M7-SPRINT-PLAN.md`

Matches the PLAN M7 deliverable table and success criteria at unit level. Commit message claims unit tests green; live plan-gated E2E remains recommended (sprint plan).

## Findings

### Critical

None. Endpoint paths, method verbs, and bodies align with the official docs corrections recorded in PLAN §M7 (user-groups paths, policy-scoped time-off, period-based approval, invoice required fields, custom-field type enum). Unit tests assert those paths. No auth/secret mishandling introduced. M7 success criteria (8 commands on help, group members, webhook event validation, time-off status, custom-field types) are covered at this SHA.

### Major

1. **Relative-date formatting marks local time as UTC (timezone bug)** — `internal/cli/m7_helpers.go` (`parseDateTimeCLI`, `parseDateOnlyCLI`). Relative inputs (`today`, `this-month`, …) come from `duration.ParseRelativeDate` as **local** `time.Time` values, but the CLI formats them with a literal `Z` suffix (`2006-01-02T15:04:05.000000Z`) and date-only paths then re-parse as UTC. For non-UTC users, schedule range endpoints, time-off start/end, invoice issued/due, and expense dates can shift by the local offset. This is the class of bug later noted in PLAN as “bare-date UTC bug” fixed by M11’s `ParseHumanDateTime`; **it is present and material at this SHA**. Matters because M7 is date-heavy and wrong day boundaries create bad invoices/time-off, not just cosmetic output.

2. **Pagination is incomplete / inconsistent across the M7 list surface** — `internal/cli/pagination_m7.go` + several list commands. (a) Single-page paths return `PageInfo{Page, PageSize}` **without** `TotalPages`/`TotalResults`, so `printPaginationSummary` never prints “Page X of Y” on the default path (same synthetic-paginaton class already flagged in M4/M5). (b) `schedule list` never reads `--page` / `--page-size` / `--no-pagination` (`internal/cli/schedule.go`) despite global flags existing — always a single unparameterized request. (c) `time-off list` and `approval list` also ignore page flags (`internal/cli/time_off.go`, `internal/cli/approval.go`), so large result sets truncate with no warning. (d) `fetchPaginatedExpenses`/`fetchPaginatedInvoices` `--no-pagination` loops still discard real API totals and rebuild synthetic `PageInfo`. Matters: commit message claims “pagination,” and users will assume full workspace lists.

3. **Name resolution only sees the first list page** — `internal/cli/m7_helpers.go` (`resolveGroupByName`, `resolvePolicyRef`, `resolveExpenseCategory`, `resolveCustomFieldByName`). Each lists once with empty/default opts (no walk of pages). Groups/policies/categories/custom-fields beyond page 1 cannot be resolved by name even when `--no-pagination` list works. Same architectural debt as earlier M4/M5 resolvers, but M7 multiplies call sites (time-off, expense, invoice client, groups, custom-fields). Matters for correctness of `--policy`/`--category`/`--name` on large workspaces.

4. **Invoice status alias map is lossy and surprising** — `pkg/clockify/m7_enums.go` (`invoiceStatusAliases`, `MapInvoiceStatus`). Spec names `DRAFT` and `FINALIZED` both map to `UNSENT`; `ACCEPTED` maps to `PAID`. Official enum is only `UNSENT|SENT|PARTIALLY_PAID|OVERDUE|PAID|VOID`. Documented in command help, but a user passing `DRAFT` or `FINALIZED` silently gets the same API state with no distinction, and `ACCEPTED` → `PAID` asserts a business transition the API never had. Matters: create/update can mutate invoice financial state based on a misleading CLI name.

### Minor

1. **Approval list client-side status filter can still call the API with non-enum values** — `internal/cli/approval.go` + `MapApprovalListStatus`. `REJECTED`/`NEW`/`IN_PROGRESS` are passed through to `GET /approval-requests` then filtered in CLI. If the API rejects unknown `status`, the command fails instead of listing + filtering; if it ignores the param, results depend on client filter only (and user filter is also client-side on a single page).
2. **`resolvePolicyRef` type-label match is substring** — `internal/cli/m7_helpers.go`. `--type SICK` uses `strings.Contains(policyName, "SICK")`, which can false-positive on names like “Sick Leave — Contractor” vs another policy, or match partial labels. Exact/fuzzy-name path is safer; type labels should be last-resort with explicit error.
3. **Expense update request is a full replace** — `pkg/clockify/m7.go` (`UpdateExpenseRequest`): required `projectID`/`userID`/`date`/`total`/`billable` on update. CLI partial flags that omit them may send zero values. Aligns with Clockify PUT semantics if confirmed live, but partial-update UX is weaker than invoice/custom-field updates.
4. **No live E2E for plan-gated resources** — sprint plan already flags invoice/approval/webhook as needing a paid workspace. Not scored against the SHA’s unit claims, but operational risk remains at ship time.
5. **`confirmDestructive` uses `fmt.Scanln`** — `internal/cli/m7_helpers.go`. Diverges from Cobra `cmd.In()`; non-interactive stdin may hang or mis-read (same pattern as earlier milestones).
6. **Large single-commit surface** — Architect lens: 8 resources + helpers + tests in one commit is consistent with the project’s velocity pattern and reuses M4/M5 templates, but increases blast radius if any path/enum is wrong; unit coverage mitigates but does not replace live verification.

## Recommendations

1. Fix M7 date parsing at the source: either convert relative/local times to UTC before formatting, or format without a false `Z` and set an explicit timezone; prefer adopting M11’s `pkg/duration.ParseHumanDateTime` grammar everywhere (time-entry + M7 date flags + reports). **Later milestones: M11 notes this grammar was adopted for M7 date flags — addressed after this SHA; this commit’s verdict still stands.**
2. Complete pagination honesty: set real totals when the API provides them (invoice wrapped `{invoices,total}`; expense daily/weekly wrappers if usable), wire `--page`/`--page-size`/`--no-pagination` on schedule/time-off/approval lists, and warn on truncation. **M10 path tests improve M7 Get/Update/Delete coverage; pagination UX was not fully reworked in later commits.**
3. Make M7 name resolvers paginate or reuse a shared “fetch all pages” helper; require IDs when resolution is ambiguous rather than failing on first-page-only lookup.
4. Tighten invoice status UX: either reject spec names that collapse (`DRAFT`/`FINALIZED`/`ACCEPTED`) with an error pointing at official enums, or document one-to-one in help and never map `ACCEPTED`→`PAID` silently.
5. Approval list: stop sending non-enum status to the API; always fetch with a valid enum (or unfiltered) and filter client-side after pagination, or document REJECTED-only local filter clearly.
6. Optional follow-up: live harness cases for policy-scoped time-off and period-based approval (M10 env-gated integration can extend).

## Conflicts / unresolved

No material conflict between the Backend Engineer and Architect lenses (in-session review; no long-running subagents).

- **Backend Engineer:** prioritizes Major #1 (timezone), Major #2/#3 (pagination + first-page resolvers), Major #4 (status alias correctness), Minor #1/#2 (approval filter + policy matching).
- **Architect:** agrees the layering (`pkg/clockify` types/enums → `internal/api` → thin CLI) and official path corrections are sound and consistent with M4/M5; flags Major #6 (bulk surface) and recommends consolidating pagination/resolver helpers instead of new one-off `pagination_m7.go` copies of known-bad patterns.

**Resolution (confidence-weighted):** both agree on the same verdict and the same priority order; Architect defers to Backend on API-correctness findings. Unresolved only for live regional behavior of invoice/approval/webhook (plan-gated) — same residual risk PLAN already records for M7; not an expert disagreement.

## Verdict

**approve with follow-ups**

M7 delivers full secondary-resource API coverage with correct official path corrections, enum validation-before-network, CRUD CLI, `--file`/`--stdin`/`--confirm`, and unit tests matching the M7 deliverable table. Timezone-broken date formatting, incomplete pagination (especially schedule/time-off/approval), first-page name resolution, and lossy invoice status aliases are real non-blocking gaps at this SHA; treat them as explicit follow-ups rather than a failed milestone. Later M11/M10 work improves some of these areas; this verdict is for commit `3168f3e` on its own merits.

## HEAD triage (2026-10-05)

- **Current HEAD:** `ac64914`
- **Original verdict:** unchanged (approve with follow-ups)
- **Legend:** ✅ completed at HEAD · ⚠️ partially fixed · ❌ still open · ➖ N/A or cannot verify

| # | Item | Status | Evidence |
|---|------|--------|----------|
| Major 1 | Timezone-broken date formatting (relative local time marked `Z`) | ❌ still open | `internal/cli/m7_helpers.go`: `parseDateTimeCLI` still formats with literal `Z` at lines 27, 34, 42 (`t.Format("2006-01-02T15:04:05.000000Z")`); `parseHumanStart` returns local TZ time (`m11_resolve.go:211-217`, `locationFromConfig` defaults to `time.Local`) but is then formatted with a false `Z`. `parseDateOnlyCLI` (lines 72-75) re-parses that string as UTC. Schedule/time-off/invoice/expense date flags still route through `parseDateTimeCLI`/`parseDateOnlyCLI`. |
| Major 2 | Pagination incomplete/inconsistent (schedule/time-off/approval; synthetic totals) | ❌ still open | (a) `internal/cli/pagination_m7.go` single-page paths still return `PageInfo{Page, PageSize}` without `TotalPages`/`TotalResults` (lines 15, 44, 73); `printPaginationSummary` (`m7_helpers.go:319`) never prints "Page X of Y" on that path. (b) `schedule list` (`schedule.go:28-115`) never reads `--page`/`--page-size`/`--no-pagination`; always one unparameterized `ListAssignments`. (c) `time_off.go:29-114` and `approval.go:28-123` list commands likewise ignore page flags. (d) `fetchPaginatedInvoices` discards real API pageInfo (`pagination_m7.go:149`, `_ = pageInfo`) and rebuilds synthetic totals; `fetchPaginatedExpenses` returns no PageInfo at all. Note: invoice/expense/group/webhook/custom-field lists do pass `page`/`pageSize` into their API calls (e.g. `invoice.go:41-51`, `expense.go:39-49`), but the sub-findings named in the original review remain. |
| Major 3 | Name resolution only sees first list page | ❌ still open | `internal/cli/m7_helpers.go`: `resolveGroupByName` (244), `resolvePolicyRef` (210), `resolveExpenseCategory` (185), `resolveCustomFieldByName` (269) each call a single list with empty/default opts — no page walk, no shared fetch-all helper. |
| Major 4 | Invoice status alias map lossy (`DRAFT`/`FINALIZED`→`UNSENT`, `ACCEPTED`→`PAID`) | ❌ still open | `pkg/clockify/m7_enums.go:43-48` `invoiceStatusAliases` unchanged; `MapInvoiceStatus` (58-64) still maps those aliases and its error message still lists spec names as "supported" (line 63). Tests still assert the lossy mappings (`m7_enums_test.go:41-44`). |
| Minor 1 | Approval list sends non-enum status to API then filters client-side | ❌ still open | `approval.go:47-52` still sets `opts.Status = mapped` where `MapApprovalListStatus` (`m7_enums.go:196-201`) returns `REJECTED`/`NEW`/`IN_PROGRESS` unchanged; client-side filter (approval.go:82-94) runs on a single unpaginated page. |
| Minor 2 | `resolvePolicyRef` type-label match is substring | ❌ still open | `m7_helpers.go:219-227` still uses `strings.Contains(strings.ToUpper(p.Name), upper)` for type labels (VACATION/SICK/…). |
| Minor 3 | Expense update request is a full replace | ⚠️ partially fixed | `pkg/clockify/m7.go:190-200` `UpdateExpenseRequest` still requires `projectID`/`userID`/`date`/`total`/`billable` (no `omitempty`). Mitigated at CLI: `expense update` fetches existing expense first and seeds the request from existing values (`expense.go:299-314`) before applying flag/file overrides, so omitted flags do not send zero values. |
| Minor 4 | No live E2E for plan-gated invoice/approval/webhook | ➖ cannot verify | Plan-gated paid-workspace resources; unit/httptest coverage only. No live harness added at HEAD. |
| Minor 5 | `confirmDestructive` uses `fmt.Scanln` | ❌ still open | `m7_helpers.go:147` still `fmt.Scanln(&response)`; diverges from Cobra `cmd.In()`. |
| Minor 6 | Large single-commit surface | ➖ N/A | Historical property of commit 3168f3e itself; not an actionable HEAD code gap. |
| Check | `--file`/`--stdin`/`--confirm` present on M7 commands | ✅ confirmed at HEAD | Present on invoice create/update/delete, expense create/update/delete, schedule create/delete, time-off create/delete, custom-field create/update/delete (e.g. `invoice.go:145-146,245-246,364,388`; `expense.go:155-156,268-269,396,423`; `schedule.go:132-133,318,344`; `time_off.go:128-129,288,324`; `custom_field.go:100-101,171-172,283,313`). |
| Check | Enum validation before network | ✅ confirmed at HEAD | `pkg/clockify/m7_enums.go` intact: `MapInvoiceStatus`, `MapTimeOffStatus`, `ValidateTimeOffType`, `MapApprovalState`/`MapApprovalListStatus`/`MapApprovalType`/`MapApprovalPeriod`, `MapSeriesUpdateOption`, `MapCustomFieldStatus`, `MapWebhookEvents`/`MapWebhookEvents` aliases — all validate/alias before any API call in CLI handlers. |

### Summary
At HEAD `ac64914` (M10), the four Major findings from the M7 review remain open. M11 wired `parseHumanStart`/`ParseHumanDateTime` into `parseDateTimeCLI`'s first branch, but the literal-`Z` formatting of local times (Major 1) is unchanged, so the timezone bug is still live. Pagination flag wiring exists on invoice/expense/group/webhook/custom-field lists, but schedule/time-off/approval lists still ignore page flags and single-page paths still synthesize `PageInfo` without real totals (Major 2). First-page name resolvers (Major 3) and the lossy invoice status alias map (Major 4) are untouched. Of the original verdict's praise points, `--file`/`--stdin`/`--confirm` and enum-validation-before-network both hold at HEAD. Only Minor 3 shows partial movement (CLI merge-with-existing on expense update).

### Completed at HEAD
- `--file`/`--stdin` JSON input on M7 create/update commands; `--confirm` on M7 delete commands (verified across invoice, expense, schedule, time-off, custom-field).
- Enum validation/aliasing before network for invoice status, time-off type/status, approval state/list-status/type/period, schedule series, custom-field status, webhook events.
- Invoice/expense/group/webhook/custom-field list commands pass `--page`/`--page-size`/`--no-pagination` into their list calls (partial improvement over the M4/M5 synthetic-paginaton pattern on the API-call side).
- Expense update CLI mitigates the full-replace request model by seeding the PUT body from the fetched existing expense.

### Still open at HEAD
- Major 1: timezone bug — relative/human local times still formatted with a false `Z` in `parseDateTimeCLI`/`parseDateOnlyCLI`.
- Major 2: pagination honesty — schedule/time-off/approval lists still unparameterized; single-page `PageInfo` still lacks `TotalPages`/`TotalResults`; invoice wrapped totals still discarded (`_ = pageInfo`); expenses still return no `PageInfo`.
- Major 3: all four M7 name resolvers still resolve only the first/default list page.
- Major 4: invoice status alias map still lossy (`DRAFT`/`FINALIZED`→`UNSENT`, `ACCEPTED`→`PAID`).
- Minor 1: approval list still sends `REJECTED`/`NEW`/`IN_PROGRESS` to `GET /approval-requests` then filters client-side on one page.
- Minor 2: `resolvePolicyRef` type-label match still substring (`strings.Contains`).
- Minor 5: `confirmDestructive` still uses `fmt.Scanln` instead of `cmd.In()`.
- Minor 4 (operational): no live E2E for plan-gated invoice/approval/webhook.

**Triage verdict:** original verdict stands — approve with follow-ups. At HEAD, all four Major findings and most Minor findings are still open; only Minor 3 and the `--file`/`--stdin`/`--confirm` + enum-validation checks are green.
