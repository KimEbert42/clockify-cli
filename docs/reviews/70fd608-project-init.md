# Commit Review — 70fd608

- **Commit:** 70fd608bc3df06721e112cde0f85b52391ad1bd9
- **Subject:** chore: project init
- **Date:** 2026-09-21
- **Experts engaged:** Tech Team Lead
- **Skill mode:** Auto-Assignment (kims-team)
- **Review date:** 2026-10-05
- **Related plan:** PLAN.md §5.7 Review Backlog #1 (P3)

## Diff summary

Single-file commit (1 file, +1254 lines): adds `clockify-cli-spec.md`, a comprehensive Clockify CLI design specification (Version 0.1.0 Draft, Status: Design Document, dated 2026-09-21). No code, tests, or build artifacts. The spec covers design principles (discoverability, progressive disclosure, pipe-friendly, idempotent-by-default, <50ms startup), a full resource-first command tree, global flags, resource ID resolution, auth/config (YAML profiles, OS keychain, regions), output formats (table/JSON/CSV/YAML), error categories E001–E010, retries, destructive-op confirmation, shell completion, CI/CD exit codes, caching, distribution/packaging, and a complete command reference with workflow appendices.

**Message accuracy:** "chore: project init" is accurate — this is a design-phase foundation doc, not scaffolding code. M0 code deliverables land in later commits (`6342b33` et al.).

## Findings

### Critical

None. Design-document-only commit; no runtime code to break at this tree.

### Major

| # | Issue | Location | Why it matters |
|---|--------|----------|----------------|
| 1 | **Command-shape contradiction.** Spec is internally inconsistent: §1.3 aliases, §2.4, §6.3, and appendices show verb-first (`clockify list projects`, `clockify get time-entry`, `clockify create time-entry`), while the §2.1 tree and §8 command reference show resource-first (`clockify project list`, `clockify time-entry start`). Two different CLIs are sketched in one document. | `clockify-cli-spec.md` §1.3 vs §2.1 vs §8.8 | Implementers must pick one; ambiguity at design time risks divergent UX, dead aliases, and wasted M0 work. Implementation later settled on resource-first + aliases. |
| 2 | **Credential-storage story contradicts itself.** §3.2 config example shows plaintext `ck_live_*` keys in YAML with comment "Encrypted at rest (see 3.3)", while §3.3 declares OS keychain (keyring) as the *primary* method and only `CLOCKIFY_API_KEY`/`--insecure-store` as fallbacks. It is never stated whether YAML ever stores raw keys. | §3.2–§3.3 | Security-sensitive design ambiguity. Later M0 review (`6342b33`) found `config set` could persist env-overloaded secrets into YAML — consistent with this unresolved spec point. |
| 3 | **Exit-code tables disagree.** §6.6 "CI/CD Mode" lists codes 0–6 only; Appendix C lists 0–12 (including config conflict, server, validation, update-required). Same spec, two contracts. | §6.6 vs Appendix C | CI scripts and tests can only honor one table; the shorter one under-specifies the E00X category mapping. |
| 4 | **Speculative API hosts presented as fact.** §3.6 region table maps `global/eu-de/us/uk/au` to concrete hostnames (`api.clockify.me`, `euc1.clockify.me`, `use2.clockify.me`, etc.) with no "verify against official docs" marker. | §3.6 | Design docs that assert unverified external endpoints can ship wrong defaults into the HTTP client. (Hindsight: hosts were later confirmed via docs per PLAN post-M1 note — not scored against this commit.) |

### Minor

| # | Issue | Location |
|---|--------|----------|
| 5 | `--api-key` as a global flag is a security footgun (shell history, process lists, CI logs) even though keyring/env are presented as preferred; spec does not warn. | §2.2, §3.4 |
| 6 | Error learn-more URL `https://clockify.github.io/cli/errors/E001` is aspirational/nonexistent at this commit. | §5.1 |
| 7 | Name-based resource resolution ("fuzzy match with warning") is specified for general use including destructive commands; design should force `--id` for delete/stop-type ops. | §2.3, §5.5 |
| 8 | Spec version/status header (0.1.0 Draft / Design Document) and "Future" plugin section are appropriate for an init commit — no action. | header, §6.5 |

## Recommendations

