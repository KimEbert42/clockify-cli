# Commit Review — 6342b33

- **Commit:** 6342b33e1b48f04ce311d158b0e16e6a9892032a
- **Subject:** feat: M0 - Project scaffolding and core infrastructure
- **Date:** 2026-09-21
- **Experts engaged:** Architect, Full Stack Developer, Backend Engineer
- **Skill mode:** Auto-Assignment (kims-team)
- **Review date:** 2026-10-05
- **Related plan:** PLAN.md §5.7 Review Backlog #3 (P0)

## Diff summary

Single large M0 foundation commit (34 files, +2612 lines): Go module `github.com/kim/clockify-cli`, Makefile, sprint plan, and internal packages for API client (retry, rate limit, regions), CLI (root/config/configure/status/aliases/completion), YAML config + profiles + env, stub keychain, CLIError E001–E010, multi-format output, resource resolver, empty `pkg/duration` stub, plus unit tests for config/errors/client/output/resolve/cli.

**Notable structural gap at this tree:** no `cmd/` package or `main.go` (verified via `git ls-tree -r 6342b33`), despite `Makefile` targeting `./cmd/clockify` and the commit message claiming a working &lt;50ms / &lt;20MB binary.

## Findings

All three experts independently reviewed the tree at `6342b33`. Findings below are **aggregated**; severity is the consensus level.

### Critical (all experts agree)

| # | Issue | Location | Why it matters |
|---|--------|----------|----------------|
| 1 | **No entrypoint — binary cannot be built.** `cli.Execute()` exists but nothing calls it; no `os.Exit` wiring for `CLIError.ExitCode()`. | tree @ `6342b33` (no `cmd/`); `Makefile` builds `./cmd/clockify` | M0 DoD (buildable CLI, `--help`, startup/size claims) unmet at this commit. |
| 2 | **Keychain is env-only stub; auth does not persist.** `setKeyring`/`getKeyring` only `os.Setenv`/`os.Getenv("CLOCKIFY_API_KEY")`. Configure prints “saved successfully” but the key dies with the process. | `internal/config/credentials.go` | Overclaims durable credential storage; M0 auth criteria fail across invocations. |
| 3 | **`config set`/`unset` wipe profiles and can persist secrets.** Save path builds `ProfileConfig{Version, Default}` without the `profiles` map; env-overloaded `APIKey` can be written into `config.yaml`. | `internal/cli/config.go` (~93–152); `internal/config/config.go` env merge | Data loss for multi-profile; CI/env secrets become durable on-disk credentials. |
| 4 | **Flag precedence not wired.** Documented `flag > env > config > default`; `MergeFlags` exists and is unit-tested but never called from any command. Global flags are inert. | `internal/cli/root.go` flags; `internal/config/config.go` `MergeFlags` | Users/CI believe flags work; they silently do not. Blocks correct M1+ command behavior. |
| 5 | **Error formatting path broken.** `Format` calls `As(err, &cliErr)` (`**CLIError`) but custom `As` only accepts `target.(*CLIError)` — rich `ERROR [E00X]` path never runs. Also no unwrap of `%w`. | `internal/errors/format.go` | Every error falls back to `ERROR: %v`; CI contract and learn-more UX lost. Tests do not assert output shape (suite stays green). |
| 6 | **HTTP resilience incorrect.** Rate limiter sleeps under mutex, refills full burst (not sustained 45/50 rps); retry ignores `ctx.Done()`, retries all transport errors, reuses request body across attempts, `defer resp.Body.Close()` inside retry loop. | `internal/api/{ratelimit,retry,client}.go` | Connection leaks, retried writes with empty bodies, thundering herd, no cancellation. Hold-firm path per sprint plan. |

### Major (strong consensus)

