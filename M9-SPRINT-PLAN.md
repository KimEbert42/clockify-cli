# M9: Distribution & Packaging — Sprint Plan

**Milestone:** M9
**Sprint Duration:** 2 × 1-day sprints (compressed from original 1–2 days)
**Developer:** 1 (solo)
**Start Date:** 2026-10-04
**End Date:** 2026-10-05 (target)
**Status:** ✅ **COMPLETE**
**Completed:** 2026-10-04
**Prerequisite:** M8 complete — commit `969ca42` (working tree clean)
**Spec reference:** `clockify-cli-spec.md` §7 (Distribution & Packaging), §8.21 (`update`), §8.22 (`version`)
**PLAN.md reference:** PLAN.md §M9 (lines 518–546), §3.3, §6.1

**Implementation summary:** Shipped GoReleaser + GitHub Actions release pipeline, fixed ldflags clobber (`main.*` injection), enriched `clockify version` (platform/go/JSON), created missing tags `v0.8.0-m7`/`v0.9.0-m8`, implemented `clockify update`/`--check`/`--version`, passive auto-update check (24h TTL, CI/json/quiet/`auto_update` skips), `auto_update` config key, npm wrapper scaffold, and homebrew tap formula config. Snapshot release verified: linux/amd64 ~3.9MB, startup ~10ms. Live brew/npm/update publish awaits a public GitHub remote + first release tag.

---

## 1. Sprint Structure

| Sprint | Theme | Days | Goal |
|--------|-------|------|------|
| **S1** | Build & Release Infrastructure | D1 | Goreleaser config, GitHub Actions release pipeline, `version` command enrichment, ldflags fix, missing git tags, local release dry-run |
| **S2** | Self-Update, Auto-Check & Package Managers | D2 | `clockify update`, auto-update check + cache, Homebrew tap formula, npm wrapper package, E2E verification, PLAN.md update, tag |

**Why this split:** S1 is pure build/packaging infrastructure — no runtime CLI behavior beyond `version`. S2 is the user-facing update/install surface and needs release assets to exist (or mocks) before it can be verified end-to-end.

---

## 2. Current State Assessment

### Already Implemented (M0–M8)

| Feature | Status | Notes |
|---------|--------|-------|
| CLI framework + global flags | ✅ | Cobra; `--output`, `--quiet`, `--no-color`, etc. |
| Build metadata injection | ⚠️ Partial | Makefile has ldflags for `internal/cli.buildVersion/Commit/Date`; `cmd/clockify/main.go` also has package vars |
| `clockify version` | ⚠️ Basic | One-liner: `clockify %s (%s) built on %s` — no platform, go version, or `--output json` |
| Cross-compile targets | ⚠️ Makefile only | `make build-all` covers linux/darwin/windows amd64+arm64; no archives, checksums, or universal macOS |
| GitHub Actions / release pipeline | ❌ | No `.github/` directory |
| GoReleaser config | ❌ | Not present |
| `clockify update` | ❌ | Not implemented |
| Auto-update check | ❌ | Not implemented |
| Homebrew formula / tap | ❌ | Not present |
| npm package wrapper | ❌ | Not present |
| README | ❌ | Not present (M10 owns full README; minimal install section may be needed for brew/npm docs) |
| Git tags `v0.8.0-m7`, `v0.9.0-m8` | ❌ | Missing — M7/M8 commits exist but tags were never created |
| Git remote | ❌ | **No remote configured** — releases + `update` require a public GitHub repo |

### What M9 Needs to Build

| Component | Status | Notes |
|-----------|--------|-------|
| `version` enrichment | ❌ | Spec §7.4 format + JSON; platform, go version; fix ldflags clobber bug |
| GoReleaser config | ❌ | `.goreleaser.yml` — archives, checksums, universal darwin, brew tap |
| GitHub Actions | ❌ | `.github/workflows/release.yml` (tag push) + `ci.yml` (test/lint) |
| `clockify update` | ❌ | `--check`, `--version`, download + replace binary |
| Auto-update check | ❌ | First command of day, 24h TTL cache, skip rules |
| `auto_update` config key | ❌ | `clockify config set auto_update false` |
| Homebrew tap formula | ❌ | Template + publish path (`kim/homebrew-tap` or similar) |
| npm wrapper package | ❌ | `npm install -g clockify-cli` postinstall binary download |
| Makefile fixes | ❌ | ldflags target, dist targets, align with goreleaser |
| Release dry-run | ❌ | Local `goreleaser release --snapshot --clean` |
| Git tags + remote | ❌ | Housekeeping before first real release |

