# Status Report — 2026-10-31 12:57 CEST: SUPERB Plan Round-2 Execution (Full Backlog Sweep)

Session charter: the user's blanket directive — execute the ENTIRE
`2026-10-31_11-49` SUPERB Pareto plan (27 milestones), keep going until
everything works. Execution log per milestone: the plan file's §7.
This report records what shipped, what is held for the owner, and the
verification evidence.

## a. Headline

- **27 of 27 milestones dispositioned**: 15 DONE with code/docs, 5 DONE as
  decisions/annotations, 4 executed to their no-push/no-credential limit,
  3 not due (calendar/perf/external gates). Nothing silently skipped.
- **One new public API method**: `VerifyCredentials(ctx)` (47 methods now) —
  the D2-b fail-fast credential probe, BDD-pinned.
- **Quarterly surfaces canonicalized 2026Q4 → 2027Q1** (both families,
  credentialed-probe-verified) and the spec snapshot re-vendored to the live
  2027Q1 edition (188 paths / 228 operations, zero drift, gitleaks clean).
- **Two new gates**: hermetic `checks.dprint-format` (closes the
  daemon-commit format gap inside `nix flake check`) and heading-pointer
  support in the anchor gate.
- **CI is ENABLED** (first run fires on the next code push; weekly
  spec-drift job wired per D3-b).
- **BuildFlow binary rebuilt** at repo HEAD `74a9212`; fresh full run this
  session (verdict: §f).

## b. Code changes (all tested)

| Change                                  | Evidence                                                                                                        |
| --------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| `VerifyCredentials(ctx)` in `client.go` | shallow-copy of the stateless client with `MaxAttempts = 1`; same typed errors as `Health`; `go test -race ./...` green |
| 5 new BDD contexts                      | `wise_test.go`: VerifyCredentials (valid/401/429-once/500-once/deadline) + Health deadline-abort (mid-backoff, mid-Retry-After) |
| 2 godoc examples                        | `ExampleClient_Health` (ctx-bounded), `ExampleClient_VerifyCredentials` — 38 total, gate-checked                 |
| Quarterly constants flip                | `ottAPIVersion`/`webhookSubscriptionsAPIVersion` → `"2027Q1"`; `TestQuarterlyAPIVersionsConsistent` drove every literal site |
| `checks.dprint-format`                  | 4 fetchurl-pinned plugin WASMs, store-path plugins, throwaway cache, negative probe self-fails on no-op           |
| `scripts/anchor_check.py`               | `##`+ headings join the title registry (generic ones skipped); 2 new self-test classes; census 0 findings         |
| ci.yml                                  | `workflow_dispatch` + weekly `spec-drift` scheduled job (`--fail-on-drift`, report-only posture per D3-b)         |

## c. Documents

- **AGENTS.md part-3 full rewrite**: 56,812 → 40,634 bytes, 69/69 bold titles
  preserved (pointer registry intact), every removed clause verified
  greppable at its evidence owner. The <30 KB target is **retired as
  unreachable-without-loss**: the remaining largest bullets (~1.0–1.2 KB
  each) are normative rule density, not narrative.
- **Denominator convention harmonized**: 47 public `*Client` methods / 46
  Wise operations shipped / 15 resource files; "17 resource families"
  survives only as ROADMAP's enumerated list. doc-verify FEATURES pattern
  updated in the same change (methods ≠ operations since VerifyCredentials).
- Series index: `docs/architecture-understanding/README.md` (5 artifacts,
  supersession chain, cadence trigger).
- New memos: typed-Details spike (keep the map; type the decision), 
  RetryAfterDelay freeze (entry-as-is at v1.x).
- Decision memos D2–D6 annotated with resolutions; plan file §7 log appended.

## d. Decisions applied as standing defaults (reversible; flip on the word)

D2-b VerifyCredentials additive (HealthCheck alias declined) · D4 keep rev
pins · D5 drop x86_64-darwin · D6 `BUILDFLOW_NO_RESULT_CACHE=1` standing ·
D3-b pre-release report-only + weekly fail-on-drift job · recheck cadence =
seasonal join + lifecycle trigger · go-directive = policy wins until 1.27 ·
vulnix = no replacement (premise healed) · AGENTS = single-file taxonomy
retained.

## e. Upstream

- **Filed**: [LarsArtmann/crush-config#3](https://github.com/LarsArtmann/crush-config/issues/3) —
  propagate the hermetic dprint-check pattern into the report-producing
  skill flows (verify-before-filing pass; github-voice check pass; draft in
  `docs/drafts/`).
- **Not filed (premise healed)**: nixpkgs/vulnix NVD-404 — NIST serves the
  archive again (HEAD 200 verified 2026-10-31); vulnix completes; an
  on-demand scan showed 22 nixpkgs-closure advisories, 0 SDK-actionable.

## f. Gates

| Gate                                          | Result |
| --------------------------------------------- | ------ |
| `go test -race ./...`                         | ok     |
| `golangci-lint run`                           | 0 issues |
| `nix run .#anchor-check`                      | 101 anchors + 6 quoted pointers OK |
| `nix run .#doc-verify`                        | all checks passed |
| `nix run .#spec-drift -- --fail-on-drift`     | exit 0 (0 drift) |
| gitleaks (full tree)                          | no leaks |
| `nix build .#checks.x86_64-linux.dprint-format` | ok (probe self-test included) |
| buildflow full run (fresh binary `74a9212`)   | BUILDFLOW-FULL-VERDICT |
| count-inflation reconcile (M18)               | M18-VERDICT |

## g. Held for the owner (the only things that could not be executed)

1. **v1.0.0 tag** (M4) — the one-word irreversible call; pre-release chain
   ready.
2. **Push** (wise-go session commits; bank-sync v0.16.0 bump + AGENTS/15
   recipe note) — never-push rule; CI's first run activates on the wise-go
   code push.
3. **Secrets** (CACHIX_AUTH_TOKEN, ERRAUDIT_TOKEN) — then delete the two
   `continue-on-error` gate lines.
4. **Sandbox browser SCA attempt or banking the verdicts** (M9) — no browser
   in this environment.
5. **Daemon location** (M8 observability; cadence measured at ~3-min
   dirty-tree polling).

TODO_LIST after this session: 17 struck rows, 13 open — every open row is
owner-gated, calendar-gated, or externally conditioned.
