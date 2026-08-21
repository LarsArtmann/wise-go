# Status Report: nix-private-go-repos Migration

**Date:** 2026-08-21 23:10 CEST  
**Session:** Application of `nix-private-go-repos` skill to `wise-go`  
**Reporter:** Crush

---

## Executive Summary

Migrated `wise-go` from a raw `buildGoModule` + `vendorHash.nix` setup to the recommended `flakeModules.go-standard` module from `github.com/LarsArtmann/go-nix-helpers`. The project now fetches its two LarsArtmann dependencies (`go-branded-id`, `go-error-family`) as pinned `git+ssh://` flake inputs and injects them into the sandbox build via `mkPreparedSource`, eliminating reliance on the public Go proxy for private-org dependencies. All targeted verification commands pass.

---

## a) FULLY DONE

1. Read and internalized the `nix-private-go-repos` skill and its two options.
2. Diagnosed the current project state:
   - No `vendor/` directory tracked.
   - Two LarsArtmann deps in `go.mod`: `go-branded-id v0.5.1` and `go-error-family v0.10.0`.
   - Existing flake already set `GOPRIVATE = "github.com/larsartmann"` in dev shells but did not use `mkPreparedSource`.
   - Confirmed the current build passes via the public Go proxy (deps are currently public).
3. Migrated `flake.nix` to use `go-standard`:
   - Added `go-nix-helpers` as a flake input.
   - Added `go-branded-id` and `go-error-family` as `git+ssh://` flake inputs pinned to the exact commits matching the versions in `go.mod`.
   - Imported `inputs.go-nix-helpers.flakeModules.go-standard`.
   - Configured `go-standard.pname`, `description`, `src` fileset, `vendorHash`, `deps`, `enableCheck = false`, `lintAsCheck = false`.
   - Set `extraBuildAttrs.env.GOEXPERIMENT = "jsonv2"` and `shellExtraEnv.GOEXPERIMENT = "jsonv2"`.
   - Preserved `devShellExtraPackages` for `go-tools` and `lychee`.
4. Preserved original project-specific checks:
   - `checks.test` with `go test -race -coverprofile=coverage.out -covermode=atomic`.
   - `checks.links` offline lychee check over living docs.
5. Resolved the new `vendorHash` after the dependency-source change:
   - New hash: `sha256-pD7YJby0gIv3JmCTCA8lw7nWLTGhx1r0UJfujR6REFE=`.
6. Verified the migration end-to-end:
   - `nix build .#default` ✅
   - `nix flake check` ✅ (6 checks pass)
   - `nix develop -c bash -c 'go mod tidy && go build ./...'` ✅
   - `nix develop -c bash -c 'GOEXPERIMENT=jsonv2 go test -race ./...'` ✅
7. Confirmed the devShell now has:
   - `GOPRIVATE=github.com/larsartmann/*,github.com/LarsArtmann/*`
   - `GOWORK=off`
   - `GOTOOLCHAIN=local`
   - `GOEXPERIMENT=jsonv2`
8. Updated `flake.lock` with the new inputs.

---

## b) PARTIALLY DONE

1. **CI authentication for private flake inputs.** The local migration is complete, but CI runners still need SSH access to fetch `git+ssh://` inputs. The skill outlines two strategies (deploy keys or `GITHUB_TOKEN` + `insteadOf`), but neither has been implemented or tested in a workflow.
2. **Understanding of unrelated working-tree changes.** Several Go source files, `AGENTS.md`, and `CHANGELOG.md` were modified during the session (likely by the auto-git daemon or build tooling). I left them untouched per the safety rule not to revert changes one did not author, but their intent/origin is not fully clear.
3. **Documentation updates for the new flake.** `AGENTS.md` and `README.md` still describe the old raw `buildGoModule` structure; no migration note has been added yet.

---

## c) NOT STARTED

1. Pinning `go-nix-helpers` itself to a tag/commit (currently tracks `master`).
2. Setting up CI SSH deploy keys or `GITHUB_TOKEN` + `insteadOf` rewriting.
3. Cross-system verification (`aarch64-linux`, `x86_64-darwin`, `aarch64-darwin`).
4. Adding a govulncheck flake check (it is only in the devShell via `go-standard` defaults).
5. Exposing `lintAsCheck` if a hermetic `checks.lint` is desired.
6. Updating project docs (`AGENTS.md`, `README.md`, `CHANGELOG.md`) to reflect the new flake architecture.
7. Resolving or reverting the unrelated auto-generated source/docs changes.
8. Fresh-clone build test from a machine without the Nix cache.

