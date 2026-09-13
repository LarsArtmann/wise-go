# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).

## P1 — Release readiness

[ ] Tag `v0.10.0` — `GetTransferReceipt` + `GetTransferPayoutInfo` are shipped in
code and CHANGELOG (`fe896a8`, dated 2026-09-13) but the tag does not exist yet
(`git tag` stops at v0.9.0). Confirm the two new methods are covered by tests,
then tag on approval.
**BLOCKED: needs the user's explicit approval (tagging is irreversible).**

[x] Re-run/extend the v1.0 API audit before tagging `v1.0.0` — DONE 2026-09-13:
the audit doc now carries a re-audit section (33 methods / 92 exported types /
22 const groups via `go doc -all`, growth lineage, godoc pass over the 3 new
methods — one stale `ListBalances` claim found and fixed — and risk-register
items 8–10 for receipt-404 semantics, MT103 nilability, and the two-pass
requirements flow). Remaining for v1.0.0: only the user-gated tag.

## P2 — User-gated

[ ] Add credentialed Wise sandbox integration tests — the workflow
(`.github/workflows/sandbox-live.yml`, active, dispatch-gated) and test skeleton
(`sandbox_live_test.go`) are key-drop-ready; only a first recorded run against
`api.wise-sandbox.com` is missing, plus a CHANGELOG entry once a run succeeds.
**BLOCKED: needs a sandbox API key (`WISE_SANDBOX_API_KEY`) from the user.**

[ ] Lock the public API at v1.0 — audit is green (see P1 re-audit note);
remaining: tag `v1.0.0`.
**BLOCKED: needs the user's explicit approval (tagging is irreversible).**

[ ] Typed recipient `Details` — typed per-corridor structs vs `map[string]string`

- key constants. Carried unanswered through five status reports
  (2026-08-19_17-14 g.2, 18-15 g.2, 09-50 g.3, 20-57 f.5, 22-31 d). The v1.0 audit
  confirms the map is the only shape consumers depend on today, so v1.0 can freeze
  the map and add typed accessors later.
  **BLOCKED: needs the user's design decision.**

[ ] Set the `CACHIX_AUTH_TOKEN` secret (and confirm the `larsartmann` cache
exists) — the CI cachix step (`cachix-action` pinned to verified v15 commit
`ad2ddac`) is `continue-on-error: true` until then.
**BLOCKED: needs the user's cachix token.**

## P3 — Tier-3 API surface (next feature work)

[x] Webhook subscription CRUD (profile level) — DONE 2026-09-13:
`Create/List/Get/DeleteProfileWebhookSubscription` shipped against
`/2026Q3/profiles/{profileId}/subscriptions` with wire types, public
`WebhookSubscription`, branded `WebhookSubscriptionID` (UUID string per spec),
validation, BDD, README, and FEATURES rows. App-level operations
(+`test-notifications`, spec: app-level only) are DEFERRED pending the
client-credentials token decision — see ROADMAP "Application-level webhook
subscriptions (deferred 2026-09-13)". Plan item #49's "8"→9 count corrected.

[x] Typed webhook event decoding — DONE 2026-09-13: `WebhookEvent` envelope +
`WebhookEventType` open enum (33 documented constants) + `ParseWebhookEvent`
(tolerant timestamps, unknown-event passthrough) + typed payload accessors for
`transfers#state-change`, `transfers#payout-failure`, and `balances#credit`
(the plan's deposits#* events are not documented on the live webhook-event
reference — dropped per verification). Unknown-event + corruption tests
included.

## P4 — Tooling & quality

[ ] Adopt `go-retry` v0.4.0 in place of the failsafe-go executor — swap the
`failsafe-go` retry executor in `client.go` for
`github.com/larsartmann/go-retry` `retry.Do`; delete
`classifyExhaustedRetries` (go-retry's exhaustion error carries the final
typed error via `WithCause` and `errors.Is`/`errors.AsType` traverse to it —
verified by probe 2026-08-21, no shim needed); feed Wise's `Retry-After`
(`RateLimitError.RetryAfter`, including the HTTP-date form) through
`Config.DelayFunc`, which failsafe-go's policy cannot express. Before executing,
record why the in-house library overrides the how-to-golang failsafe-go mandate
for this repo (maintained-by-owner, zero-dep, Retry-After-aware).

[ ] Re-enable CI on GitHub — the workflow file is refreshed (commit `14523ae`)
but the workflow is still `disabled_manually` server-side (last run 2026-07-05).
First provide SSH auth for the `git+ssh://` flake inputs (deploy key or
`GITHUB_TOKEN` + `insteadOf`), verify `nix flake check` passes in CI, then
re-enable. Until then the coverage badge stays frozen at its last CI-measured
value.

[x] Re-run `govulncheck` on the go1.26.7 toolchain — DONE 2026-09-13 (govulncheck
1.8.0, `GOEXPERIMENT=jsonv2`): **"No vulnerabilities found."** The 4 reachable
findings (GO-2026-6218, -6090, -5972, -5026) were stdlib bugs fixed in
go1.26.6; the go1.26.7 toolchain clears them all. CI re-runs this in the
`govulncheck` job once the workflow is re-enabled.

[x] Close the dedup-refactor follow-ups — DONE 2026-09-13: `TestRequireID`
pins the contract (`[rejection:wise.<domain>.invalid_request] <field> is
required`, int64 + string rows); `fetchByID` now takes the branded ID and
routes through `requireID` (4 callers updated, message contract unchanged);
`SourceOfFundsOther`/`TransferNature` were NOT added — spec verification
showed the create-transfer schema's details accept exactly the five keys
`CreateTransferRequest` already carries, so the `nolint` at `transfers.go:205`
stays as the spec-verified marker; art-dupl re-inspected (ONE suppressed
group today, the TransferRequirement mirror — the "6 groups" figure was
stale); `requireNonEmpty` declined with rationale in AGENTS.md.

[x] Error-context polish — DONE 2026-09-13: `checkError` surfaces unreadable
response bodies instead of silently empty ones; `map*` errors carry raw wire
values (`total amount %v %q`, `amount %v %q`, `reserved amount %v %q`);
`SCAChallengeError.ErrorContext` exposes the 2FA verdict headers;
`AuthError`/`NotFoundError` keep their promoted contexts (pinned by
`TestErrorContexts`); retryability contract pinned by `TestIsRetryableContract`;
convention documented in AGENTS.md.

[ ] Quality long-tail micro-batch (each ≤30 min, recurring across 8+ old reports):
`classifyTransactionType` takes `amount float64` (`transactions.go:178`) — pass
cents instead; `wiseDateFormat` constant for the inline `"2006-01-02"` layout
(`users.go:139`); `WithUserAgent` option; `fmt.Stringer` for public enums;
`errorfamily.RegisterClassification` call; surface `Profile.UserID`/`PublicID`;
pin the gofumpt action version (`ci.yml` uses `@latest`); benchmarks + fuzz
tests for date/money parsing; split the `wise_test.go` monolith; migrate the
deprecated `exhaustruct` linter to `exhaustruct_v5` (golangci-lint v2.13
deprecation warning, seen 2026-09-13).

[ ] GOEXPERIMENT ergonomics — pin direnv/home-manager setup so `jsonv2` is set
without relying on `.buildflow.yml` env injection (user-machine work; the
workaround is documented in CONTRIBUTING.md).
