# wise-go

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/wise-go.svg)](https://pkg.go.dev/github.com/larsartmann/wise-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/larsartmann/wise-go)](https://goreportcard.com/report/github.com/larsartmann/wise-go)

The unofficial Go SDK for the [Wise](https://wise.com) (TransferWise) API.

Wise does not provide an official Go SDK. This library fills that gap with strongly-typed structs, automatic retries, and idiomatic Go error handling.

## Features

- **Strongly-typed results** — Monetary amounts as `int64` cents (no `float64` money), dates as `time.Time`, enums for profile/balance/transaction types
- **Automatic retries** — Exponential backoff on 429 (rate limit), 5xx, and network errors via [failsafe-go](https://github.com/failsafe-go/failsafe-go)
- **Typed errors** — `AuthError`, `RateLimitError`, `NotFoundError`, `ServerError` with structured error details from the Wise API
- **Minimal dependencies** — `failsafe-go` for retries, `go-error-family` for behavioral error classification
- **Sandbox support** — One-line switch to the Wise sandbox environment

## Installation

```bash
go get github.com/larsartmann/wise-go
```

Requires Go 1.26 or later.

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
            fmt.Printf("  %s %s: %d cents\n", b.Currency, b.Name, b.AmountCents)
        }
    }

    // List transactions for a balance within a date range
    resp, err := client.ListTransactions(ctx, wise.ListTransactionsRequest{
        ProfileID: profiles[0].ID,
        BalanceID: balances[0].ID,
        Currency:  "EUR",
        From:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
        To:        time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, tx := range resp.Transactions {
        sign := "+"
        if tx.TotalCents < 0 {
            sign = ""
        }

        fmt.Printf("%s%d cents — %s (%s)\n", sign, tx.TotalCents, tx.Description, tx.Type)
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
// []ProfileResult{
//   {ID: 12345, Type: ProfileTypePersonal, Name: "John Doe", Email: "john@example.com", CreatedAt: ...},
//   {ID: 67890, Type: ProfileTypeBusiness, Name: "Acme Corp", Email: "billing@acme.com", CreatedAt: ...},
// }
```

### Balances

```go
// List visible, non-investment balances for a profile
balances, err := client.ListBalances(ctx, profileID)
// []BalanceResult{
//   {ID: 100, Currency: "EUR", Type: BalanceTypeStandard, Name: "Main Account", AmountCents: 123456, ...},
// }

// Get a specific balance by ID
balance, err := client.GetBalance(ctx, profileID, balanceID)
```

`ListBalances` filters to `Visible: true` and `InvestmentState == "NOT_INVESTED"` only. `GetBalance` delegates to `ListBalances` + linear scan (Wise has no single-balance endpoint).

### Transactions

```go
resp, err := client.ListTransactions(ctx, wise.ListTransactionsRequest{
    ProfileID: 12345,
    BalanceID: 100,
    Currency:  "EUR",
    From:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
    To:        time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
    Type:      "CARD_PAYMENT", // optional filter
})
// resp.Transactions → []Transaction
// resp.HasMore → always false (Wise returns all in one response)
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

- `AmountCents` — absolute value (always positive)
- `TotalCents` — signed value (negative for debits, positive for credits)
- All amounts use `int64` minor units (cents) to avoid IEEE 754 floating-point errors

## Error Handling

The SDK returns typed errors you can match with `errors.As`:

```go
import "errors"

balances, err := client.ListBalances(ctx, profileID)
if err != nil {
    var rateLimit *wise.RateLimitError
    var auth *wise.AuthError
    var notFound *wise.NotFoundError
    var server *wise.ServerError

    switch {
    case errors.As(err, &rateLimit):
        fmt.Printf("rate limited, retry after %s\n", rateLimit.RetryAfter)
    case errors.As(err, &auth):
        fmt.Printf("auth failed: %s\n", auth.Message)
    case errors.As(err, &notFound):
        fmt.Printf("not found: %s\n", notFound.Message)
    case errors.As(err, &server):
        fmt.Printf("server error (%d): %s\n", server.StatusCode, server.Message)
    default:
        var apiErr *wise.APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("API error (%d): %s\n", apiErr.StatusCode, apiErr.Message)
        }
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

| Decision                          | Rationale                                                                                                      |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| Monetary amounts as `int64` cents | `float64` causes precision loss (e.g., `0.1 + 0.2 ≠ 0.3`). Cents are safe for arithmetic and storage.          |
| Two-layer type system             | Raw API types mirror JSON exactly. Result types expose clean Go types. Mapping functions convert between them. |
| `failsafe-go` for retries         | Purpose-built HTTP retry with backoff, not a generic CQRS middleware.                                          |
| Flat package structure            | Single `package wise` — no sub-packages for 8 files. Import path is the API.                                   |
| BDD tests with Ginkgo             | `httptest.Server` mock API responses. Tests verify both happy paths and error classification.                  |

## Testing

```bash
go test ./...
```

Tests use `net/http/httptest` to mock the Wise API — no network access required.

## Project Status

Early development. Covers core read endpoints: profiles, balances, and transactions.

Not yet implemented: transfers, recipients, quotes, webhooks, and write operations.

## License

Proprietary — see [LICENSE](LICENSE).
