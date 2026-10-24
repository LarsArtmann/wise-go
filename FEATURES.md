# Features

Honest inventory of wise-go features by status. Code is the source of truth — every
claim here can be verified against the implementation.

## Status vocabulary

Eight labels answering one question: does working code exist — and if not, why not?

| Status               | When it applies                                                                                                                                                                                        |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| FULLY_FUNCTIONAL     | Code present AND working (tests pass or exercised).                                                                                                                                                    |
| PARTIALLY_FUNCTIONAL | Ships but has known gaps, edge-case bugs, or missing pieces.                                                                                                                                           |
| BROKEN               | Code exists but fails. Intentionally unused right now — zero BROKEN rows is itself information; the failure state must stay nameable.                                                                  |
| DISABLED             | Code exists and is correct, but execution is externally switched off (server-side disable, missing credentials). Unblock it and the row becomes FULLY_FUNCTIONAL.                                      |
| PLANNED              | No code yet; **completes an already-shipped resource family** — next on the build path (implementation-plan tiers 1–3).                                                                                |
| DEMAND_GATED         | No code yet; the **whole family is unshipped** and gets built only on a real consumer demand signal (plan tier 3 / ROADMAP.md Axis 1).                                                                 |
| ON_HOLD              | No code yet; a **pending scope/design decision** — not demand — blocks it.                                                                                                                             |
| OUT_OF_SCOPE         | Wise surface wise-go intentionally does not target: specialised, partner-only (client credentials / restricted access), or sandbox-only. Revisit only on a real demand signal (see ROADMAP.md Axis 1). |

## Wise API endpoint coverage

Full audit of the live Wise Platform API reference
(`docs.wise.com/api-reference/preview`, audited **2026-09-16**) against the SDK.
This is the single source of truth for endpoint existence; the resource sections
below cover behavior (parsing, filtering, validation), not endpoint inventory.

- **215 documented REST operations** across **51 reference categories**, plus
  **29 webhook-event definitions** in the preview sweep (**31** distinct event
  types counting the settlement and Swift events from the webhook-event index).
  Re-derivation note (2026-10-16): the vendored 2027Q1 spec
  (`docs/reviews/wise-api-openapi.json`) now carries **188 paths / 227
  operations** — the +12 over the 215 live-preview census (2026-09-16) are the
  2027Q1 additions triaged into the sections below; the category count is a
  live-preview-pages figure, not a spec-object count, so the two denominators
  measure different things by design.
- **wise-go ships 46 of those documented operations** through **43 of the 46
  public `*Client` methods** — the other three are `Authenticate`/`Health`
  (convenience wrappers) and `ClearSCAChallenge` (a multi-call loop);
  `TriggerOTT`/`VerifyOTT` each cover three channel paths. The count is
  gate-checked by `nix run .#doc-verify` against the compiled surface.
- **51 operations have no code yet**, split by why: **20 PLANNED** (complete
  already-shipped families), **26 DEMAND_GATED** (whole families awaiting a
  consumer demand signal — plan tier 3), **5 ON_HOLD** (app-level webhook
  scope decision). Tiers per
  [the implementation plan](docs/planning/archived/2026-08-19_wise-api-full-implementation-plan.md).
- **123 operations are OUT_OF_SCOPE** (cards, KYC/KYB, SCA factors, JOSE,
  simulations, embedded flows, cases, disputes, settlement allocations, …).

### Summary by category

| Category                                                                                                | Ops | Shipped | State                                                               |
| ------------------------------------------------------------------------------------------------------- | --- | ------- | ------------------------------------------------------------------- |
| Balances                                                                                                | 10  | 4       | core shipped; delete/movements PLANNED; regulatory ops DEMAND_GATED |
| Balance statements                                                                                      | 1   | 1       | all 6 formats shipped                                               |
| Bank account details                                                                                    | 5   | 1       | read shipped; ordering PLANNED; returns DEMAND_GATED                |
| Batch groups                                                                                            | 7   | 0       | DEMAND_GATED (tier 3)                                               |
| Profiles                                                                                                | 19  | 2       | reads shipped; writes PLANNED; KYB OUT_OF_SCOPE                     |
| Quotes                                                                                                  | 4   | 3       | complete except PATCH update (PLANNED)                              |
| Recipients                                                                                              | 8   | 5       | core shipped; deactivate/compat/confirmations PLANNED               |
| Transfers (standard)                                                                                    | 11  | 11      | complete: payments-list, US receipt, NOC shipped 2026-10-16         |
| Transfers (third-party)                                                                                 | 2   | 2       | complete: originator create + get shipped 2026-10-16                |
| Users                                                                                                   | 6   | 2       | reads shipped; writes OUT_OF_SCOPE                                  |
| Multi-currency account                                                                                  | 4   | 1       | MCA read shipped; configuration PLANNED                             |
| Webhook subscriptions                                                                                   | 9   | 4       | profile-level shipped; app-level ON_HOLD                            |
| SCA one-time tokens                                                                                     | 1   | 1       | shipped (v0.11.0)                                                   |
| SCA OTP channels                                                                                        | 8   | 6       | trigger/verify shipped; enrollment OUT_OF_SCOPE                     |
| Currencies                                                                                              | 1   | 1       | shipped                                                             |
| Exchange rates                                                                                          | 1   | 1       | shipped                                                             |
| Delivery estimates                                                                                      | 1   | 1       | shipped                                                             |
| Comparison                                                                                              | 1   | 0       | PLANNED (tier 2 leftover)                                           |
| Activity                                                                                                | 1   | 0       | DEMAND_GATED                                                        |
| Contacts                                                                                                | 1   | 0       | DEMAND_GATED                                                        |
| Addresses                                                                                               | 5   | 0       | DEMAND_GATED                                                        |
| Payin deposit details                                                                                   | 1   | 0       | DEMAND_GATED                                                        |
| Direct debit accounts                                                                                   | 2   | 0       | DEMAND_GATED                                                        |
| Bulk settlement                                                                                         | 1   | 0       | DEMAND_GATED (needs client credentials)                             |
| GPI tracking                                                                                            | 1   | 0       | DEMAND_GATED                                                        |
| OAuth token                                                                                             | 1   | 0       | OUT_OF_SCOPE                                                        |
| Claims, cases, KYC, SCA factors, cards, disputes, spend, JOSE, simulation, settlement, incoming, payins | 122 | 0       | OUT_OF_SCOPE                                                        |

