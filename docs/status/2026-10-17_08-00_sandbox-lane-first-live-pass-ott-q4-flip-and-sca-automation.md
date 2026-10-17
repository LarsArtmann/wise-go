# Sandbox Lane Executed: First Live Pass, OTT Q4 Flip, SCA Automation

**Date:** 2026-10-17 08:00 CEST · **Session start:** ~06:30 (resumed from the 2026-10-16 08:52 TODO-execution session)
**Repo state at report time:** HEAD `594a23a`, tree clean, `go test ./...` + `golangci-lint run` green.
**Trigger:** user supplied the long-blocked `WISE_SANDBOX_API_KEY` (P1/P2 sandbox lane unblocked), then demanded full automation of everything, including a simulated browser for SCA.

---

## a) FULLY DONE

1. **Sandbox key delivered and secured** — user created the account (`1792209474679@test.dev` at `wise-sandbox.com`, pre-funded £1M-equivalent in GBP/EUR/USD/AUD) and a personal API token; stored at `~/.secrets/wise-sandbox.key` (mode 600, deliberately OUTSIDE the repo — the auto-git daemon must never sweep it into history). The token itself was pasted in chat (sandbox-only, accepted risk).
2. **First credentialed sandbox-live run — PASS** (`sandbox_live_test.go`, the P2 item carried since 2026-08-08): `GetMe` → id=200157260 "Franco Orozco"; `ListTransfers` → 0 (fresh profile, auto-pagination terminates); `GetTransfer` clean SKIP (`WISE_TEST_TRANSFER_ID` unset). `go test -run TestSandboxLive` exit 0.
3. **OTT `2026Q3 → 2026Q4` flip — EXECUTED end-to-end** (P1 item, parked since 2026-10-05):
   - Credentialed probe discovered the REAL discrimination: status code is 404 everywhere, but **body shape differs** — app-level `{"code":"ott_not_found",...}` on served prefixes vs router-level `{"error":"Not Found",...}` on unknown ones. Served window measured: **2026Q2..2027Q1**; controls `9999Q1`/`2027Q2` → router 404.
   - Constant flipped in `client.go` + every reference in lockstep (13 files): `ott.go` doc comments, `internal/raw/ott.go`, `ott_test.go` mock paths, `spec_conformance_test.go` example, README, FEATURES (7 rows), AGENTS (2 bullets rewritten with probe evidence), DOMAIN_LANGUAGE (row renamed "Quarterly versioned surfaces"), ROADMAP (2), CONTRIBUTING (ritual examples genericized), CHANGELOG `[Unreleased]` Changed entry, TODO_LIST item removed.
   - Ritual verified: repo-wide straggler grep clean (only factual "vendored at 2026Q3" snapshot notes remain), `go test ./...` green, `nix fmt` 0 changed.
   - **Q4 then verified with live data**: `GetOTTStatus` on `/2026Q4/one-time-token/status` returned action `BALANCE__GET_STATEMENT`, validity 59m59s, 1 pending challenge — the flip is validated by real responses, not just 404-shape.
   - Chose Q4 (not the spec-canonical 2027Q1): matches the documented TODO procedure, current calendar quarter, uniformity with `webhookSubscriptionsAPIVersion`; 2027Q1 canonicalization stays with the tracked spec-refresh item. Sandbox serving Q2..Q1 shows prefixes are long-lived aliases — low-flip-risk either way.
