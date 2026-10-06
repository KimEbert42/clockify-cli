# Commit Review — f0d6787

- **Commit:** f0d6787e109ef1946cad9632f24a3c2a3d59f2dc
- **Subject:** feat: M4 - Write Operations for project, client, tag CRUD
- **Date:** 2026-10-03
- **Experts engaged:** Backend Engineer, Full Stack Developer
- **Skill mode:** Auto-Assignment (kims-team)
- **Review date:** 2026-10-05
- **Related plan:** PLAN.md §5.7 Review Backlog

## Diff summary

Implements M4 write operations: create/update/delete for projects, clients, and tags.

- `internal/api/resources.go` — 9 write methods (Create/Update/Delete × project/client/tag) on `/workspaces/{wid}/{resource}[/{id}]`; updates correctly use **PUT** (Clockify API semantics; PLAN §5.3 later corrected the "PATCH" wording).
- `pkg/clockify/types.go` — 6 request types (`Create/UpdateProjectRequest`, `Create/UpdateClientRequest`, `Create/UpdateTagRequest`) + request types for time-entry (pre-existing area extended).
- `internal/cli/{project,client,tag}.go` — full subcommand trees: create (`--name`…, `--file`/`--stdin`, `--if-not-exists`), update (flag or JSON partial merge over GET-seeded body), delete (interactive confirm + `--confirm`); fuzzy `resolve*ByName` helpers.
- `internal/cli/{project,client,tag}_test.go`, `pkg/clockify/types_test.go` — help, httptest list/create/update/delete/if-not-exists paths (~15+ new tests).
- `.gitignore` — `clockify` → `/clockify` so `pkg/clockify/` is no longer ignored (correct fix).
- `cmd/clockify/main.go` — wires `SetBuildInfo` (error exit codes still blanket `os.Exit(1)` until M8).
- `PLAN.md` — M4 marked complete.

Milestone fit: deliverables and success criteria in PLAN §M4 are substantially represented in the surface (flags, confirm, if-not-exists, JSON input, name resolution). Several claims are not fully correct in the implementation — see findings.

## Findings

### Critical

None. No auth bypass, secret exposure, or destructive action without confirmation on the happy path.

### Major

1. **JSON input silently drops `estimate` on project create** — `internal/cli/project.go` (create command, JSON branch ~lines 170–190). `--file`/`--stdin` unmarshal into a local `CreateProjectRequest`, then copy only `Name`, `ClientID`, `Billable`, `Color`, `Note` into flag variables. A fresh request is rebuilt later; `req.Estimate` is set **only** when the `--estimate` flag is non-empty. Any `estimate` object in the JSON payload is discarded. This contradicts the commit message (“`--file`/`--stdin` JSON input for all write commands”) and loses user data on the documented input path.

2. **Name resolution lists one unpaged window (`page=0&pageSize=0`)** — `resolveProjectByName` (`internal/cli/project.go:482`), `resolveClientByName` (`internal/cli/client.go:404`), `resolveTagByName` (`internal/cli/tag.go:329`) all call `List*/Get*` with empty opts. `appendQueryParams` (`internal/api/resources.go`) emits `page=0&pageSize=0` (Itoa(0) is non-empty, so zero values are not skipped — unlike `buildProjectQueryParams`, which skips `<= 0`). Clockify paging is 1-based; this is at best a single default/invalid page. Consequences: `get/update/delete --name`, and `--if-not-exists`, can miss resources beyond that window → false “not found”, duplicate creates, failed updates on larger workspaces. Tests pass only because httptest handlers ignore paging.

3. **Update seed omits `Archived`** — project/client/tag update commands seed the PUT body from GET with only the modeled fields (project: Name/ClientID/Billable/Color/Note; client: Name/Email/Address/Note; tag: Name). `Archived` is never seeded. If Clockify PUT is a full-object replace (PLAN §5.3 acknowledges PUT + partial bodies was an open question), any update can clear archive state. Flag path also only supports `--archive` (true), never un-archive via flags; JSON path can set it.

