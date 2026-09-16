# Features

Honest inventory of wise-go features by status. Code is the source of truth — every
claim here can be verified against the implementation.

## Status vocabulary

| Status               | When it applies                                                                                                                                                                                        |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| FULLY_FUNCTIONAL     | Code present AND working (tests pass or exercised).                                                                                                                                                    |
| PARTIALLY_FUNCTIONAL | Ships but has known gaps, edge-case bugs, or missing pieces.                                                                                                                                           |
| BROKEN               | Code exists but does not work / is disabled / fails.                                                                                                                                                   |
| PLANNED              | Documented but **no code exists yet**; demand-gated per the implementation plan.                                                                                                                       |
| OUT_OF_SCOPE         | Wise surface wise-go intentionally does not target: specialised, partner-only (client credentials / restricted access), or sandbox-only. Revisit only on a real demand signal (see ROADMAP.md Axis 1). |

## Wise API endpoint coverage

Full audit of the live Wise Platform API reference
(`docs.wise.com/api-reference/preview`, audited **2026-09-16**) against the SDK.
This is the single source of truth for endpoint existence; the resource sections
below cover behavior (parsing, filtering, validation), not endpoint inventory.

- **215 documented REST operations** across **51 reference categories**, plus
  **29 webhook-event definitions** in the preview sweep (**31** distinct event
  types counting the settlement and Swift events from the webhook-event index).
- **wise-go ships 41 of those documented operations** through **38 of the 41
  public `*Client` methods** — the other three are `Authenticate`/`Health`
  (convenience wrappers) and `ClearSCAChallenge` (a multi-call loop);
  `TriggerOTT`/`VerifyOTT` each cover three channel paths. The count is
  gate-checked by `nix run .#doc-verify` against the compiled surface.
- **51 operations are PLANNED** (demand-gated, tiers 2–3 of
  [the implementation plan](docs/planning/2026-08-19_wise-api-full-implementation-plan.md)).
- **123 operations are OUT_OF_SCOPE** (cards, KYC/KYB, SCA factors, JOSE,
  simulations, embedded flows, cases, disputes, settlement allocations, …).

### Summary by category

| Category                                                                                                | Ops | Shipped | State                                            |
| ------------------------------------------------------------------------------------------------------- | --- | ------- | ------------------------------------------------ |
| Balances                                                                                                | 10  | 4       | core shipped; lifecycle ops PLANNED              |
| Balance statements                                                                                      | 1   | 1       | all 7 formats shipped                            |
| Bank account details                                                                                    | 5   | 1       | read shipped; ordering PLANNED                   |
| Batch groups                                                                                            | 7   | 0       | PLANNED (tier 3)                                 |
| Profiles                                                                                                | 19  | 2       | reads shipped; writes PLANNED; KYB OUT_OF_SCOPE  |
| Quotes                                                                                                  | 4   | 3       | complete except PATCH update (PLANNED)           |
| Recipients                                                                                              | 8   | 5       | core shipped; compatibility/confirmation PLANNED |
| Transfers (standard)                                                                                    | 11  | 8       | core + receipts shipped; 3 ops PLANNED           |
| Transfers (third-party)                                                                                 | 2   | 0       | PLANNED (correspondent demand)                   |
| Users                                                                                                   | 6   | 2       | reads shipped; writes OUT_OF_SCOPE               |
| Multi-currency account                                                                                  | 4   | 1       | MCA read shipped; configuration PLANNED          |
| Webhook subscriptions                                                                                   | 9   | 4       | profile-level shipped; app-level decision open   |
| SCA one-time tokens                                                                                     | 1   | 1       | shipped (v0.11.0)                                |
| SCA OTP channels                                                                                        | 8   | 6       | trigger/verify shipped; enrollment OUT_OF_SCOPE  |
| Currencies                                                                                              | 1   | 1       | shipped                                          |
| Exchange rates                                                                                          | 1   | 1       | shipped                                          |
| Delivery estimates                                                                                      | 1   | 1       | shipped                                          |
| Comparison                                                                                              | 1   | 0       | PLANNED                                          |
| Activity                                                                                                | 1   | 0       | PLANNED                                          |
| Contacts                                                                                                | 1   | 0       | PLANNED                                          |
| Addresses                                                                                               | 5   | 0       | PLANNED                                          |
| Payin deposit details                                                                                   | 1   | 0       | PLANNED                                          |
| Direct debit accounts                                                                                   | 2   | 0       | PLANNED                                          |
| Bulk settlement                                                                                         | 1   | 0       | PLANNED (needs client credentials)               |
| GPI tracking                                                                                            | 1   | 0       | PLANNED                                          |
| OAuth token                                                                                             | 1   | 0       | OUT_OF_SCOPE                                     |
| Claims, cases, KYC, SCA factors, cards, disputes, spend, JOSE, simulation, settlement, incoming, payins | 122 | 0       | OUT_OF_SCOPE                                     |

