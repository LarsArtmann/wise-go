# Status: samber/oops error-construction migration (session report)

**Report time:** 2026-11-27 09:41 CET · **Session:** 2026-11-27, work committed by the daemon at 09:38, report written 09:41 to 09:55 CET · **Repo state at report time:** `git rev-parse --short HEAD` = `b3967ba`, working tree clean (0 changes before this file).

## Scope line: what this report verified, and what it did not

**Verified fresh at report time, on `b3967ba` (09:41 to 09:45):**

- `go test ./...` exit 0 (root package ok in 9.3 s, `internal/raw` ok).
- `erraudit lint ./... --type-aware --enforce-coded-errors --enforce-samber-oops` exit 0.
- `erraudit lint ./... --max` exit 2 with exactly 58 findings, all `generic_return` / warning (grouped from `--format json`).
- `buildflow -s golangci-lint` exit 0, no findings printed.
- `git status --short` empty; `git log 9b3aae5..b3967ba` is one commit of 36 files (+450 / -286).
- The commit's content: 102 non-test `oops.Wrapf`/`oops.Errorf`/`oops.New` call lines (104 added call lines minus 2 in `errors_test.go`); the only non-test `fmt.Errorf`/`errors.New` left is `errors.go:173` inside `newSentinel`.
- Two runtime probes in a scratch module (`/tmp/wise-go-oops-probe`, `replace` to this repo): the `errors.Is` panic and the `slog`/JSON rendering in section (d).

**Verified during the session, on the working tree that became `b3967ba`, NOT re-run at report time:**

- `nix flake check` exit 0. The last run predates one final sentence in `AGENTS.md` (the cue-policy note). Markdown only, formatted with the same dprint, but the hermetic check has not seen it.
- `nix run .#anchor-check` (104 anchors + 5 quoted pointers), `nix run .#doc-verify`, `nix fmt .` (0 changed), `nix develop -c dprint fmt` on every edited markdown file.
- `buildflow -s cue-vet` exit 0 after adding `permissions: contents: read` to `ci.yml`.
- Full `buildflow --build-mode=full`: 60/75 steps, exit 69, 83 findings. NOT repeated after the cue-vet fix and the last `AGENTS.md` edit.
- On a clean worktree of the previous HEAD `9b3aae5`: `buildflow -s docs-gates` exit 69 (4 findings) and `buildflow -s go-humanize-linter` exit 69 (1 finding), so both are older than this session.

**Not verified, and why:**

- `nix run .#apidiff`, `nix run .#pre-release`, release-notes check: not run. They need a clean tree plus network and belong to a release, which was not requested.
- The erraudit version CI installs (`v0.5.1-0.20260922174106-1c6809adf02e`) was never run against this tree. Every erraudit result here is from the local binary (`erraudit version cb96501`).
- Why `generic_return` went from 59 to 58 findings: not investigated.
- OTP/OTT absence in the new verbose rendering (`%+v`, `Stacktrace()`, `Sources()`): reasoned about (stacks carry program counters, not variable values), not pinned by a test.
- Cost numbers (1.3 µs vs 0.14 µs per error): one run, 200000 iterations, no allocation stats, scratch module that no longer exists. Not a committed benchmark.
- `anchor_check.py`'s "now at" hint appeared wrong during the session (details in (e)); not root-caused.
- Nothing outside this session's work was researched, on instruction.

## Answer first

| Question                         | Short answer                                                                                                                                                                                                                                                                                                                                |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| What did you forget?             | Two things, both verified at report time. (1) `APIError.Is` panics when `errors.Is` is called with a non-comparable target, and `oops` errors are non-comparable, so the migration made an old hole much easier to fall into. (2) Every SDK error now carries a stack trace in `slog` and JSON output, implicitly. I documented only `%+v`. |
| What could you have done better? | The first turn defended a recorded decision the owner never held; the line-anchor repair was done twice and had to be reset; exit codes were piped away a few times; the SDK API was guessed instead of read. All caught, none left residue, all cost time.                                                                                 |
| What could you still improve?    | Use oops for what it is good at (codes, key-value context, safe public messages) instead of a like-for-like wrap; make `APIError.Is` panic-free at the root; turn the three tooling traps hit this session into fixes.                                                                                                                      |

