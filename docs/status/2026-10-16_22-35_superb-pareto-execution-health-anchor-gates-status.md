# Status: SUPERB Pareto-plan execution close-out — Health honesty, anchor gates, truth infrastructure

> Session 2026-10-16 ~21:50–22:35 CEST. Input: "READ, UNDERSTAND, RESEARCH, REFLECT. Break this down
> into multiple actionable steps. Execute and Verify them one step at the time. Repeat until done."
> Scope: this session's run only. **Format note:** `.md` at explicit user demand (status-report skill
> default is HTML; override honored, flagged per skill contract).
>
> **Concurrent-session context (dominates verification claims):** at session start the 21:13 SUPERB
> plan was already ~70% executed by parallel sessions (M1 `baba096`, M3 docs-half, M8/M9/M10
> `19ab4e5`/`3c185f7`, M12 counts `9611379`/`41b09dc`, M13 webhooks `443b01a`, M14 SCA matrix
> `703d80a`, M15 smoke `78a4f35`, M16 originator `5e5e1f0`, M20 art-dupl scaffold `cb7b271`). This
> session VERIFIED those claims against code and gates rather than redoing them, then executed the
> remainder. A release-prep session stayed active THROUGHOUT (benchstat baseline `7b211fc`, pin-sync
> CI job `13bad72`, lint nits `4b2b9dd`, an in-flight `.go` godoc sweep, and an annotate pass moving
> two 2026-10-07 status docs into HEAD). Its dirty tree is why `apidiff` could not run (gorelease
> refuses uncommitted changes — tried twice, recorded, not worked around).

## Verdict

Everything executable without the user, without keys, and without owning the parallel session's
code lane is done and gated: the Health story is honest AND test-pinned AND consumable
(DI recipe compile-verified against v0.13.0 + master), the anchor-drift class has a real mechanical
gate that caught the parallel session re-drifting ott.go mid-session (30 real drifts fixed across
two sweeps), spec-drift gained an extraction floor, the coverage floor is single-sourced, the
curated linter list is a gate, and the full battery is green. Two self-inflicted script bugs and
a false-first-green are honestly recorded in §d. Remaining: only user/key-gated items, the
release close owned by the 21:45 plan, and one blocked gate (apidiff).

---

## a) FULLY DONE

1. **State reconciliation first** — read the 21:43 docs-health audit, TODO_LIST, git log, and
   grep-verified each 21:13-plan item's landing before touching anything; no shipped work was
   redone, no parallel-session file was touched (`wise_test.go` re-read immediately before its
   single edit).
