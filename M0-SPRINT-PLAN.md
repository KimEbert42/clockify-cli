# M0: Project Scaffolding & Core Infrastructure — Sprint Plan

**Milestone:** M0
**Sprint Duration:** 4 × 1-week sprints (5 working days each)
**Developer:** 1 (solo)
**Start Date:** TBD
**End Date:** TBD (4 weeks from start)

---

## 1. Sprint Structure

| Sprint | Theme | Days | Goal |
|--------|-------|------|------|
| **S1** | Skeleton & Config | D1–D5 | Buildable binary, config system, profile support |
| **S2** | Output & Errors | D6–D10 | Formatters, error handling, exit codes |
| **S3** | HTTP Client & API | D11–D15 | Clockify API client, retry, rate limit, configure wizard |
| **S4** | Integration & Polish | D16–D20 | ID resolution, aliases, completion, version, linter, perf |

---

## 2. Sprint 1: Skeleton & Config System (Days 1–5)

### User Stories

**US-1.1: Project Skeleton**
> As a developer, I want a Go module with proper directory structure so that the project follows Go conventions and is ready for development.

**Acceptance Criteria:**
- `go.mod` exists with correct module path and Go version
- Directory structure matches architecture spec exactly
- `cmd/clockify/main.go` compiles and prints "clockify" with `--help`
- `.gitignore` excludes binaries, IDE files, secrets
- `make build` produces working binary
- Binary runs on current platform without errors

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.1 | `go mod init github.com/kim/clockify-cli` + set Go 1.22+ | 0.5h | `go.mod` | `go mod tidy` succeeds |
| T1.2 | Create directory skeleton (`cmd/`, `internal/`, `pkg/`, `internal/cli/`, `internal/config/`, `internal/api/`, `internal/output/`, `internal/errors/`, `internal/resolve/`) | 0.5h | All dirs | `find . -type d` shows structure |
| T1.3 | Write `cmd/clockify/main.go` with minimal Cobra root command | 2h | `cmd/clockify/main.go`, `internal/cli/root.go` | `go build -o clockify ./cmd/clockify && ./clockify --help` |
| T1.4 | Add global flags to root: `--profile`, `--workspace`, `--output`, `--region`, `--api-key`, `--no-color`, `--verbose`, `--debug`, `--quiet`, `--confirm`, `--page`, `--page-size`, `--no-pagination` | 2h | `internal/cli/root.go` | `./clockify --help` shows all flags |
| T1.5 | Create `.gitignore` (binaries, `.env`, `*.exe`, IDE files) | 0.5h | `.gitignore` | `git status` clean after build |
| T1.6 | Create `Makefile` with `build`, `clean`, `test`, `lint` targets | 1.5h | `Makefile` | `make build` produces `clockify` binary |
| T1.7 | Add `internal/cli/version.go` with `clockify version` subcommand | 1h | `internal/cli/version.go` | `./clockify version` outputs version info |
| T1.8 | Add placeholder files for all `internal/` packages (empty `.go` files with package declarations) | 1h | All `internal/*/` dirs | `go build ./...` compiles |

**US-1.2: Config System Core**
> As a user, I want the CLI to read/write a YAML config file so that I can persist my settings across sessions.

**Acceptance Criteria:**
- Config file created at `~/.clockify/config.yaml` (or `$XDG_CONFIG_HOME/clockify/config.yaml`)
- Config struct supports: `version`, named profiles, per-profile `api_key`, `workspace_id`, `region`, `output`, `date_format`, `time_format`, `week_starts`
- `clockify config list` shows current config
- `clockify config get <key>` returns a single value
- `clockify config set <key> <value>` writes to config file
- `clockify config unset <key>` removes a key
- `clockify config path` shows config file path
- Config precedence: flag > env > profile > default > built-in

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.9 | Define config structs (`Config`, `Profile`) in `internal/config/config.go` | 2h | `internal/config/config.go` | `go build ./internal/config/` |
| T1.10 | Implement `internal/config/paths.go` — platform-specific config/cache dirs | 1.5h | `internal/config/paths.go` | Unit test: correct path on current OS |
| T1.11 | Implement config read/write with `viper` — `Load()`, `Save()`, `SetDefault()` | 3h | `internal/config/config.go` | Write config, read back, values match |
| T1.12 | Implement precedence chain: `Resolve(key)` merges flag → env → profile → default | 2h | `internal/config/config.go` | Test: env overrides config file |
| T1.13 | Implement `internal/config/credentials.go` — keyring integration with env fallback | 2h | `internal/config/credentials.go` | `go build ./internal/config/` |
| T1.14 | Wire `config list`, `config get`, `config set`, `config unset`, `config path` commands | 3h | `internal/cli/config.go` | `clockify config set test_key test_val && clockify config get test_key` → `test_val` |
| T1.15 | Add `CLOCKIFY_*` env var binding to all global flags | 1h | `internal/cli/root.go` | `CLOCKIFY_OUTPUT=json ./clockify config list` uses json |
| T1.16 | Write unit tests for config precedence, read/write, paths | 2h | `internal/config/*_test.go` | `go test ./internal/config/` passes |

