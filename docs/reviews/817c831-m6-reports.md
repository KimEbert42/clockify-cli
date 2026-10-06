# Commit Review — 817c831

- **Commit:** 817c8316308971806754a8c171383b3a21c37c82
- **Subject:** feat: M6 - Reports & Analytics
- **Date:** 2026-10-03
- **Experts engaged:** Backend Engineer, Full Stack Developer
- **Skill mode:** Auto-Assignment (kims-team)
- **Review date:** 2026-10-05
- **Related plan:** PLAN.md §5.7 Review Backlog

## Diff summary

Implements M6 Reports & Analytics against the Clockify Reports API. Adds `pkg/clockify/reports.go` (filter/request/response types), `internal/api/reports.go` (POST summary/detailed/weekly with retry, raw-body passthrough, CSV/PDF detection, flexible response parse), `internal/api/regions.go` `ReportsBaseURL` (global `reports.api.clockify.me/v1`; regional `{host}/report/v1`; `CLOCKIFY_REPORTS_URL` override), `internal/cli/report.go` (summary/detailed/weekly commands: relative dates, project/task/user/client/group filters, dual-client name resolution, `--export csv|pdf`, `--file` dual semantics), `internal/output/csv.go` yaml-tag mapping for struct→CSV, root registration, region/api/cli tests, and PLAN.md M6 completion notes. 10 files, +1718/−56. Matches M6 deliverable table and official API facts recorded in PLAN (regional path prefix, required per-type filter objects, approvalState mapping).

## Findings

### Critical

None. Endpoints, auth header, request bodies, and regional base URLs align with the documented Clockify Reports API and with unit tests (`internal/api/regions_test.go`, `internal/api/reports_test.go`). M6 success criteria (summary totals, detailed CSV export, weekly daily breakdown, billable filter) are covered by CLI tests at this SHA.

### Major

1. **`--file` dual meaning can overwrite the params input file** — `internal/cli/report.go` (`runReport`, export/Csv passthrough path). When `--export` is empty, `--file` is read as JSON report parameters. Later, if `api.LooksLikeCSV(contentType, body)` is true, the code writes the response body to `filePath` (`os.WriteFile(filePath, body, 0644)`). Same path variable serves two roles with no guard against clobbering the input. Mitigated by `Accept` preferring JSON, but any API/middlebox that returns CSV-looking bytes (or `CLOCKIFY_REPORTS_URL` pointing at a CSV fixture) silently destroys the params file. Spec §8.11 documents this dual meaning; the implementation needs a safe split (distinct flags or refuse write-back to a path that was just parsed as input).

2. **Params-file merge is incomplete relative to spec intent** — `internal/cli/report.go` (`readReportParamsFile`, `mergeReportFileParams`). PLAN/spec treat `--file` (no `--export`) as JSON report parameters, but merge only handles `billable`, `description`, `exportType`, `approvalState`, and `dateRangeType`. Structured filters (`projects`/`users`/`clients`/`tasks`/`tags`) and date overrides other than `start`/`end` are ignored without error. Users can ship a params file that looks complete while filters silently do nothing — wrong reports rather than failed ones.

3. **`postReport` duplicates the main client's retry/auth/error path** — `internal/api/reports.go` (`postReport`). A parallel implementation of rate-limit, retry backoff, header setup, and status mapping exists beside `Do()`/`resources.go`. Network failures are always wrapped as `errors.E005RateLimit` even when the failure is not rate limiting. Drift risk as the client evolves (pagination, proxy, logging) and misleading error codes for auth/DNS/proxy failures. Reports correctly cannot reuse JSON unmarshaling from `Do()`, but the transport/retry/error scaffolding should be shared.

4. **Report row pagination is not exposed** — `internal/cli/report.go` (`kind == "detailed"` sets `DetailedFilter{PageSize: 500}` only). Clockify detailed reports paginate via `detailedFilter.page`/`pageSize`; the CLI never sets `page`, never walks subsequent pages, and has no flag for page size. Large date ranges will silently return the first page only unless the API collapses everything into one response. Same class of gap as the unpaginated list resolvers already flagged in M4/M5 reviews — worth fixing here because reports are aggregate-facing.

### Minor

