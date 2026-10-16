# wise-go — Comprehensive Improvement Plan

> **Resolution (2026-10-07 docs-health pass — supersedes the 2026-09-13 note):**
>
> ~~90% shipped~~ **fully shipped through v0.12.0** — the D-tier
> (InvestmentState, DetailType constants, Doer), E-tier (README sections,
> EndOfStatementBalance), F-tier (badge automation, WithLogger, godoc examples,
> 121+ BDD specs, GetProfile, Money/Currency v0.4.0, HasMore removal,
> internal/raw) all landed. ~~Still open: Retry-After→backoff, `WithUserAgent`,
> enum Stringers, benchmarks, wise_test.go split~~ **all resolved** — Retry-After
> now steers the go-retry delay (v0.12.0), `WithUserAgent` shipped, benchmarks and
> the test-file split landed. Only `RegisterClassification`/enum Stringers remain
> deliberate non-goals (routed). v1.0 lock: audit green, tag user-gated
> (TODO_LIST P2). **Archived 2026-10-16** (docs-health pass): every table row
> carries an inline verdict; no living doc references this plan
> (grep-verified).

> Generated 2026-07-18 19:59 · Pareto-driven · Source: 13-skill review session + brutal self-review

## Context

On 2026-07-18, a comprehensive 13-skill review pass ran over wise-go (code-quality-scan,
deduplicate-code, data-model-review, naming-review, full-code-review, architecture-review,
architecture-visualization, go-modularize, nix-flake-migration, docs-health, update-old-docs,
copywriting, frontend-design). The session produced:

- **4 code fixes** (CARD_PAYMENT classification bug, branded-ID formatting, 4 new tests, UTC docs)
- **8 HTML review reports** + 2 D2 architecture diagrams
- **`flake.nix`** adoption (devShells + checks + treefmt)
- **3 living docs** (FEATURES.md, TODO_LIST.md, ROADMAP.md)
- **README polish** + status-report annotation

This plan takes every finding from every report and every item from the prior session's
50-item TODO list, applies Pareto prioritization, and breaks them into executable tasks.

The codebase baseline: **0 lint issues, 0 duplication clones, 94.8% test coverage, build/vet/test/lint all green.**

---

## Pareto Breakdown

### The 1% that delivers 51% of the result

These are the safety-critical, visibility-critical, and context-critical items. Without them,
everything else is at risk or invisible.

| ID     | Task                                                                                                                          | Why it's the 1%                                                                                                                                                            |
| ------ | ----------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~C1~~ | ~~Run full `nix flake check` (not `--no-build`), fix failures~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner | ~~Without this, the flake gives false confidence. A broken flake is worse than no flake.~~                                                                                 |
| ~~C2~~ | ~~Update CHANGELOG `[Unreleased]` with 2026-07-18 fixes~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner       | ~~Without this, the CARD_PAYMENT bug fix is invisible in release notes. Downstream users won't know to upgrade.~~                                                          |
| ~~C3~~ | ~~Update AGENTS.md with flake.nix workflow + new gotchas~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner      | ~~Future AI sessions need to know about flake.nix (buildflow hook requires flake.nix git-tracked), the CARD_PAYMENT classification fix, and the UTC timezone assumption.~~ |

### The 4% that delivers 64% of the result

Zero-risk internal improvements that eliminate entire categories of potential bugs.

| ID     | Task                                                                                                                                    | Why it's the 4%                                                                                                                                   |
| ------ | --------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~D1~~ | ~~Type `InvestmentState` + generic `parseEnum[T]` helper~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                | ~~Eliminates the untyped-string-constant smell AND the duplicated switch-parser pattern in one move. Internal-only, zero API break.~~             |
| ~~D2~~ | ~~Export `DetailType` constants for `ListTransactionsRequest.Type`~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner      | ~~The valid filter values are currently undiscoverable (unexported constants). Exporting them costs nothing and makes the API self-documenting.~~ |
| ~~D3~~ | ~~Accept `Doer` interface in `WithHTTPClient` instead of `*http.Client`~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner | ~~`*http.Client` satisfies `Do(req) (*Response, error)` implicitly. Backward-compatible. Unlocks mocking for downstream users.~~                  |
| ~~D4~~ | ~~Add nix CI job to `.github/workflows/ci.yml`~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                          | ~~Ensures the flake doesn't drift. Without CI, the flake is a local-only artifact that rots silently.~~                                           |
| ~~D5~~ | ~~Add `.envrc` for direnv~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                               | ~~`use flake` — one line, zero maintenance, makes `nix develop` automatic on `cd`.~~                                                              |

