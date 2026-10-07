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

[ ] Typed recipient `Details` — typed per-corridor structs vs `map[string]string`
key constants. Carried unanswered through seven status reports (2026-08-19_17-14
g.2 through 2026-10-05_14-53 g2 — all in `docs/status/`). The v1.0 audit confirms
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
**BLOCKED: needs the user's approval to push and enable.**

[ ] GOEXPERIMENT ergonomics — pin direnv/home-manager setup so `jsonv2` is set
without relying on `.buildflow.yml` env injection (user-machine work; the
workaround is documented in CONTRIBUTING.md). Carried since
`docs/status/archived/2026-07-23_03-49_buildflow-env-fix-and-golangci-restore.md`.
**BLOCKED: user-machine change.**

## P3 — Quality & tooling (unblocked)

[ ] Wire `md-go-validator` into `nix flake check` (or `.buildflow.yml`) as a
fenced-Go snippet parse gate — README block #27's invalid shape (fixed 2026-09-29)
was found by hand; nothing currently fails if a future edit breaks a snippet.
Source: `docs/status/2026-09-29_07-37_README-webhook-snippet-validation-fix.md` §f.

[ ] Extend `doc-verify` count-claims to `docs/releases/*.md` (the release-notes
method-count lines are a drift blind spot — AGENTS/FEATURES/ROADMAP/audit are
gated, `docs/releases/` is not). Source:
`docs/status/2026-10-07_03-57_session-status-unblocked-prep-release-readiness.md` §d6/f#8.

[ ] Build coverage headroom: 90.1% total against a 90.0 floor is a hair-trigger;
raise to ~92% (cheapest wins: JSON-error-path mappers, `checkError` body-read
branches) before the next feature PR. Source:
`docs/status/2026-10-06_14-16_q4-flip-execution-and-gates-status.md` §b3/e4.

[ ] Root-cause the `golangci_lint_ls` phantom on `webhooks.go:142` ("File is not
properly formatted") that recurs while CLI `golangci-lint run` is clean — diff the
LSP's golangci invocation (config, working dir, GOEXPERIMENT) against the CLI.
Source: `docs/status/2026-10-06_14-16_q4-flip-execution-and-gates-status.md` §d4.

[ ] Write the quarterly-surface rollover ritual (probe → flip → repo-wide straggler
grep → changelog → doc-verify) into CONTRIBUTING — the Q4 flip proved a value-flip
is not done until `grep -rn <old-value>` is clean. Source:
`docs/status/2026-10-07_03-57_session-status-unblocked-prep-release-readiness.md` f#36.

[ ] File the three evidenced BuildFlow-repo defects (embedded `erraudit` false
positives at HEAD; `nix run .#reinstall` not switching the profile; `file-size-check`
scanning 0 files in root-package layouts) — evidence is captured in AGENTS.md.
Source: `docs/status/2026-10-07_04-20_dual-phase-repair-and-tool-reevaluation-status.md` §b1.

[ ] `art-dupl` enforced-gate decision — accept the suppressed baseline as policy,
or enforce an exit-code gate (`art-dupl -t 2 --type-aware`); records the choice in
AGENTS.md. Source: `docs/status/2026-09-27_23-42_dedup-pass2-gomod-flipflop-rootcause-status.md` §c4.

[ ] erraudit class-wide policy — fix the 29 `context_loss`/`ignored` advisories or
suppress the rules with rationale; per-finding drift is the worst option. Source:
`docs/status/2026-09-27_23-42_dedup-pass2-gomod-flipflop-rootcause-status.md` §c11.

[ ] Go 1.27 migration plan — decide the condition/date for moving the flake
toolchain + `GOTOOLCHAIN` policy off the 1.26 pin (erraudit needs ≥ 1.27); until
then the modernizer's 1.27-syntax pushes are reverted per incident. Source:
`docs/status/2026-10-07_03-45_buildflow-red-to-green-repair-status.md` §g1.

[ ] Add a CAMT (`.xml`) `GetStatement` test — the only statement format still
without direct coverage (the conformance suite covers `json` + the exempt file
formats); add exact content-type assertions mirroring PDF/XLSX. Source:
`docs/status/2026-10-05_21-47_pareto-todo-execution-status.md` §f7.

[ ] Commit a `benchstat` baseline file so benchmark runs have a comparison point
(the hot mappers/parsers are in `bench_test.go`). Source:
`docs/status/2026-10-05_21-47_pareto-todo-execution-status.md` §f.

[ ] Add a fuzz target / corpus entry for `VerifyWebhookSignature` +
`ParseWebhookPublicKey` (webhook funcs were added after the original fuzz round)
and bench `VerifyWebhookSignature` (hot path for every delivery). Source:
`docs/status/2026-10-05_21-47_pareto-todo-execution-status.md` §f11–12.

[ ] Add a `nix run .#pre-release` flake app chaining the local gates (build, vet,
race test, lint, `nix flake check`, `doc-verify`, `apidiff`, dirty-tree check) so
a release relies on one command instead of session discipline. Source:
`docs/status/2026-10-07_04-57_v0120-release-session-status.md` §e3/f10.

[ ] Compare `docs/releases/*.md` against each tag before `gh release create` — a
release-notes check for split code spans and relative links (a reflowed span
shipped broken in the v0.12.0 release body). Source:
`docs/status/2026-10-07_04-57_v0120-release-session-status.md` §d1/e5.

[ ] Make the spec-conformance coverage guard assert its floors under `go test
-shuffle` (today it SKIPs if shuffled before the recorder, dorming the
vacuous-pass protection). Source:
`docs/status/2026-10-07_04-57_v0120-release-session-status.md` §e4/f27.

[ ] Add a `go-retry` `Retry-After`-honored log line (observability — callers
cannot currently see that Wise's hint steered a delay). Source:
`docs/status/2026-10-07_04-57_v0120-release-session-status.md` §f30.

[ ] Fuzz `decodeExchangeRates` (array / single-object / empty / corrupt) and
bench the rates array-decode path. Source:
`docs/status/2026-10-07_04-57_v0120-release-session-status.md` §f28–29.

[ ] Add `nix flake check --all-systems` (aarch64/darwin fleet coverage). Source:
`docs/status/2026-10-07_04-57_v0120-release-session-status.md` §f31.

## Harvested-and-closed pointer

Everything this list used to carry — including the v0.12.0 cycle's webhook
subscription CRUD, typed event decoding, OTT endpoints, `WithUserAgent`,
`Profile.UserID`/`PublicID`, `TestRequireID`, `fetchByID` routing, exhaustruct_v5,
govulncheck zero-findings, benchmarks/fuzz/raw round-trips, test-file split,
concurrency test, coverage gate, templates, ADRs, `doc-verify`/`apidiff` apps,
CI tool pinning, the `erraudit` CI gate, the 90% coverage floor, the Q4 webhook
rollover, the go-retry v0.6.0 executor, the v0.10.0/v0.11.0/v0.12.0 GitHub
Release objects (published 2026-10-07), and the declined ideas — is recorded in
[CHANGELOG.md](CHANGELOG.md) (`[0.12.0]` and the v0.10.0/v0.11.0 / 2026-09-13
sections) and annotated inline in the corresponding `docs/status/` reports.
