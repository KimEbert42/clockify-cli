# M1: Authentication & Workspace Context — Sprint Plan

**Milestone:** M1
**Sprint Duration:** 2 × 1-week sprints (5 working days each)
**Developer:** 1 (solo)
**Start Date:** TBD
**End Date:** TBD (2 weeks from start)

---

## 1. Sprint Structure

| Sprint | Theme | Days | Goal |
|--------|-------|------|------|
| **S1** | Keyring & Configure Fixes | D1–D5 | Real keyring integration, API validation, workspace detection fix |
| **S2** | Multi-Profile & Polish | D6–D10 | Profile switching, clear-cache, status validation, E2E verification |

---

## 2. Current State Assessment

### Already Implemented (from M0)
| Feature | Status | Issues |
|---------|--------|--------|
| `clockify configure` wizard | ✅ Works | API key not validated, workspace selection bug |
| `clockify status` | ✅ Works | No live API validation |
| `clockify config` (list/get/set/unset/path) | ✅ Works | Missing `clear-cache` |
| `--profile` flag | ✅ Works | Basic profile loading implemented |
| `CLOCKIFY_*` env vars | ✅ Works | Viper binding in place |
| Region mapping | ✅ Works | All 5 regions defined |

### Bugs to Fix
| Bug | Location | Impact |
|-----|----------|--------|
| Keyring is a stub | `internal/config/credentials.go` | API keys not persisted across sessions |
| Workspace selection always picks first | `internal/cli/configure.go:72` | User choice ignored |
| No API key validation in configure | `internal/cli/configure.go` | Invalid keys accepted |
| No `config clear-cache` | `internal/cli/config.go` | Cache cannot be cleared |

---

## 3. Sprint 1: Keyring & Configure Fixes (Days 1–5)

### User Stories

**US-1.1: Real Keyring Integration**
> As a user, I want my API key stored securely in my OS keychain so that it persists across sessions and is not stored in plaintext.

**Acceptance Criteria:**
- `credentials.go` uses `zalando/go-keyring` for Linux (Secret Service), macOS (Keychain), Windows (Credential Manager)
- `GetAPIKey()` reads from keyring, falls back to env var `CLOCKIFY_API_KEY`
- `SetAPIKey()` writes to keyring, supports per-profile keys
- `DeleteAPIKey()` removes from keyring
- `clockify configure` stores API key in keyring
- `clockify status` shows masked key from keyring

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.1 | Add `github.com/zalando/go-keyring` dependency | 0.5h | `go.mod` | `go mod tidy` succeeds |
| T1.2 | Implement `getKeyring()` with SecretService backend | 2h | `internal/config/credentials.go` | `go build ./internal/config/` |
| T1.3 | Implement `setKeyring()` with SecretService backend | 1.5h | `internal/config/credentials.go` | `go build ./internal/config/` |
| T1.4 | Implement `deleteKeyring()` with SecretService backend | 1h | `internal/config/credentials.go` | `go build ./internal/config/` |
| T1.5 | Add env var fallback: if keyring fails, use `CLOCKIFY_API_KEY` | 1h | `internal/config/credentials.go` | Works without keyring installed |
| T1.6 | Write unit tests for keyring operations (mock keyring) | 2h | `internal/config/credentials_test.go` | `go test ./internal/config/` passes |
| T1.7 | Test on Linux with `gnome-keyring` or `seahorse` | 1h | — | Manual verification |

**US-1.2: API Key Validation in Configure**
> As a user, I want `clockify configure` to validate my API key against the Clockify API before saving so that I know my credentials are correct.

**Acceptance Criteria:**
- After API key prompt, call `GET /api/v1/user` with the key
- Valid key: show user name/email, proceed to workspace selection
- Invalid key (401/403): show error, re-prompt or exit
- Network error: warn but allow save (offline mode)

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.8 | Add `ValidateAPIKey()` function using `api.Client.Do()` | 1.5h | `internal/config/credentials.go` | Returns user info or error |
| T1.9 | Wire validation into `configure.go` after API key prompt | 2h | `internal/cli/configure.go` | Invalid key shows error |
| T1.10 | Add `--offline` flag to skip validation | 0.5h | `internal/cli/configure.go` | `--offline` skips API call |
| T1.11 | Write unit tests for validation (httptest mock) | 1.5h | `internal/cli/configure_test.go` | `go test ./internal/cli/` passes |