1. **Command shape:** treat resource-first (`resource action`) as canonical, matching §2.1/§8; keep aliases as sugar for verb-style habits. Already resolved in implementation (`internal/cli/*` resource-first + `aliases.go`) — note only.
2. **Credentials:** one rule in the spec — OS keychain primary; YAML never stores raw API keys; env/flag only for CI. Later M1 `7214ea1` (real keyring + API validation) addressed this in code.
3. **Exit codes:** publish a single table (Appendix C 0–12) as the CI contract and delete/shorten §6.6. Later M0 implementation review flagged exit-code mapping as needing rework; M10 quality docs work covered the documented contract.
4. **Region hosts:** mark §3.6 as provisional until verified against official Clockify docs; PLAN already records later confirmation via spec + docs.
5. **Security nits:** document `--api-key` as discouraged; prefer env/keychain. Deletions should require `--id` (or type-to-confirm), not fuzzy `--name` alone.
6. **Process:** for future design-doc commits, prefer marking speculative external facts (hosts, URLs) with TODO/verify notes so catch-up reviews do not re-litigate them as defects.

## Conflicts / unresolved

None between experts — single-expert light review per PLAN §5.7 P3 guidance (docs/chore). No multi-expert disagreement to resolve.

## Verdict

**approve with follow-ups**

On-message project-init commit delivering a strong, comprehensive design baseline that correctly frames M0 principles and the eventual CLI surface. Notable internal inconsistencies (command-shape, credential storage, exit-code tables, speculative hosts) are design-doc quality issues, not blockers for a draft spec; most were already resolved or corrected by later implementation milestones — still recorded honestly against this SHA.

## HEAD triage (2026-10-05)

- **Current HEAD:** `ac64914`
- **Original verdict:** unchanged (approve with follow-ups)
- **Legend:** ✅ completed at HEAD · ⚠️ partially fixed · ❌ still open · ➖ N/A or cannot verify

Key triage fact: `clockify-cli-spec.md` is byte-identical to its state at `70fd608` (`git log --follow` shows only the init commit; `git diff 70fd608 HEAD -- clockify-cli-spec.md` is empty). All spec-document defects below remain verbatim in the doc. Implementation-level resolutions are recorded separately in the evidence column; this is a light/docs project-init review, so most items are design-doc hygiene, not runtime blockers.