### Balances (10 operations)

| Method | Path                                                | Status           | wise-go                                       |
| ------ | --------------------------------------------------- | ---------------- | --------------------------------------------- |
| POST   | `/v4/profiles/{id}/balances`                        | FULLY_FUNCTIONAL | `CreateBalance` (`balances.go:107`)           |
| GET    | `/v4/profiles/{id}/balances`                        | FULLY_FUNCTIONAL | `ListBalances` (`balances.go:31`)             |
| GET    | `/v4/profiles/{id}/balances/{id}`                   | FULLY_FUNCTIONAL | `GetBalance` (`balances.go:61`)               |
| GET    | `/v4/profiles/{id}/total-funds/{currency}`          | FULLY_FUNCTIONAL | `GetTotalFunds` (`balances.go:195`)           |
| DELETE | `/v4/profiles/{id}/balances/{id}`                   | PLANNED          | close balance (zero balance required)         |
| POST   | `/v4/profiles/{id}/balance-movements`               | PLANNED          | convert/move between balances                 |
| GET    | `/v4/profiles/{id}/balance-capacity`                | DEMAND_GATED     | regulatory deposit limit (hold-limit regions) |
| POST   | `/v4/profiles/{id}/excess-money-account`            | DEMAND_GATED     | excess-funds sweep target                     |
| GET    | `/v4/profiles/{id}/balances/hold-limit-breach`      | DEMAND_GATED     | hold-limit breaches (SG/MY)                   |
| POST   | `/v4/profiles/{id}/balances/hold-limit-breach/{id}` | DEMAND_GATED     | close breach via one-time refund              |

### Balance statements (1 operation, 6 formats)

| Method | Path                                                                                     | Status           | wise-go                                                                                        |
| ------ | ---------------------------------------------------------------------------------------- | ---------------- | ---------------------------------------------------------------------------------------------- |
| GET    | `/v1/profiles/{id}/balance-statements/{id}/statement.{json,csv,pdf,xlsx,camt,mt940,qif}` | FULLY_FUNCTIONAL | `ListTransactions` (`transactions.go:16`), `GetStatement` raw download (`transactions.go:276`) |

### Bank account details (5 operations)

Wise's 2027Q1 roll removed `GET /v1/profiles/{id}/account-details` and relocated
the read to `GET /balances/account-details/{id}`, keyed by the details' own ID;
wise-go migrated in the same shape (live probe 2026-10-16: old path 404s).

| Method | Path                                                      | Status           | wise-go                                            |
| ------ | --------------------------------------------------------- | ---------------- | -------------------------------------------------- |
| GET    | `/balances/account-details/{id}` (2027Q1)                 | FULLY_FUNCTIONAL | `GetBankAccountDetails` (`account_details.go:158`) |
| POST   | `/v1/profiles/{id}/bank-details`                          | PLANNED          | issue local+international details                  |
| POST   | `/v1/profiles/{id}/account-details-orders`                | PLANNED          | order account details                              |
| GET    | `/v1/profiles/{id}/account-details-orders`                | PLANNED          | list orders                                        |
| POST   | `/v1/profiles/{id}/account-details/payments/{id}/returns` | DEMAND_GATED     | return a received payment                          |

### Multi-currency account (4 operations)

| Method | Path                                                                    | Status           | wise-go                                             |
| ------ | ----------------------------------------------------------------------- | ---------------- | --------------------------------------------------- |
| GET    | `/v1/profiles/{id}/multi-currency-account`                              | FULLY_FUNCTIONAL | `GetMultiCurrencyAccount` (`account_details.go:29`) |
| GET    | `/borderless-accounts-configuration/profiles/{id}/available-currencies` | PLANNED          | balance-eligible currencies                         |
| GET    | `/borderless-accounts-configuration/profiles/{id}/payin-currencies`     | PLANNED          | details-eligible currencies                         |
| GET    | `/multi-currency-account/eligibility`                                   | PLANNED          | eligibility by profile or location                  |

