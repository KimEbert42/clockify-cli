# Commit Review — 872f91b

- **Commit:** 872f91b70b71ce1cf5cbedf664edd39b2572ad84
- **Subject:** feat: M1 complete — multi-profile, cache, status validation
- **Date:** 2026-09-21
- **Experts engaged:** Backend Engineer, Full Stack Developer
- **Skill mode:** Auto-Assignment (kims-team)
- **Review date:** 2026-10-05
- **Related plan:** PLAN.md §5.7 Review Backlog

## Diff summary

12 files, +1275/−218. Completes M1 Sprint 2 on top of the prior keyring commit (`7214ea1`):

- **Multi-profile:** `configure --profile`, `config list` prints all profiles, `config delete --profile` (confirm + keyring cleanup), `config set/unset --profile` isolate changes per profile.
- **Cache:** `internal/config/cache.go` + `paths.go` — `CacheDir`, `ProfileCacheDir`, `Write/Read/CacheExists/ClearCache` (XDG-aware); `config clear-cache` wired with `--profile`.
- **Status validation:** live `GET /user` check with `valid` / `invalid` / `unreachable` indicators; JSON output adds `valid` field; API key masked.
- **Testability:** `config` commands refactored to factory functions (`newConfigCmd` et al.); +21 unit tests (37 total). `go test ./...` verified green at this SHA.
- **PLAN.md** updated to mark M1 complete (`v0.2.0-m1`).

## Findings

### Critical

None. Builds, tests pass, and the headline M1 success criteria in PLAN.md (keychain storage, masked status, `--profile` context switch, env-var auth, precedence chain) are implemented.

### Major

1. **`clear-cache` default profile breaks cache isolation** — `internal/config/paths.go:33-39` maps `""`/`"default"` to the cache *root*; `ClearCache` (`internal/config/cache.go:13-21`) does `os.RemoveAll` on that root. So `config clear-cache` (no profile) deletes **all** profiles' caches (named-profile subdirs live under the root), while the CLI message just says "Cache cleared." — `internal/cli/config.go` `newConfigClearCacheCmd`. Users who run the default command believe they only cleared one profile.
2. **`configure` wipes unrelated profile fields** — `internal/cli/configure.go` builds `newCfg` with only `Region` + `WorkspaceID` and assigns it wholesale to `pc.Default` / `pc.Profiles[profile]`. Re-running `configure` (or `configure --profile work`) silently discards `Output`, `DateFormat`, `TimeFormat`, `WeekStarts` previously set via `config set`. This commit's own diff introduced the wholesale profile assign; merge-on-save should preserve non-prompted keys.
3. **`config.Load` swallows read/parse errors** — `internal/config/config.go` `Load()`: `os.ReadFile` and `yaml.Unmarshal` failures are ignored (`if err == nil` chains), so a corrupt `config.yaml` silently behaves like defaults. `LoadRaw` correctly returns parse errors; `Load` should too (or at least warn), otherwise users get confusing "empty" config with no diagnostic.
4. **`status` JSON `valid` field deviates from M1 acceptance criteria** — M1-SPRINT-PLAN US-2.3 specifies `--output json` includes `valid: true/false`; `statusOutput.Valid` is a **string** enum (`"valid"|"invalid"|"unreachable"|"unknown"`) — `internal/cli/status.go:20`. More informative, but breaks the stated schema and any consumer written against the plan. Decide: keep string + document, or add boolean `valid` alongside `valid_status`.
5. **Network/error classification by string matching is fragile** — `validateKeyStatus` (`internal/cli/status.go:89-103`) inspects `err.Error()` substrings ("context deadline exceeded", "connection refused", …). The API client already wraps failures with `internal/errors` codes (`E006Network`, `E001Auth`) — `internal/api/client.go`. A TLS/proxy/DNS-class error that lacks those substrings is reported as **invalid** ("re-authenticate") even when the key is fine. Use `errors` code / `errors.As` instead.
6. **Milestone claim vs. incomplete US-1.2 offline path** — the commit message and PLAN mark M1 complete, but Sprint 1 acceptance "Network error: warn but allow save (`--offline`)" (T1.10) is unimplemented: `configure` hard-fails on any `validateAPIKey` error, including network failures. Validation call sites themselves came from `7214ea1`; this SHA *declares* the milestone done while that criterion is still open.