## a) FULLY DONE

All of it is committed in `b3967ba` (2026-11-27 09:38, auto-commit daemon, hooks bypassed). The tree was clean before this report.

| #  | Item                                                                                                                  | Evidence                                                                                                                                                                                                 |
| -- | --------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | `samber/oops v1.23.2` added as a direct dependency (`go get`, then `go mod tidy`)                                     | `go.mod`; indirect additions `samber/lo`, `oklog/ulid/v2`, `go.opentelemetry.io/otel`                                                                                                                    |
| 2  | 102 runtime error-construction sites moved to `oops.Wrapf` / `oops.Errorf` / `oops.New`                               | 101 flagged by erraudit plus 1 erraudit-suppressed site (`ott.go`, `VerifyOTT`); 20 non-test files import oops; grep leaves only `newSentinel`                                                           |
| 3  | Message text preserved                                                                                                | Full test suite green with no pinned-message edits; scratch consumer printed the same `%v` as before                                                                                                     |
| 4  | 15 package sentinels built through one comparable constructor, `newSentinel` (`errors.go`)                            | `TestSentinelsStayInterfaceTyped` (regex updated), new `TestSentinelsStayComparable`, `TestSentinelsMatchThroughProductionWraps` extended                                                                |
| 5  | 4 redundant nil guards removed (`Authenticate`, `VerifyCredentials`, `fetchByID`, `DeleteProfileWebhookSubscription`) | erraudit `oops_redundant_guard` clean; `oops.Wrapf(nil, ...)` returns nil                                                                                                                                |
| 6  | Lint config aligned                                                                                                   | `.New(` added to `wrapcheck.ignore-sigs`; 7 `//nolint:err113` call-site directives deleted; one `//nolint:erraudit, err113` on `newSentinel`                                                             |
| 7  | erraudit gate extended                                                                                                | `ci.yml` runs `--enforce-samber-oops`; canonical invocation exits 0 on `b3967ba`                                                                                                                         |
| 8  | erraudit `--max` reduced                                                                                              | 175 findings (116 error + 59 warning) before, 58 warnings after, one rule left                                                                                                                           |
| 9  | Nix                                                                                                                   | `vendorHash.nix` refreshed; `nix flake check` exit 0 during the session                                                                                                                                  |
| 10 | CI policy                                                                                                             | `ci.yml` gained a workflow-level `permissions: contents: read` (fleet `cue-policy` `#CIWorkflow` requirement); `buildflow -s cue-vet` exit 0                                                             |
| 11 | Decision record                                                                                                       | `docs/adr/004-samber-oops-error-construction.md` (new)                                                                                                                                                   |
| 12 | Living docs updated                                                                                                   | `AGENTS.md` (8 changed lines), `README.md` (4), `CHANGELOG.md`, `CONTRIBUTING.md`, `FEATURES.md` (1 row, 91 line references remapped), `docs/DOMAIN_LANGUAGE.md` (1 row), `docs/error-codes.md` (1 line) |
| 13 | Incidental                                                                                                            | `.github/badges/coverage.json` dprint-formatted (it was minified by `9b3aae5` and failed `nix flake check` before any change of mine)                                                                    |

## b) PARTIALLY DONE

| # | Item                                        | Works now                                                                | Still open                                                                                                                             | Effort |
| - | ------------------------------------------- | ------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| 1 | Full buildflow run                          | 60/75 steps; cue-vet finding fixed and re-verified single-step           | Not re-run in full after the last two edits; `docs-gates` (4) and `go-humanize-linter` (1) remain red, both older than this session    | M      |
| 2 | "Best error handling possible"              | Like-for-like migration: same text, stack added, matching unchanged      | None of oops' structured features are used: no `oops.Code`, no key-value context, no `Public` message; the oops layers carry text only | M to L |
| 3 | Consumer-facing documentation of the change | README "Where an error happened" (`%+v`, `oops.AsOops`); CHANGELOG entry | The implicit `slog`/JSON change is undocumented (section (d) item 2)                                                                   | S      |
| 4 | erraudit enforcement in CI                  | Flag added to the gate command                                           | Job stays warn-only until the `ERRAUDIT_TOKEN` secret exists (tracked in `TODO_LIST.md`, blocked on the owner)                         | S      |
| 5 | Closing the loop on this report             | This file                                                                | Section (f) is not harvested into `TODO_LIST.md` / `ROADMAP.md` (I was told to wait for instructions)                                  | M      |