### Currencies / rates / delivery estimates / comparison (4 operations)

| Method | Path                                  | Status           | wise-go                                            |
| ------ | ------------------------------------- | ---------------- | -------------------------------------------------- |
| GET    | `/v1/currencies`                      | FULLY_FUNCTIONAL | `ListCurrencies` (`currencies.go:22`)              |
| GET    | `/v1/rates`                           | FULLY_FUNCTIONAL | `GetExchangeRate` (`rates.go:18`)                  |
| GET    | `/v1/delivery-estimates/{transferId}` | FULLY_FUNCTIONAL | `GetDeliveryEstimate` (`delivery_estimates.go:33`) |
| GET    | `/comparisons`                        | PLANNED          | provider comparison (tier 2 of the plan)           |

### Profiles (19 operations)

| Method | Path                                                      | Status           | wise-go                           |
| ------ | --------------------------------------------------------- | ---------------- | --------------------------------- |
| GET    | `/v2/profiles`                                            | FULLY_FUNCTIONAL | `ListProfiles` (`profiles.go:31`) |
| GET    | `/v2/profiles/{id}`                                       | FULLY_FUNCTIONAL | `GetProfile` (`profiles.go:13`)   |
| POST   | `/profiles/personal-profile`                              | PLANNED          | create personal profile           |
| POST   | `/profiles/business-profile`                              | PLANNED          | create business profile (v5)      |
| PUT    | `/profiles/{id}/personal-profile`                         | PLANNED          | update personal profile           |
| PUT    | `/profiles/{id}/business-profile`                         | PLANNED          | update business profile           |
| GET    | `/profiles/{id}/business-profile/business-representative` | OUT_OF_SCOPE     | KYB data management               |
| PUT    | `/profiles/{id}/business-profile/business-representative` | OUT_OF_SCOPE     | KYB data management               |
| POST   | `/profiles/{id}/verification-documents`                   | OUT_OF_SCOPE     | KYC document management           |
| PUT    | `/profiles/{id}/verification-documents`                   | OUT_OF_SCOPE     | KYC document management           |
| POST   | `/profiles/{id}/directors`                                | OUT_OF_SCOPE     | KYB data management               |
| GET    | `/profiles/{id}/directors`                                | OUT_OF_SCOPE     | KYB data management               |
| PUT    | `/profiles/{id}/directors`                                | OUT_OF_SCOPE     | KYB data management               |
| POST   | `/profiles/{id}/ubos`                                     | OUT_OF_SCOPE     | KYB data management               |
| GET    | `/profiles/{id}/ubos`                                     | OUT_OF_SCOPE     | KYB data management               |
| PUT    | `/profiles/{id}/ubos`                                     | OUT_OF_SCOPE     | KYB data management               |
| POST   | `/profiles/{id}/update-window`                            | OUT_OF_SCOPE     | KYB update window                 |
| DELETE | `/profiles/{id}/update-window`                            | OUT_OF_SCOPE     | KYB update window                 |
| POST   | `/profiles/{id}/verification-status/bank-transfer`        | OUT_OF_SCOPE     | KYC verification status           |

### Users (6 operations)

| Method | Path                             | Status           | wise-go                   |
| ------ | -------------------------------- | ---------------- | ------------------------- |
| GET    | `/v1/me`                         | FULLY_FUNCTIONAL | `GetMe` (`users.go:49`)   |
| GET    | `/v1/users/{id}`                 | FULLY_FUNCTIONAL | `GetUser` (`users.go:66`) |
| POST   | `/user/signup/registration_code` | OUT_OF_SCOPE     | partner onboarding flow   |
| POST   | `/users/exists`                  | OUT_OF_SCOPE     | partner onboarding flow   |
| PUT    | `/users/{id}/contact-email`      | OUT_OF_SCOPE     | partner onboarding flow   |
| GET    | `/users/{id}/contact-email`      | OUT_OF_SCOPE     | partner onboarding flow   |

### Quotes (4 operations)

| Method | Path                            | Status           | wise-go                                       |
| ------ | ------------------------------- | ---------------- | --------------------------------------------- |
| POST   | `/v3/quotes`                    | FULLY_FUNCTIONAL | `CreateUnauthenticatedQuote` (`quotes.go:18`) |
| POST   | `/v3/profiles/{id}/quotes`      | FULLY_FUNCTIONAL | `CreateQuote` (`quotes.go:39`)                |
| GET    | `/v3/profiles/{id}/quotes/{id}` | FULLY_FUNCTIONAL | `GetQuote` (`quotes.go:61`)                   |
| PATCH  | `/v3/profiles/{id}/quotes/{id}` | PLANNED          | update quote with recipient                   |

### Recipients (8 operations)

