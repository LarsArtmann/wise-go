# Status Report — wise-go Session 2026-07-18 (Evening)

> Generated 2026-07-18 20:21 · Session scope: execute paste.txt workflow (commit, Pareto plan, push) + recover from buildflow Verschlimmbesserung

---

## Session Summary

| Metric                                | Value                                          |
| ------------------------------------- | ---------------------------------------------- |
| Commits pushed                        | 3 (`1140068`, `4f0604c`, `afaf73d`, `a899a5e`) |
| Files changed                         | ~25                                            |
| Build / Vet / Test / Lint             | ALL GREEN                                      |
| `nix flake check` (full, with builds) | PASS                                           |
| `go test -race`                       | PASS (2.46s)                                   |
| Coverage                              | 94.8%                                          |
| Lint issues                           | 0                                              |
| Verschlimmbesserungs recovered from   | 1 (buildflow auto-configure)                   |

---

## a) FULLY DONE ✅

### Committed and pushed

| Commit    | What                                                                                                                                                                                                                                                                                              | Verification                                        |
| --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------- |
| `1140068` | flake.nix: switched from broken `stdenvNoCC.mkDerivation` to `buildGoModule` with correct `vendorHash`. Full `nix flake check` now passes end-to-end (format + sandboxed test).                                                                                                                   | `nix flake check` → all checks passed               |
| `4f0604c` | Three zero-risk improvements: typed `InvestmentState` + generic `parseEnum[T]` helper (eliminates duplicated switch-parsers), exported `DetailType*` constants (API discoverability), `Doer` interface in `WithHTTPClient` (backward-compatible mocking unlock).                                  | build/vet/test/lint all green                       |
| `afaf73d` | CHANGELOG `[Unreleased]` populated, AGENTS.md updated with nix workflow + 3 new gotchas, comprehensive Pareto-driven improvement plan at `docs/planning/` with mermaid execution graph.                                                                                                           | —                                                   |
| `a899a5e` | Dep upgrade recovery: go-branded-id v0.3.2 + go-error-family v0.7.0 (both now use `encoding/json/v2`), wired `GOEXPERIMENT=jsonv2` into flake devShells + buildGoModule checkPhase, updated vendorHash, restored curated `.golangci.yml` after buildflow replaced it with 40+ irrelevant linters. | build/vet/test -race/lint/nix flake check all green |

### Skills completed (from earlier session, now committed)

All 13 skills from the prior session are committed and pushed: code-quality-scan, deduplicate-code, data-model-review, naming-review, full-code-review, architecture-review, architecture-visualization, go-modularize, nix-flake-migration, docs-health, update-old-docs, copywriting, frontend-design.

---

## b) PARTIALLY DONE ⚠️

| Item                       | What's done                                                                                                                   | What's missing                                                                                                                     |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| **Pareto plan execution**  | Plan written with mermaid graph at `docs/planning/`; Phase 1 (1%→51%) and Phase 2 (4%→64%) items C1-C3 and D1-D5 all executed | Phase 3 (20%→80%) items E1-E10 not started; Phase 4-5 (other 20%) not started                                                      |
| **CI integration for nix** | `flake.nix` works locally, `nix flake check` passes                                                                           | No `nix:` job added to `.github/workflows/ci.yml` yet                                                                              |
| **GOEXPERIMENT in CI**     | Set in flake devShells + buildGoModule                                                                                        | NOT set in `.github/workflows/ci.yml` — GitHub Actions will break on next push because the dep upgrade needs `GOEXPERIMENT=jsonv2` |

---

## c) NOT STARTED ❌

- `.github/workflows/ci.yml` does NOT set `GOEXPERIMENT=jsonv2` — **CI WILL BREAK on the next push** because go-branded-id v0.3.2 and go-error-family v0.7.0 require it
- No `nix` CI job added
- README sections for mocking and middleware (E1, E2)
- `govulncheck` not run locally (blocked by security policy on `go install`)
- `CONTRIBUTING.md` not reviewed
- `EndOfStatementBalance` not exposed
- Retry-After not wired into failsafe-go backoff
- All v0.3.0 breaking changes (Money/Currency, enum normalization, etc.)

---

## d) TOTALLY FUCKED UP 💥

### 1. buildflow auto-configure replaced the curated .golangci.yml — CAUGHT AND FIXED

**What happened:** During pre-commit hook execution, buildflow's `golangci-lint-auto-configure` step replaced the project's carefully curated linter list (63 linters, tuned over months) with a generic "enable everything" config containing 100+ linters. This added irrelevant database linters (`arangolint`, `clickhouselint`, `sqlclosecheck` — no DB in this project), a `depguard` config that blocked our own dependencies (`go-branded-id`, `go-error-family`, `failsafe-go`), `varnamelen` flagging idiomatic Go HTTP handler parameters (`w` for ResponseWriter, `r` for Request), and `mnd` flagging `100` in cents conversion and `500` in HTTP status checks. Result: 36 false positive issues.

**How I caught it:** Ran `golangci-lint run` after noticing the unexpected `.golangci.yml` diff. Saw the 36 issues. Read the diff. Understood it was a Verschlimmbesserung.