**US-1.3: Workspace Selection Fix**
> As a user, I want `clockify configure` to respect my workspace selection so that I can choose the correct workspace.

**Acceptance Criteria:**
- Workspace list displays with numbers (1, 2, 3...)
- User input is parsed correctly (integer selection)
- Selected workspace ID is stored in config
- Invalid input shows error and re-prompts

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.12 | Fix workspace selection parsing in `configure.go` | 1.5h | `internal/cli/configure.go` | Selected workspace is stored |
| T1.13 | Add input validation for workspace selection | 1h | `internal/cli/configure.go` | Invalid input shows error |
| T1.14 | Test workspace selection with multiple workspaces | 1h | — | Manual verification |
| T1.15 | Write unit tests for workspace selection | 1h | `internal/cli/configure_test.go` | `go test ./internal/cli/` passes |

### Sprint 1 Definition of Done

- [ ] API key stored in OS keyring (not plaintext)
- [ ] Keyring fallback to env var works
- [ ] `clockify configure` validates API key against API
- [ ] Workspace selection bug fixed
- [ ] `clockify status` shows masked key from keyring
- [ ] All unit tests pass: `go test ./...`
- [ ] No lint errors: `golangci-lint run`

### Daily Checkpoints — Sprint 1

| Day | End-of-Day Verification |
|-----|------------------------|
| D1 | `go-keyring` dependency added; `getKeyring()` compiles |
| D2 | `setKeyring()` and `deleteKeyring()` work |
| D3 | `ValidateAPIKey()` calls API; `configure` validates key |
| D4 | Workspace selection bug fixed; input validation works |
| D5 | All S1 tests pass; `golangci-lint run` clean |

---

## 4. Sprint 2: Multi-Profile & Polish (Days 6–10)

### User Stories

**US-2.1: Multi-Profile Support**
> As a user, I want to create and switch between named profiles (work, personal) so that I can manage multiple Clockify accounts.

**Acceptance Criteria:**
- `clockify configure --profile work` creates/updates named profile
- `clockify --profile work status` shows work profile context
- `clockify --profile work project list` uses work profile's workspace
- `clockify config list` shows all profiles
- `clockify config get --profile work workspace_id` shows work workspace
- Default profile is `"default"`

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.1 | Verify `--profile` flag loads correct profile | 1h | `internal/config/config.go` | `--profile work` loads work config |
| T2.2 | Add `--profile` flag to `configure` command | 1h | `internal/cli/configure.go` | `configure --profile work` creates work profile |
| T2.3 | Ensure `config list` shows all profiles | 1h | `internal/cli/config.go` | `config list` shows default + work |
| T2.4 | Add `config delete --profile <name>` command | 1.5h | `internal/cli/config.go` | `config delete --profile work` removes work profile |
| T2.5 | Write tests for multi-profile operations | 2h | `internal/config/config_test.go` | `go test ./internal/config/` passes |

**US-2.2: Config Clear-Cache**
> As a user, I want `clockify config clear-cache` so that I can clear any cached data and force fresh API calls.

**Acceptance Criteria:**
- `clockify config clear-cache` removes cached workspace/project/tag data
- `clockify config clear-cache --profile work` clears work profile cache
- Shows confirmation of what was cleared
- Cache directory determined by platform (XDG on Linux)

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.6 | Define cache directory path in `internal/config/paths.go` | 1h | `internal/config/paths.go` | Returns correct cache path |
| T2.7 | Implement `ClearCache()` function | 1.5h | `internal/config/cache.go` | Deletes cache directory contents |
| T2.8 | Wire `clear-cache` subcommand | 1h | `internal/cli/config.go` | `config clear-cache` works |
| T2.9 | Add `--profile` support to clear-cache | 0.5h | `internal/cli/config.go` | Clears profile-specific cache |
| T2.10 | Write unit tests for cache operations | 1h | `internal/config/cache_test.go` | `go test ./internal/config/` passes |

**US-2.3: Status Command Enhancement**
> As a user, I want `clockify status` to validate my API key is still valid so that I know my credentials work.

