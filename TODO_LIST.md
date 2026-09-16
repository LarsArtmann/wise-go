# TODO List

Short- and mid-term actionable work for wise-go. Each item is bounded, has clear
ownership, and can be completed in one sitting. For long-term vision and raw ideas
see [ROADMAP.md](ROADMAP.md). For shipped features see [FEATURES.md](FEATURES.md).
Completed work lives in [CHANGELOG.md](CHANGELOG.md), never here.

## P1 — Release readiness

[ ] Publish the GitHub Release objects for **v0.10.0 and v0.11.0** — both tags
exist on origin and are served by the module proxy, but `gh release list` still
shows v0.9.0 as Latest. v0.10.0 notes are drafted at
`docs/releases/v0.10.0-release-notes.md`; cut v0.11.0 notes from the CHANGELOG
`[0.11.0]` section. Source: `docs/status/archived/2026-09-13_15-41_pareto-tail-resume-24x-25x-status.md` §b.4, re-verified
2026-09-16.
**BLOCKED: needs the user's approval to publish releases.**

[ ] Pin the CI-installed tools — `gofumpt` (`ci.yml:92`) and `govulncheck`
(`ci.yml:113`) install with `@latest` (non-reproducible), and the `apidiff`
flake app runs `gorelease@latest`. Pin all three to fixed versions. Sources:
`docs/status/archived/2026-09-13_14-16_pareto-full-execution-status.md` §f.13–14, `_15-41` §f.7; verified still
`@latest` 2026-09-16.

## P2 — User-gated

[ ] Add credentialed Wise sandbox integration tests — the workflow
(`.github/workflows/sandbox-live.yml`, active, dispatch-gated) and test skeleton
(`sandbox_live_test.go`) are key-drop-ready; only a first recorded run against
`api.wise-sandbox.com` is missing, plus a CHANGELOG entry once a run succeeds.
Carried since `docs/status/archived/2026-08-08_05-15_wise-sandbox-integration-status.md`.
**BLOCKED: needs a sandbox API key (`WISE_SANDBOX_API_KEY`) from the user.**

[ ] Lock the public API at v1.0 — audit is green (re-audit + growth lineage in
`docs/reviews/2026-08-21_v1.0-api-audit.md`, now covering the 41-method
v0.11.0 surface); remaining: tag `v1.0.0`.
**BLOCKED: needs the user's explicit approval (tagging is irreversible).**

[ ] Typed recipient `Details` — typed per-corridor structs vs `map[string]string`
+ key constants. Carried unanswered through six status reports
(2026-08-19_17-14 g.2, 18-15 g.2, 09-50 g.3, 20-57 f.5, 22-31 d.1 — all in
`docs/status/archived/`). The v1.0
audit confirms the map is the only shape consumers depend on today, so v1.0 can
freeze the map and add typed accessors later.
**BLOCKED: needs the user's design decision.**

[ ] Set the `CACHIX_AUTH_TOKEN` secret (and confirm the `larsartmann` cache
exists) — the CI cachix step (pinned to verified v15 commit `ad2ddac`) is
`continue-on-error: true` until then.
**BLOCKED: needs the user's cachix token.**

[ ] Re-enable CI on GitHub — the auth blocker is RESOLVED (2026-09-13): all
flake inputs moved from `git+ssh` to public `github:` URLs pinned by rev, so CI
needs no secrets; the workflow file is refreshed (`GOLANGCI_LINT_VERSION`
v2.13, no-auth nix job, 90% coverage gate). Remaining: push master, re-enable
the workflow server-side (`gh workflow enable ci`), and watch the first run.
Until then the coverage badge stays frozen at its last CI-measured value.
**BLOCKED: needs the user's approval to push and enable.**