---

## 3. Prerequisites & Blockers

| # | Item | Impact | Action |
|---|------|--------|--------|
| P1 | **No git remote** | Releases, brew, npm, and `update` all need a public GitHub repo | User creates/pushes repo; set remote. Plan assumes `github.com/kim/clockify-cli` (matches `go.mod`) unless user specifies otherwise |
| P2 | **Missing tags** `v0.8.0-m7`, `v0.9.0-m8` | Version history incomplete; first release tag should follow sequence | Create on respective commits during S1 housekeeping |
| P3 | **Homebrew tap repo** | `brew install clockify/tap/clockify` needs a separate tap repo (e.g. `kim/homebrew-tap`) | Create tap repo OR ship formula template and document manual tap setup; GoReleaser can push formula when `HOMEBREW_TAP_GITHUB_TOKEN` is set |
| P4 | **npm org/name** | Spec says `clockify-cli` on npm | User confirms whether to publish under personal scope (`@kim/clockify-cli`) or unscoped `clockify-cli` (name may be taken) |
| P5 | **goreleaser not installed locally** | Cannot dry-run S1 without it | `go install github.com/goreleaser/goreleaser/v2@latest` or download binary |

**Module path decision:** Spec §7.2 says `go install github.com/clockify/cli@latest`. Actual module is `github.com/kim/clockify-cli`. **Decision: use the real module path everywhere** (`go install github.com/kim/clockify-cli/cmd/clockify@latest`). Document the deviation in PLAN.md. Do not rename the module in M9 (would break all import paths).

---

## 4. Bug Fix Required Before Releases

### 4.1 ldflags clobber bug

**Problem:** Makefile injects into `internal/cli` package vars:

```
-X github.com/kim/clockify-cli/internal/cli.buildVersion=$(VERSION)
```

But `cmd/clockify/main.go` then does:

```go
cli.SetBuildInfo(version, commit, date) // version/commit/date default to "dev"/"none"/"unknown"
```

This **overwrites** the ldflags values with defaults. `make build` currently produces binaries that report `dev` even when VERSION is set.

**Fix (recommended):** Inject ldflags into the **main package** (goreleaser convention) and keep `SetBuildInfo` as the single propagation path:

```
-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)
```

Update Makefile LDFLAGS accordingly. GoReleaser `ldflags` uses the same pattern (`-X main.version={{.Version}}` etc.).

**Alternative:** Remove the `SetBuildInfo` call from main and rely solely on ldflags into `internal/cli`. Worse for goreleaser (main package injection is standard) — do not use.

**Verify:** `make build && ./clockify version` shows git describe version, not `dev`.

---

## 5. Spec ↔ Implementation Decision Log

