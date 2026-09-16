# Status Report — FEATURES.md Endpoint-Coverage Matrix Audit

- **Written:** 2026-09-16 16:35 CEST
- **Session scope:** Full audit of the live Wise Platform API reference
  (`docs.wise.com/api-reference/preview`) against wise-go, and a complete
  FEATURES.md rewrite into a per-endpoint coverage matrix. Nothing else was
  touched on purpose.
- **Format note:** `.md` per explicit user instruction (skill default is HTML).

---

## TL;DR

The Wise API preview reference documents **215 REST operations across 51
categories** (plus 29 event definitions in-category; 31 distinct event types
with the webhook-event index). wise-go ships **41** of those operations through
**38 of its 41 `*Client` methods**; **51 operations are demand-gated PLANNED**;
**123 are OUT_OF_SCOPE**. FEATURES.md now lists every single endpoint with
method, path, status, and evidence. The `doc-verify` gate is green. The test
suite is **red (48 failures)** from a *parallel session's* in-progress OpenAPI
conformance harness — unrelated to this session's docs-only change.

---

## a) FULLY DONE

| # | Item | Evidence |
|---|------|----------|
| 1 | Live preview surface enumerated completely: 51 categories, every operation page collected (including the two transfer sub-tabs `standard-transfer` / `third-party-transfer` that the aggregate transfer page hides — found via sitemap) | session fetch sweep; `docs.wise.com/sitemap.xml` |
| 2 | SDK surface enumerated: 41 `*Client` methods (39 endpoint + `Authenticate`/`Health`; `TriggerOTT`/`VerifyOTT` each fan out to 3 channel paths; statement endpoint served by 2 methods) | `rg '^func \(c \*Client\)'` over non-test sources |
| 3 | Coverage mapping built and cross-checked: **215 = 41 shipped + 51 PLANNED + 123 OUT_OF_SCOPE** | FEATURES.md summary bullets + per-category tables |
| 4 | FEATURES.md rewritten: full endpoint matrix (92 method-shaped rows + packed batch/address/single-op rows), every shipped row cites `file:line` | `FEATURES.md:21-` ("Wise API endpoint coverage") |
| 5 | Status vocabulary extended with `OUT_OF_SCOPE` (defined honestly: specialised / partner-only / sandbox-only) | `FEATURES.md:6-14` |
| 6 | AGENTS.md updated: stale "~135+ endpoints across 29 categories" replaced with the 2026-09-16 audit numbers + pointer that FEATURES.md is THE endpoint-inventory source; added the "category page fetchable as `<category>.md`" discovery | `AGENTS.md:59` |
| 7 | Verification: `nix fmt` clean; `nix run .#doc-verify` **all checks passed** (63 links OK / 0 errors, `go doc` renders, count claims match compiled surface: 41 methods, 24 `Example*` funcs) | gate output this session |
| 8 | Mechanical row-count proof: 92 endpoint rows = 41 FULLY + 32 PLANNED + 19 OUT; packed rows add 7 (batch) + 12 (single-op) PLANNED and 104 OUT → totals reconcile exactly | session grep verification |
| 9 | Work committed by the auto-daemon with content verified intact (`FEATURES.md`, `AGENTS.md`, plus a foreign test file swept in) | commit `94f3a3f` |
| 10 | Foreign-file discipline: discovered `spec_conformance_test.go` / `zz_spec_conformance_coverage_test.go` modified by a parallel session mid-run; inspected, judged, **not touched** | `git status`, file headers |

## b) PARTIALLY DONE

