#!/usr/bin/env bash
# scripts/review-commits.sh — catch-up commit reviews via opencode + kims-team
#
# Completes PLAN.md §5.7 Review Backlog by running `opencode run` for each
# unreviewed commit. Each session loads the kims-team skill, reviews the tree
# at that SHA, writes docs/reviews/<short>-<slug>.md, and updates PLAN.md.
#
# Usage:
#   scripts/review-commits.sh              # review all pending (priority order)
#   scripts/review-commits.sh --list       # show pending only
#   scripts/review-commits.sh --dry-run    # show commands without running
#   scripts/review-commits.sh --only c6e07b1
#   scripts/review-commits.sh --only c6e07b1,ac64914
#   scripts/review-commits.sh --commit     # after each review, git commit the review file
#   scripts/review-commits.sh --force      # re-review even if docs/reviews already exists
#
# Env:
#   OPENCODE_BIN     opencode binary (default: opencode)
#   REVIEW_LOG_DIR   log directory (default: .review-logs)
#   REVIEW_MODEL     optional -m provider/model for opencode run
#   REVIEW_TIMEOUT   seconds per review before kill (default: 420)

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

OPENCODE_BIN="${OPENCODE_BIN:-opencode}"
REVIEW_LOG_DIR="${REVIEW_LOG_DIR:-$ROOT/.review-logs}"
REVIEW_MODEL="${REVIEW_MODEL:-}"
REVIEW_TIMEOUT="${REVIEW_TIMEOUT:-420}"
REVIEW_DIR="$ROOT/docs/reviews"
PLAN="$ROOT/PLAN.md"

ONLY=""
DRY_RUN=0
LIST_ONLY=0
DO_COMMIT=0
FORCE=0

usage() {
  sed -n '2,20p' "$0" | sed 's/^# \{0,1\}//'
  exit 0
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage ;;
    --list) LIST_ONLY=1; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    --only) ONLY="${2:-}"; shift 2 ;;
    --only=*) ONLY="${1#*=}"; shift ;;
    --commit) DO_COMMIT=1; shift ;;
    --force) FORCE=1; shift ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      ;;
  esac
done

mkdir -p "$REVIEW_DIR" "$REVIEW_LOG_DIR"

# Backlog: short_sha|slug|milestone|experts|priority|subject
# Priority order matches PLAN.md §5.7 (P0 remaining → P3).
BACKLOG=(
  "c6e07b1|m3-time-entry|M3|Full Stack Developer, Backend Engineer, QA Engineer|P0|feat: M3 - Time Entry core daily driver"
  "ac64914|m10-quality-docs|M10|QA Engineer, Tech Team Lead, Security DevOps Engineer|P0|feat: M10 - Quality, Testing & Documentation (v1.0.0)"
  "7214ea1|keyring-api-validation|M1|Security DevOps Engineer, Backend Engineer|P1|feat: real keyring integration and API validation"
  "872f91b|m1-complete|M1|Backend Engineer, Full Stack Developer|P1|feat: M1 complete - multi-profile, cache, status validation"
  "2dc8d27|m2-sprint1-reads|M2|Backend Engineer|P1|feat: M2 Sprint 1 - workspace/user read operations"
  "df7e578|m2-complete-pagination|M2|Backend Engineer, Full Stack Developer|P1|M2 complete: project, client, tag commands + pagination"
  "f0d6787|m4-write-crud|M4|Backend Engineer, Full Stack Developer|P1|feat: M4 - Write Operations for project, client, tag CRUD"
  "4388f3e|m5-tasks|M5|Backend Engineer|P1|feat: M5 - Task Management CRUD under projects"
  "817c831|m6-reports|M6|Backend Engineer, Full Stack Developer|P1|feat: M6 - Reports & Analytics"
  "3168f3e|m7-secondary-resources|M7|Backend Engineer, Architect|P1|feat: M7 - Secondary Resources (full API coverage)"
  "969ca42|m8-completion-power|M8|Full Stack Developer, DevOps Engineer|P2|feat: M8 - Shell Completion & Power Features"
  "91b362b|m9-distribution|M9|DevOps Engineer, Security DevOps Engineer|P2|feat: M9 - Distribution & Packaging"
  "7c70a1a|m11-ergonomics|M11|Full Stack Developer, QA Engineer|P2|feat: M11 - Daily-Driver Ergonomics"
  "70fd608|project-init|M0|Tech Team Lead|P3|chore: project init"
  "83f61fc|create-plan|M0|Architect|P3|doc: create plan"
  "0aac85a|m0-plan-status|M0|Tech Team Lead|P3|docs: update PLAN.md with M0 completion status"
)

