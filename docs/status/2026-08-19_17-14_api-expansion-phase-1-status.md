# wise-go Status Report — API Expansion Phase 1

**Date:** 2026-08-19 17:14  
**Branch:** master  
**Latest commit:** `341bf37` feat(api): expand SDK with quotes, recipients, transfers, and rates endpoints  
**Ahead of origin:** 5 commits (pushed)

---

## a) FULLY DONE

1. **Comprehensive API implementation plan** at `docs/planning/2026-08-19_wise-api-full-implementation-plan.md`. Maps all ~135 endpoints across 29 Wise API categories, applies Pareto prioritisation (1% / 4% / 20% / 80%), and includes a Mermaid dependency graph.
2. **Generic HTTP write helper** in `client.go`: `Client.request` + `Client.doRequest` support POST bodies, retries, auth headers, correlation ID, SCA token, and response decoding under the existing failsafe-go retry policy.
3. **GetProfile** — `GET /v2/profiles/{id}` (`profiles.go`).
4. **GetExchangeRate** — `GET /v1/rates` (`rates.go`) with current/historical query params and currency validation.
5. **GetTransfer** — `GET /v1/transfers/{id}` (`transfers.go`).
6. **Quotes API** — `CreateUnauthenticatedQuote`, `CreateQuote`, `GetQuote` (`quotes.go`). Includes `QuoteID` (UUID string), `Quote`, `CreateQuoteRequest`, `PayIn`/`PayOut` enums, `QuoteStatus`.
7. **Recipients API** — `ListRecipients`, `GetRecipient`, `CreateRecipient` (`recipients.go`). Includes `Recipient`, `CreateRecipientRequest`, `ListRecipientsRequest` with automatic pagination.
8. **CreateTransfer** — `POST /v1/transfers` (`transfers.go`). Includes `CreateTransferRequest` with idempotency key and optional transfer details.
9. **`fetchByID` helper** in `helpers.go` to eliminate duplicated get-by-ID boilerplate.
10. **BDD tests** for every new endpoint in `wise_test.go` using `httptest` mocks.
11. **Documentation updates:** `CHANGELOG.md`, `FEATURES.md`, `TODO_LIST.md`, `AGENTS.md`.
12. **Nix build updates:** added new `.go` files to `flake.nix` fileset, updated `vendorHash`.
13. **Quality gates passed:** `go test -race ./...`, `golangci-lint run`, `nix flake check`.
14. **Changes committed and pushed** to GitHub.

---

## b) PARTIALLY DONE

1. **Full API coverage** — only the 1% Pareto core (transfer flow) is implemented. The plan documents the remaining ~130 endpoints.
2. **`Quote` type** — simplified to essential fields (ID, Source/Target Money, PayIn/PayOut, Rate, Created, ExpirationTime, Status, Profile). Does not yet expose `paymentOptions`, fees, notices, `guaranteedTargetAmount`, etc.
3. **`Recipient.Details`** — exposed as `map[string]string`. This is pragmatic but loses the typed, currency-specific shape Wise returns. Callers must know the required fields for their corridor.
4. **`CreateTransferRequest`** — covers the common case plus reference/source-of-funds/transfer-purpose fields. Does not yet model the full transfer-requirements validation flow.
5. **OpenAPI spec** — discovered mid-implementation, downloaded to `docs/reviews/wise-api-openapi.json` and used to correct the quote-ID type. Not yet used to generate or validate all types.
6. **Pagination** — implemented for `ListRecipients` and already existed for `ListTransfers`. No generic `Page[T]` abstraction yet.

---

## c) NOT STARTED

### Core transfer flow completion
- `CancelTransfer` (`PUT /v1/transfers/{id}/cancel`)
- `GetDeliveryEstimate` (`GET /v1/delivery-estimates/{id}`)
- `ValidateTransferRequirements` (`POST /v1/transfer-requirements`)
- Fund transfer (`POST /v1/profiles/{id}/transfers/{id}/payments`)

### Statements
- CSV / PDF / XLSX / CAMT.053 / MT940 / QIF statement formats

### Users
- `GET /me`, `GET /users/{id}`

### Balances expanded
- `CreateBalance`
- `GetBalance` via direct endpoint (`GET /v4/profiles/{id}/balances/{id}`)
- `GetTotalFunds`

### Bank account details & MCA
- `GetBankAccountDetails`
- `GetMultiCurrencyAccount`

### Webhooks
- Signature verification helper
- Subscription CRUD

### Batch groups, direct debit, bulk settlement

### Card issuance, KYC, SCA, disputes, digital wallets

---

## d) TOTALLY FUCKED UP

Nothing is broken or shipped in a dangerous state. Honest missteps:

1. **QuoteID started as `int64`** — I initially assumed quote IDs were numeric like other Wise IDs. The OpenAPI spec revealed they are UUID strings, requiring a mid-implementation type change in `ids.go`, `internal/raw/types.go`, `types.go`, `quotes.go`, and tests.
2. **Linter churn** — `funlen`, `cyclop`, `unused`, `wrapcheck`, `gci`, and `varnamelen` warnings required several refactor rounds on `client.go`, `helpers.go`, and `quotes.go`. This was avoidable with closer attention to the existing `.golangci.yml` config up front.
3. **Auto-commit granularity** — the auto-git daemon produced 5 commits for one logical work unit. The history is correct but noisier than a single curated commit would have been.
4. **OpenAPI discovery timing** — finding the downloadable OpenAPI spec (`index.json`) after writing the first draft of types meant some rework. A quick check for machine-readable specs before hand-authoring types would have saved time.

---

## e) WHAT WE SHOULD IMPROVE

1. **Start future endpoints from the OpenAPI spec** rather than the prose API reference. The spec is authoritative for field types, optionality, and ID formats.
2. **Expand `Quote` to include `paymentOptions`** and fee breakdowns — these are essential for consumers to present pay-in/pay-out choices.
3. **Add typed recipient-detail helpers** or at least currency-specific constants/docs so consumers are not guessing field names.
4. **Add unit tests for validation edge cases** (e.g. mismatched quote currencies, missing customerTransactionId, empty recipient details).
5. **Add error-response tests** for the new POST endpoints (400 validation, 409 duplicate transfer).
6. **Add README examples** showing the full flow: create quote → create recipient → create transfer.
7. **Add sandbox integration tests** with real credentials to verify wire formats against Wise's sandbox.
8. **Consider a service-client substructure** once the resource count clearly exceeds the ~8 threshold documented in `ROADMAP.md`.
9. **Add per-request correlation ID support** (`WithRequestCorrelationID` via context or request option).
10. **Run the full API surface through the OpenAPI spec** to identify any other ID-type mismatches or missing required fields.

---

## f) Up to 50 Things to Get Done Next

### Immediate (core transfer flow — highest value)
1. `CancelTransfer`
2. `GetDeliveryEstimate`
3. `ValidateTransferRequirements`
4. `FundTransfer` (balance funding)
5. Expand `Quote` with `paymentOptions`, fees, and notices
6. Add README example for quote → recipient → transfer
7. Add error-response BDD tests for POST endpoints
8. Add validation edge-case unit tests
9. Add `GetTransfer` error tests (404, auth, SCA)
10. Add `CreateQuote` validation tests

### Near-term (high value, self-contained)
11. `GetMe` / `GetUser`
12. `GetStatement` with format parameter (CSV/PDF/XLSX)
13. Webhook signature verification helper
14. `CreateBalance`
15. Direct `GetBalance` by ID (new endpoint)
16. `GetTotalFunds`
17. `GetBankAccountDetails`
18. `GetMultiCurrencyAccount`
19. `ListCurrencies`
20. Per-request correlation ID override

### Medium-term (completeness)
21. `GetQuoteAccountRequirements`
22. `GetAccountRequirements` (recipient-first flow)
23. `CreateRecipient` with refund and email recipient support
24. `CheckAccountQuoteCompatibility`
25. Batch groups API
26. Direct debit accounts API
27. Bulk settlement API
28. Pay-in deposit details API
29. Third-party transfers API
30. Sandbox simulation helpers

### Observability & quality
31. Sandbox integration tests workflow
32. Request/response logging hook (`WithLogger`)
33. Metrics hook (`WithMetrics`)
34. mTLS documentation / `WithMTLS` option
35. Context-aware retry cancellation
36. Add godoc examples for new public types
37. API audit for exported symbols ahead of v1.0
38. Full OpenAPI-derived type review
39. Add property-based tests for Money/Quote/Transfer mapping
40. CI speed: add Cachix binary cache

### Long-term / specialized
41. Cards API
42. Card orders API
43. Card transactions API
44. Spend limits & controls
45. KYC review API
46. SCA factor APIs (PIN, facemaps, device fingerprints, OTP)
47. Disputes API
48. Digital wallet push provisioning
49. Cases / partner support API
50. OAuth token helper

---

## g) Questions I Cannot Answer Without You

1. **Should I continue expanding the core transfer flow next** (`CancelTransfer`, `GetDeliveryEstimate`, `ValidateTransferRequirements`) **or pivot to a different high-value area** like Webhooks, Statements, or Sandbox integration tests?

2. **For the polymorphic `Recipient.Details` field, do you want typed currency-specific structs** (e.g. `GBPBankDetails`, `IBANDetails`) **or keep the `map[string]string` approach** for flexibility?

3. **Should sandbox integration tests be written before adding more endpoints**, or should the SDK grow the mock-tested surface first and add live tests later?

---

_Report complete. Waiting for instructions._