### Balances (10 operations)

| Method | Path                                                | Status           | wise-go                               |
| ------ | --------------------------------------------------- | ---------------- | ------------------------------------- |
| POST   | `/v4/profiles/{id}/balances`                        | FULLY_FUNCTIONAL | `CreateBalance` (`balances.go:101`)   |
| GET    | `/v4/profiles/{id}/balances`                        | FULLY_FUNCTIONAL | `ListBalances` (`balances.go:31`)     |
| GET    | `/v4/profiles/{id}/balances/{id}`                   | FULLY_FUNCTIONAL | `GetBalance` (`balances.go:61`)       |
| GET    | `/v4/profiles/{id}/total-funds/{currency}`          | FULLY_FUNCTIONAL | `GetTotalFunds` (`balances.go:170`)   |
| DELETE | `/v4/profiles/{id}/balances/{id}`                   | PLANNED          | close balance (zero balance required) |
| POST   | `/v4/profiles/{id}/balance-movements`               | PLANNED          | convert/move between balances         |
| GET    | `/v4/profiles/{id}/balance-capacity`                | PLANNED          | regulatory deposit limit              |
| POST   | `/v4/profiles/{id}/excess-money-account`            | PLANNED          | excess-funds sweep target             |
| GET    | `/v4/profiles/{id}/balances/hold-limit-breach`      | PLANNED          | hold-limit breaches (SG/MY)           |
| POST   | `/v4/profiles/{id}/balances/hold-limit-breach/{id}` | PLANNED          | close breach via one-time refund      |

### Balance statements (1 operation, 7 formats)

| Method | Path                                                                                     | Status           | wise-go                                                                                        |
| ------ | ---------------------------------------------------------------------------------------- | ---------------- | ---------------------------------------------------------------------------------------------- |
| GET    | `/v1/profiles/{id}/balance-statements/{id}/statement.{json,csv,pdf,xlsx,camt,mt940,qif}` | FULLY_FUNCTIONAL | `ListTransactions` (`transactions.go:16`), `GetStatement` raw download (`transactions.go:276`) |

### Bank account details (5 operations)

| Method | Path                                                      | Status           | wise-go                                            |
| ------ | --------------------------------------------------------- | ---------------- | -------------------------------------------------- |
| GET    | `/v1/profiles/{id}/account-details`                       | FULLY_FUNCTIONAL | `GetBankAccountDetails` (`account_details.go:144`) |
| POST   | `/v1/profiles/{id}/bank-details`                          | PLANNED          | issue local+international details                  |
| POST   | `/v1/profiles/{id}/account-details-orders`                | PLANNED          | order account details                              |
| GET    | `/v1/profiles/{id}/account-details-orders`                | PLANNED          | list orders                                        |
| POST   | `/v1/profiles/{id}/account-details/payments/{id}/returns` | PLANNED          | return a received payment                          |

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
| GET    | `/v1/rates`                           | FULLY_FUNCTIONAL | `GetExchangeRate` (`rates.go:16`)                  |
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
| POST   | `/v1/accounts`                                | FULLY_FUNCTIONAL | `CreateRecipient` (`recipients.go:106`)             |
| GET    | `/v1/accounts/{id}`                           | FULLY_FUNCTIONAL | `GetRecipient` (`recipients.go:88`)                 |
| GET    | `/v1/quotes/{id}/account-requirements`        | FULLY_FUNCTIONAL | `GetQuoteAccountRequirements` (`quotes.go:94`)      |
| POST   | `/v1/quotes/{id}/account-requirements`        | FULLY_FUNCTIONAL | `RefreshQuoteAccountRequirements` (`quotes.go:163`) |
| DELETE | `/v1/accounts/{id}`                           | PLANNED          | deactivate recipient                                |
| POST   | `/v1/accounts/{id}/quotes/{id}/compatibility` | PLANNED          | pre-transfer compatibility check                    |
| PATCH  | `/v1/accounts/{id}/confirmations`             | PLANNED          | accept verification mismatch                        |

