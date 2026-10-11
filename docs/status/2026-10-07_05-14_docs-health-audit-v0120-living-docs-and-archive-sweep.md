# Status Report — docs-health full audit (v0.12.0 living docs + annotate/archive sweep)

> **Superseded (2026-10-07 docs-health pass, second run):** this audit's §f
> follow-ups landed — `md-go-snippets` wired as `checks.md-go-snippets`,
> `release-notes-check` and `pre-release` apps shipped, `doc-verify` extended to
> `docs/releases/*`, coverage raised to 93.1%, the spec-conformance guard made
> shuffle-proof, and the LSP phantom root-caused. The remaining §g PR/policy
> questions are standing. Retained as the second docs-health audit record (see
> also `2026-09-16_15-30`, the canonical convention record).

- **Date:** 2026-10-07 05:14 CEST
- **Repo:** wise-go @ master (auto-commit daemon active; a **concurrent release session** was also editing docs mid-run)
- **Scope:** This session only — the operator ordered: "View ALL `**/2026-0*` files! Execute the docs-health SKILL! TODO_LIST/CHANGELOG/AGENTS/README/ROADMAP/FEATURES superb; archive fully-done + inline-struck reports." **No Go source was modified.** Findings only from this run.
- **Format note:** `.md` per explicit user instruction (status-report skill's canonical format is HTML — one-off override, same as the 2026-09-13/09-16 reports).

---

## a) FULLY DONE (with evidence)

| #  | Item                                                                                                                                                                                                                                                                                                                                                        | Evidence                                                     |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| 1  | All 2026 point-in-time reports viewed — 17 strict `2026-0*` `.md` + 9 October `.md` (26 total); 8 generated HTML/D2 artifacts classified **LEAVE-ALONE** per AGENTS.md                                                                                                                                                                                      | agent extractions + direct reads; `find`/`grep` inventory    |
| 2  | **VERIFY**: all six living docs checked against code — 41 exported `*Client` methods, 25 `Example*` funcs, `go 1.26.0`, `const Version = "0.12.0"`, 33 `WebhookEventType` constants, `webhookSubscriptionsAPIVersion=2026Q4` / `ottAPIVersion=2026Q3`                                                                                                       | `go doc`/grep counts                                         |
| 3  | `CHANGELOG.md`: `wise.Version` entry `"0.11.0"` → `"0.12.0"` (the `[0.12.0]` release bumps it — was self-contradictory)                                                                                                                                                                                                                                     | edit                                                         |
| 4  | `ROADMAP.md`: release state now lists **v0.12.0** (2026-10-07, Releases published); shipped-list bullet, v0.x ledger, and 6 new raw ideas (Go 1.27 plan, md-validator gate, art-dupl gate, erraudit policy, daemon/`go.mod` guard, release-lint)                                                                                                            | edits; `doc-verify` "ROADMAP 41 methods" ok                  |
| 5  | `FEATURES.md`: +3 rows (OpenAPI conformance gate, `apidiff`, SECURITY.md); fixed stale `types.go:640-672` → `644-678`                                                                                                                                                                                                                                       | edits                                                        |
| 6  | `AGENTS.md`: +2 new gotchas (docs fenced-Go snippet shapes; three evidenced BuildFlow upstream defects); SDK-surface clause → "unchanged at 41 through v0.12.0"; +`md-go-validator` Build & Dev line                                                                                                                                                        | edits                                                        |
| 7  | `TODO_LIST.md`: **deleted all 10 completed `[x]` items** (they now live in CHANGELOG per docs-health) → **0 `[x]` remain**; reconciled the now-DONE GitHub-Release item; pointer updated                                                                                                                                                                    | `grep -c '^\[x\]'` = 0                                       |
| 8  | `TODO_LIST.md`: HARVESTED ~10 new open items from the recent reports (md-validator gate, doc-verify→`docs/releases`, coverage headroom, LSP phantom, rollover ritual, BuildFlow defects, art-dupl/erraudit/Go-1.27 decisions, CAMT test, benchstat, fuzz webhooks, `#pre-release` app, release-notes lint, shuffle-guard, Retry-After log, `--all-systems`) | edits                                                        |
| 9  | `README.md` verified fresh (v0.12.0 status line, 41 methods, six file formats, three deps)                                                                                                                                                                                                                                                                  | read + spot checks                                           |
| 10 | `docs/releases/`: fixed the stale v0.10.0 draft header and a **broken inline code span** in the published v0.12.0 body (`Type:` split across a newline)                                                                                                                                                                                                     | edits                                                        |
| 11 | **ANNOTATE**: dated resolution banners with inline `~~…~~ done` markers on 15 files — 11 status reports, 2 reviews (twelve-factor, api-docs-study), 2 planning docs (2026-07-18, 2026-08-19)                                                                                                                                                                | edits; `grep '~~'` present on all active reports except none |
| 12 | **ARCHIVE**: 5 fully-resolved reports `git mv`'d — `status/archived/` {16-35 coverage matrix, 16-45 conformance, 17-56 vocabulary, 18-53 adr003} + `planning/archived/` {16-54 conformance plan}; wrote two `archived/README.md` manifests                                                                                                                  | `git ls-files`, `ls`                                         |
| 13 | Gates green after the sweep: `nix flake check` **all 8 checks passed**; `nix run .#doc-verify` **all count-claims + 64 links OK**; `go test ./...` green; archived completeness gate `grep -rLn '~~'` **clean**; internal link scan **0 broken**                                                                                                            | tool output                                                  |

## b) PARTIALLY DONE (what works, what remains)

1. **Inline-strikethrough depth on the kept reports.** Works: every kept report's _top claim_ is struck inline and a dated banner carries the verdict + routing. Remains: the interior `§a/§b/§f` item lists are resolved at **section** granularity, not item-by-item (`~~item~~ done at <hash>` per row). The 2026-09-16 precedent struck item-by-item for the archived set; the kept set here got section-level treatment.
2. **Archive manifests.** Works: `docs/status/archived/README.md` + `docs/planning/archived/README.md` list filename → classification → reason. Remains: they are hand-written, not the skill's `assets/` scripts.
3. **HARVEST coverage.** Works: the actionable, bounded items are in `TODO_LIST.md`. Remains: many `(f)` 50-row "brainstorm" lists in the kept reports are not individually routed — judged ROADMAP fuel, but a strict reading says surfacing them is incomplete.
4. **Concurrent-session handling.** Works: I detected the other session and hand-reconciled two collisions (its stale ROADMAP clause, its broken release-notes span) without clobbering its work. Remains: no protocol — it was monitoring + timing, not a claim convention.

## c) NOT STARTED (observed/known, deliberately untouched this session)

1. Wire `md-go-validator` into `nix flake check` or `.buildflow.yml` (routed to TODO_LIST P3).
2. Extend `doc-verify` count-claims to `docs/releases/*.md`.
3. Coverage headroom: 90.1% vs the 90.0 floor.
4. Root-cause the `golangci_lint_ls` phantom on `webhooks.go:142`.
5. Write the quarterly-surface rollover ritual into CONTRIBUTING.
6. File the three BuildFlow upstream defects (erraudit FPs, reinstall profile-lock, file-size 0-scan).
7. `art-dupl` enforced-gate / erraudit class-wide policy / Go 1.27 plan decisions.
8. Any Go source change (none — docs-only session, as scoped).

## d) TOTALLY FUCKED UP (radical honesty)

1. **Edit-tool vs daemon races, twice.** Edits to `2026-09-16_16-35` and `2026-09-16_18-53` were refused ("modified since last read") because the auto-daemon reformats/commits between my read and write. Known daemon-era footgun (AGENTS.md documents it); I still hit it. Cost: one wasted round-trip each.
2. **A claim written before the event settled.** My ROADMAP release-state edit said "GitHub Release objects pending user approval" — true when written, false ~minutes later when the concurrent session **published** them. I caught it only because that session's own report flagged the stale clause. Violates "don't encode time-sensitive claims mid-flight."
3. **Delegated reading of 5 reports to sub-agents.** For throughput I used agents to extract structured item lists; 4 of 5 timed out and only one returned. I then read the two planning files directly, but the agent-summarized reports were not line-read by me — a completeness risk for ANNOTATE that I only partially mitigated by re-verifying counts against code.
4. **Archived-file citations left dangling in one place, fixed lazily.** `TODO_LIST` cited `docs/planning/2026-09-16_16-54_…` (now `archived/`); I updated that one. Other living docs citing moved reports were checked but the sweep was opportunistic, not exhaustive.

## e) WHAT WE SHOULD IMPROVE

1. **Daemon-aware edit loop:** View immediately before every Edit; if refused, re-View and retry (no guessing). Treat it as normal, not exceptional.
2. **Freeze-release state:** never write "pending approval"/"unreleased" claims into living docs while a release/release-publish is in flight; write the outcome.
3. **Use the skill's scripts for bulk annotation** (`annotate-rows.py` / `annotate-status-items.py`) instead of hand-authored banners when a section has 10+ numbered rows.
4. **Read reports myself for ANNOTATE.** Agent summaries are fine for triage, not for verbatim-resolution edits — the precision loss is real.
5. **A session-start foreign-edit check:** `git status` + mtime scan before touching shared docs, so concurrent-session collisions are safe by construction.
6. **Re-run the headline gate after the daemon's next commit sweep** (it committed my work mid-session; verify the committed content matches intent).

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

**This session's direct follow-ups**

1. Run `nix run .#doc-verify` once more after the daemon settles HEAD (confirm committed content).
2. Review the daemon commits from this session (`d44707e`, `0a01123`, `57d499a`, `553f385`) with `git show --stat` — confirm they contain exactly my doc edits.
3. Push local master (ahead of origin) once approved — docs-only post-tag commits.
4. Item-by-item strikethrough pass on the highest-traffic kept reports (2026-10-05/06/07 series).
5. Update remaining living-doc citations that point at freshly-archived paths (audit `ROADMAP`/`AGENTS`/`FEATURES`).
6. Sweep `docs/status/` once more for any other fully-resolved report (annotate-then-archive).
7. Convert the two hand-written archive manifests to the skill's scripted form if desired.

**Docs-health tooling**
8. Wire `md-go-validator` into `nix flake check` (README/docs fenced-Go parse gate).
9. Extend `doc-verify` count-claims to `docs/releases/*.md` method-count lines.
10. Add a release-notes lint (split code spans + relative links) before `gh release create`.
11. Add a `nix run .#pre-release` gate app chaining build/vet/race/lint/flake/doc-verify/apidiff.

**Quality / repo hygiene (harvested into TODO_LIST P3)**
12. Raise coverage headroom to ~92% before the next feature PR.
13. Root-cause the LSP golines phantom vs CLI `golangci-lint`.
14. Add the CAMT (`.xml`) `GetStatement` test.
15. Commit a `benchstat` baseline.
16. Fuzz `VerifyWebhookSignature` + `ParseWebhookPublicKey`; bench the webhook verify path.
17. Fuzz `decodeExchangeRates`; bench the rates array-decode path.
18. Make the spec-conformance coverage guard assert floors under `-shuffle` instead of skipping.
19. Add a `go-retry` `Retry-After`-honored log line.
20. `nix flake check --all-systems`.
21. Write the quarterly-surface rollover ritual into CONTRIBUTING.

**Decisions pending**
22. `art-dupl` enforced-gate decision.
23. erraudit class-wide policy (fix vs suppress-with-rationale).
24. Go 1.27 migration plan (toolchain + `GOTOOLCHAIN` policy).
25. `WebhookResource.AccountID` naming revisit (spec: "recipient/balance account ID").

**Upstream (BuildFlow repo)**
26. File embedded-erraudit false positives (29 at HEAD).
27. File `nix run .#reinstall` profile-lock bug.
28. File `file-size-check` 0-scan bug for root-package layouts.
29. Align `go-version-auto-configure` vs `go-mod-normalize` vs `go-structure-linter:repair`.

**Release / CI (mostly user-gated)**
30. Set `ERRAUDIT_TOKEN` → make the erraudit gate blocking.
31. Set `CACHIX_AUTH_TOKEN`.
32. Re-enable CI on GitHub + watch the first green run.
33. Run `nix run .#apidiff` vs v0.11.0 and archive the delta to `docs/releases/v0.12.0-apidiff.md`.
34. Verify the v0.10.0/v0.11.0 published note bodies against their tags.
35. Decide `v1.0.0` timing (audit green; gated on approval).

**Endpoint expansion (demand-gated, from the FEATURES matrix)**
36. `DELETE /v4/profiles/{id}/balances/{id}` (close balance).
37. `POST /v4/profiles/{id}/balance-movements`.
38. Bank-details ordering ops.
39. Profile write ops (personal/business create/update).
40. Quotes `PATCH`.
41. Recipients deactivate / compatibility / confirmations.
42. Remaining standard-transfer read ops.
43. MCA configuration op.
44. Comparison op.
45. Decide app-level webhook subscriptions (client-credentials mode).

**Long-tail / raw ideas**
46. `WithMetrics` hook.
47. Typed `BadRequestError`.
48. Webhook end-to-end README quickstart.
49. `ParseWebhookEvent` RFC3339 fast path.
50. Document the repo's release ritual (CI-disabled exception, apidiff-before-tag, no `--prerelease`).

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF (3)

1. **Push approval.** Local `master` is ahead of `origin` with docs-only, post-tag commits (my sweep + the concurrent session's cleanup). Push them now, or hold?
2. **Archived-citation policy.** When a report is `git mv`'d into `archived/`, should living-doc citations be rewritten to the `archived/` path (I did this for one), or left as historical references to the original location?
3. **Concurrent-session protocol.** Another session was editing `AGENTS.md`/`FEATURES.md`/`ROADMAP.md` while I ran. Do you want me to (a) keep hand-reconciling by watching `git status`/mtimes, (b) serialize by claiming files, or (c) treat concurrent doc edits as out-of-scope and stop?

---

_Point-in-time snapshot; goes stale. Feed section (f) into `docs-health` HARVEST; annotate (don't rewrite) when bringing this report current._
