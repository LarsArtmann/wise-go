# Roadmap

Long-term direction and raw ideas for wise-go. These are **not** actionable tasks
(actionable work lives in [TODO_LIST.md](TODO_LIST.md)) — they are the strategic
context that explains why those tasks exist and where the project is heading.

## Vision

wise-go aims to be the **canonical Go SDK for the Wise (TransferWise) API**: the
library that is obviously correct, obviously typed, and obviously safe to depend on.
Every monetary amount is `Money` (cents paired with `Currency`). Every entity ID is
branded. Every error is typed and classifiable. Every API call retries intelligently.

As of v0.4.0, the type-safety redesign is shipped: paired `XxxCents`/`XxxCurrency`
fields are collapsed into `Money`, raw wire types are hidden behind `internal/raw`,
and enum values are normalized. The SDK covers the read side of three resources
(profiles, balances, transactions). The roadmap expands the surface along four
axes — completeness, type-safety, observability, and scale — while preserving the
architectural decisions that make the codebase maintainable.

## Axis 1: Completeness (API surface)

Today: read-only, 3 resources. The Wise API is much larger.

### Near-term (post-v0.4.0)

- **Write operations** — POST/PATCH/DELETE helpers in `client.go` mirroring
  `get`/`getWithQuery`. First candidates: transfers, recipients.
- **Quotes API** — `ListQuotes` / `CreateQuote`. Quotes are required before creating
  transfers, so they unblock the transfers workstream.
- **Recipients API** — `ListRecipients` / `CreateRecipient`. Required for transfers
  to non-Wise accounts.

### Medium-term

- **Webhooks** — `VerifyWebhookSignature` helper + typed webhook event structs.
  Wise's webhook signature scheme is well-documented; the helper is high-value and
  self-contained.
- **Statements (CSV/PDF)** — `GetStatement` with format parameter. Today the SDK
  consumes `statement.json`; the API also offers `statement.csv` and `statement.pdf`.

### Long-term

- **Pagination support** — if Wise ever moves to cursor-based pagination for
  transactions. The SDK can grow a `Page[Cursor]` type and `FetchMore` method
  without breaking the existing API. Today there is no pagination.

### Out of vision

- **Streaming / WebSocket** — Wise does not expose a streaming API. Do not speculatively build one.
- **ORM-like abstractions** — the SDK is a thin anti-corruption layer, not a
  domain model. Resist abstractions over Wise's resources.

## Axis 2: Type-safety (data-model evolution)

The v0.4.0 release shipped the major type-safety redesign: `Money`/`Currency` value
objects, enum normalization, raw-type encapsulation, and dead-code removal. The
full analysis lives at `docs/brainstorming/2026-07-18_data-model-review.html`.

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

### Near-term refinements

- **`ListTransactionsRequest.Type` typed enum** — currently `string`; should be a
  typed enum matching the exported `DetailType*` constants to prevent consumers
  from sending invalid filter values to the Wise API.

### v1.0 — the API lock

- **Lock the public API surface** — freeze exported symbols; future breaking changes
  require v2. Needs a formal API audit and godoc review pass before tagging.

### Beyond

- **Sealed transaction union (maybe)** — an interface-based union
  (`CreditTx`/`DebitTx`/`ExchangeTx`) was rejected in the data-model review as
  wrong ergonomics for statement iteration. Revisit only if a real consumer need
  emerges.
- **Generic `Page[T]`** — if pagination lands. Today there is no pagination, so no
  type to genericize.

## Axis 3: Observability (operational hooks)

Today the SDK is a black box: callers cannot inspect requests, responses, or retry
decisions without wrapping the HTTP transport themselves (documented in README).

### Near-term

- **Request/response logging hook** — `WithLogger` option for structured request
  logging (method, URL, status, duration, retry count).
- **Request ID propagation** — `X-Request-ID` header injection for distributed tracing.
- **Context-aware retry** — thread `context.Context` cancellation through the retry
  policy so callers can abort in-flight retries.

### Medium-term

- **Metrics hook** — `WithMetrics` option exposing counters/histograms for
  Prometheus or OpenTelemetry (request count, latency, retry count, error rate).
- **mTLS documentation** — Transport wrapping is documented but mTLS configuration
  is not. Add a dedicated section.

## Axis 4: Scale (architecture)

The flat-package design is correct for the present scope. The trigger for evolution
is clear.

### Trigger: resource count crosses ~6–8

When the SDK grows past profiles + balances + transactions + transfers + recipients

- quotes + webhooks + one more, the flat `client.ListX` surface becomes noisy.
  Move to a **service-client sub-structure**:

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
  v0.5.0+ will add API surface (write operations, quotes, recipients).
- **v1.0** — public API freeze. After v1.0, breaking changes require v2 and a
  deliberate migration path.

## Non-goals

- **Supporting non-Go languages** — out of scope; this is a Go SDK.
- **Re-implementing retries / circuit breakers** — `failsafe-go` does this well.
  Do not replace it without cause.
- **Auto-generation from OpenAPI** — Wise does not publish a complete OpenAPI spec.
  Hand-written types are correct; auto-gen would lose the two-layer boundary.
- **Caching / local state** — the SDK is stateless. Caching is the caller's job.
- **`Money` arithmetic** — `Money` pairs cents+currency to prevent mismatched amounts
  at the serialization boundary. It is deliberately not a financial math library
  (`Add`/`Sub`/`IsNegative`/`Equal` etc. are out of scope). Arithmetic is the
  consumer's domain logic; the SDK is an anti-corruption layer, not a domain model.