### Transfers — standard (11 operations)

| Method | Path                                         | Status           | wise-go                                                        |
| ------ | -------------------------------------------- | ---------------- | -------------------------------------------------------------- |
| POST   | `/v1/transfers`                              | FULLY_FUNCTIONAL | `CreateTransfer` (`transfers.go:221`)                          |
| GET    | `/v1/transfers`                              | FULLY_FUNCTIONAL | `ListTransfers` (`transfers.go:39`)                            |
| GET    | `/v1/transfers/{id}`                         | FULLY_FUNCTIONAL | `GetTransfer` (`transfers.go:117`)                             |
| PUT    | `/v1/transfers/{id}/cancel`                  | FULLY_FUNCTIONAL | `CancelTransfer` (`transfers.go:257`)                          |
| POST   | `/v1/transfer-requirements`                  | FULLY_FUNCTIONAL | `ValidateTransferRequirements` (`transfer_requirements.go:20`) |
| POST   | `/v1/profiles/{id}/transfers/{id}/payments`  | FULLY_FUNCTIONAL | `FundTransfer` (`transfers.go:286`)                            |
| GET    | `/v1/transfers/{id}/invoices/bankingpartner` | FULLY_FUNCTIONAL | `GetTransferPayoutInfo` (`transfers.go:160`)                   |
| GET    | `/v1/transfers/{id}/receipt.pdf`             | FULLY_FUNCTIONAL | `GetTransferReceipt` (`transfers.go:138`)                      |
| GET    | `/v1/transfers/{id}/payments`                | PLANNED          | list completed funding payments                                |
| GET    | `/v1/transfers/{id}/us-combined-receipt.pdf` | PLANNED          | US combined tax receipt                                        |
| GET    | `/v1/transfers/{id}/documents/noc`           | PLANNED          | no-objection certificate (India FIRC)                          |

### Transfers — third-party (2 operations)

| Method | Path                                           | Status  | wise-go                                       |
| ------ | ---------------------------------------------- | ------- | --------------------------------------------- |
| POST   | `/v2/profiles/{id}/third-party-transfers`      | PLANNED | originator transfers (correspondent partners) |
| GET    | `/v2/profiles/{id}/third-party-transfers/{id}` | PLANNED | third-party transfer status                   |

### Webhook subscriptions (9 operations)

| Method | Path                                                              | Status           | wise-go                                                |
| ------ | ----------------------------------------------------------------- | ---------------- | ------------------------------------------------------ |
| POST   | `/2026Q3/profiles/{id}/subscriptions`                             | FULLY_FUNCTIONAL | `CreateProfileWebhookSubscription` (`webhooks.go:90`)  |
| GET    | `/2026Q3/profiles/{id}/subscriptions`                             | FULLY_FUNCTIONAL | `ListProfileWebhookSubscriptions` (`webhooks.go:117`)  |
| GET    | `/2026Q3/profiles/{id}/subscriptions/{id}`                        | FULLY_FUNCTIONAL | `GetProfileWebhookSubscription` (`webhooks.go:149`)    |
| DELETE | `/2026Q3/profiles/{id}/subscriptions/{id}`                        | FULLY_FUNCTIONAL | `DeleteProfileWebhookSubscription` (`webhooks.go:178`) |
| POST   | `/applications/{clientKey}/subscriptions`                         | PLANNED          | app-level scope decision pending (ROADMAP.md Axis 1)   |
| GET    | `/applications/{clientKey}/subscriptions`                         | PLANNED          | app-level scope decision pending                       |
| GET    | `/applications/{clientKey}/subscriptions/{id}`                    | PLANNED          | app-level scope decision pending                       |
| DELETE | `/applications/{clientKey}/subscriptions/{id}`                    | PLANNED          | app-level scope decision pending                       |
| POST   | `/applications/{clientKey}/subscriptions/{id}/test-notifications` | PLANNED          | app-level scope decision pending                       |

