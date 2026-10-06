# Commit Review — 83f61fc

- **Commit:** 83f61fc037f3ceeac8ed44e47981690f61738fda
- **Subject:** doc: create plan
- **Date:** 2026-09-21
- **Experts engaged:** Architect (light, in-session)
- **Skill mode:** Auto-Assignment (kims-team)
- **Review date:** 2026-10-05
- **Related plan:** PLAN.md §5.7 Review Backlog #2 (P3)

## Diff summary

Single-file commit: adds `PLAN.md` (557 lines, no product code). Introduces the full project roadmap for clockify-cli: Now/Next/Later horizons, four-dimension scoring model, milestone definitions M0–M10 with deliverables/success criteria/confidence, sequencing rationale, technical & project risks, key decisions/open questions, and an estimated timeline appendix (25–34 weeks solo). Tree at this commit also contains `clockify-cli-spec.md` (design spec) already present from earlier work; this commit does not modify it.

## Findings

### Critical

None.

### Major

None for a planning document. The plan is coherent on its own terms: critical path (M0→M1→M2→M3) matches user value, risks are concrete (API quirks, timer races, keychain headless, reports host), and open questions are explicitly deferred with resolution approaches rather than left vague.

### Minor

1. **Scoring table rank labels** — `PLAN.md` §3.1: M0 and M1 both score **4.30** (correctly shown as rank 1 tied), but M6 is labeled **“3 (tied)”** at 3.55 with no other milestone at that score; M2 is “4” at 3.45. Cosmetic inconsistency in the ranking column only; the weighted scores themselves are consistent with the formula in §1.2.
2. **Time-horizon vs. full-plan duration mismatch** — §1.1 maps horizons to Q1=weeks 1–8, Q2=9–16, Q3=17–24, but Appendix §6 estimates 25–34 weeks solo (M7 alone runs through ~week 25; M9/M10 beyond that). Horizons are incomplete vs. the stated total effort; either extend “Later” or note that post-M3 work spans a fourth quarter.
3. **Milestone duration overlap in appendix** — M3 (weeks 10–13) and M4 (12–16) / M6 (15–20) are presented as cumulative ranges that partially overlap while §3.2 diagrams them as sequential/parallel-after-M2. Acceptable for rough planning, but the appendix table can be read as contradicting the sequencing diagram.
4. **Duplicate success criteria (startup &lt;50ms, size &lt;20MB)** — appear in both M0 and M9. Intentional re-verification is fine; a one-line note that M9 measures the release binary would remove ambiguity.
5. **Commit body is empty** — subject `doc: create plan` and Signed-off-by are fine for a docs-only commit; no body linking to the spec or §5.7 process (process/backlog was added later in current PLAN.md — noted without scoring).

## Recommendations

1. Fix the §3.1 rank labels (remove the spurious “3 (tied)” on M6).
2. Reconcile §1.1 quarterly horizons with Appendix §6 total effort (extend Later or re-cut Now/Next/Later).
3. Clarify whether M4/M5/M6 timelines are parallel to M3 or sequential after it, and align appendix ranges with §3.2.
4. Optionally annotate M9 as the release measurement gate for size/startup.
5. Later milestones already implemented many plan assumptions (Cobra, keyring, E00X errors, reports); the plan’s own closing note to re-review after M0 is the right control — honor that decision point when capacity is available.

## Conflicts / unresolved

None between experts — single-expert light review per PLAN §5.7 P3 guidance (docs/chore). No multi-expert disagreement to resolve.

## Verdict

approve with follow-ups

Strong initial roadmap: sound prioritization, explicit risks, and a credible M0–M10 sequence aligned to the design spec. Acceptable as the project plan for M0; minor internal inconsistencies in scoring labels and timeline horizons should be cleaned up in a follow-up docs pass rather than blocking work.

## HEAD triage (2026-10-05)

- **Current HEAD:** `ac64914`
- **Original verdict:** unchanged
- **Legend:** ✅ completed at HEAD · ⚠️ partially fixed · ❌ still open · ➖ N/A or cannot verify