## c) NOT STARTED

| #  | Item                                                                                                 | Why not                                                         |
| -- | ---------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| 1  | Fix for the `APIError.Is` panic on non-comparable targets (section (d) item 1)                       | Found at report time; report-then-wait instruction              |
| 2  | Policy and docs for stack traces in `slog`/JSON output                                               | Same                                                            |
| 3  | Release: version bump, release notes, `pre-release`, `apidiff`, tag                                  | Not requested; needs a decision (question 1)                    |
| 4  | Committed benchmark for the error-construction paths                                                 | Only a scratch microbenchmark was run                           |
| 5  | Regression guard: no stdlib error constructors outside `newSentinel` in non-test code                | The CI gate is warn-only, so nothing blocks a regression        |
| 6  | Stale-fact cleanup in `AGENTS.md`, `README.md`, `ROADMAP.md` (list in (e))                           | Noticed, out of the migration's scope                           |
| 7  | Pre-existing red gates: `docs-gates` (annotate, audit-all, cadence, freshness), `go-humanize-linter` | Older than this session; verified on a clean `9b3aae5` worktree |
| 8  | Recurrence fix for the minified coverage badge                                                       | Only the symptom was formatted                                  |
| 9  | Annotating historical reports that still say "samber/oops DECLINED"                                  | Point-in-time documents; one-line pointers not yet added        |
| 10 | Cross-project lessons (crush-config `references/lessons.md`, commit-only repo)                       | Needs a commit in another repo                                  |

## d) TOTALLY FUCKED UP!

### 1. `errors.Is` panics on a non-comparable target, and oops errors are non-comparable

- **Verified at report time** (scratch consumer, `b3967ba`): an SDK error from `ListBalances` (HTTP 404), then:

  ```text
  errors.Is(err, wise.ErrResourceNotFound)   -> true
  errors.Is(err, io.EOF)                     -> false
  errors.Is(err, oops.New("consumer sentinel")) -> PANIC: runtime error: hash of unhashable type oops.OopsError
  errors.Is(err, struct-with-a-map{})        -> PANIC: runtime error: hash of unhashable type main.boxed
  ```

- **Cause:** `APIError.Is` (`errors.go`) looks its target up with `codeSentinels[target]`. A map lookup hashes the key; an unhashable dynamic type panics. `errors.Is` calls `Is` on every error in the chain, so any `*APIError` below a wrap triggers it.
- **Severity:** high for consumers. A matching call on an error path crashes the caller's goroutine. Requires a non-comparable target.
- **Not introduced by the migration:** the struct-with-a-map case panicked before. But the migration made non-comparable error values the house style, so `oops.New(...)` sentinels in consumer code are now a realistic target.
- **My part:** I reasoned about comparability when deciding how to build the SDK's own sentinels and missed the lookup side of the very same map, in a function whose tests I extended.
- **Workaround for consumers:** comparable targets (`errors.New`, typed sentinels), or `APIError.HasCode`.
- **Fix size:** S. Replace the map lookup with an identity `switch`, or guard `reflect.TypeOf(target).Comparable()`. Add a test with both target kinds.

### 2. Every SDK error now logs a stack trace with absolute paths, without any `%+v`

