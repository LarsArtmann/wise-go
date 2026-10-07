# Status Report — dual-phase repair & tool re-evaluation (wise-go)

> **Superseded (2026-10-07 docs-health pass):** this is the current state for the buildflow repair thread. The gate is green on `go 1.26`; the three upstream BuildFlow defects are documented in AGENTS.md and routed to TODO_LIST.md. Retained (not archived) because its §b/§c/§f items remain open.

- **Date:** 2026-10-07 04:20 CEST
- **Scope:** This session only. Phase 1 (03:17–03:45): buildflow exit 69 → green repair. Phase 2 (03:50–04:15): operator challenged three judgment calls (upgrades, erraudit, max_file_size); all three re-verified with evidence. No unrelated research.
- **Outcome:** buildflow `--fix --build-mode=full` on the freshly installed binary **202b114** → **exit 0**, measured correctly. Three upstream BuildFlow defects discovered and evidenced. One self-caught analytical error (documented honestly in d).

---

## Executive summary

Phase 1 repaired a repo broken by its own guardrails: a daemon sweep deleted `.buildflow.yml` (the skip_steps file preventing the documented go 1.26↔1.27 flipflop), bumped go.mod to 1.27, and rewrote `errors_test.go` to Go 1.27-only syntax. All restored from git history; two real defects fixed on the way (invalid `exhaustruct_v5.exclude` key, dead SECURITY.md link); gate green.

Phase 2 answered "why do you hate upgrades / erraudit / the default max_file_size?" with evidence instead of argument:

1. **samber/lo**: full census of all 9 flagged sites (was a 2-site sample) — 6 idiomatic keeps, 3 marginal fits, 0 decisive; fleet adoption 34/322 repos. Decline stands, now complete.
2. **erraudit**: rebuilt and reinstalled buildflow (3bb229e → 202b114, 160 commits), retested — the **identical 29 findings** persist at HEAD. They are permanent upstream false positives (`_, ok := errors.AsType[T](err)` misreads; cleanup `_ = Body.Close()`; context_loss message demands) vs the canonical standalone gate that is clean on the same tree. Skip reinstated with double-binary evidence. En route I made, published, and retracted a wrong "fixed upstream" conclusion (see d1).
3. **max_file_size**: it is a lint threshold in lines (default 350), not a scan cap. The new file-size-check tool scans **0 files** in this repo at both 350 and 700 — empirically inert (upstream defect). 700 kept and now documented in-repo.

Three new upstream BuildFlow bugs are evidenced and queued: (a) embedded erraudit false positives at HEAD; (b) `nix run .#reinstall` does not switch the nix profile (stays locked at the old rev); (c) file-size-check scans 0 files in root-package Go repos.

---

## Timeline