| Topic | Spec Says | Reality / PLAN | M9 Decision |
|-------|-----------|----------------|-------------|
| Go install path | `github.com/clockify/cli@latest` | Module is `github.com/kim/clockify-cli` | Use real module path; document deviation |
| Homebrew | `clockify/tap/clockify` | Needs separate tap repo | Tap repo name: `clockify` org or personal; formula generated by GoReleaser; default assumption `kim/homebrew-tap` with short name `clockify` |
| npm | `npm install -g clockify-cli` | Unscoped name may be taken | Prefer scoped `@kim/clockify-cli` if unscoped unavailable; document both |
| pip / AUR / Snap | Listed in §7.2 | **Not in PLAN.md M9 deliverables** | **Out of scope** — defer to post-M10 backlog |
| `clockify version` format | `clockify-cli 1.2.3` + commit/built/go/platform | Current: `clockify %s (%s) built on %s` | Match spec labels; include go version + platform; support `--output json` |
| `clockify update` | `update`, `--check`, `--version X.Y.Z` | — | Implement all three; binary replace on Unix + Windows-aware download |
| `auto_update` config | `clockify config set auto_update false` | Config struct has no such key | Add `auto_update` (bool, default true) to config; also honor `CLOCKIFY_NO_UPDATE_CHECK=1` and CI env |
| Auto-update timing | First command of the day, cached 24h | Reuse M8 TTL cache pattern | Cache file under profile cache dir (e.g. `update-check.json`); skip network when fresh |
| Exit code 12 | Update required (version incompatible) | Spec appendix | Optional: reserve for future min-version enforcement; **not required** in M9 unless version gate is trivial |
| macOS universal binary | "Universal binary" | Separate amd64/arm64 builds in Makefile | GoReleaser `universal_binaries` on darwin; verify with `lipo -info` when on macOS |
| Windows package | `.exe + .zip` | — | GoReleaser zip archive for windows; binary name `clockify.exe` |
| CI skip rule | No check if `CLOCKIFY_NO_UPDATE_CHECK=1` or in CI | — | Detect `CI=true`, `GITHUB_ACTIONS`, `CLOCKIFY_NO_UPDATE_CHECK` |

---

## 6. Sprint 1: Build & Release Infrastructure (Day 1)

### User Stories

**US-1.1: Fix build metadata injection**

> As a user, I want `clockify version` to show the real release version baked in at build time.

**Acceptance Criteria:**
- [ ] ldflags inject into `main` package vars (`main.version`, `main.commit`, `main.date`)
- [ ] `SetBuildInfo` propagates to `internal/cli` without clobbering
- [ ] `make build` → `./clockify version` shows git describe version
- [ ] `go build ./cmd/clockify` (no ldflags) still works, shows `dev`

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.1 | Move ldflags targets to `main.*`; update Makefile LDFLAGS | 0.5h | `Makefile`, `cmd/clockify/main.go` (no logic change expected) | `make build && ./clockify version` |
| T1.2 | Unit test for version display + JSON shape | 1h | `internal/cli/root.go` (extend versionCmd), `cli_test.go` or `version_test.go` | tests |

**US-1.2: Enrich `clockify version`**

> As a user, I want version output that matches spec §7.4 and is machine-readable.

**Acceptance Criteria:**
- [ ] Table/text output includes: version, commit, built date, go version, platform (`linux/amd64` etc.)
- [ ] `clockify version --output json` returns structured JSON (version, commit, date, go, platform)
- [ ] Global `--output` flag works with version command
- [ ] Help text documents both modes

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.3 | Extend `versionCmd` — text format per spec + JSON via existing `formatOutput` pattern | 1h | `internal/cli/root.go` | Manual + unit |
| T1.4 | `version_test.go` — text contains version/commit/platform; JSON unmarshals | 1h | `internal/cli/version_test.go` | tests |

**US-1.3: Housekeeping — tags and remote**

> As a maintainer, I want the tag sequence intact before the first distribution release.

**Acceptance Criteria:**
- [ ] `v0.8.0-m7` tagged on M7 commit `3168f3e`
- [ ] `v0.9.0-m8` tagged on M8 commit `969ca42`
- [ ] Git remote configured (user action — document expected URL)
- [ ] `go test ./...` still green

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.5 | Create missing tags `v0.8.0-m7`, `v0.9.0-m8` | 0.25h | git | `git tag -l` |
| T1.6 | Confirm/set remote; push tags when remote exists | 0.25h | git | `git remote -v` |

**US-1.4: GoReleaser configuration**

> As a maintainer, I want reproducible multi-platform release builds on tag push.

