# Status: Webhooks Support Review + Self-Review — 2026-08-28 08:37

Session scope: user asked "How is our Webhooks support?" — researched, verified, fixed one docs gap, reported. Then this self-review. **No other work was touched.**

---

## a) FULLY DONE

1. **Webhook support status assessed end-to-end.** Evidence gathered from `webhooks.go`, `internal_test.go`, `example_test.go`, `README.md` (Webhooks section), `ROADMAP.md`, three planning/status docs, and the authoritative OpenAPI spec.
2. **Verification-side capability confirmed:**
   - `ParseWebhookPublicKey` (`webhooks.go:27`) — PKIX + PKCS#1 PEM fallback, RSA-only, clear errors on garbage/non-RSA.
   - `VerifyWebhookSignature` (`webhooks.go:58`) — RSA-SHA256 PKCS1v15 over raw body; rejects malformed base64, empty signature, nil key.
   - `HeaderWebhookSignature` ("X-Signature-SHA256") + `HeaderDeliveryID` ("X-Delivery-Id") constants (`webhooks.go:16,22`).
   - README section covers raw-bytes-verification, 401/403 rejection, redelivery + `X-Delivery-Id` dedup semantics.
   - Godoc example `ExampleVerifyWebhookSignature` present.
3. **Webhook test suite verified green under `-race`** (valid / tampered / wrong-key / malformed / empty-sig / nil-key / empty-payload / 5 MiB / tampered-final-byte).
4. **Gap inventory extracted from the OpenAPI spec:** 9 webhook-subscription operations across 5 paths —
   - App level (client-credentials token): `POST|GET /applications/{clientKey}/subscriptions`, `GET|DELETE /applications/{clientKey}/subscriptions/{subscriptionId}`, `POST .../test-notifications`.
   - Profile level (user token): `POST|GET /profiles/{profileId}/subscriptions`, `GET|DELETE /profiles/{profileId}/subscriptions/{subscriptionId}`.
5. **FEATURES.md docs gap found and fixed:** webhooks were tracked nowhere; added a "Webhooks" section (3 FULLY_FUNCTIONAL, 2 PLANNED rows). `nix fmt` clean (0 changes needed).

## b) PARTIALLY DONE