| Time (CEST) | Event                                                                                                                                                                           |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 03:05–03:13 | Daemon sweep (3 commits): `.buildflow.yml` deleted, go.mod → 1.27, `errors_test.go` → Go 1.27-only promoted-field literals                                                      |
| 03:17       | Operator's buildflow run: 3 step failures, invalid golangci config, exit 69                                                                                                     |
| 03:18–03:45 | Phase 1: root cause, restorations, config fixes, vendorHash verified fresh (nix build green), gate green (exit 0, but exit code was mis-measured — see d3)                      |
| 03:45       | Status report v1 written                                                                                                                                                        |
| 03:50       | Operator challenge: "why do you hate upgrades / erraudit / the default max_file_size?"                                                                                          |
| 03:52       | BuildFlow rebuild attempt #1 fails — concurrent session mid-refactor at BuildFlow HEAD (gomod_tidy_pregate: undefined `runner`, signature mismatch)                             |
| 03:58       | Attempt #2 at newer HEAD 0e9f5e301 compiles; `nix run .#reinstall` prints REINSTALL-OK — **but does not switch the nix profile**                                                |
| 04:00       | lo census: all 9 sites viewed and judged; fleet grep (34/322); AGENTS.md updated                                                                                                |
| 04:05       | **Error**: single-step erraudit probe read as "0 findings" (jq `null \| length` = 0 on a config-skipped step) → skip lifted, AGENTS.md rewritten to "fixed upstream"            |
| 04:08       | Full run contradicts it: footer says `buildflow 3bb229e`, gate fires with the same 29; `BUILDFLOW-EXIT=0` echo proven a lie (`PIPESTATUS` unsupported → captured `tail`'s exit) |
| 04:10       | Diagnosed both failures; profile switched manually to the 202b114 store path; genuine retest: **29/29 findings identical** at the fresh binary                                  |
| 04:12       | Skip reinstated with corrected rationale; AGENTS.md corrected again (incl. the jq-null trap); status report v1 annotated with addendum                                          |
| 04:15       | Final full run on 202b114: **exit 0**, measured with plain `$?`                                                                                                                 |

---

## a) FULLY DONE

| # | Item                                                                                                                                                                                                                                                                                                                                                                                           | Verification                                                                                                                                                     |
| - | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **Phase 1 repair complete** (go.mod `go 1.26.0`, `.buildflow.yml` restored, `errors_test.go` reverted to 1.26-compatible, `exhaustruct_v5.ignore-patterns`, SECURITY.md → bugcrowd.com/wise, vendorHash verified fresh)                                                                                                                                                                        | carried from report v1; `go vet`/`go test`/`nix build`/`golangci config verify` all green; survives repeated full buildflow runs incl. go-mod-update + normalize |
| 2 | **samber/lo re-ruled on a full census** — all 9 sites individually reviewed (currencies.go:30, errors.go:192, ott.go:142/346/421/434, quotes.go:500, recipients.go:195, transfer_requirements.go:241): 6 keeps (incl. ott.go:346 prepend-cons the detector over-matches; recipients.go:195 keyed map transform lo.Map can't express), 3 marginal `lo.Map` fits, 0 decisive; fleet 34/322 repos | AGENTS.md bullet rewritten with census + fleet number; decision closed                                                                                           |
| 3 | **Buildflow binary refreshed 3bb229e → 202b114** (BuildFlow HEAD 0e9f5e301, tree was clean at build time)                                                                                                                                                                                                                                                                                      | `buildflow --version` = 202b114; full-run footer shows 202b114; `nix profile` manually repointed after the reinstall app failed to                               |
| 4 | **Embedded erraudit verdict settled with double-binary evidence** — identical 29 findings (8 `[ignored]` + 21 `[context_loss]`; same file:line, same rules) at 3bb229e AND 202b114; canonical standalone gate exit-0 clean on the identical tree; skip restored with rationale + the jq-null trap documented inline                                                                            | finding-set diff extracted and compared; `.buildflow.yml` skip carries the full evidence trail                                                                   |
| 5 | **max_file_size settled with data** — semantics researched in BuildFlow source (lines; file-size-check lint; default 350); tool empirically inert here at BOTH 350 and 700 (`filesScanned: 0`); 700 kept with documented rationale (when the scan works: 350 ⇒ 11+ flags incl. wise_test.go 4229; 700 ⇒ the 3 true outliers)                                                                   | both thresholds tested live and reverted; comment now lives in `.buildflow.yml`                                                                                  |
| 6 | **Final gate green on the fresh binary** — full `--build-mode=full` run: exit 0, erraudit skipped-via-config (as designed), findings gate clear                                                                                                                                                                                                                                                | exit code captured with plain `$?` this time; run log retained (/tmp/bf-final.log)                                                                               |
| 7 | **Memory corrected and hardened** — AGENTS.md erraudit entry rewritten twice, ending at the verified truth incl. both traps (binary freshness; `findings: null` vs `.summary.total`); report v1 annotated with an addendum rather than rewritten                                                                                                                                               | daemon-swept; current head state is the corrected one                                                                                                            |

## b) PARTIALLY DONE

| # | Item                                                        | State                                                                                                                                                                                    | Gap                                                                                                                                                                       |
| - | ----------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **Upstream BuildFlow bug reports**                          | Evidence fully prepared for three defects (erraudit false positives at HEAD incl. client.go:135–155/189/394 + canonical-gate comparison; reinstall profile-lock; file-size-check 0-scan) | Nothing filed — BuildFlow repo had a concurrent session mid-refactor tonight; filing is a decision for the operator (see g2)                                              |
| 2 | **Harvest of both reports' (f) sections into TODO_LIST.md** | Report v1 got an addendum; this report supersedes it                                                                                                                                     | TODO_LIST.md still carries the concurrent release-notes session's uncommitted edits at report time — touching it risks a clobber; harvest remains blocked on that landing |
| 3 | **Lychee verdict on the new SECURITY.md URL**               | Fixed and manually verified 200                                                                                                                                                          | Never isolated inside a buildflow run (`buildflow -s lychee --format finding`) — still owed one clean empirical confirmation                                              |
| 4 | **Green-run warning inventory**                             | Known classes: vulnix (20, nixpkgs-level), go-auto-upgrade (9 jsonv1tov2 refusals + 12 lo suggestions, documented), lychee-private-links (fleet policy), 9 unavailable tools (unnamed)   | `buildflow doctor --verbose` never run; the 9 tools are still a dark number; no one-shot per-tool finding sweep was captured                                              |
| 5 | **gopls hygiene**                                           | LSP restarted mid-phase-1; `go vet` authoritative green                                                                                                                                  | The stale 1.27 `UnsupportedFeature` diagnostics were never re-confirmed visually cleared                                                                                  |

## c) NOT STARTED (observed, deliberately untouched)

1. Vulnix posture: 20 nix-store toolchain advisories (binutils/bison/coreutils/gcc; several high) — needs a nixpkgs bump decision, not project code.
2. lychee private-links fleet policy (GITHUB_TOKEN vs exclude) — recorded undecided fleet-wide.
3. CI re-enable on GitHub (pre-existing documented path; tonight's green local gates satisfy a precondition).
4. Release coordination: whether this session's user-visible fix (SECURITY.md link) belongs in the concurrent v0.11.0 release notes (see g3).
5. Go 1.27 toolchain endgame (see g1) — every occurrence was reverted per policy; the policy question itself was escalated, not decided.
6. OTT 2026Q4 rollover probe and sandbox-key pass — pre-existing documented TODO items, untouched (out of session scope).

## d) TOTALLY FUCKED UP

1. **My own false "verified clean" conclusion (worst of the session, self-caught):** after reinstalling, I ran `buildflow -s erraudit --format finding`, jq printed `cannot iterate over: null` — a config-skipped step's `findings: null` — and `null | length` = 0, which I reported as "COUNT=0 … the fresh binary reports ZERO findings." On that misread I **lifted the skip and rewrote AGENTS.md to claim the false positives were fixed upstream**. It was wrong within minutes: the full run's footer (`buildflow 3bb229e`) and gate output (same 29) contradicted it. Three compounding mistakes: trusted `null | length` over the visible jq error; never checked `.summary.total`; and confirmation bias — I wanted the fresh binary to vindicate the rebuild. Caught by cross-checking, corrected within the hour; the wrong version sat in the tree (and was daemon-committed) for roughly 20 minutes, so history contains a wrong memory entry later corrected — accepted, not rewritten.
2. **A false verification claim in my own narration:** `BUILDFLOW-EXIT=0` was printed while the run had actually failed (exit 1) — `PIPESTATUS` is not supported by this shell, so the fallback silently captured `tail`'s exit code. I reported an exit-0 success for a failing run in chat. Fixed by re-measuring with a temp file + plain `$?`; the final green exit is genuine. Lesson recorded: verify exit codes with the simplest mechanism available, and never let a formatter command stand between the tool and `$?`.
3. **Premature config mutation on unverified evidence:** the skip removal and the AGENTS.md "fixed upstream" rewrite preceded the retest that should have gated them. Test-first-then-edit would have made the wrong conclusion impossible to commit (literally).
4. **The starting state of phase 1 remains the baseline fucked-up:** the repo's own automation (daemon + buildflow step) deleted the file that prevented the exact corruption shipped in the same commit, and silently migrated code to a language version the pinned toolchain cannot compile. Nothing in the commit messages signaled either.

## e) WHAT WE SHOULD IMPROVE (self-critique)

1. **Verify the instrument before trusting the reading.** Both phase-2 mistakes were measurement failures, not analysis failures: a jq idiom that maps "skipped" to "clean", and an exit-code echo that didn't measure what it claimed. The census, the file-size tests, and the finding-set diff were all solid — the two wrong turns came from unverified tooling behavior. New standing habit: every "green" claim needs the number from the authoritative field (`.summary.total`, raw `$?`), not a derived expression.
2. **Sequence edits after evidence, never before.** Lifting the skip before the retest converted a private misread into a committed config + memory change. One-command discipline: run the check, read the authoritative field, then edit.
3. **Reconcile contradictions immediately, not after the next run.** The footer version (`3bb229e`) contradicted the reinstall claim (`202b114`) in the same output I was celebrating. I should treat any internal inconsistency in tool output as a stop-and-diagnose signal, not noise to summarize past.
4. **Check for concurrent activity before building/basing on a foreign repo.** BuildFlow attempt #1 raced another session's mid-refactor (build failed on their half-landed code). Checking HEAD movement across two reads would have predicted it. No damage done, one wasted build.
5. **The erraudit split-brain should have been my first hypothesis in phase 1**, not something reached after running both binaries — AGENTS.md documented the CI-only canonical gate; the divergence class was pre-documented.
6. **Memory hygiene under churn:** AGENTS.md's erraudit entry changed direction twice within an hour. The rule "update memory immediately" collided with "verify before encoding" — verification must win; interim states belong in the report, not the memory file.
7. **Carried from report v1 (still true):** itemize green-run warnings once; run `doctor --verbose` for the 9 unavailable tools; isolate lychee's verdict on the new link; re-confirm gopls state after restarts.

## f) NEXT TASKS (brainstorm — up to 50, impact-ordered; NOT a commitment list)

**P0 — lock in tonight's state**

1. Harvest both status reports' (b)/(c)/(f) into TODO_LIST.md once the concurrent release-notes edit lands (docs-health HARVEST).
2. File upstream: embedded erraudit false positives at HEAD (evidence: client.go:135–155/189/394 finding set, identical at 3bb229e + 202b114; canonical standalone gate clean on identical tree) — pending g2.
3. File upstream: `nix run .#reinstall` leaves the nix profile locked at the old rev (profile entry `rev=3bb229e` untouched after REINSTALL-OK; manual `nix profile remove/add` of the store path required).
4. File upstream: file-size-check scans 0 files in root-package Go repos (0 findings at 350 AND 700; `filesScanned: 0` in summary; affects every repo with root .go files).
5. File upstream (minor): `-s <skipped-tool> --format finding` emits `findings: null` with no skipped marker — jq `null | length` masquerades as 0 findings; emit `.summary.total=0` + `skipped: true` instead.
6. Re-run `buildflow doctor`: binary-freshness check should now be OK on 202b114; enumerate the 9 unavailable tools by name while there.
7. Verify gopls diagnostics are clean post-restart (the six stale 1.27 `UnsupportedFeature` errors should be gone).
8. Isolate lychee's verdict on `bugcrowd.com/wise` (`buildflow -s lychee --format finding`).
9. One-shot sweep: `-s <tool> --format finding` per remaining warning-class tool (vulnix, go-auto-upgrade, lychee) → itemized log so "exit 0 with warnings" has a finite known list.
10. Watch one more full run for result-cache behavior (two consecutive runs tonight showed 0% hit rate — if it persists at unchanged trees, that's another upstream signal).

**P1 — prevent recurrence**
11. Fleet sweep for tonight's failure signature: `git log --diff-filter=D -- .buildflow.yml` across ~/projects (if one sweep deleted it here, siblings may share it).
12. Root-cause WHO deleted `.buildflow.yml` (buildflow step vs daemon sweep vs concurrent session) — still unattributed.
13. Upstream guard request: buildflow refuses to run (or warns loudly) when its own config file is absent in a repo whose AGENTS.md documents one.
14. Guard request #2: a preflight/CI check that fails when `grep '^go ' go.mod` ≠ the repo's documented pin (tonight's damage was silent for 7 minutes).
15. Ask upstream about daemon commit message hygiene: "auto-commit 1 changed file(s)" bundled config deletion + go.mod corruption with zero signal.
16. Consider anchoring the `exhaustruct_v5.ignore-patterns` entry (`\.Cmd$`) if v5 pattern semantics differ from v3 — currently untestable either way (no `exec.Cmd{` literals exist).
17. Re-check the max_file_size threshold (and the 3 violating files) after upstream fixes file-size-check's scanning — the 700-vs-350 rationale is written for that day.
18. Confirm `nix-hash-fix` clears now that go.mod is sane (preflight had it at 10/10 failures during the broken window).

**P2 — release coordination (concurrent session active)**
19. Decide whether the SECURITY.md link fix is release-notes-worthy or infra-only (g3).
20. After the release session lands: `nix flake check` (links fileset still covers SECURITY.md) + `nix run .#doc-verify` (count-claims freshness).
21. Verify the release-notes session's TODO_LIST closure doesn't conflict with harvest items (no double-bookkeeping).
22. Confirm changelog-only-commit gotcha doesn't bite their release-notes commit (dprint excludes CHANGELOG.md).

**P3 — toolchain endgame (needs operator, g1)**
23. Decide the Go 1.27 plan: conditions + date for flake toolchain + GOTOOLCHAIN policy change; the modernizer's pushed changes become the migration's first commit.
24. If 1.26 stays: fleet-wide decision to skip-list the 1.27-pushing dispositions rather than revert per incident.
25. Track nixpkgs go_1_26 patches; bump flake input when security patches land (pairs with P0 vulnix posture).
26. erraudit local install becomes possible at 1.27 (requires go ≥ 1.27) — if the toolchain moves, revisit the CI-only constraint.

**P4 — upstream BuildFlow quality (beyond the three filed bugs)**
27. Reconcile embedded-erraudit vs standalone gate permanently: flag-compatible flags, or per-repo-class documentation of which gate is authoritative.
28. Note the summary-count quirk: erraudit reported `filesScanned: 0` while `filesAffected: 9` — field semantics are inconsistent in the finding JSON.
29. Check whether the embedded analyzer's `context_loss` rule has an exclusion mechanism the repo could use in lieu of skip (preferred if it exists — narrower than a step skip).
30. `gomod_tidy_pregate` (the file the concurrent session was refactoring) — after it lands, review whether it interacts with wise-go's go.mod pin the way go-version-auto-configure does.
31. BuildFlow repo: review `git log 3bb229e..HEAD` highlights eventually — 160 commits of behavior drift tonight only surfaced through failures.

**P5 — environment / advisories**
32. Install `interrogate` (or consciously exclude); identify the other 8 unavailable tools (see P0-6).
33. Add dprint to devShells.default (warning: running via `nix run nixpkgs#dprint` without project deps).
34. lychee private-links fleet decision (GITHUB_TOKEN vs exclude regex).
35. CI re-enable on GitHub: push + enable + first green run (preconditions met tonight).
36. After CI enablement: coverage-badge job unfreezes; erraudit CI job re-checked against the canonical invocation.
37. Set GITHUB_TOKEN for CI lychee once policy lands.

**P6 — docs / memory**
38. Cross-project lesson (crush-config `references/lessons.md`, by commit): "bundled-analyzer vs canonical-gate divergence + jq null-length trap + reinstall-doesn't-switch-profile" — generalizes beyond this repo.
39. AGENTS.md one-liner: health check `git log --diff-filter=D -- .buildflow.yml` for new sessions.
40. Confirm no feature changes this session need FEATURES.md/CHANGELOG entries beyond the SECURITY.md link (none expected).
41. If the erraudit skip ever lifts, delete its rationale from `.buildflow.yml` in the same commit (skips with dead rationale mislead).
42. Name the daemon-commit pattern in AGENTS.md ("config-deletion + policy-violation bundled in one heuristic commit") for faster recognition next time.

**P7 — smaller / deferred**
43. wise_test.go (4229 lines) is the elephant for any future file-size enforcement — eventual split candidate list: wise_test, internal_test, types.go (800).
44. `agentic_fetch` of Wise pages returns huge payloads; prefer targeted HEAD-style checks for link verification.
45. Never let formatter commands sit between a tool and `$?` when measuring exit codes (PIPESTATUS unavailable in this shell) — personal habit, already applied.
46. When duplicating findings JSON for comparison, strip ANSI + keep stderr out (`2>/dev/null`) — the 2>&1 merge broke jq once tonight.
47. OTT 2026Q4 rollover probe (documented flip procedure) — parked, pre-existing.
48. Sandbox-key pass (paired with 47 per TODO_LIST) — parked, pre-existing.
49. After erraudit upstream fix lands someday: re-diff the 29 vs canonical gate output and document the delta in AGENTS.md before lifting the skip.
50. Keep lo decision parked: revisit only if the fleet converges on samber/lo as standard (currently 34/322).

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF (3)

1. **Toolchain endgame:** What is the actual plan for Go 1.27 — a condition or date for moving the flake toolchain + `GOTOOLCHAIN` off the 1.26 pin? Tonight I again answered a modernizer push by reverting (per documented policy); if the migration is planned, repeated reverts fight the future; if not, the 1.27-pushing dispositions should be skip-listed fleet-wide instead of per-incident reverts. (Carried from report v1; still the highest-leverage open decision.)
2. **Upstream filing now or later:** The BuildFlow repo had a concurrent session mid-refactor in exactly this area tonight (gomod_tidy_pregate). Should I file the three evidenced bugs (embedded erraudit false positives at HEAD; `nix run .#reinstall` profile-lock; file-size-check 0-scan) in the BuildFlow repo now, or are they known/being handled by whoever is active there — and if I file, in one combined issue or three separate ones?
3. **Release coordination:** Should this session's user-visible change (SECURITY.md reporting link → bugcrowd.com/wise) be folded into the concurrent v0.11.0 release notes / CHANGELOG, or is tonight's work infra-only and excluded?

---

_Point-in-time snapshot, 2026-10-07 04:20 CEST. Supersedes the 03:45 report for current state (that file carries a same-session addendum; both are point-in-time per the docs-health ANNOTATE convention). Section (f) is the HARVEST input for TODO_LIST.md — blocked only by the concurrent uncommitted TODO_LIST.md edit. Then: WAIT FOR INSTRUCTIONS._
