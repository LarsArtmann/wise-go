# wise-go

[![Go Reference](https://pkg.go.dev/badge/github.com/larsartmann/wise-go.svg)](https://pkg.go.dev/github.com/larsartmann/wise-go)
[![Coverage](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fraw.githubusercontent.com%2Flarsartmann%2Fwise-go%2Fmaster%2F.github%2Fbadges%2Fcoverage.json&query=%24.message&label=coverage&color=%24.color)](https://github.com/larsartmann/wise-go/actions)
[![Lint](https://img.shields.io/badge/lint-0%20issues-success)](https://github.com/larsartmann/wise-go)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)

<!-- Coverage badge reads .github/badges/coverage.json, committed by the
     coverage-badge CI job. Measurement basis: total statement coverage over
     the whole module (GOEXPERIMENT=jsonv2 go test -coverprofile
     -covermode=atomic ./... | go tool cover -func). Never edit by hand. -->

The unofficial Go SDK for the [Wise](https://wise.com) (TransferWise) API.

Wise publishes no official Go SDK. An OpenAPI spec exists, but it reflects Wise's wire types directly — `float64` for money, untyped string IDs, inconsistent date formats. **wise-go fills that gap** with hand-written types that make invalid states hard to reach: monetary amounts as `int64` cents (never `float64`), branded IDs that prevent mixing `ProfileID` with `BalanceID` at compile time, and behavioral error classification so you can retry on intent rather than string-matching status codes.

> **Status: active development (v0.15.0).** The core transfer flow is implemented
> end-to-end — profiles, users, balances, transactions and statement file downloads
> (CSV/PDF/XLSX/CAMT.053/MT940/QIF), exchange rates, quotes (with `paymentOptions`
> fees), recipients, transfers (create / get / list / cancel / **fund** /
> receipt / MT103 payout info / payments list / US combined receipt / Indian
> NOC), third-party (originator) transfers, delivery estimates, and
> transfer-requirements validation — plus webhook subscriptions with typed event
> decoding, webhook signature verification, programmatic SCA challenge clearing
> (one-time-token endpoints), observability hooks, and balance lifecycle
> management. See
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
  - [Users](#users)
  - [Balances](#balances)
  - [Transactions](#transactions)
  - [Statements](#statements)
  - [Transfer receipts & payout paperwork](#transfer-receipts--payout-paperwork)
  - [Quotes](#quotes)
  - [Recipients](#recipients)
  - [Exchange rates](#exchange-rates)
  - [Transfers](#transfers)
  - [Core transfer flow](#core-transfer-flow-quote--recipient--transfer)
  - [Multi-Currency Account & bank details](#multi-currency-account--bank-details)
  - [Currencies](#currencies)
- [Mocking the Client](#mocking-the-client)
- [Request Middleware](#request-middleware)
- [Dependency Injection & Health Checks](#dependency-injection--health-checks)
- [Error Handling](#error-handling)
- [Strong Customer Authentication (SCA)](#strong-customer-authentication-sca)
- [Webhooks](#webhooks)
  - [Webhook subscriptions](#webhook-subscriptions)
  - [Typed event decoding](#typed-event-decoding)
- [Design Decisions](#design-decisions)
- [Testing](#testing)
  - [Sandbox verification](#sandbox-verification)
- [Project Status](#project-status)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Money is never `float64`** — Every amount is `int64` minor units (cents). No IEEE-754 representation error, ever.
- **Branded IDs prevent entity-mixing bugs** — `ProfileID`, `BalanceID`, `TransactionID`, `TransferID`, `RecipientID`, and `QuoteID` are distinct types; passing one where another belongs is a compile error.
- **Automatic retries with backoff** — Exponential backoff with jitter on 429 (rate limit), 5xx, and network errors via [go-retry](https://github.com/larsartmann/go-retry), honoring Wise's `Retry-After` (capped at the configured max delay). Auth, not-found, and client errors fail immediately.
- **Typed, classifiable errors** — `AuthError`, `RateLimitError` (with parsed `Retry-After`), `NotFoundError`, `ServerError`. Each carries its Wise API detail and implements `ErrorCode()` / `ErrorFamily()` / `IsRetryable()` from [go-error-family](https://github.com/larsartmann/go-error-family).
- **Two-layer type system** — Raw wire types live in `internal/raw`; result types expose clean Go with `Money` value objects and branded `Currency`. The mapping is the only bridge.
- **Write operations** — create quotes (authenticated and unauthenticated), recipients, and transfers; fund transfers from a balance; cancel transfers; validate transfer requirements; fetch delivery estimates.
- **SCA challenge support** — SCA-protected endpoints (e.g. balance statements for UK/EEA profiles) surface as `*SCAChallengeError` with the one-time approval token; complete the challenge with `WithSCAApprovalToken` and retry. See [SCA](#strong-customer-authentication-sca).
- **Statement & receipt downloads** — balance statements in CSV/PDF/XLSX/CAMT.053/MT940/QIF, transfer receipts (PDF), and MT103 payout proofs returned as raw bytes with client-side validation.
- **Webhook subscriptions & typed events** — register, list, and delete profile webhook subscriptions, then parse deliveries with `ParseWebhookEvent`: typed payloads for transfer state changes, payout failures, and balance credits; unknown event types never break your handler.
- **Identify your integration** — `WithUserAgent` replaces Go's default client agent with your own (`bank-sync/1.0`), as Wise recommends.
- **Tolerant timestamp handling** — Wise emits four different timestamp formats. One parser accepts them all (zoneless = UTC), and outgoing query timestamps are normalized to UTC `Z` (Wise rejects zone offsets with 422).
- **Sandbox support** — One-line switch to the Wise sandbox environment.
- **Minimal dependencies** — Three focused production deps: `go-retry`, `go-branded-id`, `go-error-family`.

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

    // List balances for the first profile (kept outside the loop so the
    // transaction query below can reference it)
    balances, err := client.ListBalances(ctx, profiles[0].ID)
    if err != nil {
        log.Fatal(err)
    }

    for _, p := range profiles {
        fmt.Printf("%s (%s)\n", p.Name, p.Type)

        // List balances for each profile
        profileBalances, err := client.ListBalances(ctx, p.ID)
        if err != nil {
            log.Fatal(err)
        }

        for _, b := range profileBalances {
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

**Retry composition — multiply carefully.** The SDK already retries the
transient class (429, 5xx, network) with backoff and `Retry-After` honoring;
a consumer-side retry loop around SDK calls multiplies with it (5 SDK
attempts × 3 loop passes = 15 hits per logical operation, each backing off).
Set `WithRetry` to the total attempts you want and let the SDK own the
delays; if you loop anyway, stop when the error is not retryable (`IsRetryable`
is absent on Rejection-family errors like auth and 4xx) so permanent
failures fail fast instead of pacing out.

### Observability

```go
// WithLogger observes every HTTP attempt, retries included — latency
// tracking and retry visibility without wrapping the transport.
client := wise.New("key", wise.WithLogger(wise.RequestLogFunc(func(e wise.RequestLog) {
    slog.Info("wise request",
        "method", e.Method, "url", e.URL, "status", e.Status,
        "duration", e.Duration, "attempt", e.Attempt, "err", e.Error)
})))

// Per-request correlation ID: one traceable X-External-Correlation-Id per
// logical operation, overriding the client-wide WithCorrelationID value.
ctx := wise.WithRequestCorrelationID(ctx, "bank-sync-2026-08-21-run-42")
_, err := client.ListTransfers(ctx, req)
```

`RequestLog.Attempt` is 1-based; a value above 1 means the request was retried.

### mTLS (client-certificate authentication)

Wise serves compliance-heavy integrations from dedicated mTLS endpoints
(`api-mtls.wise.com`, sandbox `api-mtls.wise-sandbox.com`) that authenticate
the client certificate in addition to the API token. The SDK supports this
without a dedicated option: swap the base URL and inject a TLS-configured
`http.Client` via `WithHTTPClient`:

```go
clientCert, err := tls.LoadX509KeyPair("client-cert.pem", "client-key.pem")
if err != nil {
    log.Fatal(err)
}

transport := &http.Transport{
    TLSClientConfig: &tls.Config{
        Certificates: []tls.Certificate{clientCert},
        MinVersion:   tls.VersionTLS12,
    },
}

client := wise.New("api-token",
    wise.WithBaseURL("https://api-mtls.wise.com"),
    wise.WithHTTPClient(&http.Client{Transport: transport, Timeout: 30 * time.Second}),
)
```

The same `WithHTTPClient` pattern covers custom root CAs, corporate proxies,
and connection-pool tuning; retries, logging, and typed errors work unchanged.

> **Injected transports see everything.** A custom `Transport` (or
> `http.RoundTripper` middleware) observes the `Authorization` header and the
> full request/response bodies — bank details, balances, transfer amounts.
> Whatever you log or forward from that layer must be redacted accordingly;
> the SDK's own `RequestLog` deliberately carries metadata only (method, URL,
> status, duration, attempt), never bodies or tokens.

## API Reference

### Authentication

```go
// Validate API key (calls ListProfiles internally)
err := client.Authenticate(ctx)

// Readiness probe: authenticated ListProfiles through the retry executor.
// A *RateLimitError proves reachability (answered, throttled); an *AuthError
// proves reachability with bad credentials; worst case spends the full
// retry budget (~2 min). Not a cheap liveness check.
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

### Users

```go
// The account behind the API token — the identity of ListProfiles' userId values
me, err := client.GetMe(ctx)
// me.ID, me.Name, me.Email, me.Active
// me.Details → *UserDetails (nil when Wise has none on file):
//   FirstName, LastName, PhoneNumber, DateOfBirth (UTC date; zero when unset),
//   Occupation, Avatar, PrimaryAddress, Address (*UserAddress)

// A user by ID — personal tokens may only read their own user
user, err := client.GetUser(ctx, userID)
```

`Details` and `Details.Address` are pointers: nil means Wise has nothing on
file, not an error. `DateOfBirth` is parsed from Wise's date-only wire format
as a UTC midnight `time.Time`.

### Balances

```go
// List visible, non-investment balances for a profile
balances, err := client.ListBalances(ctx, profileID)
// []Balance{
//   {ID: 100, Currency: "EUR", Type: BalanceTypeStandard, Name: "Main Account", Amount.Cents: 123456, ...},
// }

// Get a specific balance by ID (direct per-balance endpoint)
balance, err := client.GetBalance(ctx, profileID, balanceID)

// Open a new balance (savings balances require a name)
created, err := client.CreateBalance(ctx, wise.CreateBalanceRequest{
    ProfileID: profileID,
    Currency:  wise.Currency("USD"),
    Type:      wise.BalanceTypeSavings,
    Name:      "Vacation",
})

// Profile-wide funds overview converted to one currency
total, err := client.GetTotalFunds(ctx, profileID, wise.Currency("EUR"))
// total.Worth     → cash + invested portfolio valuation (Money)
// total.Available → cash + approved overdraft limit (Money)
```

`ListBalances` filters to `Visible: true` and `InvestmentState == "NOT_INVESTED"`
only. `GetBalance` reads the direct per-balance endpoint
(`GET /v4/profiles/{id}/balances/{id}`) and does **not** filter — hidden and
invested balances are retrievable individually even though `ListBalances`
drops them.

### Transactions

```go
resp, err := client.ListTransactions(ctx, wise.ListTransactionsRequest{
    ProfileID: wise.NewProfileID(12345),
    BalanceID: wise.NewBalanceID(100),
    Currency:  wise.Currency("EUR"),
    From:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
    To:        time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
    Type:      wise.StatementTypeFlat, // optional statement layout (COMPACT or FLAT)
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

### Statements

`GetStatement` downloads a balance statement as a **file** and returns its raw
bytes — for imports, archival, or PDF rendering. Use `ListTransactions` for
the typed JSON statement.

```go
pdf, err := client.GetStatement(ctx, wise.GetStatementRequest{
    ProfileID: profileID,
    BalanceID: balanceID,
    Currency:  wise.Currency("EUR"),
    From:      intervalStart, // UTC
    To:        intervalEnd,   // UTC
    Type:      wise.StatementTypeFlat,  // optional statement layout (COMPACT or FLAT)
    Locale:    "de",                      // optional 2-char statement language
    Format:    wise.StatementFormatPDF,
})
```

Six formats: `StatementFormatCSV`, `StatementFormatPDF` (includes Wise
branding), `StatementFormatXLSX` (Excel), `StatementFormatCAMT` (CAMT.053
XML), `StatementFormatMT940`, and `StatementFormatQIF`. Unknown formats are
rejected client-side. The interval rules mirror `ListTransactions`; the
interval must not exceed **469 days** (Wise limit) — longer periods are
rejected client-side with an actionable error, so split them into multiple
requests. `Locale` (2-character language code) localizes column headers and
descriptions. Like the JSON statement, the endpoint is **SCA-protected** for
UK/EEA profiles — see
[SCA](#strong-customer-authentication-sca).

### Transfer receipts & payout paperwork

`GetTransferReceipt` downloads the branded confirmation receipt for a
paid-out transfer as raw PDF bytes. Wise answers 404 until the transfer
reached `outgoing_payment_sent` — treat `*wise.NotFoundError` as "no receipt
(yet)", not a hard failure. Unlike statements, the endpoint is **not
SCA-protected**.

```go
pdf, err := client.GetTransferReceipt(ctx, transferID)
```

`GetTransferPayoutInfo` returns the banking-partner proof of payment,
including the SWIFT MT103 message (`MT103` is `*string`, `nil` for non-SWIFT
corridors):

```go
info, err := client.GetTransferPayoutInfo(ctx, transferID)
```

The rest of the payout paperwork set, all raw-byte downloads unless noted:

```go
// Completed funding payments of a transfer ([]TransferPayment)
payments, err := client.ListTransferPayments(ctx, transferID)

// US combined receipt (transfer confirmation + funding proof in one PDF;
// required by some US institutions). 404 = does not exist yet.
usPdf, err := client.GetUSCombinedReceipt(ctx, transferID)

// Indian FIRC no-objection certificate. 404 before payout = not issued yet.
noc, err := client.GetTransferNOC(ctx, transferID)
```

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
route — use the account-requirements endpoint to discover them per corridor
(the map is flat: corridors that require nested values such as an `address`
object are outside what the map can express today):

```go
requirements, err := client.GetQuoteAccountRequirements(ctx, wise.QuoteAccountRequirementsRequest{
    QuoteID: quote.ID,
})

// Second pass: when a returned field is flagged RefreshRequirementsOnChange,
// submit the partial form with that field set and Wise returns the revised
// field set for the route (e.g. legalEntityType reveals further fields).
requirements, err = client.RefreshQuoteAccountRequirements(ctx, wise.RefreshQuoteAccountRequirementsRequest{
    QuoteID: quote.ID,
    Recipient: wise.CreateRecipientRequest{
        Currency: quote.TargetCurrency,
        Type:     requirements[0].Type,
        Details:  map[string]string{"legalEntityType": "PRIVATE"},
    },
})
```

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

Funding a transfer from the profile's balance:

```go
funding, err := client.FundTransfer(ctx, profileID, transferID)
// funding.Status → wise.FundingStatusCompleted / FundingStatusCreated / FundingStatusRejected
// funding.BalanceTransactionID → the debit applied to the balance (BALANCE funding)
```

A rejected funding (e.g. `FundingErrorCodeBalanceInsufficientFunds`) is a
_successful API call that declined_ — top up the balance and call `FundTransfer`
again. `ErrorCode` is an open enum: known codes are constants, unknown codes
pass through. The endpoint is SCA-protected for UK/EEA profiles.

Moving money **on behalf of a third-party originator** — your own quote funds
the transfer, but the compliance data names the real originator. Reuse the
same `Originator.Reference` for every transfer of one originator (Wise keys
compliance on it); `OriginalTransferID` is your idempotency key:

```go
transfer, err := client.CreateThirdPartyTransfer(ctx, wise.CreateThirdPartyTransferRequest{
    ProfileID:          profileID,
    OriginalTransferID: "our-originator-transfer-correlation-id",
    TargetAccount:      recipientID,
    QuoteID:            quoteID,
    Reference:          "INV-2026-014",
    Originator: wise.Originator{
        LegalEntityType: wise.OriginatorLegalEntityTypePrivate,
        Reference:       "originator-ref-42",
        Name:            wise.OriginatorName{GivenName: "Ana", FamilyName: "Martins"},
        DateOfBirth:     time.Date(1988, 4, 12, 0, 0, 0, 0, time.UTC),
        Nationality:     "PT",
        Address: wise.OriginatorAddress{
            FirstLine: "Rua das Flores 12", City: "Lisboa", CountryCode: "PT", PostCode: "1200-192",
        },
    },
})
// fetch it back with client.GetThirdPartyTransfer(ctx, profileID, transferID)
```

### Core transfer flow: quote → recipient → transfer

The 1% Pareto core. A consumer with `GetProfile`, `CreateQuote`, `CreateRecipient`,
and `CreateTransfer` can move money end-to-end:

```go
ctx := context.Background()

// 1. Quote — locks the rate for 30 minutes.
//    PreferredPayIn=BALANCE keeps quote fees consistent with transfer fees when
//    funding from a multi-currency balance.
quote, err := client.CreateQuote(ctx, profileID, wise.CreateQuoteRequest{
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
    ProfileID:         profileID,
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

// Cross-reference the response against your prepared request to learn which
// required fields are still unsatisfied (select choices are checked against
// their allowed values):
transferReq := wise.CreateTransferRequest{
    QuoteID:               quote.ID,
    TargetAccount:         recipient.ID,
    CustomerTransactionID: wise.NewCustomerTransactionID("22244c35-9fe8-4c32-b7fd-d05c2a7734bf"),
}
if missing := wise.MissingTransferDetails(requirements, transferReq); len(missing) > 0 {
    log.Fatalf("transfer details still missing: %v", missing)
}

// 3. Transfer — customerTransactionId is your idempotency key (a UUID).
//    Reusing it with the same quote + targetAccount returns the existing
//    transfer instead of creating a duplicate.
transfer, err := client.CreateTransfer(ctx, wise.CreateTransferRequest{
    QuoteID:               quote.ID,
    TargetAccount:         recipient.ID,
    CustomerTransactionID: wise.NewCustomerTransactionID("22244c35-9fe8-4c32-b7fd-d05c2a7734bf"),
})
if err != nil {
    log.Fatal(err)
}

// 4. Fund — debit the balance to trigger processing of the payout. This is the
//    final money-movement step; until it succeeds the transfer stays
//    waiting_for_funds. An underfunded balance is not an error return but a
//    rejected result — top up and call FundTransfer again.
funding, err := client.FundTransfer(ctx, profileID, transfer.ID)
if err != nil {
    log.Fatal(err)
}
if funding.Status == wise.FundingStatusRejected {
    log.Fatalf("funding rejected (%s): %s", funding.ErrorCode, funding.ErrorMessage)
}

// 5. Track — the delivery estimate for the now-funded transfer.
estimate, err := client.GetDeliveryEstimate(ctx, transfer.ID, "Europe/Berlin")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("expected arrival: %s\n", estimate.EstimatedDeliveryDate)

// 6. Cancel — only possible before the transfer is processed.
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

### Multi-Currency Account & bank details

```go
// The profile's account container: holder of the currency balances
account, err := client.GetMultiCurrencyAccount(ctx, profileID)
// account.RecipientID → the account's own recipient entry

// One set of receiving bank details by its own ID (the 2027Q1 surface;
// the old profile-wide list endpoint was removed by Wise)
detail, err := client.GetBankAccountDetails(ctx, detailsID)
// AccountDetails{State: AccountDetailsStateActive, Country: "GB", Identification: ...}
```

Two patterns worth knowing:

- **Top up via transfer** — `MultiCurrencyAccount.RecipientID` is a real
  recipient. Create a transfer to it (from an external-funded quote) and the
  matching currency balance gets credited: the standard way to move money
  _into_ the Multi-Currency Account.
- **Identification is a typed oneOf** — `detail.Identification` carries
  exactly one of `BE` (IBAN), `GB` (sort code + account number), or `US`
  (routing + account number); anything else from the wire is surfaced as a
  corruption error. Ordering new details still goes through Wise's
  account-details-orders endpoint (not in the SDK yet).

### Currencies

```go
// Public reference data — no authentication required
currencies, err := client.ListCurrencies(ctx)
// []CurrencyInfo{
//   {Code: "EUR", Symbol: "€", Name: "Euro", CountryKeywords: [...], SupportsDecimals: true},
// }
```

Populate currency pickers and validate corridor choices against Wise's list
(code, symbol, name, country keywords, decimal support).

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

## Dependency Injection & Health Checks

wise-go is a plain-Go library: it imports no DI framework and owns no
resources (no goroutines, no `Shutdown`). That makes it a drop-in for any
composition root. Below is the compile-verified recipe for
[samber/do](https://github.com/samber/do) v2, including the two-line adapter
that plugs the readiness probe into `do`'s health-check interface:

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/larsartmann/wise-go"
    "github.com/samber/do/v2"
)

// healthAdapter adapts the SDK's Health method to samber/do's health-check
// interface. The adapter is needed because the SDK method is named Health,
// not HealthCheck (wise-go deliberately imports no DI framework).
type healthAdapter struct{ client *wise.Client }

func (h healthAdapter) HealthCheck(ctx context.Context) error {
    return h.client.Health(ctx)
}

func main() {
    injector := do.New()

    do.Provide(injector, func(i do.Injector) (*wise.Client, error) {
        return wise.New(os.Getenv("WISE_API_KEY"), wise.WithSandbox()), nil
    })

    do.ProvideNamed(injector, "wise.health", func(i do.Injector) (do.HealthcheckerWithContext, error) {
        return healthAdapter{client: do.MustInvoke[*wise.Client](i)}, nil
    })

    client, err := do.Invoke[*wise.Client](injector)
    if err != nil {
        slog.Error("wise client unavailable", "err", err)
        os.Exit(1)
    }

    checker := healthAdapter{client: client}
    if err := checker.HealthCheck(context.Background()); err != nil {
        slog.Warn("wise API not ready", "err", err)
    }
}
```

`Health` is a readiness probe, not a cheap liveness check: it runs an
authenticated `ListProfiles` through the retry executor, so a
`*RateLimitError` means the API answered under throttling, an `*AuthError`
means bad credentials, and the worst case spends the full retry budget
(~2 min). Wrap the adapter with your own timeout or retry policy if your
orchestrator expects sub-second probes.

> Naming note: whether the SDK should also expose a `HealthCheck(ctx)` alias
> (so `do.HealthcheckerWithContext` containers satisfy the interface without
> an adapter) is a tracked open decision; the recipe above works unchanged
> either way.

## Error Handling

The SDK returns typed errors you can match with `errors.AsType` (Go 1.26+):

```go
import "errors"

balances, err := client.ListBalances(ctx, profileID)
if err != nil {
    if rl, ok := errors.AsType[*wise.RateLimitError](err); ok {
        fmt.Printf("rate limited (scope: %s), retry after %s\n", rl.RateLimitedBy, rl.RetryAfter)
    } else if sca, ok := errors.AsType[*wise.SCAChallengeError](err); ok {
        fmt.Printf("SCA approval required (token issued: %t)\n", sca.TwoFAApprovalToken() != "")
    } else if auth, ok := errors.AsType[*wise.AuthError](err); ok {
        fmt.Printf("auth failed: %s\n", auth.Message)
    } else if nf, ok := errors.AsType[*wise.NotFoundError](err); ok {
        fmt.Printf("not found: %s\n", nf.Message)
    } else if br, ok := errors.AsType[*wise.BadRequestError](err); ok {
        fmt.Printf("rejected (%d): %s\n", br.StatusCode, br.Message)
    } else if srv, ok := errors.AsType[*wise.ServerError](err); ok {
        fmt.Printf("server error (%d): %s\n", srv.StatusCode, srv.Message)
    } else if apiErr, ok := errors.AsType[*wise.APIError](err); ok {
        fmt.Printf("API error (%d): %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

| HTTP Status            | Error Type          | Retried? |
| ---------------------- | ------------------- | -------- |
| 400                    | `BadRequestError`   | No       |
| 401, 403               | `AuthError`         | No       |
| 403 + Wise 2FA headers | `SCAChallengeError` | No       |
| 404                    | `NotFoundError`     | No       |
| 429                    | `RateLimitError`    | Yes      |
| 5xx                    | `ServerError`       | Yes      |
| Network error          | Wrapped `error`     | Yes      |

Mapper failures (unparseable timestamps, invalid currency codes in a response)
carry `Corruption` classification — permanent, not retryable — so a blanket
"retry on transient" wrapper fails fast instead of looping.

### How to match an error

Three matchers, one decision tree:

1. **Need the error's fields?** (status code, `RetryAfter`, rate-limit scope,
   the SCA token) — extract the type with `errors.AsType`.
2. **Branching on a Wise granular code with a curated sentinel?** (`not.valid`,
   `resource_not_found`, `rate_limit_exceeded`, ...) — match by value with
   `errors.Is`. This works through **any** `fmt.Errorf("%w")` wrap chain: the
   SDK's contextual wraps ("create transfer for quote ...: ...") never break
   the match.
3. **Any other granular code?** (the open enum — Wise adds codes over time) —
   read the structured details with `APIError.Details` or branch with
   `APIError.HasCode`. Both are promoted onto **every** typed error, so
   extract the type your HTTP status maps to (table above) and call them
   there.

```go
transfers, err := client.ListTransfers(ctx, wise.ListTransfersRequest{
    ProfileID: wise.NewProfileID(12345), // optional, defaults to personal profile
})
if err != nil {
    // Curated Wise codes match by value through any wrap chain:
    if errors.Is(err, wise.ErrRateLimitExceeded) {
        fmt.Println("wise throttled us; back off and retry")
    } else if errors.Is(err, wise.ErrResourceNotFound) {
        fmt.Println("resource gone; refresh before retrying")
    }

    // Any code (curated or not) branches on the structured details. The
    // status drives the type — here a 400 is *BadRequestError — and
    // HasCode/Details are promoted onto every typed error:
    if br, ok := errors.AsType[*wise.BadRequestError](err); ok {
        for _, detail := range br.Details {
            fmt.Printf("wise says %q: %s\n", detail.Code, detail.Message)
        }

        if br.HasCode(wise.WiseErrorCodeNotValid) {
            fmt.Println("wise rejected a value as not valid; do not retry")
        }
    }
}
```

`errors.AsType` and `errors.Is` are complementary, not competing: types carry
the fields, sentinels carry the codes. A 429 whose body says
`rate_limit_exceeded` satisfies both; a 5xx whose body says `system_error` is
still a retryable `*ServerError` — the status drives the type, the body code
drives the match.

Every error the SDK emits is named in one central catalog (`errors.go`): the
typed errors above with their `ErrorCode()` values, one named constant per
`errorfamily` error code, the `WiseErrorCode` constants for Wise's granular
wire codes, and exported sentinel errors. Two sentinel families exist:

- **Wise code sentinels** (wave 1): `ErrNotValid`, `ErrInvalidValue`,
  `ErrParameterMissing`, `ErrParameterInvalid`, `ErrInvalidRequest`,
  `ErrUnauthorized`, `ErrForbidden`, `ErrResourceNotFound`,
  `ErrOperationNotSupported`, `ErrRateLimitExceeded`, `ErrSystemError`. An
  `APIError` matches one when any structured detail carries the wire code the
  sentinel stands for:

  ```go
  if errors.Is(err, wise.ErrNotValid) {
      fmt.Println("wise rejected a value as not valid")
  }
  ```

- **Response-shape sentinels**: `ErrEmptyRateList` (rates payload with zero
  entries), `ErrAccountIdentificationCardinality` (account details
  `accountIdentification` not exactly one member), `ErrOTTAlternativesShape`
  (SCA challenge alternatives neither object nor string form), and
  `ErrTransferRefundAmountMissing` (`transfers#refund` webhook payload without
  amount/currency). The SDK returns these wrapped but always matchable:

  ```go
  if errors.Is(err, wise.ErrEmptyRateList) {
      fmt.Println("the rates response decoded to zero entries")
  }
  ```

Every error also exposes `ErrorContext()` for log pipelines; `APIError`-family
errors include `status_code` and, when Wise sent its structured envelope, the
first granular code as `wise_code`. The full reference of granular codes lives
in [docs/error-codes.md](docs/error-codes.md).

One error type carries its verdict in headers instead of the body:
`SCAChallengeError` (403 + Wise's two-factor headers) never contains granular
codes. Clear the challenge through the one-time-token endpoints, then retry
with the token from `TwoFAApprovalToken()`; the full flow is in
[Strong Customer Authentication](#strong-customer-authentication-sca).

### Consumer recipe: transfer with funding fallback

The pattern bank-sync runs in production: branch on the typed error for
transport decisions, on the granular codes for Wise's reasons, and on the
funding result's own `FundingErrorCode` for pay-in problems.

```go
transfer, err := client.CreateTransfer(ctx, createReq)
if err != nil {
    switch {
    case errors.Is(err, wise.ErrNotValid):
        // Wise rejected a value; retrying identical input will fail again.
        return fmt.Errorf("fix the transfer request, do not retry: %w", err)
    case errors.Is(err, wise.ErrResourceNotFound):
        // The quote or recipient vanished; refresh, then build a new request.
        return fmt.Errorf("refresh the referenced resource: %w", err)
    case errors.Is(err, wise.ErrRateLimitExceeded):
        // Transient; the SDK's retry path honors Retry-After automatically.
        return fmt.Errorf("throttled, retry later: %w", err)
    default:
        return fmt.Errorf("create transfer: %w", err)
    }
}

funded, err := client.FundTransfer(ctx, profileID, transfer.ID)
if err != nil {
    return fmt.Errorf("fund transfer: %w", err)
}

if funded.Status == wise.FundingStatusRejected {
    if funded.ErrorCode == wise.FundingErrorCodeBalanceInsufficientFunds {
        // Top up the balance, then fund the same transfer again; the
        // transfer itself was created fine.
        return ErrTopUpRequired
    }
}
```

## Strong Customer Authentication (SCA)

Some Wise endpoints (balance statements for UK/EEA profiles among them) are
SCA-protected: the API answers **HTTP 403 with an empty body**. The verdict and
the one-time token live in the `x-2fa-approval-result` / `x-2fa-approval`
response headers. The SDK detects this and returns `*SCAChallengeError` instead
of a bare `AuthError`.

**Which reads need SCA?**

| Endpoint family                                                     | Personal token (UK/EEA profile) | Personal token (other regions) | Basis                      |
| ------------------------------------------------------------------- | ------------------------------- | ------------------------------ | -------------------------- |
| Balance statements (`GetStatement`, `ListTransactions`)             | **SCA**                         | no                             | live-verified 2026-08      |
| `FundTransfer`                                                      | **SCA**                         | no                             | godoc + live behavior      |
| `GetTransfer`, `ListTransfers`                                      | no                              | no                             | live-verified 2026-10-16   |
| Transfer receipts (`GetTransferReceipt`, `GetUSCombinedReceipt`)    | no                              | no                             | not SCA-protected per docs |
| `GetTransferPayoutInfo`, `ListTransferPayments`                     | no                              | no                             | same family as receipts    |
| Profiles, balances, quotes, recipients, rates, currencies, webhooks | no                              | no                             | standard token reads       |

SCA protection keys off the **profile's registration region**, not the token
type: a business profile registered outside the UK/EEA reads statements with a
bare API key. When a challenge hits, the OTT flow clears it programmatically:

```go
// One correlation ID for the whole logical operation — the challenge and the
// replay both carry it in X-External-Correlation-Id.
ctx = wise.WithRequestCorrelationID(ctx, "bank-sync-2026-08-21-001")

resp, err := client.ListTransactions(ctx, req)
if err != nil {
    if sca, ok := errors.AsType[*wise.SCAChallengeError](err); ok {
        // 1. The one-time token (OTT) — fetch it explicitly; it never
        //    appears in error messages.
        ott := sca.TwoFAApprovalToken()

        // 2. Clear the challenge over a phone channel: trigger the SMS,
        //    prompt the user for the code it delivered, verify, repeat
        //    until every required challenge is passed.
        status, clearErr := client.ClearSCAChallenge(ctx, ott, wise.OTTChannelSMS,
            func(ctx context.Context, challenge wise.OTTChallenge, channel wise.OTTChannel, phoneHint string) (string, error) {
                fmt.Printf("code sent to %s — enter OTP: ", phoneHint)
                var code string
                if _, err := fmt.Scanln(&code); err != nil {
                    return "", err
                }
                return code, nil
            })
        if clearErr != nil {
            log.Fatal(clearErr)
        }
        fmt.Println("cleared:", status.Cleared())

        // 3. Replay the request — with the same ctx so Wise support can
        //    correlate both attempts:
        scaClient := wise.New("api-key", wise.WithSCAApprovalToken(ott))
        resp, err = scaClient.ListTransactions(ctx, req)
    }
}
```

The one-time-token endpoints (`GetOTTStatus`, `TriggerOTT`, `VerifyOTT`,
`ClearSCAChallenge`) live on the quarterly versioned surface
(`/2026Q3/one-time-token/...`) and authorize with the same personal API
token. The OTT is a distinct type, `wise.OneTimeToken` (construct with
`wise.NewOneTimeToken`), so it can never be passed where an OTP code or any
other string is expected. Challenges that no phone channel can clear (e.g. a
PIN-only challenge) surface as a Rejection naming the available types — clear
those by viewing a statement on wise.com instead. Wise's PSD2 flavor can also
issue a **tokenless** challenge (403 with the verdict header but no
`x-2fa-approval`): `TwoFAApprovalToken()` then returns the zero
`OneTimeToken` — still blocked, but with nothing to clear programmatically.

`GET /v1/transfers` (and the other transfer endpoints) are **not** SCA-protected,
which is why `ListTransfers` is the reliable source for outgoing transfer
history with personal API tokens.

## Webhooks

Wise signs every webhook delivery with `X-Signature-SHA256`: a Base64-encoded
RSA-SHA256 signature of the raw request body, verifiable with the per-subscription
public key Wise shows in the dashboard. Verify every delivery before trusting it:

```go
// Parse the key once when the handler is constructed; a bad key should fail
// loudly there, not on the first delivery. Register the handler once:
//
//	http.HandleFunc("/webhooks/wise", newWebhookHandler(subscriptionPublicKeyPEM))
func newWebhookHandler(subscriptionPublicKeyPEM string) http.HandlerFunc {
    webhookKey, err := wise.ParseWebhookPublicKey([]byte(subscriptionPublicKeyPEM))
    if err != nil {
        log.Fatal(err)
    }

    return func(w http.ResponseWriter, r *http.Request) {
        body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "read body", http.StatusBadRequest)

            return
        }

        // Verify the raw bytes exactly as received — before any re-marshalling,
        // which would change the signed input.
        if !wise.VerifyWebhookSignature(body, r.Header.Get(wise.HeaderWebhookSignature), webhookKey) {
            http.Error(w, "invalid signature", http.StatusUnauthorized)

            return
        }

        // Trustworthy delivery: parse the envelope and its typed payload —
        // see [Typed event decoding](#typed-event-decoding) below.
    }
}
```

Reject deliveries with HTTP 401/403 when verification fails. Wise redelivers
until you acknowledge, so process **idempotently**: `X-Delivery-Id` (unique
per delivery attempt — exposed as `wise.HeaderDeliveryID`) is the
deduplication key — record it and skip already-seen deliveries. Signature
verification alone is not enough: a verified delivery can still be a
redelivery of an event you already processed.

### Webhook subscriptions

Subscriptions are the standing instruction for Wise to deliver events to your
HTTPS endpoint. The SDK covers the profile-level CRUD (user API token);
application-level subscriptions need a client-credentials token and are
deliberately out of scope (see ROADMAP):

```go
profileID := wise.NewProfileID(12345)

sub, err := client.CreateProfileWebhookSubscription(ctx, profileID,
    wise.CreateWebhookSubscriptionRequest{
        Name:      "Payout watcher",
        TriggerOn: wise.WebhookEventTransfersStateChange, // open enum; any event type string works
        Delivery:  wise.WebhookDelivery{Version: "4.0.0", URL: "https://example.com/hooks/wise"},
    })

subs, err := client.ListProfileWebhookSubscriptions(ctx, profileID)
sub, err := client.GetProfileWebhookSubscription(ctx, profileID, sub.ID)
err := client.DeleteProfileWebhookSubscription(ctx, profileID, sub.ID)
```

Wise's subscription endpoints live on the quarterly versioned API surface
(`/2026Q4/profiles/{profileId}/subscriptions`) — the SDK sends that path
automatically. There is no update operation: subscriptions are created and
deleted, never edited.

### Typed event decoding

`ParseWebhookEvent` decodes the envelope Wise POSTs (schema version,
subscription ID, event type, sent-at) and keeps the payload raw. Typed
accessors decode the payload per event type; unknown event types never fail —
new Wise events parse with the data left raw for hand decoding:

```go
event, err := wise.ParseWebhookEvent(body) // body = the verified raw bytes
if err != nil {
    // Undecodable or malformed envelope: reject the delivery.
    http.Error(w, "malformed event", http.StatusBadRequest)

    return
}

switch event.EventType {
case wise.WebhookEventTransfersStateChange:
    state, err := event.TransferStateChange() // *TransferStateChangeData
    // state.CurrentState / state.PreviousState are TransferStatus values.
case wise.WebhookEventTransfersPayoutFailure:
    failure, err := event.TransferPayoutFailure() // *TransferPayoutFailureData
case wise.WebhookEventBalancesCredit:
    credit, err := event.BalanceCredit() // *BalanceCreditData with Money amounts
default:
    // Unknown event type: log and acknowledge — Wise redelivers otherwise.
}
```

Available typed payloads today: `transfers#state-change`,
`transfers#payout-failure`, and `balances#credit`. Malformed envelopes and
payloads are corruption-classified errors (`wise.webhook.decode`), so a
broken sender is distinguishable from a routing mistake.

## Design Decisions

| Decision                          | Rationale                                                                                                                                              |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Monetary amounts as `int64` cents | `float64` causes precision loss (e.g., `0.1 + 0.2 ≠ 0.3`). Cents are safe for arithmetic and storage.                                                  |
| Two-layer type system             | Raw wire types in `internal/raw` mirror JSON exactly. Result types expose clean Go with `Money` value objects. Mapping functions convert between them. |
| `go-retry` for retries            | In-house, zero-dependency retry with backoff and jitter; the only executor that expresses Wise's `Retry-After` as first-class policy (ADR 003).        |
| Flat package structure            | Single `package wise`; wire types hidden in `internal/raw`. The import path is the API.                                                                |
| BDD tests with Ginkgo             | `httptest.Server` mock API responses. Tests verify both happy paths and error classification.                                                          |

## Testing

```bash
GOEXPERIMENT=jsonv2 go test ./...
```

Tests use `net/http/httptest` to mock the Wise API — no network access required. The `GOEXPERIMENT=jsonv2` prefix is mandatory (see [Installation](#installation)).

### Spec conformance

Every mock exchange is also validated against Wise's machine-readable
OpenAPI contract, vendored at `docs/reviews/wise-api-openapi.json`: the
suite fails when a request path, required parameter, or response payload
no longer matches what Wise documents. Deliberate divergences (Wise's
loose live timestamp shapes, legacy wire surfaces the spec bundle does
not describe) are exempted explicitly in `spec_conformance_test.go` with
the reasoning inline. Refresh the snapshot from
`https://docs.wise.com/_bundle/api-reference/@preview/index.yaml?download`
to pick up contract changes.

### Sandbox verification

A small live-test suite (`sandbox_live_test.go`) exercises the real sandbox at
`api.wise-sandbox.com`. It is **key-gated**: every test skips unless
`WISE_SANDBOX_API_KEY` is set, so the default suite stays offline and CI never
fails on a missing secret.

Locally:

```bash
WISE_SANDBOX_API_KEY=<your-sandbox-key> GOEXPERIMENT=jsonv2 \
  go test -run TestSandboxLive -v -count=1 ./...
```

In CI: the [Sandbox Live Tests](.github/workflows/sandbox-live.yml) workflow
runs the same suite on manual dispatch (`workflow_dispatch`). Add the
`WISE_SANDBOX_API_KEY` secret to the repository first; dispatching without it
is a green no-op (all tests skip, visible as `SKIP` lines in the log). Current
coverage: a `GetMe` smoke test, a `ListTransfers` read (must succeed — the
endpoint is not SCA-protected), and a `GetTransfer` by-ID test that stays
silent unless `WISE_TEST_TRANSFER_ID` names a transfer the key owns.

## Project Status

| Doc                                        | What it holds                                                                  |
| ------------------------------------------ | ------------------------------------------------------------------------------ |
| [FEATURES.md](FEATURES.md)                 | Honest feature inventory by status, including the per-endpoint coverage matrix |
| [ROADMAP.md](ROADMAP.md)                   | Long-term direction and raw ideas                                              |
| [CONTRIBUTING.md](CONTRIBUTING.md)         | Development setup, quality gates, release process                              |
| [CHANGELOG.md](CHANGELOG.md)               | What changed in each version                                                   |
| [docs/adr/](docs/adr/)                     | Architecture decision records (Money, package boundary, retry executor)        |
| [docs/error-codes.md](docs/error-codes.md) | Every granular Wise error code: sentinel, typical status, retry guidance       |

## Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on development setup, code style, and submitting pull requests.

## License

Apache-2.0 — see [LICENSE](LICENSE).

## Trademarks

"Wise" and "TransferWise" are registered trademarks of Wise Payments Limited. This project is not affiliated with, endorsed by, or sponsored by Wise Payments Limited. All trademarks are the property of their respective owners.
