# Commit Review — 0aac85a

- **Commit:** 0aac85a5fb94a90c1a86715df8bd35161978604b
- **Subject:** docs: update PLAN.md with M0 completion status
- **Date:** 2026-09-21
- **Experts engaged:** Tech Team Lead (light, in-session)
- **Skill mode:** Auto-Assignment (kims-team)
- **Review date:** 2026-10-05
- **Related plan:** PLAN.md §5.7 Review Backlog #4 (P3)

## Diff summary

Docs-only commit: `PLAN.md` +41/−32. Records M0 as complete:

- Header status → `M0 Complete — Active Development`
- M0 section gains `✅ COMPLETE — Tagged v0.1.0-m0`, completion date, deliverable table with Status/Notes columns (all ✅), success-criteria checklists all `[x]`, Actual Effort `~1 day`
- Appendix timeline gains a Status column (`M0 DONE (1 day)`, M1 `⏳ Next`)
- Footer note: M0 actual (~1 day, 34 files, 2612 lines of Go, all success criteria met)

No product code changed. Review is of the **status claims vs the tree at this SHA / the M0 feat commit `6342b33`** and PLAN M0 criteria.

## Findings

### Critical

1. **M0 marked COMPLETE with binary success criteria checked, but no entrypoint exists.**
   - Location: PLAN.md M0 success criteria + deliverable table (tree @ `0aac85a`)
   - Issue: Tree has `internal/` + `pkg/` but **no `cmd/` and no `package main` anywhere**. `Makefile` targets `./cmd/clockify`. Criteria checked `[x]` include `clockify --help`, "Binary starts in <50ms (measured: 5ms)", "Binary size <20MB (measured: 12MB)".
   - Why it matters: A binary cannot be built or measured at this SHA. The "measured" figures are not attributable to this code state; the milestone's core DoD (buildable CLI) is unmet. Same gap was flagged Critical in `docs/reviews/6342b33-m0-scaffolding.md`; `main.go` only lands later (M4 `f0d6787`). Status doc presents unmet criteria as done.

2. **OS keychain deliverable marked ✅ on an env-var-only stub; success criterion quietly weakened.**
   - Location: PLAN.md M0 deliverable "OS keychain integration | ✅ | Env var fallback implemented"; success criterion "configure wizard … stores credentials" (dropped "in keychain")
   - Issue: At this SHA, `internal/config/credentials.go` `getKeyring`/`setKeyring`/`deleteKeyring` are `os.Getenv`/`os.Setenv`/`os.Unsetenv` only ("keyring not available" on read miss). No keyring library in tree. M0 scope text still says "`keyring` library for secure credential storage".
   - Why it matters: Credentials do not persist across processes; marking the deliverable complete overstates security/auth DoD. Real keyring (`zalando/go-keyring`) landed in M1 `7214ea1`. Softening the criterion to hide the gap undermines PLAN as an audit trail.

### Major

3. **"HTTP client successfully authenticates with Clockify API" checked `[x]` without validation code.**
   - Location: PLAN.md M0 success criteria
   - Issue: No live API-key validation path exists at this SHA. Auth validation lands in M1 (`7214ea1` — `GET /user` before store).
   - Why it matters: Criterion claims an end-to-end behavior the milestone did not deliver.

