# Commit Review — 91b362b

- **Commit:** 91b362bc07f99494c9765f8eb57ee2d92fa90463
- **Subject:** feat: M9 - Distribution & Packaging
- **Date:** 2026-10-04
- **Experts engaged:** DevOps Engineer, Security DevOps Engineer
- **Skill mode:** Auto-Assignment (kims-team)
- **Review date:** 2026-10-05
- **Related plan:** PLAN.md §5.7 Review Backlog

## Diff summary

Implements M9 Distribution & Packaging (+2380/−43 across 22 files). In-session expert review of the tree at this SHA (not HEAD).

- `.goreleaser.yml` (new): v2 config — CGO off, linux/darwin/windows × amd64/arm64 (skip windows/arm64), universal darwin, tar.gz/zip archives, sha256 `checksums.txt`, changelog filters, `kim/clockify-cli` GitHub release, Homebrew tap publish to `kim/homebrew-tap`
- `.github/workflows/ci.yml` (new): vet + test + build + `./clockify version` smoke on push/PR
- `.github/workflows/release.yml` (new): GoReleaser on `v*` tags; `permissions: contents: write`; secrets `GITHUB_TOKEN` + `HOMEBREW_TAP_GITHUB_TOKEN`
- `Makefile`: ldflags retargeted to `-X main.version/commit/date`; `release-snapshot` target
- `cmd/clockify/main.go`: prints errors to stderr; maps `CLIError.ExitCode()`
- `internal/update/` (new): GitHub Releases client, `Check`/`Install`, tar/zip extract, atomic-ish binary replace, `ShouldSkip`/`MaybeCheck` passive check with 24h TTL cache; unit tests via httptest
- `internal/cli/update.go` (new): `clockify update` / `--check` / `--version`; passive check hook; dev-build refusal (exit 10)
- `internal/cli/root.go`: `PersistentPreRun` passive update check; enriched `version` (commit/date/go/platform + `--output json`)
- `internal/config/config.go` + `internal/cli/config.go`: `auto_update` key (empty = enabled) + `ParseAutoUpdate`
- `npm/`: wrapper package scaffold (`install.js` postinstall downloads release binary; `package.json` bin link)
- `PLAN.md` + `M9-SPRINT-PLAN.md`: milestone plan/status

Matches PLAN.md M9 deliverable table and sprint plan at unit level (pipeline present, update/check/version implemented, skip rules tested, ldflags fix correct, version JSON). Live brew/npm/update publish remains blocked on remote + first tag (documented in PLAN).

## Findings

### Critical

1. **Self-update replaces the running binary with no integrity check — Security DevOps**
   - Location: `internal/update/update.go` — `Install` / `downloadFile` (download asset → extract → `ReplaceExecutable`)
   - Issue: HTTPS to `browser_download_url` is the only trust gate. GoReleaser already publishes `checksums.txt` (sha256), but `Install` never fetches or verifies it, and there is no signature/provenance check. A compromised release asset, account, or MITM path that serves a bad archive silently overwrites the user's CLI.
   - Why it matters: this is the privileged update path for a tool that runs with API keys in the keyring. Checksum verification against `checksums.txt` (and ideally cosign/sigstore later) is table stakes for a self-updater.
   - Note (not scored as fixed): still absent at HEAD in `internal/update/update.go` and `npm/install.js` as of this review date.

2. **npm postinstall command injection via untrusted asset name — Security DevOps**
   - Location: `npm/install.js` — `extractBinary` (`execSync(\`unzip -o -j "${archivePath}" ...\`)` and `execSync(\`tar -xzf "${archivePath}" ...\`)`)
   - Issue: `archivePath` is `path.join(binDir, asset.name)` where `asset.name` comes from the GitHub API JSON response, interpolated into a shell string. A crafted release asset name (or hostile repo/API content) breaks out of the quotes and executes arbitrary commands during `npm install -g`.
   - Why it matters: postinstall runs with the installing user's privileges on every global install; this is a classic npm supply-chain primitive.
   - Mitigation direction: never build shell strings from API data — use `child_process.execFile`/`spawn` with argv arrays, or pure-JS extractors; verify checksums before extract.

### Major

1. **Tag-triggered release likely fails without Homebrew tap token — DevOps**
   - Location: `.goreleaser.yml` `brews:` (token `{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}`); `.github/workflows/release.yml` env
   - Issue: every `v*` tag runs full GoReleaser publish including Homebrew formula push to `kim/homebrew-tap`. If the secret is unset/empty (PLAN notes the tap repo does not exist yet), the release job fails after artifacts may already be partially processed.
   - Why it matters: first public tags will be blocked or flaky until tap repo + secret exist; release is the critical path for brew/npm/`clockify update`.
   - Fix: gate the `brews` section on secret presence (split job / `if:` / release `skip_publish` until tap exists).