### SCA one-time tokens and OTP channels (9 operations)

| Method | Path                                         | Status           | wise-go                                                |
| ------ | -------------------------------------------- | ---------------- | ------------------------------------------------------ |
| GET    | `/2026Q3/one-time-token/status`              | FULLY_FUNCTIONAL | `GetOTTStatus` (`ott.go:188`)                          |
| POST   | `/2026Q3/one-time-token/sms/trigger`         | FULLY_FUNCTIONAL | `TriggerOTT` (`ott.go:209`, `OTTChannelSMS`)           |
| POST   | `/2026Q3/one-time-token/sms/verify`          | FULLY_FUNCTIONAL | `VerifyOTT` (`ott.go:235`, `OTTChannelSMS`)            |
| POST   | `/2026Q3/one-time-token/whatsapp/trigger`    | FULLY_FUNCTIONAL | `TriggerOTT` (`ott.go:209`, `OTTChannelWhatsApp`)      |
| POST   | `/2026Q3/one-time-token/whatsapp/verify`     | FULLY_FUNCTIONAL | `VerifyOTT` (`ott.go:235`, `OTTChannelWhatsApp`)       |
| POST   | `/2026Q3/one-time-token/voice/trigger`       | FULLY_FUNCTIONAL | `TriggerOTT` (`ott.go:209`, `OTTChannelVoice`)         |
| POST   | `/2026Q3/one-time-token/voice/verify`        | FULLY_FUNCTIONAL | `VerifyOTT` (`ott.go:235`, `OTTChannelVoice`)          |
| POST   | `/application/users/{id}/phone-numbers`      | OUT_OF_SCOPE     | restricted access (client credentials + Wise approval) |
| DELETE | `/application/users/{id}/phone-numbers/{id}` | OUT_OF_SCOPE     | restricted access (client credentials + Wise approval) |

### Batch groups (7 operations) — PLANNED

Create/get/complete-or-cancel a group, add transfers, fund from balance,
direct-debit payment initiation create/get. All seven PLANNED (tier 3 of the
plan; bulk payments demand-gated).

### Remaining single-operation categories (PLANNED, demand-gated)

| Category              | Method + Path                                                                                            | Status  | wise-go                    |
| --------------------- | -------------------------------------------------------------------------------------------------------- | ------- | -------------------------- |
| Activity              | GET `/v1/profiles/{id}/activities`                                                                       | PLANNED | account activity feed      |
| Contacts              | POST `/v1/profiles/{id}/contacts`                                                                        | PLANNED | Wisetag/email/phone lookup |
| Addresses             | POST `/v1/addresses`, GET `/v1/addresses`, GET `/v1/addresses/{id}`, GET/POST `/v1/address-requirements` | PLANNED | address book (5 ops)       |
| Payin deposit details | GET `/v1/profiles/{id}/transfers/{id}/deposit-details/bank-transfer`                                     | PLANNED | bank-transfer pay-in info  |
| Direct debit accounts | POST/GET `/v1/profiles/{id}/direct-debit-accounts`                                                       | PLANNED | ACH/EFT funding accounts   |
| Bulk settlement       | POST `/settlements`                                                                                      | PLANNED | needs client credentials   |
| GPI tracking          | POST `/v1/gpi-tracking`                                                                                  | PLANNED | SWIFT gpi lookup           |

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
  `types.go:640-672`, verified against Wise's live webhook-event reference
  2026-09-13); unknown event types pass through by design.