4. **`--if-not-exists` is not strict idempotency and swallows errors** — e.g. `internal/cli/project.go` (~194): `resolveProjectByName`; on **any** error (not-found, network, auth) it proceeds to create. Combined with fuzzy matching, `--if-not-exists --name "Web"` can return a different project named "Website" as “already exists”, or, on a resolve failure, create a near-duplicate. Violates the success-criterion spirit (“returns existing resource without error if it already exists” — should be exact-name / controlled semantics).

### Minor

5. **`--file`/`--stdin` JSON wins over flags; flags ignored** — when JSON input is present, CLI flags are not applied (JSON branch overwrites). Later PLAN §5.5 (M8) records the opposite precedence for time-entry (“JSON flags override file/stdin”). Inconsistent contract across write commands at this SHA.

6. **`parseEstimateDuration` duplicates `pkg/duration` with different rules** — `internal/cli/project.go:448`. Supports only `h`/`d`/`m` with an 8h/day assumption; rejects `1h30m` and bare seconds that `pkg/duration.Parse` (M3) accepts for time-entry estimates. UX inconsistency between `time-entry` and `project` estimate inputs.

7. **`project --client` accepts ID only** — no `resolveClientByName` on create/update despite the commit advertising client name resolution (that resolution covers client get/update/delete, not project’s `--client` reference). Later M11 “names everywhere” addressed this class of gap.

8. **Dead code** — `projectHasConflict` (`internal/cli/project.go:511`) is defined and unused in this commit.

9. **Weak test coverage vs claimed behavior** — tests cover help, list filters, get-by-name, create/update happy paths, delete `--confirm`, and `--if-not-exists` exact match. **No tests for `--file`/`--stdin`** on any write command; no estimate-parse tests; delete-without-confirm test comments admit stdin cannot be mocked (`project_test.go` ~568–581). The pagination bug in finding 2 is therefore invisible to the suite.

10. **Confirmation prompt has no TTY guard** — `fmt.Scanln` on stdin; non-TTY safely aborts (empty response ≠ `y`), so CI is not destructive, but behavior is implicit rather than explicit (M11 later added strict TTY guards for interactive paths).

## Recommendations

1. Copy **all** JSON fields (including `Estimate`) when building the create request; prefer unmarshalling into the final request struct instead of a partial flag-variable bridge.
2. Resolve names via paginated fetch (reuse `fetchPaginated*` / `DoWithPagination` or pass `Page:1, PageSize:<large>` and loop) or API `search=` query param; align `appendQueryParams` zero-value handling with `buildProjectQueryParams`.
3. Seed `Archived` from the GET result on update (or document and test Clockify PUT field-preservation semantics against the live API).
4. Make `--if-not-exists` exact-name only (or explicit `--if-match fuzzy`); do not swallow non-NotFound errors from the existence check.
5. Define one precedence rule for flags vs `--file`/`--stdin` and apply it to all write commands; align with the M8 rule (“explicit flags win”) when that lands.
6. Reuse `pkg/duration.Parse` for project estimates; drop the local parser.
7. Add httptest cases for `--file`/`--stdin` (valid + invalid JSON + estimate-in-JSON), multi-page resolution, and archived preservation.
8. Wire `resolveClientByName` into project create/update `--client`.

**Later milestones (note, not scored):** M5 reused the CRUD template (including the same patterns); M8 fixed CLIError exit codes and documented flag>JSON precedence for time-entry; M11 added names-everywhere resolution and TTY-guarded interactives. None of these retroactively fix this SHA’s JSON-estimate drop or unpaged resolvers.

## Conflicts / unresolved

**None between experts.** Session synthesis (subagents intentionally not spawned — catch-up review, hang-avoidance guidance): Backend Engineer lens (API paths/PUT methods/request types sound; estimate-drop, unpaged resolution, Archived seed = correctness bugs) and Full Stack Developer lens (precedence inversion, missing file/stdin tests, ID-only `--client`, estimate parser drift) converge on the same verdict. No confidence-weighted split, escalation, or consensus split required.

Open question for user/live API: whether Clockify project/client/tag PUT without `archived` preserves archive state (finding 3) — verify against docs.clockify.me or a test workspace.