| # | Item | Gap |
|---|------|-----|
| 1 | **Docs-health loop closed?** No — the 51 PLANNED operations exist only in FEATURES.md. Section (f) below was NOT harvested into `TODO_LIST.md`/`ROADMAP.md` (waiting for your instructions, per your command) | HARVEST pending |
| 2 | Webhook-event reconciliation: documented the 33-constants vs 31-live-events divergence with sources and dates, but did **not** reconcile them 1:1 (e.g. where `swift-in#credit` / `swift-message-received` live in the *preview* nav was never located; the preview sitemap section was truncated) | `FEATURES.md` Webhook events subsection |
| 3 | Pre-existing-failure proof: established the 48 failures come from the foreign harness (`attachConformance`, `wise_test.go:172/186`), but did not prove they pre-date this session by running the suite at the pre-harness commit in a throwaway worktree (would have been non-destructive) | test output: all failures are `[AfterEach] conformance.finish` schema mismatches |
| 4 | Local spec freshness: confirmed `docs/reviews/wise-api-openapi.json` is 2026-08-08-era (173 paths / 209 methods vs 215 live ops) and noted in AGENTS.md that new endpoints must be checked against the live reference — but did not refresh the local spec snapshot | `AGENTS.md:59` |
| 5 | CHANGELOG: the FEATURES.md overhaul is not recorded there (docs-only change; also the changelog-only-commit structural trap argues for folding it into the next code-touching commit) | — |

## c) NOT STARTED

| # | Item | Note |
|---|------|------|
| 1 | Any of the 51 PLANNED endpoints — zero implementation begun (correct: this was an audit task) | biggest-value candidates: quote PATCH, balance-movements, recipient deactivate/compatibility |
| 2 | Legacy-surface decision: the `api-reference/legacy` sitemap section revealed ~20+ operations that exist ONLY in legacy (e.g. `profilecreatev1`, `profileextensions*`, `balancegetv1`, `balanceconvertv1`, `ottstatusgetv1`, `ottpinverify`, `cardpermissionsget/update`, `cardphonenumberupdate`, `usersecurity*` 9 ops, `kycreviewgetv1`, `verificationuploadevidencesv3`). The matrix covers the **preview** surface only — per your instruction — but these were noticed and are not even mentioned in FEATURES.md as a caveat | needs your scope call (Question 1) |
| 3 | Pinned evidence snapshot of the live reference under `docs/reviews/` (so FEATURES.md numbers are checkable offline later) | — |
| 4 | doc-verify extension to gate the FEATURES.md coverage totals (41/51/123) against reality, like the method/Example counts | — |
| 5 | Everything else in the repo's TODO_LIST/ROADMAP (releases, CI re-enable, v1.0.0 tag, sandbox key, …) — untouched, as instructed | see (f) |

## d) TOTALLY FUCKED UP

Nothing this session produced is destroyed or backwards — but three things are genuinely bad right now:

1. **The main test suite is RED: 154 passed / 48 failed.** Every failure is the parallel session's half-finished conformance harness validating httptest fixtures against `docs/reviews/wise-api-openapi.json` (e.g. `GET /v1/rates` — "got object, want array"; schema resolves to `null`). Until that harness is fixed or gated, `nix flake check` and any release gate that runs tests will fail. Not authored here, but it is the repo's current state.
2. **Commit-history attribution is mangled by the daemon:** `94f3a3f` ("auto-commit 3 changed file(s) (heuristic)") bundles my FEATURES.md+AGENTS.md rewrite with a foreign `spec_conformance_test.go` change under a meaningless message. The docs milestone is invisible in history.
3. **My first FEATURES.md draft shipped 3 wrong totals** (43 wire endpoints / 39 methods / 49 planned; Quotes summary row said 4/4). Caught during mechanical verification *after* writing, fixed pre-commit — but the right order is count-first-then-write.

## e) WHAT WE SHOULD IMPROVE

1. **Count mechanically before writing count-heavy docs.** The grep-row verification took one command; doing it first would have avoided the 3-error draft entirely.
2. **Close the HARVEST loop in the same session.** Section (f) items belong in `TODO_LIST.md`/`ROADMAP.md`; right now they live only in this snapshot.
3. **Prove "pre-existing failure" cheaply and rigorously** (worktree checkout of the last green commit, run suite, discard) instead of asserting it from file inspection.
4. **Pin evidence.** A snapshot of the live preview reference in `docs/reviews/` would make the matrix's 215/41/51/123 claims independently verifiable and let doc-verify gate them.
5. **Surface legacy-vs-preview divergence explicitly in the doc** (one caveat line minimum) once you decide the canonical target surface.
6. **Daemon interplay:** check `git log` immediately before and after manual work (done here), and prefer explicit commits for milestones when you authorize them, so heuristic daemon commits don't swallow them.

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

