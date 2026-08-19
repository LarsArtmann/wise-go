# wise-go Status Report — Core Transfer Flow Completion (Phase 1.5)

**Date:** 2026-08-19 18:15 CEST
**Branch:** master
**Latest commit:** `dcd1194` docs: reflect completed core transfer flow in README and status reports
**Ahead of origin:** 0 (all pushed)
**Working tree:** clean

---

## a) FULLY DONE

1. **`CancelTransfer`** — `PUT /v1/transfers/{transferId}/cancel` (`transfers.go:231`).
   - Added `Client.put` HTTP helper (`client.go`).
   - Validates non-zero `TransferID` before the network (Rejection).
   - Returns the mapped cancelled `*Transfer`.
   - BDD tests: happy path (asserts `PUT` method, maps `cancelled` status) + zero-ID rejection without API call.

2. **`GetDeliveryEstimate`** — `GET /v1/delivery-estimates/{transferId}` (`delivery_estimates.go`).
   - New `raw.DeliveryEstimate` wire type (`internal/raw/types.go`).
   - New public `DeliveryEstimate` type (`EstimatedDeliveryDate time.Time` + formatted string).
   - Optional IANA `timezone` query parameter.
   - BDD tests: happy path (asserts `timezone=Asia/Singapore` forwarded, zone-suffixed timestamp parsed) + zero-ID rejection.

3. **`ValidateTransferRequirements`** — `POST /v1/transfer-requirements` (`transfer_requirements.go`).
   - New raw + public dynamic-form types: `TransferRequirement`, `TransferRequirementForm`, `TransferRequirementField` (key/name/type/required/refreshRequirementsOnChange/min-max-length/validationRegexp/valuesAllowed), `TransferRequirementValue`.
   - New `ValidateTransferRequirementsRequest` (targetAccount + quoteUuid required; optional `customerTransactionId`, `originatorLegalEntityType PRIVATE|BUSINESS`, `Details` block with reference/sourceOfFunds/sourceOfFundsOther/transferPurpose/transferPurposeSubTransferPurpose/transferPurposeInvoiceNumber/transferNature).
   - Client-side validation before network (Rejection), wire-omitempty details builder, full raw→public mapper chain.
   - BDD tests: happy path assertions on the decoded request body and the mapped dynamic form (required/refresh flags, maxLength, regexp, valuesAllowed enum) + missing-quote rejection.

4. **`Quote` expansion** (`types.go`, `quotes.go`, `internal/raw/types.go`).
   - `PaymentOptions []QuotePaymentOption` — disabled, estimatedDelivery (parsed), formattedEstimatedDelivery, `QuoteFee` (transferwise/payIn/discount/partner/total), source/target `Money`, payIn/payOut, payInProduct, feePercentage.
   - `Notices []QuoteNotice` — text, link, type (WARNING/INFO/BLOCKED; BLOCKED = must not create transfer).
   - `RateType` (FIXED/FLOATING), `ProvidedAmountType` (SOURCE/TARGET), `GuaranteedTargetAmountAllowed`, `GuaranteedTargetAmount`.
   - Bidirectional: `toWire` unchanged (request side), `mapQuote` expanded; extracted `parseQuoteCreated` + `mapQuoteMonetary` to keep the mapper under `funlen`.
   - BDD tests for payment options + notices mapping (fee total, cents conversion, enums, nil link).

5. **`parseWiseTimestamp` fourth layout** (`helpers.go`) — `2006-01-02T15:04:05.000+0000` (delivery-estimate style, milliseconds + numeric zone). AGENTS.md gotcha updated. Unit test row added + malformed-case test extracted (fixes pre-existing funlen).

6. **Mapper Corruption classification unit tests** (`internal_test.go`) — `TestMapQuoteParseErrorsAreCorruption` (bad createdTime, bad currency, bad payment-option delivery) and `TestMapDeliveryEstimateParseErrorIsCorruption` (incl. an explicit wire-layout regression guard).