**US-1.3: Profile Support**
> As a user, I want named profiles so that I can switch between work/personal accounts easily.

**Acceptance Criteria:**
- `--profile <name>` flag selects which config profile to use
- Profiles stored as top-level keys in `config.yaml`
- Default profile is `"default"`
- Switching profile changes all config values (workspace, region, etc.)
- `clockify config list` shows all profiles

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.17 | Implement profile loading in `internal/config/config.go` — `LoadProfile(name)` | 1.5h | `internal/config/config.go` | `clockify --profile work config list` shows work profile |
| T1.18 | Validate profile exists, show error if not found | 0.5h | `internal/config/config.go` | `clockify --profile nonexistent config list` → error |
| T1.19 | Write tests for profile switching | 1h | `internal/config/*_test.go` | `go test ./internal/config/` passes |

### Sprint 1 Definition of Done

- [ ] `go build ./cmd/clockify` compiles without errors
- [ ] `./clockify --help` shows all global flags
- [ ] `./clockify version` outputs version info
- [ ] `./clockify config set/get/list/unset/path` works end-to-end
- [ ] Config file created/read/written at correct platform path
- [ ] Profile switching works with `--profile`
- [ ] Environment variables override config file values
- [ ] `make build` produces binary
- [ ] All unit tests pass: `go test ./...`
- [ ] No lint errors: `golangci-lint run`

### Daily Checkpoints — Sprint 1

| Day | End-of-Day Verification |
|-----|------------------------|
| D1 | `go build ./cmd/clockify && ./clockify --help` shows global flags |
| D2 | `./clockify config set x y && ./clockify config get x` → `y` |
| D3 | `./clockify config list` shows all profiles; `--profile` flag works |
| D4 | `CLOCKIFY_API_KEY=test ./clockify config list` shows env override |
| D5 | All S1 tests pass; `golangci-lint run` clean; M1 snapshot tagged |

---

## 3. Sprint 2: Output Formatting & Error Handling (Days 6–10)

### User Stories

**US-2.1: Output Formatter Interface**
> As a developer, I want a `Formatter` interface so that all commands can produce consistent output in multiple formats.

**Acceptance Criteria:**
- `output.Formatter` interface defined: `Format(w io.Writer, data interface{}, headers []string) error`
- Auto-detection: TTY → table, piped → JSON
- `--output` flag selects format globally
- All formatters implement the interface

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.1 | Define `Formatter` interface in `internal/output/formatter.go` | 1h | `internal/output/formatter.go` | `go build ./internal/output/` |
| T2.2 | Implement `internal/output/detect.go` — `IsTTY()`, `DetectFormat()` | 1h | `internal/output/detect.go` | Unit test: piped vs TTY detection |
| T2.3 | Implement `GetFormatter(format string, quiet bool) Formatter` factory | 1h | `internal/output/formatter.go` | Returns correct formatter for each format string |
| T2.4 | Wire `--output` flag to formatter selection in root command | 1h | `internal/cli/root.go` | `./clockify --output json config list` → JSON |

**US-2.2: Table Formatter**
> As a user, I want table output that is human-readable with proper alignment and column headers.

**Acceptance Criteria:**
- Table output aligned with auto-width columns
- Headers displayed (unless `--quiet`)
- Empty data shows "No results found"
- Column widths adapt to content

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.5 | Implement `internal/output/table.go` — `TableFormatter` struct | 3h | `internal/output/table.go` | `./clockify config list` → aligned table |
| T2.6 | Add `--quiet` support — suppress headers/footers | 0.5h | `internal/output/table.go` | `./clockify --quiet config list` → no headers |
| T2.7 | Write unit tests for table formatting | 1h | `internal/output/*_test.go` | `go test ./internal/output/` passes |