**Acceptance Criteria:**
- `clockify status` calls `GET /api/v1/user` to validate key
- Valid key: shows green checkmark or "Valid" status
- Invalid key: shows red warning with suggestion to re-run `configure`
- `--output json` includes `valid: true/false` field
- Network error: shows "unable to validate" warning

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.11 | Add API validation to `status` command | 1.5h | `internal/cli/status.go` | `status` validates key |
| T2.12 | Add visual indicator (Valid/Invalid) | 0.5h | `internal/cli/status.go` | Shows checkmark or warning |
| T2.13 | Add `valid` field to JSON output | 0.5h | `internal/cli/status.go` | `--output json` includes valid |
| T2.14 | Handle network error gracefully | 0.5h | `internal/cli/status.go` | Network error shows warning |
| T2.15 | Write tests for status validation | 1h | `internal/cli/status_test.go` | `go test ./internal/cli/` passes |

**US-2.4: End-to-End Verification**
> As a team lead, I want all M1 success criteria verified so that the milestone is complete.

**Acceptance Criteria:**
- [ ] `clockify configure` stores API key in OS keychain
- [ ] `clockify status` shows current auth context with masked key
- [ ] `clockify --profile work list projects` switches context correctly
- [ ] `CLOCKIFY_API_KEY=... clockify status` works without config file
- [ ] Config precedence chain verified: flag > env > config > default

**Tasks:**
| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.16 | Run full E2E verification against success criteria | 2h | — | All 5 criteria pass |
| T2.17 | Fix any issues found during E2E verification | 1.5h | Various | Re-verify all criteria |
| T2.18 | Create M1 completion commit and tag | 0.5h | — | `git tag v0.2.0-m1` |

### Sprint 2 Definition of Done

- [ ] Multi-profile support works end-to-end
- [ ] `config clear-cache` clears cached data
- [ ] `clockify status` validates API key against API
- [ ] All 5 M1 success criteria verified
- [ ] All unit tests pass: `go test ./...`
- [ ] No lint errors: `golangci-lint run`
- [ ] M1 completion tag created

### Daily Checkpoints — Sprint 2

| Day | End-of-Day Verification |
|-----|------------------------|
| D6 | `--profile work` loads work config; `config list` shows profiles |
| D7 | `config clear-cache` works; cache directory logic correct |
| D8 | `clockify status` validates key; shows valid/invalid indicator |
| D9 | All E2E criteria pass; issues fixed |
| D10 | All S2 tests pass; M1 tag created |

---

## 5. Task Dependency Map

```
S1 (Keyring & Configure Fixes)
├── T1.1 (go-keyring dep) ──► T1.2 (getKeyring) ──► T1.3 (setKeyring) ──► T1.4 (deleteKeyring)
│   └──► T1.5 (env fallback) ──► T1.6 (keyring tests) ──► T1.7 (manual test)
├── T1.8 (ValidateAPIKey) ──► T1.9 (wire into configure) ──► T1.10 (offline flag)
│   └──► T1.11 (validation tests)
├── T1.12 (fix workspace selection) ──► T1.13 (input validation) ──► T1.14 (manual test)
│   └──► T1.15 (workspace tests)

S2 (Multi-Profile & Polish)
├── T2.1 (verify profile loading) ──► T2.2 (profile flag on configure) ──► T2.3 (config list profiles)
│   └──► T2.4 (config delete profile) ──► T2.5 (profile tests)
├── T2.6 (cache path) ──► T2.7 (ClearCache) ──► T2.8 (clear-cache cmd)
│   └──► T2.9 (profile support) ──► T2.10 (cache tests)
├── T2.11 (status validation) ──► T2.12 (visual indicator) ──► T2.13 (json output)
│   └──► T2.14 (network error) ──► T2.15 (status tests)
└── T2.16 (E2E verification) ──► T2.17 (fix issues) ──► T2.18 (tag)
```

### Critical Path

```
T1.1 → T1.2 → T1.3 → T1.5 → T1.8 → T1.9 → T1.12 → T1.13
  → T2.1 → T2.2 → T2.4
  → T2.6 → T2.7 → T2.8
  → T2.11 → T2.12
  → T2.16 → T2.18
```

