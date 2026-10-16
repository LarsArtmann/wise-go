# Archived status reports

Closed history. A report lands here only when every item is resolved — marked
inline as `~~item~~ done at <hash>`, `Won't implement`, `NOT-DO`, or routed to
TODO_LIST/ROADMAP. Do not open an archived file unless you need the original
wording for provenance.

## 2026-10-16 docs-health archive sweep

- `2026-10-07_04-57_v0120-release-session-status.md` — RESOLVED + ARCHIVE:
  v0.12.0 shipped and verified end-to-end; every §b–§g follow-up carries an
  inline verdict from the 2026-10-16 maximalist pass (pre-release app,
  release-notes-check, shuffle-proof guard, RetryAfterDelay, fuzz/bench all
  shipped; credential-gated and user-gated residuals routed).
- `2026-10-07_06-07_p3-quality-tooling-sweep-status.md` — RESOLVED + ARCHIVE:
  the P3 tooling sweep landed (CAMT test, webhook fuzz/bench, conformance
  guard, coverage batch, md-go-snippets, release-notes-check, pre-release);
  the 2026-10-16 pass struck the remaining §b–§g items inline.
- `2026-10-07_06-34_p3-sweep-resume-defect-filing-lsp-rootcause-status.md` —
  RESOLVED + ARCHIVE: four tool-repo defects filed (#34 closed and absorbed,
  #35/#36 + md-go-validator #8 watch-gated), LSP phantom root-caused and
  fixed; the 2026-10-16 pass itemized the previously banner-granular §f.
- `2026-10-07_07-09_p3-resume-execution-prerelease-green-status.md` — RESOLVED +
  ARCHIVE: first end-to-end pre-release green + `--shuffle` verified; CI
  wiring and meta.description shipped since; user-gated residuals routed to
  the consolidated ask.
- `2026-10-07_07-35_go-retry-pin-drift-fix-status.md` — RESOLVED + ARCHIVE:
  the pin drift was fixed in-session (`3ed4fb4`); the class was closed
  structurally by `checks.pin-sync` + the peeled-commit realignment in v0.13.0
  (`7368a21`); the cross-project lesson landed in crush-config (`0607ed3`).
  Residuals (tag-pin strategy, fleet sweep, upstream proposals) routed to
  TODO_LIST/ROADMAP.
- `2026-10-07_08-22_v0.13.0-release-pin-sync-drift-discovery-status.md` —
  RESOLVED + ARCHIVE: v0.13.0 shipped; the 2026-10-11 session landed the
  fuzz/bench/typing tail; the `.buildflow.yml` erraudit-skip question was
  answered (BuildFlow #34 closed — the removal was correct, verified
  2026-10-16). The one open seed (hermetic tag-pin strategy) is routed to
  TODO_LIST as a user decision.

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
