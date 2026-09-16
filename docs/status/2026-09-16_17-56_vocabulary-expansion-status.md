# Status Report — Status-Vocabulary Expansion + Endpoint-Matrix Session (part 2)

- **Written:** 2026-09-16 17:56 CEST
- **Session scope:** Continues `2026-09-16_16-35_features-endpoint-coverage-matrix-status.md`.
  Delta since that report: FEATURES.md status vocabulary expanded from 5 to **8
  labels** with every label applied to real rows; coverage totals re-split and
  re-verified; gates re-run green. Plus noticed repo changes from the parallel
  session.
- **Format note:** `.md` per explicit user instruction (skill default is HTML).

---

## TL;DR

FEATURES.md now carries an 8-label status vocabulary — `FULLY_FUNCTIONAL`,
`PARTIALLY_FUNCTIONAL`, `BROKEN` (deliberately zero rows), `DISABLED` (new),
`PLANNED` (refocused: "completes a shipped family"), `DEMAND_GATED` (new),
`ON_HOLD` (new), `OUT_OF_SCOPE` — and every label is exercised by real,
verified rows. The old catch-all "51 PLANNED" split honestly into
**20 PLANNED / 26 DEMAND_GATED / 5 ON_HOLD**; endpoint arithmetic still
reconciles exactly at **215 = 41 + 20 + 26 + 5 + 123**. `nix fmt` and
`doc-verify` are green. Two things noticed from outside this session: the
parallel session is actively landing commits (latest: `8f6a8ad`), and that
commit bumped the go directive to ≥1.27.1 so **plain host `go test` no longer
starts** (`GOTOOLCHAIN=local`, host has 1.26.7) — Nix-based gates are
unaffected. Actual test-suite color (red 48 → ?) is now **unknown** and
unverifiable from the host shell.

---

## a) FULLY DONE

| # | Item | Evidence |
|---|------|----------|
| 1 | Vocabulary expanded 5 → 8 labels, each definition honest and non-overlapping (one question answered: does working code exist — and if not, why not?) | `FEATURES.md:6-19` |
| 2 | `DISABLED` created AND used: CI workflow row flipped from misleading `PARTIALLY_FUNCTIONAL` to `DISABLED`; new row added for credentialed sandbox tests (workflow dispatch-gated, skeleton key-drop-ready) | `FEATURES.md` Build & tooling |
| 3 | `DEMAND_GATED` created AND used: all 26 zero-shipped-family ops relabeled (batch groups 7, addresses 5, hold-limit ops 4, third-party transfers 2, returns 1, activity/contacts/payin-deposit/DD/bulk/GPI 7) | per-category tables |
| 4 | `ON_HOLD` created AND used: the 5 app-level webhook subscription ops (scope decision pending per ROADMAP Axis 1) | Webhook subscriptions table |
| 5 | `PLANNED` refocused honestly: now means "completes an already-shipped family, next on build path" — 20 ops (quote PATCH, recipient deactivate/compatibility/confirmations, transfers ×3, balance delete/movements, bank-details ordering ×3, MCA config ×3, profiles writes ×4, comparison) | per-category tables |
| 6 | `BROKEN` kept defined with explicit "zero rows is itself information" semantics — no dishonest usage invented to satisfy "use every label" | vocabulary table |
| 7 | Coverage arithmetic re-verified mechanically after relabel: method-shaped rows 41 F / 20 P / 7 DG / 5 OH / 19 O = 92; packed rows → totals **215 = 41 + 20 + 26 + 5 + 123** exactly | session grep audit |
| 8 | Summary bullets + summary-by-category "State" column updated to the new split (no stale "51 PLANNED" anywhere) | `FEATURES.md` coverage section |
| 9 | "Deferred (demand-gated, not started)" section renamed "Deferred architecture" — no label-collision with the new `DEMAND_GATED` | `FEATURES.md` (bottom) |
| 10 | Verification: `nix fmt` 0 changed (tables already canonical); `nix run .#doc-verify` **all checks passed** post-parallel-commit `8f6a8ad` (links 63 OK / 0 errors, count claims match compiled surface) | gate output |
| 11 | Work committed (daemon swept it into `00f9d1f`, content = 1 file: FEATURES.md) | git log |
| 12 | First multiedit batch had 3 failed matches (my own reconstruction errors); caught by audit grep, fixed with exact text, re-verified — no silent partial state | session audit |