- **Typed payloads for 3 events**: `transfers#state-change`,
  `transfers#payout-failure`, `balances#credit` (`ParseWebhookEvent`).

## Client core

| Feature                                 | Status           | Evidence                                                                                |
| --------------------------------------- | ---------------- | --------------------------------------------------------------------------------------- |
| `wise.New(apiKey, opts...)` constructor | FULLY_FUNCTIONAL | `client.go:43`; functional-options pattern                                              |
| Bearer-token authentication             | FULLY_FUNCTIONAL | `client.go:355` `setHeaders`                                                            |
| Sandbox environment (`WithSandbox`)     | FULLY_FUNCTIONAL | `options.go:64`; `SandboxURL` const in `types.go:24`                                    |
| Custom base URL (`WithBaseURL`)         | FULLY_FUNCTIONAL | `options.go:71`                                                                         |
| Custom HTTP timeout (`WithTimeout`)     | FULLY_FUNCTIONAL | `options.go:78`                                                                         |
| Custom retry policy (`WithRetry`)       | FULLY_FUNCTIONAL | `options.go:86`; exponential backoff via failsafe-go                                    |
| Custom HTTP client (`WithHTTPClient`)   | FULLY_FUNCTIONAL | `options.go:97`; accepts `Doer` interface (`client.go:27`)                              |
| Correlation ID (`WithCorrelationID`)    | FULLY_FUNCTIONAL | `options.go:110`; sets `X-External-Correlation-Id` header on all requests               |
| Retry with backoff (429, 5xx, network)  | FULLY_FUNCTIONAL | `client.go:100` `isRetryable`; verified by wise_test.go retry suite                     |
| `Authenticate(ctx)`                     | FULLY_FUNCTIONAL | `client.go:120`; delegates to `ListProfiles`                                            |
| `Health(ctx)`                           | FULLY_FUNCTIONAL | `client.go:130`; delegates to `Authenticate`                                            |
| `WithUserAgent` option (v0.11.0)        | FULLY_FUNCTIONAL | `options.go:81`; custom `User-Agent` on every request, default preserved (BDD)          |
| Shared-client concurrency (v0.11.0)     | FULLY_FUNCTIONAL | `wise_test.go:246`; 16 parallel requests, per-request correlation-ID isolation, `-race` |

## Transaction behavior

| Feature                                                  | Status               | Evidence                                                                             |
| -------------------------------------------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| `ListTransactionsRequest.Type` filter forwarding         | FULLY_FUNCTIONAL     | `transactions.go:33`; BDD-tested                                                     |
| Request validation (empty currency, inverted date range) | FULLY_FUNCTIONAL     | `transactions.go:208`; returns `wise.transactions.invalid_request` (`:204`)          |
| Transaction type classification                          | FULLY_FUNCTIONAL     | `transactions.go:177`; CARD_PAYMENT / CARD_REFUND split fixed 2026-07-18             |
| Cross-currency transaction mapping                       | FULLY_FUNCTIONAL     | `transactions.go:135` `mapExchange`; uses transaction currency, not request currency |
| `Transaction.Exchange` (`*TransactionExchange`)          | FULLY_FUNCTIONAL     | `types.go:119`; nil for non-conversion transactions                                  |
| `EndOfStatementBalance` exposure                         | FULLY_FUNCTIONAL     | `types.go:192`; surfaced as `Money` on `ListTransactionsResponse`                    |
| `Transaction.Date` UTC semantics                         | PARTIALLY_FUNCTIONAL | `helpers.go:75` parses as UTC; Wise sends no TZ — documented in field comment        |
| Pagination                                               | PLANNED              | Wise returns all transactions in one response; no endpoint requires it yet           |

## Balance filtering behavior

