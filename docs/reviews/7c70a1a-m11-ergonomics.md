# Commit Review — 7c70a1a

- **Commit:** 7c70a1a197bd7c7626fd355b5e3dc36d16066abf
- **Subject:** feat: M11 - Daily-Driver Ergonomics
- **Date:** 2026-10-04
- **Experts engaged:** Full Stack Developer, QA Engineer
- **Skill mode:** Auto-Assignment (kims-team)
- **Review date:** 2026-10-05
- **Related plan:** PLAN.md §5.7 Review Backlog

## Diff summary

Implements M11 Daily-Driver Ergonomics (+2928/−366 across 25 files) against PLAN §M11 / `M11-SPRINT-PLAN.md`. In-session expert review of the tree **at this SHA** (temp worktree at `7c70a1a`, removed after review; product code untouched).

- `pkg/duration/human.go` (+184) + `human_test.go`: `ParseHumanDateTime` — `now`, `HH:MM`, am/pm, `yesterday HH:MM`, bare dates, naive `YYYY-MM-DDTHH:MM[:SS]`, ±offsets, RFC3339; local TZ via `locationFromConfig`
- `internal/cli/m11_resolve.go` (new, 218): `resolveTaskRef`, `resolveTagRefs`, `resolveTimeEntrySelector` (`current`/`last`), `ambiguousRefError` (E004 + candidates), `looksLikeID`
- `internal/cli/m11_power.go` (new, 538): clone, split, multi-entry update, flexible start parsing
- `internal/cli/picker.go` (new, 299): hand-rolled pickers + `interactiveRequested` / `interactiveBlockedReason` (CI/quiet/confirm/machine-output/non-TTY guards)
- `internal/cli/aliases.go`: `in`/`out` aliases; `ge`/`ce` retargeted to `time-entry get`/`create`
- `internal/cli/time_entry.go` (±): names on flags, `current`/`last`, interactive wiring, completion on project/task/tags
- `internal/cli/report.go`: presets `today|yesterday|this-week|last-week|this-month|last-month`
- `internal/cli/task.go`: `task done`, `task quick-add` (`--quiet` prints ID)
- `internal/output/template.go` + `detect.go` + `workspace.go` `formatOutput`: `--output template` / `--output ids`
- `internal/config/config.go`: `interactive`, `timezone` keys; `internal/cli/config.go` get/set/unset
- PLAN.md / M11-SPRINT-PLAN.md marked complete; tag `v0.11.0-m11`

Suite at this SHA: `go test ./...` green (except a synthetic alias-flag probe added only during this review, not committed).

## Findings

### Critical

1. **`in`/`out`/`ge`/`ce` aliases do not inherit target local flags — headline daily-driver command is broken** — `internal/cli/aliases.go:16-27`. `registerAliases` copies only `Use`/`Short`/`RunE`/`Run` onto a new root command; it never copies (or references) the target’s local flags. Root persistent flags (`--profile`, `--output`, `--quiet`, `--confirm`, …) work; anything local to `time-entry start` (`--project`, `--task`, `--tags`, `--description`, `--billable`, `-i/--interactive`, `--start`) does not. Confirmed at this SHA with a one-off probe: `newRootCmd().Execute()` on `in --project Website Redesign` returns `unknown flag: --project`. PLAN M11 success criterion #1 (`clockify in --project "Website Redesign" …`) and any `clockify in -i` / `clockify out` with flags fail before `RunE`. Matters because the entire M11 value proposition is “no IDs typed” via these aliases; `time-entry start --project …` still works, but the advertised ergonomic entrypoint does not.

### Major

1. **`split` is not atomic — mid-sequence API failure corrupts the source entry** — `internal/cli/m11_power.go` `runTimeEntrySplit`. After confirmation the code updates the original entry to end at the first split point, then creates remaining segments one-by-one. Validation-before-write is correct, but a create failure after the first `UpdateTimeEntry` leaves a shortened original and zero/partial segments with no rollback and no warning that the entry was already mutated. Matters for a destructive multi-write on billing data.