## Verdict

**needs rework**

The commit establishes the M4 CRUD surface (correct API verbs/paths, confirmation, if-not-exists flag, tests for happy paths), but core documented behaviors are wrong at this SHA: JSON input drops `estimate`, name resolution cannot see the full catalog (`page=0&pageSize=0`), update PUT bodies may lose `Archived`, and `--if-not-exists` semantics are fuzzy and error-swallowing. These are correctness issues in the milestone’s own success criteria, not polish items.

## HEAD triage (2026-10-05)

- **Current HEAD:** `ac64914`
- **Original verdict:** unchanged
- **Legend:** ✅ completed at HEAD · ⚠️ partially fixed · ❌ still open · ➖ N/A or cannot verify

| # | Item | Status | Evidence |
|---|------|--------|----------|
| C1 | `.gitignore` fix — `clockify` → `/clockify` so `pkg/clockify/` is tracked | ✅ | `.gitignore:2` = `/clockify`; `git ls-files pkg/clockify` shows tracked files; `git check-ignore pkg/clockify/types.go` → not ignored |
| C2 | `internal/api/resources.go` — 9 write methods (Create/Update/Delete × project/client/tag), updates use PUT | ✅ | `resources.go:134-229` — `CreateProject/UpdateProject/DeleteProject`, `CreateClient/UpdateClient/DeleteClient`, `CreateTag/UpdateTag/DeleteTag`; updates call `c.Do(ctx, "PUT", …)` |
| C3 | `cmd/clockify/main.go` — `SetBuildInfo` wired; error-exit note “blanket `os.Exit(1)` until M8” | ✅ | `main.go:19` `cli.SetBuildInfo(...)`; `main.go:22-26` now `errors.As` → `cliErr.ExitCode()` for `CLIError` (M8 landed the exit-code fix the diff-summary note anticipated) |
| 1 | Major — JSON input silently drops `estimate` on project create | ❌ | `internal/cli/project.go:200-204` JSON branch still copies only `Name/ClientID/Billable/Color/Note`; `req` rebuilt at `:220-226` without `Estimate`; `req.Estimate` set only when `--estimate` flag non-empty (`:228-238`). Rec 1 not implemented |
| 2 | Major — name resolution lists one unpaged window (`page=0&pageSize=0`) | ❌ | `resolveProjectByName` (`project.go:507-508`), `resolveClientByName` (`client.go:425-426`), `resolveTagByName` (`tag.go:349-350`) still call `List*` with empty opts; `appendQueryParams` (`resources.go:358-373`) skips only `v == ""` — `strconv.Itoa(0)="0"` is emitted → `page=0&pageSize=0`. `fetchPaginated*` helpers exist (`pagination.go`) but resolvers do not use them; completion cache (`resource_cache.go:188-204`) is not wired into resolve. Rec 2 not implemented |
| 3 | Major — update seed omits `Archived` | ❌ | Project seed `project.go:310-316` (Name/ClientID/Billable/Color/Note), client seed `client.go:271-276`, tag seed `tag.go:219-221` — none seed `Archived` from GET. Flag path still only `--archive` (true), no un-archive flag (`project.go:393`, `client.go:346`, `tag.go:270`). JSON path can set `Archived` (was already true at original SHA). Rec 3 not implemented |
| 4 | Major — `--if-not-exists` not strict idempotency; swallows errors | ❌ | `project.go:212-218`, `client.go:189-196`, `tag.go:146-153`: `if err == nil` only — **any** resolve error (network/auth/NotFound) proceeds to create. Fuzzy single-match fallback still in all three resolvers (`project.go:519-528`, `client.go:437-445`, `tag.go:361-369`). Rec 4 not implemented |
| 5 | Minor — `--file`/`--stdin` JSON wins over flags; flags ignored | ❌ | Project create JSON branch overwrites flag vars (`project.go:200-204`) and skips flag application; update uses mutually exclusive JSON/flag branches (`project.go:356`, same pattern `client.go:313`, `tag.go:249`). M8 documented “explicit flags override JSON fields” for time-entry (`time_entry.go:305` comment) — CRUD surface never aligned. Rec 5 not implemented |
| 6 | Minor — `parseEstimateDuration` duplicates `pkg/duration` with different rules | ❌ | `project.go:473-505` still the local `h`/`d`/`m` + 8h/day parser; task create uses `duration.Parse` (`task.go:356`). Inconsistency between `project --estimate` and `task`/`time-entry` estimates remains. Rec 6 not implemented |
| 7 | Minor — `project --client` accepts ID only | ❌ | `project.go:249` (create) and `:389` (update) flag help still “client ID”; no `resolveClientByName` / `resolveClientRef` call in `project.go` (grep confirms). `resolveClientRef` exists (`m7_helpers.go:170-179`) but is used by M7 resources, not project `--client`. Rec 8 not implemented |
| 8 | Minor — dead code `projectHasConflict` | ❌ | `project.go:540-547` still defined; no call sites anywhere in the codebase (only self-definition + review-doc mentions) |
| 9 | Minor — weak test coverage vs claimed behavior | ❌ | `project_test.go`/`client_test.go`/`tag_test.go` (628/484/391 lines): no `--file`/`--stdin` cases, no estimate-parse tests, no multi-page resolution tests, no archived-preservation tests; delete-without-confirm comment still present (`project_test.go:571-574`). (`task_test.go` has `--file` cases — M5, not this surface.) Rec 7 not implemented; suite still passes (`go test ./internal/cli/ -run 'TestProject\|TestClient\|TestTag'` → ok) precisely because httptest fixtures ignore paging |
| 10 | Minor — confirmation prompt has no TTY guard | ❌ | Delete confirm still bare `fmt.Scanln` on stdin: `project.go:450`, `client.go:402`, `tag.go:326`. M11 TTY guards (`picker.go:43-47`, `output.IsTerminalStdin/Stdout`) are wired to interactive pickers only — CRUD delete prompts unchanged; behavior still implicit abort on non-TTY |

