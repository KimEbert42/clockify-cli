# M10: Quality, Testing & Documentation — Sprint Plan

**Milestone:** M10
**Sprint Duration:** 3 × 1-day sprints (PLAN.md estimate 1–2 days dedicated + ongoing)
**Developer:** 1 (solo)
**Start Date:** 2026-10-05
**End Date:** 2026-10-07 (target)
**Status:** ✅ **COMPLETE**
**Completed:** 2026-10-05
**Prerequisite:** M11 complete — commit `7c70a1a` (working tree clean); tags through `v0.11.0-m11`
**Spec reference:** `clockify-cli-spec.md` §5 (Error Handling), §7 (Distribution), quality budgets (binary <20MB, startup <50ms)
**PLAN.md reference:** PLAN.md §M10, §3.2 sequencing (M10 after M11), §6.3 (tag `v1.0.0`)

**Goal:** Ship the first stable release — production-quality code (coverage, errors, security), complete user documentation (README + man pages), integration-test harness, performance proof — then tag **`v1.0.0`**.

**Implementation summary (2026-10-05):** Shipped core coverage tests (resolve/output/api/config/errors/update), `ERROR [E00X]` format + `docs/errors/E00X-shortname.md`, API-key redaction in debug, README, man pages (`make man`), env-gated `integration/` suite, CI coverage/size/lint, Makefile `cover`/`man`/`perf`. Core packages ≥80%. Binary ~11.5MB / startup ~0–50ms. Tag `v1.0.0`.

**Baseline measured 2026-10-05 (working tree @ `7c70a1a`):**

| Metric | Value | Budget / Target | Status |
|--------|-------|-----------------|--------|
| `go test ./...` | Green (10 packages, 0 failures) | Must stay green | ✅ |
| Total statement coverage | **51.7%** | ≥80% on **core** packages | ⚠️ Gap |
| Test functions | 158 across 32 files | — | ✅ Strong count, uneven coverage |
| Binary (linux/amd64, stripped) | ~9.4MB | <20MB | ✅ |
| Startup (`clockify version`) | ~8ms | <50ms | ✅ |
| Root README | ❌ Missing | Required by M10 | ⬜ |
| Man pages | ❌ Missing | Required by M10 | ⬜ |
| Integration tests | ❌ None | Required by M10 | ⬜ |
| Git remote | ❌ None | Blocks live publish (not code) | ⚠️ Operational |
| Lint in CI | ❌ `golangci-lint` in Makefile only | Add to CI | ⬜ |
| Error format vs spec §5.1 | Partial (no `Learn more` URL; numeric `ERROR [N]` vs `ERROR: … E001`) | Align or document ADR | ⚠️ Audit |

---

## 1. Sprint Structure

| Sprint | Theme | Days | Goal |
|--------|-------|------|------|
| **S1** | Coverage + Error + Security Audit | D1 | Core packages ≥80% coverage; API/method httptest gaps closed; resolve/output covered; error-format ADR + high-traffic CLIError migration; security checklist green |
| **S2** | Documentation | D2 | Root README (install, config, top workflows incl. M11 daily loop); man pages from help text; command/error reference; docs match real flags |
| **S3** | Integration + Performance + v1.0.0 | D3 | Gated integration suite (`//go:build integration`); performance audit (Makefile + recorded numbers); CI hardening; PLAN.md closeout; tag **`v1.0.0`** |

**Why this split:** S1 hardens the code before docs freeze a surface that still changes; S2 is pure writing/generation once the surface is final; S3 adds live-API confidence and the release ritual. S1 is the only sprint that can still change behavior (error format, redaction) — do it before man pages/README quote error text.

---

## 2. Current State Assessment

### Already Implemented (M0–M11)

| Area | Status | Notes |
|------|--------|-------|
| Unit tests | ✅ | 158 `func Test*`; all packages green; `pkg/clockify` 86.1%, `pkg/duration` 88.3% |
| Core CLI surface | ✅ | Full command tree registered in `internal/cli/root.go` |
| Error type + exit codes | ✅ | `internal/errors`: E001–E010, `CLIError`, `ExitCode()`, `Format()` |
| Keychain + masking | ✅ | `go-keyring`; `maskAPIKey`; config written `0600` |
| Perf budgets | ✅ | ~9.4MB / ~8ms — **no perf work needed to pass**, only to document + CI-guard |
| Distribution scaffolding | ✅ | GoReleaser, CI smoke, npm wrapper, brew formula draft |
| M11 ergonomics | ✅ | Names, human times, pickers, clone/split, aliases, presets — **must appear in README/man** |

