# Status: P3 unblocked sweep — webhook typing, fuzz corpus, coverage floor 92, CI wiring

**Date:** 2026-10-11 18:34 CEST
**Session:** ~11:15–18:30 CEST, single continuous session.
**Scope of this report:** this session only (per instruction). Baseline at session start: clean tree at `a165182`, all gates green.
**Trigger:** generic "READ, UNDERSTAND, RESEARCH, REFLECT; execute and verify one step at a time" — session self-directed from TODO_LIST's unblocked P3 items.
**Method:** 7-item todo list executed sequentially; every change verified before the next (build → targeted tests → full gates at the end).

---

## a) FULLY DONE

1. **Webhook label surface typed** (TODO item, carried since 2026-09-27): `decodeWebhookEvent`, `decodeWebhookPayload`, and `parseWebhookOccurredAt` now take `WebhookEventType` instead of `string` (webhooks.go:331/344/362); the three event-decoder call sites use the `WebhookEvent*` constants instead of string literals. Bonus carried item done in passing: `parseWebhookOccurredAt`'s `value` param renamed to `rawOccurredAt` (carried since 2026-09-27 22:02 c2). Error messages are byte-identical — pinned by the existing tests, which pass.
2. **Account-requirements verbs typed** (same TODO item): `accountRequirementsVerb` type with `verbGetAccountRequirements` / `verbRefreshAccountRequirements` (quotes.go:170–176); both call sites converted. Internal only, no public API change.
3. **Fuzz corpus seeds checked in** (TODO item): 12 hand-built seed files under `testdata/fuzz/{FuzzParseWebhookPublicKey,FuzzVerifyWebhookSignature,FuzzDecodeExchangeRates}/` — includes a real 2048-bit PKIX RSA key, a PKCS#1 RSA key, an EC key (wrong-type rejection path), garbage-DER PEM, base64/length-edge signature strings, and the **live `/v1/rates` numeric-zone regression shape** as an exchange-rates seed. Corpus verified: `go test -run Fuzz` loads and passes every seed file.
4. **`FuzzParseWiseDate` added** (TODO item, helpers_test.go:468): pins never-panics, error⇒zero-time, empty⇒zero-time, accepted⇒UTC-date-layout round-trip. Live-fuzzed 5s clean. Seeds include leap-day accept/reject pair.
5. **`docs/bench/README.md` one-pager** (TODO item): conventions, rerun/compare commands, append-only baseline table.
6. **`./testdata` added to the flake source fileset** — the sandboxed `checks.test` derivation now vendors the fuzz corpus, so `nix flake check` exercises it (verified: flake check green after the change).
7. **Coverage floor 90 → 92** (TODO item): promoted in BOTH enforcement sites — flake.nix:153 (`checks.test` checkPhase) and the CI coverage gate (ci.yml) — comments refreshed with the measured 92.8% (2026-10-11).
8. **CI wiring for release-notes-check** (TODO item): explicit `nix run .#release-notes-check` step added to the CI nix job; `md-go-snippets` was already covered inside `nix flake check` (which that job runs) — documented in the step comment.
9. **LSP-narrative de-dup** (TODO item): removed the struck resolved ROADMAP entry (AGENTS.md gotcha is the canonical home) plus the already-shipped "benchstat baseline file" raw idea. ROADMAP holds future ideas only (for the lines I touched).
10. **Docs maintenance** (docs-health rules applied): completed TODO items deleted (5 items — history lives in CHANGELOG `Unreleased`, which was filled); AGENTS.md refreshed (coverage numbers 93.1/90 → 92.8/92; fileset gotcha now mentions `./testdata` auto-inclusion).
11. **Two pre-existing breakages fixed on sight** (found by the full gate run):
    - `scripts/pin_sync.py` had a shebang but mode 100644 → ruff EXE001 failed `buildflow --fix`, blocking 10 downstream steps. Fixed via exec bit (staged; the script is invoked as `scripts/*.py` in a `scripts/` dir). Root cause: added 2026-10-07 with the pin-sync gate, mode never correct in git.
    - `.github/SECURITY.md` missing the OSSF-mandated **Security Practices** section → securitymd error-severity finding tripped the buildflow findings gate. Added a truthful, project-specific section (secret handling, webhook signature-first policy, gate battery).
12. **Final verification battery, all green:** `buildflow --build-mode full` exit 0 (59 steps success, warnings only) · `nix flake check` all 8 checks passed · `go test -race ./...` green · `golangci-lint run ./...` 0 issues · `nix run .#doc-verify` all passed · `nix run .#release-notes-check` clean · canonical `art-dupl -t 2 --type-aware` 0 actionable (16 groups, unchanged baseline) · `go.mod` still `go 1.26.0` (flipflop guard intact).
13. **Daemon-commit audit:** every auto-commit from this session was diff-verified to contain exactly my intended changes (`3ac3cf3`, `e0dfee1`, `3700c90`, `34a4f48`, `9fe4f31`, plus the earlier `e395766` sweep that picked up the corpus). Working tree clean at report time.