## b) PARTIALLY DONE

| # | Item | Gap |
|---|------|-----|
| 1 | **HARVEST loop** — still not closed. The 16:35 report's (f) list and today's vocabulary split have not been routed into `TODO_LIST.md`/`ROADMAP.md`. Worse: the split *changed the data* — the 20 `PLANNED` ops are now the concrete near-term menu TODO_LIST should carry | waiting for your go |
| 2 | Test-suite color — **unknown**, and newly unverifiable on the host: `go test` aborts with "go.mod requires go >= 1.27.1 (running go 1.26.7; GOTOOLCHAIN=local)" after parallel commit `8f6a8ad` pinned the directive. `doc-verify` (Nix-provisioned Go) is green, but whether the 48 conformance failures are fixed was NOT re-checked | needs a Nix-based test run |
| 3 | Cross-file consistency — ROADMAP.md still uses informal "demand-gated" prose; FEATURES.md now has a formal `DEMAND_GATED` label with different, stricter meaning (whole-family-unshipped). Not yet reconciled | docs-health VERIFY would flag later |
| 4 | CHANGELOG — the doc overhaul (matrix + vocabulary) still unrecorded (deliberately folded into next code-touching commit due to the changelog-only-commit trap) | — |
| 5 | Plan-doc hygiene — `docs/planning/2026-08-19_…` still carries stale statuses (`GetQuoteAccountRequirements` PLANNED→actually DONE) and inverted method names vs code; ANNOTATE pass not done | — |

## c) NOT STARTED

| # | Item |
|---|------|
| 1 | Any of the 41 unimplemented-in-scope endpoint ops (20 PLANNED + 26 DEMAND_GATED; the 5 ON_HOLD await a decision) — zero implementation begun |
| 2 | Legacy-only surface decision (~20+ ops exist only under `api-reference/legacy`; matrix still covers preview only) — untouched, awaits your scope call |
| 3 | Pinned offline snapshot of the live preview reference in `docs/reviews/` |
| 4 | doc-verify extension to gate the coverage totals (41/20/26/5/123) like the method/Example counts |
| 5 | Legacy-vs-preview caveat line in FEATURES.md |
| 6 | Everything in the repo's pre-existing TODO_LIST (releases, CI re-enable, v1.0.0 tag, sandbox key, ADR 003, …) — deliberately untouched |

## d) TOTALLY FUCKED UP

Nothing destroyed — but three real problems, one of them new and foreign:

1. **The host toolchain broke out from under plain Go commands** (foreign, `8f6a8ad`): go.mod now demands ≥1.27.1, host has 1.26.7 with `GOTOOLCHAIN=local`, so `go test`/`go build` on the host shell fail before compiling a single file. Anyone following CONTRIBUTING-style raw `go test ./...` hits a wall; only Nix-provisioned toolchains work. My earlier session claim "tests: 154/48" is the last known color — current color unknown.
2. **Daemon attribution again**: the vocabulary expansion landed as `00f9d1f "chore: auto-commit 1 changed file(s) (heuristic)"` — a milestone-sized semantic change (status-language redesign) with a meaningless message. History does not tell the story.
3. **My first multiedit batch missed 3/17 edits** because I reconstructed `old_string`s from memory instead of copying exactly (double-pipe `||` typo, paraphrased CI row). Self-caught within one audit pass, but it's the same class of sloppiness as the earlier count errors: **verify-at-write-time beats verify-after**.

## e) WHAT WE SHOULD IMPROVE