**Acceptance Criteria:**
- [ ] `.goreleaser.yml` builds: linux amd64/arm64, darwin amd64/arm64 (+ universal), windows amd64
- [ ] Archives: `.tar.gz` for unix, `.zip` for windows
- [ ] Checksums file generated
- [ ] ldflags: `-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}`
- [ ] Snapshot dry-run succeeds locally: `goreleaser release --snapshot --clean`
- [ ] Binary size checked <20MB; startup smoke <50ms (informational)

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.7 | Write `.goreleaser.yml` (builds, archives, checksums, universal_binaries, changelog) | 2h | `.goreleaser.yml` | `goreleaser check` |
| T1.8 | Homebrew section stub (publish in S2; config ready with tap name) | 0.5h | `.goreleaser.yml` | check |
| T1.9 | Snapshot release dry-run + fix issues | 1.5h | dist/ (gitignored) | artifacts exist; `./dist/*/clockify version` |
| T1.10 | Add `dist/` to `.gitignore` | 0.1h | `.gitignore` | clean `git status` |

**US-1.5: GitHub Actions workflows**

> As a maintainer, I want CI on PR/push and releases on version tags.

**Acceptance Criteria:**
- [ ] `.github/workflows/ci.yml`: go test, vet, (golangci-lint if available), on push/PR
- [ ] `.github/workflows/release.yml`: on `v*` tags → goreleaser release with `GITHUB_TOKEN`
- [ ] Release workflow builds all platforms and attaches assets to GitHub Release
- [ ] Secrets documented: `GITHUB_TOKEN` (auto), optional `HOMEBREW_TAP_GITHUB_TOKEN` (S2)

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.11 | `ci.yml` — test + lint matrix (go 1.27) | 1h | `.github/workflows/ci.yml` | push to branch (when remote exists) |
| T1.12 | `release.yml` — goreleaser on tag push | 1h | `.github/workflows/release.yml` | Workflow file valid YAML; dry-run via act optional |
| T1.13 | Makefile `release-snapshot` target wrapping goreleaser | 0.25h | `Makefile` | `make release-snapshot` |

**US-1.6: Makefile alignment**

> As a maintainer, I want local builds and release snapshots to share one ldflags story.

**Acceptance Criteria:**
- [ ] `make build`, `make build-all`, `make test`, `make check` still work
- [ ] `make release-snapshot` documented
- [ ] No references to old `internal/cli.build*` ldflags paths

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.14 | Update Makefile targets + help | 0.5h | `Makefile` | all targets |

**Sprint 1 Exit Criteria:**
- [ ] `make build && ./clockify version` shows real version/commit/platform
- [ ] `clockify version --output json` valid JSON with version, commit, date, go, platform
- [ ] `goreleaser release --snapshot --clean` produces archives + checksums for all target platforms
- [ ] Tags `v0.8.0-m7`, `v0.9.0-m8` exist
- [ ] CI + release workflows present and YAML-valid
- [ ] `go test ./...` green
- [ ] Binary size <20MB (spot-check linux/amd64)

**Sprint 1 Estimated:** ~8h (1 focused day)

---

## 7. Sprint 2: Self-Update, Auto-Check & Package Managers (Day 2)

### User Stories

**US-2.1: `clockify update` command**

> As a user, I want to upgrade the CLI to the latest release (or a specific version) from GitHub Releases.

**Acceptance Criteria:**
- [ ] `clockify update` — checks latest release, downloads matching asset, replaces current binary
- [ ] `clockify update --check` — reports available version; does not install
- [ ] `clockify update --version 1.2.3` — installs specific version
- [ ] Asset selection by GOOS/GOARCH (linux/darwin/windows; darwin arm64 vs amd64)
- [ ] Replace is atomic-ish: download to temp file, chmod +x, rename over current executable
- [ ] Windows: download `.zip`, extract `clockify.exe`, replace
- [ ] Clear errors when: not installed from release binary (dev build), network fail, version not found
- [ ] `--check` JSON output when `--output json` (current, latest, update_available)
- [ ] Unit tests with httptest GitHub Releases API; no live network in tests
- [ ] Exit codes via existing `CLIError` patterns where appropriate

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.1 | GitHub Releases API client (list releases, latest, by tag; parse asset URLs) | 1.5h | `internal/update/github.go` (new pkg or under `internal/cli`) | httptest |
| T2.2 | Binary install: download, verify (optional sha256 from checksums file), replace current exe | 2h | `internal/update/install.go` | Unit + manual on temp copy |
| T2.3 | `internal/cli/update.go` — `update` / `--check` / `--version`; register in root | 1.5h | `internal/cli/update.go`, `root.go` | Manual + unit |
| T2.4 | `update_test.go` — check, version selection, error paths | 1.5h | `internal/cli/update_test.go` | tests |