| # | Item | Status | Evidence |
|---|------|--------|----------|
| F1 | Scoring table rank labels (Finding 1) — M6 “3 (tied)” at 3.55 with no tie | ❌ | `PLAN.md` §3.1 line 684 still shows `**3.55** \| 3 (tied)`; no other milestone at that score. Related label defect: M10 score **2.15** labeled rank 10 vs M9 score **1.95** labeled rank 9 (lines 687–688) — ranks inverted relative to scores. |
| F2 | Time-horizon vs full-plan duration mismatch (Finding 2) | ⚠️ | §1.1 lines 15–19 unchanged (Later still Q3/weeks 17–24). The appendix “25–34 weeks” total that conflicted with the week-24 cap is **gone** — Appendix §6.1 (lines 964–979) is now Original Est. vs Actual per milestone with no cumulative week total. Original estimates for M7–M10 (4–5 + 2–3 + 2–3 + 2–3 weeks) still exceed the Later window if read sequentially; no Q4/extension note added to §1.1. |
| F3 | Milestone duration overlap in appendix (Finding 3) | ✅ | Overlapping cumulative week ranges (M3 10–13 vs M4 12–16 / M6 15–20) removed. §6.1 presents only per-milestone Original Est. vs Actual; nothing left to contradict §3.2’s parallel-after-M3 diagram (lines 705–708). Residual mention “Was planned weeks 10–13” (line 1008) is a single historical note, not a table contradiction. |
| F4 | Duplicate success criteria startup &lt;50ms / size &lt;20MB in M0 and M9 (Finding 4) | ❌ | Still duplicated: M0 lines 78–79 (measured 5ms / 12MB); M9 line 555 (measured ~3.9MB / ~10ms); M10 lines 596–597 (~11.5MB / warm). No note anywhere in `PLAN.md` that M9/M10 measure the **release** binary; grep for “release binary” in PLAN.md hits only the npm wrapper deliverable (line 531). |
| F5 | Commit body empty (Finding 5) | ➖ | Historical commit metadata for `83f61fc`; not fixable in current `PLAN.md`/docs. N/A to HEAD triage of plan-doc issues. |
| R1 | Recommendation 1 — fix §3.1 rank labels | ❌ | Same evidence as F1; §3.1 rank column untouched at HEAD. |
| R2 | Recommendation 2 — reconcile §1.1 horizons with Appendix §6 total effort | ⚠️ | Same as F2: total effort removed from the appendix rather than horizons extended or Now/Next/Later re-cut; §1.1 still ends at Q3/week 24. |
| R3 | Recommendation 3 — clarify M4/M5/M6 parallel vs sequential; align appendix with §3.2 | ✅ | Contradicting appendix week ranges removed (see F3); §3.2 remains the authoritative sequencing view (M4→M5 and M6 branch in parallel after M3). Explicit prose clarification was not added, but the reported reading contradiction no longer exists. |
| R4 | Recommendation 4 — optionally annotate M9 as release measurement gate | ❌ | Optional item not done; M9/M10 success criteria (lines 555, 596–597) still lack a release-binary note. |
| R5 | Recommendation 5 — honor post-M0/M1 re-review decision point | ✅ | Closing notes lines 1014–1016 record the post-M1 review outcomes (estimates revised after M0–M5 actuals; batch/versioning open questions resolved). Decision point was exercised. |
| N1 | (new) §3.1 rank inversion M9 vs M10 | ❌ | Line 687 M9 **1.95** → rank 9; line 688 M10 **2.15** → rank 10. Scores imply M10 ahead of M9. |
| N2 | (new) M10 section missing `### M10:` heading | ❌ | Lines 577–581: after M9’s `---`, the section begins `**Status:** ✅ **COMPLETE**` with no `### M10: Quality, Testing & Documentation` title. Grep `^### M` in PLAN.md lists M0–M9 and M11 only. |
| N3 | (new) Duplicate conflicting Confidence lines in M10 | ❌ | Line 607 “Confidence: High (85%)” vs line 617 “Confidence: Medium-High (85%)” in the same M10 section. |
| N4 | (new) Closing note contradicts §3.6 report-host correction | ❌ | Line 1016: “report API host confirmed via spec §3.6” vs M6 notes (line 369) and §5.1 (line 812): spec §3.6 `reports.<host>` pattern is **incorrect** for regional workspaces; regional = main host + `/report/v1`. |
| N5 | (new) Stale open-question language re: rate limits / M6 spike | ❌ | §5.2 line 821: “document if discovered during M6 spike” — M6 is complete (2026-10-03); question still marked 🔲 partially resolved. |

### Summary

Original verdict **approve with follow-ups** is unchanged. Of 10 original review items (5 findings + 5 recommendations): **3 ✅**, **2 ⚠️**, **4 ❌**, **1 ➖**. The Appendix restructure resolved the appendix-vs-sequencing contradiction (F3/R3) and removed the 25–34 week total (softening F2/R2), but scoring labels, horizon coverage, and the optional M9 release-binary annotation remain open. Five additional internal `PLAN.md` inconsistencies were observed at HEAD (N1–N5), notably the missing M10 heading and the closing note’s stale §3.6 claim.

### Completed at HEAD

- F3 — appendix no longer presents overlapping cumulative week ranges that contradicted §3.2.
- R3 — §6.1 aligned with §3.2 by removal; parallel M4/M5 and M6-after-M3 sequencing stands unchallenged.
- R5 — post-M1 review decision point documented in closing notes (lines 1014–1016).

### Still open at HEAD

- F1 / R1 / N1 — §3.1 rank labels: spurious “3 (tied)” on M6; new M9/M10 rank inversion.
- F2 / R2 — §1.1 horizons still capped at Q3/week 24 with no extension note; appendix no longer totals effort (mismatch softened, not reconciled).
- F4 / R4 — startup &lt;50ms / size &lt;20MB duplicated across M0/M9/M10 with no release-binary annotation.
- N2 — M10 heading missing.
- N3 — conflicting Confidence lines in M10.
- N4 — closing note “confirmed via spec §3.6” contradicts the plan’s own §3.6 correction.
- N5 — §5.2 rate-limit question still references post-M6 spike documentation.
- F5 — commit body (historical; not actionable at HEAD).