1. **Never reconstruct edit targets from memory** — copy exact text from the immediately-prior view, even when the file "hasn't changed". The 3 failed edits cost a round trip.
2. **Close HARVEST in-session** — two reports in a row now end with "waiting for instructions" while (f)-items sit entombed. The vocabulary split makes this urgent: the 20 `PLANNED` ops are precisely the near-term backlog TODO_LIST should name.
3. **Re-check repo color after foreign commits land mid-session** — my "suite is red (48)" statement is now stale; a Nix-based test run is the only valid color source since the go-directive bump.
4. **Label governance**: the vocabulary is now expressive enough; the risk flipped from "too few labels" to "label drift" — ROADMAP/TODO_LIST prose should adopt the same terms (PLANNED vs DEMAND_GATED vs ON_HOLD) so cross-file status stays translatable.
5. **Fold doc milestones into explicit commits** (when authorized) instead of letting the daemon write "heuristic" history over semantic changes.

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

*Ranked by impact. (S)=session finding, (R)=pre-existing TODO_LIST/ROADMAP. Blocked items marked. Updated for the vocabulary split.*

| # | Task | Impact | Effort | Origin |
|---|------|--------|--------|--------|
| 1 | Re-establish a valid test-suite color via Nix (`nix flake check` or dev-shell `go test`) — host toolchain is bricked by the go-directive bump | High | S | S |
| 2 | Decide host-toolchain policy: bump host Go to 1.27.1 / drop `GOTOOLCHAIN=local` / document Nix-only Go — **needs your call** | High | XS | S |
| 3 | HARVEST: route the 20 `PLANNED` ops into TODO_LIST as the near-term menu; the 26 `DEMAND_GATED` + 5 `ON_HOLD` into ROADMAP with the new labels | High | S | S |
| 4 | Coordinate the conformance harness (parallel session landed `8f6a8ad` mid-flight): is it done/green now? — **needs your answer (Q3)** | High | S | S |
| 5 | Decide canonical surface: preview-only vs include ~20 legacy-only ops — **needs your answer (Q1)** | High | XS | S |
| 6 | Pick the first implementation target from the 20 `PLANNED` ops — my Pareto pick unchanged: quote PATCH + recipient deactivate/compatibility — **needs your answer (Q2)** | High | M | S |
| 7 | Implement `PATCH /v3/profiles/{id}/quotes/{id}` (update quote with recipient) | High | S | S |
| 8 | Implement `DELETE /v1/accounts/{id}` (recipient deactivate) | High | S | S |
| 9 | Implement recipient compatibility check | High | S | S |
| 10 | Implement recipient confirmations accept (`PATCH /accounts/{id}/confirmations`) | Med | S | S |
| 11 | Implement `GET /v1/transfers/{id}/payments` (funding payments list) | Med | S | S |
| 12 | Implement `POST /v4/profiles/{id}/balance-movements` (balance conversion) | High | M | S |
| 13 | Implement balance close (`DELETE …/balances/{id}`) | Med | S | S |
| 14 | Implement transfer NOC document (India FIRC) | Med | S | S |
| 15 | Implement US combined receipt PDF | Low | S | S |
| 16 | Implement bank-details ordering trio (issue local+international, orders create/list) | Med | M | S |
| 17 | Implement MCA configuration endpoints (available/payin currencies, eligibility) | Med | S | S |
| 18 | Implement profiles writes (personal/business create + update ×4) | High | M | S |
| 19 | Implement comparison (`GET /comparisons`) — tier-2 leftover | Med | S | S |
| 20 | Resolve the 5 `ON_HOLD` app-level webhook ops (option a/b, ROADMAP) — **BLOCKED on your decision** | Med | XS | R+S |
| 21 | Reconcile webhook event constants (33) vs live events (31); locate preview swift-in events | Med | M | S |
| 22 | Annotate the 2026-08-19 plan doc (stale statuses, inverted method names) via docs-health ANNOTATE | Med | S | S |
| 23 | Pin a live-reference snapshot into `docs/reviews/` for offline evidence | Med | S | S |
| 24 | Extend doc-verify to gate coverage totals (41/20/26/5/123) + fail on empty extraction | Med | S | S |
| 25 | Align ROADMAP/TODO_LIST prose with the new label vocabulary (PLANNED/DEMAND_GATED/ON_HOLD) | Med | S | S |
| 26 | README: one-line pointer to the FEATURES.md matrix; re-verify "16 resources" claim | Low | XS | S |
| 27 | CHANGELOG entry for the doc overhaul, folded into next code-touching commit | Low | XS | S |
| 28 | Publish GitHub Releases for v0.10.0 + v0.11.0 — **BLOCKED on your approval** | High | S | R |
| 29 | Tag v1.0.0 (audit green) — **BLOCKED on your explicit approval** | High | XS | R |
| 30 | Re-enable CI on GitHub (push + `gh workflow enable ci`) — **BLOCKED on your approval**; flips the DISABLED row to FULLY_FUNCTIONAL | High | XS | R |
| 31 | Provide `WISE_SANDBOX_API_KEY` → first credentialed sandbox run — **BLOCKED on you**; flips the new DISABLED row | High | S | R |
| 32 | Decide typed recipient `Details` — **BLOCKED on your design decision** | Med | XS | R |
| 33 | Set `CACHIX_AUTH_TOKEN` — **BLOCKED on you** | Low | XS | R |
| 34 | Accept/reject ADR 003 → go-retry v0.4.0 swap — **BLOCKED on ADR** | Med | M | R |
| 35 | Pin CI tools (`gofumpt`, `govulncheck`, `gorelease` — all `@latest`) | Med | S | R |
| 36 | Mirror 90% coverage floor into sandboxed `checks.test` checkPhase | Med | S | R |
| 37 | erraudit pass over webhook+OTT code + curated config + gate; settle samber/oops | Med | M | R |
| 38 | Add `.github/SECURITY.md` | Med | XS | R |
| 39 | CONTRIBUTING currency: `apidiff`/`doc-verify`, 90% gate, Ginkgo `-count` note — plus now the Nix-only-Go reality after the directive bump | Med | S | R+S |
| 40 | Quality micro-batch: statement PDF/XLSX content-type asserts; Wise example-signature test; subscriptions pagination shape; godoc example for `ListProfileWebhookSubscriptions`; `wise.Version` | Med | M | R |
| 41 | GOEXPERIMENT direnv/home-manager pin — **user-machine change** | Low | S | R |
| 42 | After harness lands: extend conformance coverage toward all 41 shipped endpoints (floors today: 25 templates / 60 exchanges) | Med | M | S |
| 43 | Decide SCA phone-number enrollment scope (OUT_OF_SCOPE today) — explicit Won't-implement or backlog | Low | XS | S |
| 44 | Direct-debit accounts + bulk settlement (rides on #20's client-credentials decision) | Low | M | S |
| 45 | Refresh/replace the 2026-08-08 local OpenAPI spec (or supersede via #23 snapshot) | Med | S | S |
| 46 | ROADMAP Axis-1 "Today" paragraph: point at FEATURES.md matrix as the live inventory | Low | XS | S |
| 47 | Add DEMAND_GATED families to ROADMAP long-term list explicitly (they're implied by tiers today) | Low | XS | S |
| 48 | Verify the daemon-swept commits (`00f9d1f`, `c29d985`, `b8204d4`) contain exactly what their files claim — heuristic history audit | Low | S | S |
| 49 | Consider `git commit` authorization for doc milestones so history stops lying (policy question, yours) | Low | XS | S |
| 50 | After v1.0 lock: re-run apidiff (`nix run .#apidiff`) to confirm the matrix's 41-method surface is what gorelease sees | Med | S | R |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Canonical surface:** preview-only (as instructed) or must the ~20+ legacy-only operations also be inventoried — matrix rows or a documented exclusion note?
2. **Next implementation target:** which of the 20 `PLANNED` ops do I build first? My pick stands: quote PATCH + recipient deactivate/compatibility (small, closes core-flow holes).
3. **Parallel session + host toolchain:** is the other session still actively landing commits (I keep hands off), and do you want the host Go / `GOTOOLCHAIN=local` situation fixed (bump host, or declare Nix-shell-only Go in CONTRIBUTING)?

---

*Waiting for instructions.*
