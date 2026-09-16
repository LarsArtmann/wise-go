# Status: Docs-Health Full Audit (BUILD + HARVEST + VERIFY + ANNOTATE + ARCHIVE) — 2026-09-13 12:29 CEST

**Session scope:** user ordered "View ALL `**/2026-0*` files! Execute the docs-health SKILL! PROPERLY! FUCKING SUPERBLY!" with all seven living docs required superb and fully-done reports archived. This report covers ONLY that run. **No Go source was modified.**
**Format note:** written as `.md` per explicit user instruction (skill default is HTML).
**End state:** `go build` ✅ · `go test ./...` ✅ (README drift guard incl.) · `golangci-lint` 0 issues · coverage **91.3%** · `nix flake check` **all checks passed** · branch in sync with origin/master · daemon committed everything as `16f389d` (27 files, 8 renames verified R95–R100).

---

## a) FULLY DONE

1. **All 42 `**/2026-0*` files read and classified** — 31 markdown (status/planning/reviews) read completely; 7 HTML scan artifacts (data-model-review, code-quality-scan, naming-review, full-code-review, modularity, ASSESSMENT, nix-flake-migration proposal) and 4 D2/SVG diagrams classified point-in-time skill outputs → LEAVE ALONE per the "so what" test. Two sub-agents extracted per-item verdicts from the 17 oldest reports (May–Aug 8); every DONE verdict was grep-verified against current code before use.
2. **Deep code verification before every doc claim** — 33 public `*Client` methods (grep + manual name-list exclusion of 14 private helpers); CI state via `gh api` (`disabled_manually`, last run 2026-07-05 — the `on:` block in ci.yml is a red herring); v0.10.0 present in CHANGELOG/commit `fe896a8` but **untagged**; `failsafe-go v0.9.7` (docs said v0.9.6); `go 1.26.7`; flake on `go-standard` + `mkPreparedSource` + `git+ssh` inputs; `/mnt/buildcache` mount healthy again; `exhaustruct` nolint still at `transfers.go:205`; `fetchByID` still hand-rolls its zero check (`helpers.go:37`); no `TestRequireID` exists.
3. **Seven living docs rebuilt/fixed:**
   - **README** — status callout v0.8.1 → v0.10.0 with the full capability list (statements, receipts/MT103, webhooks, observability); "Transfer receipts & MT103" TOC entry added; new Features bullet for file downloads.
   - **FEATURES** — `ci.yml` row corrected to PARTIALLY_FUNCTIONAL with the GitHub-disabled evidence; `nix flake check` row updated to `go-standard`; godoc examples 18 → 20; the "Out of scope (not yet started)" split brain killed (Statements FULLY_FUNCTIONAL row deleted — it duplicated Transactions; heading → "Deferred (demand-gated, not started)").
   - **AGENTS** — CI entry rewritten (file refreshed `14523ae` but workflow disabled server-side; re-enable needs SSH auth for `git+ssh` inputs); coverage-badge gotcha now states the badge is FROZEN while CI is disabled; endpoint count 32 → 33 with the v1.0-audit-predates-v0.10.0 caveat; `nix flake check` wording → go-standard; failsafe-go v0.9.7; new gotcha: **`GetTransferReceipt` 404 = "no receipt yet"**.
   - **TODO_LIST** — fully rebuilt: ghost reference to nonexistent `docs/planning/2026-08-21_23-48_go-retry-cap-and-semantics-hardening.md` removed; "shipped 2026-08-21" trophy block deleted; new P1 (tag v0.10.0, v1.0 re-audit), P2 user-gated (sandbox key, v1.0.0 tag, typed recipients, Cachix token), P3 (webhook subscriptions CRUD, typed event decoding), P4 (go-retry, CI re-enable, govulncheck@1.26.7, dedup follow-ups, erraudit polish, quality micro-batch incl. `exhaustruct`→`exhaustruct_v5`).
   - **ROADMAP** — 31/32 → 33 methods everywhere; v0.10.0 release-state row added; receipts/payout in shipped list; Axis-4 resource list updated.
   - **DOMAIN_LANGUAGE** — `ProfileResult` → `Profile` (v0.4.0 rename), "unknown" removed from TransactionType, BalanceType lowercase-public/wire-UPPERCASE corrected; Money/Transfer/Quote/Recipient/SCA terms added.
   - **CHANGELOG** — `[0.10.0]` gained a Changed entry (toolchain 1.26.7, nix pins, CI file refresh).