| Method | Path                                          | Status           | wise-go                                             |
| ------ | --------------------------------------------- | ---------------- | --------------------------------------------------- |
| GET    | `/v2/accounts`                                | FULLY_FUNCTIONAL | `ListRecipients` (`recipients.go:24`)               |
| POST   | `/v1/accounts`                                | FULLY_FUNCTIONAL | `CreateRecipient` (`recipients.go:108`)             |
| GET    | `/v1/accounts/{id}`                           | FULLY_FUNCTIONAL | `GetRecipient` (`recipients.go:88`)                 |
| GET    | `/v1/quotes/{id}/account-requirements`        | FULLY_FUNCTIONAL | `GetQuoteAccountRequirements` (`quotes.go:95`)      |
| POST   | `/v1/quotes/{id}/account-requirements`        | FULLY_FUNCTIONAL | `RefreshQuoteAccountRequirements` (`quotes.go:158`) |
| DELETE | `/v1/accounts/{id}`                           | PLANNED          | deactivate recipient                                |
| POST   | `/v1/accounts/{id}/quotes/{id}/compatibility` | PLANNED          | pre-transfer compatibility check                    |
| PATCH  | `/v1/accounts/{id}/confirmations`             | PLANNED          | accept verification mismatch                        |

### Transfers — standard (11 operations)

| Method | Path                                         | Status           | wise-go                                                        |
| ------ | -------------------------------------------- | ---------------- | -------------------------------------------------------------- |
| POST   | `/v1/transfers`                              | FULLY_FUNCTIONAL | `CreateTransfer` (`transfers.go:219`)                          |
| GET    | `/v1/transfers`                              | FULLY_FUNCTIONAL | `ListTransfers` (`transfers.go:39`)                            |
| GET    | `/v1/transfers/{id}`                         | FULLY_FUNCTIONAL | `GetTransfer` (`transfers.go:117`)                             |
| PUT    | `/v1/transfers/{id}/cancel`                  | FULLY_FUNCTIONAL | `CancelTransfer` (`transfers.go:255`)                          |
| POST   | `/v1/transfer-requirements`                  | FULLY_FUNCTIONAL | `ValidateTransferRequirements` (`transfer_requirements.go:20`) |
| POST   | `/v1/profiles/{id}/transfers/{id}/payments`  | FULLY_FUNCTIONAL | `FundTransfer` (`transfers.go:284`)                            |
| GET    | `/v1/transfers/{id}/invoices/bankingpartner` | FULLY_FUNCTIONAL | `GetTransferPayoutInfo` (`transfers.go:161`)                   |
| GET    | `/v1/transfers/{id}/receipt.pdf`             | FULLY_FUNCTIONAL | `GetTransferReceipt` (`transfers.go:138`)                      |
| GET    | `/v1/transfers/{id}/payments`                | FULLY_FUNCTIONAL | `ListTransferPayments` (`transfers.go:433`)                    |
| GET    | `/v1/transfers/{id}/us-combined-receipt.pdf` | FULLY_FUNCTIONAL | `GetUSCombinedReceipt` (`transfers.go:508`)                    |
| GET    | `/v1/transfers/{id}/documents/noc`           | FULLY_FUNCTIONAL | `GetTransferNOC` (`transfers.go:530`)                          |

### Transfers — third-party (2 operations)

| Method | Path                                           | Status           | wise-go                                                    |
| ------ | ---------------------------------------------- | ---------------- | ---------------------------------------------------------- |
| POST   | `/v1/profiles/{id}/third-party-transfers`      | FULLY_FUNCTIONAL | `CreateThirdPartyTransfer` (`third_party_transfers.go:95`) |
| GET    | `/v1/profiles/{id}/third-party-transfers/{id}` | FULLY_FUNCTIONAL | `GetThirdPartyTransfer` (`third_party_transfers.go:145`)   |

### Webhook subscriptions (9 operations)

| Method | Path                                                              | Status           | wise-go                                                |
| ------ | ----------------------------------------------------------------- | ---------------- | ------------------------------------------------------ |
| POST   | `/2026Q4/profiles/{id}/subscriptions`                             | FULLY_FUNCTIONAL | `CreateProfileWebhookSubscription` (`webhooks.go:92`)  |
| GET    | `/2026Q4/profiles/{id}/subscriptions`                             | FULLY_FUNCTIONAL | `ListProfileWebhookSubscriptions` (`webhooks.go:119`)  |
| GET    | `/2026Q4/profiles/{id}/subscriptions/{id}`                        | FULLY_FUNCTIONAL | `GetProfileWebhookSubscription` (`webhooks.go:156`)    |
| DELETE | `/2026Q4/profiles/{id}/subscriptions/{id}`                        | FULLY_FUNCTIONAL | `DeleteProfileWebhookSubscription` (`webhooks.go:190`) |
| POST   | `/applications/{clientKey}/subscriptions`                         | ON_HOLD          | app-level scope decision pending (ROADMAP.md Axis 1)   |
| GET    | `/applications/{clientKey}/subscriptions`                         | ON_HOLD          | app-level scope decision pending                       |
| GET    | `/applications/{clientKey}/subscriptions/{id}`                    | ON_HOLD          | app-level scope decision pending                       |
| DELETE | `/applications/{clientKey}/subscriptions/{id}`                    | ON_HOLD          | app-level scope decision pending                       |
| POST   | `/applications/{clientKey}/subscriptions/{id}/test-notifications` | ON_HOLD          | app-level scope decision pending                       |

### SCA one-time tokens and OTP channels (9 operations)