1. **Typed-event inventory** — established that Wise documents webhook-event payload schemas under the `webhook-event` tag and that the SDK has zero typed decoding. I did NOT enumerate the full event-type list (my spec extraction only surfaced 6 referenced doc pages: transfers#state-change, cards#transaction-state-change, cards#card-production-status-change, cards#3ds-challenge, swift-in#credit, batch-payment-initiations#state-change). Good enough for a verdict, not good enough to scope an implementation.
2. **Planning-doc discrepancy noticed but not harvested:** implementation plan item #49 says "Webhook subscriptions — 8" endpoints; the spec has 9 operations. Reported in conversation; not corrected in the doc, not added to TODO_LIST.md.

## c) NOT STARTED

1. **Subscription management CRUD** (all 9 operations) — no raw types, no result types, no branded `WebhookSubscriptionID`, no client methods, no tests.
2. **Typed webhook event decoding** — no `WebhookEvent` envelope struct, no per-event payloads, no `ParseWebhookEvent`.
3. **`test-notifications` trigger support** (send dummy delivery to an existing subscription).
4. **README/API-reference + CHANGELOG entries for any of the above** (nothing to write yet).

## d) TOTALLY FUCKED UP (honest accounting — nothing broke, but three real misses)

1. **Imprecise "tests green" claim.** I ran only `-run 'Webhook'`, then said "tests green" in the summary. Webhook tests were green; the full suite was never run this session. My only change was markdown, so risk was ~zero, but the wording overstated verification.
2. **Walked past a split brain in the file I was editing.** FEATURES.md's "Out of scope (not yet started)" section contains a row "Statements (CSV/PDF) — FULLY_FUNCTIONAL" — a status/heading contradiction sitting directly under my new section. I saw it in the View output and didn't flag or fix it (pre-existing, but I was already in the file).
3. **Silently worked around broken build caches.** `/mnt/buildcache/go-build`, `/mnt/buildcache/go-mod`, and the golangci-lint LSP cache (`/mnt/buildcache/golangci-lint`) all fail with "no such device". I hand-rolled `GOCACHE`/`GOMODCACHE` env overrides and moved on without flagging the infrastructure issue, adding it to TODO_LIST.md, or documenting the workaround in AGENTS.md. This affects every future session's default `go test` / lint path.

## e) WHAT WE SHOULD IMPROVE (from this session's observations only)

1. **Say exactly what was tested.** "Webhook tests green under -race with cache overrides" ≠ "tests green".
2. **Fix-on-sight discipline applies to docs too:** when editing a doc file, contradictions in that file are in scope — fix or explicitly ticket them.
3. **Environment breakage is project knowledge:** workarounds for broken caches belong in AGENTS.md (Build & Dev section) the moment they're discovered, not carried as session-local shell folklore.
4. **FEATURES.md additions should trigger a TODO_LIST/ROADMAP cross-check** so PLANNED rows actually exist somewhere as actionable items (my two PLANNED rows are currently tracked only in FEATURES.md and the Tier-4 plan table).

## f) NEXT WORK (candidates, not commitments — webhook-focused unless noted)

### Webhook subscriptions CRUD (Tier-4 #49)

1. `internal/raw` wire types for `WebhookSubscription` (id, name, url, channel, event types, headers, created_at).
2. Public `WebhookSubscription` result type + branded `WebhookSubscriptionID` (`id.ID[WebhookSubscriptionBrand, int64]` — verify int64 vs UUID in spec first).
3. `CreateProfileWebhookSubscription(ctx, ProfileID, req)` + request type.
4. `ListProfileWebhookSubscriptions(ctx, ProfileID)`.
5. `GetProfileWebhookSubscription(ctx, ProfileID, WebhookSubscriptionID)`.
6. `DeleteProfileWebhookSubscription(ctx, ProfileID, WebhookSubscriptionID)` (204 handling).
7. App-level variants taking `clientKey` + client-credentials token story (see question 2).
8. `TestWebhookSubscription` — `POST .../test-notifications` (app level only per spec).
9. BDD tests per method (happy + 401/404 + validation).
10. Godoc examples + README "Webhook subscriptions" section + TOC entry.
11. Add new `.go` files to `flake.nix` fileset unions (Go + links checks).
12. CHANGELOG entry + version bump + release.
13. Correct plan item #49 "8" → 9 operations (or annotate).

### Typed webhook event decoding

14. `WebhookEvent` envelope: `data`, `subscription_id`, `event_type`, `schema_version`, `sent_at`.
15. `WebhookEventType` typed enum + constants (`WebhookEventTransfersStateChange`, `WebhookEventBalancesCredit`, …).
16. `ParseWebhookEvent(payload []byte) (WebhookEvent, error)` — decode-then-dispatch, tolerant timestamps via existing `parseWiseTimestamp`.
17. Enumerate the full `webhook-event` schema list from the spec (my 6-page extraction was incomplete).
18. Payload structs: transfers#state-change (resource + currentStatus), balances#credit, deposits#completed, deposits#top-up-failed as the high-value first four.
19. Decide: strict per-version schemas vs tolerant envelope + `json.RawMessage` data (design conversation).
20. Tests: fixtures per event type + unknown-event-type forward compatibility.
21. README: replace "decode the event JSON and process it" hand-wave with `ParseWebhookEvent` usage.
22. Optional: `wise.VerifyAndParseWebhookEvent(body, sig, key)` one-call helper (design decision — composition vs convenience).

### Docs/hygiene (session-found)

23. Fix FEATURES.md "Out of scope" section: move the Statements FULLY_FUNCTIONAL row out; heading currently lies.
24. Harvest the two new PLANNED rows into TODO_LIST.md as actionable tasks.
25. Document the GOCACHE/GOMODCACHE override for this machine in AGENTS.md Build & Dev (or fix `/mnt/buildcache` mounting).
26. Investigate why `/mnt/buildcache/*` is "no such device" (stale mount? missing volume?) and restore — the golangci-lint LSP is dead for the same reason (visible in every diagnostics block this session).
27. Re-verify CHANGELOG has the webhook verification helpers recorded under the right released version (v0.8.x-era; I did not check this session).

### Possible, lower certainty (flag, don't assume)

28. Consider a `WebhookHandler` http.HandlerFunc wrapper in the SDK (verify → dedup hook → parse) — could be YAGNI; consumer frameworks differ.
29. Expose delivery-attempt retry/backoff guidance constants if Wise documents redelivery cadence (not verified this session).

## g) QUESTIONS (cannot be answered from repo/spec alone)

1. **Token model for app-level subscriptions:** app-level endpoints require client-credentials tokens (`clientKey` from Wise tech support), but wise-go today is user-API-key only. Should the SDK stay user-token-only and ship just the 4 profile-level operations, or do you want the client-credentials flow (which pulls Tier-4 #36 `POST /oauth/token` into scope)?
2. **Priority:** ROADMAP's medium-term currently holds sandbox integration tests + account-requirements refresh. Should webhook subscriptions CRUD / typed events be pulled forward ahead of those, or keep Tier-4 ordering?
3. **Event-decoding scope if/when built:** minimal high-value set (transfers#state-change, balances#credit, deposits#*) or full coverage of every documented event type (cards, batch payments, swift-in, …) up front?

---

**State of the working tree:** one file modified this session (`FEATURES.md`, new Webhooks section), formatted, uncommitted (per no-commit rule — note the auto-git daemon may sweep it).
