# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).
Completed work lives in [CHANGELOG.md](CHANGELOG.md), never here.

## P1 — Release readiness

[ ] Execute the remaining SUPERB Pareto plan items
(`docs/planning/2026-10-16_19-45_SUPERB-pareto-plan-transfer-truth-close-and-fleet-hygiene.md`;
M1/M2/M3/M6/M7 already shipped 2026-10-16): M5
`GET /transfers/{id}/payments`, M8 US combined receipt, M9 NOC document,
M10 typed `transfers#active-cases`/`transfers#refund` payloads, M12 README
SCA/OTP endpoint matrix, M13 inventory-honesty sweep (FEATURES gaps +
count-claim updates), M14 live-smoke harness, M15 originator transfers.
Source: transfer-category re-audit
`docs/reviews/2026-10-16_transfer-category-reaudit.md` (true category ~15
ops: 9 shipped, 4 planned, deposit-details pay-in already DEMAND_GATED;
the re-audit's "banking-partner invoice" row was a false gap — it IS
`GetTransferPayoutInfo`).

[ ] Run the sandbox live-verification pass (`WISE_SANDBOX_API_KEY`) against the
five spec-vs-live assumptions settled by the conformance work: statement
`type` COMPACT/FLAT, `X-idempotence-uuid` on balance creation, statement
`details.type` enum, `/v1/rates` array shape, text/plain funding errors.
Source: docs/planning/archived/2026-09-16_16-54_pareto-plan-openapi-conformance-green.md, 2026-09-16.

[ ] Flip `ottAPIVersion` 2026Q3 → 2026Q4 once the OTT surface verifies live.
The OTT surface is **probe-blind**: unauthenticated requests to `/one-time-token/*`
404 on every prefix including the known-good 2026Q3, so Q4 cannot be verified
without credentials. Procedure: with `WISE_SANDBOX_API_KEY` (pairs with the sandbox
pass above), probe `GET /2026Q4/one-time-token/status` for a non-404, change
`ottAPIVersion` in `client.go`, run `go test ./...`, commit. Probes 2026-10-05.

## P2 — User-gated

[ ] Add credentialed Wise sandbox integration tests — the workflow
(`.github/workflows/sandbox-live.yml`, active, dispatch-gated) and test skeleton
(`sandbox_live_test.go`) are key-drop-ready; only a first recorded run against
`api.wise-sandbox.com` is missing, plus a CHANGELOG entry once a run succeeds.
Skip path re-verified 2026-10-07 (no key → clean SKIP, suite green).
Carried since `docs/status/archived/2026-08-08_05-15_wise-sandbox-integration-status.md`.
**BLOCKED: needs a sandbox API key (`WISE_SANDBOX_API_KEY`) from the user.**

[ ] Lock the public API at v1.0 — audit is green (re-audit + growth lineage in
`docs/reviews/2026-08-21_v1.0-api-audit.md`, now covering the 41-method
v0.12.0 surface); remaining: tag `v1.0.0`.
**BLOCKED: needs the user's explicit approval (tagging is irreversible).**

[ ] `RequestLog.RetryAfterDelay` API freeze — lock the delay-decision-entry shape
(entry riding the `LogRequest` channel, `Method`/`URL` empty, honored wait as a
`time.Duration`) for v1.0, or grow a dedicated observer interface. Evidence:
`nix run .#apidiff` vs v0.12.0 shows it is the ONLY compatible delta (gorelease
suggests v0.13.0). Source: `docs/status/2026-10-07_07-09_p3-resume-execution-prerelease-green-status.md` §g2.
**BLOCKED: needs the user's design decision.**

[ ] Typed recipient `Details` — typed per-corridor structs vs `map[string]string`
key constants. Carried unanswered through seven status reports (2026-08-19_17-14
g.2 through 2026-10-05_14-53 g2 — all under `docs/status/`, several now in `archived/`). The v1.0 audit confirms
the map is the only shape consumers depend on today, so v1.0 can freeze the map
and add typed accessors later.
**BLOCKED: needs the user's design decision.**

[ ] Set the `CACHIX_AUTH_TOKEN` secret (and confirm the `larsartmann` cache
exists) — the CI cachix step (pinned to verified v15 commit `ad2ddac`) is
`continue-on-error: true` until then.
**BLOCKED: needs the user's cachix token.**