1. **`WeekStart` hardcoded to `MONDAY`** — `internal/cli/report.go` (`runReport`). Ignores workspace week-start preference; no CLI flag. Acceptable default, but weekly grouping will look wrong for workspaces using Sunday start.
2. **`timeZone` never set on `ReportRequest`** — types allow it; CLI formats dates without offset (`formatReportDate` → `YYYY-MM-DDTHH:MM:SS.ssssss`). Date-only parse yields UTC (`time.Parse`); relative dates use `pkg/duration` local calendar. Boundary days may shift for non-UTC users.
3. **`LooksLikeCSV` body heuristic is brittle** — `internal/api/reports.go`. Non-JSON text containing a comma and not starting with `{`/`[` is treated as CSV. Fine for the happy path; fragile under error pages or proxy HTML.
4. **`fromStruct` CSV writes `"<nil>"` for empty fields** — `internal/output/csv.go`. `fmt.Sprintf("%v", nil)` should be `""` for cleaner CSV consumers.
5. **`--status` vocabulary vs API enum** — `validReportStatuses` accepts `IN_PROGRESS`/`NEW`/`APPROVED`/`REJECTED`, then `mapReportStatus` collapses the first two to `UNAPPROVED` and rejects `REJECTED` with a clear error. Documented in PLAN as API truth; help text could say so up front.
6. **No live Reports-API integration test** — PLAN marks live E2E as still recommended; env-gated harness later (M10) is noted without re-scoring here.
7. **Filter resolution inherits unpaginated list calls** — `resolveProjectRef`/`resolveUserRef`/`resolveClientByName` (from M4/M5) resolve report filters from first-page lists only. Out of scope to re-score M4/M5, but M6 increases call sites.

## Recommendations

1. Guard or split `--file` semantics: refuse CSV/JSON write-back when the same path was consumed as params input, or add `--params` / keep `--file` only for export output.
2. Either implement full params-file field mapping (filters + dates) or reject unknown/ignored keys with an error; document the supported subset in help.
3. Extract shared transport helpers from `Do()` for report POSTs; classify network errors with the correct error code (not blanket `E005RateLimit`).
4. Add `--page`/`--page-size` for detailed reports (or auto-paginate until empty) and show a truncation warning when a non-final page is returned.
5. Optional: `--week-start` and a `timeZone` flag/`CLOCKIFY_TIMEZONE`; set `ReportRequest.TimeZone` when provided.
6. Replace `"<nil>"` with empty string in `output.CSVFormatter.fromStruct`.
7. Later milestones (M7/M10) should re-check whether detailed-report pagination and params-file completeness were addressed; this SHA's verdict stands regardless.

## Conflicts / unresolved

None between Backend Engineer and Full Stack Developer lenses: both agree the commit implements the documented API correctly and that the issues above are follow-ups, not blockers. Backend lens prioritizes shared transport/error classification and pagination; Full Stack lens prioritizes `--file` UX safety and silent-filter gaps. No consensus break requiring escalation. Flagged for user decision only if live regional Reports API behavior differs from the documented `/report/v1` path (PLAN already carries that risk; `CLOCKIFY_REPORTS_URL` exists for overrides).

## Verdict

**approve with follow-ups**

M6 delivers summary/detailed/weekly reports with correct official regional Reports API base URLs, dual-client filter resolution, CSV/PDF export, and solid unit coverage matching the M6 deliverable table. `--file` dual-use overwrite risk, incomplete params merge, transport duplication, and missing detailed-report pagination are real but non-blocking gaps at this SHA; they should be scheduled as explicit follow-ups rather than treated as a failed milestone.

## HEAD triage (2026-10-05)

- **Current HEAD:** `ac64914`
- **Original verdict:** unchanged
- **Legend:** ✅ completed at HEAD · ⚠️ partially fixed · ❌ still open · ➖ N/A or cannot verify

