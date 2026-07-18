# Roadmap

Long-term direction and raw ideas for wise-go. These are **not** actionable tasks
(actionable work lives in [TODO_LIST.md](TODO_LIST.md)) — they are the strategic
context that explains why those tasks exist and where the project is heading.

## Vision

wise-go aims to be the **canonical Go SDK for the Wise (TransferWise) API**: the
library that is obviously correct, obviously typed, and obviously safe to depend on.
Every monetary amount is `int64` cents. Every entity ID is branded. Every error is
typed and classifiable. Every API call retries intelligently.

Today the SDK covers the read side of three resources (profiles, balances,
transactions). The roadmap expands the surface along three axes — completeness,
type-safety, and scale — while preserving the architectural decisions that make the
codebase maintainable.

## Axis 1: Completeness (API surface)

Today: read-only, 3 resources. The Wise API is much larger.

### Near-term (post-v0.3.0)

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
  transactions. The current `HasMore` stub is forward-compat for this; the SDK can
  grow a `Page[Cursor]` type and `FetchMore` method without breaking the existing
  API.

### Out of vision

- **Streaming / WebSocket** — Wise does not expose a streaming API. Do not speculatively build one.
- **ORM-like abstractions** — the SDK is a thin anti-corruption layer, not a
  domain model. Resist abstractions over Wise's resources.

## Axis 2: Type-safety (data-model evolution)

The 2026-07-18 data-model review identified the path from "disciplined" to "invalid
states unrepresentable." The full redesign lives at
`docs/brainstorming/2026-07-18_data-model-review.html`.

### v0.3.0 — the breaking type release

- **`Money` value object** — collapse every paired `XxxCents int64` / `XxxCurrency string`
  into a single `Money` field. Transaction goes from 20 fields to 14. Mismatched
  currency/amount becomes unrepresentable.
- **`Currency` branded type** — `type Currency string` with a constructor that
  validates ISO 4217 (3-letter uppercase ASCII). The empty-currency case becomes
  unconstructible.
- **Typed `InvestmentState` enum** — join `ProfileType`/`BalanceType`/`TransactionType`
  in the typed-enum pattern.
- **Exported `DetailType` constants** — for `ListTransactionsRequest.Type` filter.
- **Enum casing normalization** — one rule for all SDK enums.
- **Reconcile `TransactionTypeUnknown`** — use it as the real fallback or remove it.

### v1.0 — the type-system lock-in

- **Move raw wire types to `internal/raw`** — `Profile`, `Balance`,
  `StatementTransaction`, etc. become invisible to consumers. Public surface
  shrinks to result types only. The `Result` suffix can then be dropped.
- **Remove `ListTransactionsResponse.HasMore`** — return `[]Transaction` directly.
  The field is always `false` today; v1.0 cleans up the lie.

### Beyond

- **Sealed transaction union (maybe)** — an interface-based union
  (`CreditTx`/`DebitTx`/`ExchangeTx`) was rejected in the data-model review as
  wrong ergonomics for statement iteration. Revisit only if a real consumer need
  emerges.
- **Generic `Page[T]`** — if pagination lands. Today there is no pagination, so no
  type to genericize.

## Axis 3: Scale (architecture)

The 2026-07-18 architecture review identified the current flat-package design as
correct for the present scope and named the trigger for evolution.

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
wrappers over the new service clients), then a v1.0 breaking change that removes
the wrappers. See `docs/architecture-understanding/2026-07-18_10-08_*.html` step 6.

### Trigger: a consumer asks to inject a narrow service interface

The real-world trigger. If a downstream user wants to mock only
`wise.Client.ListProfiles` in their tests, that is the signal to extract a
`ProfileService` with a narrow interface. Today, consumers should define
consumer-side interfaces themselves (Go proverb: "accept interfaces, return
structs"); document this in README in the meantime.

### Never

- **Per-resource Go modules** — `wise-profiles`, `wise-balances`, etc. Every
  consumer imports all three; the boundaries have no composability payoff.
  See `docs/modularization/2026-07-18_ASSESSMENT.html`.
- **Domain-core / infrastructure split** — one HTTP backend, one retry library.
  The seam would be unused.

## Release strategy

- **v0.x** — breaking changes accepted but coordinated. Each breaking release gets
  a `BREAKING:` line in CHANGELOG.md with a migration note. The v0.3.0 release
  (Money/Currency redesign) is the next planned break.
- **v1.0** — public API freeze. After v1.0, breaking changes require v2 and a
  deliberate migration path.

## Non-goals

- **Supporting non-Go languages** — out of scope; this is a Go SDK.
- **Re-implementing retries / circuit breakers** — `failsafe-go` does this well.
  Do not replace it without cause.
- **Auto-generation from OpenAPI** — Wise does not publish a complete OpenAPI spec.
  Hand-written types are correct; auto-gen would lose the two-layer boundary.
- **Caching / local state** — the SDK is stateless. Caching is the caller's job.
