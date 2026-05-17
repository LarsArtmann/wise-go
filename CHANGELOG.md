# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- Client with Bearer token authentication
- `ListProfiles` — list all profiles for the authenticated user
- `ListBalances` — list visible, non-investment balances for a profile
- `GetBalance` — get a specific balance by ID
- `ListTransactions` — list transactions for a balance within a date range
- `Authenticate` / `Health` — validate API key connectivity
- Automatic retries with exponential backoff (429, 5xx, network errors)
- Typed error hierarchy: `APIError`, `RateLimitError`, `AuthError`, `NotFoundError`, `ServerError`
- Strongly-typed result types with `int64` cents and `time.Time` dates
- Transaction type classification (card, credit, debit, exchange, fee, refund, transfer, payment)
- Sandbox environment support via `WithSandbox()`
- Functional options: `WithBaseURL`, `WithTimeout`, `WithRetry`, `WithHTTPClient`, `WithNow`
