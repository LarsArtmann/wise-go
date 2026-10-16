# Status Report — erraudit Enforcement Triage (2026-10-16 22:16)

Session scope: triage of the full-flag `erraudit lint` run on wise-go master, the
`--enforce-generic-return` ruling, and the AGENTS.md disposition. Written after the
fact; the session was interrupted mid-report by a new directive (see NOT STARTED).

## a) FULLY DONE

- Ran the user's exact full-flag command:
  `erraudit lint ./... --type-aware --enforce-go-error-family --no-suppress --enforce-samber-oops --enforce-generic-return --explain --disable-extensions --enforce-coded-errors --enforce-deferred-close`
  → 55 findings on my tree state, ALL `generic_return` (bare `error` returns).
- Attribution matrix (exit codes captured without pipes):
  - Canonical gate `erraudit lint ./... --type-aware --enforce-coded-errors` → **exit 0, clean**.
  - Same command minus `--enforce-generic-return` → **exit 0, 0 findings** (generic-return is the sole source).
  - `--enforce-generic-return` alone → 55 findings.
- Census of the 55: **41 exported client methods + 14 unexported helpers/mappers**
  (getRaw, request, doRequest, executeWithLogging, fetchByID, quoteAccountRequirements,
  decodeExchangeRates, listRecipientsPage, mapBalance, mapTransaction, mapExchange,
  listTransfersPage, toTransfer, mapTransfer).
- Grounded flag semantics in `erraudit --help`: generic-return is **off by default,
  "returning error is standard Go practice"** — an opt-in regime, not a default rule.
- Rendered `erraudit tree --inline-wraps --sort wraps`: 4 package-level sentinels,
  each with exactly 1 wrap site; ~130 remaining wrap sites target unnamed runtime
  errors (the APIError family) and cannot become hierarchy edges by design.
- Ruling recorded: **all 55 = deliberate non-fix as a class** (union error surface per
  method; `errors.AsType` consumption contract requires `error` returns; 41 exported
  signature changes would break consumers pre-v1.0 for zero capability).
- Disposition documented in `AGENTS.md` Gotchas beside the strong-id/samber-lo
  precedent entries.
- Noticed concurrent uncommitted changes from another session (`FEATURES.md`
  anchor refresh, `scripts/anchor_check.py` improvements) — read, left untouched.

## b) PARTIALLY DONE

- This status report (was interrupted; now written).
- The "other flags verified 0-finding clean" claim was verified WITH
  `--disable-extensions` active (mirroring the user's command). If built-in
  extensions add checks, the clean verdict holds only under disabled extensions.
  Re-verify without the flag during the next gate run.

## c) NOT STARTED

- **NEW DIRECTIVE (received 22:16, after this session's triage)**: "all my errors
  should have names and be defined in a central location (as it makes sense)" —
  i.e., name + centralize the error catalog (typed structs are already central in
  `errors.go`; the 4 sentinels are scattered per-file and unexported; ~130 wrap
  sites wrap unnamed class instances). Design + implementation begins immediately
  after this report. NOT the per-method return-type regime — that stays declined.
- Owner decision pending: adopt the verified-clean enforcement flags into the
  canonical erraudit gate + CI job, or keep the minimal pinned invocation.

## d) TOTALLY FUCKED UP

Nothing destroyed, but three honest misses:

1. **Un-reconciled count discrepancy**: the user's paste showed 56 findings, my run
   showed 55. Likely cause: the concurrent session edited `webhooks.go` between the
   two runs (its FEATURES.md diff moves an anchor webhooks.go:149→156). The census
   number written into AGENTS.md may have been stale on arrival. Not caught or
   flagged at the time.
2. **Verification-context caveat surfaced late**: I reported "all other flags
   0-finding clean" without noting `--disable-extensions` was active (see b).
3. First AGENTS.md edit attempt failed on read-before-edit (process slip; recovered
   by viewing the region first).

## e) WHAT WE SHOULD IMPROVE

- State the verification context (which flags were ON/OFF) when claiming "clean".
- Reconcile tool counts against tree state when another session is concurrently
  editing the measured files.
- When recording census numbers in AGENTS.md, timestamp them against the commit
  they were measured on.

## f) NEXT (session-derived, honest — not padded to 50)

1. Re-run `--enforce-generic-return` on a settled tree; reconcile 55-vs-56; fix the
   AGENTS.md census number if stale.
2. Error-catalog work per the new directive: central sentinel declaration file,
   export decision, naming, tests pinning `errors.Is` matches.
3. Re-verify "other flags clean" WITHOUT `--disable-extensions`.
4. Owner decision: adopt clean flags (`--enforce-go-error-family`,
   `--enforce-samber-oops`, `--enforce-deferred-close`, `--no-suppress`) into the
   canonical gate + ci.yml erraudit job.
5. Feed the empirically verified flag semantics back into the
   `go-error-modernization` skill's `references/cli-and-flags.md` (via its source
   repo — the fan-out symlink is read-only-by-policy).
6. Investigate what `erraudit extension` subcommand manages; whether any built-in
   extension would add findings here.
7. Confirm the 4 sentinels are matched in tests via `errors.Is` (tree reported
   "0 sentinel checks in function bodies" — test files may not be counted; verify).
8. Check whether the daemon auto-commit mixed this session's AGENTS.md bullet with
   the concurrent anchor-check session's files; annotate history if mixed.
9. README error-handling section: document the (new) named error catalog once it
   exists.
10. At v1.0 API freeze: re-review the generic-return decline (or permanently close).

## g) QUESTIONS (cannot figure out myself)

1. The concurrent session editing `FEATURES.md` / `scripts/anchor_check.py` — keep
   hands off entirely, or do you want it verified/reviewed once it lands?
2. Adopt the verified-clean enforcement flags into the canonical erraudit gate now
   (stricter, zero current noise), or keep the minimal pinned invocation until CI
   is re-enabled?
3. For the error-catalog directive: exported sentinels (`wise.ErrXxx`, consumers
   can `errors.Is`) is my default for an SDK — confirm, or keep them unexported
   with typed-error matching as the only consumer surface?