## b) PARTIALLY DONE

1. **Live fuzzing of the three seeded targets** — corpus seeds run under plain `go test`, and `FuzzParseWiseDate` got a 5s live fuzz, but I did NOT live-fuzz the three seeded targets this session (they were live-fuzzed 8s each on 2026-10-07 pre-corpus). The corpus itself is verified only in seed-execution mode.
2. **Coverage-drop forensics** — I measured 92.8% vs the documented 93.1% (2026-10-07) and confirmed the 0.3pt drop predates this session (daemon commits between 10-07 and today). I did NOT identify WHICH commit/code caused it. The promotion to the 92 floor proceeded with 0.8pt headroom instead of the rationale's assumed 1.1pt.
3. **ROADMAP resolved-item cleanup** — I removed the LSP duplicate and the shipped baseline idea, but the same duplication pattern remains for two sibling struck entries (art-dupl DECIDED line, erraudit RESOLVED line) that also summarize AGENTS.md content. Deliberate scope cut, now documented here.

## c) NOT STARTED

*(session-scope: items visible in TODO_LIST this session but blocked or untouched — nothing unblocked was skipped)*

1. Sandbox live-verification pass + OTT `2026Q3→2026Q4` flip probe — **blocked: `WISE_SANDBOX_API_KEY`**.
2. Sandbox integration-test first recorded run — blocked: same key.
3. v1.0.0 tag — blocked: user approval (irreversible).
4. `RequestLog.RetryAfterDelay` API freeze decision — blocked: user design decision.
5. Typed recipient `Details` — blocked: user design decision.
6. `CACHIX_AUTH_TOKEN` / `ERRAUDIT_TOKEN` secrets — blocked: user tokens.
7. CI re-enable (push + `gh workflow enable` + first green run) — blocked: push approval.
8. `--all-systems` darwin matrix decision — blocked: user decision.
9. Go 1.27 migration — blocked: 4 unlock conditions (nixpkgs stable 1.27 etc.).
10. Consume the benchstat baseline — deliberately not applicable (no perf-touching PR this session).
11. Cross-repo items (crush-config `references/lessons.md` recording; `check-skill-fanout.sh` doc/path mismatch) — different repo, untouched.

## d) TOTALLY FUCKED UP!

1. **My own verification pipe lied to me once:** I ran `buildflow ... | tail -6; echo "exit: $?"` — the reported `exit: 0` was **tail's** exit code, not buildflow's, while buildflow had actually FAILED (1 step). I caught it because the summary still said ERROR, then re-ran three times until I had a true `exit=0` with the failure explained (transient, passed on identical rerun). Class: pipeline masking — a known fleet failure mode I re-committed anyway. The fix habit (no pipes on exit-code gates) is not written down anywhere in this repo's AGENTS.md.
2. **The first buildflow run failed on something that predates me and nobody noticed since 2026-10-07:** `pin_sync.py`'s missing exec bit means **no full buildflow `--fix` run has been green since the pin-sync gate landed** (ruff-check-fix fails, 10 downstream steps blocked). Either nobody ran full mode with `--fix` for four days, or the failure was absorbed. Process smell, not just a file-mode smell.
3. **securitymd's error was also pre-existing and first surfaced in MY run** — same story: the findings gate would have tripped on any full run since the tool landed. Two independent "first green run in days" discoveries in one session says full-mode gate runs are too rare.

## e) WHAT WE SHOULD IMPROVE!

1. **Never pipe a gate command before capturing its exit code** (`cmd > log; echo $?` or `set -o pipefail`). Should become an AGENTS.md gotcha.
2. **Run `buildflow --build-mode full` (or at least `--fix`) after landing infra-touching commits** — the pin_sync/securitymd rot proves days-long blind spots otherwise.
3. **The coverage floor is a two-place split brain** (flake.nix + ci.yml), explicitly "mirrored" by comment. Today's promotion updated both in one change — the next one might not. Improvement: a check that greps both files for the same floor value (or a single shared source).
4. **`docs/bench/README.md`'s benchstat command is UNVERIFIED** — I wrote instructions I never executed (benchstat isn't confirmed installed in the devShell). Doc-claims discipline says: run it once and pin the output shape, or mark the command as requiring `nix-shell`/install.
5. **`buildflow doctor` for the "9 tools unavailable (health check failed)"** — stable across all runs this session, unexamined. Could be missing-on-PATH nix-run tools; unknown blast radius.
6. **statix reported 27 warnings (repeated-keys pattern in flake.nix) and vulnix 21 advisories (binutils/gcc/bison/coreutils drvs)** — all warning-severity, none gate-tripping, none triaged this session. The vulnix ones are toolchain-level, not repo code, but 7.5–8.6-high CVEs deserve at least a recorded disposition.
7. **art-dupl flagged my fuzz-target additions** (helpers_test.go:440/472, bench_test.go:97 — structural similarity between fuzz targets). Canonical run: 0 actionable. Accepted similarity class, but the next fuzz target should extract the shared fuzz-body skeleton if a fourth appears.
8. **ROADMAP still carries two struck resolved entries** (art-dupl DECIDED, erraudit RESOLVED) duplicating AGENTS.md policy bullets — same de-dup I did for LSP, unfinished on purpose.
9. **The coverage 0.3pt drop (93.1→92.8) is unattributed.** If it matters enough to document in AGENTS.md, it matters enough to bisect once.