7. **flake.nix** — added `delivery_estimates.go` + `transfer_requirements.go` to the buildGoModule fileset.

8. **Documentation** — CHANGELOG (Unreleased Added), FEATURES (5 rows moved FULLY_FUNCTIONAL, 2 PLANNED rows removed), TODO_LIST (3 items checked), AGENTS.md (timestamp gotcha), plan doc (tier-1 rows 1–4, 6–13 → DONE; only `GetQuoteAccountRequirements` remains), README (v0.8.0 status banner, TOC, Features bullets, Quotes/Recipients/Exchange-rates API-reference sections, full quote→recipient→transfer example with validation/delivery-estimate/cancel steps).

9. **Quality gates** — `go test ./...` 77 specs pass; `go test -race -count=1 ./...` pass; `GOEXPERIMENT=jsonv2 golangci-lint run` 0 issues; `nix flake check` all checks passed; BuildFlow pre-commit hook passed (warnings only).

10. **Commits** — 11 commits this session (5 daemon + 1 manual doc commit on top of 5 earlier session commits), all pushed to origin/master. Tree clean.

---

## b) PARTIALLY DONE

1. **`Quote` type** — expanded, but still omits: `pricingConfiguration` overrides (partner-only), the deep `price` breakdown tree (priceSetId, items, deferredFee, calculatedOn), `rateExpirationTime` (we surface `expirationTime` only), `user` field, `targetAmountAllowed`/`providedAmountType` request side. Consumer-visible display needs (fees + delivery + notices) are covered.
2. **`CreateTransferRequest`** — still a flat struct; no `originatorLegalEntityType` on the create request itself, no dynamic-details injection driven by the requirements response. `ValidateTransferRequirements` output is not yet machine-linked into the transfer request.
3. **`Recipient.Details`** — remains `map[string]string` (pragmatic). Typed currency-specific helpers still not built.
4. **Generic pagination** — `ListRecipients` + `ListTransfers` each hand-roll loops; no shared `Page[T]`.
5. **Correlation ID** — `WithRequestCorrelationID` per-request override still documented-but-unimplemented (options.go doc references it).
6. **Delivery-estimate timestamp** — parseable via the tolerant parser, but the parser's layout list is now 5 candidates; a future Wise format change needs another row (no structural fix).

---

## c) NOT STARTED

### Core transfer flow (remaining)

- `FundTransfer` (`POST /v1/profiles/{profileId}/transfers/{transferId}/payments`) — the last piece for a fully fleshed end-to-end flow (fund the transfer, then poll delivery).
- Automating requirement→transfer-details feedback loop (use `ValidateTransferRequirements` output to populate `CreateTransferRequest.Details`).

### Tier 2 (plan doc)

- `GetQuoteAccountRequirements` (`GET /v1/quotes/{quoteId}/account-requirements`) — only remaining tier-1 row in the plan.
- `GetMe`/`GetUser`, `GetStatement` formats, Webhook signature verification, `CreateBalance`, direct `GetBalance`, `GetTotalFunds`, `GetBankAccountDetails`, `GetMultiCurrencyAccount`, `ListCurrencies`, per-request correlation ID.

### Deeper tiers

- Batch groups, direct debit, bulk settlement, payin deposit details, third-party transfers, sandbox simulations.

### Observability & quality

