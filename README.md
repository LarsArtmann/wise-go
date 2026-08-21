# wise-go

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/wise-go.svg)](https://pkg.go.dev/github.com/larsartmann/wise-go)
[![Coverage](https://img.shields.io/badge/coverage-84.2%25-yellowgreen)](https://github.com/larsartmann/wise-go)
[![Lint](https://img.shields.io/badge/lint-0%20issues-success)](https://github.com/larsartmann/wise-go)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)

The unofficial Go SDK for the [Wise](https://wise.com) (TransferWise) API.

Wise publishes no official Go SDK. An OpenAPI spec exists, but it reflects Wise's wire types directly — `float64` for money, untyped string IDs, inconsistent date formats. **wise-go fills that gap** with hand-written types that make invalid states hard to reach: monetary amounts as `int64` cents (never `float64`), branded IDs that prevent mixing `ProfileID` with `BalanceID` at compile time, and behavioral error classification so you can retry on intent rather than string-matching status codes.

> **Status: active development (v0.8.1).** The core transfer flow is implemented
> end-to-end: profiles, balances, transactions, exchange rates, quotes (with
> `paymentOptions` + fees), recipients, transfers (create / get / list /
> cancel), delivery estimates, and transfer-requirements validation. See
> [FEATURES.md](FEATURES.md) for the honest inventory and [ROADMAP.md](ROADMAP.md)
> for what's next.

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
  - [Quotes](#quotes)
  - [Recipients](#recipients)
  - [Exchange rates](#exchange-rates)
  - [Transfers](#transfers)
  - [Core transfer flow](#core-transfer-flow-quote--recipient--transfer)
- [Mocking the Client](#mocking-the-client)
- [Request Middleware](#request-middleware)
- [Error Handling](#error-handling)
- [Strong Customer Authentication (SCA)](#strong-customer-authentication-sca)
- [Design Decisions](#design-decisions)
- [Testing](#testing)
- [Project Status](#project-status)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Money is never `float64`** — Every amount is `int64` minor units (cents). No IEEE-754 representation error, ever.
- **Branded IDs prevent entity-mixing bugs** — `ProfileID`, `BalanceID`, `TransactionID`, `TransferID`, `RecipientID`, and `QuoteID` are distinct types; passing one where another belongs is a compile error.
- **Automatic retries with backoff** — Exponential backoff on 429 (rate limit), 5xx, and network errors via [failsafe-go](https://github.com/failsafe-go/failsafe-go). Auth, not-found, and client errors fail immediately.
- **Typed, classifiable errors** — `AuthError`, `RateLimitError` (with parsed `Retry-After`), `NotFoundError`, `ServerError`. Each carries its Wise API detail and implements `ErrorCode()` / `ErrorFamily()` / `IsRetryable()` from [go-error-family](https://github.com/larsartmann/go-error-family).
- **Two-layer type system** — Raw wire types live in `internal/raw`; result types expose clean Go with `Money` value objects and branded `Currency`. The mapping is the only bridge.
- **Write operations** — create quotes (authenticated and unauthenticated), recipients, and transfers; cancel transfers; validate transfer requirements; fetch delivery estimates.
- **SCA challenge support** — SCA-protected endpoints (e.g. balance statements for UK/EEA profiles) surface as `*SCAChallengeError` with the one-time approval token; complete the challenge with `WithSCAApprovalToken` and retry. See [SCA](#strong-customer-authentication-sca).
- **Tolerant timestamp handling** — Wise emits four different timestamp formats. One parser accepts them all (zoneless = UTC), and outgoing query timestamps are normalized to UTC `Z` (Wise rejects zone offsets with 422).
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

### Quotes

```go
// Authenticated quote locks the rate for 30 minutes and can back a transfer.
// preferredPayIn=BALANCE keeps fees consistent when funding from a balance.
quote, err := client.CreateQuote(ctx, profileID, wise.CreateQuoteRequest{
    SourceCurrency: wise.Currency("EUR"),
    TargetCurrency: wise.Currency("USD"),
    SourceAmount:   &wise.Money{Cents: 100_000, Currency: wise.Currency("EUR")},
    PreferredPayIn: wise.PayInBalance,
    PayOut:         wise.PayOutBankTransfer,
})
// quote.PaymentOptions → per pay-in/pay-out fee breakdown + estimated delivery
// quote.Notices → messages to show the user (BLOCKED means: don't use it)

// Fetch an existing quote
quote, err = client.GetQuote(ctx, profileID, quote.ID)

// Illustrative quote without a user token (no ID, cannot back a transfer)
quote, err = client.CreateUnauthenticatedQuote(ctx, req)
```

### Recipients

```go
recipient, err := client.CreateRecipient(ctx, wise.CreateRecipientRequest{
    ProfileID:         profileID,
    Currency:          wise.Currency("GBP"),
    Type:              "sort_code",
    AccountHolderName: "Jane Doe",
    Details: map[string]string{
        "sortCode":      "040075",
        "accountNumber": "37778842",
    },
})

recipient, err = client.GetRecipient(ctx, recipient.ID)
recipients, err := client.ListRecipients(ctx, wise.ListRecipientsRequest{
    ProfileID: profileID,
})
```

`Recipient.Details` is a `map[string]string`; required keys vary by currency and
route — use the account-requirements endpoint to discover them per corridor.

### Exchange rates

```go
// Current rate
rate, err := client.GetExchangeRate(ctx, wise.Currency("EUR"), wise.Currency("USD"), time.Time{})
// Historical rate (pass a time.Time; UTC recommended)
rate, err = client.GetExchangeRate(ctx, wise.Currency("EUR"), wise.Currency("USD"), someTime)
```

### Transfers

```go
transfers, err := client.ListTransfers(ctx, wise.ListTransfersRequest{
    ProfileID: wise.NewProfileID(12345),                          // optional, defaults to personal profile
    From:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),       // optional createdDateStart
    To:        time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),  // optional createdDateEnd
    Status:    []wise.TransferStatus{wise.TransferStatusDelivered}, // optional status filter
})
// transfers → []wise.Transfer (auto-paginated, 100 per page)
```

Unlike balance statements, `GET /v1/transfers` is **not SCA-protected** and is available to personal API tokens in all regions — it is the reliable source for outgoing transfer history (amount, fees via rate, status, recipient, reference, timestamps). It does not include deposits, card payments, conversions, or interest; those remain statement-only.

`TransferStatus` is an open string enum: documented lifecycle values are provided as constants (`TransferStatusDelivered`, `TransferStatusCancelled`, …), and unknown values from Wise pass through unchanged.

`Transfer.Created` is parsed tolerantly (RFC3339 or Wise's space-separated format; zoneless values are UTC).

### Core transfer flow: quote → recipient → transfer

The 1% Pareto core. A consumer with `GetProfile`, `CreateQuote`, `CreateRecipient`,
and `CreateTransfer` can move money end-to-end:

```go
ctx := context.Background()

// 1. Quote — locks the rate for 30 minutes.
//    PreferredPayIn=BALANCE keeps quote fees consistent with transfer fees when
//    funding from a multi-currency balance.
quote, err := client.CreateQuote(ctx, profile.ID, wise.CreateQuoteRequest{
    SourceCurrency: wise.Currency("EUR"),
    TargetCurrency: wise.Currency("USD"),
    SourceAmount:   &wise.Money{Cents: 100_000, Currency: wise.Currency("EUR")},
    PreferredPayIn: wise.PayInBalance,
    PayOut:         wise.PayOutBankTransfer,
})
if err != nil {
    log.Fatal(err)
}

// A BLOCKED notice means this quote must not back a transfer.
for _, n := range quote.Notices {
    if n.Type == wise.QuoteNoticeTypeBlocked {
        log.Fatalf("quote blocked: %s", n.Text)
    }
}

// 2. Recipient — a bank account to send to. Required details vary by currency
//    and route; the account-requirements endpoint discovers them per corridor.
recipient, err := client.CreateRecipient(ctx, wise.CreateRecipientRequest{
    ProfileID:         profile.ID,
    Currency:          wise.Currency("GBP"),
    Type:              "sort_code",
    AccountHolderName: "Jane Doe",
    Details: map[string]string{
        "sortCode":      "040075",
        "accountNumber": "37778842",
    },
})
if err != nil {
    log.Fatal(err)
}

// 2b. Optional: discover required details before creating the transfer.
//     Fields flagged RefreshRequirementsOnChange mean re-calling this once the
//     field is populated reveals lower-level required fields.
requirements, err := client.ValidateTransferRequirements(ctx, wise.ValidateTransferRequirementsRequest{
    TargetAccount: recipient.ID,
    QuoteID:       quote.ID,
})
if err != nil {
    log.Fatal(err)
}
_ = requirements // inspect .Fields: required, allowed values, validation regexps

// 3. Transfer — customerTransactionId is your idempotency key (a UUID).
//    Reusing it with the same quote + targetAccount returns the existing
//    transfer instead of creating a duplicate.
transfer, err := client.CreateTransfer(ctx, wise.CreateTransferRequest{
    QuoteID:               quote.ID,
    TargetAccount:         recipient.ID,
    CustomerTransactionID: "22244c35-9fe8-4c32-b7fd-d05c2a7734bf",
})
if err != nil {
    log.Fatal(err)
}

// 4. Track — poll GetTransfer until it leaves waiting_for_funds, then
//    fund the transfer from a balance and poll until delivered.
estimate, err := client.GetDeliveryEstimate(ctx, transfer.ID, "Europe/Berlin")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("expected arrival: %s\n", estimate.EstimatedDeliveryDate)

// 5. Cancel — only possible before the transfer is processed.
cancelled, err := client.CancelTransfer(ctx, transfer.ID)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("cancelled: %s\n", cancelled.Status)
```

`Quote.PaymentOptions` lists every pay-in/pay-out combination with its fee
breakdown (`Fee.Total` is the value to display) and estimated delivery; a
`QuoteNotice` of type `BLOCKED` means the quote must not be used to create a
transfer.

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
        fmt.Printf("rate limited (scope: %s), retry after %s\n", rl.RateLimitedBy, rl.RetryAfter)
    } else if sca, ok := errors.AsType[*wise.SCAChallengeError](err); ok {
        fmt.Printf("SCA approval required, token: %s\n", sca.TwoFAApprovalToken())
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

| HTTP Status            | Error Type          | Retried? |
| ---------------------- | ------------------- | -------- |
| 401, 403               | `AuthError`         | No       |
| 403 + Wise 2FA headers | `SCAChallengeError` | No       |
| 404                    | `NotFoundError`     | No       |
| 429                    | `RateLimitError`    | Yes      |
| 5xx                    | `ServerError`       | Yes      |
| Network error          | Wrapped `error`     | Yes      |

Mapper failures (unparseable timestamps, invalid currency codes in a response)
carry `Corruption` classification — permanent, not retryable — so a blanket
"retry on transient" wrapper fails fast instead of looping.

## Strong Customer Authentication (SCA)

Some Wise endpoints (balance statements for UK/EEA profiles among them) are
SCA-protected: the API answers **HTTP 403 with an empty body**. The verdict and
the one-time token live in the `x-2fa-approval-result` / `x-2fa-approval`
response headers. The SDK detects this and returns `*SCAChallengeError` instead
of a bare `AuthError`:

```go
resp, err := client.ListTransactions(ctx, req)
if err != nil {
    if sca, ok := errors.AsType[*wise.SCAChallengeError](err); ok {
        // 1. Send the one-time token to the user (push notification, email, ...).
        token := sca.TwoFAApprovalToken()

        // 2. After the user approves the challenge in the Wise app, replay it:
        scaClient := wise.New("api-key", wise.WithSCAApprovalToken(token))
        resp, err = scaClient.ListTransactions(ctx, req)
    }
}
```

`GET /v1/transfers` (and the other transfer endpoints) are **not** SCA-protected,
which is why `ListTransfers` is the reliable source for outgoing transfer
history with personal API tokens.

## Design Decisions

| Decision                          | Rationale                                                                                                                                              |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Monetary amounts as `int64` cents | `float64` causes precision loss (e.g., `0.1 + 0.2 ≠ 0.3`). Cents are safe for arithmetic and storage.                                                  |
| Two-layer type system             | Raw wire types in `internal/raw` mirror JSON exactly. Result types expose clean Go with `Money` value objects. Mapping functions convert between them. |
| `failsafe-go` for retries         | Purpose-built HTTP retry with backoff, not a generic CQRS middleware.                                                                                  |
| Flat package structure            | Single `package wise`; wire types hidden in `internal/raw`. The import path is the API.                                                                |
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
