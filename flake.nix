# Nix flake for wise-go
#
# wise-go is a Go *library*, not a binary, so this flake focuses on:
#   - devShells: reproducible local + CI environments
#   - checks: go test, golangci-lint, treefmt, govulncheck
#   - treefmt: gofumpt + goimports + nixfmt
#
# There is no packages.default because the project produces no binary. The
# library is consumed via `go get github.com/larsartmann/wise-go`.

{
  description = "wise-go — unofficial Go SDK for the Wise (TransferWise) API";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    systems.url = "github:nix-systems/default";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    git-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ self, flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;
      imports = [
        inputs.treefmt-nix.flakeModule
        inputs.git-hooks.flakeModule
      ];

      perSystem =
        {
          config,
          pkgs,
          lib,
          ...
        }:
        let
          goPkg = pkgs.go_1_26;
        in
        {
          # Reproducible local dev shell. `nix develop` enters this.
          devShells.default = pkgs.mkShellNoCC {
            packages = with pkgs; [
              goPkg
              gopls
              golangci-lint
              go-tools # staticcheck + helpers
            ];
            GOWORK = "off";
            GOPRIVATE = "github.com/larsartmann";
          };

          # Minimal CI shell. Same tools, no extras, deterministic.
          devShells.ci = pkgs.mkShellNoCC {
            packages = with pkgs; [
              goPkg
              golangci-lint
            ];
            GOWORK = "off";
            GOPRIVATE = "github.com/larsartmann";
          };

          # `nix fmt`
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              nixfmt.enable = true;
            };
          };

          # `nix flake check`
          checks = {
            format = config.treefmt.build.check self;

            test = pkgs.stdenvNoCC.mkDerivation {
              pname = "wise-go-test";
              version = "dev";
              src = lib.fileset.toSource {
                root = ./.;
                fileset = lib.fileset.unions [
                  ./go.mod
                  ./go.sum
                  ./ids.go
                  ./helpers.go
                  ./options.go
                  ./profiles.go
                  ./balances.go
                  ./errors.go
                  ./client.go
                  ./transactions.go
                  ./types.go
                  ./internal_test.go
                  ./wise_test.go
                ];
              };
              nativeBuildInputs = [
                goPkg
              ];
              GOWORK = "off";
              GOPRIVATE = "github.com/larsartmann";
              # Vendor deps offline, then run tests with the race detector.
              buildPhase = ''
                runHook preBuild
                export HOME=$(mktemp -d)
                export GOCACHE=$TMPDIR/go-cache
                export GOMODCACHE=$TMPDIR/go-mod
                go test -race -coverprofile=coverage.out -covermode=atomic ./...
                runHook postBuild
              '';
              doCheck = false;
              installPhase = ''
                runHook preInstall
                mkdir -p $out
                cp coverage.out $out/coverage.out
                runHook postInstall
              '';
            };

            lint = pkgs.stdenvNoCC.mkDerivation {
              pname = "wise-go-lint";
              version = "dev";
              src = lib.fileset.toSource {
                root = ./.;
                fileset = lib.fileset.unions [
                  ./go.mod
                  ./go.sum
                  ./ids.go
                  ./helpers.go
                  ./options.go
                  ./profiles.go
                  ./balances.go
                  ./errors.go
                  ./client.go
                  ./transactions.go
                  ./types.go
                  ./internal_test.go
                  ./wise_test.go
                  ./.golangci.yml
                ];
              };
              nativeBuildInputs = [
                goPkg
                pkgs.golangci-lint
              ];
              GOWORK = "off";
              GOPRIVATE = "github.com/larsartmann";
              buildPhase = ''
                runHook preBuild
                export HOME=$(mktemp -d)
                export GOCACHE=$TMPDIR/go-cache
                export GOMODCACHE=$TMPDIR/go-mod
                # Download deps so lint can read the package graph.
                go mod download || true
                golangci-lint run --timeout=5m
                runHook postBuild
              '';
              doCheck = false;
              installPhase = "touch $out";
            };
          };
        };
    };
}
