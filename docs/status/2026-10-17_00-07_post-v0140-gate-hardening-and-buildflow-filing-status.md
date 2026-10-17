# Status: Post-v0.14.0 Gate Hardening, BuildFlow Filings, and Docs-Truth Sweep

**Session:** 2026-10-16 ~23:00 → 2026-10-17 00:07 CEST · **Repo:** wise-go @ `94672d4` (clean tree, all local gates green) · **Author:** execution session resumed from the 21:45 SUPERB plan (user directive: "Execute and Verify them one step at the time. Repeat until done."), following the 22:44 release-train status report.

**Context:** multiple parallel crush sessions ran throughout (the error-system excellence session CLOSED its epic tonight at `b60bd30` + `eda00fe`; the auto-commit daemon swept every ~2 min and committed 23 commits in this window). This report covers this session's run and what it noticed — not a re-audit of the error-system work (its own report landed at `docs/status/2026-10-16_23-30_error-system-execution-status.md`).

---

## a) FULLY DONE

1. **M9 linter-allowlist finished** — `checks.linter-allowlist` (already built by a parallel session) negative-tested in BOTH directions this session: positive = 108-linter exact match; negative = injected `lll` detected with exactly one diff line. Policy recorded in the AGENTS buildflow gotcha (was "queued in TODO_LIST"). **The gate then caught a real lll re-add the same night** (see d3).
2. **M10 `--min-endpoints` override (T40)** — `spec_drift.py`: floor lifted to module constant `EXTRACTION_FLOOR = 40`, CLI override added, verified both directions (`--min-endpoints 45` → exit 3; `--min-endpoints 10` → exit 0). 2027Q1 16-path FEATURES triage (T42) verified already complete (bank-details + multi-currency sections); T43's totals ride doc-verify/anchor-check, both green.
3. **M11 negative fixtures (T44–T46)** — new `scripts/negative_fixtures.py`, 9/9 cases: pin-sync ok/stale-lock/tag-mismatch/missing-input/unpinned (each asserted on exit code AND output substring) + spec-drift broken-JSON (exit 3)/missing-endpoint (exit 1)/`--fail-on-drift` (exit 2)/warn-only (exit 0). Wired as hermetic `checks.script-fixtures` in `nix flake check`; green in the sandbox (log verified: "9/9 cases passed").
4. **M15 BuildFlow defects filed** — source-level verified at BuildFlow `f7029b39e`, verify-before-filing gates passed (no existing issues; current-source read despite stale local binary), check-draft 0 FAIL/0 WARN, both landed and confirmed via `gh issue list`:
   - [#64](https://github.com/LarsArtmann/BuildFlow/issues/64) — `FindingsFromOutput`'s `parseLineFormat` fallback counts one "finding" per rendered output line (statix: 2 real findings → "27 findings reported"; the repo's own gotchas #142–#235 name this disease class; the usage/crash guard doesn't cover multi-line diagnostic renders).
   - [#65](https://github.com/LarsArtmann/BuildFlow/issues/65) — the health-check phase runs language-inapplicable tools: 9 of the 9 "unavailable" are JS/TS+Python tools (jest, knip, madge, publint, svelte-check, vitest, vue-tsc, js-coverage, interrogate) in a Go-only repo, while jest's own trigger carries `OnlyLanguage(JavaScript)`. **This CORRECTS the 2026-10-11 attribution** ("on-demand tools count as unavailable") — the on-demand trio sits correctly in "skipped via config"; AGENTS.md was wrong for a day and is now fixed.
5. **Art-dupl baseline verification (§g-Q2 of the 22:44 report)** — answered without waiting: gate is exit-code-based (in-source directives), **23 groups, exit 0**. The one asymmetry — the new `AccountIdentification` raw/public mirror pair lacking directives — fixed with both-sites `art-dupl:accept` directives (filtered 7→8, shown 2→1); build green after.
6. **LSP tripwire (§f-29 of the 22:44 report)** — the `DuplicateDecl` phantoms in `internal/raw/webhooks.go` were a SECOND phantom class: stale gopls cache (build + vet green throughout). A gopls restart cleared all project diagnostics. AGENTS gotcha rewritten to the two-class form (env `GOEXPERIMENT` class vs stale-cache class).
7. **M17 crush-config lessons** — pipe-masking lesson verified already present; added two new lessons: bundled-analyzer binary freshness (the 2026-10-07 erraudit trap, generalized) + mechanical snippet parse-gates (md-go-validator pattern, bank-sync drift as evidence). Committed in crush-config at `5c10455`, pre-commit passed.
8. **M18 + M20 decision memos** — `docs/planning/2026-10-16_decision-memos.md` with D1–D5: live-harness shape (keep env-gated suite), Health P1b (additive `VerifyCredentials` in v1.1, no `Health` behavior change), spec-drift posture (pre-release stays report-only; `--fail-on-drift` goes scheduled), rev-vs-tag pins (keep rev pins; tags hide drift behind the lock), darwin matrix (drop x86_64-darwin, reversible). All four TODO_LIST decision rows now point at their memo.
9. **M19 user-asks bundle completed** — the parallel session's `2026-10-16_open-user-asks.md` (5 asks) extended to 8: added CI token secrets (CACHIX/ERRAUDIT with exact commands), the darwin matrix decision, and v1.0.0 timing (now sharpened by the error-system epic closing tonight).
10. **M16 docs truth** — DOMAIN_LANGUAGE: 22/22 backticked terms hit Go code (zero drift) + 4 new v0.14.0 entities added (OneTimeToken record, Account details, Transfer originator, Typed webhook payload); FEATURES inventory got the re-derivation note (vendored 2027Q1 spec = 188 paths / **227 operations**, +12 over the 215 live-preview census — different denominators, noted as deliberate); open-banking comparison got its point-in-time header; T66's 19-59 review annotations were already resolved by the parallel session.
11. **M24 prune, extraction half** — the five worst incident narratives (branching-flow, samber/lo census, strong-id, erraudit generic-return, art-dupl) + full histories moved to `docs/reviews/2026-10-16_agents-incident-narratives.md`; AGENTS.md 57,987 → **53,216 bytes** with zero constraint loss and two memory corrections folded in (9-unavailable attribution; LSP two-class).
12. **M26 harvest** — 21:45 plan: 13 more rows struck with verdicts/hashes (M8–M12, M14–M20, M24-partial); TODO_LIST: 4 rows closed with DONE verdicts (spec-drift hardening remainder, negative-scenario tests, benchstat baseline, BuildFlow filings), 1 new row (AGENTS prune continuation), 4 decision rows wired to memos.
13. **Final gate battery, all green at `94672d4`** — `nix flake check` exit 0 (now including `checks.script-fixtures` and the de-lll'd allowlist), 3× `go test -shuffle=on` exit 0, anchor-check 101/101 across 8 living docs, `doc-verify` all checks passed, build green, tree clean (daemon-swept).

## b) PARTIALLY DONE

1. **AGENTS.md prune to <30 KB** — 53,216 bytes; the extraction half is done (T85/T86), the compression half (T87: ~40 remaining 400–900 B constraint bullets, bullet-by-bullet under the carry-every-concept rule) needs a dedicated session. TODO_LIST row added.
2. **M4 CI resurrection** — pin-sync job + refreshed workflow landed (prior sessions); remaining: `gh workflow enable ci`, first green run, badge unfreeze, cachix/erraudit disposition, art-dupl blocking decision. All user-ratified.
3. **BuildFlow #64/#65/#63** — filed and evidenced, but unfixed upstream; the fleet keeps eating lll re-adds until #63 (auto-configure `WithUserDisabled`) lands. My single-step A/B of the auto-configure step did NOT reproduce the re-add, so the write attribution for tonight's dirty `.golangci.yml` remains inconclusive (parallel session vs buildflow) — the removal was justified by POLICY (`d8e28b5` + the gate), not by proven authorship. **→ update 2026-10-17: two more BuildFlow defects filed (#66 file-size-check inert-here fix, #67 result-cache scoping) plus branching-flow #15/#16; live reproductions in the 01-55 report; tracked in TODO_LIST (security-window + buildflow/infra micro-batch rows)**
4. **M21 bank-sync migration** — code migration committed by the parallel session (`9e9f5176`/`b9e89dff` era); the redeploy + FX-footer verify rides the host NixOS config and is surfaced as user ask #3.
5. **Error-system excellence epic** — executed and CLOSED tonight by the parallel session (`b60bd30` matrix/fuzz/conformance fixtures, `eda00fe` task-list close-out); outside this session's scope, noted as observed context. Its consumer-facing follow-through (bank-sync recipe against the new error surface) is unverified by me.
6. **§f harvest of the 22:44 report** — its 50 items are still not routed into TODO_LIST/ROADMAP (deferred then because TODO_LIST was foreign-dirty; the tree is clean NOW, so the next session should run docs-health HARVEST on BOTH the 22:44 report and this one).
7. **Buildflow binary staleness** — doctor advisories the binary is `83db580` while the BuildFlow repo HEAD is `f7029b39e`; tonight's triage read CURRENT source to compensate, but the binary itself needs the manual profile swap (upstream #35 workaround) before the next verdict-trust moment.

## c) NOT STARTED

1. **M22 sandbox-key lane** — 5-assumption live pass, OTT `2026Q3→2026Q4` probe + `ottAPIVersion` flip, first recorded sandbox-live workflow run. Blocked on `WISE_SANDBOX_API_KEY` (user ask #2).
2. **M23 Go 1.27 migration** — conditions still unmet (nixpkgs on 1.26.8); plan exists, execution gated.
3. **M25 baseline consumption** — benchstat vs the fresh v0.14.0 baseline in the next perf-touching PR (PR-coupled).
4. **M27 v1.0.0 tag** — one command on your word; sharpened question in g2.
5. **Scheduled `--fail-on-drift` job** (D3's recommendation) — needs CI enable first.
6. **Health P1b `VerifyCredentials`** (D2) — v1.1 additive, after user design word.
7. **Darwin matrix execution** (D5) — one-line change after user pick (ask #7).
8. **CI token secrets** — exact commands ready in the asks doc (ask #6).
9. **spec_drift `go/parser` extractor rewrite** — the long-arc TODO row (kills the call-shape-bug class).
10. **nixpkgs bump window** (`nix flake update nixpkgs` + vendorHash cascade) — needs a quiet window (ask #4), also the vulnix lever.

## d) TOTALLY FUCKED UP

Nothing at HEAD is broken — tree clean, every gate green. The honest failures were process-level:

1. **The daemon committed a self-inconsistent tree mid-session** (`58b5f90`, 31 files): `.golangci.yml` WITH `lll` alongside an allowlist check that requires 108-without-lll — master was red-by-construction for ~2 minutes of history. Root cause: I ran the full `buildflow --fix` battery (mandated after infra commits) while another session was mid-edit; the daemon swept the mixed state. Resolution was clean (lll removed per policy, allowlist green), but the mixed-authority commit is exactly the race the AGENTS daemon gotcha warns about — I ran into it anyway.
2. **Fixture first-design was wrong three times**: fake `.nix` broke statix (5 parse errors — MY fixtures), then missing trailing semicolons, then the `ok` case lost an input, then a single-line `require` broke pin_sync's line-anchored regex, then the runner wrote into what would be a read-only sandbox. Five iterations, each caught by tests BEFORE landing — but committed-fixture-go.mod-as-real-module was foreseeable (repo-scan tooling) and cost the buildflow run a false "5 modules need go mod tidy".
3. **An invalid A/B nearly drove a wrong conclusion**: my auto-configure attribution test sampled `git show HEAD:.golangci.yml` AFTER the daemon had already committed the lll state — the "clean" side of the A/B wasn't clean. Caught by diffing the saved copies; attribution left honestly inconclusive instead of asserted.
4. **The 22:44 report's 3 questions got asked twice in prose without an answer loop** — should have lived in the open-asks doc from the start (fixed tonight: asks doc now carries 8, and this report stops re-asking in prose; see g).
5. **Forgot-to-close-loops**: (a) the 22:44 §f 50-item harvest — deferred while TODO_LIST was foreign-dirty and then not revisited when it went clean (now explicitly routed, b6/c10); (b) the post-rewrite `checks.script-fixtures` sandbox re-verification initially returned no-output from a bad `nix log` glob — second command form verified it; small, but the first "no output" could have been misread as success in a sloppier moment (the 22:44 session's misread-as-success lesson nearly bit again).

## e) WHAT WE SHOULD IMPROVE

1. **View-before-edit discipline in daemon-racing sessions** — three edit refusals tonight (account_details.go ×2, AGENTS.md ×1) were all "bash-read doesn't update the edit tool's read tracking". Rule: `view` immediately before `edit`, always, no bash-sed shortcut.
2. **Generate fixtures at runtime by default** — any committed file named like a real project file (go.mod, flake.nix, package.json) WILL be eaten by repo-scan tooling. Tempdir-generation should be the first design, not the fifth.
3. **A/B attribution hygiene** — snapshot the exact commit BEFORE sampling state; verify both sides of the A/B are actually in the claimed state before running the experiment.
4. **Full `buildflow --fix` runs and parallel sessions don't mix** — schedule the mandated full run for single-session windows, or accept mixed-authority daemon commits as the price and re-verify authorship after.
5. **Volatile numbers don't belong in AGENTS** — the art-dupl "23 clone groups" count rots on every endpoint addition; keep the gate command, drop the census number (or point at the narratives file).
6. **Question lifecycle** — user-gated asks go straight into `open-user-asks.md` with exact unblock steps; prose re-asking is waste (this report's g section is the third ask for CI enable).
7. **Fixture README** — `scripts/testdata/spec_drift/` deserves a two-line README stating which case expects which exit, so the next maintainer doesn't reverse-engineer `negative_fixtures.py`.
8. **Benchstat into the devShell** — TODO notes it's missing; tonight's baseline work re-tripped the install friction.
9. **Normalize the narratives file** — it's currently raw extracted bullets; title-form them when next touched so the pointer from AGENTS reads as a document, not a dump.
10. **Rebuild the buildflow binary** — the stale-binary advisory will eventually produce a verdict made on dead code (the exact 2026-10-07 trap the new crush-config lesson names); the manual profile swap (upstream #35) is 2 minutes.
11. **POST-filing upstream tracking** — #64/#65/#63 have no local reminder to check for maintainer response; a TODO row per filed issue (or one "upstream watch" row) would close the loop.
12. **CHANGELOG for gate work** — tonight's script-fixtures/linter-allowlist/--min-endpoints live only in AGENTS/plan/TODO; the repo's convention folds gate changes into the next release's CHANGELOG, which hasn't happened yet.

## f) UP TO 50 THINGS TO GET DONE NEXT (brainstorm — most are ROADMAP fuel, not commitments)

| #  | Task                                                                                                                                                                     | Lane             |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------- |
| 1  | Enable CI (`gh workflow enable ci`) + watch first run                                                                                                                    | user word        |
| 2  | Post-first-green: art-dupl blocking decision (warn-first 2 weeks vs promote)                                                                                             | user word        |
| 3  | Set `CACHIX_AUTH_TOKEN` secret (commands in asks doc)                                                                                                                    | user word        |
| 4  | Set `ERRAUDIT_TOKEN` secret                                                                                                                                              | user word        |
| 5  | Push master when tree is clean + lint-green                                                                                                                              | user word        |
| 6  | Execute darwin decision (D5): drop x86_64-darwin or pin 26.05                                                                                                            | user word        |
| 7  | Tag v1.0.0 on your word (pre-release `--shuffle` chain first)                                                                                                            | user word        |
| 8  | Finish AGENTS prune <30 KB (dedicated bullet-by-bullet session)                                                                                                          | wise-go          |
| 9  | HARVEST both status reports' §f (22:44's 50 items + this one) into TODO_LIST/ROADMAP                                                                                     | docs-health      |
| 10 | `WISE_SANDBOX_API_KEY` lane: 5-assumption live pass                                                                                                                      | blocked on key   |
| 11 | OTT `2026Q4` probe; flip `ottAPIVersion` if non-404                                                                                                                      | blocked on key   |
| 12 | First recorded sandbox-live workflow run + CHANGELOG entry                                                                                                               | blocked on key   |
| 13 | bank-sync redeploy + FX footer verify (host config bump)                                                                                                                 | consumer         |
| 14 | Health P1b: implement `VerifyCredentials` in v1.1 after design word                                                                                                      | user word        |
| 15 | Scheduled spec-drift `--fail-on-drift` CI job (weekly)                                                                                                                   | needs CI         |
| 16 | BuildFlow #64 fix: statix `--json`/`--oneline` dedicated parser                                                                                                          | BuildFlow repo   |
| 17 | BuildFlow #65 fix: applicability filter for health checks                                                                                                                | BuildFlow repo   |
| 18 | BuildFlow #63 fix: auto-configure `WithUserDisabled` (stops lll fleet recurrence)                                                                                        | BuildFlow repo   |
| 19 | Rebuild buildflow binary (manual profile swap per #35)                                                                                                                   | BuildFlow repo   |
| 20 | spec_drift extractor rewrite on `go/parser`                                                                                                                              | wise-go long-arc |
| 21 | `nix flake update nixpkgs` window + vendorHash cascade (also the vulnix lever)                                                                                           | quiet window     |
| 22 | Verify coverage badge unfreezes on first master push post-enable                                                                                                         | needs CI         |
| 23 | apidiff re-check now that the error-system surface landed post-tag                                                                                                       | wise-go          |
| 24 | CHANGELOG entries for tonight's gate work (script-fixtures, linter-allowlist, `--min-endpoints`)                                                                         | wise-go          |
| 25 | Verify the error-plan's "snippets compile vs v0.13.0 AND master" mechanical check actually landed                                                                        | wise-go          |
| 26 | Validate the bank-sync consumer recipe against the NEW error surface (`APIError.Details`, `WiseErrorCode`)                                                               | consumer         |
| 27 | Disposition buildflow's bundled art-dupl step (warned 20 findings at its own threshold) — skip_steps like erraudit, or accept as noise                                   | wise-go          |
| 28 | TODO rows per filed upstream issue (#64/#65/#63) so responses get noticed                                                                                                | wise-go          |
| 29 | Drop the volatile art-dupl group count from AGENTS (keep the gate command)                                                                                               | wise-go          |
| 30 | Fixture README for `scripts/testdata/spec_drift/` (case → expected exit table)                                                                                           | wise-go          |
| 31 | benchstat into the devShell                                                                                                                                              | wise-go          |
| 32 | go.mod guard step: assert `grep '^go '` stays 1.26.x until the 1.27 flip (cheap daemon-sweep tripwire)                                                                   | wise-go          |
| 33 | pin-sync app on a CI schedule (currently dispatch/pre-release only)                                                                                                      | needs CI         |
| 34 | Negative fixture: spec_drift METHOD-GONE case (only MISSING is covered)                                                                                                  | wise-go          |
| 35 | Negative fixture: pin-sync rev-pinned-but-lock-entry-missing                                                                                                             | wise-go          |
| 36 | Normalize the narratives file into titled sections                                                                                                                       | wise-go          |
| 37 | Archive the 21:45 plan when M4/M22/M23/M25/M27 close (fully struck)                                                                                                      | docs-health      |
| 38 | Annotate + archive the 23:30 error-system status report when its tail ships                                                                                              | docs-health      |
| 39 | ROADMAP: error-system wave-2 items (their decision-gated tier)                                                                                                           | docs-health      |
| 40 | ROADMAP: post-probe OTT rollover follow-up                                                                                                                               | docs-health      |
| 41 | Spot-check the `jwsjwe` method-change FEATURES row reflects POST post-2027Q1                                                                                             | wise-go          |
| 42 | Consider `--fail-on-findings` CI posture experiment after CI settles                                                                                                     | needs CI         |
| 43 | Backport the linter-allowlist gate pattern to sibling repos (fleet)                                                                                                      | fleet            |
| 44 | crush-config lesson: daemon-mixed-authority commits during parallel buildflow runs                                                                                       | crush-config     |
| 45 | Discoverability: link decision memos + open asks from README or CONTRIBUTING                                                                                             | wise-go          |
| 46 | coverage-floor-parity: confirm belt-and-braces role is documented where a stranger would look                                                                            | wise-go          |
| 47 | Gitleaks allowlist: re-verify after next spec re-vendor (new fixtures may add example secrets)                                                                           | wise-go          |
| 48 | Decide LSP tripwire institutionalization (crushrc guard already warns; document the restart remedy in AGENTS — done tonight — and stop)                                  | wise-go          |
| 49 | v0.14.1 decision: does the post-tag error-surface + gate work warrant a patch release before v1.0?                                                                       | user word        |
| 50 | Retire this session's own scratch: `/tmp` evidence files are ephemeral — the durable repro evidence for #64/#65 lives in the issues; nothing to do, recorded for honesty | none             |

## g) THREE QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **CI enable + art-dupl posture** (asked twice, still unanswered): shall I run `gh workflow enable ci` now — and after the first green, promote art-dupl to blocking or keep it warn-first for two weeks? Everything else about CI is ready; secrets are optional-degrade.
2. **v1.0.0 timing, sharpened by tonight**: the error-system epic CLOSED tonight (structured `APIError.Details`, `WiseErrorCode`, matrix/fuzz/conformance fixtures all landed post-v0.14.0). Tag v1.0.0 on the current tree (46 methods + the new error surface, all gates green), or hold for the sandbox-key live pass first?
3. **Is the curated linter list final?** Tonight I removed a re-added `lll` from `.golangci.yml` under the `d8e28b5` policy without proven attribution (parallel session vs buildflow auto-configure). If you actually WANT lll now (deliberate policy change), the allowlist must change deliberately in the same commit — otherwise I keep treating every lll re-add as damage. Which is it?

---

**Override note:** skill default is a styled HTML dashboard; this report is `.md` on your explicit path instruction.
**Per the harness contract this file is not manually committed — the auto-commit daemon sweeps it.**

WAITING FOR INSTRUCTIONS.
