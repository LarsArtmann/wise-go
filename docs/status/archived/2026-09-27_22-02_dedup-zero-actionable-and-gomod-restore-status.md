# Status Report — Dedup to Zero-Actionable + go.mod Restore

> **Resolved (2026-10-07 docs-health pass):** ~~this report's "0 actionable" snapshot did not hold~~ — the resurfaced groups were fixed the same night and dedup is stable at 0 actionable through v0.12.0 (see the superseding `2026-09-27_23-42_dedup-pass2-gomod-flipflop-rootcause-status.md`). The go.mod restore held: `go 1.26` is the settled directive. Open items (art-dupl enforced-gate decision, webhook label typing, erraudit class-wide policy) are routed to TODO_LIST.md / ROADMAP.md.

**Date:** 2026-09-27 22:02 CEST
**Scope:** This session only (per instruction: no unrelated research). Session = art-dupl triage → 3 extractions → 8 accept directives → incidental discovery and fix of an 11-day broken build → AGENTS.md updates → full verification battery.
**Honesty note:** This report doubles as the brutal self-review. Sections (d) and (e) include failures from THIS session, not just the repo's history.

> **Update 2026-09-27 ~23:00 (second pass, same day):** the "0 actionable" snapshot above did not hold. The requireID-trio directives were placed on the enclosing functions' doc comments, which suppresses nothing for body-interior clones (the report's own gotcha in (a)#7) — the trio resurfaced actionable in the next run, and the quotes account-requirements and webhook decode/parse call-site blocks were still duplicated at the call sites. Second pass: moved the trio directives in-body (verified suppressing), extracted `quoteAccountRequirements` (quotes.go) and `decodeWebhookEvent[T]` + `raw.OccurredAtCarrier`/`raw.WebhookOccurredAt` (webhooks.go, internal/raw/webhooks.go) to eliminate groups #2/#3 for real. The go.mod bump also recurred via buildflow's gomod tooling (fleet go-version flipflop, 16 changes in 20 commits per buildflow preflight) and was re-restored. `TestSpecConformanceCoverage` additionally proved shuffle-unsafe (buildflow coverage runs shuffle; the zz_ guard can execute before the recorder) and now skips only in that degenerate order. Verified twice: art-dupl 0 actionable / 14 suppressed; full suite green in canonical and shuffled order.

> **Update 2026-09-27 ~23:10 (resolution):** the bump writers were root-caused: buildflow's `go-structure-linter:repair` ("Applied fix file=go.mod rule=go-version") and `go-version-auto-configure` both push the directive to 1.27; `go-mod-normalize` deliberately rewrites patch-pins to `go 1.26`. Fixed project-side via `skip_steps` in `.buildflow.yml`; steady state is `go 1.26` (toolchain-compatible — don't restore 1.26.7, normalize rewrites it). End state all green: `nix flake check` all checks passed, golangci-lint 0 issues, full suite ok (canonical + shuffled), art-dupl 0 actionable. Remaining buildflow findings (erraudit 29, go-auto-upgrade 12, cqrs-lint 2) are pre-existing advisories in untouched code, none introduced by this pass.

---

## Verification Snapshot (end of session, all green)

| Gate                                                                          | Result                                      |
| ----------------------------------------------------------------------------- | ------------------------------------------- |
| `go build ./...`                                                              | OK                                          |
| `go test ./...` (incl. spec-conformance gate + coverage guard)                | ok (3.0s)                                   |
| `golangci-lint run`                                                           | 0 issues                                    |
| `art-dupl --sort total-tokens -t 2 --type-aware --rich-text --explain --html` | **0 actionable** (14/14 suppressed), exit 0 |
| `nix flake check` (fmt, pre-commit, sandboxed race/coverage, links)           | all checks passed                           |

---

## a) FULLY DONE

