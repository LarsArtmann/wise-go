{
  description = "wise-go — unofficial Go SDK for the Wise (TransferWise) API";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "git+ssh://git@github.com/LarsArtmann/go-nix-helpers?ref=master";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    git-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    go-branded-id = {
      url = "git+ssh://git@github.com/LarsArtmann/go-branded-id?rev=e61b48b0f00e217e3475d8f1caf272455401f6eb";
      flake = false;
    };

    go-error-family = {
      url = "git+ssh://git@github.com/LarsArtmann/go-error-family?rev=8ec5aeb6d3f6f45a8315436d934f1a761a07f4f8";
      flake = false;
    };
  };

  outputs =
    inputs:
    let
      fs = inputs.nixpkgs.lib.fileset;
      sourceFiles = fs.unions [
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
        ./transfers.go
        ./types.go
        ./rates.go
        ./quotes.go
        ./recipients.go
        ./delivery_estimates.go
        ./transfer_requirements.go
        ./users.go
        ./webhooks.go
        ./account_details.go
        ./currencies.go
        ./internal/raw/types.go
        ./internal/raw/transfers.go
        ./internal_test.go
        ./example_test.go
        ./wise_test.go
        ./sandbox_live_test.go
        ./readme_guard_test.go
        ./README.md
      ];
      src = fs.toSource {
        root = ./.;
        fileset = sourceFiles;
      };
    in
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.go-nix-helpers.flakeModules.go-standard
        inputs.git-hooks.flakeModule
      ];

      go-standard = {
        pname = "wise-go";
        description = "Unofficial Go SDK for the Wise (TransferWise) API";
        inherit src;
        vendorHash = import ./vendorHash.nix;

        deps = {
          "github.com/larsartmann/go-branded-id" = inputs.go-branded-id;
          "github.com/larsartmann/go-error-family" = inputs.go-error-family;
        };

        # Library: run tests in the dedicated checks.test derivation instead of
        # the package build so we can preserve the race + coverage profile.
        enableCheck = false;
        lintAsCheck = false;

        # go-branded-id v0.5.1 + go-error-family v0.10.0 import encoding/json/v2.
        extraBuildAttrs.env.GOEXPERIMENT = "jsonv2";
        shellExtraEnv.GOEXPERIMENT = "jsonv2";

        # Keep the extra tools from the original devShell (lychee for link
        # checks, go-tools for staticcheck helpers).
        devShellExtraPackages = pkgs: [
          pkgs.go-tools
          pkgs.lychee
        ];
      };

      perSystem =
        {
          config,
          pkgs,
          lib,
          ...
        }:
        {
          # Hermetic test check with race detection and coverage output,
          # matching the original buildGoModule test check.
          checks.test = config.packages.default.overrideAttrs (_old: {
            pname = "wise-go-test";
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
          });

          # Offline link check over the living docs: catches ghost file
          # references (relative links) without network access.
          checks.links =
            pkgs.runCommand "markdown-links"
              {
                nativeBuildInputs = [ pkgs.lychee ];
                # lychee builds its HTTP client eagerly; even offline it
                # needs a CA bundle to initialize.
                SSL_CERT_FILE = "${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt";
                src = lib.fileset.toSource {
                  root = ./.;
                  fileset = lib.fileset.unions [
                    ./README.md
                    ./FEATURES.md
                    ./ROADMAP.md
                    ./TODO_LIST.md
                    ./CHANGELOG.md
                    ./CONTRIBUTING.md
                    ./AGENTS.md
                    ./LICENSE
                    ./.github/workflows
                    ./docs
                  ];
                };
              }
              ''
                cd $src
                lychee --offline --no-progress \
                  README.md FEATURES.md ROADMAP.md TODO_LIST.md CHANGELOG.md CONTRIBUTING.md AGENTS.md
                touch $out
              '';
        };
    };
}