2. **M4 — `Describe("Health")` typed-error contract** (wise_test.go): 429 → `*RateLimitError`
   (asserts `RateLimitedBy`, `IsRetryable`), bad key → `*AuthError`, retry-exhausted 5xx → typed
   `*ServerError` through the `ErrExhausted` chain (fast-retry client per the suite's own idiom).
   Focused Ginkgo run: 4 of 4 Health specs passed; full `go test -race ./...` green after.
3. **M5 — README "Dependency Injection & Health Checks"**: samber/do v2 composition root +
   two-line `healthAdapter`, compiled in a scratch module against **v0.13.0 (do v2.1.0) AND
   master** (replace directive + GOEXPERIMENT=jsonv2). The scratch build caught a genuinely
   unparseable draft (`if err := healthAdapter{…}`) — the snippet rule earning its keep again.
   ToC entry added; md-go-snippets 40 valid / 0 errors; doc-verify links green; T27 naming-note
   footnote points at the alias decision.
4. **M6 — anchor sweep, 30 real drifts found and fixed** in FEATURES.md across two sweeps:
   first 16 (4 blank-line, 12 symbol-moved: balances/quotes/transfers/third_party/webhooks),
   then 14 more after the parallel session's godoc edits re-drifted
   ott/account_details/rates/recipients — all re-anchored at grep-verified `func` declaration
   lines (not the script's comment-line hints). Final: **93/93 anchors verified, exit 0.**
5. **M7 — the anchor-check gate**: `scripts/anchor_check.py` (symbol-anchored + bare checks,
   living-docs-only scope with the point-in-time exclusion rationale in its docstring,
   `--self-test` covering all four drift classes), hermetic `checks.anchor-check`,
   `apps.anchor-check` (self-test + sweep), pre-release chain entry after doc-verify, AGENTS.md
   gotcha rewritten from "queued in TODO_LIST" to the full gate description. Negative-tested via
   crafted fixture AND self-test.
6. **spec-drift hardening**: extraction-floor guard (exit 3 below 40 resolved endpoints; the
   silent call-shape/regex-rot class), negative-tested at floor=44 against the real 43-endpoint
   extraction; `PYTHONDONTWRITEBYTECODE=1` in the flake app. Real run: live 2027Q1, 188 paths,
   43 SDK endpoints all match, exit 0 — which also **formally verifies M1** (`baba096`'s
   account-details migration) and the re-vendored snapshot.
7. **Coverage floor single-sourced**: ci.yml carries THE floor (92.0); `checks.test` reads it at
   check time via the `ciYml` derivation attr; `checks.coverage-floor-parity` rewritten to pin
   the shape (exactly one floor in ci.yml, zero numeric floors hardcoded in flake.nix).
8. **`checks.linter-allowlist`**: the `.golangci.yml` enabled set must equal the pinned
   108-linter curated allowlist (buildflow re-added `lll` twice on 2026-10-16; policy is now a
   gate). Negative-tested with a deliberate `lll` re-add — diff + policy message shown.
9. **`.markdownlint.json`** (MD013 off) — the on-demand markdownlint item closed.
10. **M2 — gate battery recorded**: `go test -race ./...` 0 · `golangci-lint run` 0 issues ·
    `nix flake check` 0 (with all three new checks green inside it) · doc-verify 0 (re-run after
    every doc batch) · release-notes-check 0 · spec-drift 0 · anchor-check 93/93 ·
    md-go-snippets 40/0. All exit codes captured unpiped.
11. **M24/M27 — harvest and annotation**: TODO_LIST (P1 row updated — SCA matrix landed
    `703d80a`; four completed P3 items DELETED per the trophy-case rule; spec-drift item trimmed
    to its remainder; Health decision updated to shipped-vs-code-half), CHANGELOG `[Unreleased]`
    (7 entries under Added/Changed), 19:59 architecture review annotated inline (F-1 docs-half,
    F-3 done, F-5 resolved at `baba096` — only the user-gated code-half remains), AGENTS anchor
    gotcha rewritten.
12. **M25 — the single user-ask bundle drafted** (10 decisions/keys) and delivered in the closing
    message of the execution run.
13. **Daemon-commit verification** — every auto-commit sweep was spot-checked for content
    (README DI section, flake checks, script, CHANGELOG, FEATURES fixes all confirmed in HEAD
    trees via `git show HEAD:<file>`).

## b) PARTIALLY DONE

1. **apidiff** — blocked, not failed: gorelease refuses the parallel session's dirty tree (two
   attempts, 150 s apart). This session's delta adds no exported API (tests, docs, scripts,
   flake only), so the landing session's run covers it; recorded as the one open battery item.
2. **spec-drift hardening** — floor guard + no-pyc done; the 2027Q1 16-path FEATURES triage and
   the `go/parser` extractor rewrite remain routed (TODO_LIST carries the trimmed item).
3. **TODO_LIST P1** — updated to current truth, but its remainders (v0.14.0 release close, CI
   resurrection, smoke-harness decision, live smoke) are 21:45-plan/user-gated, not executed.
4. **Health decision item** — docs-half + typed-error tests + README recipe shipped; the
   code-half (fail-fast path / `VerifyCredentials` / `HealthCheck` alias) stays user-gated.
5. **The 21:43 audit's §f tail** — f7 (anchor-check) closed here; f1–f6 maximalist annotate
   passes appear to be actively worked by another session (two 2026-10-07 status docs moved in
   HEAD during this run — observed, not touched); f8 remainder, f9–f14 open.

## c) NOT STARTED (observed, routed, zero execution this session)

