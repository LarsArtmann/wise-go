# Status: P3 TODO Execution Sweep — 2026-10-07 06:07

> **Shipped (2026-10-07 docs-health pass):** the §a sweep all landed and is
> recorded in CHANGELOG `[Unreleased]` / AGENTS.md — the CAMT test, webhook
> fuzz/bench, the `RequestLog.RetryAfterDelay` log line, the shuffle-proof
> conformance guard, coverage 93.1%, the benchstat baseline, `md-go-snippets`,
> `doc-verify`→`docs/releases/*`, `release-notes-check`, and the `pre-release`
> app; the art-dupl policy decision (§g1) is answered. The x86_64-darwin and
> `RequestLog.RetryAfterDelay`-freeze questions are harvested to `TODO_LIST.md`;
> the rest is routed to `ROADMAP.md`.

Session goal: execute the unblocked P3 (Quality & tooling) items from
TODO_LIST.md end-to-end — READ/UNDERSTAND → break down → execute → verify,
one step at a time. P1/P2 items were skipped by design (all user-blocked).
Working tree is CLEAN at report time (auto-daemon swept everything); full
suite green (`go test .` → ok, 3.8s); `golangci-lint run` → 0 issues.

---

## a) FULLY DONE (implemented + verified this session)

1. **CAMT (`.xml`) GetStatement test** — new Ginkgo Context mirroring the
   PDF/XLSX pattern: mock serves `Content-Type: application/xml`, asserts
   raw-byte round-trip of a camt.053.001.02 document. The last statement
   format without direct coverage is now covered. (wise_test.go)
2. **Webhook fuzz targets** — `FuzzParseWebhookPublicKey` (invariants: never
   panics, no key+error together, PKIX PEM round-trip) and
   `FuzzVerifyWebhookSignature` (never panics; genuine pair always verifies;
   tampered payload with genuine signature always rejects), each with
   hand-picked seeds + generated-key setup. Live-fuzzed 8s each: clean.
   (internal_test.go)
3. **`BenchmarkVerifyWebhookSignature`** — per-delivery RSA-SHA256 hot path:
   ~22.8 µs/op, 10 allocs/op. (bench_test.go)
4. **`FuzzDecodeExchangeRates`** — array/single-object/empty/corrupt/null
   seeds; invariants: never panics; accepted payloads decode as array (≥1
   entry) or tolerated single object; an exact source/target match in the
   payload must be the returned entry. Live-fuzzed 8s: clean.
5. **`BenchmarkDecodeExchangeRates`** — 4-entry array decode, requested pair
   not first: ~1.2–1.7 µs/op, 4 allocs/op.
6. **Retry-After-honored log line** — `RequestLog.RetryAfterDelay` field
   (options.go); the go-retry `DelayFunc` (client.go) now emits a
   delay-decision entry (Status=429, honored wait in Duration, Method/URL
   empty, hint possibly capped) through the existing Logger channel; BDD
   test pins the full 3-entry sequence (429 exchange → delay entry → 200
   exchange) with honored=min(30s,1ms)=1ms. `Retry-After: 0` still produces
   NO delay entry (falls through to backoff) — the existing "on retries"
   test stays green.
7. **Spec-conformance coverage guard is shuffle-proof** — the floors moved
   from the per-test `zz_` guard into a `TestMain` post-run assertion gated
   on a new `conformanceSuiteCompleted` flag set by `TestWiseClient` after
   `RunSpecs` returns. Verified: normal run ok; **10/10 shuffled runs ok**;
   `-run` filtered runs pass vacuously with an explanatory note; negative
   test (floor 99999 vs 180 actual exchanges) fails loudly, then restored.
   Statement-variant floor raised 3 → 4 (actual 6 after the CAMT test).
   Actuals now: 180 exchanges / ≥25 templates / 6 statement variants.
8. **Coverage headroom: 90.8% → 93.1%** (floor 90.0, target was ~92%) via a
   white-box batch in internal_test.go: `TestBrandNames` (all 11 phantom
   brands, previously 0%), corruption-path tables for
   `mapExchangeRate`/`mapMultiCurrencyAccount`/`mapUser`/`mapTransaction`
   (incl. exchange from-amount branch), validation tables for
   `CreateBalanceRequest.validate`/`GetStatementRequest.validate`,
   `TestBalanceTypeWire`, and `TestCheckErrorUnreadableBody` (the
   "response body could not be read" branch, via a failing reader).
