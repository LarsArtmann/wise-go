# Summary

<!-- What does this PR change and why? One or two sentences. -->

## Changes

-

## Checklist

- [ ] `go test -race ./...` passes (requires `GOEXPERIMENT=jsonv2`; `nix develop` sets it)
- [ ] `golangci-lint run` reports 0 issues
- [ ] `nix fmt .` applied (gofumpt + goimports + treefmt)
- [ ] New public API has godoc (and an `example_test.go` entry where useful)
- [ ] `CHANGELOG.md` updated under `[Unreleased]` for user-visible changes
- [ ] `flake.nix` fileset lists any new `.go` files (sandboxed build reads only that list)
