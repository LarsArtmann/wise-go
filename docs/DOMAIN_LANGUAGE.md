# Domain Language

A **Unified Language** for `wise-go` — shared across Customer, Product Owner, Developer, and AI.
Inspired by Domain-Driven Design (DDD) Ubiquitous Language.

## Glossary

| Term           | Definition                                                                                                        | Context          |
| -------------- | ----------------------------------------------------------------------------------------------------------------- | ---------------- |
| wise-go        | The Go SDK for the Wise API                                                                                       | Project name     |
| Wise           | The financial platform (formerly TransferWise)                                                                    | External service |
| API Key        | Bearer token for authenticating with the Wise API                                                                 | Authentication   |
| Sandbox        | Wise test environment at `api.wise-sandbox.com` (V2; V1 `api.sandbox.transferwise.tech` deprecated June 30, 2026) | Development      |
| 2026Q3 surface | The quarterly versioned API base (`https://api.wise.com/2026Q3`) that hosts the webhook subscription endpoints    | Webhooks         |

## Entities

Objects with identity and lifecycle in the Wise domain.

| Term        | Definition                                         | Context                                            |
| ----------- | -------------------------------------------------- | -------------------------------------------------- |
| Profile     | A personal or business account on Wise             | Has one or more balances; identified by `int64` ID |
| Balance     | A currency-denominated account holding funds       | Belongs to a profile; has amount in cents          |
| Transaction | A movement of money (credit, debit, exchange, fee) | Belongs to a balance; has signed total in cents    |

## Value Objects

Immutable objects defined by attributes.

| Term                | Definition                                   | Context                                                                                                                              |
| ------------------- | -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| Cents               | Monetary amount in minor units (int64)       | `1234.56 EUR` → `123456` cents; avoids float64 precision loss                                                                        |
| BalanceAmount       | Wise API monetary value with currency        | Wire format: `{value: 1234.56, currency: "EUR"}`; converted to cents via `Cents()`                                                   |
| TransactionType     | Classification of a transaction              | Enum: card, credit, debit, exchange, fee, refund, transfer, payment (`unknown` removed in v0.4.0 — never returned by the classifier) |
| ProfileType         | Kind of profile                              | Enum: personal, business                                                                                                             |
| BalanceType         | Kind of balance                              | Enum: standard, savings (lowercase public values; wire uses UPPERCASE)                                                               |
| InvestmentState     | Whether a balance is invested in Wise        | Values: `NOT_INVESTED`, `INVESTED`; only non-invested balances are returned                                                          |
| TransactionExchange | Currency-conversion details on a transaction | From/to amounts in cents, rate; nil for non-conversion transactions                                                                  |
| Money               | Paired cents + currency value object         | `Money{Cents int64, Currency Currency}`; no arithmetic by design (see AGENTS.md)                                                     |
| WebhookEventType    | Which domain event triggers a delivery       | Open enum (33 documented constants); open by design so unseen event types never break handlers                                       |

## Entities (extended)

| Term         | Definition                                                                             | Context                                                                              |
| ------------ | -------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Transfer     | An outgoing money movement between a quote and a recipient                             | Has lifecycle status (`TransferStatus` open enum)                                    |
| Quote        | A locked exchange-rate offer that can back a transfer                                  | UUID-string ID (`QuoteID`), unlike int64 entity IDs                                  |
| Recipient    | A payout destination (bank account, etc.)                                              | `details` is polymorphic per currency/corridor                                       |
| Subscription | A registration telling Wise to POST events for a profile to an HTTPS URL               | UUID-string ID (`WebhookSubscriptionID`); profile-level CRUD on the 2026Q3 surface   |
| Delivery     | A single webhook POST from Wise to the subscription's URL                              | Signed (`X-Signature-SHA256`); dedup on `X-Delivery-Id` (unique per attempt)         |
| Envelope     | The JSON wrapper of every delivery (`data`, `event_type`, `schema_version`, `sent_at`) | Parsed by `ParseWebhookEvent`; unknown event types pass through with the payload raw |
| SCA          | Strong Customer Authentication (3-D Secure-like challenge)                             | HTTP 403 with empty body; one-time token in response headers                         |

## Raw vs Result Types

The SDK uses a two-layer type system:

| Layer        | Purpose                                 | Example                                                     |
| ------------ | --------------------------------------- | ----------------------------------------------------------- |
| Raw types    | Mirror Wise JSON wire format exactly    | `raw.Profile` with `CreatedAt string` and `Type string`     |
| Result types | Strongly-typed public API for consumers | `Profile` with `CreatedAt time.Time` and `Type ProfileType` |

Mapping functions (`mapProfile`, `mapBalance`, `mapTransaction`, `toWebhookSubscription`, `toWebhookResource`) convert between layers.

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
