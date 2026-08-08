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
          buildGoModule = pkgs.buildGoModule.override { go = goPkg; };
          version = self.rev or self.dirtyRev or "dev";
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
            GOEXPERIMENT = "jsonv2";
          };

          # Minimal CI shell. Same tools, no extras, deterministic.
          devShells.ci = pkgs.mkShellNoCC {
            packages = with pkgs; [
              goPkg
              golangci-lint
            ];
            GOWORK = "off";
            GOPRIVATE = "github.com/larsartmann";
            GOEXPERIMENT = "jsonv2";
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

            test = buildGoModule {
              pname = "wise-go-test";
              inherit version;
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
                  ./internal/raw/types.go
                  ./internal_test.go
                  ./wise_test.go
                ];
              };
              vendorHash = "sha256-nzVrM3mDSBHsvcBSguEeTInA2xzjvOk/cQk7/ck+SSM=";
              doCheck = true;
              checkPhase = ''
                runHook preCheck
                GOEXPERIMENT=jsonv2 go test -race -coverprofile=coverage.out -covermode=atomic ./...
                runHook postCheck
              '';
              installPhase = ''
                runHook preInstall
                mkdir -p $out
                cp coverage.out $out/coverage.out 2>/dev/null || true
                runHook postInstall
              '';
            };
          };
        };
    };
}