### Minor

1. **No tests for live status validation path** — T2.15 asked for status-validation tests; coverage only exercises `"unknown"` and `formatValidIndicator` unit cases (`internal/cli/cli_test.go`). The `valid`/`invalid`/`unreachable` classification (httptest mock of `/user`) is untested — and is exactly where the string-matching bug lives.
2. **Partial-profile display inconsistency** — `config set --profile work <key>` on a non-existent profile stores a zero-value `Config` then one key; `config list` prints raw profile values (empty fields) while `config get --profile work <key>` merges via `Load` and shows inherited defaults. Same profile, two answers.
3. **"Refactor CLI commands to factory functions" is partial** — only `config.*` commands became factories; `configureCmd`, `statusCmd`, `completionCmd` remain package-level vars (`internal/cli/root.go`), limiting test isolation for those commands (status/configure tests mutate shared state carefully instead).
4. **`config delete` confirmation reads `os.Stdin` via `fmt.Scanln`** — `internal/cli/config.go` `newConfigDeleteCmd`; bypasses the `bufio.Reader` pattern used in `configure`, ignores `--quiet`, and is untestable.
5. **Profile names used unsanitized in filesystem/keyring paths** — `ProfileCacheDir` and keyring user IDs join raw profile strings (`../../x` or names with `/` can escape the cache dir or collide). Low practical risk for a personal CLI; worth a `filepath.Base`/charset guard.
6. **Cache write API is dead code at this SHA** — `WriteCache`/`ReadCache`/`CacheExists` have no production callers yet (only tests). Fine as M1 scaffolding (criteria only required `clear-cache`), but note there is nothing cached to clear until later milestones.

## Recommendations

1. Give `ClearCache` a true per-profile contract: either namespace `default` under `ProfileCacheDir("default")` like other profiles, or on default clear only default-owned entries (and fix the confirmation message to say what was removed). *Not addressed in later milestones as of this review date.*
2. Make `configure` merge into existing profile config (load → update Region/WorkspaceID/API key only → save) so preferences survive re-runs. *Still wholesale-assign at current HEAD.*
3. Propagate `Load` read/parse errors (or log a warning + exit non-zero when `CLOCKIFY_DEBUG`/verbose). *Still silent at current HEAD.*
4. Resolve the `valid` JSON schema: either follow the plan (`bool`) or amend M1-SPRINT-PLAN/PLAN.md to document the string enum; pick one before writing more consumers.
5. Replace substring matching in `validateKeyStatus` with `internal/errors` code checks (`E006Network` → unreachable; `E001Auth` → invalid; else unknown/error).
6. Implement `configure --offline` (or: warn-and-allow on network errors only) to close US-1.2; add httptest coverage for status validation classification.
7. When touching profile set/unset/list, make raw vs merged display consistent (print effective config in `list`, or note "inherited from default").

## Conflicts / unresolved

**None material.** Backend Engineer and Full Stack Developer perspectives (synthesized in-session per review protocol; no subagents spawned — both converge):

- **Agreement:** cache-root isolation (finding 1), `configure` field-wipe (2), silent `Load` errors (3), and error-classification fragility (5) are the priority defects.
- **Minor tension:** Backend prefers `valid` as an explicit status string for diagnostics; Full Stack prefers the plan's boolean for scripting ergonomics. Confidence-weighted toward the plan's acceptance criteria → treat as schema decision flagged for user choice (finding 4), not a hard defect on its own.
- No escalation required.

## Verdict

**approve with follow-ups**

The commit delivers M1 Sprint 2's multi-profile, cache-clear, and status-validation features; tests pass at the SHA (verified) and PLAN's five M1 success criteria are met. However "M1 complete" overstates the milestone (offline configure path still missing), and several correctness defects — default-profile cache wipe, configure field-wiping, silent config parse failures, fragile validation classification — warrant follow-up fixes; they remain open in current HEAD.

## HEAD triage (2026-10-05)

- **Current HEAD:** `ac64914` (feat: implement M10 - Quality, Testing & Documentation (v1.0.0))
- **Original verdict:** unchanged — **approve with follow-ups**
- **Legend:** ✅ completed at HEAD · ⚠️ partially fixed · ❌ still open · ➖ N/A or cannot verify

