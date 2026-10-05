# M11: Daily-Driver Ergonomics — Sprint Plan

**Milestone:** M11
**Sprint Duration:** 3 × 1-day sprints (PLAN.md estimate 2–3 days)
**Developer:** 1 (solo)
**Start Date:** 2026-10-04
**End Date:** 2026-10-07 (target)
**Status:** ✅ **COMPLETE**
**Completed:** 2026-10-04
**Prerequisite:** M9 complete — tag `v0.10.0-m9` on `91b362b` (working tree clean)
**Spec reference:** `clockify-cli-spec.md` §6.4 (interactive create), §8.8 (time-entry) — clone/split/`in`/`out`/`current`/`last`/template are **not in spec**; M11 additions documented for M10 docs
**PLAN.md reference:** PLAN.md §M11, §4.3 ADR (interactive lib), §3.2 sequencing (M11 after M9, before M10)

**Implementation summary:** Shipped all P0–P3 M11 deliverables in one day. Hand-rolled interactive pickers (no `huh`/bubbletea — binary ~17.6MB, startup ~4ms). Human time grammar in `pkg/duration` with local TZ. Name resolution + CLIError candidates on time-entry flags. clone/split/selectors/multi-update, aliases (in/out + fixed ge/ce), report presets, task shortcuts, template/ids output. `go test ./...` green.

---

## 1. Sprint Structure

| Sprint | Theme | Days | Goal |
|--------|-------|------|------|
| **S1** | Names + Human Time | D1 | `--project`/`--task`/`--tags` accept names on time-entry commands; `--start`/`--end` accept `HH:MM`, `yesterday 09:00`, offsets, `now`; timezone conversion; ambiguity errors with candidates + E-codes |
| **S2** | Interactive Pickers | D2 | Opt-in TTY fuzzy pickers for `start`/`create` when flags omitted; strict non-TTY/CI guards; library ADR resolved; task-name completion gap fixed |
| **S3** | Power Loop + Closeout | D3 | `clone`, `split`, `current`/`last` selectors, multi-entry update, `in`/`out` aliases (+ fix dead `ge`/`ce`), report presets, task shortcuts; P3 template if time allows; E2E + PLAN.md + tag `v0.11.0-m11` |

**Why this split:** S1 is pure resolution + parsing — no new dependencies, highest reuse, unblocks everything else. S2 is the only sprint with UX risk (hangs in pipes, binary size) and needs a library evaluation first. S3 is feature assembly on a stable base and can cut P2/P3 without losing the M11 core value proposition.

---

## 2. Current State Assessment

### Already Implemented (M0–M9)