### The 20% that delivers 80% of the result

User-facing documentation, security verification, and data-surfacing improvements.

| ID      | Task                                                                                                                                                 | Why it's the 20%                                                                                         |
| ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| ~~E1~~  | ~~README: "Mocking the client" section~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                               | ~~Every downstream user who writes tests will need this. Prevents a class of support questions.~~        |
| ~~E2~~  | ~~README: "Request middleware via WithHTTPClient" section~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                            | ~~Shows how to add tracing/logging. Closes the "middleware gap" from the architecture review.~~          |
| ~~E3~~  | ~~Run `govulncheck ./...` locally~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                    | ~~CI runs it; we should catch vulnerabilities before CI does.~~                                          |
| ~~E4~~  | ~~Run `go test -race ./...` locally~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                  | ~~CI uses `-race`; local verification should match.~~                                                    |
| ~~E5~~  | ~~Review `CONTRIBUTING.md` for drift~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                 | ~~11.5 KB file never opened this session. May reference outdated patterns.~~                             |
| ~~E6~~  | ~~Expose `EndOfStatementBalance` on `ListTransactionsResponse`~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                       | ~~Wise returns it; the SDK decodes it and throws it away. Zero-cost data the caller wants.~~             |
| ~~E7~~  | ~~Wire `Retry-After` into failsafe-go backoff policy~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                 | ~~Currently parsed but not used to inform backoff delays. Affects rate-limit recovery time.~~            |
| E8      | Register error types with `errorfamily.RegisterClassification` **→ routed — deliberate non-goal (RegisterClassification) per the resolution banner** | `go-error-family v0.6.1` exposes this API; wise-go doesn't use it yet. Enables CLI-level error handling. |
| ~~E9~~  | ~~Extract `wiseDateFormat` constant from `parseWiseDate`~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                             | ~~Magic string `"2006-01-02 15:04:05"` repeated in tests.~~                                              |
| ~~E10~~ | ~~Document `GetBalance` O(n) cost in doc comment~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                     | ~~Callers should know this calls ListBalances + linear scan.~~                                           |

### The other 20% (to reach 100%)

Long-tail quality polish, deep investigations, and future-release preparation.

| ID      | Task                                                                                                                                                                       | Category                 |
| ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ |
| ~~F1~~  | ~~Investigate 5.2% coverage gap (`go tool cover -func`)~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                    | ~~Quality~~              |
| ~~F2~~  | ~~Add benchmarks for `Cents()`, `classifyTransactionType`, `mapTransaction`~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                | ~~Performance~~          |
| ~~F3~~  | ~~Add `Example_*` test functions for godoc~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                 | ~~Documentation~~        |
| ~~F4~~  | ~~Add `fmt.Stringer` for `ProfileType`, `BalanceType`, `TransactionType`~~ **Won't implement — Won't implement — enum Stringers are deliberate non-goals per the banner.** | ~~Developer experience~~ |
| ~~F5~~  | ~~Consider `go:generate` for enum string-to-value maps~~ **Won't implement — Won't implement — declined with the Stringers per the banner.**                               | ~~Maintainability~~      |
| ~~F6~~  | ~~Add `WithUserAgent` option~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                               | ~~Feature~~              |
| ~~F7~~  | ~~Add `WithLogger` option for structured debug logging~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                     | ~~Feature~~              |
| ~~F8~~  | ~~Run `brutal-self-review` skill~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                           | ~~Quality~~              |
| ~~F9~~  | ~~Run `status-report` skill (capstone HTML dashboard)~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                      | ~~Documentation~~        |
| ~~F10~~ | ~~Run `library-deep-dive` on failsafe-go~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Quality~~              |
| ~~F11~~ | ~~Run `library-deep-dive` on go-branded-id~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                 | ~~Quality~~              |
| ~~F12~~ | ~~Run `library-deep-dive` on go-error-family~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                               | ~~Quality~~              |
| ~~F13~~ | ~~Run `bdd-testing` skill to expand Ginkgo contexts~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                        | ~~Quality~~              |
| ~~F14~~ | ~~Add integration test skeleton behind `--live` build tag~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                  | ~~Testing~~              |
| ~~F15~~ | ~~Consider splitting `internal_test.go` by domain~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                          | ~~Maintainability~~      |
| ~~F16~~ | ~~Add `GetProfile` method if Wise supports it~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                              | ~~Feature~~              |
| ~~F17~~ | ~~v0.3.0: Introduce `Money` + `Currency` types (breaking)~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                  | ~~Data model~~           |
| ~~F18~~ | ~~v0.3.0: Normalize enum casing (breaking)~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                 | ~~Data model~~           |
| ~~F19~~ | ~~v0.3.0: Reconcile `TransactionTypeUnknown` (breaking)~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                    | ~~Data model~~           |
| ~~F20~~ | ~~v0.3.0: Drop `Result` suffix or move raw types to `internal/raw` (breaking)~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                              | ~~Naming~~               |
| ~~F21~~ | ~~v1.0: Remove `ListTransactionsResponse.HasMore` (breaking)~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                               | ~~Data model~~           |
| F22     | v1.0: Lock the public API; write stability policy **→ routed — v1.0 lock: audit green, tag user-gated (TODO_LIST P2)**                                                     | Process                  |
| ~~F23~~ | ~~Future: Write operations (POST/PATCH/DELETE helpers)~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                     | ~~Feature~~              |
| ~~F24~~ | ~~Future: Transfers, recipients, quotes resources~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                          | ~~Feature~~              |
| ~~F25~~ | ~~Future: Webhook signature verification~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Feature~~              |

