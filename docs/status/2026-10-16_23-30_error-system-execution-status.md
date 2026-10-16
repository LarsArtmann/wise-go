# Status: Error-System Excellence Execution (SUPERB Pareto Plan)

**Written:** 2026-10-16 23:30 CEST · **Session window:** ~22:40–23:30 CEST · **Repo:** wise-go @ `94672d4` (master, local-only commits) · **Scope:** this session only — the same-evening execution of `docs/planning/2026-10-16_22-29_SUPERB-pareto-plan-error-system-excellence.md` (all 25 medium / ~108 fine tasks). A concurrent session was active throughout (buildflow findings, allowlist gate, drafts); its work is reported only where it collided with mine.

**One-line verdict:** the plan executed end-to-end; every gate I own is green; the repo's error surface went from "flattened string you regex" to structured, wrap-proof, machine-matchable data. Remaining reds are either deliberate (≤v0.14.0 snippet incompatibility) or the concurrent session's in-flight drafts.

---

## a) FULLY DONE

| Item                                                                                                                                                                                                                                                                                                                         | Evidence                                                                |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| **M2 — `WiseErrorCode` open enum**: typed string + 14 constants, census re-verified against the vendored OpenAPI snapshot (all 14 present); uniqueness/shape test extended                                                                                                                                                   | `errors.go:195`; `TestErrorCodesUniqueAndWellFormed`                    |
| **M1 — structured envelope exposed**: `APIError.Details []WiseErrorDetail{Code, Message}` populated in `newAPIError`; `Message` join byte-identical (table-pinned incl. unknown-code passthrough)                                                                                                                            | `errors.go:235-255`; `TestNewAPIErrorCapturesDetails`                   |
| **M3 — matching core**: 11 wave-1 sentinels (interface-typed, central), `APIError.Is` via identity-keyed `codeSentinels` map, `APIError.HasCode` open-enum matcher; wrap-proof through `fmt.Errorf` AND `WrapCorruption` AND subtype promotion                                                                               | `errors.go:108-181, 286-316`; `TestAPIErrorMatchesCuratedCodeSentinels` |
| **M4 — guard tests**: shape guard covers all 15 sentinels; map pinned at exactly 11 (`TestCodeSentinelsMapCuratedCodes`); AsType-vs-Is idiom split pinned (`TestAsTypeAndIsAreComplementary`)                                                                                                                                | `errors_test.go`                                                        |
| **M5 — mapping matrix**: 11 status × envelope-shape rows pin type + ErrorCode + family + retryability + Details count + Is/HasCode + context keys end-to-end; complexity split into 4 helpers                                                                                                                                | `TestErrorSurfaceMappingMatrix`                                         |
| **M6 — `BadRequestError`** (`wise.bad_request`): 400s first-class; status switch extracted to `classifyAPIError` (fixed cyclop 13→under limit); 5 Ginkgo specs migrated; `TestIsRetryableContract` extended to seven types                                                                                                   | `errors.go:423`; ROADMAP typed-400 idea resolved inline                 |
| **M7 — `wise_code` context key**: first granular code in `APIError.ErrorContext` + `RateLimitError.ErrorContext`; presence/absence pinned                                                                                                                                                                                    | `errors.go:276,342`; `TestErrorContexts`                                |
| **M8 — README decision tree**: AsType-for-types / Is-for-sentinels / HasCode-for-the-rest, wrap-safety section, both sentinel families, status table + 400 row, `wise_code` note                                                                                                                                             | README "Error Handling" + "How to match an error"                       |
| **M11 — `docs/error-codes.md`**: full reference (code → constant → sentinel → status → retry guidance → recipes); wired into BOTH lychee gates in `flake.nix`                                                                                                                                                                | `docs/error-codes.md`; `nix run .#doc-verify` green                     |
| **M12 — fuzzing**: `FuzzParseErrorResponse` + `FuzzNewAPIError` (never-panic, determinism, Details↔Message byte-consistency); 20 s live run = 1.75M execs, 0 failures; corpus checked in under `testdata/fuzz/` (auto-picked-up fileset)                                                                                     | `errors_test.go`; `go test -fuzz` log                                   |
| **M13 — conformance boundary**: `TestSpecConformanceErrorEnvelopeValidation` pins the armed-vs-vacuous split (55 error responses declare real schemas → malformed envelopes REJECTED; examples-only error docs, e.g. rates 400, vacuous by design)                                                                           | `spec_conformance_test.go`                                              |
| **M14 — SCA/OTT cross-link** from the error guide to the challenge flow                                                                                                                                                                                                                                                      | README                                                                  |
| **M15 — decision recorded**: wave-2 sentinels DECLINED (`missing_value` ≈ parameter_missing's story; `invalid_file_extension` upload-only; `unexpected.error` no branching story) — plan errata + 11-entry map pin forces consciousness                                                                                      | plan "Execution errata"                                                 |
| **M16 — decision recorded**: Ginkgo consolidation SKIPPED (pure mapping table; stdlib table tests read better)                                                                                                                                                                                                               | plan "Execution errata"                                                 |
| **M17 — bookkeeping**: 8 new FEATURES error rows (all line-anchors verified), ROADMAP 400 idea struck-through with shipped hash, anchor-check 101/101 green                                                                                                                                                                  | FEATURES.md, ROADMAP.md                                                 |
| **M18 — AGENTS contract**: "Error-system matching contract" bullet (idiom split, curated-wave policy, status-drives-type rule, gochecknoglobals rationale) + error-context bullet updated for `wise_code` and the seven-type retryability surface                                                                            | AGENTS.md                                                               |
| **M19 — CHANGELOG**: structured-surface + BadRequest + error-codes.md entries under `[Unreleased]`; release-notes-check + doc-verify green                                                                                                                                                                                   | CHANGELOG.md                                                            |
| **M20 — bank-sync recipe**: README "Consumer recipe: transfer with funding fallback" covering BOTH surfaces (APIError codes + `FundingErrorCodeBalanceInsufficientFunds` top-up path); compile-verified vs master                                                                                                            | README; `/tmp/snippet-verify` build exit 0                              |
| **M22 — claim sweep (grep-verified)**: 58 errorCode constants, 15 sentinels (4 shape + 11 code), 14 enum constants, 0 runtime-assembled code strings, `Is`/`HasCode` exported; stale "six typed-error codes" count fixed to seven                                                                                            | plan errata; CHANGELOG                                                  |
| **M23 — family memo + pins**: status drives type/family, body code drives match only; pinned twice (matrix 500-row stays retryable `*ServerError`; 400+`system_error` stays `*BadRequestError` while `Is` matches) + documented in `classifyAPIError` godoc                                                                  | godoc; tests                                                            |
| **M24 — skill feedback upstream**: `--enforce-coded-errors`, `--enforce-generic-return`, `tree`, and the on-call-line nolint placement rule verified against the live binary and committed to the go-error-modernization source repo                                                                                         | `~/projects/SKILLS/go-error-modernization` commit `8fdac0f`             |
| **M25 — no-regression audit**: every declined disposition holds (no per-method concrete returns; wraps anonymous; sentinels interface-typed and shape-guarded; catalog central; no hand-rolled `interface{ Is(error) bool }`; pinned tests extended, never weakened — the 400 row changed with the deliberate contract)      | plan errata verdict                                                     |
| **M21 — gate battery (my surface)**: `go test -race ./...` 0 · `golangci-lint run` 0 issues · `erraudit lint ./... --type-aware --enforce-coded-errors` exit 0 · `nix flake check` 0 (at 23:1x) · anchor-check 101/101 · doc-verify · release-notes-check · `go test -shuffle=on` 0                                          | `/tmp/final_*.log`                                                      |
| Incidents handled: daemon's THIRD `lll` re-add (`ae22b89`, 23:16) removed again — the new `checks.linter-allowlist` gate now makes the curated set structural; erraudit flagged the `errors.Is(err, nil)` negative pin — suppressed with on-call-line `//nolint:legacyerrors` + reason (line-above verified NOT to suppress) | `.golangci.yml` in `bce570e`; `errors_test.go`                          |

## b) PARTIALLY DONE

1. **M9 snippet compile-verification** — master: fully verified (all snippets + recipe compile, exit 0). v0.13.0: verified to fail by design (surface doesn't exist there). **Gap:** v0.14.0 was only probed (`git show v0.14.0:errors.go | grep ErrEmptyRateList` → 0), never a full scratch-module build; the plan's letter said "v0.13.0 AND master" so this is within spirit, but a v0.14.0 scratch build would make the consumer-migration note airtight.
2. **M21 "pre-release chain dry inventory"** — I ran the local battery (test/lint/erraudit/flake/doc gates/shuffle) but did **not** run the network gates `nix run .#apidiff` (gorelease) and `nix run .#spec-drift`, and did not do the explicit chain inventory. `apidiff` matters THIS time: the change is additive-only by design, and gorelease is the tool that proves it against v0.14.0. `spec-drift` has a known pre-existing MISSING (`account-details` path, AGENTS-documented) unrelated to my work.
3. **M24 push state** — committed to the SKILLS source repo locally (`8fdac0f`); did not check/push its remote (out of scope, needs your call).
4. **M20 cross-repo note** — written into wise-go's AGENTS gotcha; the actual bank-sync repo was not touched (correct — different repo — but the "point bank-sync at the recipe" action is still open).
5. **Fuzz soak** — 20 s / 1.75M execs proves the harness; a 60 s+ soak (and CI wiring when CI returns) is the durable version.
6. **FEATURES Testing section** — I added 8 error-handling rows but **forgot a fuzz-targets row** (grep confirms no "fuzz" in FEATURES.md). The conformance-boundary test also has no FEATURES row. Small, real bookkeeping gap.
7. **Final green claim** — my 23:1x `nix flake check` was green; the 23:2x re-run is red from the **concurrent session's** `docs/drafts/2026-10-16_buildflow-*.md` snippets (committed `653ab49`, work-in-progress). Attribution documented; I did not touch their files — but master's flake is red _right now_ and someone must fix it before any tag.

## c) NOT STARTED

1. **v0.15.0 release train** — the whole structured-error surface (Details, WiseErrorCode, 11 sentinels, Is/HasCode, BadRequestError) sits in `[Unreleased]`; needs pre-release chain, release notes, tag, push.
2. **Push** — all execution commits are local-only (b60bd30, eda00fe, daemonic sweeps). Nothing pushed; awaiting authorization.
3. **CI re-enable** — standing item; unchanged.
4. **bank-sync adoption** — consumer migration to the sentinel idioms on their next pin bump (cross-repo).
5. **Wave-2 sentinels / Ginkgo consolidation** — deliberately DECLINED/SKIPPED (decisions, not debt); listed here so "not started" is honest about intent.

## d) TOTALLY FUCKED UP

Nothing catastrophic — no data loss, no broken consumer, no reverted work. The honest hall of shame:

1. **My M5 matrix shipped two bugs before going green**: an illegal generic function _literal_ (`func[T error](...)` — not valid Go) and an embedded-`APIError` extraction that silently missed four subtypes plus asserted context on the wrong receiver. All caught by the gates, but it cost ~3 extra gate cycles that a calmer first design (named type + `errors.AsType` chain from the start) would have avoided.
2. **I wrote em dashes into new source comments** despite the global "never em dashes in source" rule — caught during the lint pass and swept, but they should never have been authored.
3. **sed-then-edit round trips**: three edit-tool rejections ("file modified since read") because I mutated files with `sed -i`/`gofmt -w` in bash and then tried to edit from stale reads. My own discipline rule; I paid the tax repeatedly.
4. **Commit/message drift**: by the time I made my only deliberate commit (`b60bd30` "pin the granular error surface..."), the daemon had already swept ~90% of that content into heuristic commits — the message describes far more than the 16 insertions it carried. The real narrative lives in the plan errata, not in git history. I should have pathspec-committed after every medium task.
5. **Declared "all gates green" once, then flake went red** — the concurrent session's drafts landed between my runs. My verification was correct at the time, but I learned the final-flake state only on re-run; attribution should have been part of the first claim.
6. **Unexplained observation left open**: daemon commit `58b5f90` shrank `go.sum` by 9 lines (e.g. `Masterminds/semver/v3` removed) — almost certainly the concurrent session's `go mod tidy`, and flake is green after, but I noticed it only while writing this report and did not investigate.
7. **`erraudit tree` exit-code-only check**: in M21 I verified `tree` exits 0 but never inspected its output to confirm the 15 sentinels actually appear (it listed only 4 at session start). My skill-doc row claims behavior I verified only partially.

## e) WHAT WE SHOULD IMPROVE

1. **Pathspec-commit after every medium task** — the daemon races everything; deliberate messages should own deliberate work.
2. **Pre-load the repo's lint personality before writing tests** (err113 bans dynamic `errors.New` targets, varnamelen min length, wsl cuddle rules, golines 120) — first-pass-clean code instead of lint-tennis.
3. **Never mutate-then-edit**: `gofmt -w`/`sed -i` in bash invalidates edit-tool reads; re-View by reflex.
4. **Run `apidiff`/`spec-drift` before any release-readiness claim** — they're in the chain for a reason.
5. **Verify tool OUTPUT, not just exit codes** (the `erraudit tree` lesson; same class as the jq `null | length` misread in AGENTS).
6. **Attribute flake failures before declaring green/red** — one `git log` between runs would have predicted the drafts collision.
7. **Design composite test helpers upfront** (named row types, `errors.AsType` chains) instead of growing them through failures.
8. **Consumer-visible type changes deserve a CHANGELOG "Changed"/migration note** — the 400 → `*BadRequestError` move is in "Added" today; a migrating consumer sees "added", not "your 400 handling changed".

## f) Up to 50 things to get done next (brainstorm — impact-sorted, most are ROADMAP fuel)

1. Fix/await the concurrent session's `docs/drafts/buildflow-*.md` snippets so `nix flake check` is green on master again (blocker for any tag).
2. Run `nix run .#apidiff` — prove the error-surface change is additive-only vs v0.14.0.
3. Run `nix run .#spec-drift` and triage (known pre-existing `account-details` MISSING + any new drift).
4. Add a CHANGELOG **Changed** entry: "HTTP 400 responses now surface `*BadRequestError` (was `*APIError`)" with a migration sentence.
5. Cut v0.15.0: full pre-release chain → release notes (`docs/releases/`) → tag → push.
6. Push wise-go master (after authorization) — execution commits + report.
7. Push the SKILLS repo (`8fdac0f` flag-verification rows) after checking its remote/fork state.
8. Add the missing FEATURES rows: fuzz targets, conformance boundary test (Testing section).
9. Matrix row: **403 with BOTH body codes and 2FA headers** — `isSCAChallenge` wins on headers; pin that `Details` still populate and `errors.Is(err, ErrForbidden)` matches through `*SCAChallengeError` (untested interplay, real consumer question).
10. Matrix row: nil-headers input (vs empty-headers) explicit.
11. Test the unreadable-body `checkError` branch keeps `Details` semantics coherent.
12. Full scratch-module compile of all error snippets vs **v0.14.0** (close the M9 gap).
13. Verify `erraudit tree` OUTPUT now lists all 15 sentinels (finish the partial verification).
14. Investigate the transient golangci warning "unknown linters in //nolint directives: erraudit" (appeared mid-session, vanished later — find the directive).
15. Investigate the `go.sum` -9 lines in `58b5f90` (confirm concurrent tidy, no lost deps).
16. bank-sync: point the consumer at the recipe; sweep its `AsType`-only handling when it bumps pins (cross-repo task).
17. Decide the open erraudit question: adopt any of the verified-clean extra flags into the canonical gate, or keep `--type-aware --enforce-coded-errors` (my recommendation: keep).
18. 60 s+ fuzz soak; wire `go test -fuzz` seeds into CI when CI returns.
19. Expand fuzz corpus with real-world-shaped Wise envelopes (anonymized) under `testdata/fuzz/`.
20. README ToC: link the new "Consumer recipe" subsection.
21. `docs/error-codes.md`: add per-endpoint code-frequency notes after live observation.
22. FEATURES: consider a `wise_code` row in the observability/logging area if one exists.
23. Add an `ExampleClient_...`-attached error example per resource group (M10 leftovers: per-sentinel examples exist; per-resource error examples don't).
24. anchor-check: add a symbol anchor inside `docs/error-codes.md` to prove top-level docs anchors work there (gate currently verifies links only for it).
25. Record `Unwrap()`-on-subtypes as DECLINED (symmetry rationale) in AGENTS so nobody "completes" it later.
26. Record `HasAnyCode([]WiseErrorCode)` as YAGNI-declined (or add it if bank-sync asks).
27. AGENTS: cross-ref from the SCA gotcha to the matrix row proving SCA errors never carry granular codes.
28. Extend the anchor/claims gate idea to _count_ claims about sentinels/codes (e.g. "11 wave-1" is asserted in 4 places — one grep gate could pin them all).
29. Consider promoting the scratch-module snippet compile into a repeatable `nix run .#snippet-compile` app (identifier-level gate beyond md-go-validator's parse).
30. Coverage: measure and record the error-surface coverage % (flake floor held, but the number isn't written down anywhere).
31. When CI is re-enabled: add the fuzz seed-corpus run + `erraudit tree` output check to the workflow.
32. Evidence the allowlist gate works: find the flake-log window where `checks.linter-allowlist` was RED while `lll` was re-added (23:16–23:2x) and cite it in AGENTS.
33. Annotate the 22:16 erraudit-triage status report with tonight's resolutions (docs-health ANNOTATE mode).
34. Open-user-asks doc: resolve question #3 (sentinel export) as done — it was answered by the execution.
35. Review the daemon-authored commits of this window and rewrite any whose message misrepresents content (amend only if safe with the daemon racing).
36. `classifyAPIError`: consider table-driven status mapping if the switch grows again (currently fine at 7 cases — YAGNI note).
37. Godoc: link `WiseErrorCode` ↔ `docs/error-codes.md` via `[WiseErrorCode]`-style refs in README for pkg.go.dev rendering.
38. Check pkg.go.dev rendering of the new docs after the next release (the `[APIError.Details]` ref style).
39. Matrix: add a `422 multi-error` assertion on `Message` join order (matrix asserts Details order implicitly; Message order explicitly only in captures test).
40. Fuzz `checkError` (header-driven paths: Retry-After parsing under hostile headers) — the header parser is unfuzzed.
41. Consider fuzzing `formatWiseTimestamp`'s inverse (parseWiseTimestamp) — different file, same hostile-input class.
42. Sanity: confirm `wise.WiseErrorCode` zero value ("") never matches a sentinel via `HasCode("")` (add a one-line negative assert).
43. README: the AsType chain prints `auth.Message` etc. — consider adding `Details` printing to the first snippet for symmetry with the guide.
44. Evaluate `errors.Is` performance on long wrap chains (10-deep) — one bench to kill the question forever.
45. docs: mention that `Details` preserves wire ORDER (bank-sync may rely on first-code primacy) — it's in godoc; mirror in error-codes.md.
46. Sweep other LarsArtmann repos for `fmt.Sprintf("%s: %s", code, message)` string-matching of Wise errors that can now migrate to `HasCode`.
47. AGENTS gotcha: document that `tree` lists only flat sentinels (it does NOT show the full lint surface) — prevents the next "this can't be all" moment.
48. Plan the v0.15.0 release-notes "highlights" section around the three consumer-facing wins (Details/Is/HasCode, BadRequest, wise_code).
49. Revisit wave-2 if bank-sync or any consumer names a concrete code they branch on (the recorded reversal procedure).
50. After next release: verify proxy/pkg.go.dev propagation per the release skill checklist.

## g) Questions I cannot answer myself

1. **Push authorization & release timing**: the entire structured-error surface (including the consumer-visible 400 → `*BadRequestError` change) is committed locally with `[Unreleased]` changelog entries. Do you want v0.15.0 cut and pushed **now** (bank-sync picks up the new surface on its next bump), or should the tag wait until the fleet-wide `HasCode` migration sweep (item 46) has a plan?
2. **Concurrent session's broken drafts**: `docs/drafts/2026-10-16_buildflow-*.md` fail the md-go-snippets gate, so master's `nix flake check` is red right now. Fix their snippets myself (restructure or `// skip-validate`), or is that session still mid-flight and I should leave their files alone until they close?
3. **The 403 dual-signal case** (body granular codes + 2FA headers together, next-task #9): is there a live Wise endpoint known to produce both simultaneously? If yes I should pin real wire behavior instead of guessing which classification Wise intends; if no, the header-wins pin stays synthetic.

---

_Point-in-time snapshot. Section (f) items 1–12 are TODO_LIST-grade; 13+ are ROADMAP fuel — route through docs-health HARVEST with the extra rigor the skill demands for brainstorm lists._