| Feature | Status | Notes |
|---------|--------|-------|
| Time-entry CRUD + timers | ✅ | `internal/cli/time_entry.go` — list/get/create/update/delete/start/stop/bulk |
| Relative dates on list | ✅ | `duration.ParseRelativeDate` — today/yesterday/this-week/last-week/this-month/last-month |
| Duration parsing | ✅ | `pkg/duration.Parse` — `1h`, `1.5h`, `30m`, `1h30m`, raw seconds |
| Name resolvers (elsewhere) | ✅ | `resolveProjectRef`, `resolveTaskByName`, `resolveUserRef`, `resolveTagByName`, etc. — used by report/task/schedule/expense/**not** time-entry |
| TTL name/list caches | ✅ | M8 `resource_cache.go` — projects/clients/tags/users/workspaces |
| Dynamic completion | ✅ | Cache-only `ValidArgsFunction` — never blocks on network |
| Confirmation prompts | ✅ | y/N `fmt.Scanln` + `--confirm`; `confirmDestructive` helper (M7) |
| TTY detection | ⚠️ | `output.IsTerminal()` — **stdout only**; no stdin check; no `CI` env guard |
| Aliases | ⚠️ | `lw`/`le`/`lp`/`cfg` work; **`ge`/`ce` are dead** (target `get time-entry` / `create time-entry` — wrong path order) |
| Template output | ❌ | No `text/template` support; unknown `--output` values silently fall back to table |

### What M11 Needs to Build

| Component | Status | Notes |
|-----------|--------|-------|
| Time-entry name resolution | ❌ | `--project`/`--task`/`--tags` are ID-only on list/create/update/start |
| `resolveTagRefs` (multi) | ❌ | Comma-separated names → IDs; only single `resolveTagByName` exists |
| `resolveTaskRef` (ID-or-name) | ❌ | Needs project context like `resolveProjectRef`; `resolveTaskByName` exists but is not ID-first |
| Ambiguity → CLIError + candidates | ❌ | Resolvers use plain `fmt.Errorf`; `internal/cli` has **zero** `internal/errors` CLIError uses |
| Human time grammar | ❌ | `parseStartTimeFlexible` rejects `HH:MM`, `now`, offsets; no TZ conversion |
| Timezone source | ❌ | Config has unused `UserProfile.TimeZone`; relative dates use `time.Local`; no `time.LoadLocation` |
| Interactive pickers | ❌ | No `-i` flag, no `interactive` config key, no picker UI, no stdin-TTY check |
| Task name cache/completion | ❌ | No `"tasks"` cache resource; report `--task` completion wrongly uses project names |
| `clone` / `split` | ❌ | Not implemented |
| `current` / `last` selectors | ❌ | Running lookup exists (`GetTimeEntryInProgress`); "last completed" needs list+filter |
| Multi-entry update | ❌ | `update` takes a single `--id` only |
| `in` / `out` aliases | ❌ | — |
| Report presets | ❌ | `report today\|yesterday\|...` shortcuts |
| Task shortcuts | ❌ | `task done`, `task quick-add` |
| `--output template` / `--output ids` | ❌ | P3 — greenfield in `internal/output` |

---

## 3. Prerequisites & Decisions Needed

| # | Item | Impact | Action / Decision |
|---|------|--------|-------------------|
| P1 | **Prompt library** | S2 pickers; binary size + startup budgets (<20MB, <50ms) | ADR in §5.1 — prefer `charmbracelet/huh`; cut to hand-rolled `x/term` raw-mode picker if binary grows past ~20MB or startup regresses >50ms |
| P2 | **Timezone source** | S1 human times; DST edge cases | ADR in §5.2 — default `time.Local` (OS); optional `timezone` config key read once; document; do not require API user timezone in M11 |
| P3 | **Task picker data** | S2 — tasks are project-scoped, never cached | Extend M8 cache with per-project task lists OR fetch live only when `-i` prompts for task (acceptable: pickers are opt-in) |
| P4 | **Working tree / tag** | Closeout | Prerequisite `v0.10.0-m9` on `91b362b`; confirm clean before S1 |

---

## 4. Spec ↔ Implementation Decision Log

| Topic | Spec Says | Reality / PLAN | M11 Decision |
|-------|-----------|----------------|-------------|
| Time-entry `--project/--task/--tags` | Examples use IDs (spec internally inconsistent on list `--tags`) | ID-only today | **Accept ID or name**; ID fast path first; backward compatible |
| Ambiguous name | Spec E-codes; M11 wants candidates | Plain `fmt.Errorf`, no candidates | Fail with `CLIError` (E004 validation or E003 not-found as fits) + candidate list (name + ID) + "use full ID" |
| `--start/--end` formats | RFC3339 / relative dates | Same; bare dates parse **UTC** | Extend grammar in `pkg/duration`; bare dates + `HH:MM` resolve in **local TZ** then convert to UTC for API |
| Interactive mode | Spec §6.4: `create time-entry --interactive` prompts | None; PLAN M11 refines to opt-in pickers | `-i/--interactive` on `start`/`create`; config `interactive: true`; **never** prompts when non-TTY, `--confirm`, `--quiet`, `CI`, or `--output` is json/csv/yaml |
| Prompt library | — | PLAN §4.3: evaluate `huh` vs survey vs custom | **`huh` first** (maintained, fuzzy); measure size/startup; fallback hand-rolled |
| Local time parsing | — | PLAN §4.3: hand-rolled small grammar | **Hand-rolled** fixed grammar — no NLP; easy to document |
| `clone` / `split` / `current` / `last` | Not in spec | M11 additions | Implement per PLAN.md deliverables table; document for M10 |
| `in` / `out` | Not in spec | Alias infra exists but `ge`/`ce` broken | Add `in`→`time-entry start`, `out`→`time-entry stop`; **fix** `ge`→`time-entry get`, `ce`→`time-entry create` in same change |
| Multi-entry update | Not in spec | Single-ID update | `update <id> [<id>...]` + `--stdin-ids`; confirm summarizes count unless `--confirm` |
| Report presets | Not in spec | Reports take explicit `--start/--end` | Thin wrappers mapping preset → `report summary` (or `--type detailed`) with matching range; all filter flags still apply |
| Task shortcuts | Spec `task --estimate` seconds | CLI already exceeds spec | `task done <task> --project P`; `task quick-add <name> --project P`; `--quiet` prints ID only |
| Template output | Not in spec | No template formatter | P3: `--output template --template '...'` + `--output ids`; cut first if behind |
| ID-only flags elsewhere | — | Audit result (research §6): real gap is time-entry only | Invoice `--number`, webhook/schedule/expense IDs have no name concept — **documented as intentional**; date flags on M7 commands gain HH:MM via shared parser |
| `week_starts` config | — | Config field exists; all call sites hardcode `"monday"` | Wire config `WeekStarts` into `ParseRelativeDate` callers when cheap (S1 or S3); document default |

---

## 5. Architecture Notes

### 5.1 Prompt library evaluation (S2 gate)

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **`charmbracelet/huh`** | Maintained; fuzzy filter; accessible mode; modern UX | New dep tree (bubbletea ecosystem); binary size + startup risk vs M10 budgets | **Preferred** — spike first |
| Hand-rolled (`x/term` raw + numbered/fuzzy line picker) | Zero new deps; tiny; predictable | More code; worse UX | **Fallback** if size/startup fail |
| `AlecAivazis/survey` | — | Archived | Reject |

**Spike (S2 start, ≤1h):** add `huh`, `go build`, measure `ls -la` binary + `time ./clockify version`. If size >20MB or startup >50ms → hand-rolled path. Record result in this file §5.1 outcome.

**Spike outcome:** Hand-rolled pickers chosen — no new dependencies; size/startup budgets held (~17.6MB / ~4ms).

### 5.2 Time parsing design (S1)

New file: `pkg/duration/human.go` (keep next to `relative.go`).

```text
ParseHumanDateTime(s string, loc *time.Location, now time.Time) (time.Time, error)

Grammar (fixed, documented):
  now
  -15m / +1h / -2d          → offset from now (d = days)
  HH:MM / H:MMam / H:MMpm   → today at that local time
  yesterday HH:MM           → yesterday at that local time
  today HH:MM               → today at that local time
  YYYY-MM-DD                → that date at 00:00 local (fixes today's UTC bare-date parse)
  YYYY-MM-DDTHH:MM[:SS]     → local naive → attach loc
  RFC3339                   → unchanged, zone preserved
  relative date only        → existing ParseRelativeDate (midnight local)

Reject: ambiguous NLP ("morning"), empty strings, out-of-range times.
```

- Caller passes `time.Local` (or config TZ) and injectable `now` for tests.
- API boundary: CLI still sends RFC3339/ISO8601; conversion to UTC happens at the format call site in `time_entry.go` (and later M7 date flags via `parseDateTimeCLI` adoption).
- Unify later: `parseStartTimeFlexible` → delegate to `ParseHumanDateTime`; adopt in `parseDateTimeCLI` / `parseReportDate` in S1 or S3 (same grammar, different return shapes).

### 5.3 Name resolution design (S1)

| Helper | Location | Behavior |
|--------|----------|----------|
| `resolveProjectRef` | existing `task.go:500` | Reuse as-is (ID → name fuzzy) |
| `resolveTaskRef(ctx, client, projectID, ref)` | new in `task.go` | If UUID/prefix → `GetTask`; else `resolveTaskByName(projectID, ref)` |
| `resolveTagRefs(ctx, client, refs []string) ([]string, error)` | new in `tag.go` | Each ref: UUID → pass through; else `resolveTagByName`; aggregate IDs |
| Ambiguity errors | all resolvers touched | Wrap in `errors.CLIError` with context + multi-line candidates: `  - <name>  (<id>)`; exit code from `CLIError.ExitCode()` |
| Completion wiring | `time_entry.go` | Register `completeProjectRefC` / `completeTagRefC` / new `completeTaskRefC` on list/create/update/start flags (today only create/update have project/tags) |

Resolution must be **skippable for tests** — follow existing factory-function pattern; resolve only when input is not already a UUID (fast path keeps scripts on IDs at zero extra latency).

### 5.4 Interactive picker design (S2)

```text
start/create RunE:
  if interactiveRequested(cmd, cfg) && allGuardChecksPass(cmd):
      for each missing field among project, task, tags, description:
          picker from cache (projects/tags/users) or per-project task list
  else if any required field still missing:
      actionable CLIError — "missing --project; run with -i on a TTY or pass --project <id|name>"

Guard checks (ALL must pass to prompt):
  - -i/--interactive flag OR config interactive:true
  - stdout is TTY AND stdin is TTY   (extend output.IsTerminal → stdin+stdout helper)
  - !--quiet
  - !--confirm
  - CI env unset (CI, GITHUB_ACTIONS, BUILD_ID, etc.)
  - --output not in {json, csv, yaml}  (empty or table on TTY only)
Never prompt on: pipes, redirect, CI, machine output, non-TTY.
```

Data source: `cachedList` / `cachedCompletionNames` (M8). On cache miss: run list API once (pickers are opt-in — network OK) then warm cache. Task pickers: list tasks for selected project live; optionally write `list.tasks.<ws>.<pid>.json`.

### 5.5 Power-loop features (S3)

| Feature | Behavior | API surface |
|---------|----------|-------------|
| `clone [id\|last]` | Copy project/task/tags/description/billable; `start` timer unless `--start/--end` given; any flag overrides copied field | `GetTimeEntry` + create/start path |
| `split <id> --at HH:MM [--at ...]` | Validate split points strictly inside entry; update original end; create remainder(s) with identical metadata | `UpdateTimeEntry` + `CreateTimeEntry` |
| `current` / `last` | On `get`/`update`/`delete`/`stop`: `current` → `GetTimeEntryInProgress`; `last` → list recent completed, pick max start | existing methods |
| Multi-entry update | Args as IDs + optional `--stdin-ids`; same flags applied to each; confirmation summarizes count | loop `UpdateTimeEntry` |
| `in` / `out` | Root aliases | `aliases.go` |
| Report presets | `report today` etc. → summary with preset range; `--type detailed` optional | existing report cmds |
| `task done` / `quick-add` | Status=DONE / create with defaults | existing task API |

**Backward compatibility hold-firm:** any existing script that passes UUIDs and RFC3339 keeps working unchanged. New forms are additive.

---

## 6. Sprint 1: Names + Human Time (Day 1)

### User Stories

**US-1.1: Names on time-entry resource flags**

> As a user, I want `--project Website`, `--task QA`, and `--tags billable,frontend` to work on time-entry commands without looking up UUIDs.

**Acceptance Criteria:**
- [ ] `time-entry start/create/update/list` accept project **name or ID** on `--project`
- [ ] `--task` accepts task name (scoped to project) or ID; project resolved first
- [ ] `--tags` accepts comma-separated tag **names or IDs**; mixed lists OK
- [ ] ID fast path: UUID input performs no list/search calls
- [ ] Ambiguous names fail with CLIError + candidate list (name + ID) + suggested fix
- [ ] Unknown names fail with not-found CLIError (no silent empty results on create/start)
- [ ] Existing ID-only scripts unchanged (backward compatible)
- [ ] Flag completions registered on list/start (create/update already partial)
- [ ] New `completeTaskRefC` wired; fix report `--task` completion bug if cheap
- [ ] Unit tests: resolution, ambiguity, UUID passthrough, tags multi-resolve

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.1 | `resolveTaskRef` (ID-first + name) + tests | 1h | `internal/cli/task.go`, `task_test.go` | tests |
| T1.2 | `resolveTagRefs` multi + tests | 1h | `internal/cli/tag.go`, `tag_test.go` | tests |
| T1.3 | Upgrade ambiguity errors to CLIError + candidates in project/task/tag/user resolvers | 1.5h | `task.go`, `project.go`, `tag.go`, `client.go`, `m7_helpers.go` | tests + manual ambiguous case |
| T1.4 | Wire resolution into `time-entry` list/create/update/start flags; update help text (ID or name) | 2h | `time_entry.go` | manual + tests |
| T1.5 | Completion wiring: list/start project/task/tags; `completeTaskRefC`; report `--task` fix | 1h | `time_entry.go`, `resource_cache.go`, `report.go` | `clockify completion` smoke |
| T1.6 | Audit remaining ID-only flags → decision log (already mostly done in research §6) | 0.5h | this file §4 | review |

**US-1.2: Human time input + timezone**

> As a user, I want `--start 9:00`, `yesterday 10:30`, `now`, and `-15m` to work in local time.

**Acceptance Criteria:**
- [ ] Grammar in §5.2 implemented in `pkg/duration` with table-driven tests (incl. DST edge if feasible: e.g. spring-forward local times)
- [ ] `parseStartTimeFlexible` delegates to new parser; bare dates parse in **local** TZ (bugfix vs today's UTC)
- [ ] API payloads remain RFC3339/ISO8601 (converted from local)
- [ ] Optional: `timezone` config key (IANA name) honored when set; default OS local
- [ ] `week_starts` config wired into relative-date callers (or explicitly deferred with note)
- [ ] M7 `parseDateTimeCLI` / report date parsing adopt same grammar where return shape allows (or documented deferral)
- [ ] Unit tests: each grammar form, invalid inputs, offset math, injection of `now`/`loc`

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T1.7 | `ParseHumanDateTime` + tests | 2h | `pkg/duration/human.go`, `human_test.go` | tests |
| T1.8 | Wire into `parseStartTimeFlexible`; local-TZ bare dates; UTC conversion at API boundary | 1h | `time_entry.go` | manual create with `--start 9:00` |
| T1.9 | Optional `timezone` config + WeekStarts wiring | 1h | `internal/config/config.go`, callers | unit + config set |
| T1.10 | Adopt parser in m7_helpers / report dates (or defer) | 1h | `m7_helpers.go`, `report.go` | manual schedule/report date |
| T1.11 | Full test suite + help-text audit | 0.5h | all | `go test ./...` |

**Sprint 1 Exit Criteria:**
- [ ] `clockify time-entry start --project "Website Redesign" --tags billable,frontend` resolves names
- [ ] `clockify time-entry create --start 9:00 --end 10:30 --project Website` creates correct local-time entry
- [ ] Ambiguous project name prints candidates and non-zero exit
- [ ] UUID/RFC3339 paths unchanged
- [ ] `go test ./...` green
- [ ] No new dependencies added in S1

**Sprint 1 Estimated:** ~11h (1 focused day; first day may stretch — P0 names + times are the milestone's core)

---

## 7. Sprint 2: Interactive Pickers (Day 2)

### User Stories

**US-2.1: Library spike**

> As a maintainer, I want the picker UX chosen against real size/startup numbers.

**Acceptance Criteria:**
- [ ] `huh` spike built; binary size + startup measured vs budgets
- [ ] Outcome recorded in §5.1 (huh or hand-rolled)
- [ ] Decision reflected in go.mod (or spike reverted)

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.1 | Spike huh; measure; decide; record | 1h | go.mod (spike), this file §5.1 | size/time numbers |

**US-2.2: Opt-in pickers on `start`/`create`**

> As an interactive user, I want missing flags to open fuzzy pickers on a TTY — and the same command in a pipe to fail fast, not hang.

**Acceptance Criteria:**
- [ ] `-i/--interactive` flag on `time-entry start` and `time-entry create`
- [ ] Config key `interactive: true` enables pickers when flag omitted
- [ ] Pickers for: project, task (after project), tags, description (text input)
- [ ] Data from M8 caches; cache miss → one API list then warm cache
- [ ] **Guards (all):** non-TTY stdout or stdin → no prompt; `CI`/`GITHUB_ACTIONS` set → no prompt; `--quiet` → no prompt; `--confirm` → no prompt; `--output json|csv|yaml` → no prompt
- [ ] Missing required fields without interactive eligibility → actionable CLIError (never hang, never empty create)
- [ ] Task pickers list tasks for chosen project (live or cached per-project)
- [ ] `in` not yet required; `start -i` is the interactive daily path (aliases land S3)
- [ ] Unit tests: guard matrix (table-driven); picker selection → correct IDs on mock client
- [ ] Manual E2E: TTY happy path; `echo | clockify time-entry start -i` / `CI=1 clockify time-entry start -i` fail fast with error text
- [ ] Completion still cache-only (no API during tab-complete)

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T2.2 | stdin+stdout TTY helper; CI env detection; `interactive` config key | 1h | `internal/output/detect.go` or cli helper, `internal/config/config.go` | unit |
| T2.3 | Picker implementation (huh or hand-rolled): fuzzy select + text input; cache-backed lists | 3h | `internal/cli/picker.go` (new) | manual TTY |
| T2.4 | Wire into `start`/`create` RunE with guard matrix; missing-field errors | 2h | `time_entry.go` | unit + manual |
| T2.5 | Task list path (per-project); optional task cache write | 1h | `resource_cache.go`, picker | manual |
| T2.6 | Guard-matrix unit tests + CLI tests with mock client | 1.5h | `picker_test.go`, `time_entry_test.go` | tests |
| T2.7 | Binary size/startup re-check after dep; document | 0.5h | Makefile / CI notes | `time ./clockify version` |

**Sprint 2 Exit Criteria:**
- [ ] `clockify time-entry start -i` on TTY with `-i` opens pickers and starts timer
- [ ] Same command with `CI=1` or piped stdin exits non-zero with actionable missing-flag error — **no hang**
- [ ] `--output json` never prompts
- [ ] Size/startup budgets hold (or fallback implemented)
- [ ] `go test ./...` green

**Sprint 2 Estimated:** ~10h (1 focused day + buffer for picker UX iteration)

---

## 8. Sprint 3: Power Loop + Closeout (Day 3)

### User Stories

**US-3.1: Selectors and power commands**

> As a daily user, I want `stop current`, `delete last`, `clone last`, and `split --at 12:00` without copying IDs.

**Acceptance Criteria:**
- [ ] `time-entry get/update/delete/stop` accept `current` and `last` in place of `--id`
- [ ] `current` errors clearly when no timer running
- [ ] `last` = most recent completed entry (list + max start; documented tie-break)
- [ ] `time-entry clone [id|last]` copies metadata; flags override; `--start/--end` creates completed entry else starts timer
- [ ] `time-entry split <id> --at HH:MM [--at ...]` validates points inside entry; original end updated; remainders created with same project/task/tags/description/billable
- [ ] Split failure modes: outside range, unsorted, equal-to-bounds → CLIError before any write
- [ ] Unit tests for selector resolution, clone flag overrides, split math + validation

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.1 | `current`/`last` selectors + tests | 1.5h | `time_entry.go` | tests + manual |
| T3.2 | `clone` command + tests | 2h | `time_entry.go` (or `time_entry_clone.go`) | manual + unit |
| T3.3 | `split` command + tests | 2.5h | same | manual + unit |

**US-3.2: Multi-entry update + aliases + presets**

**Acceptance Criteria:**
- [ ] `time-entry update <id> [<id>...]` applies same flags to each
- [ ] `--stdin-ids` reads IDs from stdin (pipe from `jq`); works with `--output json` upstream
- [ ] Confirmation prompt summarizes count (`Update N time entries?`) unless `--confirm`
- [ ] Partial failure: report which IDs failed; exit non-zero; continue policy documented (stop-on-first vs collect — **recommend collect + summary**)
- [ ] `clockify in` → `time-entry start`; `clockify out` → `time-entry stop`
- [ ] **Fix dead aliases:** `ge` → `time-entry get`, `ce` → `time-entry create` (spec aliases; current targets inverted)
- [ ] `report today|yesterday|this-week|last-week|this-month|last-month` presets; existing filter flags still apply
- [ ] `task done <task> --project P` sets status DONE
- [ ] `task quick-add <name> --project P` creates with defaults; `--quiet` prints ID only
- [ ] Dynamic completion wired for new commands/flags where applicable (M8 cache)
- [ ] Unit tests for multi-update, presets mapping, task shortcuts, alias registration

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.4 | Multi-entry update + stdin-ids + confirm summary | 2h | `time_entry.go` | tests + manual pipe |
| T3.5 | Aliases: add `in`/`out`; fix `ge`/`ce` | 0.5h | `aliases.go`, tests | `clockify --help` + run |
| T3.6 | Report presets | 1h | `report.go` | manual |
| T3.7 | `task done` / `task quick-add` | 1h | `task.go` | manual + unit |

**US-3.3: P3 template output (optional)**

**Acceptance Criteria (if scheduled):**
- [ ] `--output template --template '{{.Description}} {{.Duration}}'` works on list/get
- [ ] `--output ids` shorthand prints IDs only
- [ ] Invalid template → clear error; never panic
- [ ] TemplateFormatter in `internal/output` + `Detect` case + tests

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.8 | TemplateFormatter + ids shorthand + tests | 2h | `internal/output/template.go`, `formatter.go`, cli flags | manual + unit |

**US-3.4: E2E verification & closeout**

**Acceptance Criteria:**
- [ ] All PLAN.md M11 success criteria verified (unit/manual as appropriate)
- [ ] Backward-compat checklist: UUID+RFC3339 scripts unchanged; pipes never hang; CI never prompts
- [ ] Performance: startup <50ms, binary size <20MB re-checked
- [ ] `go test ./...` green; `make check` if lint available
- [ ] PLAN.md M11 marked complete; decisions logged (§4 + §5); timeline §6.1/§6.2 updated
- [ ] Spec gaps documented for M10 (clone/split/in/out/selectors/presets/shortcuts/template)
- [ ] Commit + tag `v0.11.0-m11`

**Tasks:**

| # | Task | Hours | Files | Verify |
|---|------|-------|-------|--------|
| T3.9 | Full suite + fix | 1h | all | `go test ./...` |
| T3.10 | CLI smoke matrix (see §12) | 1h | binary | checklist |
| T3.11 | Update PLAN.md M11 status/decisions/timeline | 1h | `PLAN.md` | review |
| T3.12 | Commit + tag `v0.11.0-m11` | 0.25h | git | tag exists |

**Sprint 3 Exit Criteria:**
- [ ] `clone last`, `split <id> --at 12:00`, `stop current`, `delete last` work end-to-end
- [ ] Multi-update + stdin-ids work; confirm summarizes count
- [ ] `in`/`out` work; `ge`/`ce` fixed
- [ ] Report presets + task shortcuts work
- [ ] PLAN.md updated; tag `v0.11.0-m11`; tests green

**Sprint 3 Estimated:** ~12h (1 focused day; P3 template cuts first if behind)

---

## 9. Risk Mitigation

| Risk | Impact | Likelihood | Mitigation | Sprint |
|------|--------|------------|------------|--------|
| **Interactive mode breaks pipelines** | High — core value regresses | Medium | Strict guard matrix; never prompt non-TTY/CI/json/quiet/confirm; fail-fast errors; table-driven guard tests | S2 |
| **`huh` binary size / startup** | Medium — M10 budgets | Medium | Spike first; hand-rolled fallback recorded in ADR | S2 |
| **Timezone/DST bugs** | Medium — wrong billable hours | Medium | Injected `now`/`loc` tests; document grammar; prefer local-TZ correctness over clever NLP | S1 |
| **Ambiguous names change error behavior** | Low-Medium — scripts parsing errors | Low | Candidates + E-code; ID fast path unchanged; document in M10 | S1 |
| **Split corrupts time entries** | High — data loss on wrong split | Low | Validate before any write; multi-step only after validation; unit tests on boundary times; manual E2E | S3 |
| **Multi-update partial failure** | Medium | Medium | Collect errors + summary; `--confirm` required for multi; document continue policy | S3 |
| **Alias fix breaks muscle memory** | Low | Low | `ge`/`ce` are currently dead — fix is strictly an improvement | S3 |
| **Scope creep (P2/P3)** | Medium — slips closeout | Medium | Cut order: P3 template → report presets → task shortcuts; keep P0+P1 non-negotiable | S3 |
| **Task cache complexity** | Low-Medium | Medium | Live list inside pickers OK (opt-in); cache write optional | S2 |
| **New `internal/errors` usage inconsistencies** | Low | Medium | Only upgrade resolvers touched by M11; leave M7 resolvers as-is unless trivial | S1 |

---

## 10. Trade-offs: Where to Cut Corners vs. Where to Hold Firm

### Hold Firm (Non-Negotiable)

| Area | Why |
|------|-----|
| **Never prompt in non-TTY / CI / json / quiet / confirm** | Pipelines and scripts are the product's spine |
| **Backward compatibility for IDs + RFC3339** | Existing users must not break |
| **UUID fast path with no extra API calls** | Performance + rate limits |
| **Ambiguous names fail loudly with candidates** | Silent wrong-entity writes are worse than errors |
| **Interactive is opt-in** | `-i` flag or explicit config; never default-on when piped |
| **Split validates before writes** | Destructive correctness |
| **Unit tests green + size/startup budgets** | M10 depends on M11 not regressing |

### Cut Corners (If Behind Schedule)

| Area | How to Cut | Impact |
|------|------------|--------|
| **P3 template / ids output** | Defer to post-M11 / M10 backlog | Users keep `jq` |
| **Report presets** | Defer | Explicit `--start/--end` still works |
| **Task shortcuts** (`done`/`quick-add`) | Defer | `task update --status DONE` exists |
| **M7 date-flag HH:MM adoption** | Defer; document | Time-entry is the daily path |
| **`timezone` config key** | Ship OS-local only | Config field can land later |
| **Task name cache** | Live fetch in pickers only | Completions remain cache-only |
| **`week_starts` config wiring** | Keep hardcoded monday + note | Low user impact |

### Cut Corners — Emergency Only

| Area | How to Cut | Impact |
|------|------------|--------|
| **All interactive pickers** | Ship S1 names+times only; M11 = P0 names/time + P1 power loop | Still huge daily-driver win; pickers become M11.1 |
| **clone/split** | Ship selectors + multi-update only | Partial power loop |
| **Hand-rolled vs huh debate** | Default hand-rolled to protect budgets | Slightly worse picker UX |

---

## 11. File-Level Task Summary

| File | Sprint | Tasks | Est. Hours |
|------|--------|-------|------------|
| `pkg/duration/human.go` + tests | S1 | T1.7 | 2h |
| `internal/cli/time_entry.go` (+ clone/split) | S1–S3 | T1.4, T1.8, T2.4, T3.1–T3.4 | ~8h |
| `internal/cli/task.go` / `tag.go` / resolvers | S1 | T1.1–T1.3 | 3.5h |
| `internal/cli/resource_cache.go` | S1–S2 | T1.5, T2.5 | 1.5h |
| `internal/config/config.go` | S1–S2 | T1.9, T2.2 | 1.5h |
| `internal/cli/picker.go` + tests | S2 | T2.3, T2.6 | 4.5h |
| `internal/output/detect.go` / template.go | S2–S3 | T2.2, T3.8 | 1.5h |
| `internal/cli/aliases.go` | S3 | T3.5 | 0.5h |
| `internal/cli/report.go` | S3 | T3.6 | 1h |
| `internal/cli/task.go` (shortcuts) | S3 | T3.7 | 1h |
| `PLAN.md` + this plan | S3 | T3.11 + status updates | 1h |
| E2E + tag | S3 | T3.9–T3.10, T3.12 | 2.25h |

**Total estimated:** ~33h ≈ **3 focused days** (matches PLAN.md 2–3 day estimate at this velocity).

---

## 12. CLI Smoke Matrix (S3 Exit)

| # | Command | Expect |
|---|---------|--------|
| 1 | `time-entry start --project "Name" --task "Name" --tags a,b` | Timer starts; no IDs typed |
| 2 | `time-entry create --start 9:00 --end 10:30 --project Name` | Correct local-time entry |
| 3 | `time-entry create --start yesterday --duration 2h --project Name` | Relative date + duration |
| 4 | `time-entry start --start now -i` (TTY) | Pickers → start |
| 5 | `CI=1 time-entry start -i` (piped or not) | Actionable error; **no hang** |
| 6 | `--output json time-entry start -i` | No prompt |
| 7 | `time-entry stop current` | Stops running timer |
| 8 | `time-entry delete last --confirm` | Deletes most recent completed |
| 9 | `time-entry clone last` | New timer with copied metadata |
| 10 | `time-entry split <id> --at 12:00` | Split succeeds; validate outside `--at` fails |
| 11 | `time-entry update id1 id2 --description X --confirm` | Both updated; confirm shows count |
| 12 | `le --output json \| jq -r '.[].id' \| time-entry update --stdin-ids --project Name --confirm` | Pipe-driven multi-update |
| 13 | `in` / `out` | Alias to start/stop |
| 14 | `ge <id>` / `ce --project X ...` | Fixed aliases work |
| 15 | `report today --output json` | Preset range applied |
| 16 | `task done <name> --project P` | Status DONE |
| 17 | Ambiguous `--project Web` | Candidates listed; non-zero exit |
| 18 | `--output template --template '{{.Description}}'` (if shipped) | Renders; bad template errors cleanly |
| 19 | Existing: `time-entry list --project <uuid> --start 2026-10-01 --end 2026-10-02 --output json` | Unchanged behavior |
| 20 | `clockify version` / `time ./clockify version` | <50ms; size <20MB |

---

## 13. M11 Overall Definition of Done

| Criterion | Verification | Status |
|-----------|-------------|--------|
| Names on time-entry `--project/--task/--tags` | Smoke 1 | ☐ |
| Ambiguous names → candidates + non-zero exit | Smoke 17 | ☐ |
| Human times: `HH:MM`, `yesterday 09:00`, `now`, offsets | Smoke 2–3 | ☐ |
| Local TZ for naive times; RFC3339 still works | Smoke 2, 19 | ☐ |
| Interactive pickers on TTY with `-i` / config | Smoke 4 | ☐ |
| No prompts/hangs in CI, pipes, json, quiet, confirm | Smoke 5–6 + guard tests | ☐ |
| `clone`, `split`, `current`, `last` | Smoke 7–10 | ☐ |
| Multi-entry update + stdin-ids + confirm count | Smoke 11–12 | ☐ |
| `in`/`out` aliases; `ge`/`ce` fixed | Smoke 13–14 | ☐ |
| Report presets + task shortcuts | Smoke 15–16 | ☐ |
| Backward compatible IDs/RFC3339 | Smoke 19 | ☐ |
| Completion still cache-only; new flags wired | Manual tab-complete | ☐ |
| Startup <50ms; binary size <20MB | Smoke 20 | ☐ |
| `go test ./...` green | CI / local | ☐ |
| PLAN.md M11 marked complete; decisions logged | Review | ☐ |
| Tag `v0.11.0-m11` created | `git tag -l` | ☐ |
| Spec gaps documented for M10 | PLAN.md / M10 notes | ☐ |

---

## 14. Success Criteria Mapping (PLAN.md §M11 → Sprint Tasks)

| PLAN.md Success Criterion | Sprint | Tasks |
|---------------------------|--------|-------|
| `clockify in --project "Website Redesign" --task "QA" --tags "billable,frontend"` starts a timer with no IDs typed | S1+S3 | T1.4, T3.5 |
| `clockify time-entry create --start 9:00 --end 10:30 --project Website` creates a correct entry in local time | S1 | T1.7–T1.8 |
| `clockify in` with no flags on a TTY (with `-i` or config) opens pickers; same command in a pipe or with `CI=1` fails fast — no hangs | S2 | T2.2–T2.6 |
| `clone last`, `split <id> --at 12:00`, `stop current`, `delete last` work end-to-end | S3 | T3.1–T3.3 |
| Every new flag/command has dynamic completion wired through the M8 cache | S1–S3 | T1.5, T2.x, T3.5–T3.7 |
| Existing scripts using IDs and RFC3339 timestamps keep working unchanged | All | Hold-firm list §10; Smoke 19 |

---

## 15. Daily Standup Template

```markdown
## Day X Standup — M11 Sprint Y

**Done today:**
- [List completed tasks]

**Blocked:**
- [Any blockers — e.g. waiting on library spike decision]

**Plan for tomorrow:**
- [Next tasks]

**Confidence for sprint:**
- 🟢 On track / 🟡 At risk / 🔴 Behind
```

---

## 16. Out of Scope (M11 Hard Cuts)

- Plugin system (post-M10)
- Full README / man pages (M10 owns docs; M11 documents new flags in PLAN.md + help text only)
- Invoice number / webhook name / schedule assignment name resolution (no name concept in API)
- Spec §6.4 "type the name to confirm" destructive pattern (documented deviation remains y/N + `--confirm`)
- NLP-style natural language dates ("morning", "next tuesday")
- Interactive mode for M7 resource commands (time-entry start/create only)
- PDF/chart report output
- Differential sync or local timer daemon

---

*This plan is a living document. Update task estimates and status as work progresses. Record the prompt-library spike outcome in §5.1 before committing to picker UX in S2.*

*Last updated: 2026-10-04 — M11 implemented. All sprints complete. Tag `v0.11.0-m11`. Next milestone: M10.*
