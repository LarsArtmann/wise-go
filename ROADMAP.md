# Roadmap

Long-term direction and raw ideas for wise-go. These are **not** actionable tasks
(actionable work lives in [TODO_LIST.md](TODO_LIST.md)) — they are the strategic
context that explains why those tasks exist and where the project is heading.

## Vision

wise-go aims to be the **canonical Go SDK for the Wise (TransferWise) API**: the
library that is obviously correct, obviously typed, and obviously safe to depend on.
Every monetary amount is `Money` (cents paired with `Currency`). Every entity ID is
branded. Every error is typed and classifiable. Every API call retries intelligently.

As of 2026-08-21, the type-safety redesign is complete, the core transfer flow is
live END TO END (quotes, recipients, transfers, funding, delivery estimates,
transfer-requirements and account-requirements validation, exchange rates), and
the tier-2 surface shipped: users, statement files in all six formats, webhook
signature verification, balance lifecycle (create/direct-get/total-funds),
Multi-Currency Account + bank details, and currency reference data — 31
endpoint methods, plus observability (WithLogger, per-request correlation IDs)
and write helpers under the existing retry/error architecture. The v1.0 audit
(`docs/reviews/2026-08-21_v1.0-api-audit.md`) found nothing blocking the tag.
The roadmap expands the surface along four axes — completeness, type-safety,
observability, and scale — while preserving the architectural decisions that make
the codebase maintainable.

## Axis 1: Completeness (API surface)

Today: tiers 1 and 2 of
`docs/planning/2026-08-19_wise-api-full-implementation-plan.md` are complete —
31 endpoint methods across 14 resources. The core transfer flow is live end to end:
quotes (including account requirements), recipients, transfers, funding, delivery
estimates, exchange rates, transfer-requirements validation.

### Shipped 2026-08-21 (previously near-term)

- **`FundTransfer`** (`POST /v1/profiles/{id}/transfers/{id}/payments`) — the
  money-movement loop is closed.
- **`GetQuoteAccountRequirements`** — bridges quotes → recipients.
- **Users read** — `GetMe` / `GetUser`.
- **Balances expanded** — `CreateBalance`, direct `GetBalance` endpoint,
  `GetTotalFunds`.
