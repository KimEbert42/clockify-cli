# P0 Review-Debt Sprint — Daily-Driver Trust

**Source:** PLAN.md §5.8 (Review Debt Backlog, team triage 2026-10-05) · HEAD `ac64914` (`v1.0.0`)
**Sprint window:** ~6.5 productive days (2-week calendar, 15% overhead + 20% rework buffer)
**Mode:** Phase 0 integrity — **70 debt : 30 feature** until P0 board clear
**Goal:** *Make the daily driver trustworthy* — billing correctness, headline UX, config safety, supply-chain gate, honest CI contract
**Status:** ✅ **IMPLEMENTED** 2026-10-06 · Evidence below · Next: commit + kims-team review (§5.7)
**Not in this sprint:** plugin system · new milestones · product decisions (T14/T18/T28/… parked in §5.8.2) · pure docs batch · live-API spikes until creds exist

**DoD for the sprint:**
- `go test ./...` green; `make check` green with **blocking** lint
- Regression tests for T01–T03, T06, T08, T34, T04/T05/T40
- No silent behavior change on bare `time-entry list` default (T18 stays parked)
- Supply-chain P0s closed **or** README install claims amended (G3)
- Every closed item: §5.8 status updated in the same commit (✅/⚠️ + evidence note)
- Per §5.7: each commit reviewed via `kims-team` → `docs/reviews/`

**Evidence (2026-10-06):**
- `go test -race ./...` green (all packages)
- Core coverage gate: config 88.9% · output 91.9% · errors 89.9% · update 83.3% · resolve 100% · api 88.1% · duration 87.4% · clockify 86.1%
- PLAN.md §5.8.2 P0 board + §5.8.4 + §5.8.6 updated with status
- Pre-commit review draft: `docs/reviews/p0-sprint-daily-driver-trust.md`
- Partial: T07 (minimal pagination honesty) · T15 (if-not-exists only; clear-field/Archived seed → P1)

---

## 1. Baseline (HEAD `ac64914`, working tree has review scaffolding only)

| Check | Evidence |
|-------|----------|
| Milestones M0–M11 + M10 | Complete · tags through `v1.0.0` |
| `go test ./...` | Green |
| Core coverage (M10) | api 90% · config 90% · errors 98% · output 95% · resolve 100% · update 84% · duration 88% · clockify 86% |
| Binary / startup | ~11.5MB · ~0–50ms warm — budgets hold; no size work this sprint |
| Commit reviews | 17/17 reviewed; 7 needs-rework · 10 approve-with-follow-ups |
| Open P0 themes | ~12 · unique open items overall ~45–55 after dedup |
| Git remote | ❌ None — blocks live brew/npm/release E2E (operational, not code) |

**Dominant root causes (do not treat as 45 one-off patches):**
1. Pagination/resolver honesty (replicated double-fetch + fabricated `PageInfo`)
2. Flag/alias precedence unwired
3. Docs written to spec, not code
4. CI gates that don't gate
5. False-green status DoD (milestone ✅ without tree evidence)

---

## 2. Item Register (all P0 board + sprint prereqs)

Status legend: ❌ open · ⚠️ partial · ✅ done (this sprint) · 🅿️ product decision · 🧪 needs live API