| Method | Path                                         | Status           | wise-go                                                |
| ------ | -------------------------------------------- | ---------------- | ------------------------------------------------------ |
| GET    | `/2026Q4/one-time-token/status`              | FULLY_FUNCTIONAL | `GetOTTStatus` (`ott.go:225`)                          |
| POST   | `/2026Q4/one-time-token/sms/trigger`         | FULLY_FUNCTIONAL | `TriggerOTT` (`ott.go:246`, `OTTChannelSMS`)           |
| POST   | `/2026Q4/one-time-token/sms/verify`          | FULLY_FUNCTIONAL | `VerifyOTT` (`ott.go:272`, `OTTChannelSMS`)            |
| POST   | `/2026Q4/one-time-token/whatsapp/trigger`    | FULLY_FUNCTIONAL | `TriggerOTT` (`ott.go:246`, `OTTChannelWhatsApp`)      |
| POST   | `/2026Q4/one-time-token/whatsapp/verify`     | FULLY_FUNCTIONAL | `VerifyOTT` (`ott.go:272`, `OTTChannelWhatsApp`)       |
| POST   | `/2026Q4/one-time-token/voice/trigger`       | FULLY_FUNCTIONAL | `TriggerOTT` (`ott.go:246`, `OTTChannelVoice`)         |
| POST   | `/2026Q4/one-time-token/voice/verify`        | FULLY_FUNCTIONAL | `VerifyOTT` (`ott.go:272`, `OTTChannelVoice`)          |
| POST   | `/application/users/{id}/phone-numbers`      | OUT_OF_SCOPE     | restricted access (client credentials + Wise approval) |
| DELETE | `/application/users/{id}/phone-numbers/{id}` | OUT_OF_SCOPE     | restricted access (client credentials + Wise approval) |

### Batch groups (7 operations) — DEMAND_GATED

Create/get/complete-or-cancel a group, add transfers, fund from balance,
direct-debit payment initiation create/get. All seven DEMAND_GATED (tier 3 of
the plan; bulk payments only on a consumer demand signal).

### Remaining single-operation categories (DEMAND_GATED)

| Category              | Method + Path                                                                                            | Status       | wise-go                    |
| --------------------- | -------------------------------------------------------------------------------------------------------- | ------------ | -------------------------- |
| Activity              | GET `/v1/profiles/{id}/activities`                                                                       | DEMAND_GATED | account activity feed      |
| Contacts              | POST `/v1/profiles/{id}/contacts`                                                                        | DEMAND_GATED | Wisetag/email/phone lookup |
| Addresses             | POST `/v1/addresses`, GET `/v1/addresses`, GET `/v1/addresses/{id}`, GET/POST `/v1/address-requirements` | DEMAND_GATED | address book (5 ops)       |
| Payin deposit details | GET `/v1/profiles/{id}/transfers/{id}/deposit-details/bank-transfer`                                     | DEMAND_GATED | bank-transfer pay-in info  |
| Direct debit accounts | POST/GET `/v1/profiles/{id}/direct-debit-accounts`                                                       | DEMAND_GATED | ACH/EFT funding accounts   |
| Bulk settlement       | POST `/settlements`                                                                                      | DEMAND_GATED | needs client credentials   |
| GPI tracking          | POST `/v1/gpi-tracking`                                                                                  | DEMAND_GATED | SWIFT gpi lookup           |

### Out-of-scope categories (123 operations, OUT_OF_SCOPE)