| # | Item                                                                                                                                                                                                                                                                      | Evidence                                                                                                                                                                            |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **Clone #2 eliminated:** new `accountRequirementsQuery` helper (quotes.go:124) replaces the two identical lazy-query closures in GetQuoteAccountRequirements and RefreshQuoteAccountRequirements                                                                          | Group gone from art-dupl; suite green                                                                                                                                               |
| 2 | **Clone #3 eliminated:** new generic `decodeWebhookPayload[T]` (webhooks.go:316) + `parseWebhookOccurredAt` (webhooks.go:329); all 3 event decoders (TransferStateChange, TransferPayoutFailure, BalanceCredit) refactored onto them                                      | Group gone; error strings preserved byte-for-byte, and state-change strings are test-pinned (wise_test.go:4085, 4105)                                                               |
| 3 | **Clone #4 eliminated:** new `creditOrDebit` (transactions.go:230) names the sign-based rule shared by the money-in and default branches of classifyTransactionType                                                                                                       | Group gone; zero-cents→debit edge already test-covered (helpers_test.go:42)                                                                                                         |
| 4 | **Clones #1/#5/#6 accepted** with in-source `art-dupl:accept` directives at 8 sites: GetBalance, GetQuote, FundTransfer (requireID+path two-idiom), both `toWire` ownedByCustomer tails, isPhoneChannel, isKnownStatementFormat                                           | art-dupl: 6 actionable → 0                                                                                                                                                          |
| 5 | **Build-breaking go.mod regression fixed:** `go 1.27.1` → `go 1.26.7` restored                                                                                                                                                                                            | Commit archaeology: bump entered via daemon commit 83125ea (2026-09-16, ADR-003 day); all deps require ≤1.26.7 (kin-openapi only 1.25); after fix: build/test/flake-check all green |
| 6 | **AGENTS.md updated (3 changes):** art-dupl suppression now documented as in-source-directive config with the placement gotcha; stale "one detected clone group" claim corrected; go-pin precision + recurrence warning ("check `grep '^go ' go.mod` after any `go get`") | File content; claims match this session's verified behavior                                                                                                                         |
| 7 | **Directive-placement gotcha discovered and documented:** body-interior clones are only suppressed when the directive sits INSIDE the cloned line range; doc-comment placement only works for declaration-level clones                                                    | 3 empirical art-dupl runs (see (d) #2)                                                                                                                                              |
| 8 | **Daemon mid-edit commits verified:** commit d7f655f captured final content of all touched files; remaining working-tree diff matches exactly my last refinements (go.mod, 2 directive moves)                                                                             | `git show d7f655f --stat`, `git diff`                                                                                                                                               |

---

## b) PARTIALLY DONE

| # | Item                                                          | What works                                                                                                                                    | What remains                                                                                                                                                                                                                                         | Effort |
| - | ------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| 1 | **go.mod restore is uncommitted with a generic message risk** | Fix is in the working tree; tests/flake prove it correct                                                                                      | Harness forbids me committing without explicit request; the auto-daemon will sweep it into a "chore: auto-commit" message that nowhere explains it fixed an 11-day build outage. It deserves a real commit message                                   | S      |
| 2 | **Webhook event labels are stringly-typed**                   | Typed constants exist (`WebhookEventTransfersStateChange` etc., types.go:640-644); values match the literals today, tests pin today's strings | My new helpers take `eventType string` and the 3 decoders pass literals — I **added** 3 more unguarded literals instead of typing the helper signature. Fix: helpers accept `WebhookEventType`, call sites pass constants                            | S      |
| 3 | **Webhook decoder error-classification consistency**          | Decode + occurred_at errors are corruption-classified (WrapCorruption) uniformly                                                              | BalanceCredit's two toMoney errors use plain `fmt.Errorf("map balances#credit amount: ...")` — pre-existing inconsistency I noticed and deliberately left (may be deliberate: currency-validity vs byte-corruption). Needs a decision + one-line doc | S      |

---

## c) NOT STARTED

All deliberately not started this session (scope discipline + "wait for instructions"). No code written; each needs a go-ahead or its own session:

1. **Typing the webhook label helpers** with `WebhookEventType` (see b2) — waiting on nothing, 30 min.
2. **Renaming `parseWebhookOccurredAt(value ...)`** — `value` violates the naming bar (says "contains stuff", not what it is; should be `rawOccurredAt`). I shipped it anyway.
3. **art-dupl as an enforced gate** — reverses the documented 2026-09-13 decline; needs your policy call (question g2).
4. **CI re-enable workstream** — documented "remaining before re-enable: push + enable + first green run"; untouched, priority question g3.
5. **HARVEST of section (f)** into TODO_LIST.md / ROADMAP.md via docs-health — pending your go-ahead.
6. **Canonical art-dupl invocation in AGENTS.md Build & Dev** (flags + expected 0-actionable baseline) so future runs are comparable — not written yet.

---

## d) TOTALLY FUCKED UP

1. **Master was unbuildable for ~11 days (2026-09-16 → 2026-09-27) and nobody noticed.** go.mod demanded `go >= 1.27.1`; the nix toolchain is pinned 1.26.7 with GOTOOLCHAIN=local, so EVERY `go build`/`go test` failed, locally and in the flake. Root cause: something ran `go get`/tooling with a non-nix Go 1.27 in PATH and the auto-daemon swept the directive bump into commit 83125ea among unrelated ADR-003 docs. Why it hid: CI workflow is disabled on GitHub (documented state), so no gate ever looked. Severity: blocked all local development; any consumer pinning master was broken too. Mitigation: restored 1.26.7 this session; `nix flake check` green again. Systemic fix proposed in (f) #11/#12.
2. **I guessed at art-dupl's directive semantics instead of checking, and burned 3 cycles.** Placement attempt 1 (doc comment) and 2 (standalone comment above func) did nothing; only attempt 3 (directive inside the cloned range) suppressed. I had `--help` available the whole time. Cost: 3 extra edit rounds + 3 extra full scans. Lesson now in AGENTS.md so it's the last time.
3. **I extended a split brain while claiming to remove duplication.** The typed `WebhookEventType` constants were one grep away (types.go:640). My helpers hardcode `"transfers#state-change"`-style labels at 3 call sites — currently value-identical to the constants, but the invariant is unenforced, and I added to the debt in a session whose entire purpose was deduplication. This is the most embarrassing finding; fix is S.
4. **I wrote code against imports I never verified.** Assumed webhooks.go imported `jsontext`/`time` (pattern-matched from ott.go); compiler flagged `UndeclaredName` twice. Two avoidable round-trips. Same class of sloppiness as #2: pattern-matching instead of reading.

---

## e) WHAT WE SHOULD IMPROVE

1. **Pre-flight health check before any session work:** run `go build ./...` + `git log -3` BEFORE touching anything. Would have surfaced (d1) in minute one and framed the session honestly instead of discovering mid-task. Impact: high — a broken master silently shaped 11 days of "green" local assumptions.
2. **Verify tool semantics before empirically guessing** (art-dupl directive placement). The failure pattern "try → fail → try → fail → read" should be "read → try".
3. **When extracting a helper, grep for existing typed constants of every string parameter FIRST.** A helper signature is the one moment where stringly→typed is a 1-line change; after it spreads to call sites it's a migration.
4. **Read the target file's import block before writing code into it.** Trivial, but this session proves it isn't happening by default.
5. **Daemon-hygiene as a standing rule:** verify `git show <daemon-commit> --stat` against intent after every batch of edits (did this — worked — keep it).
6. **The 2026-09-13 art-dupl decline deserves revisiting:** the in-source-directive "config" proved fragile (placement gotcha); that fragility is the strongest argument yet for a mechanical gate. Decision needed, not unilateral action.
7. **Report-format divergence:** skill default is HTML; you explicitly asked for `.md`. Honored the override, flagging it here. Not propagating it into the skill.

---

## f) NEXT TASKS (39 items, ranked by impact within category)

Legend: Impact Critical/High/Medium/Low · Effort S (<30min) / M (30min-2h) / L (>2h)

### Fix the systemic hole first

| # | Task                                                                                                                                            | Impact   | Effort | Category      |
| - | ----------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ | ------------- |
| 1 | Commit the go.mod restore with a message that explains the 11-day outage (not a daemon chore)                                                   | Critical | S      | Bug           |
| 2 | Add toolchain-drift guard: check go.mod `go` directive ≤ nix toolchain version (CI step or pre-commit) — mechanically impossible to repeat (d1) | Critical | S      | Quality       |
| 3 | Re-enable GitHub CI workflow (push, flip `disabled_manually`, first green run)                                                                  | High     | M      | Feature       |
| 4 | Root-cause which toolchain wrote `go 1.27.1` (non-nix go in PATH? which `go get`?)                                                              | Medium   | S      | Investigation |
| 5 | Add scheduled GH Action running `nix flake check` weekly so broken master can't hide behind disabled CI (depends on #3)                         | Medium   | M      | Quality       |

### Webhook split brain (shame-driven, do early)

| #  | Task                                                                                                                                                                                                | Impact | Effort | Category |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | -------- |
| 6  | Type the webhook helpers: `decodeWebhookPayload`/`parseWebhookOccurredAt` + 3 decoders take `WebhookEventType` constants instead of literals                                                        | High   | S      | Quality  |
| 7  | Add a consistency test: every decoder's event label must equal a `WebhookEventType` constant value (guards #6 forever)                                                                              | Medium | S      | Testing  |
| 8  | Rename `parseWebhookOccurredAt` param `value` → `rawOccurredAt`                                                                                                                                     | Medium | S      | Quality  |
| 9  | Decide + document BalanceCredit toMoney error classification (corruption vs plain wrap)                                                                                                             | Medium | S      | Quality  |
| 10 | Add error-message pins for payout-failure + balances#credit decoders (only state-change is pinned today, wise_test.go:4085/4105)                                                                    | Medium | S      | Testing  |
| 11 | Investigate the decoder-less constants (`transfers#active-cases`, `transfers#refund`, `transfers#case-state-change`, types.go:642-644): useful typed `trigger_on` values, or ghosts to wire/remove? | Medium | M      | Feature  |

### Dedup / quality follow-ups

| #  | Task                                                                                                                                        | Impact | Effort | Category      |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 12 | Decide: art-dupl gate (CI or pre-commit) — see question g2                                                                                  | High   | S      | Policy        |
| 13 | Document canonical art-dupl invocation + 0-actionable baseline in AGENTS.md Build & Dev                                                     | Medium | S      | Documentation |
| 14 | Run `nix run .#apidiff` to prove zero public-API drift from the refactor                                                                    | Medium | S      | Quality       |
| 15 | Direct micro-tests for the 3 new helpers (`accountRequirementsQuery("")==""`, decode corruption path, occurred_at corruption path)          | Medium | S      | Testing       |
| 16 | Calibrate one `art-dupl --no-actionability -t 2` run to confirm the 10 non-actionable groups are boilerplate                                | Low    | S      | Quality       |
| 17 | Verify TransferRequirement mirror-pair suppression survives art-dupl version updates (declaration-level directives are the fragile kind)    | Low    | S      | Quality       |
| 18 | Consider reusing `decodeWebhookPayload` inside ParseWebhookEvent's envelope decode, or document why it stays inline (different label shape) | Low    | S      | Quality       |
| 19 | Style check: helpers build error labels via concatenation while nearby WrapCorruption sites use `fmt.Sprintf` — pick one and note it        | Low    | S      | Quality       |

### Docs

| #  | Task                                                                                                                          | Impact | Effort | Category      |
| -- | ----------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 20 | HARVEST this report's section (f) into TODO_LIST.md / ROADMAP.md (docs-health)                                                | High   | S      | Documentation |
| 21 | AGENTS.md Build & Dev: document that outside `nix develop`, tests need `GOEXPERIMENT=jsonv2 go test` (I hit this immediately) | Medium | S      | Documentation |
| 22 | Skim README Design Decisions to confirm nothing describes classifyTransactionType's inline branch shape                       | Low    | S      | Documentation |
| 23 | Godoc note on webhook decoders that their error strings are test-pinned                                                       | Low    | S      | Documentation |
| 24 | Cross-link AGENTS.md art-dupl gotcha to this report                                                                           | Low    | S      | Documentation |
| 25 | Annotate/archive this report when its items ship (ANNOTATE discipline)                                                        | Low    | S      | Documentation |
| 26 | Run `nix run .#doc-verify` to confirm count-claims unaffected (internal-only refactor)                                        | Low    | S      | Documentation |

### Housekeeping / hygiene

| #  | Task                                                                                                                                   | Impact | Effort | Category      |
| -- | -------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 27 | `buildflow doctor` once (binary freshness post-session)                                                                                | Low    | S      | Housekeeping  |
| 28 | `nix flake check --all-systems` once (aarch64 currently omitted)                                                                       | Low    | M      | Quality       |
| 29 | End-of-session daemon-commit audit: `git log` vs intended changes (done mid-session, repeat at close)                                  | Low    | S      | Housekeeping  |
| 30 | Sweep `docs/status/` + `docs/reviews/`: `git mv` fully-shipped reports into `archived/` per convention                                 | Low    | S      | Documentation |
| 31 | Coverage badge: after CI re-enable (#3), verify the badge job wakes up (frozen since CI disabled)                                      | Low    | S      | Housekeeping  |
| 32 | Explicit no-release decision: this refactor touches zero exported identifiers — record that no tag is needed (this line is the record) | Low    | S      | Decision      |

### Bigger rocks (from documented repo state, NOT new research)

| #  | Task                                                                                                                                                     | Impact | Effort | Category      |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 33 | v1.0 audit follow-ups per docs/reviews/2026-08-21_v1.0-api-audit.md                                                                                      | High   | L      | Feature       |
| 34 | Demand-gated endpoints (41 shipped / 51 demand-gated per FEATURES.md): pick next by actual demand                                                        | High   | L      | Feature       |
| 35 | Refresh the 2026-08-08-era OpenAPI snapshot against the live preview reference for any NEW endpoint work (documented staleness risk)                     | Medium | M      | Research      |
| 36 | PIN challenges (JOSE/JWE) explicitly out of SDK scope — decide if that's still the right call as SCA usage grows                                         | Low    | S      | Policy        |
| 37 | Historical reports archive sweep beyond docs (any stale planning docs referencing the old art-dupl "one clone group" claim)                              | Low    | S      | Documentation |
| 38 | Consider BDD migration of the classification table (helpers_test.go) to Ginkgo DescribeTable per bdd-testing convention — only if the suite standardizes | Low    | M      | Testing       |
| 39 | After #6 lands, re-run the full battery and update the AGENTS.md baseline note to "0 actionable, typed labels"                                           | Low    | S      | Quality       |

---

## g) THREE QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Was `go 1.27.1` ever intentional?** I verified: the bump rode in via daemon commit 83125ea (2026-09-16, ADR-003 docs day), every dependency requires ≤1.26.7, and AGENTS.md pins 1.26. Everything says accidental — but only you know whether you ran `go get` with an external Go 1.27 on purpose, perhaps preparing a toolchain migration. My restore assumed accidental. If a 1.27 migration IS planned, guard #2 should track that plan instead of hard-pinning 1.26.7.

2. **Should zero-actionable art-dupl become an enforced gate now?** You declined a CI step on 2026-09-13 ("the in-source directives ARE the suppression config"). This session proved that config is fragile: two plausible directive placements silently do nothing, and it took three scan cycles to find out. The baseline is now a clean 0/14 — the cheapest moment ever to lock it in with `art-dupl check`/baseline or a CI step. Reversing your recorded decision is yours to make, not mine.

3. **Does CI re-enable jump the queue?** Master sat broken for 11 days precisely because the workflow is disabled — every other gap in section (f) is caught later or never. You documented the remaining steps (push + enable + first green run). If more endpoint work is planned before a release, I'd argue nothing else in this list protects you as much. But sequencing is a priority call only you can make.

---

_Report scope per instruction: this session's run only. No unrelated research performed. Written 2026-09-27 22:02 CEST._