**US-2.3: JSON, CSV, YAML Formatters**
> As a user, I want machine-readable output formats for scripting and piping.

**Acceptance Criteria:**
- `--output json` produces valid JSON array or object
- `--output csv` produces header row + data rows
- `--output yaml` produces valid YAML
- All formatters handle empty data gracefully

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.8 | Implement `internal/output/json.go` — `JSONFormatter` | 1.5h | `internal/output/json.go` | `./clockify --output json config list \| jq .` works |
| T2.9 | Implement `internal/output/csv.go` — `CSVFormatter` | 1.5h | `internal/output/csv.go` | `./clockify --output csv config list` → valid CSV |
| T2.10 | Implement `internal/output/yaml.go` — `YAMLFormatter` | 1.5h | `internal/output/yaml.go` | `./clockify --output yaml config list` → valid YAML |
| T2.11 | Write unit tests for all formatters | 1.5h | `internal/output/*_test.go` | `go test ./internal/output/` passes |

**US-2.4: Error Handling System**
> As a user, I want structured error messages with categories, exit codes, and suggested fixes.

**Acceptance Criteria:**
- `CLIError` struct: `Code`, `Message`, `Detail`, `Suggestion`, `ExitCode`
- Error categories E001–E010 defined
- `FormatError(err error) string` produces spec-compliant error output
- Exit codes match spec (0–12)
- `--verbose` shows HTTP method + URL
- `--debug` shows full request/response

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.12 | Define `CLIError` struct and category codes in `internal/errors/errors.go` | 1.5h | `internal/errors/errors.go`, `internal/errors/categories.go` | `go build ./internal/errors/` |
| T2.13 | Define exit codes in `internal/errors/exitcode.go` | 1h | `internal/errors/exitcode.go` | Exit codes 0–12 defined as constants |
| T2.14 | Implement `FormatError()` — produces formatted error output to stderr | 2h | `internal/errors/format.go` | Error output matches spec format |
| T2.15 | Wire Cobra `PersistentPreRun` / `PersistentPostRun` for error handling | 1.5h | `internal/cli/root.go` | Errors display correctly with suggestions |
| T2.16 | Implement verbose/debug flag handling (placeholder for HTTP logs) | 1h | `internal/cli/root.go` | `./clockify --verbose` / `--debug` flags parsed |
| T2.17 | Write unit tests for error formatting and exit codes | 1.5h | `internal/errors/*_test.go` | `go test ./internal/errors/` passes |

### Sprint 2 Definition of Done

- [ ] `Formatter` interface works with table, JSON, CSV, YAML
- [ ] Auto-detection: piped → JSON, TTY → table
- [ ] `--output json` produces valid JSON parseable by `jq`
- [ ] `--output table` produces aligned columns
- [ ] `--quiet` suppresses headers
- [ ] `CLIError` struct with all fields defined
- [ ] Error output matches spec format (ERROR, detail, suggestion)
- [ ] Exit codes 0–12 defined and usable
- [ ] All unit tests pass: `go test ./...`
- [ ] No lint errors: `golangci-lint run`

### Daily Checkpoints — Sprint 2

| Day | End-of-Day Verification |
|-----|------------------------|
| D6 | `Formatter` interface defined; auto-detect works |
| D7 | `./clockify config list` → table; `./clockify --output json config list` → JSON |
| D8 | CSV and YAML formatters work; `--quiet` suppresses headers |
| D9 | `CLIError` struct, `FormatError()`, exit codes all defined |
| D10 | All S2 tests pass; `golangci-lint run` clean; M2 snapshot tagged |

---

## 4. Sprint 3: HTTP Client & Configure Wizard (Days 11–15)

### User Stories

**US-3.1: Clockify HTTP Client**
> As a developer, I want an HTTP client that wraps `net/http` with Clockify-specific logic so that all API calls are consistent.

