# Wise API Docs Study — Changelog & API Reference Analysis

> **Resolution (2026-10-07 docs-health pass — supersedes the 2026-09-13 note):**
> tier 1–2 shipped through v0.11.0/v0.12.0 (quotes, recipients, transfers +
> receipts + MT103 payout info, funding, rates, statements, users, MCA/bank
> details, currencies, webhook signature verification, observability, mTLS docs,
> ~~webhook subscription CRUD + typed events~~ **shipped v0.11.0**, SCA
> one-time tokens, the `2026Q4` webhook surface). Still open, routed: addresses,
> balance-capacity, `Profile.currentState`/`externalCustomerId`, OAuth token
> endpoint, simulation endpoints, batch groups
> (plan tier 3/4 → ROADMAP); cards/KYC/SCA-sessions/disputes stay tier-3/4
> demand-gated; OB-package and JOSE remain declined (ROADMAP non-goals);
> x-trace-id stays WONT (intermediary-set). This study remains the canonical
> gap map cited from AGENTS.md — keep it in place.
>
> **Surface update (2026-10-17):** the shipped surface now stands at
> **46 `*Client` methods across 17 resources** (v0.15.0, unchanged method
> count — v0.15.0 shipped the structured error surface, no new endpoints): the transfer
> paperwork trio (payments list, US combined receipt, NOC) and the
> third-party (originator) create/get pair landed, and bank details moved to
> Wise's 2027Q1 `GET /balances/account-details/{id}` surface. v0.16.0 (also
> 2026-10-17) left the surface unchanged (sandbox lane + OTT `2026Q4` flip).
> The routed "still open" set above is unchanged; FEATURES.md owns the live matrix.
>
> **Archive verdict (2026-10-23, docs-health sweep — owner-ordered archiving
> of fully-resolved reports):** every §8 next-step is shipped or routed
> (struck below), every tier-table row carries its outcome, and the live
> truth is maintained where it is acted on: `FEATURES.md` (the 46/49/123
> coverage matrix), `TODO_LIST.md`, and `ROADMAP.md`. Archived to
> `docs/reviews/archived/` — this file is a point-in-time gap analysis and
> no longer the canonical gap map (the 2026-10-07/M22 "keep in place"
> disposition is superseded by the owner's archive order).

**Date:** 2026-08-08
**Sources:** [changelog](https://docs.wise.com/changelog) | [API reference](https://docs.wise.com/api-reference) | [environments](https://docs.wise.com/guides/developer/environments)

---

## Executive Summary

Studied the full Wise Platform API changelog (2023–2026) and API reference (~135+
endpoints across 29 categories). Key findings for wise-go:

1. **CRITICAL: Sandbox URL was stale** — V1 (`api.sandbox.transferwise.tech`)
   deprecated June 30, 2026 (already passed). Fixed to V2 (`api.wise-sandbox.com`).
2. **int64 ID migration (Jan 2026)** — All IDs now explicitly documented as int64.
   wise-go already uses int64 branded IDs — no change needed.
3. **Global headers documented (Apr 2026)** — `X-External-Correlation-Id` and
   `x-trace-id` now documented on all operations. Added `WithCorrelationID` option.
4. **`X-Rate-Limited-By` header (Apr 2026)** — Documented alongside `Retry-After`
   on 429 responses. Added `RateLimitedBy` field to `RateLimitError`.
5. **API surface is ~5% covered** — 3 read-only endpoints out of ~135+. The Wise
   API is vastly larger than the SDK's current scope.

---

## 1. Changelog Findings (chronological, SDK-relevance filtered)

### Critical for wise-go

| Date        | Entry                                                                                                                            | Impact on wise-go                             |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| 27 Jul 2026 | Global API versioning (`2026Q4`) — opt-in date-based version                                                                     | Future: may need version header/param support |
| 15 Jan 2026 | **int64 ID migration** — transfer ID, recipient ID, card transaction ID, user ID, profile ID, balance ID all documented as int64 | **Already correct** — branded IDs use int64   |
| 9 Apr 2026  | **Global headers documented** — `X-External-Correlation-Id`, `x-trace-id` on all operations                                      | **Added** `WithCorrelationID` option          |
| 9 Apr 2026  | **429 response documented** — `Retry-After` + `X-Rate-Limited-By` on all operations                                              | **Added** `RateLimitedBy` to `RateLimitError` |
| 24 Nov 2025 | **Sandbox URL migration** — V1 → V2 sandbox                                                                                      | **Fixed** — updated `SandboxURL` constant     |
| 31 Mar 2026 | OAuth token consolidation — Client Credentials + User Tokens merged into single endpoint                                         | Future: relevant when SDK adds OAuth support  |

### Informational (no immediate SDK change)

| Date        | Entry                                                                      | Notes                                  |
| ----------- | -------------------------------------------------------------------------- | -------------------------------------- |
| 18 Mar 2026 | Webhook schema v4.0.0 — millisecond-precision timestamps                   | Relevant when webhooks are implemented |
| 18 Mar 2026 | Event ordering guide — ordering fields for reconciling out-of-order events | Webhook implementation reference       |
| 12 Mar 2026 | `X-External-Correlation-Id` guide published                                | Confirms header usage pattern          |
| 5 Feb 2025  | Profile `currentState` field added                                         | Could enhance `Profile` struct         |
| 14 Feb 2025 | Profile `externalCustomerId` field added                                   | Could enhance `Profile` struct         |
| 5 Jun 2026  | Client credentials token format migration guide                            | Future OAuth reference                 |

### Deprecations tracked (none affect current SDK scope)

The SDK only uses 3 endpoints (`GET /v2/profiles`, `GET /v4/profiles/{id}/balances`,
`GET /v1/profiles/{id}/balance-statements/{id}/statement.json`). None are deprecated.

---

## 2. API Reference — Full Endpoint Inventory

The Wise Platform API has ~135+ endpoints across 29 categories. The SDK covers 3
(2.2%). Here is the full landscape organized by SDK-relevance tiers:

### Tier 1: Natural SDK expansion targets

These align with the existing ROADMAP and are the most common consumer needs:

| ~~Category ~~          | ~~Key Endpoints ~~                                                                                                                      | ~~wise-go Status~~                                   |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| ~~**Quotes**~~         | ~~`POST /v3/quotes`, `POST /v3/profiles/{id}/quotes`, `GET /v3/profiles/{id}/quotes/{id}`, `GET /v1/quotes/{id}/account-requirements`~~ | ~~PLANNED~~ shipped v0.8.0 (+ requirements v0.9.0)   |
| ~~**Transfers**~~      | ~~`GET /v1/transfers/{id}`, `GET /v1/delivery-estimates/{id}`, CRUD for transfers~~                                                     | ~~PLANNED~~ shipped v0.8.0 (+ funding v0.9.0)        |
| ~~**Recipients**~~     | ~~`GET /v2/accounts`, `GET /v1/accounts/{id}`, `GET /v1/account-requirements`~~                                                         | ~~PLANNED~~ shipped v0.8.0                           |
| ~~**Exchange rates**~~ | ~~`GET /v1/rates`~~                                                                                                                     | ~~Not planned~~ shipped v0.8.0                       |
| ~~**Statements**~~     | ~~`GET /v1/profiles/{id}/balance-statements/{id}/statement.{json,csv,pdf}`~~                                                            | ~~CSV/PDF PLANNED~~ shipped v0.9.0 (all six formats) |

### Tier 2: Moderate-value SDK targets

| ~~Category ~~                  | ~~Key Endpoints ~~                                                                                                                       | ~~Notes~~                                                                                                        |
| ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| ~~**Users**~~                  | ~~`GET /v1/me`, `GET /v1/users/{id}`~~                                                                                                   | ~~Simple read endpoints~~ shipped v0.9.0 (`GetMe`/`GetUser`)                                                     |
| ~~**Webhooks**~~               | ~~Profile & application subscription CRUD + signature verification~~                                                                     | ~~High-value, self-contained~~ shipped: signature verify v0.9.0, profile-level CRUD v0.10.0 (app-level ON_HOLD)  |
| ~~**Balances (expanded)**~~    | ~~`POST /v4/profiles/{id}/balances` (create), `GET /v1/profiles/{id}/total-funds/{currency}`, `GET /v1/profiles/{id}/balance-capacity`~~ | ~~Natural extension of existing balance support~~ shipped v0.9.0 (create/get/total-funds; capacity DEMAND_GATED) |
| ~~**Multi-Currency Account**~~ | ~~`GET /v4/profiles/{id}/multi-currency-account`, eligibility, available currencies~~                                                    | ~~MCA details with bank account numbers~~ read shipped v0.9.0 (configuration PLANNED)                            |
| ~~**Bank Account Details**~~   | ~~`GET /v1/profiles/{id}/account-details`, `GET /v3/profiles/{id}/bank-details`~~                                                        | ~~IBAN/SortCode/routing info~~ shipped v0.9.0; migrated to Wise's 2027Q1 surface in v0.14.0                      |
| ~~**Addresses**~~              | ~~`GET /v1/addresses`, `POST /v1/addresses`, `GET /v1/address-requirements`~~                                                            | ~~CRUD for address book~~ DEMAND_GATED (FEATURES matrix)                                                         |

### Tier 3: Specialized (card issuance, KYC, disputes, SCA)

| ~~Category ~~              | ~~Endpoints~~                                                                     | ~~Notes~~                                                                                       |
| -------------------------- | --------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| ~~**Cards**~~              | ~~approx. 15 endpoints (orders, transactions, limits, spend controls, disputes)~~ | ~~Large surface, card-issuing partners only~~ OUT_OF_SCOPE (FEATURES)                           |
| ~~**KYC Review**~~         | ~~approx. 8 endpoints (list, create, submit, verify requirements)~~               | ~~Compliance-focused~~ OUT_OF_SCOPE (FEATURES)                                                  |
| ~~**SCA**~~                | ~~One-time tokens, SCA sessions, verification~~                                   | ~~Strong customer authentication~~ OTT endpoints shipped v0.11.0; sessions/factors OUT_OF_SCOPE |
| ~~**Disputes**~~           | ~~approx. 4 endpoints (list, get, submit, upload evidence)~~                      | ~~Card dispute management~~ OUT_OF_SCOPE (FEATURES)                                             |
| ~~**Digital Wallets**~~    | ~~Payment tokens, provisioning, activation~~                                      | ~~Apple/Google Pay integration~~ OUT_OF_SCOPE (FEATURES)                                        |
| ~~**Batch Groups**~~       | ~~Batch payment initiation and tracking~~                                         | ~~Bulk payment operations~~ DEMAND_GATED (FEATURES)                                             |
| ~~**Sandbox Simulation**~~ | ~~approx. 15 simulation endpoints~~                                               | ~~Testing-only~~ OUT_OF_SCOPE (FEATURES; the SDK tests against httptest)                        |

### Not in scope for wise-go (per ROADMAP non-goals)

- Open Banking API (OB UK 3.1.11) — separate API for regulated TPPs
- JOSE/JWE encryption playground
- Facetec (biometric verification)
- Partner support cases

---

## 3. Environment URLs (corrected)

### Current documented URLs

| Environment                  | TLS                                         | mTLS                                    |
| ---------------------------- | ------------------------------------------- | --------------------------------------- |
| ~~**Production**~~           | ~~`https://api.wise.com`~~                  | ~~`https://api-mtls.transferwise.com`~~ |
| ~~**Sandbox V2 (current)**~~ | ~~`https://api.wise-sandbox.com`~~          | ~~`https://api-mtls.wise-sandbox.com`~~ |
| ~~Sandbox V1 (deprecated)~~  | ~~`https://api.sandbox.transferwise.tech`~~ | ~~—~~                                   |

### V1 → V2 migration notes (Nov 2025)

- V1 deprecation date: **June 30, 2026** (passed)
- Data created in V1 after April 1, 2025 was **not** migrated
- New credentials required for partners onboarded after April 1, 2025
- mTLS: new certificates needed for V2
- JWE/JWS: new public keys needed for V2
- No code logic changes — only configuration (URLs + credentials)

### wise-go changes made

- Updated `SandboxURL` from `api.sandbox.transferwise.tech` to `api.wise-sandbox.com`
- Updated `DOMAIN_LANGUAGE.md` glossary entry

---

## 4. Global Headers (Apr 2026 documentation)

| Header                          | Purpose                                             | wise-go Status                              |
| ------------------------------- | --------------------------------------------------- | ------------------------------------------- |
| `X-External-Correlation-Id`     | Distributed tracing across Wise API calls           | **Added** `WithCorrelationID` option        |
| `x-trace-id`                    | Internal Wise trace correlation                     | Not added (typically set by intermediaries) |
| `Authorization: Bearer {token}` | Authentication                                      | Already implemented                         |
| `Retry-After` (response)        | Seconds/date to wait before retrying                | Already parsed in `RateLimitError`          |
| `X-Rate-Limited-By` (response)  | Identifies rate-limit scope (e.g., "ip", "profile") | **Added** `RateLimitedBy` field             |

---

## 5. int64 ID Migration (Jan 2026)

Wise explicitly documented all ID fields as `int64` (64-bit). Critical fields:
transfer ID (approaching int32 upper bound), card transaction ID (past int32
upper bound), recipient ID, user ID, profile ID, balance ID.

**wise-go status: Already compliant.** All branded ID types (`ProfileID`,
`BalanceID`, `TransactionID`) are backed by `int64` via `go-branded-id`.

---

## 6. Global API Versioning (Jul 2026)

The first global date-based version (`2026Q4`) is now available (opt-in).
Previous endpoint-based versions remain available.

- Global versions do NOT support some deprecated endpoints
- `POST /v2/profiles/business-profile` not supported in global versions (use v3)
- No immediate SDK change needed — current endpoints remain available

---

## 7. Actions Taken This Session

| # | Action                                    | Files Changed                                            |
| - | ----------------------------------------- | -------------------------------------------------------- |
| 1 | Fixed stale sandbox URL                   | `types.go`, `docs/DOMAIN_LANGUAGE.md`                    |
| 2 | Added `WithCorrelationID` option          | `options.go`, `client.go`                                |
| 3 | Added `RateLimitedBy` to `RateLimitError` | `errors.go`, `client.go`                                 |
| 4 | Added correlation ID BDD tests            | `wise_test.go`                                           |
| 5 | Added `checkError` header capture tests   | `internal_test.go`                                       |
| 6 | Created this study report                 | This file                                                |
| 7 | Updated project docs                      | `TODO_LIST.md`, `ROADMAP.md`, `FEATURES.md`, `AGENTS.md` |

---

## 8. Recommended Next Steps (prioritized)

### Immediate (next release)

1. ~~**Add `GetProfile`** — `GET /v2/profiles/{profileId}` is a simple addition~~ shipped v0.8.0
2. ~~**Expand `Balance` struct** — capture `investmentState`, `cashAccountType`,~~
   ~~`bankDetails` fields already in Wise responses~~ PARTIALLY — the balance
   lifecycle shipped v0.9.0 (CreateBalance / direct GetBalance /
   GetTotalFunds) and `InvestmentState` drives list filtering, but the three
   named wire fields remain unexposed on the public `Balance` struct
   (`cashAccountType` is not even parsed raw); routed 2026-10-23 → ROADMAP
   raw ideas "Balance residual wire fields".
3. ~~**Add exchange rates** — `GET /v1/rates` is self-contained and high-value~~ shipped v0.8.0

### Near-term (v0.7–v0.8)

4. ~~**Quotes API** — prerequisite for transfers~~ shipped v0.8.0 (+ account
   requirements and the two-pass refresh v0.9.0)
5. ~~**Recipients API** — prerequisite for transfers~~ shipped v0.8.0
6. ~~**Transfers API** — the core write-operation value proposition~~ shipped
   v0.8.0 (create/get/list/cancel); funding v0.9.0; the category completed
   v0.14.0 (payments list, US combined receipt, NOC, originator transfers)
7. ~~**Statements (CSV/PDF)** — format parameter on existing endpoint~~
   shipped v0.9.0 (all six formats via `GetStatement`)

### Medium-term (v0.9–v1.0)

8. ~~**Webhook signature verification** — high-value, self-contained helper~~
   shipped v0.9.0 (subscription CRUD followed in v0.10.0)
9. ~~**Per-request correlation ID** — allow overriding per-call, not just
   per-client~~ shipped v0.9.0 (`WithRequestCorrelationID`; ctx wins over the
   client-wide `WithCorrelationID`)
10. ~~**Bank account details** — IBAN/SortCode access for MCA~~ shipped
    v0.9.0; migrated to Wise's 2027Q1 `GET /balances/account-details/{id}`
    surface in v0.14.0

### Context for future work

11. ~~**mTLS support** — `api-mtls.wise.com` / `api-mtls.wise-sandbox.com` endpoints~~
    documentation shipped (README mTLS section); client certs ride the
    `WithHTTPClient` transport seam — no SDK code needed
12. ~~**OAuth token endpoint** — consolidated auth (Mar 2026)~~ routed:
    OUT_OF_SCOPE in FEATURES (only if the SDK takes over token exchange)
13. ~~**Global API versioning** — `2026Q4` header support when it becomes
    required~~ routed: the quarterly surfaces are LIVE for webhooks and OTT
    (`2026Q4`); the 2027Q1 canonicalization belongs to the tracked
    spec-refresh item

---

**ARCHIVED 2026-10-23** (docs-health sweep; owner-ordered archiving of
fully-resolved reports — supersedes the 2026-10-07 and v0.15.0-train M22
banner-grade dispositions above). All §8 items and every tier row carry
inline outcomes; the live coverage matrix is `FEATURES.md`. File moved to
`docs/reviews/archived/`.