---

## Coarse Plan — 30 to 100 min tasks (sorted by impact/effort)

| #      | ID                                                                                                                                     | Task                                                           | Impact       | Effort        | Customer Value                                     |
| ------ | -------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------- | ------------ | ------------- | -------------------------------------------------- |
| ~~1~~  | ~~C1~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Run full `nix flake check`, fix failures~~                   | ~~Critical~~ | ~~30 min~~    | ~~Confidence in reproducibility~~                  |
| ~~2~~  | ~~C2+C3~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                | ~~CHANGELOG [Unreleased] + AGENTS.md update~~                  | ~~Critical~~ | ~~30 min~~    | ~~Downstream visibility + future session context~~ |
| ~~3~~  | ~~D1~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Type `InvestmentState` + generic `parseEnum[T]`~~            | ~~High~~     | ~~30 min~~    | ~~Type safety; eliminates duplicated parsers~~     |
| ~~4~~  | ~~D2~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Export `DetailType` constants~~                              | ~~High~~     | ~~30 min~~    | ~~API discoverability~~                            |
| ~~5~~  | ~~D3~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Accept `Doer` interface in `WithHTTPClient`~~                | ~~High~~     | ~~30 min~~    | ~~Mocking unlock for downstream~~                  |
| ~~6~~  | ~~D4~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Add nix CI job to ci.yml~~                                   | ~~High~~     | ~~30 min~~    | ~~Prevent flake drift~~                            |
| ~~7~~  | ~~E1+E2~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                | ~~README: mocking + middleware sections~~                      | ~~Medium~~   | ~~40 min~~    | ~~User-facing docs~~                               |
| ~~8~~  | ~~E3+E4~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                | ~~Run govulncheck + go test -race~~                            | ~~Medium~~   | ~~30 min~~    | ~~Security + concurrency safety~~                  |
| ~~9~~  | ~~E5~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Review CONTRIBUTING.md for drift~~                           | ~~Medium~~   | ~~30 min~~    | ~~Doc accuracy~~                                   |
| ~~10~~ | ~~E6~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Expose `EndOfStatementBalance`~~                             | ~~Medium~~   | ~~30 min~~    | ~~Data the caller wants~~                          |
| ~~11~~ | ~~E7~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Wire Retry-After into failsafe-go backoff~~                  | ~~Medium~~   | ~~60 min~~    | ~~Faster rate-limit recovery~~                     |
| ~~12~~ | ~~E8+E9+E10~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner — except RegisterClassification (routed non-goal)          | ~~Register errors + extract date const + document GetBalance~~ | ~~Medium~~   | ~~30 min~~    | ~~Developer experience~~                           |
| ~~13~~ | ~~D5~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Add `.envrc` for direnv~~                                    | ~~Low~~      | ~~5 min~~     | ~~Dev convenience~~                                |
| ~~14~~ | ~~F1~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                   | ~~Investigate 5.2% coverage gap~~                              | ~~Low~~      | ~~30 min~~    | ~~Quality assurance~~                              |
| ~~15~~ | ~~F2+F3~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                | ~~Benchmarks + Example_ tests~~                                | ~~Low~~      | ~~60 min~~    | ~~Performance visibility + godoc~~                 |
| ~~16~~ | ~~F4+F5~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner — except Stringers/go:generate (declined)                      | ~~Stringer implementations + go:generate~~                     | ~~Low~~      | ~~60 min~~    | ~~Developer experience~~                           |
| ~~17~~ | ~~F6+F7~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                | ~~WithUserAgent + WithLogger options~~                         | ~~Low~~      | ~~90 min~~    | ~~Feature additions~~                              |
| ~~18~~ | ~~F8+F9~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                                | ~~Brutal-self-review + status-report skills~~                  | ~~Low~~      | ~~90 min~~    | ~~Quality capstone~~                               |
| ~~19~~ | ~~F10-F12~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                              | ~~Library deep-dives (3 deps)~~                                | ~~Low~~      | ~~90 min~~    | ~~Dependency utilization audit~~                   |
| ~~20~~ | ~~F13-F15~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner                                                              | ~~BDD expansion + integration skeleton + test split~~          | ~~Low~~      | ~~90 min~~    | ~~Test quality~~                                   |
| ~~21~~ | ~~F17-F20~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner — v0.4.0 shipped the Money/Currency redesign                 | ~~v0.3.0 breaking release (Money/Currency/enums/naming)~~      | ~~High~~     | ~~480 min~~   | ~~Type safety~~                                    |
| ~~22~~ | ~~F21-F22~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner — HasMore removed v0.4.0; API lock user-gated (TODO_LIST P2) | ~~v1.0 release (HasMore removal + API lock)~~                  | ~~High~~     | ~~180 min~~   | ~~API stability~~                                  |
| ~~23~~ | ~~F23-F25~~ done — shipped ≤ v0.12.0 per the 2026-10-07 resolution banner — write ops/transfers/webhooks all shipped                   | ~~Future features (write ops, transfers, webhooks)~~           | ~~High~~     | ~~Many days~~ | ~~Completeness~~                                   |

