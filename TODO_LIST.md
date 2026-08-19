# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).

## P1 — v1.0 release (API lock)

[ ] Add credentialed Wise sandbox integration tests — create a protected or manually triggered live-test workflow using sandbox credentials, cover the supported read-only endpoints, and keep it separate from fork-safe mock tests. The sandbox URL was updated to V2 (`api.wise-sandbox.com`) in this session; tests must verify the new endpoint works. This is required to verify that the SDK actually interoperates with Wise's sandbox rather than only matching mocked responses.

[ ] Lock the public API at v1.0 — formal audit of every exported symbol, godoc
review pass, then tag `v1.0.0`. All v0.5.0 breaking changes are shipped and tested;
the lock is the logical next milestone. Requires explicit approval (tagging is
irreversible). The API surface is now stable: `Money`/`Currency` value objects,
branded IDs, typed enums (`DetailType`, `TransactionType`, `ProfileType`,
`BalanceType`, `InvestmentState`), and the two-layer raw/result split are all
finalized.

## P2 — CI speed

[ ] Add Cachix binary cache to the `nix:` CI job — `nix flake check` no longer
uses `--no-build` (v0.5.0 change), so the full sandboxed test runs. Without a
binary cache, building `go_1_26` from nixpkgs source takes 15+ minutes. Add
`cachix/cachix-action` with a public cache for `nixpkgs` to keep CI under 5 minutes.

## P3 — API surface expansion (from changelog & API reference study)

[x] Add `GetProfile(ctx, ProfileID)` — `GET /v2/profiles/{profileId}`.

[x] Add exchange rates endpoint — `GET /v1/rates`.

[x] Add Quotes API — `POST /v3/quotes`, `POST /v3/profiles/{id}/quotes`,
`GET /v3/profiles/{id}/quotes/{id}`.

[x] Add Recipients API — `GET /v2/accounts`, `GET /v1/accounts/{id}`,
`POST /v1/accounts`.

[x] Add `GetTransfer` and `CreateTransfer` — `GET /v1/transfers/{id}`,
`POST /v1/transfers`.

[x] Add `CancelTransfer` — `PUT /v1/transfers/{id}/cancel`.

[x] Add `GetDeliveryEstimate` — `GET /v1/delivery-estimates/{id}`.

[x] Expand `Quote` with `paymentOptions`, fees, and notices — the quote now
exposes per pay-in/pay-out option fee breakdowns, estimated delivery times,
`RateType`, `ProvidedAmountType`, guaranteed-target-amount flags, and
user-facing notices. (2026-08-19)

[x] Add `ValidateTransferRequirements` — `POST /v1/transfer-requirements`.
(2026-08-19)

[ ] Add per-request correlation ID override — currently `WithCorrelationID` sets
a client-wide header. Allow per-call override via context or request struct for
request-level tracing.

[ ] Add mTLS endpoint support — Wise documents `api-mtls.wise.com` and
`api-mtls.wise-sandbox.com`. Add `WithMTLS(tls.Config)` option or document the
`WithHTTPClient` + custom Transport pattern more explicitly.