---

## d) TOTALLY FUCKED UP

Nothing is catastrophically broken. The one real risk is that **CI will break immediately** if a workflow runs `nix flake check` without first providing SSH access to the new `git+ssh://` inputs. This is expected for the private-dep pattern but is an unaddressed operational gap.

---

## e) WHAT WE SHOULD IMPROVE

1. **Reproducibility:** Pin `go-nix-helpers` to a specific commit or tag instead of floating on `master`.
2. **CI hardening:** Choose and document the CI authentication strategy (deploy keys vs. `GITHUB_TOKEN`/`insteadOf`) before the next push to CI.
3. **Source-file hygiene:** Investigate the unrelated modifications to `AGENTS.md`, `CHANGELOG.md`, `account_details.go`, `balances.go`, etc., and either commit them intentionally or revert them cleanly.
4. **Docs:** Add a short section to `AGENTS.md` explaining the `go-standard` module, the `deps` map, and how to bump `vendorHash.nix`.
5. **Apps metadata:** The `apps.x86_64-linux.*` outputs emit warnings about missing `meta.description`; adding descriptions would clean up `nix flake check` output.
6. **Library-specific UX:** `apps.default` points to a non-existent binary for a library; consider documenting this or, if `go-standard` adds an option later, disabling it.
7. **Vendor-hash workflow:** Make the update command discoverable (`nix build .#default` with `lib.fakeHash`) so future dependency bumps are straightforward.
8. **Cross-system testing:** Run `nix flake check --all-systems` or on the available non-Linux builders to catch Darwin/Linux ARM issues early.

---

## f) Up to 50 Things We Should Get Done Next

1. Pin `go-nix-helpers` flake input to a tag or commit.
2. Add CI SSH deploy-key setup for `git+ssh://` inputs.
3. Alternatively, configure CI `GITHUB_TOKEN` + `insteadOf` SSH→HTTPS rewriting.
4. Update `.github/workflows` to run `nix flake check` with proper auth.
5. Test `nix flake check` on `aarch64-linux`.
6. Test `nix flake check` on `x86_64-darwin`.
7. Test `nix flake check` on `aarch64-darwin`.
8. Add a CI cache (Cachix or similar) to speed up repeated private-dep builds.
9. Update `AGENTS.md` with the new flake architecture and private-dep notes.
10. Update `README.md` build/dev instructions if they mention the old flake structure.
11. Add a `CHANGELOG.md` entry for the flake migration.
12. Resolve whether to keep or revert the unrelated daemon-generated source changes.
13. Investigate why the daemon modified `account_details.go`, `balances.go`, etc.
14. Run `nix fmt` once more after resolving source-file changes.
15. Verify `checks.test` still emits `coverage.out` to `$out`.
16. Add `checks.vulns` for hermetic `govulncheck`.
17. Evaluate `lintAsCheck = true` and fix any new lint findings.
18. Add `meta.description` to `apps.test`, `apps.default`, `apps.lint`, `apps.fmt`.
19. Document the exact `vendorHash.nix` update workflow.
20. Add a flake comment header explaining the library-specific choices.
21. Verify `nix run .#test` works (it should; `nix run .#default` will not for a library).
22. Confirm `go-standard` default src filtering is not sufficient before keeping the explicit fileset.
23. Review whether `enableOverlay = true` is desirable for a library.
24. Check if `packages.default` for a library is useful or confusing for consumers.
25. Add a `nix run .#update-vendor-hash` helper app if desired.
26. Test a fresh `git clone` followed by `nix flake check`.
27. Test the build with `--option substitute false` to verify true sandbox independence.
28. Verify `git config url."git@github.com:".insteadOf` is documented for new devs.
29. Add a `docs/reviews/2026-08-21_nix-private-go-repos-migration.md` retrospective.
30. Update `TODO_LIST.md` with remaining CI/docs tasks.
31. Verify `nix develop -c golangci-lint run ./...` still passes.
32. Check `buildflow` compatibility with the new flake outputs.
33. Confirm the pre-commit hook still runs `nix fmt` correctly.
34. Audit `flake.lock` for unnecessary duplicate `flake-parts` entries.
35. Consider adding `publicDeps` if the LarsArtmann deps are intended to stay public.
36. Document why the deps are treated as private even though they are currently public.
37. Verify `go mod tidy` does not inject local `replace` directives into the committed `go.mod`.
38. Ensure `flake.nix` is git-tracked (it is) for the buildflow pre-commit hook.
39. Review `go-standard`'s `enableGolangciLint` default and `apps.lint` behavior.
40. Verify devShell has `gopls` available for LSP users.
41. Run the full test suite with `-coverprofile` locally.
42. Compare the new `checks.test` output with the old one to confirm parity.
43. Make sure `checks.links` still catches broken relative links.
44. Add a note about `GOEXPERIMENT=jsonv2` being required for this codebase.
45. Review if `go-tools` and `lychee` are still the right devShell extras.
46. Verify `nix fmt` formats `flake.nix` deterministically across runs.
47. Check that no secrets or SSH keys are referenced inside `flake.nix`.
48. Consider a CI matrix job for the new `checks.test` output artifact.
49. Update `FEATURES.md` if build/dev tooling is listed there.
50. Celebrate that the migration is done and move on to API expansion.

