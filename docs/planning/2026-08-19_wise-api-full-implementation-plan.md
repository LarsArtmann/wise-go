# Wise API Full Implementation Plan

**Date:** 2026-08-19
**Source:** [docs.wise.com/api-reference](https://docs.wise.com/api-reference)
**Current SDK version:** v0.7.0 (plan baseline; SDK is now v0.16.0)

> **Update (2026-10-17 docs-health pass):** tier 1 is fully shipped — ~~row 5 (`GetQuoteAccountRequirements`, still PLANNED)~~ shipped in v0.9.0 (struck in the table below). Tier 2 shipped except row 23 (`/v1/comparisons`, still PLANNED). The SDK is now **v0.16.0** with **46 `*Client` methods across 17 resources** (v0.15.0 shipped the structured error surface and v0.16.0 the sandbox lane — no endpoint changes) — the transfer category completed 2026-10-16 (payments list, US combined receipt, NOC) plus third-party (originator) transfers, and bank details migrated to Wise's 2027Q1 `GET /balances/account-details/{id}` surface. The full per-endpoint coverage matrix lives in `FEATURES.md`, and open tier-2/3/4 work is tracked in `TODO_LIST.md` (bounded tasks) and `ROADMAP.md` (strategy). Earlier banner: ~~the SDK is now v0.12.0 with 41 methods~~ (2026-10-07); ~~v0.8.1, tier-1 fully shipped except row 5~~ (2026-08-21).
>
> **Archive verdict (2026-10-23, docs-health sweep — owner-ordered archiving
> of fully-resolved reports):** every tier-2/3/4 row now carries an inline
> outcome (struck below) and all four phases are shipped; the live truth is
> maintained in `FEATURES.md` (coverage matrix), `TODO_LIST.md`, and
> `ROADMAP.md`. Archived to `docs/planning/archived/` — the tier framework's
> residue (row 23 comparisons, profile writes, demand-gated families) is
> exactly the FEATURES matrix's PLANNED/DEMAND_GATED/OUT_OF_SCOPE rows.

---

## Executive Summary

The Wise Platform API exposes ~135 endpoints across 29 categories. wise-go v0.7.0 currently implements **5 client methods** covering the read side of 4 resources:

- `ListProfiles` (`GET /v2/profiles`)
- `ListBalances` (`GET /v4/profiles/{id}/balances`)
- `GetBalance` (client-side scan over `ListBalances`)
- `ListTransactions` (`GET /v1/profiles/{id}/balance-statements/{id}/statement.json`)
- `ListTransfers` (`GET /v1/transfers`)

This plan maps the full API surface, applies Pareto prioritisation, and defines an executable dependency graph. The goal is not to land every endpoint in one change — that would be unmaintainable — but to establish the patterns (write helpers, request types, raw/result boundaries, tests) and ship the **1% that unlocks 51% of consumer value**: the core transfer flow.

---

## Pareto Prioritisation

| Tier                 | Share of endpoints | Approx. count | Value delivered  | Target endpoints                                                                                                          |
| -------------------- | ------------------ | ------------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------- |
| **1% core**          | ~1–2%              | 3–4           | 51%              | `GetProfile`, `GetTransfer`, `CreateQuote`, `GetQuote`, `CreateRecipient`, `GetRecipient`, `CreateTransfer`               |
| **4% high-value**    | ~4%                | 5–6           | 64% (cumulative) | `ListRecipients`, `GetExchangeRate`, `CancelTransfer`, `GetDeliveryEstimate`, `GetQuoteAccountRequirements`               |
| **20% SDK-complete** | ~20%               | 27            | 80% (cumulative) | Quotes (full), Recipients (full), Transfers (full), Balances expanded, Statements CSV/PDF, Webhooks signature, Users read |
| **80% long tail**    | ~80%               | 100+          | 20%              | Cards, KYC, SCA factors, Disputes, Digital Wallets, Batch Groups, Simulations, Cases                                      |

**Principle:** ship the 1% first, because a consumer who can create a quote, add a recipient, and send a transfer gets the majority of the SDK's value. Everything else is incremental.

---

## Full Endpoint Inventory

### Tier 1 — Core transfer flow (ship first)

| # | Category | Method             | Path     | SDK method                                      | Status                                                    |
| - | -------- | ------------------ | -------- | ----------------------------------------------- | --------------------------------------------------------- |
|   | ~~1~~    | ~~Profiles~~       | ~~GET~~  | ~~`/v2/profiles/{profileId}`~~                  | ~~`GetProfile(ctx, ProfileID)`~~                          |
|   | ~~2~~    | ~~Quotes~~         | ~~POST~~ | ~~`/v3/quotes`~~                                | ~~`CreateUnauthenticatedQuote(ctx, CreateQuoteRequest)`~~ |
|   | ~~3~~    | ~~Quotes~~         | ~~POST~~ | ~~`/v3/profiles/{profileId}/quotes`~~           | ~~`CreateQuote(ctx, ProfileID, CreateQuoteRequest)`~~     |
|   | ~~4~~    | ~~Quotes~~         | ~~GET~~  | ~~`/v3/profiles/{profileId}/quotes/{quoteId}`~~ | ~~`GetQuote(ctx, ProfileID, QuoteID)`~~                   |
|   | ~~5~~    | ~~Quotes~~         | ~~GET~~  | ~~`/v1/quotes/{quoteId}/account-requirements`~~ | ~~`GetQuoteAccountRequirements(ctx, QuoteID)`~~           |
|   | ~~6~~    | ~~Recipients~~     | ~~GET~~  | ~~`/v2/accounts`~~                              | ~~`ListRecipients(ctx, ListRecipientsRequest)`~~          |
|   | ~~7~~    | ~~Recipients~~     | ~~GET~~  | ~~`/v1/accounts/{accountId}`~~                  | ~~`GetRecipient(ctx, RecipientID)`~~                      |
|   | ~~8~~    | ~~Recipients~~     | ~~POST~~ | ~~`/v1/accounts`~~                              | ~~`CreateRecipient(ctx, CreateRecipientRequest)`~~        |
|   | ~~9~~    | ~~Transfers~~      | ~~GET~~  | ~~`/v1/transfers/{transferId}`~~                | ~~`GetTransfer(ctx, TransferID)`~~                        |
|   | ~~10~~   | ~~Transfers~~      | ~~POST~~ | ~~`/v1/transfers`~~                             | ~~`CreateTransfer(ctx, CreateTransferRequest)`~~          |
|   | ~~11~~   | ~~Transfers~~      | ~~PUT~~  | ~~`/v1/transfers/{transferId}/cancel`~~         | ~~`CancelTransfer(ctx, TransferID)`~~                     |
|   | ~~12~~   | ~~Transfers~~      | ~~GET~~  | ~~`/v1/delivery-estimates/{transferId}`~~       | ~~`GetDeliveryEstimate(ctx, TransferID)`~~                |
|   | ~~13~~   | ~~Exchange rates~~ | ~~GET~~  | ~~`/v1/rates`~~                                 | ~~`GetExchangeRate(ctx, source, target, time)`~~          |

### Tier 2 — High-value, self-contained

| #      | Category                  | Method   | Path                                                                                               | Notes                                                                                                                |
| ------ | ------------------------- | -------- | -------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| ~~14~~ | ~~Users~~                 | ~~GET~~  | ~~/me~~                                                                                            | ~~Current user~~ shipped v0.9.0 (`GetMe`)                                                                            |
| ~~15~~ | ~~Users~~                 | ~~GET~~  | ~~/users/{userId}~~                                                                                | ~~User by ID~~ shipped v0.9.0 (`GetUser`)                                                                            |
| ~~16~~ | ~~Balances~~              | ~~POST~~ | ~~/v4/profiles/{profileId}/balances~~                                                              | ~~Create balance~~ shipped v0.9.0 (idempotency key added v0.12.0)                                                    |
| ~~17~~ | ~~Balances~~              | ~~GET~~  | ~~/v4/profiles/{profileId}/balances/{balanceId}~~                                                  | ~~Get balance by ID (new endpoint; may remove client-side scan)~~ shipped v0.9.0 (direct endpoint; the scan is gone) |
| ~~18~~ | ~~Balances~~              | ~~GET~~  | ~~/v1/profiles/{profileId}/total-funds/{currency}~~                                                | ~~Total funds~~ shipped v0.9.0                                                                                       |
| ~~19~~ | ~~Statements~~            | ~~GET~~  | ~~/v1/profiles/{profileId}/balance-statements/{balanceId}/statement.{csv,pdf,xlsx,xml,mt940,qif}~~ | ~~Format parameter~~ shipped v0.9.0 (six formats via raw `GetStatement`)                                             |
| ~~20~~ | ~~Webhooks~~              | ~~—~~    | ~~Signature verification helper~~                                                                  | ~~High value, no REST calls~~ shipped v0.9.0 (verify) + v0.10.0 (typed event decoding)                               |
| ~~21~~ | ~~Transfer requirements~~ | ~~POST~~ | ~~/v1/transfer-requirements~~                                                                      | ~~Validate before create~~ shipped v0.8.0                                                                            |
| ~~22~~ | ~~Currencies~~            | ~~GET~~  | ~~/v1/currencies~~                                                                                 | ~~Allowed currencies~~ shipped v0.9.0                                                                                |
| ~~23~~ | ~~Comparisons~~           | ~~GET~~  | ~~/v1/comparisons~~                                                                                | ~~Provider comparison~~ still PLANNED — routed to the FEATURES matrix (tier 2 leftover)                              |

### Tier 3 — Moderate value, broader surface

| #      | Category                   | Method             | Path                                                                                            | Notes                                                                                                                                     |
| ------ | -------------------------- | ------------------ | ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| ~~24~~ | ~~Profiles~~               | ~~GET~~            | ~~`/profiles/{profileId}`~~                                                                     | ~~v1 alias~~ superseded — `GetProfile` ships on `/v2` (v0.8.0); FEATURES owns profile rows                                                |
| ~~25~~ | ~~Profiles~~               | ~~POST~~           | ~~`/profiles/personal-profile`~~                                                                | ~~Create personal profile~~ PLANNED (FEATURES matrix)                                                                                     |
| ~~26~~ | ~~Profiles~~               | ~~POST~~           | ~~`/profiles/business-profile`~~                                                                | ~~Create business profile~~ PLANNED (FEATURES matrix, v5)                                                                                 |
| ~~27~~ | ~~Profiles~~               | ~~PUT~~            | ~~`/profiles/{profileId}/personal-profile`~~                                                    | ~~Update personal profile~~ PLANNED (FEATURES matrix)                                                                                     |
| ~~28~~ | ~~Profiles~~               | ~~PUT~~            | ~~`/profiles/{profileId}/business-profile`~~                                                    | ~~Update business profile~~ PLANNED (FEATURES matrix)                                                                                     |
| ~~29~~ | ~~Addresses~~              | ~~GET/POST~~       | ~~`/addresses`, `/addresses/{id}`, `/address-requirements`~~                                    | ~~Address book~~ DEMAND_GATED (FEATURES)                                                                                                  |
| ~~30~~ | ~~Bank account details~~   | ~~GET/POST~~       | ~~`/profiles/{profileId}/account-details`, `/profiles/{profileId}/bank-details`~~               | ~~IBAN/routing~~ reads shipped v0.9.0; migrated to Wise's 2027Q1 `/balances/account-details/{id}` in v0.14.0 (writes PLANNED in FEATURES) |
| ~~31~~ | ~~Multi-currency account~~ | ~~GET~~            | ~~`/v4/profiles/{profileId}/multi-currency-account`, `/borderless-accounts-configuration/...`~~ | ~~MCA~~ read shipped v0.9.0 (configuration PLANNED in FEATURES)                                                                           |
| ~~32~~ | ~~Batch groups~~           | ~~POST/PATCH/GET~~ | ~~`/profiles/{profileId}/batch-groups/...`~~                                                    | ~~Bulk payments~~ DEMAND_GATED (FEATURES)                                                                                                 |
| ~~33~~ | ~~Payin deposit details~~  | ~~GET~~            | ~~`/profiles/{profileId}/transfers/{transferId}/deposit-details/bank-transfer`~~                | ~~Pay-in instructions~~ DEMAND_GATED (FEATURES)                                                                                           |
| ~~34~~ | ~~Direct debit accounts~~  | ~~GET/POST~~       | ~~`/profiles/{profileId}/direct-debit-accounts`~~                                               | ~~Bulk funding~~ DEMAND_GATED (FEATURES)                                                                                                  |
| ~~35~~ | ~~Bulk settlement~~        | ~~POST~~           | ~~`/settlements`~~                                                                              | ~~Client-creds only~~ DEMAND_GATED (FEATURES)                                                                                             |

### Tier 4 — Specialised / partner-only / long tail

| #      | Category                                 | Count                                                                                                                                                                                                                    | Notes                                                                                                                                                                 |
| ------ | ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~36~~ | ~~OAuth token~~                          | ~~1~~                                                                                                                                                                                                                    | ~~`POST /oauth/token` — only if SDK takes over token exchange~~ OUT_OF_SCOPE (FEATURES)                                                                               |
| ~~37~~ | ~~JOSE playground~~                      | ~~6~~                                                                                                                                                                                                                    | ~~Signature/encryption testing~~ OUT_OF_SCOPE (FEATURES)                                                                                                              |
| ~~38~~ | ~~Users write~~                          | ~~4~~                                                                                                                                                                                                                    | ~~Signup, exists, contact email~~ OUT_OF_SCOPE (FEATURES)                                                                                                             |
| ~~39~~ | ~~Claim account~~                        | ~~1~~                                                                                                                                                                                                                    | ~~`POST /user/claim-account`~~ OUT_OF_SCOPE (FEATURES)                                                                                                                |
| ~~40~~ | ~~Profile verification~~                 | ~~8~~                                                                                                                                                                                                                    | ~~KYC review, additional verification, FaceTec~~ OUT_OF_SCOPE (FEATURES)                                                                                              |
| ~~41~~ | ~~Link requests / embedded flows~~       | ~~4~~                                                                                                                                                                                                                    | ~~iframe/webview helpers~~ OUT_OF_SCOPE (FEATURES)                                                                                                                    |
| ~~42~~ | ~~SCA / one-time tokens~~                | ~~10~~                                                                                                                                                                                                                   | ~~OTT status, SCA sessions, OTP channels~~ OTT status/trigger/verify shipped v0.11.0 (incl. `ClearSCAChallenge`); sessions and OTP enrollment OUT_OF_SCOPE (FEATURES) |
| ~~43~~ | ~~PIN / facemaps / device fingerprints~~ | ~~8~~                                                                                                                                                                                                                    | ~~JWE-encrypted SCA factors~~ OUT_OF_SCOPE (FEATURES; sca-jose module trigger documented in ROADMAP)                                                                  |
| ~~44~~ | ~~Cards~~                                | ~~approx. 25~~                                                                                                                                                                                                           | ~~Orders, transactions, limits, spend controls, disputes, sensitive details, 3DS~~ OUT_OF_SCOPE (FEATURES)                                                            |
| ~~45~~ | ~~Digital wallets~~                      | ~~4~~                                                                                                                                                                                                                    | ~~Apple/Google Pay push provisioning~~ OUT_OF_SCOPE (FEATURES)                                                                                                        |
| ~~46~~ | ~~Disputes~~                             | ~~7~~                                                                                                                                                                                                                    | ~~Card dispute management~~ OUT_OF_SCOPE (FEATURES)                                                                                                                   |
| ~~47~~ | ~~Incoming transfers~~                   | ~~1~~                                                                                                                                                                                                                    | ~~Partner-only~~ OUT_OF_SCOPE (FEATURES)                                                                                                                              |
| ~~48~~ | ~~Payins~~                               | ~~1~~                                                                                                                                                                                                                    | ~~PayNow QR~~ OUT_OF_SCOPE (FEATURES)                                                                                                                                 |
| ~~49~~ | ~~Webhook subscriptions~~                | ~~9 operations across 5 paths (2026-09-13 spec correction: 2 app-level CRUD + 2 app-level get/delete + test-notifications + 4 profile-level ops)~~ profile-level CRUD shipped v0.10.0; app-level five ON_HOLD (FEATURES) | ~~Application + profile level~~ resolved                                                                                                                              |
| ~~50~~ | ~~Cases~~                                | ~~3~~                                                                                                                                                                                                                    | ~~Partner support~~ OUT_OF_SCOPE (FEATURES)                                                                                                                           |
| ~~51~~ | ~~Sandbox simulations~~                  | ~~approx. 15~~                                                                                                                                                                                                           | ~~Testing-only helpers~~ OUT_OF_SCOPE (FEATURES)                                                                                                                      |

---

## Dependency Graph

The core flow has strict prerequisites:

```mermaid
flowchart TD
    A[GetProfile] --> B[CreateQuote]
    C[CreateRecipient] --> D[CreateTransfer]
    B --> D
    D --> E[GetTransfer]
    D --> F[CancelTransfer]
    D --> G[GetDeliveryEstimate]
    H[GetExchangeRate] -.-> B
    I[GetQuoteAccountRequirements] -.-> C
    J[ListRecipients] --> C
    K[GetRecipient] -.-> D
    L[TransferRequirements] -.-> D
```

**Execution order for the 1% tier:**

1. HTTP write helpers (`post`, `postWithQuery`, `patch`, `put`, `delete`) in `client.go`.
2. `GetProfile` — validates the `{id}` path pattern.
3. `GetExchangeRate` — no dependencies; confirms query-param helpers.
4. `GetTransfer` — reuses `mapTransfer` from `ListTransfers`.
5. `CreateQuote` + `GetQuote` — introduces `QuoteID`, `Quote` type, `PayIn`/`PayOut` enums.
6. `CreateRecipient` + `GetRecipient` + `ListRecipients` — introduces `Recipient` type and account-details polymorphism.
7. `CreateTransfer` — consumes `QuoteID` + `RecipientID`.
8. `CancelTransfer` + `GetDeliveryEstimate` — cheap follow-ups.

---

## Architectural Decisions

1. **Keep the flat `client.X` surface for now.** The resource count (~7 after this phase) is still below the ~8 threshold documented in `ROADMAP.md` for a service-client refactor.
2. **Extend the two-layer raw/result boundary.** Every new response type gets a wire-format struct in `internal/raw/types.go` and a parsed public type in `types.go`.
3. **Branded IDs for every new entity.** Add `QuoteID`, `UserID`, `AddressID`, `CardToken`, etc., as needed. The 1% tier needs `QuoteID`.
4. **Write helpers mirror read helpers.** `post(ctx, path, body, target)` encodes `body` as JSON, sets `Content-Type: application/json`, and delegates response handling to `checkError` + `jsonDecode`.
5. **Idempotency headers where documented.** `X-idempotence-uuid` for `CreateRecipient`, `CreateTransfer`, and business-profile creation. Accept an optional `IdempotenceKey` field in request structs.
6. **Validation before the network.** Reject impossible requests (e.g. `CreateTransfer` without `QuoteID` or `TargetAccountID`) with `errorfamily.NewRejection`, matching the existing `ListTransactionsRequest.validate()` pattern.
7. **No `Money` arithmetic.** `Money` remains a serialization boundary value object. Quote/transfer amounts are parsed into `Money`; consumers perform their own math.

---

## Quality Gates Per Endpoint

For every endpoint added:

- [ ] Raw wire type in `internal/raw/types.go`.
- [ ] Parsed public type in `types.go`.
- [ ] Branded ID if a new entity is referenced.
- [ ] Client method in the appropriate resource file.
- [ ] Request validation (if any fields constrain each other).
- [ ] BDD test in `wise_test.go` with `httptest` mock.
- [ ] Unit test for mapper edge cases in `internal_test.go`.
- [ ] Update `FEATURES.md`, `TODO_LIST.md`, `CHANGELOG.md`.

---

## Recommended Phasing

### Phase 1 — Foundations (this session)

- ~~HTTP POST/PATCH/PUT/DELETE helpers.~~ done v0.8.0
- ~~`GetProfile`.~~ done v0.8.0
- ~~`GetExchangeRate`.~~ done v0.8.0
- ~~`GetTransfer`.~~ done v0.8.0

### Phase 2 — Quotes + Recipients (next session)

- ~~`CreateQuote`, `GetQuote`, `GetQuoteAccountRequirements`.~~ done v0.8.0
  (requirements + two-pass refresh v0.9.0)
- ~~`ListRecipients`, `GetRecipient`, `CreateRecipient`.~~ done v0.8.0

### Phase 3 — Transfers (next session)

- ~~`CreateTransfer`.~~ done v0.8.0
- ~~`CancelTransfer`, `GetDeliveryEstimate`.~~ done v0.8.0 (funding v0.9.0)

### Phase 4 — Completion (future)

- ~~Tier 2 endpoints, then tier 3, then tier 4 as consumer demand justifies.~~
  tier 2 shipped v0.9.0 except comparisons (PLANNED); tiers 3/4 routed to the
  FEATURES matrix (PLANNED / DEMAND_GATED / OUT_OF_SCOPE rows)

---

## Out of Scope

The following are documented but deliberately excluded from the initial implementation because they require non-REST patterns, partner access, or are too specialised:

- ~~Open Banking API (separate regulated API).~~
- ~~JOSE/JWE encryption playground.~~
- ~~FaceTec biometric verification.~~
- ~~PCI-sensitive card details.~~
- ~~Partner support cases.~~
- ~~Sandbox simulation endpoints (testing-only).~~

_All six confirmed still out of scope at archive time (2026-10-23): the
FEATURES matrix carries them as OUT_OF_SCOPE rows._

---

**ARCHIVED 2026-10-23** (docs-health sweep; owner-ordered archiving of
fully-resolved reports — supersedes the 2026-10-17 M22 banner-grade
disposition above). Every tier row and phase carries an inline outcome;
the live coverage matrix is `FEATURES.md`. File moved to
`docs/planning/archived/`.