| # | Issue | Location |
|---|--------|----------|
| 7 | **Configure wizard ignores workspace selection** — always stores `workspaces[0]["id"]`; half-edited `_choice` branch. | `internal/cli/configure.go` ~60–72 |
| 8 | **403 mapped to E001Auth; E003Permission dead code** — CI cannot distinguish bad key vs insufficient role. Exit codes documented as 0–12; implementation is 2–11 + default 1. | `internal/errors/errors.go` |
| 9 | **Config load swallows all I/O/YAML errors** — corrupt config looks like defaults. | `internal/config/config.go` |
| 10 | **Resolver gaps** — `isUUID` is length-only; prefix match returns first hit without uniqueness; untested beyond `isUUID`. | `internal/resolve/resolver.go` |
| 11 | **Completion invalid shell exits 0** — no `ValidArgs`; `completion foo && echo ok` succeeds. | `internal/cli/completion.go` |
| 12 | **Verbose/debug logs go to stdout** — corrupt piped JSON/CSV. | `internal/api/client.go` |
| 13 | **Aliases inert** — register against commands that do not exist yet; silent no-op. | `internal/cli/aliases.go` |
| 14 | **Tests overstate coverage** — non-asserting `TestDetectFormat`; no credentials/rate-limiter/body-retry/profile-wipe tests; shared global `rootCmd`. | various `*_test.go` |

### Minor

- Empty `pkg/duration` stub committed without API surface.
- YAML tag options leak into CSV map keys (`api_key,omitempty`).
- Table prints `<nil>` for missing values; CSV write errors ignored.
- Unknown region silently falls back to `global`.
- `--api-key` flag is a security footgun (ps/history/logs).
- XDG default config path inconsistent when `XDG_CONFIG_HOME` unset.
- Makefile `lint` requires absent `.golangci.yml`; no CI workflow at this commit.
- Verbose logs / status can look “configured” without valid auth.

## Recommendations

**P0 — foundation blockers (should have closed before “M0 complete”):**

1. Land `cmd/clockify/main.go` mapping `CLIError.ExitCode()` → `os.Exit`; print via `errors.Format` to stderr.
2. Replace keyring stub with a real `CredentialStore` (OS keyring) + env fallback; never treat `os.Setenv` as storage; strip `api_key` from YAML saves.
3. Make `config set/unset` profile-aware and non-destructive (read-modify-write full `ProfileConfig`; regression test that `profiles:` survives).
4. Wire precedence once (`config.Resolve` / `PersistentPreRunE`: flag &gt; env &gt; profile &gt; default); make global flags actually apply.
5. Fix `errors.As`/`Format` (stdlib `errors.As`, unwrap chains); assert full output in tests.
6. Fix rate limiter (time-based refill, sleep outside lock; prefer 50 rps per spec) and HTTP retry (body rewind via `GetBody`, close/drain per attempt, honor `Retry-After`, respect context, idempotency awareness).

**P1 — before M1 auth / M3 writes:**

7. Fix configure workspace picker + validate region.
8. Map 403 → `E003Permission`; publish one exit-code table and test it as the CI contract.
9. Surface config I/O errors (`E007Config`); fail loud on corrupt YAML.
10. Harden resolver (hex UUID, unique prefix, ambiguity → CLIError + candidates).
11. Completion: `ValidArgs` for shells; non-zero exit on invalid input.
12. Logs → stderr; raise test bar (`-race`, profile wipe, secret non-persistence, body retries, error format shape).

**Already addressed in later milestones (hindsight — not scored against this commit):** real keyring (M1 `7214ea1`), time-entry/duration (M3), broader CRUD and reports (M4–M7), packaging/`main.go` distribution path (M9). Reviewing **on the commit’s own merits** vs its message and M0 PLAN criteria, the verdict stands.

## Conflicts / unresolved

| Topic | Positions | Resolution |
|-------|-----------|------------|
| Keyring stub vs M0 DoD | Sprint “cut corners” table vs “hold firm” security; commit implements env-only anyway | **Confidence-weighted → Critical defect.** Configure’s “saved successfully” overclaims; DoD requires durable auth or an explicit re-scope note. |
| Exit codes 0–12 vs 2–11+1 | Docs/spec vs implementation | **Open product choice** — implement denser mapping or publish the real table; do not leave CI users guessing. |
| `--api-key` flag | Spec convenience vs keychain-first security | **Open** — keep with warning for CI, or remove once keyring+env are solid. |
| Alias target naming | Spec verb-first vs map noun-first vs Cobra tree | **Deferred** until resource commands exist; silent skip is a footgun. |
| Missing `cmd/` | Intentional “same milestone later” vs accidental | **Does not matter for verdict** — tree fails DoD either way at this SHA. |
| Rate limit 45 vs 50 | Code vs spec/sprint | Prefer **50/s** once limiter is corrected. |