- Sandbox credential integration tests (needs user's key), `WithLogger`/`WithMetrics` hooks, mTLS docs, context-aware retry cancellation, godoc examples for new public types, v1.0 API audit, property-based tests, Cachix cache.

---

## d) TOTALLY FUCKED UP

Nothing is broken or shipped dangerously. Honest missteps this session:

1. **Import-block thrash in `wise_test.go`** — the test file originally imported both `encoding/json/v2` and `stdjson "encoding/json/v2"` (duplicate package, ST1019). "Fixing" it I churned through 6+ iterations: renamed the alias to stdlib, sed-mangled `&body` into `&2`, briefly used `json/v2`'s non-existent `NewDecoder`, left mixed tab/space indentation twice, and had an auto-formatter (dprint/nix-fmt) revert two hand edits mid-stream. Root cause: I kept doing targeted `perl`/`sed` string surgery on a file an auto-formatter owns. The stable fix was `json.UnmarshalRead` (json/v2) for request-body decoding and standard `encoding/json` under `stdjson` for the real `NewDecoder` — but I only landed on it after several wasted round trips. Lesson: when a formatter rewrites files, use the `edit` tool with whole-block context and re-verify immediately, not sed one-liners.

2. **`strPtr`/`int32Ptr` newexpr churn** — added helper functions, got `modernize` newexpr findings, nolinted them, then the user pointed out Go 1.26 has `new(value)`; I verified it works and replaced with inline `new(int32(10))` / `new("...")`. The nolint-first approach was the wrong instinct — the language feature exists and is cleaner.

3. **`mapQuote` refactor induced transient breakage** — extracting `parseQuoteCreated` I first assigned 2 values from a 3-return function and referenced an undefined `expiration`; caught by build within one step but shows the extract pattern needs a build immediately after (AGENTS.md cross-cutting lesson — I failed to apply it once).

4. **Doc-file race with the auto-commit daemon** — I edited `FEATURES.md`, the status doc, and `wise_test.go` while the daemon committed/rewrote them mid-edit ("file has been modified since last read"), forcing re-reads and re-applies. Symptom of not re-reading right before the write.

5. **README/status doc touched twice** — the daemon's `docs(docs)` commit (6776d7d) had already covered README/CHANGELOG/FEATURES/TODO/AGENTS, so my later manual doc commit (dcd1194) was the tail (plan doc + status doc + leftover README section). Slight overlap/races; final state is correct and non-duplicated.

---

## e) WHAT WE SHOULD IMPROVE

1. **Own the formatter contract** — run `nix fmt`/`gofumpt`/`golines` BEFORE editing, verify after every edit with `gofmt -l`, and prefer whole-block `edit` over `perl -pi` for anything a formatter touches.
2. **Build immediately after refactor extractions** — the AGENTS.md lesson ("delete → build → fix callers") applies equally to extract-refactors; I skipped it once.
3. **Add the missing error-path BDD tests** — 400 validation, 409 cancellation-not-allowed, 404, SCA for the new POST/PUT endpoints (ListRecipients/ListTransfers have none either).
4. **Validated edge-case tests** — `CreateTransferRequest.validate` (missing customerTransactionId/quote/target), `ValidateTransferRequirementsRequest.validate`, quote amount/currency mismatch matrix; several exist, several don't.
5. **Extract `vendorHash` to `vendorHash.nix`** — BuildFlow nix-checker flags it; cleaner diffs when deps change.
6. **Tighten the auto-commit interplay** — for doc-heavy tasks, either stage-and-commit in one shot or pause the daemon; the mid-session rewrites cost several re-reads.
7. **Wire `ValidateTransferRequirements` output into a helper** that maps discovered required fields onto `CreateTransferRequest` (or document the manual mapping) — right now the discovery is one-way.
8. **Add godoc examples** for `CancelTransfer`, `GetDeliveryEstimate`, `ValidateTransferRequirements`, and the new `Quote` fields so `go doc` shows usage.
9. **Run `govulncheck`/vulnix findings review** — BuildFlow reports stdlib vulns (GO-2026-6218 resolvePath in net/url, +31 more); likely toolchain-fixable, but currently unassessed.
10. **Consider `FundTransfer` next** to close the last core-flow gap, then re-evaluate the service-client substructure threshold (ROADMAP ~8 resources; we are at 15 methods across 8 resources).

---

## f) Up to 50 Things to Get Done Next

### Immediate (close out the transfer flow)

1. `FundTransfer` (`POST /v1/profiles/{id}/transfers/{id}/payments`) — balance funding completes the E2E.
2. Error-path BDD tests for POST/PUT endpoints (400 validation, 409 cancellation, 404, SCA, 429 with Retry-After for write ops).
3. Validation edge-case unit tests: `CreateTransferRequest.validate`, `ValidateTransferRequirementsRequest.validate`, quote amount/currency matrix.
4. `GetTransfer` error tests (404, auth, SCA).
5. Wire requirements→transfer-details feedback (helper or documented pattern).
6. Extract `vendorHash.nix` from flake.nix.
7. Godoc examples for the four new APIs.
8. Update README coverage badge (94.8% is stale; measure after new tests).

### Near-term (high value, self-contained)

9. `GetQuoteAccountRequirements` (last tier-1 row).
10. `GetMe` / `GetUser`.
11. `GetStatement` with format parameter (CSV/PDF/XLSX).
12. Webhook signature verification helper.
13. `CreateBalance`.
14. Direct `GetBalance` by ID (new v4 endpoint).
15. `GetTotalFunds`.
16. `GetBankAccountDetails`.
17. `GetMultiCurrencyAccount`.
18. `ListCurrencies`.
19. Per-request correlation ID override (`WithRequestCorrelationID` via context).

### Sandbox & integration

20. Credentialed sandbox integration test workflow (needs user's sandbox key).
21. Verify `GetDeliveryEstimate` + `ValidateTransferRequirements` wire formats against real sandbox responses.
22. Sandbox simulation helpers (fund transfer, cross-border payment confirmations).

### Type-safety & ergonomics

23. Typed recipient-detail structs (e.g. `GBPBankDetails`, `IBANDetails`) or per-currency key constants.
24. `Quote` residual fields: `rateExpirationTime`, `targetAmountAllowed`, `user`.
25. `pricingConfiguration` request/response exposure (partner tier).
26. Generic `Page[T]` pagination abstraction for List* endpoints.
27. Property-based tests for Money/quote/transfer mapping (gopter).
28. `ProvidedAmountType`/`RateType` request-side options in `CreateQuoteRequest`.

### Observability & quality

29. `WithLogger` request/response logging hook.
30. `WithMetrics` hook.
31. mTLS docs / `WithMTLS` option (`api-mtls.wise.com`).
32. Context-aware retry cancellation.
33. API audit for exported symbols ahead of v1.0.
34. `govulncheck` findings triage.
35. CI: Cachix binary cache for `nix flake check`.
36. Coverage badge automation (CI upload).

### Medium-term (completeness)

37. `GetAccountRequirements` (recipient-first flow).
38. `CreateRecipient` with refund + email-recipient support.
39. `CheckAccountQuoteCompatibility`.
40. Batch groups API.
41. Direct debit accounts API.
42. Bulk settlement API.
43. Pay-in deposit details API.
44. Third-party transfers API.

### Long-term / specialized

45. Cards API.
46. Card orders / card transactions.
47. Spend limits & controls.
48. KYC review API.
49. SCA factor APIs (PIN, facemaps, OTP devices).
50. Disputes, digital wallet push provisioning, Cases, OAuth token helper.

---

## g) Questions I Cannot Answer Without You

1. **Should I implement `FundTransfer` next to complete the end-to-end flow, or move to sandbox integration tests (Q3) first?** `FundTransfer` is the last mock-testable core-flow piece; sandbox tests prove the wire formats for everything already shipped but need your credentials.

2. **For `Recipient.Details`, do you want typed currency-specific structs (GBP sort-code, IBAN, US routing, etc.) with a fallback map, or keep `map[string]string` and add per-corridor key constants/docs?** Typed structs are a breaking change to the v0.8 surface; constants are additive.

3. **Can you provide a sandbox API key (or a protected workflow with secrets) so the SDK can be verified against `api.wise-sandbox.com`?** All 15 endpoints are currently mock-tested only; the sandbox verification is the difference between "matches our mocks" and "actually interoperates with Wise".

---

_Report complete. Waiting for instructions._