[ ] Set the `ERRAUDIT_TOKEN` secret (PAT with read access to the private
`github.com/larsartmann/erraudit` repo — the module is not on the public proxy)
so the erraudit CI job can install the tool; until then that job's install step
fails, the gate is skipped, and the job stays green (`continue-on-error` pattern,
2026-10-05). When the token lands, delete the two `continue-on-error` / `if`
lines in the `erraudit` job to make the gate blocking.
**BLOCKED: needs the user's PAT.**

[ ] Re-enable CI on GitHub — the auth blocker is RESOLVED (2026-09-13): all
flake inputs moved from `git+ssh` to public `github:` URLs pinned by rev, so CI
needs no secrets; the workflow file is refreshed (`GOLANGCI_LINT_VERSION`
v2.13, no-auth nix job, 90% coverage gate). Remaining: push master, re-enable
the workflow server-side (`gh workflow enable ci`), and watch the first run.
Until then the coverage badge stays frozen at its last CI-measured value.
Source: `docs/status/archived/2026-10-07_03-57_session-status-unblocked-prep-release-readiness.md`.
**BLOCKED: needs the user's approval to push and enable.**

## P3 — Quality & tooling (unblocked)

### Open

[ ] Fix `Health` semantics (docs half) — rewrite the `client.go:184` doc
comment and the README line to say what the method IS: a credential-validating
end-to-end readiness probe (authenticated `ListProfiles` through the retry
executor), not a cheap reachability check. Document the typed-error distinction
(a `*RateLimitError` from `Health` proves reachability with throttling) and the
worst-case latency (~4 attempts × 30 s timeout + capped delays, up to ~2 min).
Unblocked; the code-half decision lives under Decisions below. Optionally add a
short README recipe for wiring the client into a consumer-side DI container
health check (adapter pattern, no SDK changes — see
`docs/architecture-understanding/2026-10-16_19-59_samber-do-health-architecture-review.md` §5).
Source: 2026-10-16 samber/do + health architecture review, finding F-1/F-3.

[ ] Consume the committed `benchstat` baseline
(`docs/bench/2026-10-07_v0120_baseline.txt`) in the next perf-touching PR.
Note: the README rerun commands are now execution-verified (six separate
`-count 1` invocations; Ginkgo bans `-count>1`), and benchstat is NOT in the
devShell — install with `go install golang.org/x/perf/cmd/benchstat@latest`.

### Decisions (P3)

[ ] `Health` code-half — after (or with) the docs fix above, decide: keep
`Health` as the retried readiness probe as-is, give it a fail-fast path
(`MaxAttempts=1`-style, health probes should not spend ~2 min in the retry
executor), and/or add a `VerifyCredentials` alias if any consumer depends on
the current retried semantics; consider a `HealthCheck(ctx)` name-alias so
consumer `do.HealthcheckerWithContext` containers satisfy the interface
directly. Also extend the `wise_test.go` `Describe("Health")` context to pin
the typed-error distinction (429 → `*RateLimitError` proves reachability;
bad key → `*AuthError`). A true cheap liveness probe additionally needs
an endpoint choice verified against the live API (no ping/health endpoint is
documented in the vendored snapshot). The docs-half (godoc + README +
FEATURES wording) shipped 2026-10-16; the optional consumer-side DI recipe
(README vs `docs/recipes/`) is also yours to place.
**BLOCKED: user design decision.** Source: 2026-10-16 samber/do + health
architecture review, finding F-1; `docs/architecture-understanding/2026-10-16_19-59_samber-do-health-architecture-review.md`.

[ ] `nix flake check --all-systems` — verified 2026-10-07: aarch64-linux and
aarch64-darwin evaluate clean; x86_64-darwin FAILS EXTERNALLY (nixpkgs-26.11
dropped that system, not a flake bug). Decide: pin a nixpkgs-26.05-darwin
input for darwin checks, or drop x86_64-darwin from the check matrix.
**BLOCKED: user decision.**

[ ] Go 1.27 migration — plan written at
`docs/planning/2026-10-07_go-1.27-migration-plan.md` (4 unlock conditions,
single-commit flip/rollback). Condition-1 probe 2026-10-07: `nix eval
nixpkgs#go.version` = 1.26.8 → not met. Execution is blocked on those
conditions (nixpkgs stable 1.27, in-house deps verified on 1.27, erraudit local install,
buildflow skip_steps removal). Source:
`docs/status/2026-10-07_03-45_buildflow-red-to-green-repair-status.md` §g1.