---

## g) Up to 3 Questions I Cannot Figure Out Myself

1. **Should `go-nix-helpers` be pinned to a specific commit/tag for reproducibility, or is tracking `master` intentional?** The skill examples use `master`, but a library consumed by others may benefit from a locked helper revision.
2. **The working tree contains modifications to `AGENTS.md`, `CHANGELOG.md`, and several Go source files that I did not make directly. Are those changes intentional and desired, or should they be reverted before this migration is finalized?**
3. **Are `go-branded-id` and `go-error-family` intended to remain public repositories?** They currently resolve via the public Go proxy, which means the migration is defensive rather than strictly required; confirming the intent affects whether we should keep them in `deps` or move them to `publicDeps`.

---

## Raw Verification Log (This Session)

```
$ nix build .#default
# => success

$ nix flake check
# => all checks passed
# => 6 flake checks built: treefmt, pre-commit, markdown-links, wise-go-prepared-source, wise-go, wise-go-test

$ nix develop -c bash -c 'go mod tidy && go build ./...'
# => success

$ nix develop -c bash -c 'GOEXPERIMENT=jsonv2 go test -race ./...'
# => ok  	github.com/larsartmann/wise-go
# => ?  	github.com/larsartmann/wise-go/internal/raw	[no test files]

$ nix develop -c bash -c 'echo GOPRIVATE=$GOPRIVATE GOEXPERIMENT=$GOEXPERIMENT GOWORK=$GOWORK GOTOOLCHAIN=$GOTOOLCHAIN'
# => GOPRIVATE=github.com/larsartmann/*,github.com/LarsArtmann/* GOEXPERIMENT=jsonv2 GOWORK=off GOTOOLCHAIN=local
```

## Files Touched in This Session (Key Ones)

- `flake.nix` — migrated to `go-standard` module.
- `vendorHash.nix` — updated to `sha256-pD7YJby0gIv3JmCTCA8lw7nWLTGhx1r0UJfujR6REFE=`.
- `flake.lock` — added `go-nix-helpers`, `go-branded-id`, `go-error-family`; removed `systems` and `treefmt-nix`.
- `docs/status/2026-08-21_23-10_nix-private-go-repos-migration.md` — this report.

## Unrelated/Unrequested Changes Observed (Not Reverted)

- `AGENTS.md` modified.
- `CHANGELOG.md` modified.
- Multiple Go source files modified (e.g., `account_details.go`, `balances.go`, `delivery_estimates.go`, `helpers.go`, `ids.go`, `quotes.go`, `recipients.go`, `transactions.go`, `transfer_requirements.go`, `transfers.go`).
- These appear to introduce/refactor usage of a `requireID` helper. Origin is likely the auto-git daemon or another build tool; left untouched pending clarification.