**US-2.2: Auto-update check (passive)**

> As a user, I want a non-blocking notice when a new version exists — without slowing every command.

**Acceptance Criteria:**
- [ ] On first command of the day (per profile cache), check GitHub Releases for newer version
- [ ] Result cached 24h (reuse `WriteTTL`/`ReadTTL` pattern, filename like `update-check.json`)
- [ ] If newer version found and not quiet: print one-line hint to stderr (or stdout per existing patterns): `A new version of clockify is available (x.y.z). Run 'clockify update' to upgrade.`
- [ ] Skip when: `--quiet`, `--output json` (machine mode — no chatter), `CLOCKIFY_NO_UPDATE_CHECK=1`, `CI=true`, `GITHUB_ACTIONS=true`, `auto_update=false` in config
- [ ] Never fails the primary command due to update-check errors (silent on network failure)
- [ ] Does not run for `version` and `update` commands themselves
- [ ] Unit tests: cache hit skips network; env/CI skips; stale cache triggers check

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.5 | Config: add `auto_update` bool (default true) + env `CLOCKIFY_NO_UPDATE_CHECK` | 1h | `internal/config/config.go`, `config_test.go` | unit |
| T2.6 | `checkForUpdate` hook in `Execute()` (or root PersistentPreRun) with all skip rules | 1.5h | `internal/cli/root.go` or `internal/update/check.go` | unit |
| T2.7 | Hint printing respects quiet/json/no-color | 0.5h | same | manual |
| T2.8 | Tests for check logic + cache TTL | 1h | `internal/update/check_test.go` | tests |

**US-2.3: Homebrew tap**

> As a macOS/Linux user, I want `brew install clockify/tap/clockify`.

**Acceptance Criteria:**
- [ ] Tap repository exists (or formula template in-repo ready to push)
- [ ] GoReleaser `brews` section configured: repository `kim/homebrew-tap` (or user-confirmed), homepage, description, license
- [ ] Formula installs binary from GitHub Release assets
- [ ] First real tag push (or manual formula commit) makes `brew install` work
- [ ] README/docs snippet: `brew tap kim/tap && brew install clockify` (adjust to real tap name)

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.9 | Complete GoReleaser brew config; create tap repo; optional `HOMEBREW_TAP_GITHUB_TOKEN` | 1h | `.goreleaser.yml`, tap repo | Formula file appears after release (or snapshot with `--skip=publish` not applicable — verify config) |
| T2.10 | Document brew install path in PLAN.md / future README | 0.25h | `PLAN.md` | Review |

**US-2.4: npm wrapper package**

> As a cross-platform user, I want `npm install -g clockify-cli` (or scoped equivalent).

**Acceptance Criteria:**
- [ ] Package directory `npm/` (or `dist/npm/`) with `package.json`, install script
- [ ] postinstall (or preinstall) downloads correct binary from GitHub Releases for platform/arch
- [ ] Binary linked into npm global bin as `clockify`
- [ ] Works on linux/darwin/windows (windows: .exe)
- [ ] Version of npm package matches CLI release version
- [ ] Document install command in PLAN.md

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.11 | Scaffold npm package + download script (node, no heavy deps) | 2h | `npm/package.json`, `npm/install.js`, `npm/README.md` | `npm pack` + local install test |
| T2.12 | Test install on current platform; document publish steps (`npm publish`) | 0.75h | — | Manual |
| T2.13 | Name availability check / scope decision (P4) | 0.25h | — | User confirms |

**US-2.5: `go install` path**

> As a Go user, I want `go install github.com/kim/clockify-cli/cmd/clockify@latest`.

