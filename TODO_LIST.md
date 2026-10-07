# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).
Completed work lives in [CHANGELOG.md](CHANGELOG.md), never here.

## P1 — Release readiness

[ ] Run the sandbox live-verification pass (`WISE_SANDBOX_API_KEY`) against the
five spec-vs-live assumptions settled by the conformance work: statement
`type` COMPACT/FLAT, `X-idempotence-uuid` on balance creation, statement
`details.type` enum, `/v1/rates` array shape, text/plain funding errors.
Source: docs/planning/2026-09-16_16-54_pareto-plan-openapi-conformance-green.md, 2026-09-16.

[x] Decide and execute the quarterly-surface 2026Q3 → 2026Q4 rollover.
**DONE 2026-10-05 (split execution, evidence-based):** unauthenticated live probes
against api.wise.com showed `/2026Q4/profiles/{id}/subscriptions` answering 401
(surface live) while a bogus `/9999Q1/...` 404s — so the webhook subscription CRUD
was flipped to `2026Q4` (`webhookSubscriptionsAPIVersion` in `client.go`; tests,
README, and FEATURES updated; conformance gate strips any `[0-9]{4}Q[1-4]` prefix
so it passes either way — the live probe is the real gate).

[ ] Flip `ottAPIVersion` 2026Q3 → 2026Q4 once the OTT surface verifies live.
The OTT surface is **probe-blind**: unauthenticated requests to `/one-time-token/*`
404 on every prefix including the known-good 2026Q3, so Q4 cannot be verified
without credentials. Procedure: with `WISE_SANDBOX_API_KEY` (pairs with the sandbox
pass above), probe `GET /2026Q4/one-time-token/status` for a non-404, change
`ottAPIVersion` in `client.go`, run `go test ./...`, commit. Probes 2026-10-05.

[ ] Publish the GitHub Release objects for **v0.10.0 and v0.11.0** — both tags
exist on origin and are served by the module proxy, but `gh release list` still
shows v0.9.0 as Latest. Notes for BOTH releases are drafted:
`docs/releases/v0.10.0-release-notes.md` and
`docs/releases/v0.11.0-release-notes.md` (cut from the CHANGELOG `[0.11.0]`
section 2026-10-07; describes the tag as shipped, including its pre-rollover
`2026Q3` subscription surface). Remaining: two `gh release create` calls with
the note files. Source: `docs/status/archived/2026-09-13_15-41_pareto-tail-resume-24x-25x-status.md` §b.4, re-verified
2026-09-16.
**BLOCKED: needs the user's approval to publish releases.**

[x] Pin the CI-installed tools — DONE 2026-10-05: `gofumpt` v0.12.0 and
`govulncheck` (golang.org/x/vuln) v1.8.0 pinned via `GOFUMPT_VERSION` /
`GOVULNCHECK_VERSION` env vars in `ci.yml`; the `apidiff` flake app pins
gorelease to `golang.org/x/exp@v0.0.0-20261005173118-76772065c9b0`. No `@latest`
remains in workflows or flake apps (comments excepted).

[x] Add an erraudit CI gate — DONE 2026-10-05 (new item born from the erraudit
pass below): `ci.yml` has an `erraudit` job running the curated config. It is
WARN-not-fail until the private module can be fetched. Carries one user-gated
secret (see P2).

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
v0.11.0 surface); remaining: tag `v1.0.0`.
**BLOCKED: needs the user's explicit approval (tagging is irreversible).**

[ ] Typed recipient `Details` — typed per-corridor structs vs `map[string]string`

- key constants. Carried unanswered through six status reports
  (2026-08-19_17-14 g.2, 18-15 g.2, 09-50 g.3, 20-57 f.5, 22-31 d.1 — all in
  `docs/status/archived/`). The v1.0
  audit confirms the map is the only shape consumers depend on today, so v1.0 can
  freeze the map and add typed accessors later.
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
**BLOCKED: needs the user's approval to push and enable.**

[x] Adopt `go-retry` v0.6.0 in place of the failsafe-go executor — DONE
2026-09-16: ADR 003 accepted and executed. `client.go` runs
`retry.DoWithValue`; `classifyExhaustedRetries` is deleted (the `ErrExhausted`
chain carries the final typed error via `WithCause`); Wise's `Retry-After`
feeds `Config.DelayFunc`, capped at `WithRetry`'s max delay. Gates green:
`go test -race ./...`, `golangci-lint run`, `nix flake check`.