2. **Interactive pickers recreate `bufio.NewReader(os.Stdin)` per prompt — input loss across multi-step flows** — `internal/cli/picker.go` `promptLine`, `pickItem`, `pickTags`. Each call constructs a fresh reader; bufio may consume buffered stdin beyond the current line, so subsequent prompts miss typed-ahead input (classic multi-reader stdin bug). Matters for the S2 UX path (`project` → `task` → `tags` → `description`) where `-i` is the whole feature. Also: `pickItem`’s 5-attempt filter loop returns `""` on empty line (skip) which is fine, but buffering remains.

3. **Multi-entry update cannot clear fields and registers unused `--file`/`--stdin`** — `internal/cli/m11_power.go` `runTimeEntryUpdateMulti` + `internal/cli/time_entry.go` `newTimeEntryUpdateCmd`. Merge uses `description != ""`, `projectID != ""`, etc. instead of `Flags().Changed` (billable correctly uses `Changed`). Explicit empty cannot clear project/task/description/tags. `--file`/`--stdin` are registered on `update` but never read (only `--stdin-ids` is). Same empty-vs-Changed pattern flagged in the M8 review on create JSON merge — recurring defect class.

4. **`time-entry list` default scope silently changed to “today”** — `internal/cli/time_entry.go` list path: when neither `--start`/`--end` nor `--in-progress` is set, the command injects today’s local day range. Previously (and per “existing scripts … keep working unchanged” success criterion) a bare list would not be limited to today. Daily-driver UX may want this default, but it is a silent behavior change for scripts that enumerate history via `clockify le` without date flags. Should be explicit (flag or documented breaking change), not an implicit default.

### Minor

1. **`FormatTemplate` comment promises a trailing newline but does not add one** — `internal/output/template.go:11-21`. Comment says “Ensure trailing newline for CLI friendliness”; code returns without writing `\n`. Tests trim before assert, so this never fails CI. `--output template` output is therefore not always line-oriented like table/ids.

2. **`interactiveRequested` ignores richer config synonyms** — `internal/cli/picker.go:13-23` checks `cfg.Interactive` for `"true"` only, while `config.Config.InteractiveEnabled()` accepts `true/1/yes/on`. `config set interactive yes` is normalized to `"true"` on write, so common paths are fine; hand-edited YAML or partial loads that leave `"yes"` would not enable pickers.

3. **`resolveTagRefs` UUID fast-path does not verify the tag exists** — `internal/cli/m11_resolve.go:84-90` appends any hex-32 ref as an ID; task resolution falls through to `GetTask` on miss. Invalid tag IDs fail later at the API with a less actionable error than the E004 candidates path.

4. **`ParseHumanDateTime` hardcodes `"monday"` for relative dates** — `pkg/duration/human.go` (`ParseRelativeDateIn(s, "monday", …)`). `weekStarts` config is wired on list filters but not on the human-time grammar; harmless for `today`/`yesterday`, wrong for any future relative-week form in this parser.

5. **QA test gaps that let Critical #1 ship** — `internal/cli/m11_test.go`: `TestAliasesRegistered` only asserts `findCommand(root, "time-entry start")` etc. resolve; it never executes `in --project …` or checks flag sets. `TestHumanTimeOffset` discards errors and only re-checks `09:00`. No unit/integration coverage for clone, split, multi-update, report presets, or picker stdin behavior. `pkg/duration/human_test.go` is solid for the grammar itself.

6. **Report preset `this-week` end defaults to `today`** — `internal/cli/report.go` `newReportPresetCmd("this-week", "this-week", "today")` (week-to-date) while `last-week` uses a full-week end via ad-hoc `+6d` / `+1m-1d` pseudo-expressions parsed with `Sscanf`. Inconsistent semantics; pseudo-expression parsing is brittle and hardcodes monday in `duration.ParseRelativeDate`.

7. **Clone with only `--start` starts a timer** — `m11_power.go` clone: `--start` alone leaves `End` empty and starts a running timer at that time (dead `else if` branch documents the choice). Matches PLAN wording but is easy to misread when cloning a completed entry “at the same start.”

## Recommendations