2. **Windows npm package is likely broken — DevOps / Full Stack**
   - Location: `npm/package.json` (`"bin": {"clockify": "bin/clockify"}`); `npm/install.js` (win32 path renames extracted binary to `bin/clockify` without `.exe`; no `.cmd` shim)
   - Issue: on Windows the downloaded binary is `clockify.exe`, but the bin path and rename target omit `.exe`, and no npm `.cmd` shim is generated. The package will install without a working `clockify` command.
   - Why it matters: `os: ["win32"]` is declared; Windows users get a silent broken install (postinstall even exits 0 on failure).

3. **Windows self-update replace path is fragile for a running exe — DevOps**
   - Location: `internal/update/update.go` — `ReplaceExecutable` (`os.Rename(src, dst)` then `dst+".new"` rename, then direct `WriteFile(dst, data, 0755)`)
   - Issue: renaming/overwriting a running Windows executable typically fails (file lock). Last-resort direct write to the live path is unreliable; the standard pattern is rename running exe → `.old`, write new binary, optional delete `.old` on next start.
   - Why it matters: `clockify update` is a headline M9 deliverable; failure on the most common package-manager-adjacent path (windows zip) undermines it. Unix rename-over-inode works and is fine.

### Minor

1. **CI does not run `golangci-lint`** — `.github/workflows/ci.yml` vs `Makefile` `lint`/`check`. Vet+test+build only; lint remains local. Low risk once CI size/coverage gates land in M10, but inconsistent with `make check`.

2. **Passive-update exempt list misses completion subcommands** — `internal/update/check.go` `ShouldSkip` exempts `cmdName` in `{update, version, completion, help}`. Cobra sets `cmd.Name()` to the leaf (e.g. `bash`/`zsh` under `completion`), so tab-completion script generation can trigger a network check. Prefer parent-command walk or a completion-prefix check.

3. **`CompareVersions` ignores pre-release suffixes** — `internal/update/update.go` `splitVersion` truncates at first non-digit and drops `-next`; `1.0.0-rc1` compares equal to `1.0.0`. Acceptable for current tag scheme (`v0.9.0-m8` parses as 0.9.0); document or adopt full semver later.

4. **npm install failure exits 0** — `npm/install.js` `main().catch` sets `process.exitCode = 0` with a manual-install hint. Documented tradeoff (offline installs), but leaves a package that claims a `clockify` bin with no binary — easy to misread as success.

5. **No download size cap; weak GitHub rate-limit posture** — `downloadFile`/`install.js` copy unbounded bodies; Go client sets no `User-Agent` (GitHub API prefers one). Minor operational polish.

6. **`FindAsset` fuzzy fallback is loose** — `internal/update/update.go` matches any archive whose name contains os/arch tokens (`all`/`universal` special cases). Prefer exact `clockify_*_{os}_{arch}` / `*_darwin_all` only; fail closed on ambiguity.

7. **`update --check` JSON path lightly tested** — `internal/cli/update_test.go` mostly tests dev-build refusal + `version --output json`; check formatting against httptest lives in `internal/update`, not wired through Cobra end-to-end.

## Recommendations

1. **Before any public release (hard prerequisite):** verify sha256 of downloaded assets against the release `checksums.txt` in both `internal/update.Install` and `npm/install.js`; fail closed on mismatch. Follow-ups can add cosign/sigstore provenance.
2. **Fix npm extract to avoid the shell:** `execFile`/`spawn` with explicit argv, or pure JS tar/zip extraction; never interpolate API-derived strings into `execSync`.
3. **Make Homebrew publish optional** until `kim/homebrew-tap` + `HOMEBREW_TAP_GITHUB_TOKEN` exist (conditional GoReleaser section or release job split), so tag releases still ship binaries/checksums.
4. **Fix Windows packaging:** npm bin → `bin/clockify.exe` on win32 (or generate `.cmd` shim); Go update path: rename running exe aside before write.
5. Wire `golangci-lint` into `ci.yml`; exempt `completion` parent command from passive checks; document `CompareVersions` semver limitations.
6. Note (no score vs this SHA): M10/M11 commits after `91b362b` do not appear to close checksum verification, npm injection, or brew-token gating — treat items 1–3 as still open for post-v1.0.0 distribution work.

## Conflicts / unresolved

No material conflict between the DevOps Engineer and Security DevOps Engineer lenses (in-session review; no long-running subagents spawned, per catch-up constraints).