### Summary
- 10 original findings re-verified at HEAD `ac64914`: **0 completed, 0 partially fixed, 10 still open**.
- 3 diff-summary context items (gitignore fix, API write methods/PUT verbs, entrypoint `SetBuildInfo` + exit codes): **all verified present** at HEAD; the “blanket `os.Exit(1)` until M8” note is superseded — `main.go` now maps `CLIError` to its exit code.
- Recommendations 1–8 map 1:1 onto findings 1–7/9 and finding 7 (rec 8) — none implemented on the project/client/tag CRUD surface; later milestones (M5 template reuse, M8 time-entry precedence + CLIError, M11 names/TTY for **other** commands) did not retroactively fix these paths.
- Historical verdict stands: **needs rework**.

### Completed at HEAD
- C1 `.gitignore` `/clockify` fix (pkg/clockify tracked).
- C2 All 9 API write methods present with correct verbs (POST create, PUT update, DELETE).
- C3 Entrypoint `SetBuildInfo` + CLIError-aware exit codes (M8 improvement beyond the original note).

### Still open at HEAD
- Finding 1 — project create JSON input still drops `estimate`.
- Finding 2 — `resolveProjectByName`/`resolveClientByName`/`resolveTagByName` still unpaged (`page=0&pageSize=0` via `appendQueryParams`).
- Finding 3 — update PUT seed still omits `Archived`; flag path still archive-only.
- Finding 4 — `--if-not-exists` still swallows all resolve errors + fuzzy matching.
- Finding 5 — CRUD `--file`/`--stdin` still JSON-wins/flags-ignored; inconsistent with M8 time-entry rule.
- Finding 6 — project `--estimate` still uses local `parseEstimateDuration`, not `pkg/duration.Parse`.
- Finding 7 — project create/update `--client` still ID-only.
- Finding 8 — `projectHasConflict` still dead code.
- Finding 9 — no `--file`/`--stdin`, estimate, multi-page, or archived-preservation tests on the M4 CRUD surface.
- Finding 10 — CRUD delete confirm still has no explicit TTY guard.