**How I fixed it:** Restored the original curated linter list, kept the cosmetic YAML 4-space reformatting. Verified: 0 issues.

**Lesson:** `buildflow --fix` / `buildflow auto-configure` is dangerous for projects with curated linter configs. Added a warning to AGENTS.md.

### 2. Dep upgrade broke the build — CAUGHT AND FIXED

**What happened:** buildflow's `go-auto-upgrade` bumped `go-branded-id` v0.3.1→v0.3.2 and `go-error-family` v0.6.1→v0.7.0. Both new versions import `encoding/json/v2`, which requires `GOEXPERIMENT=jsonv2` to build. Without it: `build constraints exclude all Go files in encoding/json/v2`.

**How I caught it:** Ran `go build ./...` after noticing the unexpected `go.mod` diff. Saw the build error.

**How I fixed it:** Wired `GOEXPERIMENT=jsonv2` into flake.nix devShells, buildGoModule checkPhase, and verified all gates pass.

### 3. CI WILL BREAK — NOT YET FIXED ⚠️

**What I missed:** `.github/workflows/ci.yml` does not set `GOEXPERIMENT=jsonv2`. The next CI run on push will fail with the same `encoding/json/v2` build constraint error I just fixed locally. **This is the most urgent unfixed issue.**

### 4. I didn't catch that `nix flake check` was broken on first attempt

**What happened:** My initial `flake.nix` used `stdenvNoCC.mkDerivation` for the test check, which failed inside the Nix sandbox because Go modules can't be fetched offline and `git` is not available. I punted this to the user in the previous session saying "skipped here to avoid 5+ minute first-time Go-module fetch." In reality it was a real bug — the derivation was structurally broken, not just slow.

**How I fixed it:** Switched to `buildGoModule` with the correct `vendorHash`. Full `nix flake check` now passes.

---

## e) WHAT WE SHOULD IMPROVE 🎯

| #   | Improvement                                                                                                     | Urgency                                        |
| --- | --------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| 1   | **Fix CI: add `GOEXPERIMENT=jsonv2` to ci.yml**                                                                 | 🔴 CRITICAL — CI will break on next push       |
| 2   | Pin `GOEXPERIMENT=jsonv2` globally for the project (`.envrc`, Makefile, or shell rc)                            | High                                           |
| 3   | Consider whether the dep upgrade to v0.3.2/v0.7.0 is worth the jsonv2 requirement, or pin back to v0.3.1/v0.6.1 | Medium — depends on what v0.3.2/v0.7.0 give us |
| 4   | Add `buildflow auto-configure` to the list of things to never run blindly                                       | Medium                                         |
| 5   | Continue executing the Pareto plan Phase 3+ items (E1-E10)                                                      | Medium                                         |
| 6   | Add nix CI job to ci.yml                                                                                        | Medium                                         |

---

## f) UP TO 50 THINGS TO GET DONE NEXT 📋

### P0 — CRITICAL (CI is about to break)

| #   | Task                                                                         | Effort |
| --- | ---------------------------------------------------------------------------- | ------ |
| 1   | Add `env: GOEXPERIMENT: jsonv2` to `.github/workflows/ci.yml` all three jobs | 5 min  |
| 2   | Verify CI passes after the fix                                               | 10 min |

### P1 — High urgency (this session's unfinished work)

| #   | Task                                                                  | Effort |
| --- | --------------------------------------------------------------------- | ------ |
| 3   | Add `nix` job to ci.yml (cachix/install-nix-action + nix flake check) | 30 min |
| 4   | Add README "Mocking the client" section                               | 10 min |
| 5   | Add README "Request middleware via WithHTTPClient" section            | 10 min |
| 6   | Review `CONTRIBUTING.md` for drift                                    | 15 min |
| 7   | Expose `EndOfStatementBalance` on `ListTransactionsResponse`          | 30 min |

### P2 — Medium (from the Pareto plan Phase 3)

| #   | Task                                                           | Effort |
| --- | -------------------------------------------------------------- | ------ |
| 8   | Wire `Retry-After` into failsafe-go backoff policy             | 60 min |
| 9   | Register error types with `errorfamily.RegisterClassification` | 30 min |
| 10  | Extract `wiseDateFormat` constant                              | 5 min  |
| 11  | Document `GetBalance` O(n) cost                                | 5 min  |
| 12  | Investigate 5.2% coverage gap                                  | 30 min |
| 13  | Add `WithUserAgent` option                                     | 20 min |
| 14  | Add `WithLogger` option                                        | 45 min |

### P3 — Low (quality polish)

| #   | Task                                           | Effort |
| --- | ---------------------------------------------- | ------ |
| 15  | Add benchmarks for hot paths                   | 45 min |
| 16  | Add `Example_*` test functions for godoc       | 30 min |
| 17  | Add `fmt.Stringer` for enum types              | 30 min |
| 18  | Consider `go:generate` for enum maps           | 60 min |
| 19  | Run `brutal-self-review` skill                 | 45 min |
| 20  | Run `status-report` skill (capstone dashboard) | 30 min |
| 21  | Run `library-deep-dive` on failsafe-go         | 45 min |
| 22  | Run `library-deep-dive` on go-branded-id       | 30 min |
| 23  | Run `library-deep-dive` on go-error-family     | 30 min |
| 24  | Run `bdd-testing` skill                        | 60 min |
| 25  | Add integration test skeleton (live build tag) | 60 min |