| # | Item | Status | Evidence |
|---|------|--------|----------|
| F1 | Command-shape contradiction (verb-first §1.3/§2.4/§6.3 vs resource-first §2.1/§8) | ⚠️ | Spec unchanged — both shapes still present in the doc. Implementation canonicalized resource-first: `internal/cli/aliases.go` (`lp`→`project list`, `in`/`out`→`time-entry start/stop`, `ge`/`ce`→`time-entry get/create`), README "Full command surface" tree, man/ and docs/cli/ all resource-first. Spec text not reconciled. |
| F2 | Credential-storage contradiction (§3.2 plaintext `ck_live_*` in YAML vs §3.3 keychain-primary) | ⚠️ | Code-side resolved at HEAD: real keyring via `zalando/go-keyring` (`internal/config/credentials.go`); `configure` stores key via `SetAPIKey` and writes YAML with only Region/WorkspaceID (`internal/cli/configure.go:92,101-116`); `config set` has no `api_key` case — rejects with "unknown config key" (`internal/cli/config.go:158-186`); config saved `0600` (`internal/config/config.go:180`); README security notes. Spec §3.2–3.3 still contradictory. |
| F3 | Exit-code tables disagree (§6.6 lists 0–6 vs Appendix C 0–12) | ❌ | Spec unchanged — both tables still in the doc verbatim. HEAD publishes the real contract elsewhere: README "Exit codes" table (0, 2–11) + `docs/errors/E00X-*.md` pages each stating its exit code; `internal/errors/errors.go` `CLIError.ExitCode()`. Spec §6.6 vs Appendix C contradiction not fixed in the reviewed document. |
| F4 | Speculative API hosts presented as fact (§3.6, no verify marker) | ❌ | Spec unchanged — §3.6 still asserts hosts as fact. Hosts themselves now confirmed via official docs (PLAN post-M1/M6 notes; `internal/api/regions.go` matches spec's API-host column). Note: PLAN §812 records spec §3.6's reports-host pattern (`reports.<host>` for regional) as **incorrect** — real regional reports use same host + `/report/v1` (`internal/api/regions.go` `ReportsBaseURL`). Recommendation to mark §3.6 provisional was never applied. |
| F5 | `--api-key` global flag footgun; spec warns nowhere (§2.2, §3.4) | ⚠️ | Flag still registered at HEAD with a neutral help string (`internal/cli/root.go:70`). README security notes now warn: "Never pass `--api-key` on shared machines; use profiles/keychain" and prefer env only in ephemeral CI secrets (README:335-340). Spec still carries no warning. |
| F6 | Learn-more URL `clockify.github.io/cli/errors/E001` aspirational (§5.1) | ⚠️ | Spec unchanged — fictional URL still at `clockify-cli-spec.md:408`. Implementation at HEAD prints real repo URLs: `ErrorCode.LearnMoreURL()` → `github.com/kim/clockify-cli/blob/main/docs/errors/...` (`internal/errors/errors.go:81-83`), and `docs/errors/E001-auth.md` … `E010-internal.md` exist and are linked from README. Spec not updated to match the shipped ADR. |
| F7 | Name-based fuzzy resolution allowed on destructive ops; should force `--id` (§2.3, §5.5) | ⚠️ | Not forced at HEAD: delete commands accept `--id` **or** `--name` with by-name resolution (e.g. `internal/cli/client.go:365-387`, `tag.go`, `custom_field.go`, `group.go`), gated by a y/N `confirmDestructive` prompt unless `--confirm` (`internal/cli/m7_helpers.go:132-153`). Confirmation exists (satisfies "or type-to-confirm" in part — requires typing `y`, not the name), but `--id` is not mandatory and `--confirm` bypasses everything. M11 names-everywhere work extends name resolution further. |
| F8 | Spec version/status header (0.1.0 Draft / Design Document) and Future plugin section appropriate | ➖ | No action was ever requested; header remains accurate as a historical design-phase marker. Nothing to fix at HEAD. |
| R1 | Command shape: resource-first canonical, aliases as sugar | ✅ | Done in implementation: resource-first surface is the documented, shipped CLI (README, man pages, docs/cli); aliases live in `internal/cli/aliases.go` with verb-style targets fixed to resource-first paths (`ge`→`time-entry get`, `ce`→`time-entry create`). |
| R2 | Credentials: one rule — keychain primary, YAML never stores raw keys, env/flag only for CI | ⚠️ | Code implements the rule (see F2). The recommendation's deliverable was a single rule **in the spec** — §3.2–3.3 still contradicts itself; `--insecure-store` from the spec was never implemented. |
| R3 | Exit codes: publish single table (0–12) as CI contract, shorten §6.6 | ⚠️ | Spec §6.6/Appendix C contradiction untouched. M10 shipped a real, documented CI contract instead: README table + per-code `docs/errors/` pages (0, 2–11; code 12 update-required reserved/unimplemented per M9 plan). Practical contract exists; spec tables still disagree. |
| R4 | Region hosts: mark §3.6 provisional until verified; PLAN records confirmation | ⚠️ | PLAN records official-docs confirmation (M6: regional reports = same host + `/report/v1`; main hosts confirmed). Spec §3.6 was never annotated provisional and is partially wrong on reports hosts (see F4). |
| R5 | Security nits: document `--api-key` discouraged; deletions require `--id`/type-to-confirm | ⚠️ | README documents `--api-key` as discouraged (✅ docs half). Deletions do not require `--id`; by-name delete + y/N confirm remains (⚠️). |
| R6 | Process: future design-docs should mark speculative external facts TODO/verify | ➖ | Process guidance with no verifiable artifact at HEAD beyond this spec (which was not re-annotated). No later design-doc commits exist to check adoption against. |

### Summary
Project-init review (docs-only commit, P3 light review) — original verdict "approve with follow-ups" holds. The design spec `clockify-cli-spec.md` was never revised after `70fd608`, so all four Major findings and most Minor findings remain present in the document itself. However, the substance those findings worried about was largely absorbed by later milestones: resource-first command shape + fixed aliases, keyring-primary credentials with YAML free of raw keys, a real published exit-code contract, and real learn-more URLs backed by `docs/errors/` pages. Two spec-level defects (exit-code table disagreement, unmarked/incorrect hosts in §3.6) remain outright open in the doc; the rest are half-fixed (implementation or README resolved; spec text not).

### Completed at HEAD
- **R1** — resource-first canonical command shape with working aliases (`internal/cli/aliases.go`, README command surface, man/docs/cli).
- Implementation-side resolutions noted in F1, F2, F5, F6 evidence: keychain storage + no-YAML-secrets via `config set`, README security warnings, real learn-more URLs + error docs pages.

### Still open at HEAD
- **F3** (❌) — spec §6.6 (0–6) vs Appendix C (0–12) still disagree; README publishes a third, real contract (0, 2–11).
- **F4** (❌) — §3.6 hosts still unmarked as provisional; reports-host column also wrong for regional workspaces (PLAN:812).
- **F1, F2, F5, F6, F7** (⚠️) — resolved in code/README where noted, but spec text unchanged: command-shape contradiction, credential-storage contradiction, missing `--api-key` warning, fictional learn-more URL, non-forced `--id` on destructive ops.
- **R2–R5** (⚠️) — recommendation deliverables that targeted the spec document were not applied there.
- **F8, R6** (➖) — no action / process guidance, not scoreable.