9. **Benchstat baseline committed** — `docs/bench/2026-10-07_v0120_baseline.txt`
   (6 benchmarks × `-count=6`, standard bench format). Ritual documented in
   CONTRIBUTING (Benchmarks section). NOTE: `benchstat` itself is not
   installed locally; the file format is standard but I could not
   round-trip it through benchstat.
10. **md-go-validator wired into `nix flake check`** — new input
    `github:LarsArtmann/md-go-validator/v1.3.0` (source-only, `flake = false`)
    built inside `checks.md-go-snippets` via
    `(pkgs.buildGoModule.override { go = pkgs.go_1_27; })` with
    `proxyVendor`, `GOEXPERIMENT=jsonv2`, and vendorHash
    `sha256-QUgeh99RqCc80oh1UNJDH38Llm8jMW3hQkKmPGZr3NE=`. Scope: living
    docs + every `docs/` subdir (30 blocks in living docs, 5 valid / 3
    skipped in docs/, 0 errors). Verified both directions: green on the
    real tree; a deliberately broken ```go fence appended to README FAILED
    the check; restored → green.
11. **doc-verify count-claims extended to release notes** — the LATEST
    `docs/releases/*-release-notes.md` (sort -V) must claim the current
    method count; historical notes stay frozen at their tags. Today:
    `ok: docs/releases/v0.12.0-release-notes.md: 41 client methods`.
    Fixed shellcheck SC2012 (`ls` → `find`) that writeShellApplication
    rejected.
12. **`nix run .#release-notes-check`** — release-notes pre-flight for
    `gh release create`: detects (a) code spans split across lines (odd
    backtick count outside fences — the v0.12.0 reflow defect class) and
    (b) repo-relative link targets (GitHub Release bodies can't resolve
    them). Heuristics negative-tested on a crafted bad file; all three
    existing notes clean.
13. **`nix run .#pre-release`** — one-command release gate: dirty-tree →
    build → vet → race tests → golangci-lint → `nix flake check` →
    doc-verify → release-notes-check → apidiff. Verified the first link
    (correctly FAILS on a dirty tree); the full chain has NOT yet run
    end-to-end on a clean tree (see b).
14. **CONTRIBUTING.md** — three new sections: "Release gates"
    (release-notes-check + pre-release), "Quarterly API-surface rollover
    ritual" (probe → flip → straggler grep → changelog → verify, with the
    probe-blind OTT caveat and the independent-rollover warning), and
    "Benchmarks" (baseline location + regenerate/compare commands).
    doc-verify + snippet gate re-verified green after the edit.
15. **art-dupl enforced-gate decision — DECIDED (not yet recorded)** —
    grounded in a live run: `art-dupl -t 2 --type-aware .` → 16 clone
    groups, 0 actionable, exit 0. Decision: **accept the suppressed
    baseline as policy, no enforced gate**; revisit trigger = CI
    re-enable, where it should land as a warn-first (continue-on-error)
    job and be promoted after two green weeks. Rationale: the repo is
    dependency-light; buildflow has no custom-command step to host a gate;
    CI is disabled (a gate today would run nowhere and give false
    confidence); suppression config is in-source so future enforcement is
    a one-liner. **Still needs: the AGENTS.md record** (see c).
16. **Discovered + worked around: md-go-validator's own flake cannot build
    its own package** — its package.nix uses `pkgs.go` (1.26.7,
    GOTOOLCHAIN=local) while its go.mod requires ≥ 1.27, so its
    `packages.default` fails (`go: go.mod requires go >= 1.27`). This is a
    4th tool-repo defect discovered this session (file alongside the
    BuildFlow three; see c). The wise-go check sidesteps it by building the
    tool with `go_1_27` from this repo's nixpkgs.

## b) PARTIALLY DONE

1. **golangci_lint_ls phantom (webhooks.go:142)** — investigation started,
   NOT root-caused. Facts assembled:
   - webhooks.go:142 is exactly the `*mapped, //nolint:branching-flow:panic`
     line (the documented suppressed dereference).
   - "File is not properly formatted" is golangci-lint v2's FORMATTER
     finding message (gci/goimports/gofumpt/golines are configured).
   - CLI: golangci-lint 2.14.0 (nix store), built with go1.27.1, run inside
     the devShell with GOEXPERIMENT=jsonv2 + GOTOOLCHAIN=local.
   - LSP: crushrc wires `golangci_lint_ls` through a host-managed wrapper
     `~/.local/bin/golangci-lint-lsp-wrapper` → `golangci-lint-langserver`
     → `golangci-lint` from PATH, with `--env GOTOOLCHAIN auto` and caches
     pinned to /mnt/buildcache. **The wrapper does NOT set
     GOEXPERIMENT=jsonv2** — it relies on "inherit the session env", so a
     session/daemon context without GOEXPERIMENT gets package-load failures
     (jsonv2 build constraints) and possibly formatter-only verdicts.
   - Version skew risk: the devShell historically pins v2.13 (ci.yml) while
     the system/PATH binary is 2.14.0 — if the langserver resolves a
     different golangci-lint than the CLI shell does, formatter versions
     disagree → phantom.
   - Live probe at 06:05: `lsp_diagnostics(webhooks.go)` returned **empty**
     — the phantom did NOT reproduce this session (it is intermittent).
   - Related observation: gopls has been reporting a STALE
     `internal_test.go:4 "bytes" unused` error all session although the
     package compiles — the LSP layer (gopls AND golangci) sees mid-edit or
     differently-configured states. Best next probe: restart
     golangci_lint_ls with GOEXPERIMENT=jsonv2 added to the wrapper, then
     edit webhooks.go:142 region and watch for the diagnostic; and diff
     `golangci-lint cache clean` behavior between /mnt/buildcache and the
     CLI's default cache.
2. **pre-release full-chain run** — app built and its dirty-tree gate
   verified, but a complete end-to-end pass (through apidiff, which needs
   network) has not run on a clean tree yet.
3. **erraudit class-wide policy** — the substance is already recorded in
   AGENTS.md (2026-10-07 entries: embedded analyzer = upstream false
   positives, skip_steps'd, canonical standalone gate governs, never
   per-finding fixes). Missing: one explicit sentence framing it as the
   CLASS-WIDE disposition (the TODO's actual ask: "per-finding drift is the
   worst option").
4. **All-systems flake coverage** — `nix flake check --all-systems`
   evaluated: aarch64-linux and aarch64-darwin check derivations evaluate
   clean; **x86_64-darwin FAILS EVALUATION because nixpkgs-26.11 dropped
   that system** (external ecosystem change, not a repo defect). Blocked on
   a user decision: pin a `nixpkgs-26.05-darwin` input for darwin systems,
   or drop x86_64-darwin from the flake's system list.

## c) NOT STARTED

1. **AGENTS.md update for this session's work** — the memory-write sweep
   was interrupted by this report. Items owed: coverage-guard is now
   shuffle-proof via TestMain (replace the old "may SKIP under -shuffle"
   paragraph); new checks/apps (md-go-snippets, release-notes-check,
   pre-release) + the md-go-validator input/vendorHash/golines-override
   gotchas; RequestLog.RetryAfterDelay entry semantics; coverage now 93.1%;
   conformance actuals 180 exchanges / 6 statement variants; art-dupl
   policy decision; erraudit class-wide framing; x86_64-darwin
   all-systems blocker.
2. **TODO_LIST.md harvest** — mark the 11 completed P3 items done with
   pointers; rewrite the all-systems item with the x86_64-darwin finding;
   the three decision items get one-line resolutions pointing at
   AGENTS.md / the plan doc.
3. **Go 1.27 migration plan document** — designed but not written. Planned
   shape (docs/planning/2026-10-07_go-1.27-migration-plan.md): flip only
   when ALL of (1) nixpkgs-unstable's default `pkgs.go` is a stable 1.27.x
   line, (2) in-house deps (go-branded-id, go-error-family, go-retry)
   verified on 1.27 (apidiff + race + conformance green), (3) erraudit
   installs locally under the new toolchain (CI job becomes redundant),
   (4) buildflow's go-version-auto-configure + go-structure-linter
   skip_steps are REMOVED (their 1.27 push becomes correct, ending the
   flipflop war). Single-commit flip (flake pin + GOTOOLCHAIN policy +
   go.mod directive + docs), single-commit rollback. Today's standing
   workarounds (buildGoModule go_1_27 for md-go-validator; erraudit
   CI-only) are the documented cost of waiting.
4. **File the evidenced tool-repo defects** — BuildFlow three (embedded
   erraudit false positives at 202b114/3bb229e; `nix run .#reinstall` not
   switching ~/.nix-profile; file-size-check scanning 0 files in
   root-package layouts) **plus the new 4th**: md-go-validator's flake
   cannot build its own package (go pin above). None filed yet; all have
   captured evidence in AGENTS.md / this report.

## d) TOTALLY FUCKED UP (own mistakes this session, honestly)

1. **Broke Ginkgo nesting in wise_test.go** while inserting the
   Retry-After test — my Context landed OUTSIDE the WithLogger Describe and
   the "Retry cancellation" Describe lost its indentation (vet: undefined
   `entries`). My first scripted repair failed a pattern-match assertion
   (wasted a cycle), then two exact-match edits fixed it. Lesson: for
   structural insertions into Ginkgo containers, match on the enclosing
   container's closing brace, not the following header line.
2. **First coverage-guard fix was WRONG** — I replaced "skip if exchanges
   == 0 under shuffle" with the same emptiness check in TestMain; a 6-run
   shuffle loop caught a 1/6 failure because
   `TestSpecConformanceValidatorCanFail` records 1 exchange, so the guard
   can run mid-order with a PARTIAL count, not just zero. The correct fix
   (suiteRan flag set after RunSpecs) came only after that failure.
   Lesson: the vacuous condition was "suite not finished", not "nothing
   recorded" — I pattern-matched the old code instead of the failure mode.
3. **Small compile/toolchain stumbles** (each caught by tests/build, fixed,
   but avoidable with a verify-after-every-edit discipline):
   `f.Add(nil)` type error; an unused `"io"` import whose scripted removal
   didn't match (daemon raced the file) until sed; `go = pkgs.go_1_27`
   silently ignored inside buildGoModule (needed `.override { go = ... }`);
   vendorHash computed under go 1.26 not matching the go 1.27 vendor
   derivation; `errors.AsType[*APIError]` where 500 maps to `*ServerError`;
   shellcheck SC2012; forbidigo rejecting `fmt.Println/Printf` in TestMain.
4. **Lint was run too late** — I wrote ~450 lines of test code across four
   files before the first `golangci-lint run`, which then surfaced 6
   findings (2× forbidigo, err113, golines long line, nonamedreturns,
   wsl_v5) + 1 more after refactoring. All fixed to 0 issues, but each
   would have been cheaper immediately after the batch that introduced it.
5. **Edit-tool races with the auto-commit daemon** burned several retries
   ("file modified since read") — I should have gone straight to
   re-read-then-edit instead of attempting blind re-applies.

## e) WHAT WE SHOULD IMPROVE

1. **Verify-after-every-change, not after every feature** — the fuzz/bench
   batch and the lint batch each accumulated avoidable fix-up cycles.
2. **Negative-test every gate as part of "done"** — the broken-snippet and
   floor-99999 negative tests caught real gaps; make it a ritual for every
   new check (done this session for md-go-snippets and the coverage guard;
   release-notes-check got heuristic-level negative tests only).
3. **Structural edits need structural tools** — Ginkgo container insertion
   via text matching is fragile; prefer anchoring on unique closing braces
   or using lsp_replace_symbol-style boundaries.
4. **Shuffle/parallel robustness belongs in the first test run**, not a
   follow-up loop — one `go test -shuffle=on` early would have surfaced the
   partial-count race immediately.
5. **Environment parity between CLI gates and LSP gates** — the
   GOEXPERIMENT divergence (wrapper vs devShell) is the prime phantom
   suspect; host-managed wrappers should pin the repo-critical env.
6. **Record decisions the moment they are made** — art-dupl's policy was
   decided mid-session but not written down before this report; an
   interrupted session would have lost it.

## f) NEXT (up to 50, ordered by impact)

1. Write the AGENTS.md memory sweep (item c1 — highest leverage for future
   sessions).
2. Harvest TODO_LIST.md: mark 11 items done, update all-systems + decision
   items (c2).
3. Write docs/planning/2026-10-07_go-1.27-migration-plan.md (c3).
4. Record art-dupl policy + erraudit class-wide framing in AGENTS.md.
5. File the 3 BuildFlow defects + 1 md-go-validator defect (evidence in
   hand; use verify-before-filing + github-voice).
6. Run `nix run .#pre-release` end-to-end on the (now clean) tree — first
   full-chain validation including apidiff.
7. Root-cause the golangci_lint_ls phantom: add GOEXPERIMENT=jsonv2 to the
   wrapper, restart LSP, force the diagnostic (touch webhooks.go), diff
   formatter behavior between the wrapper's golangci-lint resolution and
   the devShell's.
8. Fix the stale gopls `bytes unused` phantom (restart gopls; if it
   persists after a clean build, file against gopls or crush LSP wiring).
9. Install/verify benchstat (`go run golang.org/x/perf/cmd/benchstat@...`)
   and round-trip the committed baseline through it once.
10. Decide x86_64-darwin: darwin-pinned nixpkgs input vs dropping the
    system (user decision; either unblocks `--all-systems` as a gate).
11. Wire `nix flake check --all-systems` (or the evaluable subset) into
    pre-release/CI once 10 is decided.
12. Re-run the six-shuffle suite loop after any future guard change (the
    1/6 catch proves single runs are insufficient).
13. Promote the coverage floor? 93.1% actual vs 90.0 floor — consider
    raising the enforced floor to 92% to lock the headroom in.
14. Add `nix run .#pre-release` mention to CONTRIBUTING's quality-gates
    checklist (it is documented in Release gates but not in the checklist).
15. Extend doc-verify to also gate `docs/releases/*.md` claim COUNTS in
    historical files against their TAGS (needs tag checkout; heavier).
16. Add corpus seeds to testdata/ for the three new fuzz targets (checked-in
    interesting inputs: empty array, single-object, malformed PEM, huge
    payload header).
17. Bench the rates single-object fallback path (legacy shape) for parity
    with the array bench.
18. Consider a `RequestLog` docs example showing the delay-decision entry
    (godoc) so consumers discover RetryAfterDelay.
19. Evaluate emitting the delay entry ONLY when a Logger is installed
    (already true) AND document that DelayFunc logging is client-wide, not
    per-request.
20. Re-check `.buildflow.yml` interplay: does buildflow's test step shuffle
    interact with the new TestMain guard as expected? (Run buildflow once
    and read `.summary.total`, not `findings`.)
21. After the next buildflow update, re-verify file-size-check still scans
    0 files (documented inert) and re-test the erraudit skip rationale.
22. Re-check md-go-validator releases: when upstream fixes its flake go
    pin, consider consuming `packages.default` instead of vendoring the
    build in checks.md-go-snippets.
23. VendorHash hygiene: note the go-1.27 vendorHash divergence in
    vendorHash.nix or a comment (done in flake comment; verify it
    survives nix fmt).
24. Consider pinning golangci-lint version in the devShell to match
    ci.yml's v2.13 (or bump ci.yml to 2.14) to kill the LSP/CLI version
    skew class.
25. Sweep remaining ~45 functions under 100% coverage for the next
    headroom push (quotes.go toWire/parseQuoteCreated error paths,
    ott.go ClearSCAChallenge branches, webhooks.go 415/430).
26. Add a regression test pinning that `Retry-After: 0` emits NO delay
    entry (currently only implied by the "on retries" test).
27. Consider a BDD test for TWO consecutive 429s (two delay entries,
    attempt numbers 1 and 2).
28. Sandbox live pass (P1) — still the release blocker; pairs with the OTT
    Q4 flip.
29. OTT 2026Q4 flip (P1) — blocked on sandbox credentials.
30. CI re-enable (P2) — then art-dupl warn-first job + erraudit blocking
    flip + coverage-badge unfreeze all become actionable.
31. When CI lands: add md-go-snippets-equivalent (md-go-validator) as a CI
    step so the gate also runs off-machine.
32. v1.0.0 tag (P2, user-gated).
33. Typed recipient Details decision (P2, user-gated).
34. CACHIX_AUTH_TOKEN / ERRAUDIT_TOKEN secrets (P2, user-gated).
35. GOEXPERIMENT ergonomics via direnv/home-manager (P2, user-machine) —
    would ALSO kill the LSP env-divergence class.
36. Add the CAMT content-type assertion to the spec-conformance statement
    variant tally documentation (FEATURES.md statement-format list if it
    enumerates formats).
37. Check FEATURES.md/CHANGELOG need entries for: Retry-After log line
    (observability feature), new fuzz/bench coverage, pre-release app —
    fold into the next release cycle's changelog section now.
38. Sweep README for a mention of the new pre-release command (release
    ritual section if present).
39. Re-run `nix flake lock` freshness check (md-go-validator input just
    added; confirm narHash pinned).
40. Confirm the daemon's commits contain what I wrote (AGENTS gotcha:
    verify a daemon commit's message vs content for the session's sweep).
41. Consider extracting the "latest release notes" helper in doc-verify
    into a shared snippet if a third consumer appears (YAGNI today).
42. Add `docs/bench/README.md` one-pager (what the files are, how to
    regenerate) if the dir grows past one baseline.
43. Evaluate `-count=10` baselines on the next regeneration (6 runs gave a
    noisy ParseWebhookEvent spread: 1262–1750 ns/op).
44. Look into the ParseWebhookEvent bench variance (JSON decode of a
    multi-line payload; -benchtime tuning may stabilize).
45. Add fuzz target for `parseWiseTimestamp`'s sibling `parseWiseDate`
    (statement dates) if not covered by the existing timestamp fuzz.
46. Extend `FuzzNewCurrency` with the Money-currency pairing path
    (`toMoney`) if corrupt-currency fixtures aren't already table-covered.
47. Check whether `readBody`'s size-limit branch (if any) is covered; the
    unreadable-body test covers the error branch only.
48. Document the md-go-validator scope decision (living docs + docs/ tree;
    CHANGELOG has 0 Go blocks) in the check's nix comment (done) and
    AGENTS.md (pending c1).
49. Revisit `minExemptStatementVariants` after the sandbox pass — if Wise
    adds formats, the floor message enumerates csv/pdf/xlsx/xml only.
50. Schedule the next dedup pass (art-dupl manual, per policy) after the
    next feature PR lands.

## g) QUESTIONS (cannot self-answer)

1. ~~**art-dupl policy**: I decided "accept the suppressed baseline as
   policy; no enforced gate; revisit at CI re-enable (warn-first job)".
   Confirm — or do you want the exit-code gate wired into buildflow/CI
   NOW (as continue-on-error) despite CI being disabled?~~ RESOLVED
   2026-10-07 (P3-resume session): decided as proposed — baseline is
   policy, no enforced gate, revisit warn-first at CI re-enable.
   Recorded in AGENTS.md (art-dupl bullet) and TODO_LIST Closed 2026-10-07.
2. **x86_64-darwin**: nixpkgs-26.11 dropped the system, so
   `nix flake check --all-systems` can never pass against
   nixpkgs-unstable. Pin a `nixpkgs-26.05-darwin` input for darwin
   outputs, or drop x86_64-darwin from the flake's supported systems?
3. **RequestLog.RetryAfterDelay API shape**: pre-v1.0-freeze, is the
   delay-decision entry on the existing `Logger`/`RequestLog` channel
   (Method/URL empty, Status=429) acceptable to lock in for v1.0, or do
   you want a dedicated observation interface (e.g. optional
   `DelayObserver`) before the surface freezes?

---

*Report: 2026-10-07 06:07 · suite green · lint 0 issues · tree clean ·
coverage 93.1% · conformance 180 exchanges / 6 statement variants.*