---

## Fine Plan — ≤12 min tasks (sorted by impact/effort)

| #      | Parent                                                                                     | Task                                                              | Time       |
| ------ | ------------------------------------------------------------------------------------------ | ----------------------------------------------------------------- | ---------- |
| ~~1~~  | ~~C1~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Check background `nix flake check` result~~                     | ~~2 min~~  |
| ~~2~~  | ~~C1~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~If check failed: read error, identify root cause~~              | ~~5 min~~  |
| ~~3~~  | ~~C1~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~If check failed: fix derivation (likely fileset issue)~~        | ~~10 min~~ |
| ~~4~~  | ~~C1~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Re-run `nix flake check` to verify fix~~                        | ~~5 min~~  |
| ~~5~~  | ~~C2~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add `[Unreleased]` section to CHANGELOG with CARD_PAYMENT fix~~ | ~~5 min~~  |
| ~~6~~  | ~~C2~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add branded-ID formatting fix to CHANGELOG~~                    | ~~2 min~~  |
| ~~7~~  | ~~C2~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add type filter test + UTC docs to CHANGELOG~~                  | ~~3 min~~  |
| ~~8~~  | ~~C2~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add flake.nix adoption to CHANGELOG~~                           | ~~2 min~~  |
| ~~9~~  | ~~C3~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add flake.nix workflow to AGENTS.md Build & Dev section~~       | ~~5 min~~  |
| ~~10~~ | ~~C3~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add "flake.nix must be git-tracked for buildflow" gotcha~~      | ~~3 min~~  |
| ~~11~~ | ~~C3~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add CARD_PAYMENT/CARD_REFUND classification gotcha~~            | ~~3 min~~  |
| ~~12~~ | ~~C3~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add UTC timezone gotcha~~                                       | ~~2 min~~  |
| ~~13~~ | ~~D1~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add `type InvestmentState string` to types.go~~                 | ~~2 min~~  |
| ~~14~~ | ~~D1~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Promote constants to typed `InvestmentState`~~                  | ~~2 min~~  |
| ~~15~~ | ~~D1~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add `parseEnum[T ~string]` generic helper~~                     | ~~5 min~~  |
| ~~16~~ | ~~D1~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Refactor `parseProfileType` onto `parseEnum`~~                  | ~~5 min~~  |
| ~~17~~ | ~~D1~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Refactor `parseBalanceType` onto `parseEnum`~~                  | ~~5 min~~  |
| ~~18~~ | ~~D1~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Run tests + lint to verify~~                                    | ~~2 min~~  |
| ~~19~~ | ~~D2~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Rename `wiseDetail*` constants to exported `DetailType*`~~      | ~~5 min~~  |
| ~~20~~ | ~~D2~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Update `classifyTransactionType` to use new names~~             | ~~3 min~~  |
| ~~21~~ | ~~D2~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Update `ListTransactionsRequest.Type` field doc~~               | ~~2 min~~  |
| ~~22~~ | ~~D2~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Run tests + lint to verify~~                                    | ~~2 min~~  |
| ~~23~~ | ~~D3~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Define unexported `doer interface{ Do(...) }` in client.go~~    | ~~3 min~~  |
| ~~24~~ | ~~D3~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Change `Client.httpClient` field type to `doer`~~               | ~~2 min~~  |
| ~~25~~ | ~~D3~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Change `WithHTTPClient` to accept `doer`~~                      | ~~3 min~~  |
| ~~26~~ | ~~D3~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Run tests + lint to verify backward compat~~                    | ~~2 min~~  |
| ~~27~~ | ~~D4~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add `nix:` job to ci.yml using cachix/install-nix-action~~      | ~~10 min~~ |
| ~~28~~ | ~~D5~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Create `.envrc` with `use flake`~~                              | ~~2 min~~  |
| ~~29~~ | ~~E1~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Write "Mocking the client" README section~~                     | ~~10 min~~ |
| ~~30~~ | ~~E2~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Write "Request middleware" README section~~                     | ~~10 min~~ |
| ~~31~~ | ~~E3~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Run `govulncheck ./...` and capture output~~                    | ~~5 min~~  |
| ~~32~~ | ~~E4~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Run `go test -race ./...` and verify~~                          | ~~5 min~~  |
| ~~33~~ | ~~E5~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Read CONTRIBUTING.md, identify drift~~                          | ~~10 min~~ |
| ~~34~~ | ~~E6~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Add `EndOfStatementBalance Money` to ListTransactionsResponse~~ | ~~10 min~~ |
| ~~35~~ | ~~E7~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Read failsafe-go backoff docs for custom delay function~~       | ~~10 min~~ |
| ~~36~~ | ~~E8~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Call `errorfamily.RegisterClassification` in init()~~           | ~~10 min~~ |
| ~~37~~ | ~~E9~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0)  | ~~Extract `wiseDateFormat` constant~~                             | ~~2 min~~  |
| ~~38~~ | ~~E10~~ done — shipped per the 2026-10-07 resolution banner (tiers landed through v0.12.0) | ~~Add O(n) cost doc to `GetBalance`~~                             | ~~2 min~~  |