1. **Fix alias flag inheritance (Critical):** in `registerAliases`, copy the target command’s local flag set onto the alias (or make aliases thin wrappers that set `cmd.Flags().AddFlagSet(target.Flags())` before execute / use Cobra `SetFlags` + shared run). Add a regression test that executes `in --project X` / `in -i` and asserts no `unknown flag`. This alone blocks calling M11 “done” against its success criteria.
2. **Make `split` safer:** either (a) create all new segments first from a copy plan, then update the original last, or (b) wrap in a compensating rollback that restores the original end time if any create fails; at minimum surface a clear error stating the entry was already shortened and how to repair.
3. **Reuse one stdin reader** across `maybePromptMissing` prompts; thread it through `pickItem`/`pickTags`/`promptLine`. Add a unit test with a piped multi-line stdin.
4. **Align multi-update merge with `Flags().Changed`** (including clearing); either implement or remove unused `--file`/`--stdin` on `update`.
5. **Make list’s default date range opt-in or documented:** prefer requiring `--start/--end`/preset, or add `--all`/`--range`, and call out the change in release notes / M10 docs. Backward-compat criterion should be tested explicitly.
6. **Honor the template trailing-newline contract** (or delete the comment); consider `templateFuncs` docs for what fields `--output ids` reflects on.
7. **Use `cfg.InteractiveEnabled()`** in `interactiveRequested`; route human relative dates through `weekStartFromConfig`.
8. **QA:** expand alias tests to full `Execute` with local flags; add httptest-backed tests for clone/split/update-multi happy path + split failure path; delete vacuous assertions.

**Later milestones:** M10 quality/docs (`ac64914`) and any post-v1.0.0 fixes may address some of these; this review scores **7c70a1a on its own merits** against the M11 message and PLAN §M11 criteria. If HEAD already fixed the alias flag copy or list default, record that as a follow-up fix commit rather than rewriting history — `v0.11.0-m11` points at this SHA.

## Conflicts / unresolved

- **Full Stack Developer vs QA Engineer:** no substantive disagreement. FSD weighted implementation defects (aliases, split atomicity, merge semantics); QA weighted test blind spots that allowed the alias defect and weak offset/interactive coverage to pass. Both converge on **needs rework**.
- **Conflict resolution (kims-team):** confidence-weighted — alias flag inheritance is objective (cobra parse error at this SHA); no escalation needed. Consensus report: same verdict from both perspectives.
- **Product decision flagged for user:** whether bare `time-entry list` defaulting to “today” is intentional M11 UX or an accidental regression vs “scripts keep working.” Recommend explicit decision + docs either way.

## Verdict

**needs rework**

Strong M11 surface area (human time grammar, name resolution with candidates, clone/split/selectors, pickers, presets, template/ids) and a green unit suite at this SHA, but the headline ergonomic entrypoints (`in`/`out` with flags) fail at flag parse, `split` can corrupt entries on partial API failure, and interactive stdin handling is unsafe — against explicit PLAN success criteria. Fix alias flag inheritance and split atomicity before treating M11 as accepted.

## HEAD triage (2026-10-05)

- **Current HEAD:** `ac64914`
- **Original verdict:** unchanged (`needs rework` at 7c70a1a — historical verdict stands)
- **Legend:** ✅ completed at HEAD · ⚠️ partially fixed · ❌ still open · ➖ N/A or cannot verify

Basis: every M11-relevant file is byte-identical between `7c70a1a` and `ac64914` (`git diff --stat 7c70a1a..ac64914` empty for `aliases.go`, `picker.go`, `m11_power.go`, `m11_resolve.go`, `m11_test.go`, `time_entry.go`, `report.go`, `template.go`, `human.go`, `human_test.go`). `ac64914` (M10 — Quality, Testing & Documentation) is the only commit after `7c70a1a` and added generated command docs, CI, and test/docs polish — it did not touch M11 runtime code. New corroborating evidence at HEAD: generated docs `docs/cli/clockify_in.md` list only `-h, --help` as local options for `in`, and `docs/cli/clockify_le.md` documents no default date range for `le`.

