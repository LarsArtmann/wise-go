# Draft issue for LarsArtmann/md-go-validator

## Title

flake cannot build its own package: go.mod requires 1.27, buildGoModule gets nixpkgs default go 1.26.7

## Body

```markdown
## Symptom

`nix build .#default` on `md-go-validator` `5e30858` (v1.3.0) fails while `go build` in a go ≥ 1.27 toolchain succeeds:

```
go: go.mod requires go >= 1.27 (running go 1.26.7; GOTOOLCHAIN=local)
error: Cannot build '/nix/store/…-md-go-validator-5e30858.drv'
```

## What breaks

- Consumers cannot consume the flake — `github:LarsArtmann/md-go-validator/v1.3.0#packages…` is unbuilt by construction. wise-go had to consume the repo as a `flake = false` SOURCE input and rebuild the tool itself with `(pkgs.buildGoModule.override { go = pkgs.go_1_27; })` (working check at wise-go `checks.md-go-snippets`, 2026-10-07).

## Root cause (source)

`package.nix:21` uses plain `buildGoModule` with no go override:

```nix
buildGoModule {          # package.nix:21 — nixpkgs default go (1.26.7)
  pname = "md-go-validator";
  inherit version vendorHash src;
  proxyVendor = true;
  GOEXPERIMENT = "jsonv2";
```

while `go.mod:3` is `go 1.27`. The derivation pins `GOTOOLCHAIN=local`, so the go command refuses rather than toolchain-switching. `GOEXPERIMENT` is already set — only the toolchain pin is missing.

## Fix

```nix
(pkgs.buildGoModule.override { go = pkgs.go_1_27; }) {
  …
}
```

Note: `vendorHash` may differ between the 1.26 and 1.27 module graphs (it did in the wise-go consumer); recompute after the override. Also drop the dependency on whatever go ≤ 1.26 assumption `flake.nix`'s `runtimeInputs = [ pkgs.go_1_26 ]` (flake.nix:62) makes if it feeds the same build.

## To verify

`nix build .#default` succeeds on a clean store; `nix run .#default -- --help` prints usage; `nix flake check` green.
```

---

💘 Generated with Crush