| Feature                                       | Status           | Evidence                                                       |
| --------------------------------------------- | ---------------- | -------------------------------------------------------------- |
| Filter visible + non-investment balances      | FULLY_FUNCTIONAL | `balances.go:43`; drops `Visible: false` and invested balances |
| Fetch hidden or invested balances             | FULLY_FUNCTIONAL | `GetBalance` direct endpoint retrieves them individually       |
| Profile name construction (personal/business) | FULLY_FUNCTIONAL | `profiles.go:31` name/BusinessName branches                    |

## Error handling

| Feature                                           | Status           | Evidence                                                            |
| ------------------------------------------------- | ---------------- | ------------------------------------------------------------------- |
| `APIError` base type                              | FULLY_FUNCTIONAL | `errors.go:42`; embeds into all subtypes; carries `Headers` (`:49`) |
| `RateLimitError` (HTTP 429) with `RetryAfter`     | FULLY_FUNCTIONAL | `errors.go:71`; parses delta-seconds + HTTP-date                    |
| `RateLimitError.RateLimitedBy` (429 header)       | FULLY_FUNCTIONAL | `errors.go:75`; captures `X-Rate-Limited-By` header                 |
| `AuthError` (HTTP 401, 403)                       | FULLY_FUNCTIONAL | `errors.go:103`                                                     |
| `SCAChallengeError` (403 + 2FA headers, v0.6.0)   | FULLY_FUNCTIONAL | `errors.go:116`; `TwoFAApprovalToken()` returns the OTT             |
| `WithSCAApprovalToken` option (v0.6.0)            | FULLY_FUNCTIONAL | `options.go:142`; sends cleared OTT as `x-2fa-approval`             |
| `NotFoundError` (HTTP 404)                        | FULLY_FUNCTIONAL | `errors.go:142`                                                     |
| `ServerError` (HTTP 5xx)                          | FULLY_FUNCTIONAL | `errors.go:151`                                                     |
| `ErrorCode()` / `ErrorFamily()` / `IsRetryable()` | FULLY_FUNCTIONAL | All implement go-error-family interfaces                            |
| `errors.As` matching                              | FULLY_FUNCTIONAL | Demonstrated in README; tested                                      |

## Type system

| Feature                                               | Status           | Evidence                                                                                  |
| ----------------------------------------------------- | ---------------- | ----------------------------------------------------------------------------------------- |
| Branded `ProfileID` / `BalanceID`                     | FULLY_FUNCTIONAL | `ids.go:32,38`; phantom types prevent entity-ID mixing at compile time                    |
| Branded `TransactionID`                               | FULLY_FUNCTIONAL | `ids.go:41`                                                                               |
| `Money` value object (`Cents` + `Currency`)           | FULLY_FUNCTIONAL | `types.go:59`; paired cents/currency makes mismatch unrepresentable                       |
| `Currency` branded type with ISO 4217 validation      | FULLY_FUNCTIONAL | `types.go:39` `NewCurrency`; validates 3-letter uppercase ASCII                           |
| Two-layer raw/result split (`internal/raw` boundary)  | FULLY_FUNCTIONAL | Wire types in `internal/raw/types.go`; parsed types in `types.go`; `helpers.go:47` bridge |
| `ProfileType`, `BalanceType`, `TransactionType` enums | FULLY_FUNCTIONAL | `types.go:139,147,164`                                                                    |
| `InvestmentState` typed enum                          | FULLY_FUNCTIONAL | `types.go:155`; used for balance filtering (`balances.go:43`)                             |
| `DetailType` typed enum + constants                   | FULLY_FUNCTIONAL | `transactions.go:159`; typed filter for `ListTransactionsRequest.Type`                    |
| Enum casing normalization (lowercase SDK values)      | FULLY_FUNCTIONAL | `BalanceType` normalized; `ProfileType`/`TransactionType` already lowercase               |

## Wire format hardening

