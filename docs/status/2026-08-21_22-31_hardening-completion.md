# Hardening Plan Completion — Final Status Report

**Date:** 2026-08-21 22:31 CEST
**Plan:** `docs/planning/2026-08-21_21-00_post-execution-hardening-plan.md` (Table B: 60 tasks)
**Progress:** **60 of 60 executable tasks done and verified** (Tier 1%: 23/23 · Tier 4%: 20/20 · Tier 20%: 17/17).
**Verification state at completion:** `go test -race -count=1 ./...` green — **143 Ginkgo specs** + plain test funcs; `golangci-lint` **0 issues**; coverage **90.8%** (goal ≥90 met); `nix flake check` **all checks passed**; working tree clean; **23 commits ahead of origin/master, not pushed** (per directive).

---

## a) Completed in this session (resumed at 31/60)

### Bookkeeping (pre-9.x)

- 8.x godoc examples verified committed (daemon-authored 20810a7; all 10 examples intact, +214 lines).
- The 21:46 status report tracked as 9d9d7a5 together with six new AGENTS.md operational gotchas (dprint changelog-only commit failure, flake links-fileset rule, daemon races/sweeps, ExceededError value-matching, newAPIError signature) and the hardening plan's **G4 marked stale** (`docs/DOMAIN_LANGUAGE.md` has existed since 2026-08-08).

### Tier 4% completion — 9.x Spec-hardening (4f1973d + daemon 4669530)

- **9.1** `GetStatementRequest.validate` rejects intervals > 469 days client-side (`maxStatementInterval`, transactions.go); boundary (exactly 469 days) proven to pass. Param name verified against the OpenAPI spec first.
- **9.2** `Locale` field → `statementLocale` query param (spec: 2-character language codes) + BDD asserting the wire param and README/CHANGELOG documentation.
- **9.3** Exchange-rate UTC-Z regression spec: constructs a CEST `time.Time`, asserts the wire `time` value equals `...Z` — the exact v0.8.1 bug class.
- **9.4** First MT940 + QIF happy-path BDD (raw SWIFT/QIF byte passthrough).
- **9.5** AGENTS.md `Accept-Minor-Version` convention entry.
- **9.6** Tier gate: race + lint + `nix flake check` green.

### Tier 20% — 10.x POST account-requirements (759a2ec + 748394b)

- **10.1** No new raw types needed: POST reuses the GET's response schema and `CreateRecipientRequest` as the body (spec-verified: `recipient-create-request` oneOf-free body, same `account-requirements-response`).
- **10.2** `RefreshQuoteAccountRequirements` + `RefreshQuoteAccountRequirementsRequest`: quote/currency/type/details validated client-side; wire body deliberately **omits unset fields** (no empty `accountHolderName`, no zero `profile`) because Wise resolves the dependent form, not the recipient.
- **10.3** Two-pass BDD: GET-flagged `RefreshRequirementsOnChange` field → POST partial form → revised requirements reveal `address.state`; plus 404, corruption, and rejection contexts. **SDK is now 32 endpoint methods.**
- **10.4** Validate table + omit-empties wire table in internal_test.go.
- **10.5** FEATURES/CHANGELOG/README rows; goconst resolved **properly**: named `wireKey{Currency,Type,Details}` constants (helpers.go) applied across all wire builders — the deferred 11.x goconst decision landed early, as production constants rather than test-label appeasement.

### Tier 20% — 11.x Type-model polish (133b367)

- **11.1** `AccountID` brand + `NewAccountID`; `MultiCurrencyAccount.ID` strongly typed (safe: the field shipped untyped in this same unreleased batch).
- **11.2** `wise.HeaderDeliveryID` constant; README webhook section wired to it.
- **11.3** `FundTransferResult` doc records the oneOf→one-struct flattening tradeoff and its revisit trigger.
- **11.4** `parseWiseTimestamp` errors: single readable line (value + accepted layouts) replacing the five-error joined dump.

### Tier 20% — 12.x Nolint/dupl reduction (55f1269)

- **12.1/12.2** `GetUser` folded into the existing `fetchByID` template (byte-identical errors, zero behavior change); both `//nolint:dupl` suppressions **removed** — dupl runs clean unsuppressed. `GetMultiCurrencyAccount` stays hand-rolled (validates profile ID, doesn't fit the by-ID shape); documented in the commit.

### Tier 20% — 13.x Docs-health (1d6e4c4)

- **13.1** ROADMAP: 31→32 methods, refresh moved from "in flight" to Shipped, release-state note (cut-ready, version = question g3).
- **13.2** AGENTS.md: coverage-badge-job entry (own concurrency group, rebase-before-push, badge lags until push) + godoc-examples-nolint context; stale 31 count fixed.
- **13.3** FEATURES.md Documentation section: 18 godoc examples (15 compile-only + 3 runnable) + README reference rows.

### Tier 20% — 14.x Contributor UX (b50d8e2 + 1b8f7c8)

- **14.1** `lychee` in devShell (verified: 0.24.2 via `nix develop -c lychee --version`).
- **14.2** `readme_guard_test.go`: parses every Go code fence in README, asserts each `client.X`/`wise.X` exists in the AST-derived exported API set. Sanity floors prevent vacuous passes (caught my own broken sanity-name on first run: constructor is `New`, not `NewClient`); **negative-verified** — injected stale symbols fail the test, restored README passes. Added to flake fileset along with README.md itself.
- **14.3** CONTRIBUTING: links gate mechanics (fileset union rule), drift guard, badge job, local lychee command.

---

## b) Gates per group (all green)

| Group | Tests | Lint | flake check |
| ----- | ----- | ---- | ----------- |
| 9.x | race green (138 specs) | 0 | passed |
| 10.x | race green (143 specs) | 0 | passed |
| 11.x | race green | 0 | (format only; no fileset change) |
| 12.x | race green | 0 | (no fileset change) |
| 13.x | docs only | — | — |
| 14.x | race green | 0 | passed (after fileset change) |
| **Final** | **race green, 143 specs** | **0** | **passed** |

Coverage: 90.4% → **90.8%** (refresh endpoint + guard test add covered statements).

---

## c) Process notes

- The commit daemon split several groups into multiple commits (4669530/4f1973d for 9.x; 759a2ec/748394b for 10.x; b50d8e2/1b8f7c8 for 14.x). Every daemon commit was verified to contain exactly the intended changes; final tree passes all gates, which is the state that matters.
- One vacuous-test trap caught and closed: the drift guard originally would have passed with a broken extractor; sanity floors (≥10 references + 3 must-exist symbols) now make extraction failures loud.
- The README's flat-details limitation (`map[string]string` cannot express nested `address` objects) is now stated honestly in the Recipients section rather than left implicit; the full typed-recipients redesign remains gated (G3).

## d) Open gated questions — NOT answered, by design

1. **g1 — Sandbox API key** (`WISE_SANDBOX_API_KEY`): the live-test workflow dispatches green-without-key; a key drop makes it real.
2. **g2 — Cachix**: does cache `larsartmann` exist and is `CACHIX_AUTH_TOKEN` set? The CI step degrades to a warning until answered.
3. **g3 — Version call**: 0.9.0 first or straight to v1.0.0? The `[Unreleased]` block is cut-ready for either; no tag exists and none was created.

**Not pushed.** 23 commits await the maintainer's review; push and tag only on instruction.