**Acceptance Criteria:**
- `api.ClockifyClient` interface: `Do(ctx, method, path, body, result) error`
- Base URL derived from region config
- Auth header (`X-Api-Key`) attached to all requests
- Request/response logging in verbose/debug modes
- JSON marshaling/unmarshaling for request/response bodies

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.1 | Define `ClockifyClient` interface in `internal/api/client.go` | 1h | `internal/api/client.go` | `go build ./internal/api/` |
| T3.2 | Implement `internal/api/regions.go` — region-to-host mapping | 1h | `internal/api/regions.go` | `GetHost("global")` → `api.clockify.me` |
| T3.3 | Implement base `HTTPClient` struct with `Do()` method | 3h | `internal/api/client.go` | Construct client, call `Do()` with test server |
| T3.4 | Implement `internal/api/middleware.go` — auth, logging, content-type | 2h | `internal/api/middleware.go` | Request has `X-Api-Key` header |
| T3.5 | Add verbose/debug logging middleware | 1h | `internal/api/middleware.go` | `--verbose` shows method+URL; `--debug` shows headers |
| T3.6 | Write unit tests with `httptest` server | 2h | `internal/api/*_test.go` | `go test ./internal/api/` passes |

**US-3.2: Retry & Rate Limiting**
> As a user, I want automatic retries with exponential backoff so that transient failures are handled gracefully.

**Acceptance Criteria:**
- 429 Rate Limit: retry after `Retry-After` header or 2s, up to 5 retries
- 5xx Server Error: retry after 1s, 2s, 4s, 8s, 16s (max 5 retries)
- Network Error: retry after 1s, 2s, 4s (max 3 retries)
- 4xx Client Error: no retry
- Rate limit state tracked per-second counter

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.7 | Implement `internal/api/retry.go` — retry middleware with exponential backoff | 3h | `internal/api/retry.go` | Unit test: retry on 500, no retry on 400 |
| T3.8 | Implement `internal/api/ratelimit.go` — rate limiter (50 req/s) | 2h | `internal/api/ratelimit.go` | Unit test: respects rate limit |
| T3.9 | Wire retry + rate limit into HTTP client middleware chain | 1h | `internal/api/client.go` | Full middleware stack works |
| T3.10 | Write tests for retry scenarios (429, 500, network error) | 2h | `internal/api/*_test.go` | `go test ./internal/api/` passes |

**US-3.3: Configure Wizard**
> As a first-time user, I want an interactive setup wizard so that I can authenticate and configure my workspace in one step.

**Acceptance Criteria:**
- `clockify configure` prompts for: API key, workspace (auto-detected), region, output format
- API key validated against Clockify API before saving
- Credentials stored in OS keychain via `keyring`
- Config saved to `~/.clockify/config.yaml`
- `clockify configure --profile work` configures named profile
- `clockify configure --reset` clears all config

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.11 | Implement `internal/cli/configure.go` — interactive prompts using `golang.org/x/term` | 3h | `internal/cli/configure.go` | `clockify configure` prompts for API key |
| T3.12 | Add API key validation — call `/api/v1/user` with key | 1h | `internal/cli/configure.go` | Invalid key shows error |
| T3.13 | Add workspace auto-detection — call `/api/v1/workspaces` | 1h | `internal/cli/configure.go` | Lists workspaces, selects first |
| T3.14 | Store API key in keyring, config in YAML | 1h | `internal/cli/configure.go` | `clockify status` shows stored key |
| T3.15 | Implement `--profile` and `--reset` flags | 1h | `internal/cli/configure.go` | `clockify configure --reset` clears config |
| T3.16 | Write integration test for configure wizard (mock keyring) | 2h | `internal/cli/*_test.go` | `go test ./internal/cli/` passes |

**US-3.4: Status Command**
> As a user, I want `clockify status` to show my current auth context so that I can verify my configuration.

**Acceptance Criteria:**
- Shows: profile, masked API key, workspace name+ID, region, output format
- `--output json` produces machine-readable status
- Validates API key is still valid on each call

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.17 | Implement `internal/cli/status.go` | 2h | `internal/cli/status.go` | `clockify status` shows all fields |
| T3.18 | Add API key validation in status | 0.5h | `internal/cli/status.go` | Invalid key shows warning |
| T3.19 | Add `--output json` support for status | 0.5h | `internal/cli/status.go` | `clockify status --output json` → valid JSON |

### Sprint 3 Definition of Done

- [ ] `ClockifyClient.Do()` makes authenticated HTTP requests
- [ ] Region mapping works for all 5 regions
- [ ] Retry works on 429/5xx with exponential backoff
- [ ] Rate limiter respects 50 req/s limit
- [ ] `clockify configure` runs interactive wizard end-to-end
- [ ] API key stored in OS keychain (or env var fallback)
- [ ] `clockify status` shows current context
- [ ] Verbose/debug logging shows HTTP details
- [ ] All unit tests pass: `go test ./...`
- [ ] No lint errors: `golangci-lint run`