---

## Execution Graph

```mermaid
graph TD
    subgraph "Phase 0 — Safety (DONE)"
        S1[Commit existing work] --> S2[Push to origin]
    end

    subgraph "Phase 1 — 1% → 51%"
        C1[nix flake check full] --> C1a{Pass?}
        C1a -->|No| C1b[Fix derivation]
        C1b --> C1
        C1a -->|Yes| C2[CHANGELOG Unreleased]
        C2 --> C3[AGENTS.md update]
    end

    subgraph "Phase 2 — 4% → 64%"
        C3 --> D1[Type InvestmentState + parseEnum]
        D1 --> D2[Export DetailType constants]
        D2 --> D3[Accept Doer interface]
        D3 --> D4[Add nix CI job]
        D4 --> D5[Add .envrc]
    end

    subgraph "Phase 3 — 20% → 80%"
        D5 --> E1E2[README: mocking + middleware]
        D5 --> E3E4[govulncheck + race test]
        D5 --> E5[Review CONTRIBUTING.md]
        D5 --> E6[Expose EndOfStatementBalance]
        D5 --> E7[Wire Retry-After backoff]
        D5 --> E8E9E10[Register errors + extract const + doc GetBalance]
    end

    subgraph "Phase 4 — Commit + Push"
        E1E2 --> P1[Commit improvements]
        E3E4 --> P1
        E5 --> P1
        E6 --> P1
        E7 --> P1
        E8E9E10 --> P1
        P1 --> P2[Push to origin]
    end

    subgraph "Phase 5 — Other 20% (future)"
        P2 --> F1[Coverage gap investigation]
        P2 --> F2[Benchmarks + Examples]
        P2 --> F3[Stringer + go:generate]
        P2 --> F4[WithUserAgent + WithLogger]
        P2 --> F5[Skill capstones: brutal-review, status-report]
        P2 --> F6[Library deep-dives]
        P2 --> V030[v0.3.0: Money/Currency redesign]
        V030 --> V100[v1.0: API lock]
        V100 --> FUTURE[Future: write ops, transfers, webhooks]
    end
```

---

## Anti-Verschlimmbesserung Checklist

Before every change:

- [ ] Read the file before editing
- [ ] Run tests after every code change
- [ ] Run lint after every code change
- [ ] Never change public API without explicit migration path
- [ ] Never `git reset` or `git checkout`
- [ ] Never edit files I haven't read in this session
- [ ] If something feels wrong, STOP and think before acting