**Acceptance Criteria:**
- [ ] Documented install command using real module path
- [ ] `go install` works from a clean machine with Go ≥ 1.27 (module `go 1.27.1`)
- [ ] ldflags not required — version shows `dev` unless built via release/make (acceptable; document)
- [ ] Optionally: `main` package could embed version via build info (`debug.ReadBuildInfo`) as fallback when ldflags absent — **nice-to-have**, only if cheap

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.14 | Document go install; optional buildinfo fallback for version | 0.5–1h | `PLAN.md`, maybe `main.go` | Manual `go install` |

**US-2.6: E2E verification & closeout**

> As a maintainer, I want M9 verified and PLAN.md updated.

**Acceptance Criteria:**
- [ ] All PLAN.md M9 success criteria checked (unit-level where no GitHub remote/token; live where possible)
- [ ] `go test ./...` green
- [ ] `make check` / lint clean if golangci-lint available
- [ ] Snapshot release artifacts: binary size <20MB, `version` correct on at least linux/amd64
- [ ] `update --check` works against mock; works live after first GitHub release exists
- [ ] PLAN.md M9 marked complete; decisions logged; tags section updated
- [ ] Commit + tag `v0.10.0-m9`

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.15 | Full test suite + fix | 1h | all | `go test ./...` |
| T2.16 | CLI smoke: version/update help; update --check offline errors | 0.5h | binary | Manual |
| T2.17 | Update PLAN.md M9 status, deliverables, decisions, timeline | 1h | `PLAN.md` | Review |
| T2.18 | Live E2E when remote exists: tag → GH release → update --check sees it | 1h | — | Checklist |
| T2.19 | Commit + tag `v0.10.0-m9` | 0.25h | git | Tag exists |

**Sprint 2 Exit Criteria:**
- [ ] `clockify update`, `--check`, `--version` implemented and tested
- [ ] Auto-update check: 24h cache, skip rules, non-fatal on error
- [ ] `auto_update` config key + `CLOCKIFY_NO_UPDATE_CHECK` work
- [ ] GoReleaser brew config complete; tap documented
- [ ] npm package scaffolded and locally installable
- [ ] `go install` path documented
- [ ] PLAN.md updated; tests green; tag `v0.10.0-m9`

**Sprint 2 Estimated:** ~10–12h (1 focused day + release E2E buffer)

---

## 8. Architecture Notes

### 8.1 Version data flow

```
git tag v0.10.0-m9
    → GitHub Actions release.yml
    → goreleaser ldflags: -X main.version=0.10.0 -X main.commit=... -X main.date=...
    → main() → cli.SetBuildInfo(...) → versionCmd / update compare
```

### 8.2 Update check flow

```
any command Execute()
  → shouldSkipUpdateCheck()? (quiet|json|env|CI|auto_update|cmd in {version,update})
  → config.ReadTTL(profile, "update-check.json", &cached, 24h) fresh? → return
  → GET github releases latest (short timeout, ~2s)
  → compare semver to buildVersion
  → cache result; if newer && !quiet → stderr hint
  → always return nil (never block primary command)
```

### 8.3 Update install flow

```
clockify update [--check] [--version v]
  → resolve target version (latest or --version)
  → find asset matching runtime.GOOS/GOARCH (+ windows zip)
  → optional: download checksums, verify sha256
  → download to os.TempDir()
  → if zip: extract
  → chmod 0755
  → os.Executable() → rename over current binary (fallback: copy + exec)
  → print success with old → new versions
```

### 8.4 Package layout (new)

```
.goreleaser.yml
.github/workflows/ci.yml
.github/workflows/release.yml
internal/update/          # github client, install, check (if not under cli)
internal/cli/update.go
internal/cli/version_test.go
npm/package.json
npm/install.js
npm/README.md
```

*Prefer `internal/update` for testable logic; CLI file stays thin.*

---

## 9. Risk Mitigation

