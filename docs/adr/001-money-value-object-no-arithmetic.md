# ADR 001: `Money` is a value object without arithmetic

**Status:** Accepted (2026-08-08, codified 2026-09-13)
**Deciders:** Lars Artmann

## Context

The Wise API returns monetary amounts as `{value: 1234.56, currency: "EUR"}` pairs
(`raw.BalanceAmount`). A naive SDK would surface these as `float64` fields, which
loses precision in minor units and lets amounts decouple from their currency.
Downstream, consumers need to do arithmetic (sum fees, compare amounts) and are
tempted to expect the SDK to provide it.

## Decision

All public monetary amounts are `Money{Cents int64, Currency Currency}`:

- Cents are `int64` minor units (`1234.56 EUR` → `123456`), never `float64`.
- `Currency` is a validated typed string (3 uppercase ASCII letters).
- `Money` has **no** `Add`/`Sub`/`Equal`/`IsZero`/`IsNegative` methods — and none
  will be added without revisiting this ADR.

`Money` exists to pair an amount with its currency at the serialization boundary
so mismatched amounts are unrepresentable. The SDK is an anti-corruption layer:
call the API, parse responses, return typed data. Financial arithmetic is the
consumer's domain logic, not the SDK's job.

## Consequences

- Consumers doing math must operate on `Money.Cents` themselves and keep currency
  checks in their own domain code.
- No hidden cross-currency bugs from SDK-side arithmetic (there is nothing to
  misuse).
- `bench_test.go` pins `BalanceAmount.Cents` conversion at nanosecond scale; any
  future arithmetic addition would need a correctness story for currency
  mismatches (error vs panic vs unrepresentable) — that design work is why the
  methods do not exist yet.
- Related: AGENTS.md "Money is a value object, NOT a math library"; ADR 002 for
  the layering this sits in.