[ ] GOEXPERIMENT ergonomics — pin direnv/home-manager setup so `jsonv2` is set
without relying on `.buildflow.yml` env injection (user-machine work; the
workaround is documented in CONTRIBUTING.md). Carried since
`docs/status/archived/2026-07-23_03-49_buildflow-env-fix-and-golangci-restore.md`.
**BLOCKED: user-machine change.**

## P3 — Quality & tooling (unblocked)

[x] Mirror the 90% coverage floor into the sandboxed `checks.test` checkPhase —
DONE 2026-10-05: the checkPhase now computes the total from `coverage.out` and
fails below 90.0 (mirrors the ci.yml gate verbatim), so the flake check enforces
the floor instead of only measuring.

[x] erraudit pass over the v0.11.0 webhook + OTT code, curated erraudit config,
CI gate, and the samber/oops decision — DONE 2026-10-05:
`erraudit lint ./... --type-aware` and `fix --type-aware` report ZERO findings
repo-wide (webhook + OTT code included), so no migrations were needed. Curated
gate config: `erraudit lint ./... --type-aware --enforce-coded-errors` (the only
applicable opt-in: flags go-error-family constructors called with empty codes;
repo-verified clean). **samber/oops: DECLINED** — the SDK's error contract
(`go-error-family` interfaces + What/Why/Fix messages) already provides the
structure oops would add; adopting it would overlap two classification layers
with zero consumer demand. `--enforce-samber-oops` / `--enforce-go-error-family`
/ `--enforce-deferred-close` / `--enforce-generic-return` are all inapplicable
(stdlib wrapping is a deliberate convention here; reasons in AGENTS.md
Dependencies → erraudit). CI gate: `erraudit` job in ci.yml, WARN-not-fail until
`ERRAUDIT_TOKEN` exists (private module). Extend BuildFlow with an erraudit
provider remains a fleet decision (BuildFlow-repo task, not this repo).

[x] Add `.github/SECURITY.md` — DONE 2026-10-05: private vulnerability reporting
via GitHub security advisories, scope (what to report / what is out of scope),
and a 7-day triage commitment. Linked from CONTRIBUTING.md; the flake
`checks.links` fileset now includes it so the lychee link check can resolve the
relative link.

[x] `doc-verify` hardening — DONE 2026-10-05: count-claim extraction is now a
`check_count` helper that FAILS when a pattern extracts EMPTY (silent-skip class
closed; negative-tested both ways: a mismatched 41→42 claim and a bogus pattern
both fail loudly), and coverage extends beyond AGENTS.md to the FEATURES shipped-
operations claim, the ROADMAP `41 methods` claim, and the audit doc's
`41 \`*Client\` methods` lineage marker.

[x] CONTRIBUTING currency — DONE 2026-10-05: documents `nix run .#doc-verify` /
`nix run .#apidiff` (with what each catches), the 90% coverage floor (both
enforcement sites), the Ginkgo repeat note (`-count=N` false failures → separate
`-count=1` runs), golangci-lint v2.13 alignment with ci.yml, and a SECURITY.md
pointer.

[x] Quality micro-batch — DONE 2026-10-05, all five items:
`GetStatement` now has XLSX coverage (was missing entirely) plus exact-bytes
content-type assertions for PDF and XLSX (binary content types pass through
`getRaw` undecoded); `VerifyWebhookSignature` has
`TestVerifyWebhookSignatureDocumentedVector` — Wise DOES publish a worked
example (sandbox public key + delivery body + signature in
github.com/transferwise/digital-signatures-examples), and the trio is pinned
byte-for-byte and verified by the test;
`ListProfileWebhookSubscriptions` pagination shape verified against the vendored
spec (`webhookProfileSubscriptionList` 200 response is a bare JSON array of
`subscription` — no pagination envelope; the SDK returns the complete set, tests
cover two-subscription and empty cases); godoc example
`ExampleClient_ListProfileWebhookSubscriptions` added (nolint:testableexamples,
like its siblings); `wise.Version` constant added (`"0.11.0"`, mirrors the
latest tag).

## Harvested-and-closed pointer

Everything this list used to carry as `[x]` (webhook CRUD, typed event
decoding, OTT endpoints, `WithUserAgent`, `Profile.UserID`/`PublicID`,
`TestRequireID`, `fetchByID` routing, exhaustruct_v5, govulncheck zero-findings,
benchmarks/fuzz/raw round-trips, test-file split, concurrency test, coverage
gate, templates, ADRs, `doc-verify`/`apidiff` apps, declined ideas) is recorded
in [CHANGELOG.md](CHANGELOG.md) (v0.10.0/v0.11.0 and the 2026-09-13 sessions)
and annotated in the corresponding `docs/status/archived/2026-09-13_*` reports.