review_path_for() {
  local short="$1"
  local slug="$2"
  # Prefer exact planned name; also accept any existing docs/reviews/<short>-*.md
  local planned="$REVIEW_DIR/${short}-${slug}.md"
  if [[ -f "$planned" ]]; then
    echo "$planned"
    return 0
  fi
  local match
  match="$(find "$REVIEW_DIR" -maxdepth 1 -type f -name "${short}-*.md" 2>/dev/null | head -1 || true)"
  if [[ -n "$match" ]]; then
    echo "$match"
    return 0
  fi
  echo "$planned"
}

is_reviewed() {
  local short="$1" slug="$2"
  [[ -f "$(review_path_for "$short" "$slug")" ]]
}

full_sha() {
  git rev-parse "$1"
}

commit_meta() {
  local sha="$1"
  git log -1 --format='%H%n%ad%n%s' --date=short "$sha"
}

build_prompt() {
  local short="$1" slug="$2" milestone="$3" experts="$4" subject="$5"
  local full
  full="$(full_sha "$short")"

  cat <<EOF
You are completing a catch-up commit review for clockify-cli using the **kims-team** skill.

# Task

Load and follow the kims-team skill (Auto-Assignment mode). Review git commit \`$short\` and write the review artifact, then update PLAN.md.

# Commit under review

- Short SHA: \`$short\`
- Full SHA: \`$full\`
- Subject: $subject
- Milestone: $milestone
- Suggested experts: $experts

# How to inspect the code

This is a **historical review**. Read the tree **at that commit**, not current HEAD:

\`\`\`
git -C $ROOT show $short --stat
git -C $ROOT show $short:path/to/file.go
git -C $ROOT ls-tree -r --name-only $short
git -C $ROOT log -1 --format='%H%n%ad%n%s%n%b' --date=iso $short
\`\`\`

Do **not** modify product code. Do **not** check out the old commit in a way that dirties the working tree permanently (prefer \`git show\`).

# How to engage experts (important — avoid hangs)

Suggested experts: **$experts**

You may use the kims-team skill mindset, but **do not block on long-running subagents**:

1. Prefer reviewing **in this session** (read the SHA, form findings yourself).
2. If you spawn expert agents, spawn **at most 2 in parallel**, keep each task tight (one file area / one concern), and **do not wait indefinitely**.
3. If an agent does not return within a few minutes, **synthesize from what you already have** and write the review file anyway.
4. **Write the review file first**, then update PLAN.md, then print the summary line. Never leave the session without a review artifact if you have enough signal for a verdict.

# Deliverable 1 — review file

Write a markdown review to:

\`\`\`
$REVIEW_DIR/${short}-${slug}.md
\`\`\`

Use this template (PLAN.md §5.7):

\`\`\`markdown
# Commit Review — $short

- **Commit:** $full
- **Subject:** $subject
- **Date:** <YYYY-MM-DD from commit>
- **Experts engaged:** $experts
- **Skill mode:** Auto-Assignment (kims-team)
- **Review date:** <today YYYY-MM-DD>
- **Related plan:** PLAN.md §5.7 Review Backlog

## Diff summary
<files changed + what the commit does — brief>

## Findings
### Critical
### Major
### Minor
<each finding: location file:line if possible, issue, why it matters>

## Recommendations
<actionable follow-ups; note if later milestones already addressed them>

## Conflicts / unresolved
<none | options with trade-offs | flagged for user decision>

## Verdict
<approve | approve with follow-ups | needs rework>
<1–2 sentence overall judgment>
\`\`\`

If experts disagree, apply kims-team conflict resolution (confidence-weighted → escalation → consensus report) and document it under Conflicts.

# Deliverable 2 — update PLAN.md

In \`$PLAN\`, §5.7 Review Backlog:

1. Find the row containing \`$short\`.
2. Change its Review status from ⬜ Pending to:
   - \`✅ Reviewed — <verdict>\` using the verdict from your review file
3. Update the **Status** line above the table:
   - Recount reviewed vs pending (17 total commits in the backlog table).
   - Mention the newest review path: \`docs/reviews/${short}-${slug}.md\`
4. Update the **Last updated** footer at the end of PLAN.md:
   - Note the review count and that the latest completed review is \`$short\` (verdict).
   - Keep post-v1.0.0 next steps text intact.

Do not rewrite unrelated PLAN.md sections.

# Constraints

- Review the commit **on its own merits** vs its message and the PLAN milestone criteria for that milestone.
- You may note (without scoring) that later milestones fixed issues — still give an honest verdict for this SHA.
- Prefer concise, high-signal findings over exhaustive nitpicks.
- For P3 docs/chore commits, a lighter single-expert review is fine.
- **Hard deadline:** finish within ~${REVIEW_TIMEOUT}s. Incomplete but honest reviews are better than hangs.
- After writing files, print a one-line summary: \`REVIEW_DONE $short <verdict> $REVIEW_DIR/${short}-${slug}.md\`

# Working directory

$ROOT
EOF
}

extract_verdict() {
  local path="$1"
  [[ -f "$path" ]] || { echo ""; return 0; }
  if grep -q 'needs rework' "$path" 2>/dev/null; then
    echo "needs rework"
  elif grep -q 'approve with follow-ups' "$path" 2>/dev/null; then
    echo "approve with follow-ups"
  elif grep -Eqi '^\s*(\*\*)?approve' "$path" 2>/dev/null; then
    # avoid false positive from "approve" inside recommendations
    if awk 'BEGIN{v=""} /^## Verdict/{f=1;next} /^## /{f=0} f{print}' "$path" | grep -qi 'approve'; then
      if awk 'BEGIN{v=""} /^## Verdict/{f=1;next} /^## /{f=0} f{print}' "$path" | grep -qi 'needs rework'; then
        echo "needs rework"
      else
        echo "approve"
      fi
    else
      echo "reviewed"
    fi
  else
    echo "reviewed"
  fi
}

mark_plan_reviewed() {
  local short="$1" slug="$2" verdict="$3"
  local path rel
  path="$(review_path_for "$short" "$slug")"
  rel="docs/reviews/$(basename "$path")"

  python3 - "$PLAN" "$short" "$verdict" "$rel" <<'PY'
import re, sys
from pathlib import Path
plan_path, short, verdict, rel = sys.argv[1:5]
text = Path(plan_path).read_text()

# Update backlog table row
row_re = re.compile(
    r'(^\| \d+ \| `' + re.escape(short) + r'` \|.*\|) ⬜ Pending( \|)',
    re.M,
)
text2, n = row_re.subn(r'\1 ✅ Reviewed — **' + verdict + '**\2', text, count=1)

# Recount reviewed/pending from table
rows = re.findall(r'^\| \d+ \| `([0-9a-f]{7,})` \|.*?\| (✅ Reviewed[^|]*|⬜ Pending) \|', text2, re.M)
reviewed = sum(1 for _, s in rows if s.startswith('✅'))
pending = sum(1 for _, s in rows if s.startswith('⬜'))
total = len(rows)

# Update Status line
text2 = re.sub(
    r'^\*\*Status:\*\*.*$',
    f'**Status:** {pending}/{total} commits **pending** — {reviewed} reviewed. Latest: `{short}` → `{rel}` (**{verdict}**). Expected review artifact: `docs/reviews/<short-sha>-<slug>.md` per §5.7.',
    text2,
    count=1,
    flags=re.M,
)

# Update Last updated footer — append/replace review backlog sentence
footer_marker = '**Commit review process (§5.7):**'
if footer_marker in text2:
    # Replace from Commit review process through next unbroken sentence about backlog if present
    text2 = re.sub(
        r'\*\*Commit review process \(§5\.7\):\*\*.*?(?=Next \(post-v1\.0\.0\)|$)',
        f'**Commit review process (§5.7):** every commit reviewed via `kims-team` skill; reviews stored in `docs/reviews/<short-sha>-<slug>.md`. **Review backlog:** {reviewed}/{total} reviewed — latest `{short}` → `{rel}` (**{verdict}**); {pending} pending. ',
        text2,
        count=1,
        flags=re.S,
    )

Path(plan_path).write_text(text2)
print(f'PLAN updated: {short} -> {verdict} (reviewed={reviewed}, pending={pending})')
PY
}

run_review() {
  local short="$1" slug="$2" milestone="$3" experts="$4" subject="$5"
  local prompt log_file
  prompt="$(build_prompt "$short" "$slug" "$milestone" "$experts" "$subject")"
  log_file="$REVIEW_LOG_DIR/${short}-${slug}.log"

  local model_args=()
  if [[ -n "$REVIEW_MODEL" ]]; then
    model_args=(-m "$REVIEW_MODEL")
  fi

  echo "==> Reviewing $short ($milestone) — $subject"
  echo "    experts: $experts"
  echo "    log:     $log_file"
  echo "    timeout: ${REVIEW_TIMEOUT}s"

  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "    [dry-run] would run: timeout ${REVIEW_TIMEOUT} $OPENCODE_BIN run --auto --dir $ROOT ${model_args[*]:-} <prompt>"
    printf '%s\n' "$prompt" > "${log_file}.prompt"
    echo "    [dry-run] prompt written to ${log_file}.prompt"
    return 0
  fi

  set +e
  timeout --kill-after=20 "$REVIEW_TIMEOUT" \
    "$OPENCODE_BIN" run \
      --auto \
      --dir "$ROOT" \
      --title "commit-review-$short" \
      "${model_args[@]}" \
      "$prompt" \
      >"$log_file" 2>&1
  local rc=$?
  set -e

  # 124 = timeout; 137 = SIGKILL
  local timed_out=0
  if [[ $rc -eq 124 || $rc -eq 137 ]]; then
    timed_out=1
    echo "    WARN: timed out after ${REVIEW_TIMEOUT}s (rc=$rc) — see $log_file" >&2
  elif [[ $rc -ne 0 ]]; then
    echo "    ERROR: opencode run exited $rc — see $log_file" >&2
  fi

  # Accept if review file exists with a verdict section, even on timeout/error
  local rpath
  rpath="$(review_path_for "$short" "$slug")"
  if [[ ! -f "$rpath" ]]; then
    local found
    found="$(find "$REVIEW_DIR" -maxdepth 1 -type f -name "${short}-*.md" 2>/dev/null | head -1 || true)"
    if [[ -n "$found" ]]; then
      rpath="$found"
    fi
  fi

  if [[ -f "$rpath" ]] && grep -q '^## Verdict' "$rpath" 2>/dev/null; then
    local verdict
    verdict="$(extract_verdict "$rpath")"
    mark_plan_reviewed "$short" "$slug" "$verdict"
    if [[ "$timed_out" -eq 1 ]]; then
      echo "    OK (timeout but artifact present) — $rpath ($verdict)"
    else
      echo "    OK — $rpath ($verdict)"
    fi
  else
    echo "    ERROR: no usable review artifact under $REVIEW_DIR/${short}-*.md" >&2
    echo "    tail of log:" >&2
    tail -n 40 "$log_file" >&2 || true
    return 1
  fi

  if [[ "$DO_COMMIT" -eq 1 ]]; then
    local rel
    rel="$(realpath --relative-to="$ROOT" "$rpath")"
    if git status --porcelain -- "$rel" PLAN.md | grep -q .; then
      git add "$rel" PLAN.md
      git commit -m "docs: add team review for ${short}" -- "$rel" PLAN.md
      echo "    committed review for $short"
    else
      echo "    nothing to commit for $short"
    fi
  fi
}

pending=()
for row in "${BACKLOG[@]}"; do
  IFS='|' read -r short slug milestone experts priority subject <<<"$row"
  if [[ -n "$ONLY" ]]; then
    if [[ ",$ONLY," != *",$short,"* ]]; then
      continue
    fi
  fi
  if [[ "$FORCE" -eq 0 ]] && is_reviewed "$short" "$slug"; then
    continue
  fi
  pending+=("$row")
done

if [[ "$LIST_ONLY" -eq 1 ]]; then
  if [[ ${#pending[@]} -eq 0 ]]; then
    echo "No pending reviews."
  else
    echo "Pending reviews (${#pending[@]}):"
    for row in "${pending[@]}"; do
      IFS='|' read -r short slug milestone experts priority subject <<<"$row"
      printf '  [%s] %s  %s  (%s)\n' "$priority" "$short" "$subject" "$experts"
    done
  fi
  exit 0
fi

if [[ ${#pending[@]} -eq 0 ]]; then
  echo "Nothing to review — all backlog commits already have docs/reviews/ artifacts."
  exit 0
fi

echo "Review backlog: ${#pending[@]} commit(s) to process"
echo "Repo: $ROOT"
echo

failed=()
succeeded=()

for row in "${pending[@]}"; do
  IFS='|' read -r short slug milestone experts priority subject <<<"$row"
  if run_review "$short" "$slug" "$milestone" "$experts" "$subject"; then
    succeeded+=("$short")
  else
    failed+=("$short")
  fi
  echo
done

echo "========================================"
echo "Completed: ${#succeeded[@]}"
if [[ ${#succeeded[@]} -gt 0 ]]; then
  printf '  %s\n' "${succeeded[@]}"
fi
echo "Failed: ${#failed[@]}"
if [[ ${#failed[@]} -gt 0 ]]; then
  printf '  %s\n' "${failed[@]}"
  echo "Re-run failures with: scripts/review-commits.sh --only $(IFS=,; echo "${failed[*]}")"
  exit 1
fi

echo "All pending reviews finished. Check docs/reviews/ and PLAN.md §5.7."