| ID | Theme | Status | Effort | Primary files (HEAD) |
|----|-------|--------|--------|----------------------|
| **T01** | Relative **end** ranges | ❌ | S | `pkg/duration/relative.go`, `internal/cli/time_entry.go:89-94`, `m7_helpers.go`, `report.go` |
| **T02** | Single-struct table/CSV/YAML empty | ❌ | M | `internal/output/csv.go` `fromStruct` (slice-only); table path same gap |
| **T03** | Alias flag inheritance + Execute tests | ❌ | S | `internal/cli/aliases.go` (copies RunE only, not flags) |
| **T04** | sha256 verify vs `checksums.txt` | ❌ | M | `internal/update/update.go` `Install` (download→extract→replace, no verify) |
| **T05** | npm `execSync` → `execFile`/argv | ❌ | S | `npm/install.js:11,118,125` |
| **T06** | Map **403→E003Permission** (401 stays E001) | ❌ | S | `internal/errors/errors.go:132-135`; tests `resources_test.go`, `errors_test.go` assert 403→E001 today |
| **T07** | Pagination source of truth (SPIKE) | ❌ | SPIKE 0.5–1d → L | `internal/cli/pagination.go` (double-fetch projects; fabricated PageInfo); `internal/api/client.go` `DoWithPagination`; resolvers; `internal/resolve` orphaned |
| **T08** | `url.Values.Encode()` on query builders | ❌ | S | `internal/api/resources.go` `appendQueryParams` / `buildTimeEntryQueryParams`; `internal/cli/pagination.go` `buildProjectQueryParams`; M7 params maps |
| **T09** | HTTP client resilience | ❌ | M | `internal/api/client.go` `defer resp.Body.Close()` **inside** retry loops (`Do`, `DoWithPagination`); no body rewind; rate limiter under per-request lock; `postReport` path |
| **T10** | `split` atomicity | ❌ | M | `internal/cli/m11_power.go:290-337` — update original **then** create segments; mid-loop failure leaves inconsistent state, no rollback |
| **T11** | `clear-cache` per-profile only | ❌ | S | `internal/config/paths.go:38-44` maps `""`/`default` → **cache root**; `cache.go:13-21` `RemoveAll` root; `TestClearCache` enshrines wipe |
| **T12** | `configure` merge-on-save | ❌ | S–M | `internal/cli/configure.go:101-113` builds `config.Config{Region,WorkspaceID}` and **replaces** profile — drops other keys |
| **T13** | `Load` fail-loud on corrupt YAML | ❌ | S | `internal/config/config.go:42-62` ignores `yaml.Unmarshal` error; `config_extra_test.go:267` asserts silent fallback |
| **T15/T33** | Clear-field semantics; `--if-not-exists` only swallows NotFound | ❌ | M | time-entry/project/task update merge paths; CRUD create `--if-not-exists` |
| **T16** | Bulk partial failure → non-zero exit; require end/duration | ❌ | S | `internal/cli/time_entry.go:772+` collects `errors` slice; need to confirm final exit code + per-entry validation |
| **T29** | Hidden API-key input in `configure` | ❌ | S | `configure.go:31-36` plain `ReadString` (echoed) |
| **T30** | Comma-ok on `ws["id"]` type assert | ❌ | XS | `configure.go:88` `workspaces[choice-1]["id"].(string)` — panics if id non-string |
| **T34** | CI fail-closed: lint blocking, coverage threshold, `-race` | ❌ | S | `.github/workflows/ci.yml:25` `continue-on-error: true`; coverage prints but does not fail; `Makefile:36-37` `test:` no `-race` |
| **T37** | `errors.As` walk `%w`; `Format()` in `main.go`; debug/verbose → stderr | ❌ | S | `main.go:21` prints `Error: %v` not `Format()`; `errors.As` in `internal/errors/format.go:34-50` does **not** unwrap `%w` chains; debug/verbose `fmt.Printf` → stdout |
| **T40** | Homebrew `brews:` gated on secret presence | ❌ | S | `.goreleaser.yml:67-79` unconditional; `.github/workflows/release.yml` always `goreleaser release` on `v*` |
| **T02c** | Table/YAML single-struct parity | ❌ | S | Same root as T02 — handle struct (not only `[]struct`) in table + YAML if CSV alone is fixed |

**Sprint dependency notes:**
- T06 → must land **before** or **with** docs that claim exit 4 for 403 (one commit for code + README + `docs/errors/E003-permission.md` + PLAN tables).
- T07 is the only SPIKE: design first (0.5–1d); if design risk ≤1d implement fully, else **minimal honesty fix** (drop double-fetch, omit `page/pageSize` when ≤0, stop fabricating totals) and ticket the rest to P1.
- T04/T05/T40 are the **G3 supply-chain gate** — required before any public npm/brew/`update` advertising.
- Product decisions (T18 list default, T14 `--hydrated`, T28 status cache, T32 archived PUT, T22 invoice aliases, T48 `--api-key`) are **not coded** in this sprint.

---

## 3. Sprint Schedule

Capacity: ~6.5 productive days. Order is milestone-like: **trust the daily path first**, then safety gates, then structural debt.