| # | Item | Status | Evidence |
|---|------|--------|----------|
| Critical #1 | `in`/`out`/`ge`/`ce` aliases do not inherit target local flags | ❌ still open | `internal/cli/aliases.go:21-26` still builds a new `cobra.Command` copying only `Use`/`Short`/`RunE`/`Run`; no flag-set copy (`AddFlagSet(target.Flags())`). Zero diff since 7c70a1a. Corroborated by HEAD docs `docs/cli/clockify_in.md` (local options = `-h` only) and `docs/cli/clockify_le.md`. README.md:8/96 and PLAN.md:648 still advertise `clockify in --project …` as working. |
| Major #1 | `split` not atomic — source entry mutated before creates | ❌ still open | `internal/cli/m11_power.go:290-302` still calls `UpdateTimeEntry` on the original first; create loop `:306-337` returns raw `err` on failure (`:317-318`, `:334-336`) with no rollback and no warning that the entry was already shortened. |
| Major #2 | Interactive pickers recreate `bufio.NewReader(os.Stdin)` per prompt | ❌ still open | `internal/cli/picker.go:148` (`promptLine`), `:163` (`pickItem`), `:225` (`pickTags`) each construct a fresh reader; no shared stdin reader threaded through `maybePromptMissing` (`:61-138`). |
| Major #3 | Multi-entry update: merge via `!= ""`; unused `--file`/`--stdin` | ❌ still open | `internal/cli/m11_power.go:476-490` still merges project/task/description/tags on non-empty values (`tagIDs != nil`); only `billable` uses `Flags().Changed` (`:407`, `:488`). `internal/cli/time_entry.go:474-475` still registers `--file`/`--stdin` on `update`, but `runTimeEntryUpdateMulti` reads only `--stdin-ids` (`m11_power.go:370`). |
| Major #4 | `time-entry list` default scope silently "today" | ❌ still open | `internal/cli/time_entry.go:97-102` comment "Default to today if no date filters provided" still injects local 00:00–23:59 when neither `--start`/`--end` nor `--in-progress` is set. No `--all`/`--range` opt-in; `docs/cli/clockify_le.md` / `docs/cli/clockify_time-entry_list.md` still document no default range; no release-notes callout found in the M10 docs. |
| Minor #1 | `FormatTemplate` comment promises trailing newline; code adds none | ❌ still open | `internal/output/template.go:20-21` still has the comment and `return nil` without writing `\n`. `internal/cli/m11_test.go:139` and `internal/output/template_test.go` still `TrimSpace` before assert. |
| Minor #2 | `interactiveRequested` checks only `"true"` | ❌ still open | `internal/cli/picker.go:23` still `strings.EqualFold(strings.TrimSpace(cfg.Interactive), "true")`; `config.Config.InteractiveEnabled()` (`internal/config/config.go:99-108`, accepts `true/1/yes/on`) is never called from the CLI picker path. |
| Minor #3 | `resolveTagRefs` UUID fast-path does not verify the tag exists | ❌ still open | `internal/cli/m11_resolve.go:101-103` still appends any `looksLikeID` ref as an ID without an API check; contrast `resolveTaskRef` (`:77-87`) which falls through to `GetTask` on miss. |
| Minor #4 | `ParseHumanDateTime` hardcodes `"monday"` for relative dates | ❌ still open | `pkg/duration/human.go:86` and `:90` still call `ParseRelativeDateIn(..., "monday", ...)`. `weekStartFromConfig` exists (`internal/cli/m11_resolve.go:190-200`) and is used by list filters (`internal/cli/time_entry.go:81`), but `ParseHumanDateTime` takes no cfg/weekStart parameter. |
| Minor #5 | QA gaps that let Critical #1 ship (offset/interactive coverage) | ❌ still open | `internal/cli/m11_test.go` unchanged (213 lines, identical to 7c70a1a). `TestAliasesRegistered` (`:181-196`) still only `findCommand` — no `Execute` on `in --project`/`in -i`. `TestHumanTimeOffset` (`:198-213`) still discards `err`/`tm` (`_ = err`, `_ = tm`) and re-checks only `09:00`. Still no tests for clone, split, multi-update, report presets (`report_test.go` has no `this-week` preset semantics test), or picker stdin. `pkg/duration/human_test.go` remains solid for the grammar, including offsets (`:65-78`). |
| Minor #6 | Report preset `this-week` end defaults to `today`; brittle pseudo-expression parse | ❌ still open | `internal/cli/report.go:43` still `newReportPresetCmd("this-week", "this-week", "today")` vs `:44` `last-week` end `"last-week+6d"`; `:62-82` still `Sscanf`-parses `+Nd`/`+Nm-1d` with `"monday"` hardcoded at `:65`. |
| Minor #7 | Clone with only `--start` starts a timer (misread trap) | ❌ still open | `internal/cli/m11_power.go:162-180` unchanged; with only `--start`, `req.End` stays empty via the dead `else if` (`:176-180`), and the conflict-check path (`:184-192`) starts a running timer. Behavior matches PLAN wording per original review; the documentation/clarity trap remains at HEAD. |
| Rec #1 | Fix alias flag inheritance + regression test that executes `in --project …` | ❌ still open | See Critical #1 — code and test coverage identical at HEAD. |
| Rec #2 | Make `split` safer (create-then-update or compensating rollback) | ❌ still open | See Major #1 — write order unchanged. |
| Rec #3 | Reuse one stdin reader + piped multi-line stdin test | ❌ still open | See Major #2 — per-prompt `bufio.NewReader(os.Stdin)` unchanged; no picker stdin test added. |
| Rec #4 | Align multi-update with `Flags().Changed` (incl. clearing); implement or drop `--file`/`--stdin` | ❌ still open | See Major #3. |
| Rec #5 | Make list's default date range opt-in or documented; test backward-compat criterion | ❌ still open | See Major #4 — implicit today default retained; generated CLI docs still silent on the default; README examples still use explicit `--start today --end today` without noting the bare-list change. |
| Rec #6 | Honor template trailing-newline contract; document what `--output ids` reflects | ❌ still open | See Minor #1; `internal/output/template.go:81-85` `templateFuncs` still only `upper`/`lower`; `extractIDs` (`:35-79`) behavior still undocumented. |
| Rec #7 | Use `cfg.InteractiveEnabled()`; route human relative dates through `weekStartFromConfig` | ❌ still open | See Minors #2 and #4 — neither half applied. |
| Rec #8 | QA: Execute-level alias tests; httptest clone/split/update-multi; delete vacuous assertions | ❌ still open | See Minor #5 — no new tests; targeted re-run at HEAD (`go test ./internal/cli/ -run 'TestAliasesRegistered|TestHumanTimeOffset|…'`) still green vacuously. |

