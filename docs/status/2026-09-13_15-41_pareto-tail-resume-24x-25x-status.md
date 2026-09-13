# Status Report — Pareto Tail Resume: 23.5/24.x/25.x + Ops Tooling Batch

**Generated:** 2026-09-13 15:41 CEST
**Scope:** this resumption session only (resumed after the 14:16 full-execution report; the user said "Execute and Verify them one step at the time … Keep going until everything works").
**Report format:** `.md` per explicit user instruction (the status-report skill's canonical format is HTML; one-off override, not propagated into the skill).

---

## Executive snapshot

| Gate | Result |
| --- | --- |
| `go test -race -count=1 ./...` | ✅ ok (root + `internal/raw`, 201 top-level test funcs) |
| `golangci-lint run` | ✅ 0 issues |
| `nix flake check` | ✅ all checks passed (sandboxed race/coverage test + README links) |
| `nix run .#doc-verify` (new) | ✅ godoc renders, count-claims match, lychee 60 OK / 0 errors |
| `nix run .#apidiff` (new) | ✅ vs `v0.10.0`: **all-additive**, suggests v0.11.0, 0 removed/incompatible |

**The single biggest discovery:** `v0.10.0` **is already tagged and pushed** (local `git tag`, module proxy `@v/list`, and gorelease's `-base=latest` all confirm). The prior session's report and the plan carried "v0.10.0 untagged, tag on approval" as a live blocker — that premise is **stale**. What remains is only the GitHub **Release object** (`gh release list` still shows v0.9.0 as Latest).

---

## a) FULLY DONE (evidence-backed)

1. **Count-drift fix across living docs** — surface is verifiably **37 `*Client` methods / 107 exported types / 25 const groups / 15 resources / 22 `Example*` funcs** (measured via `go doc -all` + `grep -c '^func Example'`). Fixed: `AGENTS.md` API-docs bullet (33→37, 14→15, re-audit parenthetical now notes same-day webhook growth), `FEATURES.md` (20→22 examples, 14→15 resources, deferred-row 14→15), `ROADMAP.md` Axis-1 block (33→37, resource list gained "webhooks", 2026Q3 note), `docs/reviews/2026-08-21_v1.0-api-audit.md` (blockquote delta note: +4 methods/+15 types/+3 const groups; section header; webhook growth paragraph). README "Project Status" checked — carries no counts, no drift. Point-in-time reports and the plan's historical baselines deliberately untouched.
2. **CHANGELOG `[Unreleased]` micro-batch entries** — Added: `WithUserAgent`, `Profile.UserID`/`PublicID`, concurrency-pin note; Changed: internal batch (fetchByID branded-ID routing, raw-valued mapper errors, SCA error context, benchmarks, raw round-trip tests). House style matched to the v0.9.0 "Internal deduplication" entry.
3. **`docs/DOMAIN_LANGUAGE.md` webhook vocabulary** — Glossary row for the 2026Q3 surface; Entities (extended) rows for Subscription / Delivery / Envelope (verified header constants `X-Signature-SHA256` / `X-Delivery-Id` before writing); Value Objects row for the open `WebhookEventType` enum; mapper list extended with `toWebhookSubscription`/`toWebhookResource`.
4. **Task 23.5 — client concurrent-safety test** — `wise_test.go` "Concurrency" Describe: 16 goroutines, one shared `*Client`, per-goroutine correlation ID echoed into a per-goroutine rate value (any context/response cross-talk fails the assertions); server-side seen-map proves 16 distinct IDs. Passes under `-race`, stable across 3 separate `-count=1` runs (Ginkgo forbids `go test -count=N` — first `-count=3` "failure" was that, not the test).
5. **Tasks 24.1+24.2 — test-file split** — `internal_test.go` (2254 lines, `package wise`) split by concern via line-range extraction + `goimports`: `errors_test.go` (741 lines: classification, contexts, checkError/SCA, exhausted retries, corruption-classification suite, `roundTripFunc`, retry-after parsing), `helpers_test.go` (502 lines: timestamps, money/currency, classifier, mapper edge cases, both fuzz targets), remainder `internal_test.go` (1033 lines: ID guards, request validation, wire tests, webhook signature tests, transport edges). `expectRejection` stayed with its 6 validation-test callers. Build + vet + all 201 tests green; `nix fmt` clean; lint 0 issues.
6. **Flake fileset repaired for the split** — `errors_test.go`/`helpers_test.go` added to `flake.nix` `sourceFiles` (the AGENTS.md "fileset must list all Go files" gotcha — without this the sandboxed check would have silently tested a smaller surface than local).
7. **Task 24.3 — coverage gate in CI** — `.github/workflows/ci.yml` "Coverage threshold" step after Test: fails under 90% (measured 91.4% on today's tree; ~1.4pt headroom documented). Gate math unit-checked both directions locally (91.4 → pass, 89.5 → fail); YAML parses.
8. **Task 24.4 — `nix run .#apidiff`** — new flake app wrapping `gorelease -base=latest` (GOEXPERIMENT injected; documented as network-dependent, hence an app and never a sandboxed check). **Verified end-to-end against the clean /tmp snapshot: 0 removed/incompatible/changed entries — the unreleased delta is purely additive; gorelease suggests v0.11.0.** Also caught and fixed the flake fileset gap while preparing gates.
9. **Tasks 25.1+25.2 — contributor templates** — `.github/ISSUE_TEMPLATE/bug_report.yml` (sandbox-vs-live dropdown, repro required, correlation-ID guidance, redaction checkboxes), `feature_request.yml` (endpoint-link field, house-style API-shape prompt), `.github/PULL_REQUEST_TEMPLATE.md` (checklist mirrors the real gates incl. the fileset gotcha). Both YAMLs validated.
10. **Tasks 25.3–25.5 — ADR directory** — `docs/adr/001-money-value-object-no-arithmetic.md` (Accepted), `002-flat-package-raw-boundary.md` (Accepted), `003-retry-executor-go-retry-override.md` (**Proposed** — see d/e for why this is the honest shape). ADR 003 grounds every claim: failsafe-go present since `9327c5e`, the how-to-golang mandate (rules.md:86), the probe result quoted from TODO_LIST.
11. **Task 25.6 — `nix run .#doc-verify` + reports cleanup** — flake app: godoc render check, count-claims freshness (AGENTS.md methods vs `go doc -all`; FEATURES examples vs `grep -c`), lychee offline links. The example-count pattern initially matched **nothing** (silently-skipped check) — fixed to `[0-9]+ .Example.. funcs`, verified it extracts the real "22", and the FAIL branch proven via standalone comparison test. Stale `reports/` (jscpd report from Jul 18, coverage.out from Aug 21) trashed; nothing in living docs referenced them.
12. **TODO_LIST brought current** — P1 v0.10.0 item re-marked (tag exists; only Release object left, BLOCKED on approval); CI re-enable item rewritten (auth blocker resolved 2026-09-13; remaining = push + enable, user-gated); go-retry split into a done `[x]` rationale-record (ADR 003) + still-open `[ ]` migration BLOCKED on ADR acceptance; two new `[x]` batch rows for this session's infra work.
13. **AGENTS.md Build & Dev** — `apidiff` and `doc-verify` documented as the two new local gates.
14. **Final full-gate run after ALL changes** — race ✅, lint 0 ✅, `nix flake check` ✅ (run twice this session: after the split, and after the doc-verify app), `doc-verify` ✅.

## b) PARTIALLY DONE

1. **docs-health HARVEST of the 14:16 report's §f list** — this session manually updated TODO_LIST for *its own* completed items only. The previous report's ~50 next items were never formally harvested into TODO_LIST/ROADMAP. Remainder: run HARVEST over both reports' §f lists. Blocker: none. Effort: M.
2. **ADR 003 → go-retry migration** — rationale recorded (Proposed), migration untouched by design (the plan asked only for the rationale record). Remainder: user accepts/rejects ADR, then swap executor, delete `classifyExhaustedRetries`, wire `Retry-After` through `Config.DelayFunc`, gate on the 429 BDD tests. Blocker: user decision. Effort: M.
3. **Coverage-gate local parity** — the 90% threshold lives only in `ci.yml` (disabled on GitHub). The sandboxed `checks.test` measures coverage but enforces no floor. Remainder: mirror the threshold in `flake.nix` checkPhase. Blocker: none. Effort: S.
4. **v0.10.0 release** — tag ✅ + pushed ✅ + proxy ✅ + notes drafted ✅; GitHub Release object missing. Blocker: user approval to publish. Effort: S.
5. **CI enablement** — workflow file fully refreshed (v2.13 pin, no-auth nix job, coverage gate), auth blocker eliminated, but master is unpushed-by-policy and the workflow remains `disabled_manually` server-side. Blocker: user push/enable approval. Effort: S + watch first run.
6. **CONTRIBUTING.md currency** — AGENTS.md documents the two new apps; CONTRIBUTING (the contributor-facing doc) does not mention `apidiff`/`doc-verify` or the 90% gate. Effort: S.

## c) NOT STARTED (still wanted, waiting on decision/key/priority)

1. **App-level webhook subscriptions** (client-credentials token model + `TestWebhookSubscription`) — deferred 2026-09-13, ROADMAP entry exists; needs the token-model decision. *Also note: `deposits#*` payload types were substituted by verified events last session; if Wise documents them later, revisit.*
2. **Credentialed sandbox live run** — `sandbox_live.yml` + `sandbox_live_test.go` key-drop-ready; needs `WISE_SANDBOX_API_KEY`.
3. **v1.0.0 tag** — audit green (re-audit + same-day delta), everything else done; only the user-gated tag remains.
4. **CACHIX_AUTH_TOKEN secret** — cachix step stays `continue-on-error: true` until then.
5. **GOEXPERIMENT direnv/home-manager pinning** — user-machine ergonomics; workaround documented.
6. **Typed recipient `Details`** — carried unanswered through five+ status reports; v1.0 can freeze the map.
7. **Tier-3 API expansion** — the full-implementation plan's remaining tiers (per the 2026-08-19 plan + v1.0 audit).

## d) TOTALLY FUCKED UP (radical honesty)

Nothing this session broke the build, lost data, or shipped wrong behavior — all gates are green and every claim above is command-verified. But three things were genuinely bad, and one inherited item is still live:

1. **I shipped a check that could not fail — and initially trusted its green.** The first `doc-verify` run passed while the FEATURES example-count extraction matched zero text (claimed="" → comparison skipped). This is precisely the pipeline-masking failure class the global AGENTS.md warns about (`set -o pipefail`, filters that hide failures), and I briefly repeated it anyway. A gate that silently skips is worse than no gate. **Mitigation done:** pattern fixed, real-value extraction verified for BOTH counts, FAIL branch proven. **Residual risk:** the methods-count check would also silently skip if AGENTS.md's phrasing ever changes ("endpoint methods" regex). Hardening idea in (f).
2. **I propagated a stale premise from the previous session's summary into my opening plan.** I resumed believing "v0.10.0 untagged, tag on approval" and shaped the opening todo list around it; the tag's existence only surfaced mid-session via gorelease (`Base version: v0.10.0`) and was then verified three ways and corrected everywhere (TODO_LIST, final summary). The lesson already on file — *status reports are point-in-time; re-verify before treating as truth* — applied to the SUMMARY, not just to old reports. Cost: ~10 minutes of misdirected planning, no wrong output shipped.
3. **ADR 003 was first written to contradict the still-open TODO_LIST item.** My first draft declared "no migration warranted / nothing to migrate," which would have misled the next session into closing the go-retry task — the exact opposite of the plan's intent (record the override rationale *for a pending migration*). Caught mid-session by cross-checking TODO_LIST P4 before finalizing; rewrote as Proposed; renamed via `git mv`. The near-miss itself is the finding: I wrote a decision record before re-reading the decision log.
4. **(Inherited, live) LSP diagnostics are misleading on this machine** — ~70 warnings (gopls `stdversion` go1.27 complaints, tagliatelle/wsl_v5/gci in files the real CLI passes) persist in every tool result. I verified CLI-vs-LSP divergence once (`golangci-lint run` → 0 issues) and ignored the noise, but every future session pays this tax again. Not mine to fix unilaterally (editor/LSP config), listed in (f).
5. **(Minor, self-inflicted)** First `lint` after the split-flagged `makezero`/`varnamelen`/`wsl_v5`/`modernize` issues in my own new concurrency test — I knew this repo's lint profile and still wrote the non-idiomatic version first. Also typo'd `rm is banned` onto a command line (harmless; `rm` refused to remove files named "is" and "banned"). Sloppiness tax, both corrected same-session.

## e) WHAT WE SHOULD IMPROVE

1. **Gate-writing discipline: "prove the gate can fail" before trusting "all checks passed."** Concrete rule: every new check gets a negative test (wrong value → expect FAIL) and a positive extraction assertion (what value did you actually extract?). This session's doc-verify bug and the earlier `--all-features`/`head -5` masking incidents are the same class. Candidate: encode as a bullet in AGENTS.md + a step in the linter-building/quality-scan skills.
2. **Session-resume hygiene: re-verify the previous summary's factual claims (tags, releases, branch state) before planning around them.** A 30-second `git tag`/`gh release list`/proxy check at resume time would have caught the v0.10.0 staleness before any planning. Candidate: add to the resume ritual (AGENTS.md cross-cutting lessons).
3. **Decision records must be written against the decision log.** Read TODO_LIST P4 *before* drafting an ADR that references pending work. The ADR-003 rewrite cost 15 minutes and a rename.
4. **gorelease tool pinning** — `apidiff` runs `golang.org/x/exp/cmd/gorelease@latest`: reproducibility smell in a repo that pins everything else (golangci v2.13, action SHAs). Pin to a version and note the bump policy.
5. **Coverage-gate parity** — mirror the 90% floor into the sandboxed `checks.test` so the gate exists locally, not only in a CI file that is currently disabled.
6. **LSP config skew** — gopls ignores/mismatches the repo's GOEXPERIMENT+lint config and cries wolf 70×/session. Worth one dedicated fix (gopls env injection or `gopls` settings in `.golangci.yml`-adjacent config) so future sessions stop re-litigating CLI-vs-LSP.
7. **Count-claims belong in ONE doc.** The 37/15/22 drift spanned four files because the same numbers live in AGENTS.md, FEATURES.md, ROADMAP.md, and the audit doc. `doc-verify` now guards two of them; consider deriving all four docs' counts from a single generated block (a small `go run` doc-gen or expanding doc-verify to check ROADMAP/audit too).
8. **Ginkgo + `go test -count=N` trap** — cost one false-FAILURE investigation. One line in CONTRIBUTING's Testing section ("repeat via separate `-count=1` runs or `ginkgo -repeat=N`") prevents a repeat.

## f) Up to 50 things we should get done next

> Brainstorm ranked by impact; the top ~12 are TODO_LIST-grade, the rest are ROADMAP fuel for docs-health HARVEST routing. Effort: S <30min, M 30min–2h, L >2h.

| # | Task | Impact | Effort | Category |
| --- | --- | --- | --- | --- |
| 1 | Publish the v0.10.0 GitHub Release from the drafted notes (BLOCKED: approval) | Critical | S | Release |
| 2 | Push master + `gh workflow enable ci` + watch first green run + confirm badge unfreezes (BLOCKED: approval) | Critical | S | Release/CI |
| 3 | Answer the app-level webhook question (client-credentials now vs keep ROADMAP deferral) | High | S | Decision |
| 4 | Cut v0.11.0 (webhooks + WithUserAgent + Profile fields) after 1–3 — gorelease pre-verified all-additive | High | S | Release |
| 5 | Accept/reject ADR 003, then execute the go-retry migration (swap executor, delete `classifyExhaustedRetries`, `Retry-After` via `DelayFunc`, 429 BDD as gate) | High | M | Feature/Quality |
| 6 | Mirror the 90% coverage floor into `checks.test` checkPhase (local parity with the CI gate) | High | S | Quality |
| 7 | Pin gorelease to a fixed version in the `apidiff` app (kill `@latest`) | Medium | S | Quality |
| 8 | docs-health HARVEST: route both reports' §f lists into TODO_LIST/ROADMAP | High | M | Docs |
| 9 | Re-verify `docs/releases/v0.10.0-release-notes.md` against the now-known state before publishing | Medium | S | Docs |
| 10 | Verify the v0.10.0 tag object points at the intended commit (`fe896a8`) | Medium | S | Release |
| 11 | Extend doc-verify: ROADMAP + audit-doc count claims, and fail loudly when a claimed-count pattern extracts EMPTY (kill the silent-skip class) | High | S | Quality |
| 12 | Wire `apidiff` + `doc-verify` into CONTRIBUTING.md (contributor-facing gates) | Medium | S | Docs |
| 13 | Add a "prove the gate can fail" rule to AGENTS.md cross-cutting lessons (from d1) | High | S | Docs/Process |
| 14 | Add "re-verify prior-session facts (tags/releases/branch) at resume" to AGENTS.md lessons (from d2) | High | S | Docs/Process |
| 15 | Ginkgo repeat syntax note in CONTRIBUTING Testing section | Low | S | Docs |
| 16 | Fix gopls/LSP config skew (GOEXPERIMENT env, lint config) so tool-result warnings stop lying | Medium | M | Tooling |
| 17 | Sandbox live run with `WISE_SANDBOX_API_KEY` (unblocks the whole sandbox-live lane) | High | S | Feature |
| 18 | Set `CACHIX_AUTH_TOKEN` + confirm the `larsartmann` cache exists; flip cachix step to fail-hard | Medium | S | CI |
| 19 | `TestWebhookSubscription` (spec: app-level only) after #3 | Medium | M | Feature |
| 20 | Typed recipient `Details` decision (5+ reports unanswered; v1.0 wants it settled) | High | S | Decision |
| 21 | Tag v1.0.0 after #1/#2 land and gates stay green (audit is done) | High | S | Release |
| 22 | Draft the v0.11.0 release-notes skeleton early (webhooks, WithUserAgent, Profile fields, internal batch) | Medium | S | Docs |
| 23 | CI job running `apidiff` on release-prep PRs (once CI is enabled) | Medium | S | CI |
| 24 | CI job running `doc-verify` on docs-touching PRs | Medium | S | CI |
| 25 | README docs table: add `docs/adr/` index + status-report pointers | Low | S | Docs |
| 26 | Cross-link ADRs from code: AGENTS.md "Money" and retry bullets should cite ADR 001/003 | Low | S | Docs |
| 27 | Benchstat baseline file committed + workflow doc updated (bench marks exist; no baseline to diff against) | Low | S | Quality |
| 28 | Investigate `ParseWebhookEvent` 34.5µs / allocation profile (highest-latency bench) | Medium | M | Quality |
| 29 | Webhook end-to-end README quickstart (subscribe → verify → parse in one runnable block) | Medium | M | Docs |
| 30 | Godoc example for `VerifyWebhookSignature` + `ParseWebhookEvent` composition (formalizes the declined `VerifyAndParse` helper) | Low | S | Docs |
| 31 | Verify ListProfileWebhookSubscriptions pagination shape against spec (assumed single response) | Medium | S | Quality |
| 32 | Replay/idempotency guidance for webhook consumers (dedup on `X-Delivery-Id` — doc the at-least-once semantics) | Medium | S | Docs |
| 33 | Raise the coverage gate to 91% once webhook payload tests mature | Low | S | Quality |
| 34 | Add `.github/ISSUE_TEMPLATE/config.yml` (blank-issues toggle, discussion link) | Low | S | Ops |
| 35 | Add `.github/SECURITY.md` (vulnerability reporting path — a published SDK should have one) | Medium | S | Ops |
| 36 | Review dependabot.yml coverage (does it watch workflows + Go modules?) | Low | S | Ops |
| 37 | aarch64-darwin/linux flake system coverage (flake check warns "omitted systems") | Low | M | Tooling |
| 38 | Probe `CreateUnauthenticatedQuote` as a keyless live smoke test (would partially unblock live-path confidence without a key) | Medium | S | Feature |
| 39 | Sweep FEATURES.md for remaining unaudited numeric claims beyond the two doc-verify guards | Low | S | Docs |
| 40 | ROADMAP: fold gorelease's v0.11.0 suggestion into the release-planning note | Low | S | Docs |
| 41 | DOMAIN_LANGUAGE: retry/exhaustion vocabulary after ADR 003 lands | Low | S | Docs |
| 42 | CHANGELOG convention: decide whether contributor-infra changes get a standing "Internal" subsection | Low | S | Docs |
| 43 | `gitignore`/residue sweep: confirm no dangling `reports/` references in tooling configs | Low | S | Cleanup |
| 44 | Tier-3 API expansion kickoff (next Pareto pass over the ~135-endpoint plan) | Medium | L | Feature |
| 45 | Webhook signature fuzzing (malformed PEM/signature inputs beyond the edge tests) | Low | S | Quality |
| 46 | Consider goroutine-leak check (`goleak`) in the suite, given the new concurrency surface | Medium | S | Quality |
| 47 | Add sandbox-live dispatch smoke to the workflow (dispatch-gated, key-gated) | Low | S | CI |
| 48 | Property test: `formatWiseTimestamp` round-trips `parseWiseTimestamp` for UTC inputs | Low | S | Quality |
| 49 | Audit `example_test.go` examples still compile-only → mark the 3 runnable ones in README | Low | S | Docs |
| 50 | Rename ADR 003 title/filename consistency check in doc-verify (status line ↔ filename) | Low | S | Docs |

## g) Three questions I cannot answer myself

1. **Release shape:** publish the v0.10.0 GitHub Release now with the drafted receipt/payout-info notes, and fold webhooks + `WithUserAgent` + `Profile.UserID/PublicID` into a separate v0.11.0 — or skip publishing v0.10.0's release page entirely and cut one combined v0.11.0 release? I verified v0.10.0 is already tagged/pushed/proxy-listed and gorelease confirms the unreleased delta is all-additive (suggests v0.11.0), but whether to backfill a Release object for an already-proxied tag, or let it ride, is a release-strategy call only you can make.
2. **CI go-live:** approve pushing master and `gh workflow enable ci` now (first run will also unfreeze the coverage badge and exercise the new 90% gate + no-auth nix job), or keep shipping on local gates until v0.11.0 is ready so CI's first run sees the full webhook surface? Everything on my side is prepared and verified; the push/enable itself is gated on you by harness policy.
3. **ADR 003 (go-retry):** accept and let me execute the failsafe-go → go-retry migration now (executor swap, `classifyExhaustedRetries` deletion, `Retry-After` honored via `DelayFunc`, 429 BDD suite as the regression gate), or hold it until after v0.11.0 ships? I cannot weigh the timing against your release cadence — and I will not override a Proposed ADR without your acceptance.

---

**Per the skill contract:** section (f) is HARVEST input — TODO_LIST.md already carries this session's completed rows; the 50 items above should be routed (top ~12 → TODO_LIST, rest → ROADMAP) when you say go. Nothing was committed manually; the auto-commit daemon owns commits (per harness rule). **WAITING FOR INSTRUCTIONS.**