| Feature                                        | Status           | Evidence                                                  |
| ---------------------------------------------- | ---------------- | --------------------------------------------------------- |
| Tolerant timestamp parsing (4 layouts)         | FULLY_FUNCTIONAL | `helpers.go` `parseWiseTimestamp`; zoneless = UTC         |
| Outgoing query timestamps as UTC `Z` (v0.8.1)  | FULLY_FUNCTIONAL | `helpers.go:148` `formatWiseTimestamp`; Wise 422s offsets |
| `ListBalances` `types` query param (v0.5.3)    | FULLY_FUNCTIONAL | `balances.go:18,23`; live API 400s without it             |
| Mapper errors classified `Corruption` (v0.5.2) | FULLY_FUNCTIONAL | `internal_test.go`; fail-fast instead of retry loops      |

## Webhook behavior

| Feature                                                 | Status           | Evidence                                                                                              |
| ------------------------------------------------------- | ---------------- | ----------------------------------------------------------------------------------------------------- |
| `ParseWebhookPublicKey` (PKIX + PKCS#1 PEM, RSA-only)   | FULLY_FUNCTIONAL | `webhooks.go:33`; garbage/non-RSA input rejected with clear errors                                    |
| `VerifyWebhookSignature` (RSA-SHA256 over raw body)     | FULLY_FUNCTIONAL | `webhooks.go:64`; valid/tampered/wrong-key/malformed/empty/5 MiB tested                               |
| `HeaderWebhookSignature` / `HeaderDeliveryID` constants | FULLY_FUNCTIONAL | `webhooks.go:22,24`; delivery-dedup guidance in README Webhooks section                               |
| `ClearSCAChallenge` convenience loop                    | FULLY_FUNCTIONAL | `ott.go:284`; triggers, prompts via callback, verifies, repeats until every challenge passed          |
| `OTTChannel` typed channel enum                         | FULLY_FUNCTIONAL | `ott.go:24`; single lowercase-wire → UPPERCASE-challenge-type mapping                                 |
| OTT secrecy: token never in error strings               | FULLY_FUNCTIONAL | `SCAChallengeError.Error()` reports `token issued: true/false` only; value via `TwoFAApprovalToken()` |

## Build & tooling

| Feature                                                | Status               | Evidence                                                                                                                                                                                 |
| ------------------------------------------------------ | -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `go test ./...`                                        | FULLY_FUNCTIONAL     | httptest mocks; no network required                                                                                                                                                      |
| `golangci-lint run`                                    | FULLY_FUNCTIONAL     | 0 issues                                                                                                                                                                                 |
| `.github/workflows/ci.yml` (build/test/lint/vulncheck) | PARTIALLY_FUNCTIONAL | Workflow file maintained (refreshed 2026-09-12, commit `14523ae`) but still `disabled_manually` on GitHub since 2026-07-05; no runs since. `sandbox-live.yml` is active (dispatch-gated) |
| `nix flake check`                                      | FULLY_FUNCTIONAL     | Format + sandboxed test via the `go-standard` module (`checks.test` race + coverage)                                                                                                     |
| `nix fmt` (gofumpt + goimports + nixfmt)               | FULLY_FUNCTIONAL     | `flake.nix` treefmt config                                                                                                                                                               |
| BDD tests via Ginkgo + httptest                        | FULLY_FUNCTIONAL     | `wise_test.go`                                                                                                                                                                           |
| `nix run .#doc-verify` (links + godoc + count claims)  | FULLY_FUNCTIONAL     | `flake.nix` app; gates the count claims in AGENTS.md/FEATURES.md against the compiled surface                                                                                            |

## Documentation

| Feature                           | Status           | Evidence                                                                                                      |
| --------------------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------- |
| Godoc examples for the public API | FULLY_FUNCTIONAL | `example_test.go`; 24 `Example*` funcs (compile-only doc examples + 3 runnable) covering every resource group |
| README API reference              | FULLY_FUNCTIONAL | All 16 resources documented with runnable snippets + TOC                                                      |

## Deferred (demand-gated, not started)

| Feature                      | Status  | Notes                                          |
| ---------------------------- | ------- | ---------------------------------------------- |
| Service-client sub-structure | PLANNED | Trigger reached: 16 resources (see ROADMAP.md) |
