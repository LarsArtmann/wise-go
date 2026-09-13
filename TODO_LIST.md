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

[ ] Re-run/extend the v1.0 API audit before tagging `v1.0.0` — the audit
(`docs/reviews/2026-08-21_v1.0-api-audit.md`) covered 31 methods/74 types; the
surface has since grown to 33 methods (+`RefreshQuoteAccountRequirements`,
+`GetTransferReceipt`, +`GetTransferPayoutInfo`). Refresh the inventory and
risk register so the v1.0.0 freeze covers what will actually be frozen.

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
+ key constants. Carried unanswered through five status reports
(2026-08-19_17-14 g.2, 18-15 g.2, 09-50 g.3, 20-57 f.5, 22-31 d). The v1.0 audit
confirms the map is the only shape consumers depend on today, so v1.0 can freeze
the map and add typed accessors later.
**BLOCKED: needs the user's design decision.**

[ ] Set the `CACHIX_AUTH_TOKEN` secret (and confirm the `larsartmann` cache
exists) — the CI cachix step (`cachix-action` pinned to verified v15 commit
`ad2ddac`) is `continue-on-error: true` until then.
**BLOCKED: needs the user's cachix token.**

## P3 — Tier-3 API surface (next feature work)

[ ] Webhook subscription CRUD + test notifications — 9 operations in the OpenAPI
spec across 5 paths (app-level need client-credentials tokens; profile-level fit
today's user-token client). Wire types, public `WebhookSubscription` +
branded `WebhookSubscriptionID`, client methods, BDD, docs. Full breakdown:
`docs/status/2026-08-28_08-37_webhooks-review-and-self-review.md` f.1–13
(plan item Tier-4 #49 says 8 operations; the spec has 9).
**Design question first: ship profile-level only, or pull the client-credentials
token flow (OAuth) into scope?**

[ ] Typed webhook event decoding — `WebhookEvent` envelope + `WebhookEventType`
enum + `ParseWebhookEvent`, tolerant timestamps via `parseWiseTimestamp`;
fixtures + unknown-event forward-compatibility tests. Breakdown:
`docs/status/2026-08-28_08-37_webhooks-review-and-self-review.md` f.14–22.

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

[ ] Re-run `govulncheck` on the go1.26.7 toolchain — the 4 reachable findings
(GO-2026-6218, -6090, -5972, -5026) were stdlib bugs fixed in go1.26.6; the
toolchain has since moved (go.mod `go 1.26.7`), so confirm zero findings.

[ ] Close the dedup-refactor follow-ups
(`docs/status/2026-08-21_23-10_dedup-refactor-self-review.md` f): direct table
test for `requireID` (+ pin its error contract); route `fetchByID` through
`requireID` (currently two zero-ID idioms, `helpers.go:37`); add
`SourceOfFundsOther`/`TransferNature` to `CreateTransferRequest` and delete the
`//nolint:exhaustruct` at `transfers.go:205` (verify field acceptance in
`docs/reviews/wise-api-openapi.json` first); inspect the 6 suppressed art-dupl
groups; decide `requireNonEmpty` for non-ID string validations.

[ ] Error-context polish (`docs/status/2026-08-08_12-27_erraudit-review-and-error-context-improvements.md`):
capture the read error in `body, _ := readBody(resp)` (`client.go:382`); add
raw-input values to `map*` error contexts; `ErrorContext`/`IsRetryable` on
`AuthError`/`NotFoundError`; document the error-context convention in AGENTS.md.

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