### Summary

HEAD `ac64914` (M10 — Quality, Testing & Documentation) is the only commit after `7c70a1a`, and a diff of every M11-relevant file between those SHAs is empty: **none of the 12 original findings or 8 recommendations were addressed at HEAD**. Counts: ✅ 0 · ⚠️ 0 · ❌ 20 · ➖ 0. The M10 generated command docs independently corroborate the Critical alias defect (`docs/cli/clockify_in.md` exposes only `-h` for `in`) and the undocumented list default (`docs/cli/clockify_le.md`). Original verdict **needs rework** stands unchanged; M10 closed milestone-quality gates (CI, generated docs, test polish) but did not touch M11 runtime code.

### Completed at HEAD

_None. No M11 finding or recommendation is fixed at `ac64914`._

### Still open at HEAD

- **Critical #1 / Rec #1 — alias flag inheritance:** still the M11-blocking defect; `clockify in --project …` / `in -i` still fail at flag parse; README/PLAN still claim the success criterion.
- **Major #1 / Rec #2 — split atomicity:** source entry still mutated first; mid-sequence API failure still leaves a shortened original with no rollback or warning.
- **Major #2 / Rec #3 — interactive stdin buffering:** pickers still allocate a fresh `bufio.Reader(os.Stdin)` per prompt.
- **Major #3 / Rec #4 — multi-update merge semantics:** explicit empty still cannot clear project/task/description/tags; `--file`/`--stdin` still registered on `update` but unread.
- **Major #4 / Rec #5 — list default "today":** still an implicit, undocumented scope change for bare `time-entry list`/`le`; the product decision flagged in the original review remains open.
- **Minors #1–#6 / Recs #6–#8 — template newline, interactive synonyms, tag ID verification, hardcoded monday, QA coverage (offset/interactive), report preset semantics:** all unchanged at HEAD.
- **Minor #7 — clone with only `--start`:** unchanged behavior (dead-else-if comment intact); matches PLAN wording but the misread trap remains.
