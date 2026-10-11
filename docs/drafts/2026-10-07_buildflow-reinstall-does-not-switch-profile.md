# Draft issue for LarsArtmann/buildflow

## Title

nix run .#reinstall prints REINSTALL-OK but does not switch ~/.nix-profile

## Body

```markdown
## Symptom

`nix run .#reinstall` (buildflow repo, 2026-10-07) completes and prints `REINSTALL-OK`, but `~/.nix-profile` still points at the previous generation. The running `buildflow` binary keeps its old version, so analyzer verdicts made "after a reinstall" are made on a stale binary.

## What breaks

- The reinstall app's own success signal is a lie: exit 0 + `REINSTALL-OK` with no profile switch.
- Downstream decisions inherit the stale binary. Concrete instance: an erraudit verdict was accepted from a binary 160 commits behind HEAD before the staleness was noticed (wise-go session, 2026-10-07).

## Root cause (source)

The app builds/installs the store path but does not perform the profile mutation a `nix profile` install would. Verified workaround on the same day:
```

nix profile remove buildflow
nix profile install <store-path>

```
After the manual remove/add, `~/.nix-profile/bin/buildflow` matches the freshly built store path.

## Operator steps (repo)

- [ ] Make the reinstall app actually switch the profile (prefer `nix profile install` semantics inside the app), or
- [ ] print the exact manual commands + the store path and exit non-zero-equivalent guidance instead of `REINSTALL-OK`.
- [ ] Add a post-run assertion: resolved `~/.nix-profile/bin/buildflow` store path == built store path, else fail loudly.

## To verify

On a machine with an older profile generation: `nix run .#reinstall` → `readlink -f ~/.nix-profile/bin/buildflow` equals the newly built store path.
```

---

💘 Generated with Crush