| Risk | Impact | Likelihood | Mitigation | Sprint |
|------|--------|------------|------------|--------|
| **No GitHub remote / repo** | High — blocks real releases, brew, npm, live update | High (current state) | P1 prerequisite; all S1 work is local-safe (goreleaser snapshot); live E2E deferred until remote exists | All |
| **ldflags clobber** | Medium — wrong versions in releases | Certain (bug exists) | T1.1 fix + test before any tag | S1 |
| **npm name taken** | Low — install path differs | Medium | Scope to `@kim/clockify-cli`; document | S2 |
| **Homebrew tap friction** | Medium — brew users blocked | Medium | Formula template in-repo; tap create is one gh CLI command; document | S2 |
| **Self-update replaces wrong binary** | High — user left without working CLI | Low | Only replace `os.Executable()` path; refuse if not writable; keep download in temp; print clear error; optional checksum verify | S2 |
| **Dev builds (`version=dev`) confuse update** | Medium | High | `update` errors clearly when version is `dev` or not semver: "installed from source; use go install -u or package manager" | S2 |
| **Auto-update slows commands** | Medium | Low | Cache-first; 2s timeout; never fail primary command; skip in CI/json/quiet | S2 |
| **GoReleaser config drift** | Low | Medium | `goreleaser check` in Makefile; snapshot in CI (optional job) | S1 |
| **Windows replace-in-use** | Medium | Low | Document restart; try rename; fall back to writing `.new` + instructions if locked | S2 |
| **Scope creep (pip/AUR/snap/exit-12 gate)** | Medium | Medium | Hard cuts below | All |
| **Module path vs spec path** | Low | Certain | Document real path; never claim `github.com/clockify/cli` | S1/S2 |

---

## 10. Trade-offs: Where to Cut Corners vs. Where to Hold Firm

### Hold Firm (Non-Negotiable)

| Area | Why |
|------|-----|
| **Fix ldflags before first release tag** | Wrong versions undermine update + support |
| **`update` never breaks the primary command** | Auto-check is advisory only |
| **Real module path in all docs** | `go install` must actually work |
| **Checksums on releases** | Security baseline for self-update |
| **Skip auto-update in CI/json/quiet** | Scripts must stay deterministic |
| **`--check` without network mutation** | Safe in scripts |
| **Unit tests for update/check with httptest** | No live GitHub required for green CI |

### Cut Corners (If Behind Schedule)

| Area | How to Cut | Impact |
|------|------------|--------|
| **npm package** | Ship formula + go install first; npm as follow-up | Cross-platform users use brew/script |
| **Homebrew tap automation** | Manual formula commit from snapshot artifacts | Slightly more maintainer work |
| **Checksum verification in `update`** | Skip if checksums asset missing | Slightly lower assurance |
| **buildinfo version fallback** | Accept `dev` for `go install` | Documented |
| **Universal macOS binary** | Ship separate amd64/arm64 only | Users pick arch via brew |

### Cut Corners — Emergency Only

| Area | How to Cut | Impact |
|------|------------|--------|
| **`clockify update` full install** | Ship `--check` only; print download URL | Users update manually |
| **Auto-update check** | Defer entirely to M10 | Spec feature late |
| **npm wrapper** | Defer to post-M9 | One install channel missing |
| **Release workflow** | Manual `goreleaser release` from laptop | Works but not automated |
| **pip/AUR/Snap** | Already out of scope | Fine |

---

## 11. File-Level Task Summary

| File | Sprint | Tasks | Est. Hours |
|------|--------|-------|------------|
| `Makefile` | S1 | T1.1, T1.13, T1.14 | 1.25h |
| `internal/cli/root.go` + version tests | S1 | T1.2–T1.4 | 3h |
| `.goreleaser.yml` | S1/S2 | T1.7–T1.8, T2.9 | 3.5h |
| `.github/workflows/ci.yml` | S1 | T1.11 | 1h |
| `.github/workflows/release.yml` | S1 | T1.12 | 1h |
| `.gitignore` | S1 | T1.10 | 0.1h |
| `internal/update/*` | S2 | T2.1–T2.2, T2.6, T2.8 | 5h |
| `internal/cli/update.go` + test | S2 | T2.3–T2.4 | 3h |
| `internal/config/config.go` | S2 | T2.5 | 1h |
| `npm/*` | S2 | T2.11–T2.13 | 3h |
| `PLAN.md` | S2 | T2.10, T2.17 | 1.25h |
| E2E + tag | S2 | T2.15–T2.16, T2.18–T2.19 | 2.75h |