- **DevOps Engineer:** prioritizes Major #1 (release/token gating) and Major #2/#3 (Windows npm + self-update replace) as operational blockers to the milestone's install-path claims; notes pipeline design (CGO off, universal darwin, checksums artifact, CI smoke) is sound.
- **Security DevOps Engineer:** prioritizes Critical #1 (no binary integrity verification) and Critical #2 (npm shell injection) as must-fix before anyone is told to `npm i -g` or `clockify update` from a public repo; HTTPS-only trust is not enough for a privileged self-updater.

**Resolution (confidence-weighted):** both lenses agree on verdict **approve with follow-ups** — M9 meets its PLAN deliverables (pipeline, update command, skip rules, version enrichment, ldflags fix, tests), and the critical security items are hardening gaps that must land before public distribution rather than evidence the milestone design is wrong. Security DevOps owns the integrity/injection follow-ups; DevOps owns release-token and Windows packaging follow-ups. Unresolved for user decision: whether to treat checksum + npm injection as release-blocking gates (recommended) or accept HTTPS-only until first public tag.

## Verdict

**approve with follow-ups**

M9 delivers GoReleaser + GitHub Actions release path, ldflags main-package fix, enriched `version` (text/JSON), `clockify update`/`--check`/`--version`, passive auto-update with well-tested skip rules, `auto_update` config, npm/Homebrew packaging scaffolds, and unit tests — matching PLAN M9 criteria with documented operational blockers (no remote/tap yet). Honest gaps at this SHA: self-update and npm postinstall lack artifact integrity checks, npm extract is shell-injectable via asset names, release workflow is brittle without the Homebrew token, and Windows npm/self-update paths are incomplete. Verdict is for `91b362b` on its own merits; later milestones (M10/M11) do not clearly close the security or Windows packaging findings.

## HEAD triage (2026-10-05)

- **Current HEAD:** `ac64914`
- **Original verdict:** unchanged
- **Legend:** ✅ completed at HEAD · ⚠️ partially fixed · ❌ still open · ➖ N/A or cannot verify

| # | Item | Status | Evidence |
|---|------|--------|----------|
| Critical #1 | Self-update / npm install: no checksum verification against `checksums.txt` | ❌ still open | `internal/update/update.go:222-272` `Install` still downloads → extracts → `ReplaceExecutable` with no `checksums.txt` fetch or sha256 verify; `npm/install.js` likewise. Repo-wide grep for checksum/sha256/sigstore/cosign hits only `.goreleaser.yml:46-48` (publish-side) and plan/docs. `git diff 91b362b..ac64914 -- internal/update/update.go npm/install.js` = 0 lines. |
| Critical #2 | npm postinstall command injection via `execSync` + API-derived asset name | ❌ still open | `npm/install.js:11` still `require("child_process").execSync`; `:118` `execSync(\`unzip -o -j "${archivePath}" ...\`)`; `:125` `execSync(\`tar -xzf "${archivePath}" ...\`)`. No `execFile`/argv-array change; file unchanged vs 91b362b. |
| Major #1 | Homebrew tap publish not gated on secret presence | ❌ still open | `.goreleaser.yml:67-79` `brews:` section unconditional, token `{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}`; `.github/workflows/release.yml:23-31` always runs `goreleaser release --clean` on every `v*` tag, no `if:` secret check and no `skip_publish`. Both files unchanged vs 91b362b. |
| Major #2 | Windows npm package: bin path missing `.exe`, no `.cmd` shim | ❌ still open | `npm/package.json:5-7` still `"bin": {"clockify": "bin/clockify"}`; `npm/install.js:137` always `binPath = .../bin/clockify`; win32 extracts `clockify.exe` then `fs.renameSync(extracted, binPath)` (`:122-123`) renames it to the extension-less path. Comment at `:138-140` notes npm `.cmd` shim but nothing generates one; `npm/` contains only `README.md`, `install.js`, `package.json`. |
| Major #3 | Windows self-update `ReplaceExecutable` lacks running-exe rename-aside pattern | ❌ still open | `internal/update/update.go:400-422` unchanged: `os.Rename(src, dst)` → write `dst+".new"` → rename → last-resort `os.WriteFile(dst, data, 0755)` on the live path. No rename of running exe to `.old` + write + deferred cleanup. |
| Minor #1 | CI does not run `golangci-lint` | ⚠️ partially fixed | `.github/workflows/ci.yml:21-25` now has `Lint (golangci-lint)` via `golangci/golangci-lint-action@v6`, but with `continue-on-error: true` — wired into CI yet non-blocking, so lint failures do not gate builds (inconsistent with `make check` intent). |
| Minor #2 | `ShouldSkip` exempt list matches completion leaf names only | ❌ still open | `internal/update/check.go:49-52` still `switch cmdName { case "update", "version", "completion", "help": }`; `internal/cli/update.go:41` still passes `cmd.Name()` (Cobra leaf, e.g. `bash`/`zsh`). No parent-command walk or completion-prefix check. |
| Minor #3 | `CompareVersions` ignores pre-release suffixes | ❌ still open | `internal/update/update.go:183-198` `splitVersion` unchanged: truncates at first non-digit; only `-next` suffix trimmed (`:184`). `TestCompareVersionsMore` (`internal/update/update_test.go:275-290`) pins `1.0.0-next == 1.0.0` but no `-rc1`/semver-prerelease handling; only doc is comment `:165` "semver-ish strings". |
| Minor #4 | npm install failure exits 0 | ❌ still open | `npm/install.js:162-167` `main().catch` still sets `process.exitCode = 0` with manual-install hint; package still claims a `clockify` bin when no binary was installed. Unchanged vs 91b362b. |
| Minor #5 | No download size cap; weak GitHub rate-limit posture | ❌ still open | `internal/update/update.go:274-294` `downloadFile` still unbounded `io.Copy(f, resp.Body)`; `getJSON` (`:60-77`) sets `Accept` + `X-GitHub-Api-Version` but no `User-Agent`. npm `User-Agent: clockify-cli-npm` was already present at 91b362b (`git show 91b362b:npm/install.js` lines 40/74) — not a HEAD fix. No size cap on either path. |
| Minor #6 | `FindAsset` fuzzy fallback is loose | ❌ still open | `internal/update/update.go:119-128` unchanged: fallback matches any archive whose name `Contains(goos) && (Contains(arch) \|\| "all" \|\| "universal") && hasArchiveExt` — no exact-match-only policy, no fail-closed on ambiguity. |
| Minor #7 | `update --check` JSON path lightly tested | ❌ still open | `internal/cli/update_test.go:45-71` `TestUpdateCheckJSONAgainstMock` still exercises `version --output json` only; `res` (the `CheckResult`) is built then discarded via `_ = res`. No end-to-end Cobra `update --check --output json` against httptest (client construction still fixed in `newUpdateCmd`). |