**Estimated critical path duration:** ~30 hours (6 days of focused work)
**Buffer for non-critical tasks:** ~10 hours (2 days)
**Total estimated:** ~40 hours (8 days) = 1.5 weeks + buffer → fits in 2-week sprint

---

## 6. Risk Mitigation

| Risk | Mitigation | Owner | Sprint |
|------|------------|-------|--------|
| **Keyring unavailable (headless/CI)** | Env var fallback; test with `CLOCKIFY_API_KEY` in CI | Dev | S1 |
| **`go-keyring` Linux Secret Service unavailable** | Document requirement for `gnome-keyring` or `seahorse`; provide env var fallback | Dev | S1 |
| **Clockify API rate limits during validation** | Add 1s debounce between validation calls; implement retry | Dev | S1 |
| **Profile switching edge cases** | Test with 5+ profiles; validate config merge logic | Dev | S2 |
| **Cache directory permissions** | Use `os.MkdirAll` with 0700; handle permission errors gracefully | Dev | S2 |

---

## 7. Trade-offs: Where to Cut Corners vs. Where to Hold Firm

### Hold Firm (Non-Negotiable)

| Area | Why |
|------|-----|
| **Real keyring integration** | Security: API keys must not be stored in plaintext by default |
| **API key validation** | User experience: invalid keys should fail fast with clear errors |
| **Workspace selection fix** | Core UX: users must be able to select their workspace |
| **Config precedence chain** | Foundation for all flag/env/config behavior |

### Cut Corners (If Behind Schedule)

| Area | How to Cut | Impact |
|------|------------|--------|
| **Multi-profile support** | Defer to M2; only support default profile | Users can use env vars to switch |
| **`config clear-cache`** | Defer to M2; users can manually delete cache dir | Minor inconvenience |
| **Status validation** | Show stored config only; skip live API validation | Users know their key works from `configure` |
| **Offline mode flag** | Skip `--offline` flag; always validate | Users must be online during configure |
| **Unit test coverage** | Target 60% instead of 80%; focus on keyring + configure | Critical paths mandatory |

### Cut Corners — Emergency Only

| Area | How to Cut | Impact |
|------|------------|--------|
| **Visual indicators in status** | Plain text only; skip checkmark/warning | Less polished UX |
| **Config delete profile** | Manual deletion only (edit config.yaml) | Minor inconvenience |

---

## 8. File-Level Task Summary

| File | Sprint | Tasks | Est. Hours |
|------|--------|-------|------------|
| `go.mod` | S1 | T1.1 | 0.5h |
| `internal/config/credentials.go` | S1 | T1.2–T1.5, T1.8 | 7.5h |
| `internal/config/credentials_test.go` | S1 | T1.6, T1.11 | 3.5h |
| `internal/cli/configure.go` | S1 | T1.9–T1.13 | 5.5h |
| `internal/cli/configure_test.go` | S1 | T1.11, T1.15 | 2.5h |
| `internal/config/config.go` | S2 | T2.1 | 1h |
| `internal/config/config_test.go` | S2 | T2.5 | 2h |
| `internal/cli/config.go` | S2 | T2.4, T2.8, T2.9 | 3h |
| `internal/config/paths.go` | S2 | T2.6 | 1h |
| `internal/config/cache.go` | S2 | T2.7 | 1.5h |
| `internal/config/cache_test.go` | S2 | T2.10 | 1h |
| `internal/cli/status.go` | S2 | T2.11–T2.14 | 3h |
| `internal/cli/status_test.go` | S2 | T2.15 | 1h |
| `internal/cli/config.go` | S2 | T2.3 | (included above) |

---

## 9. M1 Overall Definition of Done

| Criterion | Verification | Status |
|-----------|-------------|--------|
| `clockify configure` stores API key in OS keychain | Run wizard, check keyring | ☐ |
| `clockify status` shows current auth context with masked key | `clockify status` → inspect output | ☐ |
| `clockify --profile work` switches context correctly | `clockify --profile work status` → shows work config | ☐ |
| `CLOCKIFY_API_KEY=... clockify status` works without config | Set env var, run status | ☐ |
| Config precedence chain verified: flag > env > config > default | Test each layer | ☐ |
| All unit tests pass | `go test ./...` | ☐ |
| No lint errors | `golangci-lint run` | ☐ |

---

## 10. Daily Standup Template

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