4. **"`clockify config set/get/list` works end-to-end" checked `[x]` without a runnable CLI.**
   - Location: PLAN.md M0 success criteria
   - Issue: No entrypoint (see #1). Additionally, `6342b33` review found `config set/unset` wipes the `profiles:` map and can persist env-derived secrets — "works end-to-end" is doubtful even hypothetically.
   - Why it matters: Checklist status does not reflect code reality; false green signals for downstream milestones.

### Minor

5. **Exit-code claim overbroad.**
   - Location: PLAN.md M0 deliverable "Error handling … exit codes 0–12"
   - Issue: `CLIError.ExitCode()` maps E001→2 … E004→10, E010→11, default→1. No code path returns 12 at this SHA.
   - Why it matters: CI consumers may rely on the documented range; publish the real table (already noted in `6342b33` review).

6. **Appendix timeline leaves original estimates in estimate columns after marking DONE.**
   - Location: PLAN.md §6 M0 row (`3–4 | 3–4 | ✅ DONE (1 day)`)
   - Issue: Cumulative/weeks columns unchanged while status shows 1-day actual; "Total estimated effort: 25–34 weeks" not reconciled.
   - Why it matters: Minor doc inconsistency; M0 Actual note partially mitigates.

7. **Deliverable claims `cmd/`, `internal/`, `pkg/` layout ✅.**
   - Location: PLAN.md M0 deliverable table
   - Issue: `cmd/` absent at this SHA (see #1).

**Verified accurate (no finding):** 12 global persistent flags (`internal/cli/root.go`); rate limiter `NewRateLimiter(45)` (`internal/api/client.go`); 5 regions (`internal/api/regions.go`); E001–E010 constants defined; tag `v0.1.0-m0` exists (points at feat commit `6342b33`); "34 files, 2612 lines of Go" matches the M0 feat commit stats (reasonable as milestone attribution, not this docs commit's size).

## Recommendations

1. **Correct M0 completion narrative** — either land `cmd/clockify/main.go` (with `CLIError.ExitCode()` → `os.Exit`) before declaring COMPLETE, or re-state M0 as "foundation packages complete; entrypoint/buildable binary deferred" and un-check binary/`--help` criteria.
2. **Restore accurate keychain wording** — "env-var fallback only at M0; OS keyring delivered in M1 (`7214ea1`)"; do not mark the original keychain deliverable ✅ on a stub.
3. **Annotate criteria not met at M0** — live API auth validation, durable credential storage, runnable binary metrics.
4. **Publish the real exit-code table** (2–11 + default 1) or denser mapping; drop "0–12" until implemented.
5. **Later milestones (not scored against this SHA):** real keyring (M1 `7214ea1`), API validation (M1), `cmd/clockify/main.go` (M4 `f0d6787`), packaging/build metadata (M9).

## Conflicts / unresolved

None. Single-expert in-session review (P3 docs/chore; task guidance prefers no blocking subagents). No conflicting expert positions to weight.

## Verdict

**needs rework**

Structural doc hygiene is fine (status header, checklist format, actual-effort note, footer), but the commit’s core job — recording M0 completion — overstates DoD against PLAN M0 criteria and the tree at this SHA: no buildable entrypoint despite "measured" binary metrics, keychain marked complete on an env-var stub, and auth/config end-to-end criteria checked without the code that would make them true. Fix the status narrative (or land the missing P0 foundation pieces) before treating this as a green M0 baseline.

## HEAD triage (2026-10-05)

- **Current HEAD:** `ac64914`
- **Original verdict:** needs rework (unchanged by this triage)
- **Legend:** ✅ completed at HEAD · ⚠️ partially fixed · ❌ still open · ➖ N/A or cannot verify

| # | Item | Status | Evidence |
|---|------|--------|----------|
| C1 | F1 — M0 COMPLETE with binary criteria checked, no entrypoint | ⚠️ | Tree fixed: `cmd/clockify/main.go:18-27` exists (`package main`; `os.Exit(cliErr.ExitCode())` at :24); `Makefile:23` builds `./cmd/clockify`; git log shows `cmd/clockify/main.go` first lands in `f0d6787` (M4). Docs not fixed: PLAN.md:73-79 M0 success criteria still all `[x]` (`clockify --help`, "measured: 5ms"/"measured: 12MB") with no deferral annotation — M0 section unchanged from the reviewed commit. |
| C2 | F2 — OS keychain deliverable ✅ on env-var stub; criterion weakened | ⚠️ | Tree fixed: `internal/config/credentials.go:7` imports `github.com/zalando/go-keyring`; `getKeyring`/`setKeyring`/`deleteKeyring` (:53-67) call `keyring.Get/Set/Delete` (service `clockify-cli`, per-profile keys); `go.mod:8`; landed M1 `7214ea1`. Docs not fixed: PLAN.md:65 M0 row still `OS keychain integration | ✅ | Env var fallback implemented`; criterion PLAN.md:75 still drops "in keychain". M1 section (PLAN.md:102,111) documents real keyring correctly. |
| C3 | F3 — "HTTP client successfully authenticates" checked without validation code | ⚠️ | Tree fixed: `internal/cli/configure.go:125-133` `validateAPIKey` issues `GET /user` before `config.SetAPIKey` (called at :52, stored at :92); `internal/cli/status.go:92` validates on demand (landed M1 `7214ea1`). Docs not fixed: PLAN.md:76 M0 criterion still `[x]` unannotated. |
| C4 | F4 — "config set/get/list works end-to-end" checked without runnable CLI (+ profiles-wipe risk) | ⚠️ | Tree fixed: runnable CLI (`cmd/clockify/main.go`); `internal/cli/config.go:140-195` config `set` uses `config.LoadRaw()` + `Save(pc)` preserving `pc.Profiles`; `unset` same pattern (:217-260); no `api_key` case in set/unset (unknown-key errors :186,:250) so secrets are not persisted via `config set`; `config.Save` writes 0600 (`internal/config/config.go:172-180`). Docs not fixed: PLAN.md:74 M0 criterion still `[x]` without annotation. |
| C5 | F5 — Exit-code claim "0–12" overbroad | ❌ | Code unchanged: `internal/errors/errors.go:105-129` `ExitCode()` maps E001→2, E002→3, E003→4, E004→10, E005→5, E006→6, E007→7, E008→8, E009→9, E010→11, default→1 — no path returns 12. Docs claim unchanged: PLAN.md:68 M0 row still `Error handling | ✅ | E001–E010 categories, exit codes 0–12`. (Real table now published elsewhere — see C11.) |
| C6 | F6 — Appendix timeline leaves estimates unreconciled after DONE | ✅ | PLAN.md §6.1 (:966-979) rewritten as "Original Est. vs Actual" — M0 row `3–4 weeks | ~1 day | ✅ **DONE**`; §6.2 (:983-987) reconciles cumulative actual (~13.5 dev days) with velocity explanation; "Total estimated effort: 25–34 weeks" no longer present in PLAN.md. |
| C7 | F7 — Deliverable claims `cmd/`, `internal/`, `pkg/` layout ✅ | ⚠️ | Tree now has `cmd/` (`cmd/clockify/main.go`; `Makefile:23` targets `./cmd/clockify`). Docs still misattribute: PLAN.md:62 M0 row claims `cmd/`, `internal/`, `pkg/` layout ✅ as an M0 deliverable; git log shows `cmd/clockify/main.go` first appears in `f0d6787` (M4), not M0. |
| C8 | R1 — Correct M0 completion narrative (land main.go or restate M0 as deferred) | ⚠️ | Half satisfied: `cmd/clockify/main.go` landed M4 `f0d6787` with `CLIError.ExitCode()` → `os.Exit` (:24). M0 narrative NOT re-stated: PLAN.md:55-86 M0 section unchanged — still `✅ **COMPLETE** — Tagged v0.1.0-m0` with all criteria `[x]` and no "foundation packages complete; entrypoint deferred" note. |
| C9 | R2 — Restore accurate keychain wording in M0 docs | ❌ | Not done: PLAN.md:65 M0 row still `OS keychain integration | ✅ | Env var fallback implemented` (stub marked complete); requested wording ("env-var fallback only at M0; OS keyring delivered in M1 `7214ea1`") absent from the M0 section. M1 section (PLAN.md:102) correctly documents `go-keyring` but does not correct the M0 audit-trail claim. |
| C10 | R3 — Annotate M0 criteria not met at M0 (auth validation, durable storage, binary metrics) | ❌ | Not done: PLAN.md:73-79 M0 success criteria still all `[x]` (binary size/startup metrics, HTTP auth, config e2e, configure stores credentials) with no annotations that these were delivered in M1 (`7214ea1`) / M4 (`f0d6787`). |
| C11 | R4 — Publish real exit-code table (2–11 + default 1); drop "0–12" | ⚠️ | Partially done: real table published at README.md:225-239 (0, 2–11 matching implementation) plus per-code `docs/errors/` (e.g. `docs/errors/E001-auth.md`); PLAN.md:612 and :850 document E001→2 … E010→11. Not done: PLAN.md M0 row :68 still says "exit codes 0–12"; code still has no exit-12 path (`internal/errors/errors.go:105-129`). |
| C12 | R5 — Later milestones: real keyring (M1), API validation (M1), main.go (M4), packaging (M9) | ✅ | All landed: real keyring M1 `7214ea1` (`internal/config/credentials.go:7,53-67`; `go.mod:8`); API key validation M1 (`internal/cli/configure.go:125-128`; `internal/cli/status.go:92`); `cmd/clockify/main.go` M4 `f0d6787` (git log first commit); packaging/build metadata M9 (`Makefile:1-5` ldflags into `main.*`, goreleaser targets :6,:69-70; PLAN.md M9 section :520+). |

### Summary
- Completed: 2 · Partial: 7 · Still open: 3 · N/A: 0

### Completed at HEAD
- **C6** — Appendix timeline reconciled: PLAN.md §6.1/§6.2 now show Original vs Actual (M0 `3–4 weeks → ~1 day`) and a cumulative-actual velocity note; the unreconciled "25–34 weeks" total is gone.
- **C12** — All later-milestone deliverables from Recommendation 5 have landed: real `go-keyring` (M1), API-key validation via `GET /user` (M1), `cmd/clockify/main.go` wiring `CLIError.ExitCode()` (M4), packaging/ldflags/goreleaser (M9).

### Still open at HEAD
- **C5** — PLAN.md M0 row still claims "exit codes 0–12"; `CLIError.ExitCode()` still maps 2–11 + default 1 with no return-12 path (`internal/errors/errors.go:105-129`).
- **C9** — PLAN.md M0 keychain deliverable still marked `✅ | Env var fallback implemented`; the env-var stub is still presented as completed OS keychain integration at M0.
- **C10** — PLAN.md M0 success criteria remain annotation-free: all `[x]` including binary metrics and auth/config end-to-end claims that were not met at M0.

Partially fixed (not counted as open above; tree/docs improved but PLAN.md M0 section still overstates M0 DoD): **C1, C2, C3, C4, C7, C8, C11** — entrypoint/keyring/validation/config CLI now exist in the tree and correct tables exist in README/docs/errors/later PLAN sections, but the M0 status section (PLAN.md:55-86) was never reworked, so the original findings' doc claims remain false as an M0 audit trail.

