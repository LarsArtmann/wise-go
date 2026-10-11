# Archived status reports

Closed history. A report lands here only when every item is resolved — marked
inline as `~~item~~ done at <hash>`, `Won't implement`, `NOT-DO`, or routed to
TODO_LIST/ROADMAP. Do not open an archived file unless you need the original
wording for provenance.

## 2026-10-07 docs-health archive sweep (v0.12.0 era)

- `2026-09-16_16-45_openapi-spec-conformance-status.md` — ANNOTATE + ARCHIVE: the
  OpenAPI conformance gate shipped in v0.12.0; suite reached 202/202.
- `2026-09-16_18-53_adr003-executed-go-retry-v06-migration-status.md` — ANNOTATE +
  ARCHIVE: the failsafe-go -> go-retry v0.6.0 migration executed; recorded in
  CHANGELOG `[0.12.0]`.
- `2026-09-16_17-56_vocabulary-expansion-status.md` — ANNOTATE + ARCHIVE: the
  8-label status vocabulary is live and gate-checked in FEATURES.md.
- `2026-09-16_16-35_features-endpoint-coverage-matrix-status.md` — ANNOTATE +
  ARCHIVE: the per-endpoint coverage matrix is live and gate-checked in FEATURES.md.

## 2026-10-07 docs-health archive sweep (second run — P3 build-out)

- `2026-09-27_22-02_dedup-zero-actionable-and-gomod-restore-status.md` — RESOLVED +
  ARCHIVE: the dedup extractions and the `go 1.26` restore held through v0.12.0
  (its "0 actionable" snapshot was superseded same-night by the 23-42 report).
  Residuals (webhook label typing, art-dupl gate, erraudit policy) routed to
  TODO_LIST/ROADMAP/AGENTS.
- `2026-09-27_23-42_dedup-pass2-gomod-flipflop-rootcause-status.md` — RESOLVED +
  ARCHIVE: the pass-2 extractions (`quoteAccountRequirements`,
  `decodeWebhookEvent[T]`) and the in-body directive placement are stable; the
  go.mod flipflop is settled at `go 1.26`. Residual decisions routed.
- `2026-09-29_07-37_README-webhook-snippet-validation-fix.md` — RESOLVED +
  ARCHIVE: the README webhook block passes `md-go-validator` and the check is
  wired as `checks.md-go-snippets`. Residual (Example alignment, upstream
  feedback, archived-doc scan policy) routed.
- `2026-10-05_14-53_strong-id-analysis-execution-status.md` — SHIPPED + ARCHIVE:
  the `CustomerTransactionID` / `OTTStatus.UserID` / `WebhookResource.ProfileID`
  brands landed in v0.12.0; the 30 declines are pinned in AGENTS.md. Residual
  (UUID validation depth, OTT `userId` ID-space check, `AccountID` naming)
  routed to TODO_LIST/ROADMAP.
- `2026-10-05_15-52_branching-flow-triage-panic-suppression-status.md` — SHIPPED +
  ARCHIVE: the two `//nolint:branching-flow:panic` suppressions are live and
  verified. Residual (SARIF baseline, generic-nolint test, `IdempotencyKey`
  brand) routed.
- `2026-10-05_21-47_pareto-todo-execution-status.md` — SHIPPED + ARCHIVE: all §a
  items reached v0.12.0 and the rollover questions resolved 2026-10-06; the
  residual quality items landed in the P3 sweep. Routed.
- `2026-10-06_14-16_q4-flip-execution-and-gates-status.md` — SHIPPED + ARCHIVE:
  the 2026Q4 webhook flip and the two never-run gates landed in v0.12.0; all §g
  questions resolved. Routed.
- `2026-10-07_03-57_session-status-unblocked-prep-release-readiness.md` — SHIPPED +
  ARCHIVE: the release prep landed (v0.12.0 tagged; Release objects published);
  superseded by the release session. Residual user-gated items routed to
  TODO_LIST.