| Category                                 | Ops | Reason                                                   |
| ---------------------------------------- | --- | -------------------------------------------------------- |
| Cards (list/status/permissions/PIN)      | 5   | card issuing product (tier 4)                            |
| Card kiosk collection                    | 2   | card issuing product                                     |
| Card orders (+ address validation)       | 7   | card issuing product                                     |
| Card sensitive details                   | 4   | PCI JWE flow, card issuing product                       |
| Card transactions                        | 2   | card issuing product                                     |
| Digital wallets (Apple/Google Pay)       | 4   | push provisioning, card issuing product                  |
| Disputes                                 | 7   | card dispute lifecycle                                   |
| Spend controls                           | 6   | card authorisation rules (application-scoped)            |
| Spend limits                             | 8   | card/profile/cardholder limits                           |
| 3DS challenge result                     | 1   | card issuing (push-notification 3DS)                     |
| KYC review                               | 5   | hosted-KYC / partner-KYC product                         |
| Additional verification                  | 3   | partner KYC (deprecated upstream in favor of KYC review) |
| FaceTec                                  | 1   | biometric SCA export                                     |
| Link requests (embedded flows)           | 4   | Embedded Flows product                                   |
| Claim account                            | 1   | redirect-to-Wise onboarding                              |
| SCA sessions                             | 1   | JWE-encrypted SCA factors                                |
| SCA PIN / facemaps / device fingerprints | 9   | JWE-encrypted SCA factors                                |
| JOSE (key management + playgrounds)      | 7   | mTLS/JWS/JWE partner hardening                           |
| OAuth token                              | 1   | only if the SDK takes over token exchange (tier 4 #36)   |
| Cases + case files                       | 4   | partner support operations                               |
| Settlement allocations                   | 3   | correspondent-scale funds allocation                     |
| Incoming transfers                       | 1   | partner-only (implementation manager required)           |
| Payins (PayNow)                          | 1   | SGD rail funding                                         |
| Simulation                               | 17  | sandbox-only helpers; the SDK tests against httptest     |

### Webhook events

- The preview reference documents **29 event definitions** across its category
  pages (2026-09-16 sweep); the union with the webhook-event index (Swift
  `swift-in#credit`, `swift-message-received`) is **31 distinct event types**.
- wise-go ships **33 documented `WebhookEventType` constants** (open enum,
  `types.go:644-678`, verified against Wise's live webhook-event reference
  2026-09-13); unknown event types pass through by design.
- **Typed payloads for 3 events**: `transfers#state-change`,
  `transfers#payout-failure`, `balances#credit` (`ParseWebhookEvent`).

## Client core

| Feature                                 | Status           | Evidence                                                                                                                                                 |
| --------------------------------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `wise.New(apiKey, opts...)` constructor | FULLY_FUNCTIONAL | `client.go:52`; functional-options pattern                                                                                                               |
| Bearer-token authentication             | FULLY_FUNCTIONAL | `client.go:397` `setHeaders`                                                                                                                             |
| Sandbox environment (`WithSandbox`)     | FULLY_FUNCTIONAL | `options.go:64`; `SandboxURL` const in `types.go:24`                                                                                                     |
| Custom base URL (`WithBaseURL`)         | FULLY_FUNCTIONAL | `options.go:71`                                                                                                                                          |
| Custom HTTP timeout (`WithTimeout`)     | FULLY_FUNCTIONAL | `options.go:95`                                                                                                                                          |
| Custom retry policy (`WithRetry`)       | FULLY_FUNCTIONAL | `options.go:103`; exponential backoff with jitter via go-retry                                                                                           |
| Custom HTTP client (`WithHTTPClient`)   | FULLY_FUNCTIONAL | `options.go:114`; accepts `Doer` interface (`client.go:41`)                                                                                              |
| Correlation ID (`WithCorrelationID`)    | FULLY_FUNCTIONAL | `options.go:127`; sets `X-External-Correlation-Id` header on all requests                                                                                |
| Retry with backoff (429, 5xx, network)  | FULLY_FUNCTIONAL | `client.go:145` `isRetryableError`; `Retry-After` honored via `DelayFunc` (capped)                                                                       |
| `Authenticate(ctx)`                     | FULLY_FUNCTIONAL | `client.go:185`; delegates to `ListProfiles`                                                                                                             |
| `Health(ctx)`                           | FULLY_FUNCTIONAL | `client.go:202`; credential-validating readiness probe (authenticated `ListProfiles` through the retry executor — not a cheap liveness check; see godoc) |
| `WithUserAgent` option (v0.11.0)        | FULLY_FUNCTIONAL | `options.go:88`; custom `User-Agent` on every request, default preserved (BDD)                                                                           |
| Shared-client concurrency (v0.11.0)     | FULLY_FUNCTIONAL | `wise_test.go:246`; 16 parallel requests, per-request correlation-ID isolation, `-race`                                                                  |

## Transaction behavior

| Feature                                                  | Status               | Evidence                                                                                     |
| -------------------------------------------------------- | -------------------- | -------------------------------------------------------------------------------------------- |
| `ListTransactionsRequest.Type` filter forwarding         | FULLY_FUNCTIONAL     | `transactions.go:33`; BDD-tested                                                             |
| Request validation (empty currency, inverted date range) | FULLY_FUNCTIONAL     | `transactions.go:208`; returns `wise.transactions.invalid_request` (`:204`)                  |
| Transaction type classification                          | FULLY_FUNCTIONAL     | `transactions.go:177`; CARD_PAYMENT / CARD_REFUND split fixed 2026-07-18                     |
| Cross-currency transaction mapping                       | FULLY_FUNCTIONAL     | `transactions.go:135` `mapExchange`; uses transaction currency, not request currency         |
| `Transaction.Exchange` (`*TransactionExchange`)          | FULLY_FUNCTIONAL     | `types.go:119`; nil for non-conversion transactions                                          |
| `EndOfStatementBalance` exposure                         | FULLY_FUNCTIONAL     | `types.go:192`; surfaced as `Money` on `ListTransactionsResponse`                            |
| `Transaction.Date` UTC semantics                         | PARTIALLY_FUNCTIONAL | `users.go:137` `parseWiseDate` parses as UTC; Wise sends no TZ — documented in field comment |
| Pagination                                               | PLANNED              | Wise returns all transactions in one response; no endpoint requires it yet                   |

## Balance filtering behavior

| Feature                                       | Status           | Evidence                                                       |
| --------------------------------------------- | ---------------- | -------------------------------------------------------------- |
| Filter visible + non-investment balances      | FULLY_FUNCTIONAL | `balances.go:43`; drops `Visible: false` and invested balances |
| Fetch hidden or invested balances             | FULLY_FUNCTIONAL | `GetBalance` direct endpoint retrieves them individually       |
| Profile name construction (personal/business) | FULLY_FUNCTIONAL | `profiles.go:31` name/BusinessName branches                    |

## Error handling

| Feature                                           | Status           | Evidence                                                                                            |
| ------------------------------------------------- | ---------------- | --------------------------------------------------------------------------------------------------- |
| `APIError` base type                              | FULLY_FUNCTIONAL | `errors.go:241`; embeds into all subtypes; carries `Headers` (`:252`) and `Details` (`:255`)        |
| `BadRequestError` (HTTP 400)                      | FULLY_FUNCTIONAL | `errors.go:423`; granular reason codes arrive in `Details`                                          |
| `RateLimitError` (HTTP 429) with `RetryAfter`     | FULLY_FUNCTIONAL | `errors.go:320`; parses delta-seconds + HTTP-date                                                   |
| `RateLimitError.RateLimitedBy` (429 header)       | FULLY_FUNCTIONAL | `errors.go:324`; captures `X-Rate-Limited-By` header                                                |
| `AuthError` (HTTP 401, 403)                       | FULLY_FUNCTIONAL | `errors.go:357`                                                                                     |
| `SCAChallengeError` (403 + 2FA headers, v0.6.0)   | FULLY_FUNCTIONAL | `errors.go:373`; `TwoFAApprovalToken()` returns the OTT                                             |
| `WithSCAApprovalToken` option (v0.6.0)            | FULLY_FUNCTIONAL | `options.go:161`; sends cleared OTT as `x-2fa-approval`                                             |
| `NotFoundError` (HTTP 404)                        | FULLY_FUNCTIONAL | `errors.go:410`                                                                                     |
| `ServerError` (HTTP 5xx)                          | FULLY_FUNCTIONAL | `errors.go:432`                                                                                     |
| `ErrorCode()` / `ErrorFamily()` / `IsRetryable()` | FULLY_FUNCTIONAL | All implement go-error-family interfaces                                                            |
| `errors.AsType` matching                          | FULLY_FUNCTIONAL | Demonstrated in README; tested                                                                      |
| `APIError.Details` structured error envelope      | FULLY_FUNCTIONAL | `errors.go:255`; Wise's `errors[].code`/`message` per entry, wire order                             |
| `WiseErrorCode` open enum (14 documented codes)   | FULLY_FUNCTIONAL | `errors.go:195`; unknown wire values pass through                                                   |
| Curated code sentinels (11) + `APIError.Is`       | FULLY_FUNCTIONAL | `ErrNotValid` (`errors.go:124`) ... `Is` (`errors.go:307`) matches `Details` codes through any wrap |
| `APIError.HasCode(WiseErrorCode)`                 | FULLY_FUNCTIONAL | `errors.go:286`; open-enum matcher for uncurated codes                                              |
| `wise_code` in `ErrorContext()`                   | FULLY_FUNCTIONAL | `errors.go:276` (APIError) and `errors.go:342` (RateLimitError); first granular code                |
| Granular-code reference doc                       | FULLY_FUNCTIONAL | `docs/error-codes.md`; code, sentinel, status context, retry guidance                               |

## Type system

| Feature                                               | Status           | Evidence                                                                                  |
| ----------------------------------------------------- | ---------------- | ----------------------------------------------------------------------------------------- |
| Branded `ProfileID` / `BalanceID`                     | FULLY_FUNCTIONAL | `ids.go:32,38`; phantom types prevent entity-ID mixing at compile time                    |
| Branded `TransactionID`                               | FULLY_FUNCTIONAL | `ids.go:41`                                                                               |
| `Money` value object (`Cents` + `Currency`)           | FULLY_FUNCTIONAL | `types.go:65`; paired cents/currency makes mismatch unrepresentable                       |
| `Currency` branded type with ISO 4217 validation      | FULLY_FUNCTIONAL | `types.go:49` `NewCurrency`; validates 3-letter uppercase ASCII                           |
| Two-layer raw/result split (`internal/raw` boundary)  | FULLY_FUNCTIONAL | Wire types in `internal/raw/types.go`; parsed types in `types.go`; `helpers.go:54` bridge |
| `ProfileType`, `BalanceType`, `TransactionType` enums | FULLY_FUNCTIONAL | `types.go:139,147,164`                                                                    |
| `InvestmentState` typed enum                          | FULLY_FUNCTIONAL | `types.go:155`; used for balance filtering (`balances.go:43`)                             |
| `DetailType` typed enum + constants                   | FULLY_FUNCTIONAL | `transactions.go:159`; typed filter for `ListTransactionsRequest.Type`                    |
| `OneTimeToken` typed credential (unreleased)          | FULLY_FUNCTIONAL | `ott.go:38`; `NewOneTimeToken` validates; typed on every OTT-taking surface               |
| `OTTAction` open enum (unreleased)                    | FULLY_FUNCTIONAL | `ott.go:154`; `OTTStatus.ActionType` (was bare `string`)                                  |
| Enum casing normalization (lowercase SDK values)      | FULLY_FUNCTIONAL | `BalanceType` normalized; `ProfileType`/`TransactionType` already lowercase               |

## Wire format hardening

| Feature                                        | Status           | Evidence                                                  |
| ---------------------------------------------- | ---------------- | --------------------------------------------------------- |
| Tolerant timestamp parsing (4 layouts)         | FULLY_FUNCTIONAL | `helpers.go` `parseWiseTimestamp`; zoneless = UTC         |
| Outgoing query timestamps as UTC `Z` (v0.8.1)  | FULLY_FUNCTIONAL | `helpers.go:167` `formatWiseTimestamp`; Wise 422s offsets |
| `ListBalances` `types` query param (v0.5.3)    | FULLY_FUNCTIONAL | `balances.go:18,23`; live API 400s without it             |
| Mapper errors classified `Corruption` (v0.5.2) | FULLY_FUNCTIONAL | `internal_test.go`; fail-fast instead of retry loops      |

## Webhook behavior

| Feature                                                 | Status           | Evidence                                                                                              |
| ------------------------------------------------------- | ---------------- | ----------------------------------------------------------------------------------------------------- |
| `ParseWebhookPublicKey` (PKIX + PKCS#1 PEM, RSA-only)   | FULLY_FUNCTIONAL | `webhooks.go:35`; garbage/non-RSA input rejected with clear errors                                    |
| `VerifyWebhookSignature` (RSA-SHA256 over raw body)     | FULLY_FUNCTIONAL | `webhooks.go:64`; valid/tampered/wrong-key/malformed/empty/5 MiB tested                               |
| `HeaderWebhookSignature` / `HeaderDeliveryID` constants | FULLY_FUNCTIONAL | `webhooks.go:22,24`; delivery-dedup guidance in README Webhooks section                               |
| `ClearSCAChallenge` convenience loop                    | FULLY_FUNCTIONAL | `ott.go:322`; triggers, prompts via callback, verifies, repeats until every challenge passed          |
| `OTTChannel` typed channel enum                         | FULLY_FUNCTIONAL | `ott.go:54`; single lowercase-wire → UPPERCASE-challenge-type mapping                                 |
| OTT secrecy: token never in error strings               | FULLY_FUNCTIONAL | `SCAChallengeError.Error()` reports `token issued: true/false` only; value via `TwoFAApprovalToken()` |

## Build & tooling

| Feature                                                | Status           | Evidence                                                                                                                                                                                                                                                                                                              |
| ------------------------------------------------------ | ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `go test ./...`                                        | FULLY_FUNCTIONAL | httptest mocks; no network required                                                                                                                                                                                                                                                                                   |
| `golangci-lint run`                                    | FULLY_FUNCTIONAL | 0 issues                                                                                                                                                                                                                                                                                                              |
| `.github/workflows/ci.yml` (build/test/lint/vulncheck) | DISABLED         | Workflow file is maintained and correct (refreshed 2026-09-12, commit `14523ae`) but `disabled_manually` on GitHub since 2026-07-05. Unblock: push + `gh workflow enable ci`.                                                                                                                                         |
| Credentialed sandbox integration tests                 | FULLY_FUNCTIONAL | Eight key-gated live tests in `sandbox_live_test.go` (smoke trio + rates shape, funding-404, SCA challenge shape, OTT status on the quarterly surface, CreateBalance idempotency); credentialed run green 2026-10-17, keyless runs SKIP cleanly; `sandbox-live.yml` stays dispatch-gated, key lives outside the repo. |
| `nix flake check`                                      | FULLY_FUNCTIONAL | Format + sandboxed test via the `go-standard` module (`checks.test` race + coverage)                                                                                                                                                                                                                                  |
| `nix fmt` (gofumpt + goimports + nixfmt)               | FULLY_FUNCTIONAL | `flake.nix` treefmt config                                                                                                                                                                                                                                                                                            |
| BDD tests via Ginkgo + httptest                        | FULLY_FUNCTIONAL | `wise_test.go`                                                                                                                                                                                                                                                                                                        |
| `nix run .#doc-verify` (links + godoc + count claims)  | FULLY_FUNCTIONAL | `flake.nix` app; gates the count claims in AGENTS.md/FEATURES.md against the compiled surface                                                                                                                                                                                                                         |
| OpenAPI spec-conformance gate                          | FULLY_FUNCTIONAL | `spec_conformance_test.go` + `zz_spec_conformance_coverage_test.go`; every mock exchange validated against `docs/reviews/wise-api-openapi.json` (v0.12.0)                                                                                                                                                             |
| `nix run .#apidiff` (gorelease vs latest tag)          | FULLY_FUNCTIONAL | `flake.nix` app; pins gorelease via `golang.org/x/exp`; public-surface delta report                                                                                                                                                                                                                                   |
| `.github/SECURITY.md` (vulnerability reporting)        | FULLY_FUNCTIONAL | private advisories, scope, 7-day triage; linked from CONTRIBUTING.md                                                                                                                                                                                                                                                  |

## Documentation

| Feature                           | Status           | Evidence                                                                                                      |
| --------------------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------- |
| Godoc examples for the public API | FULLY_FUNCTIONAL | `example_test.go`; 36 `Example*` funcs (compile-only doc examples + 3 runnable) covering every resource group |
| README API reference              | FULLY_FUNCTIONAL | All 16 resources documented with runnable snippets + TOC                                                      |

## Deferred architecture

| Feature                      | Status  | Notes                                          |
| ---------------------------- | ------- | ---------------------------------------------- |
| Service-client sub-structure | PLANNED | Trigger reached: 16 resources (see ROADMAP.md) |