### Coverage Gaps (core focus)

| Package | Coverage | Gap driver | M10 action |
|---------|----------|------------|------------|
| `pkg/clockify` | 86.1% | — | Keep; minor edge tests only |
| `pkg/duration` | 88.3% | — | Keep; TZ/DST edge cases if cheap |
| `internal/config` | 63.3% | paths, merge precedence | Push to ≥80% |
| `internal/errors` | 63.6% | format variants, As | Push to ≥80% |
| `internal/cli` | 51.4% | large surface; many thin RunE paths | Target ≥70% (see ADR: “core” = config/output/errors/duration/credentials/client) |
| `internal/update` | 46.0% | asset find / skip matrix | Push to ≥80% |
| `internal/output` | 39.3% | `template.go`, `yaml.go`, `detect.go` edge cases | Push to ≥80% |
| `internal/api` | 44.7% | Many methods **0%**: `resources.go` list/get/create/update/delete, M7 Get/Update, `DoWithPagination`, `handleErrorResponse`, `WithRateLimit/Verbose/Debug` | Table-driven httptest for methods (thin wrappers — high ROI) |
| `internal/resolve` | **8.7%** | Only `IsUUID` tested | Cover fuzzy/prefix resolution properly |
| `cmd/clockify` | 0% | `main` | Optional: extract + test exit-code mapping; not a “core” gate |

**“Core packages” definition for the ≥80% success criterion (PLAN M10):**
`internal/config`, `internal/output`, `internal/errors`, `internal/update`, `pkg/duration`, `pkg/clockify` (already), plus `internal/api` **client core** (`client.go`, `retry.go`, `ratelimit.go`, `regions.go`, `reports.go` error paths) and `internal/resolve`.  
`internal/cli` is measured and reported but gated at **≥70%** — full 80% on every command is diminishing returns; httptest + help/validation tests cover the contract.

### Error-Message Reality vs Spec §5.1

| Spec says | Implementation today |
|-----------|----------------------|
| `ERROR: <concise description>` | `ERROR [%d]: %s` (`internal/errors/format.go`) — numeric code, not `E001` |
| Detail + Suggested fix blocks | ✅ Present when `Detail`/`Suggestion` set |
| `Learn more: https://clockify.github.io/cli/errors/E001` | ❌ Never printed |
| Categories E001–E010 | ✅ Enum exists; most CLI paths use `fmt.Errorf` (~250+ call sites) so users see raw text, no code |