**Total estimated:** ~19–20h ≈ **2 focused days** (matches PLAN.md 1–2 day estimate at this velocity).

---

## 12. M9 Overall Definition of Done

| Criterion | Verification | Status |
|-----------|-------------|--------|
| ldflags bug fixed; `make build` shows real version | `make build && ./clockify version` | ☐ |
| `clockify version` has commit, go, platform | Manual | ☐ |
| `clockify version --output json` valid JSON | `jq .` | ☐ |
| Tags `v0.8.0-m7`, `v0.9.0-m8` exist | `git tag -l` | ☐ |
| GoReleaser snapshot builds all platforms | `make release-snapshot` | ☐ |
| Archives + checksums produced | `dist/` listing | ☐ |
| Linux amd64 binary <20MB | `ls -la dist/...` | ☐ |
| Binary starts <50ms | `time ./dist/.../clockify version` | ☐ |
| CI workflow present | `.github/workflows/ci.yml` | ☐ |
| Release workflow on `v*` tags | `.github/workflows/release.yml` | ☐ |
| `clockify update --check` implemented | Manual + unit | ☐ |
| `clockify update` installs latest/specific | Manual (needs release) | ☐ |
| Auto-update: 24h cache, skip rules | Unit tests | ☐ |
| `auto_update` config + `CLOCKIFY_NO_UPDATE_CHECK` | Unit tests | ☐ |
| Homebrew formula/tap path documented & configured | GoReleaser brews section | ☐ |
| npm package scaffolded + local install works | `npm pack` + install | ☐ |
| `go install` path documented (real module) | PLAN.md / docs | ☐ |
| `go test ./...` green | CI / local | ☐ |
| PLAN.md M9 marked complete; decisions logged | Review | ☐ |
| Tag `v0.10.0-m9` created | `git tag -l` | ☐ |
| Git remote exists + first release published | `gh release list` (when remote ready) | ☐ |

---

## 13. Out of Scope (M9 Hard Cuts)

- pip, AUR, Snap packages (spec §7.2 lists them; not in PLAN.md M9 deliverables)
- Plugin system (post-M10)
- Exit code 12 / min-version enforcement gate
- Interactive update TUI
- Differential/patch updates
- Code signing / notarization (macOS) — note as future hardening
- SBOM / provenance attestations — future hardening
- Full README (M10 owns docs; M9 only documents install paths in PLAN.md as needed)
- Renaming Go module to `github.com/clockify/cli`

---

## 14. Daily Standup Template

```markdown
## Day X Standup — M9 Sprint Y

**Done today:**
- [List completed tasks]

**Blocked:**
- [Any blockers — e.g. waiting on GitHub remote / tap repo / npm name]

**Plan for tomorrow:**
- [Next tasks]

**Confidence for sprint:**
- 🟢 On track / 🟡 At risk / 🔴 Behind
```

---

## 15. Success Criteria Mapping (PLAN.md §M9 → Sprint Tasks)

| PLAN.md Success Criterion | Sprint | Tasks |
|---------------------------|--------|-------|
| `go install` produces working binary | S2 | T2.14 |
| `brew install` works on macOS | S2 | T2.9–T2.10 (needs remote + tag) |
| `clockify update` upgrades to latest version | S2 | T2.1–T2.4 |
| `clockify version` shows correct build metadata | S1 | T1.1–T1.4 |
| GitHub Actions releases binaries on tag push | S1 | T1.7, T1.12 |
| Binary size <20MB, startup <50ms | S1 | T1.9 (snapshot check) |
| npm package install | S2 | T2.11–T2.13 |

---

*This plan is a living document. Update task estimates and status as work progresses. Live release E2E requires a GitHub remote (P1) and at least one published tag after M9 infrastructure exists.*

*Last updated: 2026-10-04 — M9 implemented and tagged `v0.10.0-m9`. Infrastructure complete; operational publish steps remain (git remote, homebrew-tap repo, npm publish).*
