# Commit Review — P0 daily-driver trust sprint

- **Commit:** (uncommitted working tree after P0-SPRINT-PLAN.md)
- **Subject:** feat: P0 review-debt sprint — daily-driver trust
- **Date:** 2026-10-06
- **Experts engaged:** Backend Engineer, Full Stack Developer, DevOps, Security DevOps, QA Engineer
- **Skill mode:** Auto-Assignment (kims-team, pre-commit review of sprint diff)

## Diff summary

Implements PLAN.md §5.8.4 P0 board:

| ID | Change |
|----|--------|
| T01 | `pkg/duration.ParseRelativeDateEndIn` + end-side CLI parsers; `--end this-week` = Sunday EOD |
| T02 | `internal/output/fromStruct` handles single structs and pointers |
| T03 | Alias commands inherit flags via `AddFlagSet`; regression tests |
| T04 | `internal/update.Install` verifies sha256 vs release `checksums.txt` (fail closed) |
| T05 | `npm/install.js` uses `execFile` argv + checksum verify |
| T06 | `FromHTTPStatus`: 403→E003Permission; tests updated |
| T07 | Minimal pagination honesty: `ListProjects/Clients/Tags/TasksPaged` single-fetch; header PageInfo; resolvers paginate; omit page≤0 |
| T08 | `url.Values.Encode()` in query builders |
| T09 | HTTP client: body close outside retry loop; body rewind; Retry-After; debug/verbose → stderr |
| T10 | Split create-then-update + rollback on failure |
| T11 | ClearCache per-profile only under `cache/profiles/<name>` |
| T12 | configure merge-on-save |
| T13 | Load fail-loud on corrupt YAML |
| T15/T33 | `--if-not-exists` swallows only E002NotFound |
| T16 | Bulk requires end/duration; partial failure → E010 exit 11 |
| T29 | Hidden API-key input (`term.ReadPassword`) |
| T30 | Comma-ok on workspace id type assert |
| T34 | CI lint blocking; `-race`; `tools/coveragegate` core ≥80% |
| T37 | `errors.As` uses stdlib unwrap; `main.go` prints `Format()` to stderr |
| T40 | GoReleaser `brews.skip_upload: auto` |

Also: PLAN.md §5.8.4/§5.8.6 status updates; README exit-code honesty (no code 12; 403→E003).

## Findings

**Backend:** Relative ends are now inclusive period ends; start tokens unchanged. Query encoding uses `url.Values` — special chars safe. Pagination single-fetch removes double project-list call. Resolvers still load up to 500 items per page loop (acceptable for CLI scale).

**Full Stack:** Alias `AddFlagSet` copies flag pointers — works for parse/Get; tests cover `in --help` parse path. Bulk validation fails fast before writes when end/duration missing.

**DevOps:** Supply-chain gate closed for update path (fail closed without checksums). npm postinstall no longer shells out with interpolated paths. Brew publish optional via `skip_upload: auto` when token missing.

**Security:** Configure key input hidden on TTY. Debug logs masked key on stderr only. Checksum verify prevents silent binary swap.

**QA:** `go test -race ./...` green across all packages. Regression tests added for T01 end ranges, T02 single-struct, T03 alias flags, T04 checksums (happy/missing/mismatch), T06 403 mapping, T08 encoding, T34 coveragegate tool.

## Recommendations

1. Commit this tree with a feat/fix message matching house style; then run `kims-team` review on the actual commit SHA (§5.7) and file `docs/reviews/<short>-p0-daily-driver-trust.md`.
2. Public npm/brew/`update` advertising still gated on first GitHub Release (operational).
3. Product decisions (T14 hydrated, T18 list default, T28 status cache) remain parked — do not code around them.
4. T07 full resolver/search API work if Clockify exposes name search later.

## Conflicts / unresolved

- Product decisions parked (not defects).
- Live integration E2E needs credentials.
- `internal/resolve` package remains unused by CLI (production resolvers live in `internal/cli`); consider delete in P2 cleanup.

## Verdict

**approve with follow-ups** — P0 sprint meets DoD (`go test -race ./...` green; supply-chain fail-closed; CI gates enforce). Follow-ups: SHA commit + formal review artifact; live publish after remote exists; P1 remainder per §5.8.