*Brainstorm ranked by impact; (S)=this session's findings, (R)=already in TODO_LIST/ROADMAP. Items 1–3 need your answers to (g) first.*

| # | Task | Impact | Effort | Origin |
|---|------|--------|--------|--------|
| 1 | Decide canonical surface (preview-only vs include legacy-only ops) and amend FEATURES.md caveat/matrix accordingly | High | XS | S |
| 2 | Pick the next PLANNED resource to implement (my recommendation: quote PATCH + recipient deactivate/compatibility — completes the core money-movement flow) | High | M | S |
| 3 | Coordinate the conformance harness: fix or `t.Skip`-gate the 48 red failures so the suite is green again | High | M | S |
| 4 | HARVEST (f) into TODO_LIST/ROADMAP per docs-health | High | S | S |
| 5 | Implement `PATCH /v3/profiles/{id}/quotes/{id}` (update quote with recipient) | High | S | S |
| 6 | Implement `DELETE /v1/accounts/{id}` (recipient deactivate; 403-on-inactive pinned) | High | S | S |
| 7 | Implement recipient compatibility check (`POST /accounts/{id}/quotes/{id}/compatibility`) | High | S | S |
| 8 | Implement `POST /v4/profiles/{id}/balance-movements` (balance↔balance conversion) | High | M | S |
| 9 | Implement `GET /v1/transfers/{id}/payments` (funding payments list) | Med | S | S |
| 10 | Implement recipient confirmations accept (`PATCH /accounts/{id}/confirmations`) | Med | S | S |
| 11 | Implement balance close (`DELETE /v4/profiles/{id}/balances/{id}`) | Med | S | S |
| 12 | Implement transfer NOC document (`documents/noc`, India FIRC demand) | Med | S | S |
| 13 | Implement US combined receipt PDF | Low | S | S |
| 14 | Implement profiles write: personal/business create + update (4 ops) | High | M | S |
| 15 | Implement addresses (5 ops incl. dynamic requirements discovery) | Med | M | S |
| 16 | Implement bank-details ordering (issue details, orders create/list, returns) | Med | M | S |
| 17 | Implement MCA configuration endpoints (available/payin currencies, eligibility) | Med | S | S |
| 18 | Implement batch groups (7 ops) — tier 3 bulk payments | Med | L | S |
| 19 | Implement activity feed (`GET /profiles/{id}/activities`) | Med | S | S |
| 20 | Implement contacts (`POST /profiles/{id}/contacts` Wisetag/email/phone lookup) | Med | S | S |
| 21 | Implement comparison (`GET /comparisons`) — the one tier-2 op still missing | Med | S | S |
| 22 | Implement payin deposit details (bank-transfer pay-in instructions) | Med | S | S |
| 23 | Implement GPI tracking lookup (`POST /v1/gpi-tracking`) | Low | S | S |
| 24 | Implement third-party transfers (2 ops, correspondent partners) | Low | M | S |
| 25 | Decide app-level webhook subscriptions (option a: stay out; option b: add client-credentials mode + `POST /oauth/token`) — then implement the 5 ops incl. `test-notifications` | Med | M | R+S |
| 26 | Reconcile webhook event constants (33) against live events (31): locate preview swift-in events, verify each constant, remove truly dead ones | Med | M | S |
| 27 | Annotate `docs/planning/2026-08-19_wise-api-full-implementation-plan.md` (docs-health ANNOTATE): line 51 `GetQuoteAccountRequirements` PLANNED→DONE; `CreateQuote`/`CreateUnauthenticatedQuote` naming inverted vs code; tier-2/3 statuses | Med | S | S |
| 28 | Snapshot the live preview reference into `docs/reviews/` (pinned evidence + offline re-verification) | Med | S | S |
| 29 | Extend `doc-verify`: gate FEATURES.md coverage totals (41/51/123) + fail loudly on empty count extraction (merges existing TODO item) | Med | S | S+R |
| 30 | README: add one-line pointer to the FEATURES.md endpoint matrix; re-verify the "16 resources" claim against it | Low | XS | S |
| 31 | CHANGELOG: record the FEATURES.md overhaul folded into the next code-touching commit (changelog-only commits structurally fail) | Low | XS | S |
| 32 | Publish GitHub Release objects for v0.10.0 + v0.11.0 — **BLOCKED on your approval** | High | S | R |
| 33 | Tag v1.0.0 (audit green) — **BLOCKED on your explicit approval** | High | XS | R |
| 34 | Re-enable CI on GitHub (push + `gh workflow enable ci` + watch first run) — **BLOCKED on your approval** | High | XS | R |
| 35 | Provide `WISE_SANDBOX_API_KEY` → run first credentialed sandbox test + CHANGELOG entry — **BLOCKED on you** | High | S | R |
| 36 | Decide typed recipient `Details` (v1.0 can freeze the map) — **BLOCKED on your design decision** | Med | XS | R |
| 37 | Set `CACHIX_AUTH_TOKEN` secret — **BLOCKED on you** | Low | XS | R |
| 38 | Accept/reject ADR 003 → then swap failsafe-go for go-retry v0.4.0 — **BLOCKED on ADR acceptance** | Med | M | R |
| 39 | Pin CI-installed tools (`gofumpt`, `govulncheck`, `gorelease` — all `@latest`) | Med | S | R |
| 40 | Mirror the 90% coverage floor into the sandboxed `checks.test` checkPhase | Med | S | R |
| 41 | erraudit pass over webhook+OTT code, curated config, CI/buildflow gate; settle samber/oops adopt-or-decline | Med | M | R |
| 42 | Add `.github/SECURITY.md` (vulnerability reporting path for a published SDK) | Med | XS | R |
| 43 | CONTRIBUTING currency: mention `apidiff`/`doc-verify`, 90% gate, Ginkgo `-count` note | Low | S | R |
| 44 | Quality micro-batch: `GetStatement` PDF/XLSX content-type asserts; `VerifyWebhookSignature` vs Wise's documented example (if published); subscriptions pagination shape check; godoc example for `ListProfileWebhookSubscriptions`; `wise.Version` const | Med | M | R |
| 45 | GOEXPERIMENT ergonomics: pin direnv/home-manager so `jsonv2` is set without buildflow injection — **user-machine change** | Low | S | R |
| 46 | After the conformance harness is fixed: extend it toward 1:1 coverage of all 41 shipped endpoints (current floors: 25 templates / 60 exchanges) | Med | M | S |
| 47 | Decide SCA phone-number enrollment scope (currently OUT_OF_SCOPE; restricted-access endpoints) — explicit Won't-implement or backlog entry | Low | XS | S |
| 48 | Direct-debit accounts + bulk settlement (needs client-credentials decision, rides on #25) | Low | M | S |
| 49 | Refresh `docs/reviews/wise-api-openapi.json` from the live source (or replace with the #28 snapshot) so schema-conformance tests validate against current reality | Med | S | S |
| 50 | Update ROADMAP.md Axis-1 "Today" paragraph: it still narrates tiers as the planning lens; add one line pointing at the FEATURES.md matrix as the live inventory | Low | XS | S |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Canonical surface:** Should the coverage matrix target the **preview** reference only (as instructed today), or must the legacy-only operations (~20+: v1 profile create/extensions, `balanceconvertv1`, legacy SCA verify variants, `usersecurity*`, legacy card permission/phone ops, …) also be inventoried — and if so, as rows in FEATURES.md or as a documented exclusion note?
2. **What next:** Of the 51 PLANNED operations, which do you want implemented first? My Pareto pick is quote PATCH + recipient deactivate/compatibility (closes the remaining core-flow holes at ~small effort); batch groups and profiles-writes are the big-ticket items. Or is another session already executing (see Q3)?
3. **The parallel conformance harness:** Is another agent/session actively finishing `spec_conformance_test.go` right now (in which case I keep hands off `wise_test.go`/spec files and wait), or should this session fix/gate the 48 red failures and finish the harness myself?

---

*Waiting for instructions. Per the status-report skill, section (f) is HARVEST
fuel for TODO_LIST.md/ROADMAP.md — say the word and I'll route it.*