No expert contradicted another on the overall verdict. Backend’s `Format`/`As` `**CLIError` claim was **verified** against the tree at this commit.

## Verdict

**needs rework**

Solid modular skeleton (cmd/internal/pkg layout, formatter interface, error type shape, region map, config precedence *design*, test scaffolding), but M0’s core promises are unmet at `6342b33`: no runnable entrypoint, credentials do not persist, profile saves destroy data, flag precedence is unwired, error formatting never takes the rich path, and HTTP retry/rate-limit paths are incorrect on hold-firm quality bars. Later milestones appear to address much of this; this commit should not be treated as a green M0 baseline without the P0 fixes documented above.

## HEAD triage (2026-10-05)

- **Current HEAD:** `ac64914`
- **Original verdict:** unchanged (`needs rework`)
- **Legend:** ✅ completed at HEAD · ⚠️ partially fixed · ❌ still open · ➖ N/A or cannot verify

| # | Item | Status | Evidence |
|---|------|--------|----------|
| C1 | No entrypoint — binary cannot be built | ✅ | `cmd/clockify/main.go` exists; `errors.As` + `CLIError.ExitCode()` → `os.Exit` (`cmd/clockify/main.go:20-27`); `go build ./cmd/clockify` succeeds |
| C2 | Keychain env-only stub; auth does not persist | ✅ | Real OS keyring via `github.com/zalando/go-keyring`; `SetAPIKey`→`keyring.Set` (`internal/config/credentials.go:39,61-63`); `GetAPIKey` env fallback then keyring (`:21-28`) |
| C3 | `config set`/`unset` wipe profiles / persist secrets | ✅ | `config set/unset` now `LoadRaw` + mutate full `ProfileConfig` + `Save` (`internal/cli/config.go:140-197,217-260`); `Profiles` map preserved (`internal/config/config.go:200-216`); no `api_key` write path; env not merged in `LoadRaw` |
| C4 | Flag precedence not wired | ❌ | `MergeFlags` still only referenced from tests (`internal/config/config.go:183-198`); no production `GetString("api-key"/"workspace"/"region")`; `loadAuth`/`loadAuthWithConfig` read config/keyring only (`internal/cli/workspace.go:154-166`, `internal/cli/user.go:165-177`); error strings still say "set --workspace" but flag is inert |
| C5 | Error formatting path broken | ⚠️ | Custom `As` now accepts `*CLIError` **and** `**CLIError` (`internal/errors/format.go:44-48`); tests assert full `Format` shape (`internal/errors/errors_test.go:96-115`); **but** `Format` is never called from production — `main.go` prints `Error: %v` (`cmd/clockify/main.go:21`); `As` still does not walk `%w` |
| C6 | HTTP resilience incorrect | ❌ | Rate limiter still `time.Sleep` under mutex + full burst refill (`internal/api/ratelimit.go:33-34`); `defer resp.Body.Close()` still inside retry loop (`internal/api/client.go:120,212`); `Do()` ignores `ctx.Done()`, no `GetBody` rewind, no `Retry-After`; default limiter still 45 (`client.go:74`). Only `reports.go` postReport is better (ctx select, fresh body reader) |
| M7 | Configure wizard ignores workspace selection | ✅ | Full picker: list + numbered choice + bounds check (`internal/cli/configure.go:65-89`) |
| M8 | 403 mapped to E001Auth; E003Permission dead | ❌ | `FromHTTPStatus` still `status == 401 \|\| status == 403 → E001Auth` (`internal/errors/errors.go:134-135`); `docs/errors/E003-permission.md:18` and `README.md:232` still document E003/403→exit 4; `ExitCode()` still 2–11 + default 1 (`errors.go:105-129`) |
| M9 | Config load swallows I/O/YAML errors | ⚠️ | `Load` still ignores read/parse errors (`internal/config/config.go:45-57`); `LoadRaw` surfaces them (`:200-216`) and is used by `config set/unset/list/delete` — but `status`, `workspace`, `time-entry`, auth paths still call silent `Load` |
| M10 | Resolver gaps | ⚠️ | `isUUID` still length-only (`internal/resolve/resolver.go:56-58`); prefix match still first-hit (`:25-31`); name path now lists candidates but returns plain `fmt.Errorf` not `CLIError` (`:42-53`). Separate M11 path has hex `looksLikeID` + `CLIError` (`internal/cli/m11_resolve.go:21-58`) |
| M11 | Completion invalid shell exits 0 | ❌ | `ValidArgs` set (`internal/cli/completion.go:27`) but no `Args: cobra.OnlyValidArgs`; `Run` discards error via `_ = runCompletionScript` (`:30`); live check: `completion foo` exits **0** |
| M12 | Verbose/debug logs go to stdout | ❌ | `fmt.Printf` still used for verbose/debug (`internal/api/client.go:103,124,195,215`); `MaskAPIKey` redaction added but stream still stdout |
| M13 | Aliases inert | ⚠️ | Targets now real commands (`internal/cli/aliases.go:5-14`); `registerAliases` copies only `Run`/`RunE` (`:21-26`) — target local flags (e.g. `time-entry create --start/--project`) are not registered on alias commands |
| M14 | Tests overstate coverage | ⚠️ | `TestDetectFormat` still non-asserting — `wantType` computed then discarded (`internal/output/formatter_test.go:77-95`); Format shape tests improved ✅; no `ratelimit_test.go`; `TestClientRetry` is GET-only (no body rewind); no CLI-level profile-wipe/secret tests; `Makefile test` is plain `go test` (no `-race`) |
| N1 | Empty `pkg/duration` stub | ✅ | `pkg/duration/{duration,human,relative}.go` + tests present |
| N2 | YAML tag options leak into CSV map keys | ✅ | Options stripped at `internal/output/csv.go:73-76` (`tag[:i]` on comma) |
| N3 | Table prints `<nil>`; CSV write errors ignored | ❌ | `table.go:31` still `fmt.Sprintf("%v", row[key])` → `<nil>` for missing keys; `csv.go:18,28` ignore `writer.Write` errors |
| N4 | Unknown region silently falls back to `global` | ❌ | `GetRegion` still returns `Regions["global"]` (`internal/api/regions.go:19-24`); silent fallback now codified by test (`regions_test.go:29-36`) — behavior unchanged |
| N5 | `--api-key` flag is a security footgun | ❌ | Still a persistent flag (`internal/cli/root.go:70`) despite keyring being live; conflict "remove once keyring solid" not resolved |
| N6 | XDG default config path inconsistent when `XDG_CONFIG_HOME` unset | ➖ | `internal/config/paths.go:14-20` is now standard XDG (`$XDG_CONFIG_HOME/clockify` vs `~/.clockify` fallback); accepted pattern, not a live defect |
| N7 | Makefile `lint` requires absent `.golangci.yml`; no CI workflow | ✅ | `Makefile:61-62` `golangci-lint run \|\| go vet ./...` (no config required); `.github/workflows/ci.yml` + `release.yml` exist |
| N8 | Verbose logs / status can look “configured” without valid auth | ⚠️ | `status` now has `Valid` field (`valid`/`invalid`/`unknown`) (`internal/cli/status.go:45-56,106-117`); still prints workspace/region/output from config regardless; verbose path still stdout |
| R1 | Land `cmd/clockify/main.go` with `ExitCode` + `errors.Format` | ⚠️ | Entrypoint + `ExitCode` → `os.Exit` ✅ (`cmd/clockify/main.go:20-27`); `errors.Format` never wired — rich `ERROR [E00X]` never printed by binary |
| R2 | Real `CredentialStore` + strip `api_key` from YAML saves | ⚠️ | Keyring + env fallback ✅; `config set`/`configure` do not write env key to YAML; but `Config.APIKey` still has `yaml:"api_key"` with no strip logic on `Save` (`internal/config/config.go:13,172-181`) |
| R3 | Make `config set/unset` profile-aware + regression test | ⚠️ | Code path ✅ (`LoadRaw`/`Save` preserves `Profiles`); library tests exist (`config_test.go` TestSaveAndLoadWithProfiles) but **no CLI-level regression test** that `profiles:` survives `clockify config set` (no `internal/cli` config test file) |
| R4 | Wire precedence once; make global flags apply | ❌ | Same as C4 — `MergeFlags` uncalled; `--api-key`/`--workspace`/`--region` inert |
| R5 | Fix `errors.As`/`Format`; assert full output in tests | ⚠️ | `As` handles `**CLIError` ✅; full-shape assertions in `errors_test.go:96-128` ✅; `%w` unwrap still missing; production still does not call `Format` |
| R6 | Fix rate limiter + HTTP retry | ❌ | Same defects as C6 still in `ratelimit.go`/`client.go` `Do()`/`DoWithPagination()`; no `GetBody`, no ctx cancel in loop, no `Retry-After`, defer-in-loop |
| R7 | Fix configure workspace picker + validate region | ⚠️ | Picker ✅; region string accepted with no validation against `api.Regions` (`internal/cli/configure.go:41-49`) |
| R8 | Map 403 → `E003Permission`; publish one exit-code table | ❌ | Mapping still 403→E001; table **is** published (`README.md:225-239`) but documents dead E003/exit-4 path — CI contract still false for 403 |
| R9 | Surface config I/O errors (`E007Config`); fail loud on corrupt YAML | ⚠️ | `LoadRaw` fails loud ✅ on set/unset/list/delete; `Load` still silent for status/workspace/auth commands |
| R10 | Harden resolver (hex UUID, unique prefix, CLIError) | ⚠️ | Name-ambiguity candidates ✅ in `internal/resolve`; hex check + CLIError only in separate M11 helpers; `internal/resolve` prefix uniqueness + `isUUID` hex still open |
| R11 | Completion: `ValidArgs`; non-zero exit on invalid input | ❌ | `ValidArgs` present but not enforced as `Args` validator; `Run` swallows error; live `completion foo` → exit 0 |
| R12 | Logs → stderr; raise test bar (`-race`, profile wipe, secret non-persistence, body retries, error format shape) | ❌ | Logs still stdout (C12); error-format shape tests ✅; Makefile `test` has no `-race`; no CLI profile-wipe/secret-persist/body-retry tests |

