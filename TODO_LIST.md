# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).

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

[ ] Automate the README coverage badge — CI-generated coverage value (no
hand-edited numbers; the 94.8% badge was stale for three versions before the
2026-08-21 audit re-measured 84.2% by hand). Add a measurement note (command +
package basis) until automation lands. Source: 2026-08-21 status report (d.2,
f.1), pareto plan task 20.

[ ] Add a markdown link checker (lychee via flake) to `nix flake check` or
pre-commit — the 2026-08-21 audit hand-grepped only 3 of 8+ docs; ghost
references (like the unverified `wise-api-openapi.json` path incident) are
exactly what a checker catches. Source: 2026-08-21 status report (e.3, f.2),
pareto plan task 21.

[ ] Housekeeping bundle — resolve the 68-vs-77 spec-count discrepancy (derive
the real `It(` count), mention `docs/reviews/wise-api-core-schemas.json` in
AGENTS.md beside the OpenAPI spec (two spec files exist; docs reference one).
Source: 2026-08-21 status report (d.5, f.10, f.11), pareto plan task 23.