### Daily Checkpoints — Sprint 3

| Day | End-of-Day Verification |
|-----|------------------------|
| D11 | `ClockifyClient` interface defined; `Do()` works with `httptest` |
| D12 | Retry middleware works; rate limiter implemented |
| D13 | `clockify configure` prompts and stores key |
| D14 | `clockify configure` validates key against API; workspace auto-detection works |
| D15 | `clockify status` works; all S3 tests pass; M3 snapshot tagged |

---

## 5. Sprint 4: Integration & Polish (Days 16–20)

### User Stories

**US-4.1: ID Resolution**
> As a user, I want to reference resources by ID, prefix, or name so that I don't have to memorize UUIDs.

**Acceptance Criteria:**
- `resolve.Resolver` interface: `Resolve(ctx, client, workspaceID, input) (string, error)`
- Exact UUID match returns immediately
- Prefix match (unique prefix of UUID) resolves correctly
- Fuzzy name match uses `sahilm/fuzzy`
- Multiple fuzzy matches → error with suggestions
- `--id` and `--name` flags on relevant commands

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T4.1 | Define `Resolver` interface in `internal/resolve/resolver.go` | 0.5h | `internal/resolve/resolver.go` | `go build ./internal/resolve/` |
| T4.2 | Implement `internal/resolve/exact.go` — exact UUID match | 0.5h | `internal/resolve/exact.go` | Unit test: valid UUID passes |
| T4.3 | Implement `internal/resolve/prefix.go` — unique prefix match | 1h | `internal/resolve/prefix.go` | Unit test: `5f8a` matches `5f8a1b2c...` |
| T4.4 | Implement `internal/resolve/fuzzy.go` — name fuzzy match with `sahilm/fuzzy` | 2h | `internal/resolve/fuzzy.go` | Unit test: "websit" matches "Website Redesign" |
| T4.5 | Wire resolvers into `ResolveInput()` composite function | 1h | `internal/resolve/resolver.go` | Exact → prefix → fuzzy fallback chain |
| T4.6 | Write unit tests for all resolver types | 1.5h | `internal/resolve/*_test.go` | `go test ./internal/resolve/` passes |

**US-4.2: Aliases**
> As a user, I want short aliases for common commands so that I can work faster.

**Acceptance Criteria:**
- `clockify lw` = `clockify workspace list`
- `clockify le` = `clockify time-entry list`
- `clockify lp` = `clockify project list`
- `clockify ge` = `clockify get time-entry`
- `clockify ce` = `clockify create time-entry`
- `clockify cfg` = `clockify config set`
- Aliases shown in `clockify --help`

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T4.7 | Implement `internal/cli/aliases.go` — alias map and Cobra registration | 2h | `internal/cli/aliases.go` | `clockify lw` runs `workspace list` |
| T4.8 | Add alias documentation to help text | 0.5h | `internal/cli/aliases.go` | `clockify --help` shows aliases |
| T4.9 | Write tests for alias resolution | 0.5h | `internal/cli/*_test.go` | `go test ./internal/cli/` passes |

**US-4.3: Shell Completion**
> As a user, I want shell completion scripts so that tab-completion works in my shell.

**Acceptance Criteria:**
- `clockify completion bash` outputs bash completion script
- `clockify completion zsh` outputs zsh completion script
- `clockify completion fish` outputs fish completion script
- `clockify completion powershell` outputs powershell completion script
- `clockify completion install --shell <shell>` installs to correct location

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T4.10 | Implement `internal/cli/completion.go` using Cobra's built-in completion | 2h | `internal/cli/completion.go` | `clockify completion bash` → valid script |
| T4.11 | Add `install` subcommand with shell detection | 1.5h | `internal/cli/completion.go` | `clockify completion install --shell bash` → installs |
| T4.12 | Test completion in bash/zsh (manual) | 0.5h | — | Tab completion works |

**US-4.4: Build & Quality**
> As a developer, I want a robust build system and linting so that code quality is maintained.