### Summary
At HEAD `ac64914` (M10 commit; M10/M11 are the only commits after `91b362b`), none of the two Critical findings or three Major findings were fixed. The only distribution-path delta since `91b362b` is `.github/workflows/ci.yml` gaining a non-blocking `golangci-lint` step plus coverage/size-budget/integration additions, and `internal/update/update_test.go` gaining more unit tests (FindAsset happy paths, `CompareVersions` `-next` pinning). All five files at the center of the security/Windows/release findings — `internal/update/update.go`, `internal/update/check.go`, `npm/install.js`, `npm/package.json`, `.goreleaser.yml`, `.github/workflows/release.yml` — have a zero-line diff vs `91b362b`. Reviewer note in the original text (Rec #6: M10/M11 do not close items 1–3) is confirmed accurate at HEAD.

### Completed at HEAD
- None of the 12 scored findings are fully fixed.
- Minor #1 is partially fixed: `golangci-lint` is now invoked in CI (`ci.yml:21-25`) but with `continue-on-error: true`, so it does not gate merges.

### Still open at HEAD
- **Critical #1** — no sha256 verification against `checksums.txt` in `internal/update.Install` or `npm/install.js` (hard prerequisite before public distribution per original Rec #1).
- **Critical #2** — npm `execSync` shell interpolation of API-derived `archivePath` remains (Rec #2).
- **Major #1** — Homebrew `brews` publish still unconditional; tag releases remain brittle without `HOMEBREW_TAP_GITHUB_TOKEN` / tap repo (Rec #3).
- **Major #2** — Windows npm bin still `bin/clockify` (no `.exe`, no `.cmd` shim) (Rec #4).
- **Major #3** — Windows `ReplaceExecutable` still lacks rename-running-exe-aside pattern (Rec #4).
- **Minor #2** — `ShouldSkip` still matches completion leaf names only (Rec #5).
- **Minor #3** — `CompareVersions` still ignores pre-release suffixes; no documentation of limitation (Rec #5).
- **Minor #4** — `npm/install.js` still exits 0 on postinstall failure.
- **Minor #5** — no download size cap; Go GitHub client still sends no `User-Agent`.
- **Minor #6** — `FindAsset` fuzzy fallback still loose (fail-open on ambiguous asset sets).
- **Minor #7** — `update --check` JSON path still not wired end-to-end through Cobra tests.