| # | Item | Status | Evidence |
|---|------|--------|----------|
| M1 | `--file` dual meaning can overwrite the params input file | ❌ | `internal/cli/report.go` L181–187 still reads `--file` as JSON params when `--export` is empty; L357–363 still does `os.WriteFile(filePath, body, 0644)` on CSV-looking responses with no guard against clobbering the input path |
| M2 | Params-file merge incomplete relative to spec intent | ❌ | `mergeReportFileParams` (`report.go` L409–425) unchanged: still only `billable`, `description`, `exportType`, `approvalState`, `dateRangeType`; structured filters (`projects`/`users`/`clients`/`tasks`/`tags`) and other date keys still silently dropped without error |
| M3 | `postReport` duplicates main client's retry/auth/error path | ❌ | `internal/api/reports.go` L47–118 still has a parallel retry loop, rate-limiter Take, header setup; non-retryable network failures still wrapped as `errors.Wrap(errors.E005RateLimit, "report request failed", err)` at L87, while `Do()` uses `E006Network` (`client.go` L117, L148) |
| M4 | Report row pagination not exposed | ❌ | `report.go` L312 still sets `DetailedFilter{PageSize: 500}` only; `Page` field exists in `pkg/clockify/reports.go` L39 but the CLI never sets it, has no report `--page`/`--page-size` flags, and never walks subsequent pages |
| m1 | `WeekStart` hardcoded to `MONDAY` | ❌ | `report.go` L243 still `WeekStart: clockify.ReportWeekStartMonday`; no `--week-start` flag; M11 config `WeekStarts` (`m11_resolve.go`) not wired into report requests |
| m2 | `timeZone` never set on `ReportRequest` | ❌ | `formatReportDate` (`report.go` L434–439) still emits offset-less timestamps; `ReportRequest.TimeZone` (`pkg/clockify/reports.go` L71) never populated; no `CLOCKIFY_TIMEZONE` handling found |
| m3 | `LooksLikeCSV` body heuristic is brittle | ❌ | `internal/api/reports.go` L172–189 unchanged: first-line-comma heuristic still treats non-JSON error/proxy text as CSV |
| m4 | `fromStruct` CSV writes `"<nil>"` for empty fields | ❌ | `internal/output/csv.go` L26 still `fmt.Sprintf("%v", row[key])`; L81 stores `item.Field(j).Interface()` with no nil→`""` mapping |
| m5 | `--status` vocabulary vs API enum not clarified in help | ❌ | `report.go` L111 help still lists `IN_PROGRESS, NEW, APPROVED, REJECTED` with no up-front note that NEW/IN_PROGRESS map to `UNAPPROVED` or that `REJECTED` is rejected (`mapReportStatus` L442–459 unchanged) |
| m6 | No live Reports-API integration test | ❌ | `integration/integration_test.go` L142–156 `TestIntegrationReportsSummary` only runs a `ListProjects` connectivity check; it never POSTs to `/reports/*` |
| m7 | Filter resolution still inherits unpaginated list calls | ❌ | Resolvers unchanged: `resolveProjectByName` (`project.go` L508, `ListProjects(Opts{})`), `resolveTaskByName` (`task.go` L629), `resolveUserRef` (`task.go` L662), `resolveClientByName` (`client.go` L426); none use `fetchPaginated*` helpers |
| R7 | Later milestones re-check (process item) | ➖ | Recommendation 7 is a process step; this triage is that re-check — substantive findings above remain open at HEAD |

### Summary

All 11 substantive findings and 6 of 7 recommendations remain open at HEAD `ac64914`. Zero completed, zero partially fixed. Commits `969ca42` (M8) and `7c70a1a` (M11) touched report-adjacent files but did not address any M6 report findings. The original "approve with follow-ups" verdict stands; the follow-ups have not been delivered.

### Completed at HEAD

None.

### Still open at HEAD

- **M1** — `--file` params/export dual-use overwrite risk (`internal/cli/report.go` L181–187 + L357–363)
- **M2** — incomplete params-file merge; structured filters silently ignored (`mergeReportFileParams` L409–425)
- **M3** — `postReport` transport/retry/error duplication; network failures still misclassified as `E005RateLimit` (`internal/api/reports.go` L47–118, L87)
- **M4** — detailed-report pagination not exposed; `PageSize: 500` only, no `Page`, no page walk (`report.go` L312)
- **m1** — `WeekStart` hardcoded to `MONDAY`; no `--week-start` flag (`report.go` L243)
- **m2** — `timeZone` never set; offset-less date formatting (`report.go` L434–439)
- **m3** — `LooksLikeCSV` brittle body heuristic (`internal/api/reports.go` L172–189)
- **m4** — `"<nil>"` in CSV output (`internal/output/csv.go` L26, L81)
- **m5** — `--status` help text still does not document API enum mapping/rejection (`report.go` L111)
- **m6** — no live Reports-API integration test; `TestIntegrationReportsSummary` is connectivity-only (`integration/integration_test.go` L142–156)
- **m7** — report filter resolvers still first-page-only list calls (task.go/project.go/client.go)