### Summary
- Completed: 7 · Partial: 13 · Still open: 13 · N/A: 1

### Completed at HEAD
- **C1** Entrypoint landed — `cmd/clockify/main.go` wires `CLIError.ExitCode()` → `os.Exit`; `go build ./cmd/clockify` succeeds.
- **C2** Real OS keyring (`zalando/go-keyring`) replaces env-only stub; env remains a read fallback only.
- **C3** `config set`/`unset` are profile-aware via `LoadRaw`/`Save`; `profiles:` preserved; no secret-write path.
- **M7** Configure workspace picker fully implemented (numbered selection + validation).
- **N1** `pkg/duration` has a real API + tests (no longer empty).
- **N2** CSV map keys strip YAML tag options (`,omitempty` etc.).
- **N7** CI workflows exist; Makefile lint no longer depends on a missing `.golangci.yml` (falls back to `go vet`).

### Still open at HEAD
- **C4 / R4** — `MergeFlags` never called; `--api-key`/`--workspace`/`--region` declared but unread; precedence still not wired.
- **C6 / R6** — rate limiter still sleeps under mutex with burst refill; `Do()` retry still reuses body, `defer`s `Body.Close()` in-loop, ignores ctx/`Retry-After`; default rate still 45.
- **M8 / R8** — 403 still mapped to `E001Auth`; `E003` remains dead code while README/docs claim exit 4.
- **M11 / R11** — invalid completion input still exits 0 (`ValidArgs` not enforced as Args validator; error discarded).
- **M12** — verbose/debug logs still `fmt.Printf` to stdout (corrupt piped JSON/CSV).
- **N3** — table still prints `<nil>` for missing fields; CSV `Write` errors still ignored.
- **N4** — unknown region still silently falls back to `global` (now with a test asserting the silence).
- **N5** — `--api-key` persistent flag still present despite live keyring.
- **R12** — logs still not stderr; test bar still missing `-race` in Makefile, CLI profile-wipe, secret non-persistence, and body-retry tests.
- **Partial (⚠️) remainders:** C5/R1/R5 `Format` never used in production + no `%w` unwrap; M9/R9 `config.Load` still swallows corrupt-YAML errors on primary commands; M10/R10 `internal/resolve` prefix/`isUUID` gaps (M11 has a parallel path); M13 aliases omit target local flags; M14 `TestDetectFormat` still non-asserting; R2 no explicit `api_key` strip on save; R3 no CLI-level profiles-survive regression test; R7 no region validation in configure; N8 status still shows config values when key unconfigured.