**Acceptance Criteria:**
- `make build` produces binary for current platform
- `make build-all` cross-compiles for all target platforms
- `make lint` runs `golangci-lint` with project config
- `make test` runs all tests
- `make check` = lint + test (for CI)
- `.golangci.yml` configured with project-appropriate rules
- Binary starts in <50ms

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T4.13 | Create `.golangci.yml` with linter config | 1h | `.golangci.yml` | `golangci-lint run` passes |
| T4.14 | Add `build-all` target to Makefile (cross-compile 5 platforms) | 1h | `Makefile` | `make build-all` produces binaries |
| T4.15 | Add `check` target to Makefile (lint + test) | 0.5h | `Makefile` | `make check` passes |
| T4.16 | Performance check: measure binary startup time | 1h | — | `time ./clockify --help` < 50ms |
| T4.17 | Run full test suite and fix any failures | 2h | Various | `go test ./...` all pass |

**US-4.5: End-to-End Verification**
> As a team lead, I want all success criteria verified so that M0 is complete.

**Acceptance Criteria:**
- [ ] `clockify --help` shows all global flags and available commands
- [ ] `clockify config set/get/list` works end-to-end
- [ ] `clockify configure` wizard runs interactively and stores credentials in keychain
- [ ] HTTP client successfully authenticates with Clockify API
- [ ] `--output json` produces valid JSON; `--output table` produces aligned table output
- [ ] Binary starts in <50ms

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T4.18 | Run full E2E verification against success criteria checklist | 2h | — | All 6 criteria pass |
| T4.19 | Fix any issues found during E2E verification | 2h | Various | Re-verify all criteria |
| T4.20 | Create M0 completion commit and tag | 0.5h | — | `git tag v0.1.0-m0` |

### Sprint 4 Definition of Done

- [ ] ID resolution: exact, prefix, fuzzy all work
- [ ] Aliases work for all specified short commands
- [ ] Shell completion scripts generated for bash/zsh/fish/powershell
- [ ] `make build-all` cross-compiles for Linux/macOS/Windows
- [ ] `.golangci.yml` configured; `make lint` passes
- [ ] Binary starts in <50ms (measured)
- [ ] All 6 M0 success criteria verified
- [ ] All unit tests pass: `go test ./...`
- [ ] M0 completion tag created

### Daily Checkpoints — Sprint 4

| Day | End-of-Day Verification |
|-----|------------------------|
| D16 | Fuzzy resolver works; prefix resolver works |
| D17 | Aliases work; completion scripts generate |
| D18 | `make build-all` cross-compiles; `.golangci.yml` passes |
| D19 | Binary startup <50ms; all tests pass |
| D20 | All 6 M0 success criteria verified; M0 tag created |

---

## 6. Task Dependency Map

```
S1 (Skeleton & Config)
├── T1.1 (go.mod) ─────────────────────────────────────┐
├── T1.2 (dirs) ──► T1.3 (root cmd) ──► T1.4 (flags) │
│                   ├── T1.7 (version)                  │
│                   └── T1.8 (placeholders)             │
├── T1.5 (.gitignore)                                   │
├── T1.6 (Makefile)                                     │
├── T1.9 (config structs) ──► T1.10 (paths)            │
│   └──► T1.11 (viper read/write) ──► T1.12 (precedence)
│       └──► T1.14 (config commands) ──► T1.15 (env vars)
├── T1.13 (credentials/keyring)                         │
└── T1.17 (profiles) ──► T1.18 (profile validation)    │
    └──► T1.19 (profile tests)                          │
                                                       │
S2 (Output & Errors) ◄─────────────────────────────────┘
├── T2.1 (Formatter interface) ──► T2.2 (detect) ──► T2.3 (factory) ──► T2.4 (wire)
├── T2.5 (table) ──► T2.6 (quiet)
├── T2.8 (JSON) ──┐
├── T2.9 (CSV) ───┤──► T2.11 (formatter tests)
├── T2.10 (YAML) ─┘
└── T2.12 (CLIError) ──► T2.13 (exit codes) ──► T2.14 (format) ──► T2.15 (wire)
    └──► T2.16 (verbose/debug)

S3 (HTTP Client & Configure) ◄─────────────────────────┘
├── T3.1 (interface) ──► T3.2 (regions) ──► T3.3 (Do) ──► T3.4 (middleware)
│   └──► T3.5 (logging)
├── T3.7 (retry) ──► T3.9 (wire)
├── T3.8 (rate limit) ──► T3.9 (wire)
└── T3.11 (configure wizard) ──► T3.12 (validate) ──► T3.13 (workspace detect)
    └──► T3.14 (store) ──► T3.15 (flags)
    └──► T3.17 (status) ──► T3.18 (validate) ──► T3.19 (json)

S4 (Integration & Polish) ◄────────────────────────────┘
├── T4.1 (interface) ──► T4.2 (exact) ──► T4.3 (prefix) ──► T4.4 (fuzzy) ──► T4.5 (composite)
├── T4.7 (aliases) ──► T4.8 (docs)
├── T4.10 (completion) ──► T4.11 (install)
├── T4.13 (.golangci.yml) ──► T4.14 (build-all) ──► T4.15 (check)
└── T4.16 (perf check) ──► T4.17 (fix failures) ──► T4.18 (E2E) ──► T4.20 (tag)
```

