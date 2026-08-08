# wise-go

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/wise-go.svg)](https://pkg.go.dev/github.com/larsartmann/wise-go)
[![Coverage](https://img.shields.io/badge/coverage-94.8%25-success)](https://github.com/larsartmann/wise-go)
[![Lint](https://img.shields.io/badge/lint-0%20issues-success)](https://github.com/larsartmann/wise-go)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)

The unofficial Go SDK for the [Wise](https://wise.com) (TransferWise) API.

Wise publishes no official Go SDK. An OpenAPI spec exists, but it reflects Wise's wire types directly — `float64` for money, untyped string IDs, inconsistent date formats. **wise-go fills that gap** with hand-written types that make invalid states hard to reach: monetary amounts as `int64` cents (never `float64`), branded IDs that prevent mixing `ProfileID` with `BalanceID` at compile time, and behavioral error classification so you can retry on intent rather than string-matching status codes.

> **Status: early development (v0.5.0).** Read-only coverage of profiles, balances, and transactions. Write operations (transfers, recipients, quotes, webhooks) are not yet implemented — see [ROADMAP.md](ROADMAP.md).

> **Design story:** [I needed a Go SDK for Wise. Nobody built one.](https://larsartmann.com/blog/when-the-api-has-no-spec-your-types-are-the-spec)

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [API Reference](#api-reference)
  - [Authentication](#authentication)
  - [Profiles](#profiles)
  - [Balances](#balances)
  - [Transactions](#transactions)
- [Mocking the Client](#mocking-the-client)
- [Request Middleware](#request-middleware)
- [Error Handling](#error-handling)
- [Design Decisions](#design-decisions)
- [Testing](#testing)
- [Project Status](#project-status)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Money is never `float64`** — Every amount is `int64` minor units (cents). No IEEE-754 representation error, ever.
- **Branded IDs prevent entity-mixing bugs** — `ProfileID`, `BalanceID`, and `TransactionID` are distinct types; passing one where another belongs is a compile error.
- **Automatic retries with backoff** — Exponential backoff on 429 (rate limit), 5xx, and network errors via [failsafe-go](https://github.com/failsafe-go/failsafe-go). Auth, not-found, and client errors fail immediately.
- **Typed, classifiable errors** — `AuthError`, `RateLimitError` (with parsed `Retry-After`), `NotFoundError`, `ServerError`. Each carries its Wise API detail and implements `ErrorCode()` / `ErrorFamily()` / `IsRetryable()` from [go-error-family](https://github.com/larsartmann/go-error-family).
- **Two-layer type system** — Raw wire types live in `internal/raw`; result types expose clean Go with `Money` value objects and branded `Currency`. The mapping is the only bridge.
- **Sandbox support** — One-line switch to the Wise sandbox environment.
- **Minimal dependencies** — Three focused production deps: `failsafe-go`, `go-branded-id`, `go-error-family`.

## Installation

```bash
go get github.com/larsartmann/wise-go
```

Requires **Go 1.26 or later** with the **`jsonv2` experiment** enabled. The `go-branded-id` and `go-error-family` dependencies use `encoding/json/v2`, which only builds when the experiment is on:

```bash
export GOEXPERIMENT=jsonv2   # add to ~/.bashrc / ~/.zshrc for persistence
```

Every invocation of the Go toolchain (`go build`, `go test`, `go vet`, `go mod tidy`) needs this variable set. On NixOS, `nix develop` sets it for you (see `flake.nix`).

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/larsartmann/wise-go"
)

func main() {
    client := wise.New("your-api-key")

    ctx := context.Background()

    // Validate your API key
    if err := client.Authenticate(ctx); err != nil {
        log.Fatal(err)
    }

    // List profiles
    profiles, err := client.ListProfiles(ctx)
    if err != nil {
        log.Fatal(err)
    }

    for _, p := range profiles {
        fmt.Printf("%s (%s)\n", p.Name, p.Type)

        // List balances for each profile
        balances, err := client.ListBalances(ctx, p.ID)
        if err != nil {
            log.Fatal(err)
        }

        for _, b := range balances {
            fmt.Printf("  %s %s: %d cents\n", b.Currency, b.Name, b.Amount.Cents)
        }
    }

    // List transactions for a balance within a date range
    resp, err := client.ListTransactions(ctx, wise.ListTransactionsRequest{
        ProfileID: profiles[0].ID,
        BalanceID: balances[0].ID,
        Currency:  wise.Currency("EUR"),
        From:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
        To:        time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, tx := range resp.Transactions {
        sign := "+"
        if tx.Total.Cents < 0 {
            sign = ""
        }

        fmt.Printf("%s%d cents — %s (%s)\n", sign, tx.Total.Cents, tx.Description, tx.Type)
    }
}
```

## Configuration

```go
// Sandbox environment
client := wise.New("sandbox-key", wise.WithSandbox())

// Custom base URL
client := wise.New("key", wise.WithBaseURL("https://custom.proxy.example.com"))

// Custom timeout (default: 30s)
client := wise.New("key", wise.WithTimeout(10*time.Second))

// Custom retry policy (default: 3 retries, 100ms–5s backoff)
client := wise.New("key", wise.WithRetry(5, time.Second, 30*time.Second))

// Custom HTTP client
client := wise.New("key", wise.WithHTTPClient(&http.Client{Timeout: 15*time.Second}))

// Compose multiple options
client := wise.New("key",
    wise.WithSandbox(),
    wise.WithTimeout(15*time.Second),
    wise.WithRetry(5, time.Second, 30*time.Second),
)
```

## API Reference

### Authentication

```go
// Validate API key (calls ListProfiles internally)
err := client.Authenticate(ctx)

// Health check (delegates to Authenticate)
err := client.Health(ctx)
```

### Profiles

```go
profiles, err := client.ListProfiles(ctx)
// []Profile{
//   {ID: 12345, Type: ProfileTypePersonal, Name: "John Doe", Email: "john@example.com", CreatedAt: ...},
//   {ID: 67890, Type: ProfileTypeBusiness, Name: "Acme Corp", Email: "billing@acme.com", CreatedAt: ...},
// }
```

### Balances

```go
// List visible, non-investment balances for a profile
balances, err := client.ListBalances(ctx, profileID)
// []Balance{
//   {ID: 100, Currency: "EUR", Type: BalanceTypeStandard, Name: "Main Account", Amount.Cents: 123456, ...},
// }

// Get a specific balance by ID
balance, err := client.GetBalance(ctx, profileID, balanceID)
```

`ListBalances` filters to `Visible: true` and `InvestmentState == "NOT_INVESTED"` only. `GetBalance` delegates to `ListBalances` + linear scan (Wise has no single-balance endpoint).

### Transactions

```go
resp, err := client.ListTransactions(ctx, wise.ListTransactionsRequest{
    ProfileID: wise.NewProfileID(12345),
    BalanceID: wise.NewBalanceID(100),
    Currency:  wise.Currency("EUR"),
    From:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
    To:        time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
    Type:      wise.DetailTypeCardPayment, // optional filter
})
// resp.Transactions → []Transaction
```

**Transaction type classification** — The SDK maps Wise detail types to Go enums:

| Wise `details.type`        | `TransactionType`         | Note                         |
| -------------------------- | ------------------------- | ---------------------------- |
| `CARD_PAYMENT`             | `TransactionTypeCard`     |                              |
| `CARD_REFUND` (amount > 0) | `TransactionTypeRefund`   |                              |
| `CARD_REFUND` (amount ≤ 0) | `TransactionTypeCard`     |                              |
| `TRANSFER`                 | `TransactionTypeTransfer` |                              |
| `PAYMENT`                  | `TransactionTypePayment`  |                              |
| `CONVERSION`, `EXCHANGE`   | `TransactionTypeExchange` |                              |
| `FEE`                      | `TransactionTypeFee`      |                              |
| Other (amount > 0)         | `TransactionTypeCredit`   | Default for positive amounts |
| Other (amount ≤ 0)         | `TransactionTypeDebit`    | Default for negative amounts |

**Amount semantics:**

- `Amount.Cents` — absolute value (always positive)
- `Total.Cents` — signed value (negative for debits, positive for credits)
- `RunningBalance.Cents` — the balance after this transaction (signed `int64`)
- `Exchange` — `*TransactionExchange` with from/to amounts and rate; `nil` for non-conversion transactions
- All amounts use `int64` minor units (cents) to avoid IEEE 754 floating-point errors

**Date timezone:** `Transaction.Date` is in UTC. Wise statement dates carry no timezone; the SDK interprets them as UTC via `time.Parse`. Convert explicitly before comparing against local-time values to avoid off-by-one-day errors at boundaries.

**Validation** — `ListTransactions` rejects invalid requests before hitting the API:

- Empty `Currency` → `"wise.transactions.invalid_request: currency is required"`
- `From` after `To` → `"wise.transactions.invalid_request: intervalStart must not be after intervalEnd"`

## Mocking the Client

The SDK returns concrete types (`*wise.Client`, `[]wise.Profile`, etc.), not interfaces.
This follows Go's "accept interfaces, return structs" proverb: consumers define narrow
interfaces for the subset of methods they actually use, keeping mocks minimal.

```go
// Define a narrow interface in your package.
type ProfileLister interface {
    ListProfiles(ctx context.Context) ([]wise.Profile, error)
}

// Your service depends on the interface, not *wise.Client.
type Service struct {
    profiles ProfileLister
}

// In tests, implement the interface with a stub.
type mockProfileLister struct {
    profiles []wise.Profile
    err      error
}

func (m *mockProfileLister) ListProfiles(ctx context.Context) ([]wise.Profile, error) {
    return m.profiles, m.err
}
```

## Request Middleware

`WithHTTPClient` accepts any type implementing the `Doer` interface
(`Do(req *http.Request) (*http.Response, error)`). `*http.Client` satisfies this
implicitly. Inject a custom client to add tracing, logging, or mTLS at the transport layer:

```go
type loggingTransport struct {
    next http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    start := time.Now()
    resp, err := t.next.RoundTrip(req)
    log.Printf("%s %s -> %d (%s)", req.Method, req.URL, resp.StatusCode, time.Since(start))
    return resp, err
}
```

Then inject it via `WithHTTPClient`:

```go
client := wise.New("key", wise.WithHTTPClient(&http.Client{
    Transport: &loggingTransport{next: http.DefaultTransport},
}))
```

## Error Handling

The SDK returns typed errors you can match with `errors.AsType` (Go 1.26+):

```go
import "errors"

balances, err := client.ListBalances(ctx, profileID)
if err != nil {
    if rl, ok := errors.AsType[*wise.RateLimitError](err); ok {
        fmt.Printf("rate limited, retry after %s\n", rl.RetryAfter)
    } else if auth, ok := errors.AsType[*wise.AuthError](err); ok {
        fmt.Printf("auth failed: %s\n", auth.Message)
    } else if nf, ok := errors.AsType[*wise.NotFoundError](err); ok {
        fmt.Printf("not found: %s\n", nf.Message)
    } else if srv, ok := errors.AsType[*wise.ServerError](err); ok {
        fmt.Printf("server error (%d): %s\n", srv.StatusCode, srv.Message)
    } else if apiErr, ok := errors.AsType[*wise.APIError](err); ok {
        fmt.Printf("API error (%d): %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

| HTTP Status   | Error Type       | Retried? |
| ------------- | ---------------- | -------- |
| 401, 403      | `AuthError`      | No       |
| 404           | `NotFoundError`  | No       |
| 429           | `RateLimitError` | Yes      |
| 5xx           | `ServerError`    | Yes      |
| Network error | Wrapped `error`  | Yes      |

## Design Decisions

| Decision                          | Rationale                                                                                                                                              |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Monetary amounts as `int64` cents | `float64` causes precision loss (e.g., `0.1 + 0.2 ≠ 0.3`). Cents are safe for arithmetic and storage.                                                  |
| Two-layer type system             | Raw wire types in `internal/raw` mirror JSON exactly. Result types expose clean Go with `Money` value objects. Mapping functions convert between them. |
| `failsafe-go` for retries         | Purpose-built HTTP retry with backoff, not a generic CQRS middleware.                                                                                  |
| Flat package structure            | Single `package wise` — no sub-packages for 8 files. Import path is the API.                                                                           |
| BDD tests with Ginkgo             | `httptest.Server` mock API responses. Tests verify both happy paths and error classification.                                                          |

## Testing

```bash
GOEXPERIMENT=jsonv2 go test ./...
```

Tests use `net/http/httptest` to mock the Wise API — no network access required. The `GOEXPERIMENT=jsonv2` prefix is mandatory (see [Installation](#installation)).

## Project Status

See [FEATURES.md](FEATURES.md) for the full feature inventory by status and [ROADMAP.md](ROADMAP.md) for long-term direction.

## Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on development setup, code style, and submitting pull requests.

## License

Apache-2.0 — see [LICENSE](LICENSE).

## Trademarks

"Wise" and "TransferWise" are registered trademarks of Wise Payments Limited. This project is not affiliated with, endorsed by, or sponsored by Wise Payments Limited. All trademarks are the property of their respective owners.