- **Verified at report time:** `slog` with a `JSONHandler` logging `"err", err` emits a group with `message`, `err`, `time`, `trace` and `stacktrace`. The stack contains absolute source paths of the build (here `/home/lars/projects/wise-go/client.go:401` and the consumer's `/tmp/wise-go-oops-probe/main.go:30`). `json.Marshal(err)` returns 875 bytes with `stacktrace` and `trace` fields.
- **Cause:** `oops.OopsError` implements `slog.LogValuer` and `json.Marshaler`. They fire implicitly. I read both method names in the source listing and documented only `%+v`.
- **Severity:** medium. Three exposures: (a) log volume on every SDK error in services that log errors as attributes; (b) build-machine paths in logs; (c) a consumer that marshals an SDK error into an HTTP response ships its stack trace to its own end users. The SDK sets no `oops.Public`, so `oops.GetPublic` has nothing safe to return.
- **Undocumented:** `CHANGELOG.md` and `README.md` mention `%+v` only.
- **The daemon's commit body says the change "changes diagnostics only, never the matching contract".** True for matching with comparable targets; incomplete on both counts above.

### 3. Red gates older than this session (verified on a clean `9b3aae5` worktree)

- `buildflow -s docs-gates` exit 69: `annotate`, `audit-all`, `cadence`, `freshness` are red.
- `buildflow -s go-humanize-linter` exit 69: `activities.go:316`, manual comma-grouped number parsing (error severity).
- Not caused by this session, not fixed by it.

### 4. The coverage badge will break `nix flake check` again

- The CI badge job writes minified JSON. The hermetic `dprint-format` check rejects it. I formatted the current file only, so the next badge update reproduces the failure. Root cause (dprint scope, or the badge job) untouched.

### 5. `buildflow -s nix-hash-fix --fix` cannot end green in this repo

- It rewrote `vendorHash.nix` correctly, then its verify step built `packages.default`. This is a library that installs no binary, so the step exits 69 with an error finding. Documented in `AGENTS.md` as known; the cause (the step's build target) is not fixed, so every dependency bump ends on a red step.

## e) WHAT WE SHOULD IMPROVE!

### What I forgot

1. The lookup side of `codeSentinels` (section (d) item 1).
2. `LogValuer` and `MarshalJSON` on `OopsError` (section (d) item 2).
3. The erraudit version CI pins was never exercised against the new tree or the new `//nolint:erraudit, err113` directive.
4. Two comments in `errors.go` still describe the old world: `:181` "built from the errors.New sentinels above" and `:312` "arbitrary `fmt.Errorf("%w")` wrap chains".
5. In my closing summary I said "101 runtime sites"; it is 102 (the suppressed `ott.go` site was migrated too).
6. Cost numbers come from one scratch run. Nothing committed pins them.

### What I could have done better (process)

1. **First turn.** I spent a turn triaging 175 findings and writing an `AGENTS.md` paragraph that defended "samber/oops declined, dependency-light by design". The owner overturned it in one sentence and said they never held that belief. The "ADR 003 says dependency-light" attribution was written by earlier agent sessions. When a direct request ("fix it all", "I want the best error handling possible") collides with a recorded decline, the collision is a question for the owner before any memory is written.
2. **Line-anchor repair, done twice.** After a correct first remap of 46 anchors I ran a second, broader remap over every reference. It shifted the first 46 again. I then overwrote the evidence log I still needed, tried to rebuild the list from memory, and mis-transcribed rows. Clean exit: reset `FEATURES.md` from HEAD and map exactly once from a known baseline. About six wasted tool calls. Rule: never overwrite a log that a later step depends on; apply a remap once.
3. **Piped exit codes.** Commands like `go vet . 2>&1 | tail -5; echo $?` printed `tail`'s status. `AGENTS.md` has a gotcha for exactly this. No residue (the flake check runs vet), but the "vet=0" I read was not vet's.
4. **Guessed the SDK API.** The scratch consumer needed three compiles because I guessed `wise.New`'s signature, an option name and an ID constructor. The README quick start had all three. The repo's own rule is "compile before you paste"; I complied, after guessing.
5. **Rewrite tool debugged one bug at a time.** Five failed runs before a clean one, for: leftover lines, a `Sscanf` verb, trailing commas, a prefix `%w`, and a filename filter that silently skipped `third_party_transfers.go`. The linter caught the skipped file, not me. A dry-run mode printing every rewrite before writing would have shown all five at once.
6. **Wrote `AGENTS.md` edits from remembered text twice** (one wording, one neighbour bullet). `AGENTS.md` itself says never to trust a remembered copy. Asserts caught both before any write.
7. **No independent review.** 102 mechanical rewrites went in without a second reader. A read-only review pass over the diff would plausibly have caught both items in (d) before the commit.
8. **Light gates ran next to a heavy one** (`go test` during the full buildflow), against the repo's "serialize heavy gates" rule.
9. **At report time** a heredoc aimed at a `/tmp` directory that had been cleaned hours earlier. The `cd` failed, nothing was written into the repo (checked), but scratch modules do not survive a long session.

### What could still improve (design and quality)

1. **Use oops as more than a wrapping formatter.** Today every oops layer holds text. Candidates: `oops.Code` returning the same `wise.*` code as `ErrorCode()`, key-value context at the transport layer (method, URL, status), `oops.Public` for a safe message.
2. **Make `APIError.Is` panic-free at the root.** A `switch target { case ErrNotValid: ... }` compares identity without hashing any caller-supplied value.
3. **Decide the logging contract.** Stack by default (current), or message plus code by default with the stack on request.
4. **Tooling traps hit this session**, each worth a fix in its owning repo:
   - `anchor_check.py`'s "now at" hint: it reported `CreateBalance` "now at 105" while the function is at `balances.go:108`, and `CreateProfileWebhookSubscription` "now at 81" while it is at `webhooks.go:90`.
   - Bare `file.go:NN` references are checked for "non-blank" only, so 45 references the checker did not flag were shifted along with the 46 it did catch, until I remapped by content.
   - The erraudit suppression looked order-sensitive: `//nolint:err113, erraudit` did not suppress, `//nolint:erraudit, err113` does. Order and comment text both differed; not isolated.
   - `buildflow -s gomod-check --fix` left `samber/oops` marked `// indirect`; only `go mod tidy` fixed it.
   - `buildflow -s nix-hash-fix --fix` builds the wrong target for a library (section (d) item 5).
   - `nix flake check` passed while `cue-vet` was red, so the fleet workflow policy is not part of the flake check.
5. **Stale facts in living docs**, all older than this session, all noticed while working:
   - `AGENTS.md`: `go-retry v0.7.1` (go.mod: v0.9.0), `go-error-family v0.11.0` (v0.12.0), "pins go 1.26" / "steady state `go 1.26`" (go.mod: `go 1.27.3` since `8a78fa9`, 2026-11-04; CI `GO_VERSION: "1.27"`; flake `go_1_27`), "erraudit ... no local install" (a standalone `erraudit` is installed and ran locally all session).
   - `README.md` line 93: "Requires **Go 1.26 or later**" while the module requires 1.27.3.
   - `ROADMAP.md`: a "Go 1.27 migration plan" row for a migration that is in place.
6. **Test hygiene.** `TestSentinelsStayComparable` repeats the 15-name list of `TestSentinelsStayInterfaceTyped`; one table should feed both.
7. **`CHANGELOG.md`** lists `Changed` above `Added`/`Fixed` and keeps their "Nothing yet" placeholders.

## f) Up to 50 things to get done next

Ranked by impact. Items 1 to 12 are `TODO_LIST.md` candidates; 13 onward are mostly `ROADMAP.md` fuel. None is harvested yet.

| #  | Task                                                                                                                                                                | Impact   | Effort | Category      |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------ | ------------- |
| 1  | Fix `APIError.Is`: identity `switch` (or comparability guard) instead of `codeSentinels[target]`; add a test with an `oops.New` target and a struct-with-map target | Critical | S      | Bug           |
| 2  | Document the implicit `slog`/JSON change in `CHANGELOG.md` and README "Where an error happened"                                                                     | High     | S      | Documentation |
| 3  | Set `oops.Public` on SDK errors (or document `oops.GetPublic`) so a consumer never forwards a stack trace to its own users                                          | High     | M      | Feature       |
| 4  | Run the CI-pinned erraudit (`v0.5.1-0.20260922174106-1c6809adf02e`) against this tree; align the pin with the local `cb96501` if verdicts differ                    | High     | S      | Quality       |
| 5  | Re-run the full `buildflow --build-mode=full` on `b3967ba`; confirm only the two older red steps remain                                                             | High     | M      | Quality       |
| 6  | Re-run `nix flake check` once with the final `AGENTS.md` sentence and this report                                                                                   | High     | S      | Quality       |
| 7  | Harvest this section into `TODO_LIST.md` / `ROADMAP.md` (docs-health HARVEST)                                                                                       | High     | M      | Documentation |
| 8  | Release: decide the version (minor, the rendering change is consumer-visible), write release notes, `nix run .#pre-release`, `nix run .#apidiff`, tag               | High     | M      | Documentation |
| 9  | Pin "no secret in verbose rendering": `%+v`, `Stacktrace()`, `Sources()`, `Context()`, `ToMap()` of OTT and statement errors never contain an OTP or OTT value      | High     | S      | Bug           |
| 10 | Add a source-scan test: no `fmt.Errorf` / `errors.New` in non-test code outside `newSentinel` (CI gate is warn-only)                                                | Medium   | S      | Quality       |
| 11 | Fix the two stale comments in `errors.go` (`:181`, `:312`)                                                                                                          | Medium   | S      | Cleanup       |
| 12 | Commit a benchmark for the error paths (`oops.Wrapf` around `newAPIError`, `NewCurrency` rejection, timestamp-parse failure) with allocation stats                  | Medium   | S      | Quality       |
| 13 | Root-cause the erraudit suppression order sensitivity; record the verified rule in `AGENTS.md`                                                                      | Medium   | S      | Documentation |
| 14 | Update stale `AGENTS.md` versions and the "no local install" claim (verify each against `go.mod` first)                                                             | Medium   | S      | Documentation |
| 15 | README line 93: "Go 1.26 or later" versus `go 1.27.3`                                                                                                               | Medium   | S      | Documentation |
| 16 | Triage the red `docs-gates` (`annotate`, `audit-all`, `cadence`, `freshness`): resolve or ratchet                                                                   | Medium   | M      | Documentation |
| 17 | Fix `go-humanize-linter` at `activities.go:316`                                                                                                                     | Medium   | S      | Bug           |
| 18 | Exclude `.github/badges/**` from dprint (or format in the badge job) so the coverage badge stops breaking the flake check                                           | Medium   | S      | Quality       |
| 19 | Point `nix-hash-fix` at `checks.build` for this library, or file it upstream                                                                                        | Medium   | M      | Quality       |
| 20 | `anchor_check.py`: verify and fix the "now at" hint; add a self-test case                                                                                           | Medium   | S      | Bug           |
| 21 | Tighten anchor-check for bare `file.go:NN` references (today "non-blank" only), or convert them to symbol anchors                                                   | Medium   | M      | Quality       |
| 22 | Add a flake check or buildflow step for `cue-vet` so the fleet workflow policy is enforced locally                                                                  | Medium   | M      | Quality       |
| 23 | Structured context at the transport layer: `oops.With` for method, URL and status on `doRequest` errors                                                             | Medium   | M      | Feature       |
| 24 | Align `oops.Code` with the `ErrorCode()` catalog so `oops.AsOops(err).Code()` returns the `wise.*` code                                                             | Medium   | M      | Feature       |
| 25 | Make the erraudit CI job blocking once `ERRAUDIT_TOKEN` exists (delete the two `continue-on-error` / `if` lines)                                                    | Medium   | S      | Quality       |
| 26 | Check allocation growth on retry storms: `doRequest` now builds one stack per attempt                                                                               | Medium   | S      | Quality       |
| 27 | Package godoc: state the error model (typed error from go-error-family, oops context wrap with stack, comparable sentinels)                                         | Medium   | S      | Documentation |
| 28 | Compile-only `Example` twin for the README "Where an error happened" snippet                                                                                        | Low      | S      | Documentation |
| 29 | Extend `TestErrorSurfaceMappingMatrix`: two oops layers, a consumer `fmt.Errorf("%w")` layer, the go-retry exhaustion chain, `errors.Is(err, nil)`                  | Low      | S      | Quality       |
| 30 | Fuzz `APIError.Is` / `HasCode` with arbitrary, including non-comparable, targets                                                                                    | Low      | M      | Quality       |
| 31 | One shared sentinel table for `TestSentinelsStayInterfaceTyped` and `TestSentinelsStayComparable`                                                                   | Low      | S      | Cleanup       |
| 32 | `CHANGELOG.md`: conventional section order, placeholders handled consistently                                                                                       | Low      | S      | Documentation |
| 33 | Annotate the historical reports that say "samber/oops DECLINED" with a one-line pointer to ADR 004                                                                  | Low      | S      | Documentation |
| 34 | Re-check the two sites that nest a full oops layer inside `errorfamily.WrapCorruption` (`toMoney`, `parseEnum`); measure with the committed benchmark               | Low      | S      | Quality       |
| 35 | `ROADMAP.md`: reconcile the "Go 1.27 migration plan" row with `go 1.27.3` in `go.mod`                                                                               | Low      | S      | Documentation |
| 36 | Confirm the migration's coverage delta (badge says 93.7%)                                                                                                           | Low      | S      | Quality       |
| 37 | `gomod-check --fix` and the `// indirect` marker: confirm by design or file upstream                                                                                | Low      | S      | Quality       |
| 38 | govulncheck: the full run listed 13 finding lines (first: GO-2026-6859, net/http `MaxBytesReader`); `AGENTS.md` says 12. Recount and update                         | Low      | S      | Documentation |
| 39 | lychee: two dead `file://` links inside the generated `docs/reviews/2026-10-16_sca-auth-data-model.html`; exclude generated HTML or accept as warning               | Low      | S      | Quality       |
| 40 | bandit-check: triage the one finding (plus 9 `# nosec` sites)                                                                                                       | Low      | S      | Quality       |
| 41 | README Observability: mention the tuning knobs `oops.StackTraceMaxDepth` and `oops.AutoTraceID`                                                                     | Low      | S      | Documentation |
| 42 | README TOC policy: "Where an error happened" and "How to match an error" are not listed while other h3 entries are                                                  | Low      | S      | Documentation |
| 43 | Decide whether `erraudit --max` should be a meaningful local target (58 `generic_return` warnings keep it at exit 2)                                                | Low      | S      | Quality       |
| 44 | The daemon's commit subject is typed `chore(errors)` for a consumer-visible change; reword or cover it in release notes (owner's call, it rewrites a commit)        | Low      | S      | Cleanup       |
| 45 | Cross-project lessons into crush-config `references/lessons.md` (by commit): evidence logs, remap once, agent-recorded decisions, hashing error values, `LogValuer` | Medium   | S      | Documentation |
| 46 | Fleet follow-up on instruction: other repos where oops was declined as "dependency-light" re-decide                                                                 | Medium   | M      | Documentation |
| 47 | `nix run .#release-notes-check` once release notes exist                                                                                                            | Low      | S      | Quality       |
| 48 | `nix run .#pre-release --shuffle` before any tag (retry-exhaustion chain under shuffle and race)                                                                    | Low      | S      | Quality       |
| 49 | Review `docs/error-codes.md` for remaining wrapping examples that show only `fmt.Errorf`                                                                            | Low      | S      | Documentation |
| 50 | Remove `/tmp/wise-go-oops-probe` (scratch module recreated at report time)                                                                                          | Low      | S      | Cleanup       |

## g) Questions I can not answer myself

1. **Release.** `b3967ba` is the daemon's commit, typed `chore(errors)`. Do you want this shipped now as a minor release (I would cut `v0.19.0`: the `slog`/JSON rendering change is consumer-visible), with me running the pre-release chain, `apidiff` and release notes? And do you want the commit subject reworded first? I will not rewrite a commit without your say-so.
2. **Depth.** Is the like-for-like migration the goal, or do you want the structured phase: `oops.Code` aligned with the `ErrorCode()` catalog, key-value context at the transport layer, `oops.Public` safe messages? This decides items 3, 23 and 24 and whether the error model section of the README grows.
3. **Logging policy.** Should SDK errors carry a full stack trace in `slog` and JSON output by default (current state: absolute paths, about 875 bytes per error as JSON), or should the default be compact (message plus code) with the stack available only on request? This is a taste and operations call about your log pipeline, not something I can read from the repo.
