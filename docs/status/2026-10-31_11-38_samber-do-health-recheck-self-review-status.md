# Status + Brutal Self-Review — samber/do + Health Architecture Recheck Session

**Date:** 2026-10-31 11:38 CET
**Session scope:** the 2026-10-31 09:33 recheck of the samber/do v2 + health-checks architecture question (report: `docs/architecture-understanding/2026-10-31_09-33_samber-do-health-architecture-recheck.md`), plus this self-review of that run. Per instruction: no new research beyond what this session did and noticed.

**Session verdict recap:** samber/do = zero usage, still the correct call for an SDK; scores 4.0 / 4.5 / 4.5 / **3.5↑**; one open user-gated decision (F-1 code-half); zero new findings.

---

## a) FULLY DONE

1. **Premise re-audit** — samber/do absent from `go.mod` and all Go sources (re-verified with grep/rg this session, not recalled).
2. **Upstream re-verification, live** — `samber/do/v2` v2.1.0 still latest; health/shutdown interfaces unchanged; NEW evidence captured: the package's own docs say `HealthCheck` "should perform a quick check", plus `ErrHealthCheckTimeout` and parallel scope fan-out — sharpened the F-1 argument.
3. **Delta log 2026-10-16 → now** — surface 41 → 46 public methods; F-1 docs-half, BDD pin, README DI recipe, F-5 spec-drift resolution all confirmed shipped; v0.14.0/v0.15.0/v0.16.0 releases accounted.
4. **Architecture re-verification with exact anchors** — statelessness (0 goroutines/tickers/mutexes in non-test code), seams (`Doer`, `Logger`, `WithRetry`, `WithHTTPClient`), retry wiring (`MaxAttempts`, `DelayFunc`), `Health` chain + honest godoc, BDD `Describe("Health")`.
5. **New mitigation evidence** — go-retry v0.7.1's backoff delay is a ctx-aware `select` (retry.go:175-181) with cancel/timeout pinned upstream: consumers CAN bound `Health` with a context deadline; documented in report §4.2 with a gate-validated snippet.
6. **Four dimensions re-scored** with delta reasoning (self-healthy 3.0 → 3.5: honesty gap closed and regression-pinned; capability gap remains user-gated). Aggregate 4.1/5.
7. **Report shipped** — ~190 lines, series-aware (4th assessment), standalone verification appendix (11 command/result rows, all run this session).
8. **Prior report annotated** — superseded-by pointer added to the 2026-10-16 report; survival past the auto-git daemon verified (`grep -c` = 1).
9. **Gates green** — dprint (two rounds), `anchor-check` exit 0, `doc-verify` exit 0 (70 OK / 0 errors), `nix flake check` exit 0 including `md-go-snippets` validating the report's ```go fence.
10. **Series + harvest protocol honored** — prior reports read first and cited (no re-derivation); TODO_LIST harvest check: every roadmap item already tracked, nothing to harvest.
11. **Self-caught defect #1 fixed in-session** — "16 resource files" was wrong (16 = total files incl. client.go; 15 resource files); caught by the post-write self-consistency sweep, all 4 occurrences fixed.
12. **Self-caught defect #2 fixed by THIS review** — TL;DR said "the two releases since the last review"; it is **three** (v0.14.0 release train ran 22:33, after the 19:59 review). Fixed with the derivation note inline.
13. **Gate-at-HEAD gap closed** — full `nix flake check` re-run at the final post-fix state (initial run had executed before the last text edits).

## b) PARTIALLY DONE

1. **Verify-after-edit discipline** — cheap doc gates re-ran after every edit; the heavy gate initially ran BEFORE the final text edits (had to reason about whether a re-run was needed instead of knowing). Closed by re-running, but the ordering was wrong the first time.
2. **Memory / done-bar items (deliberate skips, documented not silent):**
   - AGENTS.md durable-outcome line SKIPPED — the samber/do verdict already lives in TODO_LIST.md:140, README.md:841, and both reports; a fourth copy would violate one-fact-one-place.
   - Series index — `docs/architecture-understanding/` has no README index (5 files); report reachability relies on the series pointer chain. Acceptable, but weak.
   - Browser screenshot (library-deep-dive done-bar) — N/A: report is `.md` by explicit user instruction, no HTML to render.