## f) Up to 50 things we should get done next

*(honest count: 30; not padded to 50 — the remaining 20 slots would be invention)*

**Unblocked, bounded (start here):**
1. Verify the `benchstat` command in `docs/bench/README.md` actually runs in the devShell; fix or annotate.
4. Bisect the 93.1→92.8 coverage drop to a commit; recover or record.
5. Add a floor-parity check (flake.nix vs ci.yml coverage gate values must match).
6. Remove the two remaining struck resolved ROADMAP lines (art-dupl, erraudit) — finish the de-dup.
7. Live-fuzz the three seeded targets (5s each) to prove corpus+harness under mutation.
8. Run `buildflow doctor`; disposition the 9 unavailable tools.
9. Triage statix's 27 flake.nix warnings (likely generated-pattern false positives; document as deliberate non-fix if so).
10. Record the vulnix 21-advisory set's disposition (toolchain drvs, not repo code) in AGENTS.md or a review note.
11. Cross-repo: record the go-retry drift + Ginkgo `-count>1` + `gh --body-file` lessons in crush-config `references/lessons.md`.
12. Cross-repo: fix the `check-skill-fanout.sh` path/doc mismatch (crush-config).

**User-gated (one approval unblocks each):**
13. Provide `WISE_SANDBOX_API_KEY` → runs items 14–16.
14. Sandbox live-verification of the five spec-vs-live assumptions (P1).
15. OTT `2026Q3→2026Q4` probe + flip (P1).
16. First recorded sandbox integration run + CHANGELOG entry (P2).
17. Approve the v1.0.0 tag (P2; audit is green).
18. Decide `RequestLog.RetryAfterDelay` freeze vs observer interface (P2).
19. Decide typed recipient `Details` vs map freeze (P2).
20. Set `CACHIX_AUTH_TOKEN` + verify the `larsartmann` cache exists (P2).
21. Set `ERRAUDIT_TOKEN`, then delete the two `continue-on-error`/`if` lines in the erraudit CI job (P2).
22. Approve push + `gh workflow enable ci` + watch first run (P2; un-freezes the coverage badge).
23. Decide the darwin check matrix (pin nixpkgs-26.05-darwin vs drop x86_64-darwin) (P3 decision).

**Larger arcs:**
24. Go 1.27 migration when its 4 unlock conditions land (plan exists).
25. CI re-enable follow-ups: promote coverage floor gate off-machine, watch the badge job's first push-back.
26. Extract a shared fuzz-body skeleton IF a fourth similar fuzz target appears (defer until then).
27. `2026Q4` rolver ritual documentation (ROADMAP raw idea).
28. Typed `BadRequestError` (ROADMAP raw idea).

**From this session's observations (new, small):**
29. Add the "pipes mask gate exit codes" gotcha to AGENTS.md (from d.1).
30. Consider a weekly `buildflow --build-mode full --fix` cadence note in AGENTS.md Build & Dev (from d.2/d.3).

## g) Questions I can NOT figure out myself

1. **Can you provide a Wise sandbox API key (`WISE_SANDBOX_API_KEY`)?** It is the single blocker for the two P1 items (live spec-vs-live verification, OTT Q4 flip) and the P2 sandbox integration run. Without it those stay parked indefinitely.
2. **Do you approve tagging `v1.0.0` now?** The audit is green and the surface is frozen-worthy, but tagging is irreversible and TODO_LIST records it as needing your explicit approval. (Related, separable: `apidiff` suggested the unreleased additions are v0.13.1/v0.14 material — say the word and I'll cut the smaller release instead.)
3. **Do you approve pushing master and re-enabling the CI workflow?** ~6 local commits (including this session's) are unpushed; the workflow file is ready and needs no secrets. This also un-freezes the coverage badge and gives the promoted 92% floor an off-machine enforcer. I will not push without your approval.

---

*Point-in-time snapshot per session scope. Section (f) items 1–12 and 29–30 are new since TODO_LIST was last maintained this session; a HARVEST pass folded the bounded ones in. WAITING FOR INSTRUCTIONS.*
