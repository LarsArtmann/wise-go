# CI First-Run Watch Package (v0.15.0 release train, 2026-10-17)

Everything CI needs except the one user action. Verified against the final
`ci.yml` at the v0.15.0 tag (commit `415cce4`): YAML parses clean, 8 jobs,
flat graph (no `needs`), all actions pinned by SHA, tool versions pinned in
`env` (go 1.26, golangci-lint v2.13, gofumpt v0.12.0, govulncheck v1.8.0,
erraudit pseudo-version `v0.5.1-0.20260922…`), `GOEXPERIMENT=jsonv2` global.

## 1) Enable (the one user word)

```sh
gh workflow enable ci
```

Then push any non-docs commit to master (docs-only pushes are filtered by
`paths-ignore`, and tag pushes do not trigger `on: push`). Expected first-run
shape: 9 jobs fan out — `build-and-test`, `lint`, `govulncheck`, `erraudit`
(warn), `art-dupl` (warn), `coverage-badge` (master push only), `nix`,
`pin-sync`, and `fuzz-smoke` (3s per fuzz target, auto-discovered). Local pre-release exit 0 at the tag is the expectation baseline;
the runbook below covers the CI-only failure classes.

## 2) Triage runbook (CI-only failure classes)

These fail on CI but pass locally; none of them indicate an SDK defect.

| Class       | Signature                                                                           | Fix                                                                                                                                   |
| ----------- | ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| Env         | `GOEXPERIMENT` missing → `bytes imported and not used` / `undefined: json` phantoms | Already pinned workflow-wide (`env.GOEXPERIMENT: jsonv2`); if a new job is added it inherits, do not re-declare per-step.             |
| Permissions | `coverage-badge` push rejected (403)                                                | Workflow `permissions: contents: write` is set at job level; check repo Settings → Actions → Workflow permissions = "Read and write". |
| Permissions | `erraudit` install fails                                                            | Expected without `ERRAUDIT_TOKEN` secret; job is `continue-on-error` by design. Becomes blocking only after the secret exists.        |
| Toolchain   | `requires go1.27 or later`                                                          | A dep bumped past the 1.26 pin; check `grep '^go ' go.mod` first (the go-directive drift class), then pin the dep back.               |
| Cache       | golangci-lint OOM/timeout at 5m                                                     | Re-run once; persistent → bump `--timeout`, never lint-config.                                                                        |
| Cachix      | Push warning in `nix` job                                                           | By design (`continue-on-error`) until cache name + `CACHIX_AUTH_TOKEN` are verified (open question g2).                               |

Not CI-only but seen before: `nix flake check` on the runner needs the same
public `github:` inputs as locally (no SSH since the 2026-09-13 migration) —
a fetch failure there means a flake input rev went stale, which `pin-sync`
will name precisely.

## 3) Badge-job verification checklist (after first green run)

1. `.github/badges/coverage.json` committed by `coverage-bot` with
   `[skip ci]` (the job owns its own concurrency group,
   `cancel-in-progress: false` — queued, never aborted mid-push).
2. The commit rebased on tip before push (`git pull --rebase --autostash`
   step) — confirm no force-push in the job log.
3. README badge renders the CI-measured percentage. Until then the badge
   stays frozen at the last CI-measured value (CI disabled since 2026-07-05;
   locally measured 92.8% vs the 92.0 floor).

## 4) art-dupl warn → blocking decision note

The `art-dupl` job is `continue-on-error: true` per the 2026-10-07
enforced-gate decision (warn-first). Promotion criteria (both, not either):

- CI has a stabilized green baseline (≥3 consecutive full-suite greens with
  art-dupl passing inside them).
- The clone-group count is stable at the accepted baseline. Count at the
  v0.15.0 tag: **25 groups, 0 actionable, exit 0** (23 at the error-epic
  close + 2 accepted groups introduced by the error-system epic's catalog
  and matrix code). All carry in-source `// art-dupl:accept` directives.

Flip = delete the `continue-on-error: true` line in the `art-dupl` job.
Revisit only after the first-run watch period closes.