**M10 decision required (ADR in S1):** Align formatter to spec (`ERROR [E001]: …` + `Learn more:` URL using the real docs site you control — `clockify.github.io` is the **spec's fictional/official docs host**; either point at your repo’s `docs/errors/E001.md` or omit Learn more and document the deviation). Recommendation: **use E-code in brackets + Learn more → GitHub docs path in this repo** (`https://github.com/kim/clockify-cli/blob/main/docs/errors/E001.md`) so the link is real. Keep numeric `ExitCode()` unchanged (CI contract).

**CLIError migration policy (S1):** Do **not** mass-replace `fmt.Errorf`. Migrate only:
1. User-input validation (missing required flags, bad enums) → E004
2. Auth/config failures → E001 / E007
3. Not-found after resolution → E002
4. Conflict (timer already running, duplicate create without `--if-not-exists`) → E008
5. Ambiguous name (already done in M11 for resolvers) — keep

Everything else may remain plain errors; document that formal codes cover **auth/config/validation/notfound/conflict** paths.

### Security Baseline (checklist, mostly good)

| Check | Status | Notes |
|-------|--------|-------|
| Secrets in source/repo | ✅ None found | No hardcoded keys |
| Config file mode | ✅ `0600` | `internal/config/config.go:180` |
| Status masks API key | ✅ | `maskAPIKey` |
| Keyring preferred | ✅ | Env fallback documented |
| Verbose/debug header logging | ⚠️ | `X-Api-Key` set on requests; confirm **debug body/header dumps redact the key** (spec §5.3 shows `ck_live_****xxxx`) — fix in S1 if unredacted |
| `npm` install script | ✅ | Downloads release assets; no secrets |
| CI secrets | ✅ | No Clockify key in workflows today |

---

## 3. Prerequisites & Blockers

| # | Item | Impact | Action |
|---|------|--------|--------|
| P1 | **Test workspace + API key** | S3 integration suite | User provides `CLOCKIFY_API_KEY` + workspace ID (or creates a throwaway Clockify workspace). Suite **skips** when env missing — CI stays green without secrets |
| P2 | **Error docs URL decision** | S1 formatter + S2 docs | ADR: repo `docs/errors/` vs omit Learn more. Plan assumes **repo docs** (always linkable offline) |
| P3 | **golangci-lint** | S3 CI | Install or use `go vet` only if lint unavailable; prefer adding lint job |
| P4 | **git remote / GitHub repo** | Live brew/npm/update (operational, not code) | Out of M10 code scope; README documents install paths that work without remote (`go install`, binary) |
| P5 | **Working tree** | S1 start | Confirm clean @ `7c70a1a` before first commit |

---

## 4. Spec ↔ Implementation Decision Log

| Topic | Spec Says | Reality | M10 Decision |
|-------|-----------|---------|--------------|
| Coverage target | ≥80% core packages | 51.7% overall; uneven | Gate on **core set** defined §2; report CLI separately |
| Error format | `ERROR: …` + Learn more | `ERROR [N]: …` no Learn more | **ADR S1:** print `ERROR [E001]: msg` + optional Learn more → repo docs; keep exit codes |
| Learn more host | `clockify.github.io/cli/errors/E001` | Not this project’s docs | Use **real repo URLs** under `docs/errors/`; do not pretend to own clockify.github.io |
| Integration tests | Against real API | None | `//go:build integration`; env-gated; never run in default CI |
| Man pages | Generated from help text | None | `cobra/doc` or `go-md2man` → `man/clockify.1` + subcommands; Makefile `make man`; ship in release archives optional |
| README | Install + top workflows | Missing; only `npm/README.md` | Root README is source of truth; npm/README stays wrapper-specific |
| Binary/startup | <20MB / <50ms | Pass already | Document measured numbers in README + PLAN; add CI smoke timing optional |
| Lint | — | Makefile only | Add `golangci-lint` (or `go vet`) to `ci.yml` |
| Plugin system | Out of scope until post-M10 | — | Still out of scope |
| PDF export | May be unsupported | API-dependent | Document as “if API returns PDF”; no code work |

---

## 5. Architecture Notes

### 5.1 Coverage strategy (S1)

**Pattern already used:** httptest `http.HandlerFunc` + `api.NewClient(..., server.URL)` in `internal/api/*_test.go` and CLI tests via factories.

**Priority order (ROI):**
1. `internal/resolve` — pure functions; small; currently 8.7%
2. `internal/output` — `template.go`, `yaml.go`, `detect.go` auto-detect, empty/edge tables
3. `internal/api/resources.go` + M7 Get/Update/Delete — one table test over method name → expected path + method (httptest assert `r.URL.Path`, `r.Method`)
4. `internal/api/client.go` — `handleErrorResponse`, `DoWithPagination` (X-Pagination headers), `WithRateLimit` (45/s is slow to test fully — test option wiring + single Take)
5. `internal/config` / `internal/update` / `internal/errors` — fill remaining branches (precedence, cache TTL, asset name fallbacks, formatCLIError Detail/Suggestion)
6. `internal/cli` — targeted: validation errors, `--confirm` paths, template/ids output, completion help — **not** every command E2E

**Makefile addition:**
```make
cover:
	go test ./... -coverprofile=cover.out
	go tool cover -func=cover.out | tail -1
```
CI may upload `cover.out` as artifact; hard-fail threshold optional (see ADR).

### 5.2 Error format alignment (S1)

```text
ERROR [E001]: Invalid API key

  Key was rejected by the API.

  Suggested fix: Run `clockify configure` and paste a fresh key.

  Learn more: https://github.com/kim/clockify-cli/blob/main/docs/errors/E001.md
```

- `ErrorCode.String()` → `"E001"`…`"E010"`
- `Format` / `formatCLIError` updated; unit tests updated in `internal/errors`
- `docs/errors/E001.md` … `E010.md` (S2 can flesh out; S1 can stub titles)
- `ExitCode()` mapping **unchanged**

### 5.3 Man page generation (S2)

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **`spf13/cobra/doc`** | Uses live `rootCmd`; subcommand tree accurate | Generates many files | **Preferred** — `make man` writes `man/docs/clockify.1` + children |
| `go-md2man` | Simple roff | Still needs a doc tree | Fallback |
| Hand-written | — | Stale risk | Reject |

**Makefile target:** `make man` → `go run ./cmd/clockify` doc gen or small `internal/docs` generator binary (avoid adding docs code into the released binary — use `//go:build ignore` script or separate `tools/` package).

**Release:** Optionally include `man/` in GoReleaser `archives` extra files; not required for v1.0.0 code-complete.

### 5.4 Integration test harness (S3)

```go
//go:build integration

package integration_test

// Requires: CLOCKIFY_API_KEY, CLOCKIFY_WORKSPACE_ID (or configure profile)
// Skips if missing; cleans up created resources with defer.
```

**Suite scope (minimum viable for M10):**
1. Auth: `GetUser` succeeds
2. Read: `ListWorkspaces`, `ListProjects` (page 1)
3. Time entry: create short entry on test project → get → delete (cleanup)
4. CLI smoke: `configure` is interactive — **skip**; instead call API client + a few CLI commands via `exec.Command` with `CLOCKIFY_API_KEY` env
5. Reports: optional summary for today (may be empty) — prove HTTP path works

**Never** run in default `go test ./...` (build tag). Document in README “Development → Integration tests”.

### 5.5 Performance audit (S3)

Record in PLAN.md + README “Performance”:

```bash
make build
ls -la clockify
/usr/bin/time -f '%e' ./clockify version
# optional: hyperfine './clockify version' if installed
```

CI smoke already runs `version`; optionally add size guard step (`test $(stat -c%s clockify) -lt 20000000`).

---

## 6. Sprint 1: Coverage + Error + Security (Day 1)

### User Stories

**US-1.1: Core packages meet the coverage bar**

> As a maintainer, I want config/output/errors/update/resolve/api-client at ≥80% so regressions in the foundation are caught in CI.

**Acceptance Criteria:**
- [ ] `internal/resolve` ≥80% (fuzzy, prefix, multi-match, not-found)
- [ ] `internal/output` ≥80% (template, ids, yaml, csv edge, detect)
- [ ] `internal/config` ≥80% (precedence, profiles, cache TTL envelope)
- [ ] `internal/errors` ≥80% (format branches, As, FromHTTPStatus)
- [ ] `internal/update` ≥80% (skip matrix, asset fallback darwin_all)
- [ ] `internal/api` client core ≥80% (Do retry/rate-limit hooks, handleErrorResponse, DoWithPagination, reports error path); resource wrappers get path/method httptest (overall api package target ≥70%+)
- [ ] `pkg/clockify` + `pkg/duration` remain ≥85%
- [ ] `go test ./...` green; `make cover` prints total
- [ ] Coverage numbers recorded in this file §6.1

**US-1.2: Errors are actionable and match the agreed format**

> As a user, I want consistent `ERROR [E00X]:` messages with suggestion + learn-more where applicable.

**Acceptance Criteria:**
- [ ] ADR decided and implemented (E-code + Learn more URL policy)
- [ ] `internal/errors` format unit tests updated
- [ ] High-traffic validation/auth/notfound/conflict paths use `CLIError` (not mass rewrite)
- [ ] Spec deviation (if any) documented in PLAN.md §5 / this file §4
- [ ] Spot-check: `clockify project create` without `--name` (or invalid `--status`) shows code + suggestion; auth failure shows E001 + configure hint

**US-1.3: Secrets never leak in logs or status**

> As a user, I want `--verbose`/`--debug` and `status` to never print a full API key.

**Acceptance Criteria:**
- [ ] Debug/verbose path redacts `X-Api-Key` (mask like status)
- [ ] Security checklist §2 complete; fixes applied for any gaps
- [ ] Unit test: redaction helper + debug output does not contain raw key (httptest)

### Tasks

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.1 | `internal/resolve` tests to ≥80% | 1h | `internal/resolve/*_test.go` | cover |
| T1.2 | `internal/output` template/yaml/detect/ids tests | 1.5h | `internal/output/*_test.go`, new `template_test.go` | cover |
| T1.3 | API resources + M7 Get/Update/Delete httptest (path/method) | 2h | `internal/api/resources_test.go`, `m7_resources_test.go` | cover |
| T1.4 | Client core: handleErrorResponse, DoWithPagination, options, rate limit | 1.5h | `internal/api/client_test.go`, `ratelimit_test.go` | cover |
| T1.5 | config/errors/update fill gaps | 1.5h | respective `*_test.go` | cover |
| T1.6 | Error format ADR + implement `ErrorCode.String` + format + tests | 1h | `internal/errors/*`, `docs/errors/E00*.md` stubs | unit + manual |
| T1.7 | Security: redact API key in verbose/debug; checklist | 1h | `internal/api/client.go`, `status.go` if needed | unit |
| T1.8 | CLIError migration (validation/auth/notfound/conflict hotspots) | 2h | `internal/cli/{project,client,task,time_entry,m7_helpers,config}.go` | manual smoke |
| T1.9 | Makefile `cover` target; record baseline in this file | 0.5h | `Makefile`, this file | `make cover` |

**Sprint 1 exit criteria:** core coverage gates green OR remaining gaps explicitly listed with owners; error format decided + implemented; security checklist green; `go test ./...` green.

---

## 7. Sprint 2: Documentation (Day 2)

### User Stories

**US-2.1: New user can install and track time in 10 minutes**

> As a Clockify user, I want a README that gets me from zero to `clockify in` without reading the spec.

**Acceptance Criteria:**
- [ ] Root `README.md` exists
- [ ] Install: `go install`, binary, Homebrew/npm **honestly marked** (tap/remote pending)
- [ ] Quickstart: configure → status → project list → `in`/`out` → `time-entry list`
- [ ] Config: profiles, regions, precedence (flag > env > config > default), keychain note
- [ ] Top workflows (≥10), **including M11**: names, `HH:MM` times, pickers (`-i`), `clone`/`split`, `current`/`last`, report presets, `task done`, templates/ids, `--output json | jq`
- [ ] CI recipe: `--output json --quiet --no-color` + exit codes table (2 auth, 3 notfound, 10 validation, …)
- [ ] Performance numbers + budgets
- [ ] Development: tests, integration gate, man pages, release tags
- [ ] Error codes table linking `docs/errors/`

**US-2.2: Man pages match the real CLI**

> As a Unix user, I want `man clockify` (and subcommands) generated from the same Cobra tree.

**Acceptance Criteria:**
- [ ] `make man` generates man pages without errors
- [ ] Pages include global flags, common aliases (`in`/`out`/`ge`/`ce`, `lp`, `cfg`), M11 commands
- [ ] README “Manual pages” section documents install path (`$XDG_DATA_HOME/man` or `MANPATH`)
- [ ] Man output not stale: generated at release time via Makefile

### Tasks

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.1 | README: structure + install + quickstart | 2h | `README.md` | read-through |
| T2.2 | README: workflows (daily driver + M11 + reports + CI) | 2h | `README.md` | copy-paste dry run |
| T2.3 | README: config/profiles/regions/output/exit codes/perf/dev | 1.5h | `README.md` | checklist |
| T2.4 | `docs/errors/E001–E010.md` (cause, exit code, fix, example) | 1.5h | `docs/errors/` | links from README + format.go |
| T2.5 | Man page generator + Makefile + README section | 1.5h | `tools/man.go` or `internal/docs`, `Makefile` | `make man` |
| T2.6 | Cross-check every README example against current flags (`--help`) | 1h | `README.md` | manual smoke |

**Sprint 2 exit criteria:** README + man pages + error docs exist; all examples verified against binary; no “coming soon” for M11 features (they exist).

---

## 8. Sprint 3: Integration + Performance + v1.0.0 (Day 3)

### User Stories

**US-3.1: Live API smoke suite proves the daily loop**

> As a maintainer, I want an opt-in integration suite that exercises auth, read, time-entry lifecycle against a real workspace.

**Acceptance Criteria:**
- [ ] `//go:build integration` tests skip cleanly without env
- [ ] With `CLOCKIFY_API_KEY` + workspace: user/projects read pass; time-entry create→get→delete cleanup passes
- [ ] README documents how to run: `CLOCKIFY_API_KEY=… CLOCKIFY_WORKSPACE_ID=… go test -tags integration ./integration/...`
- [ ] Default `go test ./...` unchanged (no integration)
- [ ] Optional CLI exec smoke via built binary

**US-3.2: Performance budgets are measured and guarded**

> As a maintainer, I want recorded binary size and startup, and CI that fails if they regress wildly.

**Acceptance Criteria:**
- [ ] Numbers measured on clean build and written to PLAN.md §M10 + README
- [ ] CI optional: size check `<20MB`; smoke `version` already present
- [ ] No perf regression introduced by S1/S2 (re-measure)

**US-3.3: v1.0.0 is tagged and PLAN is closed**

> As a user, I want a clear v1.0.0 milestone that means “stable daily-driver CLI”.

**Acceptance Criteria:**
- [ ] `go test ./...` green; lint/vet green
- [ ] Coverage gates recorded
- [ ] PLAN.md M10 marked COMPLETE; sequencing table M10 → ✅; last-updated footer
- [ ] Git tag **`v1.0.0`** on release commit
- [ ] Milestone tags table already lists `v1.0.0` for M10 — confirm message

### Tasks

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.1 | Integration package + harness + cleanup helpers | 2h | `integration/` (new), README dev section | skip-without-env; live run if creds |
| T3.2 | Perf script / Makefile target + record numbers | 0.5h | `Makefile`, PLAN.md, README | `make perf` |
| T3.3 | CI: vet + test + (lint) + optional size guard | 0.5h | `.github/workflows/ci.yml` | CI green locally simulated |
| T3.4 | GoReleaser dry-run sanity (archives include man optional) | 0.5h | `.goreleaser.yml` | `goreleaser release --snapshot --clean --skip=publish` if available |
| T3.5 | PLAN.md closeout + tag `v1.0.0` | 0.5h | `PLAN.md` | `git tag -l v1.0.0` |

**Sprint 3 exit criteria:** Integration harness lands (live run if user provides P1); budgets re-measured; PLAN.md closed; `v1.0.0` tagged.

---

## 9. Risk Register (M10-specific)

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Chasing 80% on every CLI command wastes days | Schedule slip | Medium | Core-package definition §2; CLI ≥70%; httptest over manual E2E |
| Integration tests need a paid/real workspace | Suite incomplete | High | Env-gated skip; user-provided throwaway workspace; never block unit CI |
| Error format change breaks scripts that parse `ERROR [N]` | User breakage | Low–Med | Keep exit codes; document format in README; bump minor/major note in v1.0.0 release notes (first stable — good time) |
| Docs drift from flags | Misleading users | Medium | T2.6 flag cross-check; man generated from Cobra at release |
| Mass `fmt.Errorf` → CLIError refactor | Regressions | Medium | Explicit non-goal; only hotspots per US-1.2 |
| No git remote | Cannot verify live update/brew | Known | Out of code scope; README honest about install paths |
| `golangci-lint` not installed | Lint gate soft | Low | CI installs linter; fallback `go vet ./...` |

---

## 10. Success Criteria (from PLAN.md M10)

- [x] Unit test coverage ≥80% on **core** packages — api 90.2%, config 90.3%, errors 98.4%, output 95.1%, resolve 100%, update 84.1%, clockify 86.1%, duration 88.3%
- [x] Integration test suite present and env-gated — live run deferred (no `CLOCKIFY_API_KEY` / remote in this environment)
- [x] Binary starts in <50ms (measured warm ~0–50ms)
- [x] Binary size <20MB (measured ~11.5MB linux/amd64)
- [x] README covers install, config, top 10 workflows
- [x] Error messages follow `ERROR [E00X]` + Learn more → `docs/errors/E00X-shortname.md`
- [x] Man pages generated from help text (`make man`)
- [x] Security audit complete (keychain, 0600 config, masked status, redacted debug)
- [x] `go test ./...` green
- [x] Tag `v1.0.0`

---

## 11. Out of Scope

- Plugin system (post-v1.0.0 / post-M10 per PLAN §5.4)
- Publishing to npm/Homebrew/creating git remote (operational; needs user)
- PDF report export beyond passthrough
- Achieving 100% coverage or rewriting the entire CLI to CLIError
- Graphical output, charts
- Changing module path or breaking ID/RFC3339 script compatibility

---

## 12. Closeout Checklist

- [x] `make cover` core gates green; numbers in PLAN.md §M10 / §6.2
- [x] Error format ADR + `docs/errors/*` linked from README (short names E00X-shortname.md)
- [x] Security checklist signed off; debug redaction tested (`MaskAPIKey`)
- [x] README + `make man` complete; man/ + docs/cli/ generated
- [x] `integration/` skips without env; live suite run explicitly deferred (needs creds + remote)
- [x] Perf numbers recorded; budgets still met (~11.5MB, ~0–50ms)
- [x] PLAN.md §M10 ✅ COMPLETE; §3.3 timeline M10 ✅; footer updated
- [x] Git tag `v1.0.0`
- [ ] (User) push remote + first GitHub Release when ready — unblocks brew/npm/`clockify update` live E2E

---

*Last updated: 2026-10-05 — M10 **COMPLETE** → `v1.0.0`. Core coverage ≥80%; README + man pages + error docs; integration harness env-gated; CI size/coverage. Remaining: public remote + Release for live package-manager E2E.*
