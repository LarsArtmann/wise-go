# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).

## P1 — Release readiness

[x] Tag `v0.10.0` — RESOLVED 2026-09-13 (between sessions): the tag exists locally
and on origin (the module proxy serves v0.10.0; gorelease uses it as
`-base=latest`). What remains is the GitHub **Release object** (`gh release
list` still shows v0.9.0 as Latest); the notes are drafted at
`docs/releases/v0.10.0-release-notes.md` — publish on approval.
**BLOCKED: needs the user's approval to publish the release.**

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

[x] Record the go-retry override rationale — DONE 2026-09-13 as ADR 003
(`docs/adr/003-retry-executor-go-retry-override.md`, status Proposed): the
in-house library overrides the how-to-golang failsafe-go mandate for this repo
(maintained-by-owner, zero-dep, `Retry-After`-aware via `Config.DelayFunc`).

[ ] Adopt `go-retry` v0.4.0 in place of the failsafe-go executor — BLOCKED on
accepting ADR 003. Then: swap the
`failsafe-go` retry executor in `client.go` for
`github.com/larsartmann/go-retry` `retry.Do`; delete
`classifyExhaustedRetries` (go-retry's exhaustion error carries the final
typed error via `WithCause` and `errors.Is`/`errors.AsType` traverse to it —
verified by probe 2026-08-21, no shim needed); feed Wise's `Retry-After`
(`RateLimitError.RetryAfter`, including the HTTP-date form) through
`Config.DelayFunc`, which failsafe-go's policy cannot express.

[ ] Re-enable CI on GitHub — the auth blocker is RESOLVED (2026-09-13): all
flake inputs moved from `git+ssh` to public `github:` URLs pinned by rev, so CI
needs no secrets; the workflow file is refreshed (`GOLANGCI_LINT_VERSION`
v2.13, no-auth nix job, 90% coverage gate). Remaining: push master, re-enable
the workflow server-side (`gh workflow enable ci`), and watch the first run.
Until then the coverage badge stays frozen at its last CI-measured value.
**BLOCKED: needs the user's approval to push and enable.**

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

[x] Quality long-tail micro-batch — DONE 2026-09-13 with two recorded declines:
`classifyTransactionType` now takes cents (`totalCents int64`); `wiseDateFormat`
constant extracted (`users.go`); `WithUserAgent` shipped + tested;
`Profile.UserID`/`PublicID` surfaced from the wire; `exhaustruct_v5` migration
done (CI pin bumped to v2.13). DECLINED: `fmt.Stringer` on the public enums
(every one is a plain string type — `String()` would print exactly what fmt
already prints; 20 no-op methods are API surface without information) and
`errorfamily.RegisterClassification` (that API maps THIRD-PARTY sentinel
errors; wise-go's six error types implement the Classified interface
directly, which is go-error-family's prescribed path for owned errors).

[x] Test-infra batch — DONE 2026-09-13: client concurrent-safety test (shared
client, 16 parallel requests, per-request correlation-ID isolation, `-race`);
`internal_test.go` split by domain into `errors_test.go` (error taxonomy and
classification) + `helpers_test.go` (timestamps/money/classifier) — all 201
tests preserved; 90% coverage gate in `ci.yml` (measured 91.4%);
`nix run .#apidiff` (gorelease vs latest tag: current delta is all-additive,
suggests v0.11.0).

[x] Contributor/ops batch — DONE 2026-09-13: bug + feature issue templates and
PR template (`.github/`); ADR 001 (Money, no arithmetic) + ADR 002 (flat
package + `internal/raw` boundary) + ADR 003 (go-retry override rationale,
Proposed) in `docs/adr/`; `nix run .#doc-verify` (lychee offline links + godoc
render + count-claims freshness — the gate that would have caught the
"33 methods" drift); stale `reports/` artifacts (jscpd report, old
coverage.out) trashed.

[ ] GOEXPERIMENT ergonomics — pin direnv/home-manager setup so `jsonv2` is set
without relying on `.buildflow.yml` env injection (user-machine work; the
workaround is documented in CONTRIBUTING.md).