3. **Self-consistency sweep ran AFTER writing, not before** — it worked (caught defect #1) but the numbers should have been derived at write time; see defect #2 for the case where late nearly meant never.

## c) NOT STARTED (all pre-existing, correctly parked, none silently ignored)

1. **F-1 code-half** — additive `VerifyCredentials(ctx)` single-attempt probe (memo D2 option b, recommended for v1.1); BLOCKED: user design decision (TODO_LIST P3).
2. **`HealthCheck(ctx)` name-alias** — same decision row; only worth it with a real consumer asking.
3. **Cheap liveness probe** — needs an endpoint choice verified against the live API; demand-gated (no ping/health endpoint in the vendored spec).
4. **Service-client split (ADR 002)** — trigger (15+ resources) reached, sequencing deferred to v1.x by documented decision.
5. **AGENTS.md wording looseness (noticed this session, untouched)** — "46 endpoint methods across 17 resources" vs FEATURES' precise "46 methods / 43 carrying operations" and my file-derived "15 resource files + client.go": three different denominators (resource families vs methods vs files) that invite misreading. Gate-green (doc-verify checks the count claims), but the phrasing deserves one conscious harmonization pass.
6. **SDK-side ctx-abort test** — report §4.2's deadline-mitigation claim rests on upstream go-retry tests; no wise-go BDD test pins it locally.
7. **README ctx-timeout line** — the DI recipe (README.md:841) shows the adapter but not the `context.WithTimeout` bounding shown in the recheck report §4.2.

## d) TOTALLY FUCKED UP!

Nothing irreparable shipped; the honest ledger of this session's self-inflicted damage:

1. **Defect #1 — wrong denominator written from recall.** I read `rg -l` output showing 16 FILES (incl. client.go) and wrote "16 resource files" into the report. Briefly committed by the auto-git daemon before the fix landed. Root cause: derived the number from memory of my own tool output instead of re-deriving at write time.
2. **Defect #2 — wrong release count, lived ~2 hours.** "Two releases since the last review" contradicted my own §2 table listing three. NOT caught by any gate (free-text prose inside a table cell, in a non-swept directory) — caught only by this self-review's re-derivation. Same root cause as #1. Lesson: prose counts need the same derivation rigor as gated counts; nothing downstream would have flagged it.
3. **Edit-tool race ×2** — attempted `multiedit` immediately after external dprint formatting touched the file, without re-viewing. Two wasted round trips; the tool correctly refused.
4. **Gate ordering waste** — heaviest gate first, text edits after → one avoidable full re-run.
5. **No lies, but two inaccuracies DID ship momentarily** — stating that plainly rather than grading my own curve.

## e) WHAT WE SHOULD IMPROVE!

1. **Derive-then-write** — every count/denominator/date in a report gets pasted from a fresh command output at write time, never recalled. Both session bugs were recall errors; zero gate would catch the class in non-swept dirs.
2. **View-after-formatter** — after dprint/treefmt touches a file, ALWAYS `view` before editing. The tool enforces it; fighting it costs round trips.
3. **Heaviest gate LAST** — order: content edits → dprint → cheap gates (anchor/doc) → `nix flake check`. Guarantees the gate-at-HEAD claim instead of reasoning about staleness.
4. **End-of-session done-bar checklist** — run the skill done-bars (gate / index / memory / render-N/A) as an explicit list; this session did each consciously but ad hoc, and a checklist would have forced the release-count re-derivation.
5. **Stamp derived numbers** — report tables carrying counts could note their derivation command inline (the verification appendix does this for the appendix but §2 prose didn't benefit).
6. **Consider a series README** for `docs/architecture-understanding/` (5 files, zero index) — one file, kills the weak-reachability note in (b).

## f) Up to 50 things to get done next (session-derived, impact-sorted; brainstorm tier — ROADMAP/TODO routing via docs-health, not commitments)

| #  | Item                                                                                                                                       | Tier      |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------ | --------- |
| 1  | Decide F-1 code-half: adopt memo D2-(b) additive `VerifyCredentials(ctx)` for v1.1 (single attempt, same typed errors)                     | Decision  |
| 2  | If yes: BDD `Describe("VerifyCredentials")` mirroring Health's typed-error pin                                                             | Follow-up |
| 3  | If yes: update README DI recipe with "which probe when" guidance (compile-verified vs pinned tag AND master per repo rule)                 | Follow-up |
| 4  | If yes: record the outcome beside D2 in `docs/planning/2026-10-16_decision-memos.md` annotation style                                      | Follow-up |
| 5  | Add SDK-side BDD test: ctx deadline aborts `Health` mid-backoff (pins §4.2 locally, independent of decision #1)                            | Small     |
| 6  | Add the `context.WithTimeout` bounding snippet to README's DI section (exists only in the recheck report today)                            | Small     |
| 7  | Decide `HealthCheck(ctx)` alias: only if a real consumer surfaces; otherwise close the README naming note                                  | Decision  |
| 8  | Harmonize AGENTS.md's "46 endpoint methods across 17 resources" with FEATURES' precise denominators (methods vs operations vs files)       | Small     |
| 9  | Consider a one-line AGENTS.md note "samber/do deliberately absent; recheck 2026-10-31" ONLY if judged non-duplicative                      | Small     |
| 10 | Split the TODO_LIST F-1 row: strike the docs-half as done, keep code-half open (scannability per annotation convention)                    | Small     |
| 11 | Liveness probe research: verify a cheapest-endpoint candidate against the LIVE API before any code (demand-gated)                          | Gated     |
| 12 | Re-run this samber/do+health recheck on the next `samber/do` release that touches lifecycle/health surface (>v2.1.0)                       | Cadence   |
| 13 | Sequence the ADR-002 service-client split (trigger reached) — planning decision for the v1.x line                                          | Planning  |
| 14 | Add a README index to `docs/architecture-understanding/` (5-report series, zero index)                                                     | Small     |
| 15 | Consider godoc precision: replace "~4 attempts × 30s" in `Health` godoc with the exact default budget numbers                              | Small     |
| 16 | Add a compile-only `ExampleHealth` to `example_test.go` (existing nolint pattern) showing the ctx-deadline bound                           | Small     |
| 17 | Next release: refresh FEATURES/spec-conformance actual counts (doc-verify checks them anyway — zero extra gate work)                       | Routine   |
| 18 | Circuit breaker: keep demand-gated stance (ROADMAP.md:305); no action absent a consumer signal                                             | Gated     |
| 19 | Process rule (this repo's AGENTS or my habits): "prose counts in non-swept docs are UNGATED — derive, don't recall" — encode where durable | Process   |
| 20 | When editing any living doc next: re-verify its numbers the same session (anchor-check verifies anchors, NOT prose counts)                 | Process   |

Twenty honest items, not fifty — the session touched one report and two gates; padding to 50 would manufacture unrelated work, which the instruction explicitly forbade.

## g) Questions I can NOT figure out myself

1. **F-1 code-half (the standing block):** land `VerifyCredentials(ctx)` as memo D2-(b) recommends (additive, v1.1, no `Health` behavior change) — yes, no, or defer past v1.1? The architecture case is made; the surface-budget call is yours.
2. **Recheck cadence:** should the samber/do+health series re-run on a trigger (upstream release touching lifecycle/health, or ~quarterly), or stay on-demand only? (The seasonal-conventions entry in AGENTS.md already runs ~8-week external-claims sweeps for FEATURES/spec floors — should this series join that cadence or stay separate?)
3. **Test priority independent of #1:** want the SDK-side ctx-abort BDD test (item f#5) added NOW to pin the deadline mitigation locally, or is resting on upstream go-retry's pinned cancel/timeout tests acceptable?

---

_Point-in-time status + self-review, 2026-10-31 11:38 CET. Now waiting for instructions._