| Day | Items | Detail | Exit check |
|-----|-------|--------|------------|
| **D1** | **T01 + T02/T02c** | Relative **end** ranges + single-struct output | `--end this-week` = Monday **EOD** (local TZ); get/create/start in table/CSV/YAML non-empty |
| **D1.5** | **T03** | Alias inheritance + Execute-level regression | `clockify in --project X` parses like `time-entry start --project X`; tests **Execute** alias cmds, not just `findCommand` |
| **D2** | **T08** | Query encoding | All builders use `url.Values`; httptest asserts encoded special chars (`search=a b`, tags with `,`) |
| **D2.5** | **T11–T13 + T29/T30** | Config safety + secrets | Per-profile clear-cache; merge-on-save; Load fail-loud; hidden key; comma-ok asserts |
| **D3** | **T06 + T37** | Error contract | 403→E003; one exit-code table everywhere; `main` uses `Format()`; errors.As unwraps `%w`; debug→stderr |
| **D3.5** | **T10** | Split atomicity | Create-then-update **or** rollback on mid-failure; clear error; regression test with forced mid-fail |
| **D4** | **T34 + T09** | CI gates + HTTP client | Lint **blocks**; coverage threshold fails; `-race` in `make test`; defer-in-loop closed; body rewind on retry |
| **D4.5** | **T04 + T05 + T40** | Supply chain | Verify `checksums.txt` before replace; npm `execFile`; brews publish only if `HOMEBREW_TAP_GITHUB_TOKEN` set (or README amended) |
| **D5** | **T15/T33 + T16** | Write-path semantics + bulk | Update clear-field rules documented + tested; `--if-not-exists` swallows **only** NotFound; bulk non-zero on partial fail |
| **D5.5–6.5** | **T07** | Pagination SPIKE → implement | See §5. If spike balloons, ship minimal honesty + leave architecture P1 |

**Stretch (only if capacity remains):** T27 wire `--workspace` precedence · T44 shared confirm helper · T45–T46 estimate/`--client` consistency · batch T53/T54 docs honesty.

---

## 4. Per-Item Implementation Notes

### T01 — Relative end ranges (P0, billing correctness)

**Problem:** `ParseRelativeDate` returns **start-of-day** for period tokens (`this-week` = Monday 00:00). Callers that treat the result as an inclusive **end** collapse ranges (e.g. `--end this-week` becomes Monday 00:00 instead of Sunday 23:59:59). `time_entry.go:89-94` already special-cases end to 23:59:59, but **only** that path; reports/M7/other list paths need a single rule.

**Design:**
1. Add `ParseRelativeDateEndIn(s, weekStart, loc, now)` (or `EndOfRelativeDate`) — same tokens, end-of-day / end-of-period semantics:
   - `today` / `yesterday` → 23:59:59.999… local
   - `this-week` / `last-week` → last day of week, 23:59:59 (weekStart honored)
   - `this-month` / `last-month` → last day of month, 23:59:59
2. Keep `ParseRelativeDate` as **range start** (document clearly).
3. CLI: end-side relative tokens use `ParseRelativeDateEnd*`; start-side unchanged.
4. Call sites to audit: `time_entry.go` list, `report.go` presets + `parseReportDate`, `m7_helpers.go` date flags, schedule/time-off/expense/invoice date ends.

**Tests:** table for each token × weekStart monday/sunday × mid-period `now`; assert inclusive end > corresponding start; regression: `report this-week` / `time-entry list --end this-week`.

**Exit:** `--end this-week` spans through end of week; no bare-midnight collapse on end flags.

---

### T02 / T02c — Single-struct output (P0, headline UX)

**Problem:** `internal/output/csv.go` `fromStruct` only accepts **slices** (`val.Kind() != reflect.Slice → nil`). Single structs (get/create/update/start/stop) produce **empty** table/CSV/YAML rows. JSON works (raw encode).

**Design:**
1. Extend `fromStruct` (or add `structToMap`) to handle:
   - single `struct` → one map
   - `*struct` → one map (deref; nil → empty)
   - `[]struct` / `[]*struct` → maps (existing)