1. v0.14.0 tag + push (CHANGELOG section exists; no git tag yet) — user-gated.
2. CI re-enable (push + `gh workflow enable ci`) — user-gated; pin-sync job already landed by
   the other session (`13bad72`).
3. Negative-scenario fixtures for `pin_sync.py`/`spec_drift.py` wired into flake checks
   (anchor-check got its self-test; those two did not).
4. Upstream filing of the two BuildFlow display defects (needs verify-before-filing).
5. Consume/verify the benchstat baseline the other session cut (`7b211fc`) in the next
   perf-touching PR.
6. Cross-repo lessons (pipe-masking, ps-before-gates) into crush-config `lessons.md`;
   DOMAIN_LANGUAGE re-verify; AGENTS size-reduction session; sandbox key drop → live pass →
   OTT 2026Q4 flip; cachix/erraudit tokens.

## d) TOTALLY FUCKED UP

1. **anchor_check.py shipped with two bugs — and the first "all 93 verified" was a FALSE GREEN.**
   (a) The symbol prefix was computed from the whole-file offset instead of the line offset, so
   every symbol-anchored check silently degraded to the weak bare check — I reported "93/93
   clean" in a todo update minutes before fixing the bug exposed 12 more real drifts. (b)
   `locate_symbol`'s func pattern `func[^(]*` cannot cross receiver parens, so hints pointed at
   doc-comment lines instead of declarations. The checker failed exactly the way unchecked
   checkers fail: quietly, in the direction of green.