4. **~20 historical reports annotated INLINE** (strikethrough + hash/decision evidence, per docs-health ANNOTATE): webhooks 08-28 (f.23–f.27 + b.2/d.2/d.3/e.4 resolved), dedup + nix-migration 08-21_23-10 (f-lists verdicted), hardening 22-31 (g3 resolved: v0.9.0 `e508572` shipped, pushes landed) and 21-46 (superseded, 8.x commit landed as `20810a7`), execution self-review 20-57 (**all 50 f-items verdicted inline**), docs-health 09-50 (**all 40 f-items verdicted**), both 08-19 reports (FundTransfer + full tier-2 struck `done at e508572`), 08-08 ×5, 07-05 (appendix-only resolution converted to inline markers on all 25 f-table rows — the skill's #1 failure mode), 05-17, 05-23, 07-18 comprehensive-improvement-plan, twelve-factor, api-docs-study, v1.0 audit (31→33 growth note), open-banking (stale V1 sandbox URL struck + corrected inline).
5. **8 fully-resolved files ARCHIVED via `git mv`** (history preserved): `docs/status/archived/` ← 05-21 ×3, 07-18, 07-23 · `docs/planning/archived/` ← 08-08_02-13, 08-21_12-05 (80/80 executed), 08-21_21-00 (60/60 executed). Pre-archive reference check: zero living-doc links to any archived path.
6. **Health report printed inline** — Accuracy 9.4/10 (44/47 verified claims), Fitness 9.5/10 (all 7 living docs role-correct, 0 split brains, 0 ghost refs).
7. **Quality gates green after every edit batch** — build/test/vet, lint 0 issues, coverage 91.3% (up from 90.8%), `nix flake check` all checks passed (README links + format + sandboxed race test).
8. **Concurrent v0.10.0 ship detected and absorbed** — a foreign process committed `fe896a8` (GetTransferReceipt/GetTransferPayoutInfo) mid-audit; re-read all living docs fresh, folded the new surface into every count/claim instead of clobbering.

## b) PARTIALLY DONE

1. **HTML artifact annotation** — classified and justified as LEAVE, but not item-by-item annotated; a future run could add one-line resolution banners to each of the 7 HTML files.
2. **Coverage badge** — local truth is 91.3%, badge shows the last CI-measured 86.9%. I deliberately did NOT hand-edit (README forbids it; CI is disabled so nothing regenerates it). The freeze is now documented in AGENTS, but the number stays stale until CI runs again.
3. **TODO_LIST final micro-edit** (exhaustruct_v5 addition) was still uncommitted at report time — the auto-git daemon will sweep it; content verified on disk.
4. **Method-count rigor** — counted via grep on `func (c *Client)` with a manual private-helper exclusion list; a `go doc -all` symbol-based count (like the v1.0 audit used) would be more rigorous for the 33 number.

## c) NOT STARTED

1. **Tag `v0.10.0`** — code + CHANGELOG complete (`fe896a8`), tag does not exist. User-gated.
2. **v1.0 re-audit** for the three post-audit methods (`RefreshQuoteAccountRequirements`, receipts, payout info) before the v1.0.0 freeze.
3. **CI re-enable** — SSH auth for the `git+ssh://` flake inputs (`go-nix-helpers`, `go-branded-id`, `go-error-family`) then flipping the GitHub workflow state; nothing set up yet.
4. **Webhook subscription CRUD + typed event decoding** — the two P3 features, spec-mapped but unbuilt.
5. **First credentialed sandbox run** — key still missing; workflow + test skeleton idle.
6. All P4 micro-batches (requireID tests, exhaustruct_v5 migration, erraudit polish, benchmarks/fuzz, WithUserAgent, Stringers, etc.) — consolidated in TODO_LIST, none started this session (out of scope for a docs-only run).

## d) TOTALLY FUCKED UP (honest accounting — nothing shipped is broken; these are my process failures)

1. **Sloppy multiedit on AGENTS.md** — edit 2 accidentally REPLACED the godoc-examples gotcha line with the receipt gotcha (deleting it); edit 3 repaired it in the same call. Net result correct, but I constructed a destructive intermediate state by careless batch building. One multiedit, three ops, one of which ate a line I then had to restore.
2. **Three failed edits from reading files via bash instead of the view tool** — CHANGELOG twice ("modified since last read": the daemon raced me AND `sed` output doesn't reset the edit tool's read-tracking) and open-banking once. Each was a wasted round trip; the fix (view immediately before edit) is the known daemon-race discipline and I still fumbled it.
3. **One 2-edit multiedit landed 1/2** on `2026-08-19_17-14` (exact-match miss on an already-struck line I was extending); required a re-view + separate edit.
4. **Wrote an edit against a guessed title anchor** on the v040 report ("# v0.4.0 Implementation — Status Report" — the real title is "Status Report — 2026-08-08 02:53"); the tool rejected it, but I should have viewed before constructing.
5. **Nearly believed the file-level CI signal** — reading ci.yml's restored `on:` block made "CI is enabled" look true; only the `gh api` workflow-state check (`disabled_manually`, no runs since July) revealed the truth. FEATURES' FULLY_FUNCTIONAL row had been written from exactly that file-level illusion. File content ≠ runtime state.
6. **Hand-rolled annotation edits instead of the skill's annotate scripts** (`annotate-rows.py`/`annotate-prose.py` with dry-run) — worked, but the scripts' atomicity/refusal semantics are the safer path for the table-row batches and I skipped them.

## e) WHAT WE SHOULD IMPROVE

1. **GitHub-state checks are step 1 for any CI claim** — `gh workflow list` / `gh api .../workflows` before every doc statement about CI; the repo now has one live incident of file-vs-state divergence enshrined in FEATURES.
2. **View-before-edit is mandatory on daemon-active files** — bash `sed`/`cat` reads do not reset edit-tool tracking; this session burned 3 round trips on it.
3. **Use the annotate scripts for table-row annotation batches** — dry-run first, per the docs-health skill; hand-built multiedits are where the AGENTS.md near-miss happened.
4. **Persist the surface-count derivation** — add the canonical count command (`go doc -all`-based) to AGENTS.md so the 31/32/33 drift class dies; counts get derived, never typed.
5. **Archive policy for HTML/D2 artifacts deserves one written line** (in AGENTS or a docs/README) so future runs don't re-litigate the classification each time.
6. **The v1.0 audit should be a living checklist, not a snapshot** — it went stale within 3 weeks (3 new methods); next audit run should end with "re-verify on every release until the tag".

## f) Up to 50 things we should get done next (impact-ordered; ⏳ = user-gated; most are already routed in TODO_LIST.md)

1. ~~⏳ **Tag `v0.10.0`** — code+docs shipped, tag missing (TODO P1).~~ done (tag exists (fe896a8); only the GitHub Release object remains (TODO_LIST P1))
2. ~~**Re-run the v1.0 audit** over the 33-method surface; refresh the risk register (TODO P1).~~ done (done 2026-09-13 — re-audit section in the audit doc)
3. ⏳ **Sandbox API key → first live run** + CHANGELOG entry (TODO P2).
4. ⏳ **Tag `v1.0.0`** after 1–3 (TODO P2).
5. ⏳ **Typed recipient `Details` decision** (typed structs vs map + `DetailsKey*` constants) — five reports old (TODO P2).
6. ⏳ **Set `CACHIX_AUTH_TOKEN`** + confirm the `larsartmann` cache (TODO P2).
7. **CI re-enable path**: SSH deploy key or `GITHUB_TOKEN`+`insteadOf` for the `git+ssh` flake inputs, verify `nix flake check` in CI, flip the workflow on (TODO P4) — unfreezes the coverage badge (f.16).
8. ~~**Re-run `govulncheck` on go1.26.7** — the 4 stdlib findings should now be zero (TODO P4).~~ done (done 2026-09-13 — zero findings on go1.26.7)
9. ~~**`requireID` direct table test** + pin its error contract in `internal_test.go` (TODO P4).~~ done (done 2026-09-13 — TestRequireID)
10. ~~**Route `fetchByID` through `requireID`** — kills the two-idiom split brain (`helpers.go:37`) (TODO P4).~~ done (done 2026-09-13 — fetchByID routes through requireID)
11. ~~**Add `SourceOfFundsOther`/`TransferNature` to `CreateTransferRequest`**, delete the `//nolint:exhaustruct` (`transfers.go:205`); verify field acceptance in the OpenAPI spec first (TODO P4).~~ **Won't implement — declined 2026-09-13 — spec-verified: create accepts exactly the five existing keys.**
12. ~~**Webhook subscription CRUD** — profile-level first; app-level needs the client-credentials token story (TODO P3).~~ done (done at 5a6448c (v0.11.0))
13. ~~**Typed webhook event decoding** — envelope + `ParseWebhookEvent` + fixtures (TODO P3).~~ done (done at 5a6448c)
14. ~~**Correct plan item #49 "8" → 9 operations** in `docs/planning/2026-08-19_wise-api-full-implementation-plan.md` (noted in TODO; the plan doc itself still says 8).~~ done (done 2026-09-13 — plan item #49 corrected)
15. ~~**Migrate `exhaustruct` → `exhaustruct_v5`** (golangci v2.13 deprecation warning, seen this session) (TODO P4).~~ done (done 2026-09-13 — exhaustruct_v5 + lint pin v2.13)
16. Coverage badge unfreeze — consequence of f.7.
17. ~~**`body, _ := readBody(resp)` error capture** (`client.go:382`) (TODO P4).~~ done (done 2026-09-13 — checkError surfaces unreadable bodies)
18. ~~**Raw-input values in `map*` error contexts** (TODO P4).~~ done (done 2026-09-13 — map* errors carry raw wire values)
19. ~~**`ErrorContext`/`IsRetryable` on `AuthError`/`NotFoundError`** (TODO P4).~~ done (decided 2026-09-13 — promoted contexts pinned by TestErrorContexts; only RateLimit/Server implement IsRetryable)
20. ~~**AGENTS.md error-context convention** entry (TODO P4).~~ done (done 2026-09-13 — AGENTS error-context convention)
21. **Curated erraudit config + CI gate** (TODO P4).
22. ⏳ **samber/oops adopt-or-decline** (user) (TODO P4).
23. ~~**`classifyTransactionType` float64 → cents** (`transactions.go:178`) (TODO P4).~~ done (done 2026-09-13 (8b54f7c))
24. ~~**`wiseDateFormat` constant** (`users.go:139`) (TODO P4).~~ done (done 2026-09-13 (8b54f7c))
25. ~~**`WithUserAgent` option** (TODO P4).~~ done (done 2026-09-13 (0fab6f5))
26. **`fmt.Stringer` for public enums** (TODO P4).
27. **`errorfamily.RegisterClassification` call** (TODO P4).
28. ~~**Surface `Profile.UserID`/`PublicID`** (TODO P4).~~ done (done 2026-09-13 (8b54f7c))
29. **Pin the gofumpt action version** (`ci.yml` uses `@latest`) (TODO P4).
30. ~~**Benchmarks** for hot paths (TODO P4).~~ done (done 2026-09-13 (dbab2e4))
31. ~~**Fuzz tests** for date/money parsing (TODO P4).~~ done (done 2026-09-13 — FuzzParseWiseTimestamp + FuzzNewCurrency)
32. ~~**Split the `wise_test.go` monolith** (TODO P4).~~ done (done 2026-09-13 (5f632e7))
33. ~~**Concurrent-safety test** (v050 retrospective item) (TODO P4).~~ done (done 2026-09-13 — Concurrency Describe (wise_test.go:246))
34. ~~**`internal/raw` test file** (TODO P4).~~ done (done 2026-09-13 — internal/raw/types_test.go)
35. **`wise.Version` constant** (TODO P4).
36. ~~**gorelease breaking-change CI check** (TODO P4).~~ done (done 2026-09-13 — nix run .#apidiff)
37. ~~**Coverage-threshold gate** (badge colors today, never fails) (TODO P4).~~ done (done 2026-09-13 — 90% gate in ci.yml (flake parity tracked in TODO_LIST P3))
38. ~~**README Date-Handling section** (TODO P4).~~ done (done — README carries the UTC date-handling notes)
39. ~~**Issue/PR templates** (TODO P4).~~ done (done 2026-09-13 — bug/feature/PR templates)
40. ~~**ADR directory** for the big decisions (Money, flat package, go-retry) (TODO P4).~~ done (done 2026-09-13 — ADR 001/002/003)
41. ~~**`doc-verify` flake app** (lychee + godoc freshness for contributors) (TODO P4).~~ done (done 2026-09-13 — nix run .#doc-verify)
42. ~~**`reports/jscpd-report.json` stale path cleanup** (TODO P4).~~ done (done 2026-09-13 — reports/ trashed)
43. **`GetStatement` PDF/XLSX content-type assertions** (TODO P4).
44. **`VerifyWebhookSignature` test against Wise's documented example signature** (TODO P4).
45. ~~**Decide `Authenticate()`'s future** now that `GetMe` exists (TODO P4).~~ done (done 2026-09-13 — same as f.41)
46. ⏳ **direnv/home-manager GOEXPERIMENT pin** (user machine) (TODO P4).
47. **`nix flake check --all-systems`** (aarch64/darwin never checked — warning visible this session).
48. ~~**One-line archive-policy note for HTML/D2 artifacts** (this session's e.5).~~ done (done 2026-09-16 — archive-policy line added to AGENTS.md)
49. **Design-story blog update** — the retry-typed-error find is the material (low priority).
50. ~~**HARVEST this report's f-list into TODO_LIST** — mostly pre-routed; f.14 (plan-doc 8→9 correction) is the one new item to fold in.~~ done (done 2026-09-16 — this docs-health run harvested the list)

## g) Three questions I cannot answer myself

1. **Tag `v0.10.0` now?** Code, CHANGELOG (`[0.10.0] - 2026-09-13`), README, and FEATURES all describe it as shipped, but the tag doesn't exist and `git tag` stops at v0.9.0. Do you want me to create and push the annotated tag (with the CHANGELOG as release notes), or do you tag manually? (Also: was the concurrent `fe896a8` session of yours, so I should treat its docs as canonical?)
2. **CI re-enable appetite:** should I set up the SSH deploy keys / `GITHUB_TOKEN`+`insteadOf` for the three `git+ssh://` flake inputs and flip `ci.yml` back on (it unfreezes the coverage badge and restores the cachix/govulncheck/lychee jobs), or does CI stay off and releases keep shipping on local gates? If re-enabling: which auth strategy do you prefer?
3. **Webhook subscriptions scope:** profile-level operations only (fits today's user-token client, 4 endpoints) — or pull the client-credentials OAuth token flow into scope (unlocks the 5 app-level endpoints + `test-notifications`, but drags Tier-4 #36 `POST /oauth/token` in)? This decides whether the P3 feature is a two-sitting or a multi-release job.

---

_Point-in-time snapshot. Status reports go stale; re-verify before acting on them. The auto-git daemon will pick up this file._