### P4 — v0.3.0 (breaking changes)

| #   | Task                                                           | Effort |
| --- | -------------------------------------------------------------- | ------ |
| 26  | Introduce `Money` value object                                 | 4 hr   |
| 27  | Introduce `Currency` branded type                              | 2 hr   |
| 28  | Migrate Transaction/TransactionExchange/BalanceResult to Money | 3 hr   |
| 29  | Normalize enum casing                                          | 1 hr   |
| 30  | Reconcile `TransactionTypeUnknown`                             | 15 min |
| 31  | Drop `Result` suffix or move raw to internal/raw               | 2 hr   |

### P5 — v1.0+

| #   | Task                                         | Effort |
| --- | -------------------------------------------- | ------ |
| 32  | Remove `HasMore`; return `[]Transaction`     | 1 hr   |
| 33  | Move raw types to `internal/raw`             | 2 hr   |
| 34  | Lock public API                              | 2 hr   |
| 35  | Write operations (POST/PATCH/DELETE helpers) | 2 hr   |
| 36  | Transfers resource                           | 1 day  |
| 37  | Recipients resource                          | 1 day  |
| 38  | Quotes resource                              | 1 day  |
| 39  | Webhook signature verification               | 1 day  |

### P6 — Tooling

| #   | Task                                                                  | Effort |
| --- | --------------------------------------------------------------------- | ------ |
| 40  | Investigate buildflow `govalid-generate` failure (needs GOEXPERIMENT) | 30 min |
| 41  | Consider adding `GOEXPERIMENT=jsonv2` to buildflow config             | 15 min |
| 42  | Run `govulncheck` locally (blocked by security policy)                | 5 min  |
| 43  | Consider pinning go-branded-id back to v0.3.1 if jsonv2 is too costly | 15 min |
| 44  | Add `.envrc` support documentation                                    | 10 min |
| 45  | Add nix CI caching (cachix)                                           | 30 min |

### P7 — Documentation

| #   | Task                                                                  | Effort |
| --- | --------------------------------------------------------------------- | ------ |
| 46  | Update ROADMAP.md with jsonv2 dependency note                         | 10 min |
| 47  | Update FEATURES.md with Doer interface + DetailType constants         | 10 min |
| 48  | Document the buildflow auto-configure anti-pattern in CONTRIBUTING.md | 15 min |
| 49  | Add "How to upgrade dependencies safely" to CONTRIBUTING.md           | 20 min |
| 50  | Write ADR for GOEXPERIMENT=jsonv2 decision                            | 30 min |

---

## g) TOP 2 QUESTIONS I CANNOT FIGURE OUT MYSELF 🤔

### 1. Should we keep the dep upgrade (v0.3.2 / v0.7.0) or revert to v0.3.1 / v0.6.1?

The upgrade to `go-branded-id v0.3.2` and `go-error-family v0.7.0` brings `encoding/json/v2` as a transitive dependency, which requires `GOEXPERIMENT=jsonv2` everywhere — locally, in CI, in the nix flake, in the buildflow pre-commit hook. This is a non-trivial operational burden for every developer and every build environment.

**I cannot figure out:** what these new versions actually give us. The diffs show json/v2 adoption, but I don't know if there are bug fixes, new features, or breaking changes that matter to wise-go. I also don't know if you intentionally triggered the `buildflow --fix` that caused this, or if it was accidental.

**Options:**

- **A) Keep the upgrade** — accept the `GOEXPERIMENT=jsonv2` requirement everywhere. Update CI, buildflow, and document it.
- **B) Revert to v0.3.1 / v0.6.1** — restore the working state. The json/v2 experiment is bleeding edge; if wise-go doesn't need it, the operational cost isn't worth it.
- **C) Investigate what changed** — read the changelogs of both libs, then decide.

**What I'll do once answered:** either wire GOEXPERIMENT into CI (option A) or `go get` the old versions (option B). Both are ~10 min.

### 2. Is the `GOEXPERIMENT=jsonv2` requirement acceptable for this project long-term?

`GOEXPERIMENT=jsonv2` is a Go experiment flag. It could be removed, changed, or graduated to default in a future Go release. Building a public SDK on top of an experiment flag means:

- Every consumer of wise-go who runs `go test` must also set `GOEXPERIMENT=jsonv2`
- Or the deps need to be vendored with the experiment baked in
- Or we accept that the SDK is Go-experimental

**I cannot figure out:** whether you consider this an acceptable tradeoff for wise-go, or whether this is a signal that the LarsArtmann ecosystem libs (go-branded-id, go-error-family) are moving too fast for downstream consumers.

**What I'll do once answered:** if acceptable → document in README + CI + CONTRIBUTING. If not → coordinate a rollback across the ecosystem or pin wise-go to the pre-jsonv2 versions.