2. **Self-test written AFTER the script, not with it** — the fixture's first run tested nothing
   (doc filename outside the sweep list → "0 anchors verified, exit 0" read like success), and
   my expected line-count in the self-test was wrong (6 vs the fixture's actual 5 lines). The
   gate's own gate was the last thing built; both script bugs were found by fixture runs that
   should have been mechanical from minute one.
3. **Two flake iterations burned on string/grammar traps**: `${FLOOR}` inside a Nix `''` string
   is interpolation (undefined variable), and the parity detector's grep pattern matched my own
   new awk comparison (`(p+0) < (f+0)`) — the check failed on itself; also one Nix list-element
   semicolon syntax error. None dangerous, all avoidable by grepping the diff for `${` before
   building.
4. **A multiedit that reports "file modified since read" but actually applied** — the 14-fix
   batch's error message and the file state disagreed; I had to verify all 14 values
   individually instead of trusting either signal. Cost a full verification loop.
5. **Three edit-tool mod-time races** (FEATURES ×2, flake ×1) — grep does not reset read-state,
   view does. Known mechanic from the 21:43 audit's own report; paid the tax anyway.
6. **apidiff waited on twice** (~5 min total) in a window where the tree was guaranteed to stay
   dirty — should have accepted the block on first refusal and moved on.

## e) WHAT WE SHOULD IMPROVE

1. **Author the self-test with the tool, not after it** — "never trust a checker that hasn't
   failed against a known-bad input" is the repo's own rule; sequencing it last is what turned
   bug (d1a) into a published false green. Default order for any new gate: negative fixture →
   script → real sweep → positive report.
2. **Treat a checker's first green as provisional until its negative test has failed at least
   once.** Concretely: anchor-check's bare-check-only mode PASSED while blind to 12 of 16
   drifts; only the Beta-moved fixture would have caught that on day one.
3. **Multi-session hygiene is a mechanical habit, not a memory**: `view` immediately before
   every edit batch (resets read-state AND catches foreign edits), `git log` between composing
   and landing living-doc edits, and never pipe an exit code — all three bit this session in
   minor ways despite being written down.
4. **Nix string discipline**: before building, grep new `''` blocks for `${`; prefer `$VAR` or
   concatenation; remember list elements take no semicolons. Two of three flake failures were
   this class.
5. **When tool output and file state disagree, verify content directly and move on** — do not
   re-apply a possibly-applied batch (replace_all double-apply risk).
6. **Gated tools in a shared tree**: accept environment blocks (dirty tree) on first refusal,
   record them, and let the owning session's landing run cover them — waiting twice bought
   nothing.

## f) Up to 50 things we should get done next (honest count: 20)

1. Re-run `nix run .#apidiff` once the tree lands clean (gorelease vs v0.13.0).
2. Cut v0.14.0: `nix run .#pre-release`, tag, push, proxy + `go get` verify (user approval).
3. Re-enable CI: push, `gh workflow enable ci`, watch first run incl. the new pin-sync job.
4. Execute the Health code-half once the user picks (fail-fast / `VerifyCredentials` /
   `HealthCheck` alias / keep).
5. Triage the 16 added 2027Q1 paths into the FEATURES matrix (spec-drift remainder).
6. Rewrite the spec_drift extractor on `go/parser` (kills the call-shape-bug class).
7. Wire negative-scenario fixtures for `pin_sync.py` + `spec_drift.py` into flake checks
   (mirror anchor-check's `--self-test` pattern).
8. Maximalist annotate passes on the kept 2026-10-07 reports (f1–f6 of the 21:43 audit) —
   coordinate with the session apparently already moving them.
9. Verify the new `checks.linter-allowlist` survives the next full buildflow run (it exists
   precisely for buildflow's auto-configure).
10. Fresh anchor sweep at v0.14.0 tag time — the release-prep session's godoc sweep WILL
    re-drift anchors again; the gate will catch it, someone must run it at landing.
11. Verify the other session's benchstat baseline content (`7b211fc`) against the bench README
    conventions before the tag.
12. Drop `WISE_SANDBOX_API_KEY` → sandbox live pass on the five spec-vs-live assumptions →
    OTT 2026Q4 probe/flip.
13. Set `CACHIX_AUTH_TOKEN` + `ERRAUDIT_TOKEN` secrets; delete the two `continue-on-error`
    lines in the erraudit job when the PAT lands.
14. Upstream-file the two BuildFlow display defects (verify-before-filing pass first).
15. Consume the committed benchstat baseline in the next perf-touching PR.
16. AGENTS.md size-reduction session (48.6 KB and growing every incident).
17. DOMAIN_LANGUAGE terms re-verify against the 46-method surface.
18. api-docs-study 215/51 counts re-audit post-2027Q1.
19. Cross-repo: pipe-masking + ps-before-gates lessons into crush-config `references/lessons.md`.
20. Consider full single-sourcing of the coverage floor (CI reads the same checked-in value the
    flake reads — current state is already one source, ci.yml; this removes the grep coupling).

## g) Questions I can NOT figure out myself (3)

1. **Lane arbitration:** the release-prep session was still landing `.go` edits when this report
   was written (dirty recipients.go et al., apidiff blocked twice). Do you want every other
   session to treat the Go-code lane as exclusively theirs until their report lands (docs/
   tooling allowed in parallel), or is concurrent code work acceptable? This decides whether
   the final anchor sweep + apidiff belong to them or to a follow-up of mine.
2. **v0.14.0 timing:** tag now with what is staged (breaking OTT typing, 2027Q1 migration,
   transfer-category completion, today's gates) and push master — or hold for anything specific?
   (21:45-plan M1–M3 + the push/enable approvals all hinge on this.)
3. **Annotation depth (carried from the 21:43 audit, still unanswered):** maximalist
   item-by-item resolution on the remaining kept 2026-10-07 reports over ~2 docs passes, or is
   banner-granularity acceptable now that every open item is verifiably routed into
   TODO_LIST/ROADMAP? It decides whether f-item 8 stays queued.

---

*Point-in-time snapshot 2026-10-16 22:35 CEST. Verification appendix: `go test -race` 0 ·
`golangci-lint` 0 · `nix flake check` 0 (anchor-check + linter-allowlist + coverage-floor-parity
green inside) · doc-verify 0 · release-notes-check 0 · spec-drift 0 (live 2027Q1, 188 paths, 43
endpoints) · anchor-check 93/93 after 30 fixes · md-go-snippets 40/0 · apidiff BLOCKED (dirty
tree, parallel session). Not run: nothing else was in scope. WAITING FOR INSTRUCTIONS.*