### Critical Path

```
T1.1 → T1.3 → T1.4 → T1.9 → T1.11 → T1.12 → T1.14 → T1.15
  → T2.1 → T2.3 → T2.4
  → T2.12 → T2.14 → T2.15
  → T3.1 → T3.3 → T3.4
  → T3.11 → T3.12 → T3.14
  → T4.1 → T4.4 → T4.5
  → T4.18 → T4.20
```

**Estimated critical path duration:** ~45 hours (9 days of focused work)
**Buffer for non-critical tasks:** ~25 hours (5 days)
**Total estimated:** ~70 hours (14 days) = 3 weeks + 2 days → fits in 4-week sprint

---

## 7. Risk Mitigation

| Risk | Mitigation | Owner | Sprint |
|------|------------|-------|--------|
| **Cobra alias limitations** | Spike on Day 1: test `cobra.Command.AddCommand` with aliases; fall back to wrapper commands if needed | Dev | S1 |
| **Keyring unavailable (headless)** | Env var fallback built into `credentials.go`; test with `CLOCKIFY_API_KEY` in CI | Dev | S1 |
| **Clockify API rate limits unknown** | Test empirically in S3 with 10 rapid requests; document discovered limits; implement 50 req/s default | Dev | S3 |
| **`tablewriter` not flexible enough** | Evaluate on S2 D7; fallback to `text/tabwriter` from stdlib if needed | Dev | S2 |
| **Duration parsing complexity** | Defer to M3 (not in M0 scope); M0 only needs config and output infrastructure | Dev | S4 |
| **Binary startup >50ms** | Profile with `go tool pprof`; strip debug info with `ldflags -s -w`; remove unnecessary deps | Dev | S4 |
| **Cross-compilation issues** | Test on D18; use `GOOS`/`GOARCH` env vars; avoid CGO dependencies | Dev | S4 |
| **Scope creep (config features)** | M0 config: only YAML read/write, profiles, precedence. No cache, no encryption, no validation beyond types | Dev | S1 |

---

## 8. Trade-offs: Where to Cut Corners vs. Where to Hold Firm

### Hold Firm (Non-Negotiable)

| Area | Why |
|------|-----|
| **Config precedence chain** | Foundation for all flag/env/config behavior; getting this wrong cascades |
| **Formatter interface** | Every command depends on it; interface changes later are expensive |
| **Error struct with exit codes** | CI/CD depends on consistent exit codes; must be correct from Day 1 |
| **Keyring integration** | Security: API keys must not be stored in plaintext by default |
| **Retry with exponential backoff** | API stability depends on this; must be correct before any real API calls |

### Cut Corners (If Behind Schedule)

| Area | How to Cut | Impact |
|------|------------|--------|
| **Profile support** | Simplify to 1 profile (default) only; defer multi-profile to M1 | Users can still use env vars to switch |
| **Shell completion** | Output scripts only; skip `install` subcommand | Users can manually install scripts |
| **CSV/YAML formatters** | Implement CSV only; defer YAML to M1 | JSON + table + CSV covers 90% of use cases |
| **Fuzzy resolution** | Implement exact + prefix only; defer fuzzy to M1 | Users can use full UUIDs or prefix |
| **Verbose/debug modes** | Implement verbose only; defer debug to M1 | Verbose covers most debugging needs |
| **Region mapping** | Implement `global` only; defer other regions to M1 | Most users are on global |
| **`clockify status` API validation** | Show stored config only; skip live API validation | Users know their key works from `configure` |
| **Unit test coverage** | Target 60% instead of 80%; focus on critical paths only | Config + output + error tests mandatory |

### Cut Corners — Emergency Only