| # | Item | Status | Evidence |
|---|------|--------|----------|
| M1 | `clear-cache` default profile wipes cache root (Major 1) | ❌ | `internal/config/paths.go:38-44` still maps `""`/`"default"` to `CacheDir()` root; `ClearCache` (`internal/config/cache.go:13-25`) `RemoveAll`s that root; `newConfigClearCacheCmd` (`internal/cli/config.go:316-339`) still prints plain "Cache cleared." for default. `TestClearCache` (`internal/config/cache_test.go:44-64`) enshrines root deletion. |
| M2 | `configure` wholesale-assigns profile, wiping non-prompted fields (Major 2) | ❌ | `internal/cli/configure.go:101-113` still builds `newCfg = config.Config{Region, WorkspaceID}` and assigns it wholesale to `pc.Default` / `pc.Profiles[profile]`. No merge-on-save. |
| M3 | `config.Load` swallows read/parse errors (Major 3) | ❌ | `internal/config/config.go:42-63` still uses `if err == nil` chains for `os.ReadFile` and `yaml.Unmarshal`. Worse: `TestLoadIgnoresCorruptConfig` (`internal/config/config_extra_test.go:257-269`) asserts the silent-defaults behavior as intended. |
| M4 | status JSON `valid` is string enum vs plan's bool (Major 4) | ❌ | `internal/cli/status.go:20` still `Valid string \`json:"valid"\`` with enum values. `M1-SPRINT-PLAN.md:173` still specifies `valid: true/false`; no doc amendment found (PLAN.md:121 mentions `valid` field but not its type). Schema decision still unresolved. |
| M5 | Network/error classification by substring matching (Major 5) | ❌ | `validateKeyStatus` (`internal/cli/status.go:89-104`) still inspects `err.Error()` substrings. API client now wraps failures as `internal/errors` codes (`E006Network` at `internal/api/client.go:117,148`; `E001Auth` via `FromHTTPStatus` at `client.go:153-160`), but status never uses `errors.As`/codes. |
| M6 | `configure --offline` / US-1.2 offline path (Major 6) | ❌ | `internal/cli/configure.go:52-55` still hard-fails on any `validateAPIKey` error; no `--offline` flag anywhere in `internal/cli/`. M1-SPRINT-PLAN:291 documents a product decision to *skip* the flag ("always validate"), but the criterion itself remains unimplemented. |
| m1 | No tests for live status validation path (Minor 1) | ❌ | `internal/cli/cli_test.go` still only exercises `"unknown"` (`TestStatusNoAPIKey`, `TestStatusJSONOutput`) and `formatValidIndicator` unit cases. No httptest mock of `GET /user` for `valid`/`invalid`/`unreachable` classification — the code path behind M5 remains untested. |
| m2 | Partial-profile display inconsistency raw vs merged (Minor 2) | ⚠️ | Improved: `config list --profile X` now uses merged `config.Load(profile)` (`internal/cli/config.go:56-63`), consistent with `config get`. Still raw: bare `config list` prints `LoadRaw` per-profile values (`config.go:65-82`) without inherited-default merge/annotation. |
| m3 | Factory functions partial — `configureCmd`/`statusCmd` still package vars (Minor 3) | ⚠️ | `completionCmd` converted to `newCompletionCmd()` (`internal/cli/completion.go:12`, wired `root.go:89`). `configureCmd` (`configure.go:16`) and `statusCmd` (`status.go:24`) remain package-level vars (`root.go:87-88`). |
| m4 | `config delete` confirmation uses `fmt.Scanln`, ignores `--quiet` (Minor 4) | ⚠️ | Improved: `--confirm` flag added (`internal/cli/config.go:288,312`) and tested non-interactively (`TestConfigDeleteProfile`, `cli_test.go:274`). Still open: interactive path uses `fmt.Scanln` (`config.go:293`), not `bufio.Reader`; `--quiet` still ignored. |
| m5 | Profile names unsanitized in filesystem/keyring paths (Minor 5) | ❌ | `ProfileCacheDir` still `filepath.Join(base, profile)` with raw name (`internal/config/paths.go:38-44`); keyring user id still `profile + "-api-key"` raw (`internal/config/credentials.go:17-18,35-37,47-49`). No `filepath.Base`/charset guard in config paths. |
| m6 | Cache write API dead code at M1 SHA (Minor 6) | ✅ | Obsolete at HEAD: TTL cache has production consumers — `rememberCompletionNames`/`cachedCompletionNames`/`rememberList`/`cachedList` call `config.WriteTTL`/`ReadTTL` (`internal/cli/resource_cache.go:84,91,101,110`). Nothing cached to clear is no longer true. |
| R1 | Recommendation 1: per-profile `ClearCache` contract (default ≠ wipe all) | ❌ | Same evidence as M1 — no contract change at HEAD. |
| R2 | Recommendation 2: `configure` merge-on-save | ❌ | Same evidence as M2 — still wholesale-assign. |
| R3 | Recommendation 3: propagate `Load` read/parse errors | ❌ | Same evidence as M3 — still silent; test asserts silence. |
| R4 | Recommendation 4: resolve `valid` JSON schema (bool vs string) | ❌ | Same evidence as M4 — string kept; plan not amended to document it. |
| R5 | Recommendation 5: `internal/errors` code checks in `validateKeyStatus` | ❌ | Same evidence as M5 — substrings unchanged despite client wrapping E001/E006. |
| R6 | Recommendation 6: `configure --offline` + httptest status-validation tests | ❌ | Same evidence as M6 + m1 — neither the flag nor the live-path tests landed. |
| R7 | Recommendation 7: raw vs merged display consistency | ⚠️ | Same evidence as m2 — `list --profile` merged; bare `list` still raw. |

