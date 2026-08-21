# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).

## P1 — Close out the core transfer flow

[ ] Add `FundTransfer` — `POST /v1/profiles/{profileId}/transfers/{transferId}/payments`
(balance funding). The last mock-testable core-flow gap: after funding, the
quote → recipient → transfer → fund → delivery-estimate loop is complete end to end.
Source: `docs/status/2026-08-19_18-15_core-transfer-flow-completion.md` (f.1),
plan doc tier 1.

[ ] Add error-path BDD tests for the write endpoints — HTTP 400 validation, 409
cancel-not-allowed, 404, SCA 403, and 429-with-Retry-After for `CreateQuote`,
`CreateRecipient`, `CreateTransfer`, `CancelTransfer`,
`ValidateTransferRequirements` (zero-ID/missing-field rejections exist; the HTTP
error matrix does not — `ListRecipients`/`ListTransfers` have none either).
Source: status report 18-15 (f.2, f.4).

[ ] Add validation edge-case unit tests — `CreateTransferRequest.validate`
(`transfers.go:133`), `ValidateTransferRequirementsRequest.validate`
(`transfer_requirements.go:42`), `CreateQuoteRequest.validate` (`quotes.go:88`):
missing customerTransactionId/quote/target, amount/currency mismatch matrix.
Source: status report 18-15 (f.3).

[ ] Wire `ValidateTransferRequirements` output into transfer creation — a helper
that maps discovered required fields onto `CreateTransferRequest.Details` (or a
documented manual-mapping pattern). The discovery is currently one-way.
Source: status report 18-15 (f.5, e.7).

## P2 — v1.0 release (API lock)

[ ] Add credentialed Wise sandbox integration tests — create a protected or
manually triggered live-test workflow using sandbox credentials, cover the supported
read-only endpoints, and keep it separate from fork-safe mock tests. The sandbox URL
was updated to V2 (`api.wise-sandbox.com`); tests must verify the new endpoint works.
This is required to verify that the SDK actually interoperates with Wise's sandbox
rather than only matching mocked responses.

[ ] Lock the public API at v1.0 — formal audit of every exported symbol, godoc
review pass, then tag `v1.0.0`. The API surface is now stable: `Money`/`Currency`
value objects, branded IDs (incl. UUID `QuoteID`), typed enums (`DetailType`,
`TransactionType`, `ProfileType`, `BalanceType`, `InvestmentState`, `PayIn`/`PayOut`,
`QuoteStatus`, `RateType`, `ProvidedAmountType`), and the two-layer raw/result
split are finalized. Requires explicit approval (tagging is irreversible).

## P3 — Tier-2 API surface (plan doc tier 2 + near-term reports)

[ ] Add `GetQuoteAccountRequirements` — `GET /v1/quotes/{quoteId}/account-requirements`.
The only remaining tier-1 row in
`docs/planning/2026-08-19_wise-api-full-implementation-plan.md` (row 5).

[ ] Add `GetMe` / `GetUser` — `GET /me`, `GET /users/{userId}`. Self-contained reads.

[ ] Add `GetStatement` with format parameter — CSV/PDF/XLSX (SDK consumes
`statement.json` only today).

[ ] Add webhook signature verification helper — no REST calls; high value,
self-contained.

[ ] Add balances expansion — `CreateBalance`
(`POST /v4/profiles/{profileId}/balances`), direct `GetBalance`
(`GET /v4/profiles/{profileId}/balances/{balanceId}`, would replace the
client-side scan in `balances.go:51`), `GetTotalFunds`
(`GET /v1/profiles/{profileId}/total-funds/{currency}`).

[ ] Add MCA / account details — `GetBankAccountDetails`, `GetMultiCurrencyAccount`.

[ ] Add `ListCurrencies` — `GET /v1/currencies`.

[ ] Add per-request correlation ID override — currently `WithCorrelationID` sets
a client-wide header (`options.go:73`). Allow per-call override via context or
request struct for request-level tracing. Note: `options.go:72` doc comment
already references the unimplemented `WithRequestCorrelationID`.

[ ] Add mTLS endpoint support — Wise documents `api-mtls.wise.com` and
`api-mtls.wise-sandbox.com`. Add `WithMTLS(tls.Config)` option or document the
`WithHTTPClient` + custom Transport pattern more explicitly.

## P4 — Tooling & quality

[ ] Extract `vendorHash` from `flake.nix` into `vendorHash.nix` — BuildFlow
nix-checker flags it; cleaner diffs when dependencies change
(`flake.nix:118`). Source: status report 18-15 (f.6, e.5).

[ ] Add godoc examples for `CancelTransfer`, `GetDeliveryEstimate`,
`ValidateTransferRequirements`, and the expanded `Quote` fields
(`example_test.go` currently only covers `Currency`/`Money`). Source: status
report 18-15 (f.7, e.8).

[ ] Triage `govulncheck` findings — BuildFlow reports stdlib vulns
(GO-2026-6218 `resolvePath` in `net/url`, +31 more); likely toolchain-fixable,
currently unassessed. Source: status report 18-15 (e.9).

[ ] Add Cachix binary cache to the `nix:` CI job — `nix flake check` no longer
uses `--no-build`, so the full sandboxed test runs. Without a binary cache,
building `go_1_26` from nixpkgs source takes 15+ minutes. Add
`cachix/cachix-action` with a public cache for `nixpkgs` to keep CI under
5 minutes (only `cachix/install-nix-action` is wired today:
`.github/workflows/ci.yml:107`).