- **MCA & bank details** — `GetBankAccountDetails`, `GetMultiCurrencyAccount`.
- **Currencies** — `ListCurrencies`.
- **Webhooks** — `VerifyWebhookSignature` (RSA-SHA256 over the raw body,
  PKIX/PKCS#1 PEM keys).
- **Statements** — `GetStatement` in all six formats (CSV, PDF, XLSX, CAMT.053,
  MT940, QIF) via the raw-response path.

### Medium-term

- **Sandbox verification** — credentialed integration tests against
  `api.wise-sandbox.com`. The workflow and test skeleton are in place
  (manual dispatch, key-gated); blocked on a sandbox API key.
- **POST account-requirements refresh** — completes the recipient-side two-pass
  flow (in flight, 2026-08-21 hardening plan).

### Long-term

- **Tier 3/4 of the plan** — batch groups, direct debit, bulk settlement, cards,
  KYC, SCA factors, disputes. Only as consumer demand justifies.
- **Pagination generalization** — `ListRecipients` and `ListTransfers` hand-roll
  page loops; a shared `Page[T]` abstraction becomes worthwhile once a third
  paginated endpoint lands.

### Out of vision

- **Streaming / WebSocket** — Wise does not expose a streaming API. Do not speculatively build one.
- **ORM-like abstractions** — the SDK is a thin anti-corruption layer, not a
  domain model. Resist abstractions over Wise's resources.

## Axis 2: Type-safety (data-model evolution)

The v0.4.0 release shipped the major type-safety redesign: `Money`/`Currency` value
objects, enum normalization, raw-type encapsulation, and dead-code removal. The
full analysis lives at `docs/brainstorming/2026-07-18_data-model-review.html`.

### Shipped in v0.5.0

- **`DetailType` typed enum** — `ListTransactionsRequest.Type` is now `DetailType`
  (was `string`). The `DetailType*` constants are typed values, preventing consumers
  from sending invalid filter values to the Wise API at compile time.
- **Godoc examples** — testable `Example*` functions for `Money` and `Currency`.
- **Test coverage** — `toMoney` currency validation failure path, zero end-of-statement
  balance edge case.
- **CI hardening** — `nix flake check` no longer skips the build; gofumpt format
  check added to the lint job.

### Shipped in v0.4.0

- **`Money` value object** — paired `XxxCents int64` / `XxxCurrency string` collapsed
  into `Money { Cents, Currency }`. Mismatched currency/amount is unrepresentable.
- **`Currency` branded type** — `NewCurrency` validates ISO 4217 (3-letter uppercase
  ASCII). The empty-currency case is unconstructible via the constructor.
- **`internal/raw` boundary** — wire-format types hidden from consumers. Public
  surface shrinks to parsed result types only.
- **Enum normalization** — `BalanceType` values lowercased; all SDK enums follow one
  casing rule. `TransactionTypeUnknown` removed (never returned by the classifier).
- **Typed `InvestmentState` enum** and **exported `DetailType` constants**.

### v1.0 — the API lock

- **Lock the public API surface** — freeze exported symbols; future breaking changes
  require v2. The formal API audit and godoc review pass are complete
  (`docs/reviews/2026-08-21_v1.0-api-audit.md`, green); awaiting the maintainer's
  version call (0.9.0 first or straight to v1.0.0).

### Beyond

- **Sealed transaction union (maybe)** — an interface-based union
  (`CreditTx`/`DebitTx`/`ExchangeTx`) was rejected in the data-model review as
  wrong ergonomics for statement iteration. Revisit only if a real consumer need
  emerges.
- **Generic `Page[T]`** — if pagination lands. Today there is no pagination, so no
  type to genericize.

## Axis 3: Observability (operational hooks)

Today the SDK is observable: `WithLogger` provides structured request logging
(method, URL, status, duration, retry count), per-request correlation IDs flow
through the context (`WithRequestCorrelationID`), and context cancellation aborts
in-flight retries. Callers who need raw transport access can still wrap the HTTP
transport themselves (documented in README).

### Medium-term

- **Metrics hook** — `WithMetrics` option exposing counters/histograms for
  Prometheus or OpenTelemetry (request count, latency, retry count, error rate).

### Shipped

- **Request/response logging** — `WithLogger` + `RequestLog`/`RequestLogFunc`
  (2026-08-21).
- **Per-request correlation ID** — client-wide `WithCorrelationID` plus the
  per-call context override `WithRequestCorrelationID` (2026-08-21).
- **Context-aware retry** — cancellation propagates through the retry policy.
- **mTLS documentation** — dedicated README section (`api-mtls.wise.com`,
  `api-mtls.wise-sandbox.com`).
- **Exchange rates** — `GetExchangeRate` (`GET /v1/rates`) shipped in v0.8.0 with
  current and historical lookup.

## Axis 4: Scale (architecture)

The flat-package design is correct for the present scope. The trigger for evolution
is clear.

### Trigger: resource count crosses ~6–8 — REACHED

The SDK now has 14 resources (profiles, users, balances, multi-currency account,
bank details, transactions/statements, transfers, funding, quotes, recipients,
exchange rates, delivery estimates, transfer requirements, currencies) and 31
endpoint methods on the flat `client.X` surface — the core flow including
`FundTransfer` is complete (2026-08-21). The threshold documented here
has been crossed; the open question is WHEN to pay the refactor cost. The
recommended sequencing: v1.0 on the flat surface (the 2026-08-21 audit found
no structural blocker), then move to a **service-client sub-structure** in
one release cycle:

```go
client.Profiles().List(ctx)
client.Balances().List(ctx, profileID)
client.Transactions().List(ctx, req)
```

This is a within-package refactor for one release cycle (flat methods become thin
wrappers over the new service clients), then a v1.0/v2 breaking change that removes
the wrappers.

### Trigger: a consumer asks to inject a narrow service interface

The real-world trigger. If a downstream user wants to mock only
`wise.Client.ListProfiles` in their tests, that is the signal to extract a
`ProfileService` with a narrow interface. Today, consumers should define
consumer-side interfaces themselves (Go proverb: "accept interfaces, return
structs"); documented in README.

### Never

- **Per-resource Go modules** — `wise-profiles`, `wise-balances`, etc. Every
  consumer imports all three; the boundaries have no composability payoff.
- **Domain-core / infrastructure split** — one HTTP backend, one retry library.
  The seam would be unused.

## Release strategy

- **v0.x** — breaking changes accepted but coordinated. Each breaking release gets
  a migration table in CHANGELOG.md. v0.4.0 shipped the Money/Currency redesign;
  v0.5.0 added the `DetailType` typed enum; v0.6.x SCA support; v0.7.0 transfers
  read; v0.8.0 the full core transfer flow (quotes, recipients, transfers write,
  rates); v0.8.1 the outgoing-timestamp wire fix.
- **v1.0** — public API freeze. After v1.0, breaking changes require v2 and a
  deliberate migration path.

## Non-goals

- **Supporting non-Go languages** — out of scope; this is a Go SDK.
- **Re-implementing retries / circuit breakers** — `failsafe-go` does this well.
  Do not replace it without cause.
- **Auto-generation from OpenAPI** — Wise publishes an OpenAPI spec (downloaded to
  `docs/reviews/wise-api-openapi.json`); use it as the authoritative reference when
  hand-authoring types (it caught the UUID `QuoteID` mismatch). Auto-generating the
  SDK from it remains a non-goal: generated types would lose the two-layer
  raw/result boundary and branded-ID ergonomics.
- **Caching / local state** — the SDK is stateless. Caching is the caller's job.
- **`Money` arithmetic** — `Money` pairs cents+currency to prevent mismatched amounts
  at the serialization boundary. It is deliberately not a financial math library
  (`Add`/`Sub`/`IsNegative`/`Equal` etc. are out of scope). Arithmetic is the
  consumer's domain logic; the SDK is an anti-corruption layer, not a domain model.