[ ] Adopt `go-retry` v0.4.0 in place of the failsafe-go executor — BLOCKED on
accepting ADR 003 (`docs/adr/003-retry-executor-go-retry-override.md`,
Proposed). Then: swap the `failsafe-go` retry executor in `client.go` for
`github.com/larsartmann/go-retry` `retry.Do`; delete `classifyExhaustedRetries`
(go-retry's exhaustion error carries the final typed error via `WithCause`,
verified by probe 2026-08-21); feed Wise's `Retry-After`
(`RateLimitError.RetryAfter`, including the HTTP-date form) through
`Config.DelayFunc`, which failsafe-go's policy cannot express.
**BLOCKED: needs the user's acceptance of ADR 003.**

[ ] GOEXPERIMENT ergonomics — pin direnv/home-manager setup so `jsonv2` is set
without relying on `.buildflow.yml` env injection (user-machine work; the
workaround is documented in CONTRIBUTING.md). Carried since
`docs/status/archived/2026-07-23_03-49_buildflow-env-fix-and-golangci-restore.md`.
**BLOCKED: user-machine change.**

## P3 — Quality & tooling (unblocked)

[ ] Mirror the 90% coverage floor into the sandboxed `checks.test` checkPhase —
today the threshold lives only in `ci.yml` (disabled on GitHub), so the flake
check measures coverage but enforces no floor. Source:
`docs/status/archived/2026-09-13_15-41_pareto-tail-resume-24x-25x-status.md` §b.3/§f.6.

[ ] erraudit pass over the v0.11.0 webhook + OTT code, then a curated erraudit
config (drop the inapplicable `--enforce-samber-oops`) and a CI/buildflow gate.
The samber/oops adopt-or-decline decision rides along (user may close it).
Sources: `docs/status/archived/2026-08-08_12-27_erraudit-review-and-error-context-improvements.md` §f.7–13,
`_14-16` (= `docs/status/archived/2026-09-13_14-16_pareto-full-execution-status.md`) §f.19.

[ ] Add `.github/SECURITY.md` — a published SDK should carry a vulnerability
reporting path. Source: `docs/status/archived/2026-09-13_15-41_pareto-tail-resume-24x-25x-status.md` §f.35.

[ ] `doc-verify` hardening — fail loudly when a count-claim pattern extracts
EMPTY (kill the silent-skip class for good), and extend coverage to the
ROADMAP/audit-doc count claims. Source: `docs/status/archived/2026-09-13_15-41_pareto-tail-resume-24x-25x-status.md` §d.1/§f.11.

[ ] CONTRIBUTING currency — mention `nix run .#apidiff` / `.#doc-verify` and
the 90% gate; add the Ginkgo repeat note (`go test -count=N` is a false failure;
use separate `-count=1` runs). Sources: `docs/status/archived/2026-09-13_15-41_pareto-tail-resume-24x-25x-status.md` §b.6/§f.12/§f.15.

[ ] Quality micro-batch (each ≤30 min; sources: `_20-57` = `docs/status/archived/2026-08-21_20-57_execution-session-self-review.md` §f.30–31/§f.33,
`_14-16` §f.23–25/§f.35/§f.44, v0.50-retro §f.11):
`GetStatement` PDF/XLSX content-type assertions; `VerifyWebhookSignature` test
against Wise's documented example signature (if published); verify
`ListProfileWebhookSubscriptions` pagination shape against the spec (single
response assumed); godoc example for `ListProfileWebhookSubscriptions` (only
create has one); `wise.Version` constant.

## Harvested-and-closed pointer

Everything this list used to carry as `[x]` (webhook CRUD, typed event
decoding, OTT endpoints, `WithUserAgent`, `Profile.UserID`/`PublicID`,
`TestRequireID`, `fetchByID` routing, exhaustruct_v5, govulncheck zero-findings,
benchmarks/fuzz/raw round-trips, test-file split, concurrency test, coverage
gate, templates, ADRs, `doc-verify`/`apidiff` apps, declined ideas) is recorded
in [CHANGELOG.md](CHANGELOG.md) (v0.10.0/v0.11.0 and the 2026-09-13 sessions)
and annotated in the corresponding `docs/status/archived/2026-09-13_*` reports.
