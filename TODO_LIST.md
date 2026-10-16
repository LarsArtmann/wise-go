# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).
Completed work lives in [CHANGELOG.md](CHANGELOG.md), never here.

## P1 — Release readiness

[ ] Execute the remaining SUPERB Pareto plan items — sequenced by
`docs/planning/2026-10-16_21-45_SUPERB-v0.14-release-close-ci-resurrection-and-gate-hardening.md`.
Of the 19:45 plan: M1/M2/M3/M5/M6/M7/M8/M9/M10/M15 ALL shipped 2026-10-16
(typed webhook payloads + originator transfers landed 21:16-21:44;
46-method surface, count claims synced `41b09dc`; art-dupl CI job
scaffolded `cb7b271`). Remaining: M4 live smoke (user key), M12 SCA/OTP
README matrix (plan task M14), M13 remainder (2027Q1 path triage → plan
task M10), M14 smoke-harness decision (plan task M18), plus the v0.14.0
release close (plan tasks M1-M3) and CI resurrection (plan task M4).
Source: transfer-category re-audit
`docs/reviews/2026-10-16_transfer-category-reaudit.md` (annotated 2026-10-16:
deposit-details pay-in already DEMAND_GATED; the re-audit's "banking-partner
invoice" row was a false gap — it IS `GetTransferPayoutInfo`).

[ ] Cut v0.14.0 — staged in `[Unreleased]` (breaking `OneTimeToken`/`OTTAction`
typing, 2027Q1 account-details migration, transfer-category completion,
spec-drift gate, three new endpoints): `nix run .#pre-release`, tag,
`gh release create`, proxy + `go get` verification, apidiff vs v0.13.0
re-check. Consider scratch-verifying bank-sync compiles against master
(replace/GOWORK) BEFORE the tag — the 21:09 session asserted the
challenge→clear→replay flow compiles unchanged by reasoning only.
Source: `docs/status/2026-10-16_21-09_cross-repo-wise-go-audit-sca-typing-session.md` §f13/§f14.

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
`docs/reviews/2026-08-21_v1.0-api-audit.md`, growth lineage maintained through
the current 44-method surface); remaining: tag `v1.0.0`.
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
Add a network `pin-sync` job to ci.yml at the same time (mirroring the
sandbox-live dispatch pattern).
Until then the coverage badge stays frozen at its last CI-measured value.
Source: `docs/status/archived/2026-10-07_03-57_session-status-unblocked-prep-release-readiness.md`.
**BLOCKED: needs the user's approval to push and enable.**

## P3 — Quality & tooling (unblocked)

### Open

[ ] Build an anchor-check gate (`nix run .#anchor-check`): extract every
`*.go:NN` reference from README/FEATURES/TODO_LIST/ROADMAP and assert the
cited line still contains the claimed symbol; run the repo-wide sweep and fix
the drifts it finds. doc-verify gates counts, not line anchors — this class
rots silently (8 dead anchors found 2026-10-16, more re-drifted the same
evening). Source: `docs/status/2026-10-16_20-41_samber-do-premise-review-and-health-findings.md` §f6/§f7.

[ ] Harden `spec_drift.py`: add an extraction-floor guard (fail when the
extracted endpoint count diverges from the expected surface), set
`PYTHONDONTWRITEBYTECODE=1` in the flake app, and triage the 16 added 2027Q1
paths into the FEATURES.md matrix (payment-instruments, settlement/
allocations, saml2, simulation/*, gpi-tracking, cases/files, login/saml2/sso,
plus the `/auth/jose/playground/jwsjwe` method change). Longer arc: rewrite
the regex extractor on stdlib `go/parser` (kills the call-shape-bug class).
Source: `docs/status/2026-10-16_19-48_wise-api-spec-drift-check-session.md` §f.

[ ] Persist negative-scenario tests for the two Python gates
(`scripts/pin_sync.py`, `scripts/spec_drift.py`) — crafted stale-pin /
removed-path / broken-JSON fixtures wired into flake checks; both scripts are
currently verified by manual gauntlets only.

[ ] Add a minimal `.markdownlint.json` (MD013 off) so on-demand markdownlint
runs surface real issues (broken references, duplicate headings) without
fighting the repo's long-line style; dprint stays the formatter.

[ ] Single-source the coverage floor (flake.nix and ci.yml read one value, or
generate ci.yml's from flake.nix) — the parity check guards the split brain;
eliminate it. Source: 2026-10-11 19:20 session §e4/§f5.

[ ] Build a linter-allowlist check: the enabled-linters set in
`.golangci.yml` must match a pinned allowlist, making the curated-list policy
a gate instead of memory (`lll` was auto-re-added twice on 2026-10-16 alone).
Source: `docs/status/2026-10-16_08-52_todo-execution...md` §f18.

[ ] Cut a fresh benchstat baseline at the next release tag — the committed
one predates `BenchmarkParseWiseTimestamp` and ends with a junk
`exit status 1` line (benchstat tolerates both; append-only forbids editing).

[ ] File the two BuildFlow display defects upstream after a
verify-before-filing pass: (a) "9 tools unavailable" persists after installing
the tools and purging the result cache while `buildflow doctor` passes them;
(b) the findings counter reports rendered output lines, not findings (27
"statix warnings" where `statix check .` shows 2). Source: 2026-10-11 19:20
session §f2.

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

[ ] Spec-drift gate posture — decide whether `--fail-on-drift` should be
wired into `nix run .#pre-release` (releases then require a fresh vendored
spec; today vendored-snapshot staleness is report-only) and where spec-drift
recurs while CI is disabled (scheduled local run, buildflow step, or
manual-before-release). Source:
`docs/status/2026-10-16_19-48_wise-api-spec-drift-check-session.md` §g2/§g3.
**BLOCKED: user decision.**

[ ] Flake pinning strategy — keep `?rev=<commit-sha>` inputs (the hermetic
`checks.pin-sync` asserts structural invariants only; the network
`apps.pin-sync` is the real version detector, firing at pre-release time) or
switch to version-tag refs (`github:LarsArtmann/go-retry/v0.7.1`, flake.lock
still pinning the immutable commit) so the hermetic check can catch version
drift offline. Reverses the documented rev-pinned decision — user call.
Source: `docs/status/2026-10-07_08-22_v0.13.0-release-pin-sync-drift-discovery-status.md` §g1.
**BLOCKED: user decision.**

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