4. **Modernizer-sweep recurrence caught and repaired** (daemon commit `92bab25`, 05:55): deleted `.buildflow.yml` (106 lines, the skip_steps mitigations) and zeroed `max-issues-per-linter`/`max-same-issues` (50/10 → 0/0). Restored `.buildflow.yml` **byte-identical** from `92bab25^`, reverted the caps, kept the cosmetic quote normalization. Post-restore `go test` + `golangci-lint` green. Committed `594a23a` with the incident in the message. (`go.mod` untouched — `go 1.26.0` holds.)
5. **Assumption 4 of the 5-assumption pass — VERIFIED**: `/v1/rates` live array shape; 3 current + 3 historical pairs (EUR↔USD, GBP→EUR, USD→GBP) all decoded, `Time` non-zero both modes (the tolerant parser handles the live shapes).
6. **Assumption 5 — SDK side VERIFIED, wire fact pinned**: `FundTransfer` on a nonexistent transfer → 404; raw-HTTP probe pinned `Content-Type: application/json` with router-shape body `{"status":"404","error":"Not Found","message":"Resource not found"}` (NOT text/plain — the spec's text/plain applies to domain funding errors only). The SDK surfaces the body losslessly through the typed-error surface.
7. **SCA classification verified live, exactly as documented**: statements (CSV AND `statement.json`) on personal AND business profiles → HTTP 403, empty body, `x-2fa-approval-result="REJECTED"`, OTT issued → classified `*SCAChallengeError` with the actionable message. Business profiles are NOT SCA-exempt.
8. **Live-verification harness built** at `/tmp/wisecheck` (replace-directive module, subcommands: profiles/rates/balance/statement/txns/fund404/ott/retry/sca) — all evidence above is reproducible with one command each; key never logged.
9. **Headless-browser lane scaffolded** (user-sanctioned): chromium 155 via `nix build nixpkgs#chromium` (`/tmp/chromium-link`), `go-rod v0.116.2` module at `/tmp/wisebrowse`, login flow + statement navigation scripted, screenshots wired.
10. Session-resume bookkeeping: the 08:52 status report was still untracked — formatted and committed cleanly as `57abdc3`.

## b) PARTIALLY DONE

1. **Assumption 1 (statement `type` COMPACT/FLAT)** — both enum values are ACCEPTED by the live server (requests carry `type=COMPACT`/`type=FLAT` and reach the SCA gate; no type-related 400), but the statement CONTENT is unretrievable: the SCA challenge is **SIGNATURE** type with empty ViewData and **no phone alternatives** — the documented "clear by viewing a statement on wise.com" case. Content-level verification (row layout differences) pending SCA clearing.
2. **Assumption 2 (X-idempotence-uuid on balance creation)** — header accepted (server never complains about it; a missing required header would 400 on param validation), but **sandbox locks balance creation entirely**: 403 `non.existing.balance` for JPY/CAD STANDARD, GBP SAVINGS, on BOTH personal and business profiles. Same-key idempotency semantics therefore UNVERIFIABLE in sandbox.
3. **Assumption 3 (statement `details.type` enum)** — blocked behind the same SCA gate (`statement.json` = ListTransactions is equally SCA-protected); zero transactions visible without clearing.
4. **SCA automation (user demand "FUCKING automate it")** — three paths explored: (i) phone channels — dead by design (SIGNATURE has none; `TriggerOTT(sms)` → 404 `ott_challenge_not_found`); (ii) JOSE PIN endpoints (`/v2/profiles/{id}/pin` + `/pin/verify`) — dead: the key endpoints (`/auth/jose/response|request/public-keys`) require ClientCredentialsToken AND 404 router-level on sandbox; (iii) headless browser — built, first blind attempt FAILED (below), DOM-probe written but its run hung and was killed at report time. **Unresolved.**
5. **Permanent credentialed live tests** — evidence captured, but nothing added to `sandbox_live_test.go` yet (rates shape, OTT-status-on-Q4, funding-404, SCA-shape are all cheap key-gated test candidates).
6. **Lane recording** — CHANGELOG has the OTT-flip entry but NOT the first-live-run/assumptions entry; TODO_LIST P1 sandbox-pass item not yet resolved/struck; this report is the first written record of the assumption verdicts.

## c) NOT STARTED

1. v1.0.0 tag (user decision, carried).
2. crush-config push + home-manager activation (4+ local commits, carried).
3. 08:29 buildflow run findings review (carried from prior session).
4. Daemon-commit audit `93601d1..HEAD` beyond what this session touched.
5. `nix flake check` after the flake.lock nixpkgs bump (carried).
6. AGENTS buildflow-gotcha `lll` recurrence note (carried from prior session).
7. CI enable / cachix / erraudit tokens (user-gated, carried).

## d) TOTALLY FUCKED UP (all recovered)

1. **TODO_LIST edit regression** — while removing the OTT item I built a `new_string` that silently dropped 3 lines of the neighboring 5-assumption item (the assumptions list + source). Caught on the very next action and restored. Lesson (again): target the exact item text, never splice neighbors.
2. **Blind browser attempt** — guessed login selectors, no screenshot-first diagnosis: email+password "filled" but URL never left `/login` (SPA likely needs different triggers), then panicked on the 2FA selector timeout. Should have DOM-probed from the start; the probe then also hung (page load / `MustWaitStable` never resolved) and burned ~1.5h wall-clock in background before I killed it. Net: zero SCA progress via browser so far.
3. **Edit-tool discipline** — four failed edits from using `sed`/`head` output instead of the View tool (`you must read the file first`); plus a `.golangci.yml` edit where I typo'd the same key twice. Wasted round-trips, no damage.
4. **Harness cosmetics** — `%s` verb on int64 cents (noisy output), sloppy sed-surgery left a no-op `id = id` block in the biz-statement path (dead code in /tmp, rewritten properly in `retryStatement`).

## e) WHAT WE SHOULD IMPROVE

1. **Probe-body-shape > status-code for quarterly surfaces** — the TODO's "probe for non-404" criterion was wrong in practice (everything 404s); the app-vs-router body discriminator is the reliable signal. CONTRIBUTING ritual step 1 now says "non-404" — should be updated to document body-shape discrimination (I updated the AGENTS bullet but NOT CONTRIBUTING's step-1 wording).
2. **Parallel-session detection came late** — a v0.15.0-train session interleaved commits throughout my flip window; I only discovered it via a 132-file diff anomaly when verifying the daemon sweep. Cheap tell: unexpected commit count/log growth between consecutive commands. Check `git log` before EVERY repo commit window, not just after.
3. **Verify-then-amend was impossible here** — daemon heuristics + parallel commits landed between my edits and any commit attempt; the flip lives in two `chore: auto-commit` commits (2eff1e5, 34b753a) with no amendable HEAD. The CHANGELOG entry is the durable record. Consider: stage+commit in ONE bash invocation (add && commit) to shrink the race window.
4. **The SIGNATURE challenge is a human factor by design** — full automation requires full web-login automation (anti-bot risk, fragile). There is a real ceiling on "automate everything" here; the honest disposition is partial verification + documented residual (already the repo's language for such gates).
5. **Durable tests lag evidence** — every /tmp harness finding should land as a key-gated `TestSandboxLive*` case; the lane keeps producing throwaway proof instead of compounding suite value.

## f) NEXT (prioritized)

1. Re-run the browser DOM probe with hard timeouts + trace screenshots; fix the login flow (email→password→2FA `111111`), navigate to a statement view.
2. After UI approval: `go run . retry` → full COMPACT+FLAT statements + `details.type` census → assumptions 1+3 closed.
3. If browser stays hostile: record 1+3 as partial-verdict (params accepted; content SCA-gated), close the P1 item with the residual documented.
4. Add permanent key-gated live tests: rates shape, `GetOTTStatus` Q4, funding-404 body, SCA challenge shape (+ retry-with-OTT when cleared).
5. CHANGELOG `[Unreleased]`: first recorded sandbox-live run + the five assumption verdicts + the two new wire findings.
6. TODO_LIST: strike the P1 sandbox-pass item (with verdict), strike the P2 first-run item (CHANGELOG entry from 5).
7. AGENTS.md: sandbox-lane gotcha entry — SIGNATURE challenge shape, balance-creation locked (403 `non.existing.balance`), funding-404 content-type, OTT served window 2026Q2..2027Q1, probe body-shape discriminator.
8. `docs/error-codes.md`: add the observed top-level-`code`-only envelope variant (`non.existing.balance`) to the wire-shape notes.
9. CONTRIBUTING rollover ritual step 1: replace "non-404" with the body-shape discrimination rule.
10. Modernizer-sweep recurrence #5 note in the AGENTS buildflow gotcha (92bab25 → restored 594a23a; same signature as the 2026-10-07 incident).
11. `nix flake check` (post nixpkgs bump, carried).
12. Review the 08:29 buildflow findings (carried).
13. Daemon audit `93601d1..HEAD` — includes today's interleaved parallel-session commits (attribution table).
14. Kill/cleanup `/tmp` harness state: `/tmp/wise-ott.txt` (60-min TTL — expired by now), chromium out-link if disk matters.
15. v1.0.0 tag decision (question 2 below).
16. crush-config push + activation (question 3 below).
17. `nix run .#spec-drift` post-flip sanity (Q4 paths against live spec).
18. `nix run .#doc-verify` + `release-notes-check` after doc edits.
19. CI enable + first green run (runbook exists) — user-gated.
20. Grow `TestSandboxLiveGetTransfer` into the full quote→recipient→transfer→fund flow (P3 carried; now the only way to see domain funding errors).
21. Consider `simulation/balance/topup` + `simulation/bank-transactions/import` to seed statement data BEFORE clearing SCA (richer assumption-3 census).
22. Sweep: is the `/spend/profiles/{id}/3dsecure/challenge-result` surface relevant to the SCA story? (spec census only; low priority).
23. Bank-sync pin bump to v0.15.0 (P1, other repo).
24. AGENTS prune part 2 (P3 carried).
25. Security window watch: go 1.26.9 in nixpkgs (P3 carried).

## g) QUESTIONS (cannot resolve myself)

1. **SCA endgame:** keep investing in the headless-browser login (fragile, anti-bot risk, possibly hours) to close assumptions 1+3 fully — or accept the partial verdict (server accepts both statement types; content gated by a by-design human factor) and bank the lane? My recommendation: one more bounded browser attempt (~30 min), then partial-verdict.
2. **v1.0.0:** tag on the current tree (46 methods + error surface, all gates green) or hold for the sandbox-lane close + `RetryAfterDelay` freeze decision (P2)?
3. **crush-config:** push the local commits (2 lessons + doc fixes + any daemon strays) and run home-manager activation, or do you want to review the diff first? (Unpushed lessons are invisible to other sessions until activated.)

---

**Commit ledger (this session):** `57abdc3` (status report), `2eff1e5`+`34b753a` (daemon-swept OTT flip, content verified = exactly the 13 intended files), `594a23a` (sweep restoration). Parallel-session commits interleaved between `57abdc3` and `2eff1e5` (v0.15.0-train: a321f14, 03ad312, a84bc35, e9d1f41, 8bda014, 6c5f19a, 6714984, 29e3be7, 2606d3b, f234033 + daemon 92bab25).

**Key artifacts:** `~/.secrets/wise-sandbox.key` (600), `/tmp/wisecheck` (verification harness), `/tmp/wisebrowse` (go-rod login automation), `/tmp/wise-ott.txt` (expired OTT), screenshots `/tmp/wisebrowse/*.png` (pre-login only).