### Summary

At HEAD `ac64914` (M10, v1.0.0), **19 triaged items**: 1 ✅ completed, 4 ⚠️ partially fixed, 14 ❌ still open. The original "approve with follow-ups" verdict is unchanged. The original review's priority defects (cache-root wipe, configure field-wipe, silent `Load`, fragile validation classification) are all **still open**; the API client matured (M10 `internal/errors` codes) but `status` never adopted them. Minor polish landed incrementally: completion factory, `config delete --confirm`, merged `list --profile`, and real TTL-cache consumers (closing the dead-code note).

### Completed at HEAD

- **m6 / R-cleanup — cache write API not dead code:** TTL cache is consumed in production by M8 resource caching (`internal/cli/resource_cache.go:84,91,101,110` → `config.WriteTTL`/`ReadTTL`). The original caveat ("nothing cached to clear until later milestones") is obsolete; note this does *not* fix the default-profile wipe (M1/R1).

### Still open at HEAD

- **M1 / R1 — default `clear-cache` wipes all profiles' caches.** Root mapping + `RemoveAll` + generic message unchanged; `TestClearCache` still asserts root deletion.
- **M2 / R2 — `configure` still wholesale-assigns `{Region, WorkspaceID}`,** discarding `Output`/`DateFormat`/`TimeFormat`/`WeekStarts`/etc. set via `config set`.
- **M3 / R3 — `Load` still swallows read/parse errors;** a unit test now *documents* the corrupt-config→defaults fallback as expected behavior.
- **M4 / R4 — status JSON `valid` remains a string enum;** M1-SPRINT-PLAN still states `valid: true/false`; no amendment documenting the string choice.
- **M5 / R5 — `validateKeyStatus` still classifies by `err.Error()` substrings,** ignoring the `E001Auth`/`E006Network` codes the API client now produces.
- **M6 / R6 — `configure --offline` still unimplemented;** network failures hard-fail configure. Plan de-scoped the flag but the US-1.2 acceptance criterion is not met.
- **m1 — no live status-validation tests** (httptest `/user` mock for valid/invalid/unreachable classification).
- **m2 / R7 — bare `config list` still prints raw profile values** while `get`/`list --profile` show merged/effective config.
- **m3 — `configureCmd` and `statusCmd` remain package-level vars** limiting test isolation.
- **m4 — `config delete` interactive confirm still `fmt.Scanln`,** `--quiet` still ignored (non-interactive path via `--confirm` is fine).
- **m5 — profile names still unsanitized** in `ProfileCacheDir` and keyring user ids (`../../x` or `/` can escape/collide).
