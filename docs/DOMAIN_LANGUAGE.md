# Domain Language

A **Unified Language** for `wise-go` — shared across Customer, Product Owner, Developer, and AI.
Inspired by Domain-Driven Design (DDD) Ubiquitous Language.

## Glossary

| Term    | Definition                                               | Context          |
| ------- | -------------------------------------------------------- | ---------------- |
| wise-go | The Go SDK for the Wise API                              | Project name     |
| Wise    | The financial platform (formerly TransferWise)           | External service |
| API Key | Bearer token for authenticating with the Wise API        | Authentication   |
| Sandbox | Wise test environment at `api.sandbox.transferwise.tech` | Development      |

## Entities

Objects with identity and lifecycle in the Wise domain.

| Term        | Definition                                         | Context                                            |
| ----------- | -------------------------------------------------- | -------------------------------------------------- |
| Profile     | A personal or business account on Wise             | Has one or more balances; identified by `int64` ID |
| Balance     | A currency-denominated account holding funds       | Belongs to a profile; has amount in cents          |
| Transaction | A movement of money (credit, debit, exchange, fee) | Belongs to a balance; has signed total in cents    |

## Value Objects

Immutable objects defined by attributes.

| Term            | Definition                             | Context                                                                            |
| --------------- | -------------------------------------- | ---------------------------------------------------------------------------------- |
| Cents           | Monetary amount in minor units (int64) | `1234.56 EUR` → `123456` cents; avoids float64 precision loss                      |
| BalanceAmount   | Wise API monetary value with currency  | Wire format: `{value: 1234.56, currency: "EUR"}`; converted to cents via `Cents()` |
| TransactionType | Classification of a transaction        | Enum: card, credit, debit, exchange, fee, refund, transfer, payment, unknown       |
| ProfileType     | Kind of profile                        | Enum: personal, business                                                           |
| BalanceType     | Kind of balance                        | Enum: STANDARD, SAVINGS                                                            |

## Raw vs Result Types

The SDK uses a two-layer type system:

| Layer        | Purpose                                 | Example                                                           |
| ------------ | --------------------------------------- | ----------------------------------------------------------------- |
| Raw types    | Mirror Wise JSON wire format exactly    | `Profile` with `CreatedAt string` and `Type string`               |
| Result types | Strongly-typed public API for consumers | `ProfileResult` with `CreatedAt time.Time` and `Type ProfileType` |

Mapping functions (`mapProfile`, `mapBalance`, `mapTransaction`) convert between layers.

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