| Area | How to Cut | Impact |
|------|------------|--------|
| **Makefile cross-compilation** | Single-platform build only | Users build from source |
| **Linting** | Skip `.golangci.yml`; run `go vet` only | Lower code quality bar temporarily |
| **Performance check** | Skip startup time measurement | Risk: may not meet <50ms goal |

---

## 9. File-Level Task Summary

| File | Sprint | Tasks | Est. Hours |
|------|--------|-------|------------|
| `go.mod` | S1 | T1.1 | 0.5h |
| `cmd/clockify/main.go` | S1 | T1.3 | 2h |
| `internal/cli/root.go` | S1 | T1.3, T1.4, T1.15, T2.15, T2.16 | 8h |
| `internal/cli/version.go` | S1 | T1.7 | 1h |
| `internal/cli/config.go` | S1 | T1.14 | 3h |
| `internal/cli/configure.go` | S3 | T3.11–T3.15 | 7h |
| `internal/cli/status.go` | S3 | T3.17–T3.19 | 3h |
| `internal/cli/completion.go` | S4 | T4.10–T4.11 | 3.5h |
| `internal/cli/aliases.go` | S4 | T4.7–T4.8 | 2.5h |
| `internal/config/config.go` | S1 | T1.9, T1.11, T1.12, T1.17, T1.18 | 10h |
| `internal/config/paths.go` | S1 | T1.10 | 1.5h |
| `internal/config/credentials.go` | S1 | T1.13 | 2h |
| `internal/output/formatter.go` | S2 | T2.1, T2.3 | 2h |
| `internal/output/detect.go` | S2 | T2.2 | 1h |
| `internal/output/table.go` | S2 | T2.5, T2.6 | 3.5h |
| `internal/output/json.go` | S2 | T2.8 | 1.5h |
| `internal/output/csv.go` | S2 | T2.9 | 1.5h |
| `internal/output/yaml.go` | S2 | T2.10 | 1.5h |
| `internal/errors/errors.go` | S2 | T2.12 | 1.5h |
| `internal/errors/categories.go` | S2 | T2.12 | (included above) |
| `internal/errors/exitcode.go` | S2 | T2.13 | 1h |
| `internal/errors/format.go` | S2 | T2.14 | 2h |
| `internal/api/client.go` | S3 | T3.1, T3.3, T3.9 | 5h |
| `internal/api/regions.go` | S3 | T3.2 | 1h |
| `internal/api/middleware.go` | S3 | T3.4, T3.5 | 3h |
| `internal/api/retry.go` | S3 | T3.7 | 3h |
| `internal/api/ratelimit.go` | S3 | T3.8 | 2h |
| `internal/resolve/resolver.go` | S4 | T4.1, T4.5 | 1.5h |
| `internal/resolve/exact.go` | S4 | T4.2 | 0.5h |
| `internal/resolve/prefix.go` | S4 | T4.3 | 1h |
| `internal/resolve/fuzzy.go` | S4 | T4.4 | 2h |
| `pkg/duration/duration.go` | — | Deferred to M3 | 0h |
| `Makefile` | S1 | T1.6 | 1.5h |
| `.golangci.yml` | S4 | T4.13 | 1h |
| `.gitignore` | S1 | T1.5 | 0.5h |

---

## 10. M0 Overall Definition of Done

| Criterion | Verification | Status |
|-----------|-------------|--------|
| `clockify --help` shows all global flags and commands | Run command, inspect output | ☐ |
| `clockify config set/get/list` works end-to-end | `clockify config set x y && clockify config get x` | ☐ |
| `clockify configure` stores credentials in keychain | Run wizard, verify keychain entry | ☐ |
| HTTP client authenticates with Clockify API | Call `/api/v1/user` with valid key | ☐ |
| `--output json` produces valid JSON | `clockify config list --output json \| jq .` | ☐ |
| `--output table` produces aligned table | `clockify config list` → visual inspection | ☐ |
| Binary starts in <50ms | `time ./clockify --help` | ☐ |
| All unit tests pass | `go test ./...` | ☐ |
| No lint errors | `golangci-lint run` | ☐ |
| Cross-platform build works | `make build-all` produces binaries | ☐ |

---

## 11. Daily Standup Template

Use this at end of each day:

```markdown
## Day X Standup — Sprint Y

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

*This plan is a living document. Update task estimates and status as work progresses.*