2. Mirror in table + YAML formatters if they special-case slices.
3. Header keys: keep yaml-tag stripping + `omitempty` logic; fall back to lowercased field names.
4. Nested structs: flatten one level (existing list pattern) or emit dotted keys — pick one, document in PLAN §5.8 notes.

**Tests:** httptest-style unit tests per formatter for `Project`, `TimeEntry` (pointer + value); headers match; non-empty single row; existing multi-row tests still pass.

**Exit:** `clockify project get … --output table|csv|yaml` prints the entity; start/stop/create same.

---

### T03 — Alias flag inheritance (P0, README example)

**Problem:** `registerAliases` copies only `RunE`/`Run`/`Short` — **not** `Flags()` / `PersistentFlags()`. README `clockify in --project "Website Redesign" …` fails at parse (unknown flag on alias command). Tests only assert `findCommand` resolution.

**Design:**
1. After creating the alias command, `cmd.Flags().AddFlagSet(target.Flags())` and same for persistent flags (or set alias's parent to share flag inheritance via cobra alias pattern).
2. Also copy `ValidArgsFunction` / flag completion registration where practical.
3. Prefer thin alias commands that **delegate** to target command RunE with the same args after flag registration — avoid diverging flag sets.
4. Add Execute-level tests: `root.SetArgs([]string{"in", "--project", "X", "--output", "json"})` (httptest API) — expect parse success and same behavior as `time-entry start`.

**Exit:** README daily-driver alias example parses; `--help` on `in` shows start flags; ge/ce/out/lw/le/lp/cfg consistent.

---

### T04 + T05 + T40 — Supply chain & packaging gate (P0, G3)

**T04 — Checksum verify (update + npm):**
1. GoReleaser already publishes `checksums.txt` (`.goreleaser.yml:46-48`).
2. `internal/update.Install` after download: fetch `checksums.txt` asset from same release; parse lines `sha256  filename`; sha256 the downloaded archive; **fail closed** on mismatch (never replace binary).
3. Unit tests: httptest release with matching/mismatched/missing checksums — mismatch aborts; missing checksums → fail closed with actionable error (or explicit `--skip-verify` later — **not** silent skip).
4. `npm/install.js`: same verify after download, before extract.

**T05 — npm extract without shell:**
1. Replace `execSync(\`unzip…\`)` / `execSync(\`tar…\`)` with `execFile`/`spawn` **argv arrays** (no interpolation of API-derived paths into a shell string).
2. Prefer pure-JS extract if dependency-light; otherwise `execFile("tar", ["-xzf", archivePath, "-C", dir])`.
3. Never build command strings from asset names.

**T40 — Brews gated on secret:**
1. `.goreleaser.yml`: condition `brews` publish on env token presence (GoReleaser `skip_publish` / template, or split release job).
2. `.github/workflows/release.yml`: `if: secrets.HOMEBREW_TAP_GITHUB_TOKEN != ''` around brew tap step — **or** README install lines amended until tap exists (G3).
3. Until public remote + first Release: README must not claim `brew install` / live `clockify update` works — either gate is real or claims are true.

**Exit:** unit tests green for verify paths; npm script has no `execSync` of archive paths; release workflow won't fail solely on missing tap token; G3 satisfied.

---

### T06 + T37 — Error contract & CI exit codes (P0, false CI contract)

**T06:**
1. `FromHTTPStatus`: `403 → E003Permission`; keep `401 → E001Auth`.
2. Update `handleErrorResponse` suggestion for E003 (workspace role, not re-auth).
3. Flip assertions that expect 403→E001 (`errors_test.go:73+`, `resources_test.go:93-94,414`).
4. **One exit-code table** in README + `docs/errors/E003-permission.md` + PLAN (drop any “0–12” leftovers — code is 2–11 + default 1).
5. Spec §6.6 vs Appendix C: annotate historical contradiction; do not invent codes.

**T37:**
1. `internal/errors.As`: walk `%w` chain (std `errors.As` wrapper or custom unwrap loop).
2. `cmd/clockify/main.go`: print `clierrors.Format(err)` to stderr (E-code + Detail + Learn more), not bare `Error: %v`.
3. Debug/verbose HTTP logs → **stderr** (pipe-friendliness: JSON on stdout stays clean).
4. Tests: wrapped CLIError unwraps; Format on exit path; stderr routing.

**Exit:** `clockify status` with bad key → E001/exit 2; insufficient role → E003/exit 4; docs match code.

---

### T08 — Query encoding (P0, correctness)

**Problem:** `appendQueryParams` / `buildTimeEntryQueryParams` / `buildProjectQueryParams` concat raw strings (`search=` + value) — spaces, `&`, unicode break or split params.

**Design:**
1. Replace with `url.Values` + `.Encode()` in **all** query builders (resources.go, pagination.go, M7 list opts, reports if any query params).
2. Omit `page`/`pageSize` when `<= 0` (already partly done — keep invariant).
3. Trim tags / filter strings; reject or encode consistently.
4. httptest: search with spaces/specials; assert `r.URL.Query().Get("search")`.

**Exit:** encoded queries; special-char tests green; no raw `fmt.Sprintf` path builders left for query strings.

---

### T09 — HTTP client resilience (P0)

**Problems (M0 C6, M2 M4, M10 #4):**
1. `defer resp.Body.Close()` **inside** retry loops → defers pile up until function return (body leak under retries).
2. Retries reuse `req` without rewinding body (GET ok; POST retry broken).
3. No `Retry-After` honor on 429.
4. Rate limiter `Take()` called under client-level lock per request — serializes waiters; should be outside the per-request critical section.
5. `DoWithPagination` / `postReport` duplicate Do logic — share transport helpers.

**Design:**
1. Extract shared `doOnce` / `doWithRetries` helper used by `Do`, `DoWithPagination`, reports.
2. Close body **before** sleep/continue (explicit close, not defer-in-loop).
3. Buffer request body once; reset reader on retry (`bytes.Reader` + `GetBody` or re-create).
4. Honor `Retry-After` when present.
5. Rate limiter: take without holding mutex across sleep (already mostly token-bucket — verify).
6. Tests: retry closes N-1 bodies (httptest); POST retry body intact; Retry-After path.

**Exit:** no defer-in-loop; shared client core; body-safe retries.

---

### T10 — Split atomicity (P0, write path trust)

**Problem:** `runTimeEntrySplit` updates the original entry **first**, then creates segments. If create #2 fails, user has truncated original + orphan segments; no rollback; error may not list created IDs.

**Design (prefer create-then-update):**
1. Validate points (already strict).
2. Create all **new** segments first (from points onward, remainder).
3. Only then update original to end at first split point.
4. On mid-create failure: delete **only** segments created in this call (best-effort rollback); if rollback fails, return error listing created IDs + manual cleanup hint.
5. Alternative: keep update-first but snapshot original start/end and **restore** on failure — document choice in review.

**Tests:** httptest fails create after N successes → assert original restored or error lists IDs; multi-point split happy path.

**Exit:** no silent inconsistency; clear partial-failure error; regression test.

---

### T11–T13 + T29/T30 — Config safety bundle (P0)

| Sub | Fix |
|-----|-----|
| **T11** | Default `clear-cache` must **not** `RemoveAll` cache **root**. Per-profile only: clear `ProfileCacheDir(profile)` contents for named profiles; for default, clear default-profile cache dir **without** deleting sibling profile dirs — restructure cache layout if needed (`CacheDir()/profiles/<name>/` vs `CacheDir()/default/`). Update `TestClearCache` that enshrines root wipe. Message must say which profile was cleared. |
| **T12** | `configure` must **merge**: load existing profile, overlay only keys the wizard collected (region, workspace); never wipe other keys (output, weekStarts, interactive, timezone, auto_update…). |
| **T13** | `Load` / `LoadRaw`: corrupt YAML → return **error** (CLIError with path + hint), not silent defaults. Fix test that asserts corrupt → defaults. Missing file still OK (defaults). |
| **T29** | `configure` API key prompt: hide input (term.ReadPassword or platform equivalent); never echo key. |
| **T30** | Workspace auto-detect: `id, _ := ws["id"].(string)`; error if empty/non-string — no panic. |

**Exit:** corrupt config fails loudly; clear-cache isolated; configure preserves unrelated keys; key not echoed; no type-assert panics.

---

### T15/T33 + T16 — Write semantics & bulk (P0)

**T15/T33:**
1. Document and test **clear-field** semantics for updates (explicit empty vs omit vs null) — match Clockify API (partial PUT/POST).
2. `--if-not-exists`: only swallow **NotFound** (E002); auth/validation/conflict still fail.
3. Seed `Archived` / estimate fields on CRUD create/update where API expects them — fix lossy updates (live check T32 parked 🅿️).

**T16 — Bulk:**
1. Require per-entry `end` **or** `duration` (timer-start bulk is a separate product call — default reject incomplete).
2. Partial failure → **non-zero exit** (e.g. E010 or dedicated code) + summary `created=N failed=M`.
3. Tests: mixed success/fail file → exit ≠ 0; missing duration rejected.

---

### T34 — CI fail-closed (P0, G6)

| Change | Detail |
|--------|--------|
| Lint | Remove `continue-on-error: true` from `.github/workflows/ci.yml` lint step |
| Coverage | Fail if **core package** coverage < 80% (script reusing `make cover` awk; core list = M10 definition) |
| `-race` | `Makefile` `test:` → `go test -race ./...` (CI and local) |
| Size budget | Keep existing <20MB gate |

**Exit:** CI red on lint/coverage/race failure; Makefile matches CI contract.

---

### T07 — Pagination honesty (P0 structural, SPIKE)

**Problems:**
1. `fetchPaginatedProjects` with `!noPagination` calls `DoWithPagination` **then** `ListProjects` again — double fetch; second result wins; headers discarded.
2. Client/tag/task paths fabricate `PageInfo{Page, PageSize}` without totals from headers.
3. `noPagination` loops build synthetic PageInfo (`TotalPages: currentPage`) — wrong if API doesn't return all.
4. Name resolvers unpaged (project/task/tag/user) — can miss resources beyond page 1.
5. `internal/resolve` package orphaned (100% cov, no production callers) — wire or delete.

**Spike design (0.5–1d) — decide before coding:**
1. **Single source of truth:** `DoWithPagination` returns items + header-derived `PageInfo` (X-Pagination-*); list commands never re-fetch.
2. Omit `page`/`pageSize` from query when ≤0 (API default).
3. Unfiltered list cache only stores **complete** result sets or honest page metadata — never cache partial as full.
4. Resolvers: either paginate all pages or use search/filter API if Clockify exposes name search; if not, document first-N limitation + UUID fast path (already).
5. `internal/resolve`: **delete** if unused, or become the shared fuzzy helper used by all `resolve*Ref` functions.

**If spike → L (risk):** minimal ship = remove double-fetch, real header PageInfo, omit zero pages, delete orphan package; leave full resolver pagination as P1 with explicit docs.

**Exit:** one list request per page; honest or omitted totals; no fabricated counts; resolver policy decided + tested.

---

## 5. Process Rules (G1–G7) enforced this sprint

| Rule | Sprint application |
|------|-------------------|
| **G1** | Milestone/docs ✅ only with tree evidence at the feat SHA; PLAN status cites SHA |
| **G2** | Review recs not fixed in-landing → §5.8 row; open P0 blocks COMPLETE claims |
| **G3** | No public npm/brew/`update` advertising until T04/T05/T40 close **or** README amended in D4.5 commit |
| **G4** | Status-doc liveness: PLAN/spec/README vs HEAD — fix code **or** claim same commit |
| **G5** | No-pattern-replication checklist: queries→`url.Values`; pagination→real totals or honest omit; aliases→copy flags; formatters→struct+slices; errors→`internal/errors`; secrets→no echo; resolvers→shared helper |
| **G6** | CI must **fail**: lint blocking, coverage threshold non-zero, `-race` |
| **G7** | Boy Scout 10–15% on touched files; biweekly re-tier of §5.8 |

**Commit review:** every sprint commit → `kims-team` → `docs/reviews/<short>-<slug>.md` (§5.7). Heavier reviews for T07/T09/T10/T04; lighter for docs/chore if any.

---

## 6. Explicitly Out of Sprint

| Item | Why deferred |
|------|----------------|
| T14 `--hydrated` | Product: implement vs remove (spike needs live API) |
| T18 bare list default “today” | Product decision — no silent change |
| T28 status offline/cache | Product |
| T32 PUT without `archived` | 🧪 live workspace check |
| T20–T26 M6/M7 polish | P1 after daily-driver trust |
| T27 `--workspace` wiring | Stretch only |
| T38–T39 Windows npm/update | Unless npm/brew stays advertised post-G3 |
| T53/T54 PLAN/spec honesty batch | Docs P1; do minimal corrections only where T06 touches exit codes |
| Plugin system | Post-v1.0.0 / later horizon |
| New features | Phase 0 is 70:30 debt:feature |

---

## 7. Regression Test Matrix (must land)

| ID | Test shape | Package |
|----|------------|---------|
| T01 | Relative end tokens × weekStart | `pkg/duration` |
| T01 | CLI list/report end flags | `internal/cli` httptest |
| T02 | Single struct → table/CSV/YAML non-empty | `internal/output` |
| T03 | `in --project` Execute path | `internal/cli` |
| T04 | Checksum match/mismatch/missing | `internal/update` |
| T05 | npm extract no shell interpolation | manual/script review + unit if JS test harness |
| T06 | 403→E003, 401→E001; exit codes | `internal/errors`, `internal/api` |
| T07 | Single list call; header PageInfo; no zero-page params | `internal/cli`, `internal/api` |
| T08 | Encoded special-char queries | `internal/api` httptest |
| T09 | Body close + retry rewind | `internal/api` httptest |
| T10 | Split rollback / created-ID error | `internal/cli` |
| T11 | clear-cache isolation | `internal/config` |
| T13 | Corrupt YAML error | `internal/config` |
| T16 | Bulk partial → non-zero | `internal/cli` |
| T34 | Coverage script fails <80% core; `-race` | CI + Makefile |

---

## 8. Success Criteria (sprint close)

- [ ] T01 relative **end** ranges correct (inclusive period ends)
- [ ] T02 single-struct output non-empty in table/CSV/YAML
- [ ] T03 alias commands accept target flags; Execute tests green
- [ ] T08 all query builders use `url.Values.Encode()`
- [ ] T11–T13 + T29/T30 config safety green
- [ ] T06 403→E003; single exit-code table; T37 Format + stderr + As unwrap
- [ ] T10 split atomic or documented rollback
- [ ] T34 lint/coverage/race **fail closed** in CI + Makefile
- [ ] T04/T05/T40 supply chain gate closed **or** README install claims amended (G3)
- [ ] T15/T33 write semantics + T16 bulk exit codes
- [ ] T07 spike outcome: implement **or** minimal honesty + P1 ticket
- [ ] `go test ./...` + `-race` green; binary/startup budgets still hold
- [ ] PLAN.md §5.8 statuses updated in same commits (✅/⚠️ + evidence)
- [ ] §5.7 review artifacts for every feature commit

---

## 9. Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| T07 spike exceeds 1d | Hard timebox; ship minimal honesty; P1 architecture |
| Cache layout change (T11) breaks TTL completions | Migration note; clear-cache only removes profile subtrees; completion cache paths reviewed |
| E003 mapping breaks scripts that treated 403 as auth | Document in README + error page; exit 4 already documented as permission |
| Checksum missing on old releases | Fail closed with message to use package manager; no silent accept |
| `-race` surfaces latent races | Budget time D4; fix or quarantine known flaky tests with reason |
| No git remote / no live API | All work unit-tested via httptest; integration suite remains env-gated |

---

## 10. Follow-ups after this sprint (P1 — not in DoD)

- Full pagination architecture if only minimal landed
- Reports `--file` dual-use split (T23–T26)
- M7 tz `Z` + list page flags (T20–T22)
- Interactive stdin/confirm shared helpers (T19/T44)
- Windows npm/update paths (if distribution stays public)
- PLAN/spec historical honesty batch (T53/T54) beyond exit-code corrections
- Product decisions: T14, T18, T28, T32, invoice aliases
- Plugin system (horizon Later)
- Public remote + first GitHub Release (operational) — **after** G3

---

*Plan authored from PLAN.md §5.8 + HEAD code inspection (`ac64914` + working-tree review scaffolding). Update this document only for scope changes; status lives in PLAN.md §5.8.6.*
