{
  description = "wise-go — unofficial Go SDK for the Wise (TransferWise) API";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "github:LarsArtmann/go-nix-helpers?rev=16c3184262c55377aba2126dc20e028637f58aa0";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    git-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    go-branded-id = {
      url = "github:LarsArtmann/go-branded-id?rev=0fe9f367e6176d446634312c462477d2276ba025";
      flake = false;
    };

    go-error-family = {
      url = "github:LarsArtmann/go-error-family?rev=2d0679e03652e6244c5b6bbbcaa935fbf57e78ef";
      flake = false;
    };

    go-retry = {
      url = "github:LarsArtmann/go-retry?rev=082842a0bac8aa37df8e0ca894e677b0964c5faf";
      flake = false;
    };

    # Source-only: built in checks.md-go-snippets with this repo's nixpkgs
    # (go 1.27 for the tool's go.mod floor — its own flake currently pins a
    # go too old to build it, so its packages.default is not consumable).
    md-go-validator = {
      url = "github:LarsArtmann/md-go-validator/v1.3.0";
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
        ./third_party_transfers.go
        ./types.go
        ./rates.go
        ./quotes.go
        ./recipients.go
        ./delivery_estimates.go
        ./transfer_requirements.go
        ./users.go
        ./webhooks.go
        ./ott.go
        ./account_details.go
        ./currencies.go
        ./internal/raw/types.go
        ./internal/raw/transfers.go
        ./internal/raw/webhooks.go
        ./internal/raw/ott.go
        ./internal/raw/types_test.go
        ./internal_test.go
        ./errors_test.go
        ./helpers_test.go
        ./bench_test.go
        ./example_test.go
        ./wise_test.go
        ./ott_test.go
        ./sandbox_live_test.go
        ./readme_guard_test.go
        ./spec_conformance_test.go
        ./zz_spec_conformance_coverage_test.go
        ./docs/reviews/wise-api-openapi.json
        ./README.md
        ./testdata
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
          "github.com/larsartmann/go-retry" = inputs.go-retry;
        };

        # Library: run tests in the dedicated checks.test derivation instead of
        # the package build so we can preserve the race + coverage profile.
        enableCheck = false;
        lintAsCheck = false;

        # go-branded-id v0.5.1 + go-error-family v0.10.0 import encoding/json/v2.
        extraBuildAttrs.env.GOEXPERIMENT = "jsonv2";
        shellExtraEnv.GOEXPERIMENT = "jsonv2";

        # Keep the extra tools from the original devShell (lychee for link
        # checks, go-tools for staticcheck helpers). gitleaks/codespell/
        # markdownlint-cli are the three on-demand buildflow tools that
        # otherwise report "unavailable (health check failed)" every run.
        # dprint/ruff/vulnix are pipeline tools buildflow otherwise nix-runs
        # ad hoc with a "not in project devShell" warning on every run
        # (observed 2026-10-17).
        devShellExtraPackages = pkgs: [
          pkgs.go-tools
          pkgs.lychee
          pkgs.gitleaks
          pkgs.codespell
          pkgs.markdownlint-cli
          pkgs.dprint
          pkgs.ruff
          pkgs.vulnix
          # Python toolchain for scripts/*.py (the new binary runs the full
          # Python provider set on this repo; without these it nix-runs each
          # ad hoc with a "not in project devShell" warning). NOTE: pytest
          # and vulture have no top-level attr in nixpkgs-unstable (removed
          # ~2026-10); use python3Packages.* — top-level pkgs.pytest breaks
          # the entire flake eval (live 2026-10-17: every nix-shell step
          # died with "attribute 'pytest' missing").
          pkgs.pyupgrade
          pkgs.bandit
          pkgs.mypy
          pkgs.python3Packages.pytest
          pkgs.python3Packages.vulture
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
          # Hermetic test check with race detection and coverage output.
          # The coverage floor is SINGLE-SOURCED in ci.yml (the parity check
          # below enforces the shape): this check reads the floor from ci.yml
          # at check time, so a promotion can never split the two gates.
          checks.test = config.packages.default.overrideAttrs (_old: {
            pname = "wise-go-test";
            doCheck = true;
            ciYml = ./.github/workflows/ci.yml;
            checkPhase = ''
              runHook preCheck
              GOEXPERIMENT=jsonv2 go test -race -coverprofile=coverage.out -covermode=atomic ./...
              TOTAL=$(go tool cover -func=coverage.out | awk '/^total:/ {sub("%", "", $3); print $3}')
              echo "coverage: $TOTAL%"
              FLOOR=$(grep -oE 'p\+0 < [0-9]+(\.[0-9]+)?' "$ciYml" | grep -oE '[0-9]+(\.[0-9]+)?$' | sort -u)
              if [ -z "$FLOOR" ] || [ "$(printf '%s\n' "$FLOOR" | grep -c .)" -ne 1 ]; then
                echo "FAIL: expected exactly one coverage floor in ci.yml, found: '$FLOOR'"
                exit 1
              fi
              awk -v p="$TOTAL" -v f="$FLOOR" 'BEGIN { if ((p+0) < (f+0)) exit 1 }' || {
                echo "FAIL: coverage $TOTAL% is below the floor $FLOOR (single-sourced from ci.yml)"
                exit 1
              }
              runHook postCheck
            '';
            installPhase = ''
              runHook preInstall
              mkdir -p $out
              cp coverage.out $out/coverage.out 2>/dev/null || true
              runHook postInstall
            '';
          });

          # The coverage floor is single-sourced: ci.yml carries THE floor,
          # checks.test reads it from there at check time. This check pins the
          # shape — exactly one floor in ci.yml, none hardcoded in flake.nix —
          # so a two-site promotion or a floor removal breaks the build
          # instead of silently splitting (or dropping) the gate.
          checks.coverage-floor-parity =
            pkgs.runCommand "coverage-floor-parity"
              {
                nativeBuildInputs = [
                  pkgs.gnugrep
                  pkgs.coreutils
                ];
                flakeNix = ./flake.nix;
                ciYml = ./.github/workflows/ci.yml;
              }
              ''
                FLOOR_CI=$(grep -oE 'p\+0 < [0-9]+(\.[0-9]+)?' "$ciYml" | grep -oE '[0-9]+(\.[0-9]+)?$' | sort -u)
                COUNT=$(printf '%s\n' "$FLOOR_CI" | grep -c . || true)
                INLINE_FLAKE=$(grep -cE '\(p\+0\) < [0-9]' "$flakeNix" || true)
                if [ -z "$FLOOR_CI" ] || [ "$COUNT" -ne 1 ]; then
                  echo "FAIL: ci.yml must carry exactly one coverage floor (found: '$FLOOR_CI')"
                  exit 1
                fi
                if [ "$INLINE_FLAKE" -ne 0 ]; then
                  echo "FAIL: flake.nix hardcodes a coverage floor — the floor is single-sourced in ci.yml"
                  echo "(checks.test reads it via the ciYml derivation attr)"
                  exit 1
                fi
                echo "coverage-floor-parity OK: single floor ($FLOOR_CI) sourced from ci.yml"
                touch $out
              '';

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
                    ./.github/SECURITY.md
                    ./docs
                  ];
                };
              }
              ''
                cd $src
                lychee --offline --no-progress \
                  README.md FEATURES.md ROADMAP.md TODO_LIST.md CHANGELOG.md CONTRIBUTING.md AGENTS.md docs/error-codes.md
                touch $out
              '';

          # Fenced-Go snippet parse gate over the living docs and docs/:
          # README's webhook block #27 once shipped an invalid shape that
          # only a hand run of md-go-validator caught (fixed 2026-09-29);
          # this fails that class at check time. The tool builds with go
          # 1.27 — independent of the SDK's pinned 1.26 toolchain.
          checks.md-go-snippets =
            let
              md-go-validator = (pkgs.buildGoModule.override { go = pkgs.go_1_27; }) {
                pname = "md-go-validator";
                version = "1.3.0";
                src = inputs.md-go-validator;
                vendorHash = import ./vendorHashMdGoValidator.nix;
                proxyVendor = true;
                env.GOEXPERIMENT = "jsonv2";
              };
            in
            pkgs.runCommand "md-go-snippets"
              {
                nativeBuildInputs = [ md-go-validator ];
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
                    ./docs
                  ];
                };
              }
              ''
                cd $src
                md-go-validator README.md FEATURES.md ROADMAP.md TODO_LIST.md CHANGELOG.md CONTRIBUTING.md AGENTS.md docs/error-codes.md $(find docs -type d)
                touch $out
              '';

          # Line-anchor gate: every `Symbol` (`file.go:NN`) reference in the
          # living docs must still land — the symbol on the cited line, or a
          # non-blank line for bare anchors. doc-verify gates counts, not line
          # anchors, and this class rots silently (16 dead anchors found and
          # fixed 2026-10-16). Hermetic; scripts/anchor_check.py --self-test
          # covers the drift classes. On-demand: nix run .#anchor-check.
          checks.anchor-check =
            pkgs.runCommand "anchor-check"
              {
                nativeBuildInputs = [ pkgs.python3 ];
                src = lib.fileset.toSource {
                  root = ./.;
                  fileset = lib.fileset.unions [
                    ./README.md
                    ./FEATURES.md
                    ./ROADMAP.md
                    ./TODO_LIST.md
                    ./CONTRIBUTING.md
                    ./AGENTS.md
                    ./docs
                    ./scripts/anchor_check.py
                    ./go.mod
                    (lib.fileset.fileFilter (file: file.hasExt "go") ./.)
                  ];
                };
              }
              ''
                cd $src
                python3 scripts/anchor_check.py --repo .
                touch $out
              '';

          # Negative fixtures for the python gates (scripts/negative_fixtures.py):
          # pin-sync's stale-lock/tag-mismatch/missing-input/unpinned shapes and
          # spec-drift's broken-spec/missing-endpoint/--fail-on-drift postures,
          # each asserted on exit code AND output substring. Hermetic — every
          # case passes --offline or --spec-file, nothing is fetched.
          checks.script-fixtures =
            pkgs.runCommand "script-fixtures"
              {
                nativeBuildInputs = [ pkgs.python3 ];
                src = lib.fileset.toSource {
                  root = ./.;
                  fileset = lib.fileset.unions [
                    ./scripts/negative_fixtures.py
                    ./scripts/pin_sync.py
                    ./scripts/spec_drift.py
                    ./scripts/testdata
                    ./go.mod
                    ./docs/reviews/wise-api-openapi.json
                    (lib.fileset.fileFilter (file: file.hasExt "go") ./.)
                  ];
                };
              }
              ''
                cd $src
                export PYTHONDONTWRITEBYTECODE=1
                python3 scripts/negative_fixtures.py --repo-root .
                touch $out
              '';

          # Curated-lister gate: the enabled set in .golangci.yml must equal
          # the allowlist below. buildflow's auto-configure re-added `lll`
          # twice on 2026-10-16 alone; this turns the curated-list policy
          # from memory into a gate. To change the list: edit .golangci.yml
          # AND this allowlist in the same commit.
          checks.linter-allowlist =
            let
              allowlist = [
                "arangolint"
                "asasalint"
                "asciicheck"
                "bidichk"
                "bodyclose"
                "canonicalheader"
                "clickhouselint"
                "containedctx"
                "contextcheck"
                "copyloopvar"
                "cyclop"
                "decorder"
                "dogsled"
                "dupl"
                "dupword"
                "durationcheck"
                "embeddedstructfieldcheck"
                "err113"
                "errcheck"
                "errchkjson"
                "errname"
                "errorlint"
                "exhaustive"
                "exhaustruct_v5"
                "exptostd"
                "fatcontext"
                "forbidigo"
                "forcetypeassert"
                "funlen"
                "ginkgolinter"
                "gocheckcompilerdirectives"
                "gochecknoglobals"
                "gochecknoinits"
                "gochecksumtype"
                "gocognit"
                "goconst"
                "gocritic"
                "gocyclo"
                "godoclint"
                "godot"
                "godox"
                "goheader"
                "gomoddirectives"
                "gomodguard_v2"
                "goprintffuncname"
                "gosec"
                "gosmopolitan"
                "govet"
                "grouper"
                "iface"
                "importas"
                "inamedparam"
                "ineffassign"
                "interfacebloat"
                "intrange"
                "iotamixing"
                "ireturn"
                "loggercheck"
                "maintidx"
                "makezero"
                "mirror"
                "misspell"
                "mnd"
                "modernize"
                "musttag"
                "nakedret"
                "nestif"
                "nilerr"
                "nilnesserr"
                "nilnil"
                "nlreturn"
                "noctx"
                "nolintlint"
                "nonamedreturns"
                "nosprintfhostport"
                "paralleltest"
                "perfsprint"
                "prealloc"
                "predeclared"
                "promlinter"
                "protogetter"
                "reassign"
                "recvcheck"
                "revive"
                "rowserrcheck"
                "sloglint"
                "spancheck"
                "sqlclosecheck"
                "staticcheck"
                "tagalign"
                "tagliatelle"
                "testableexamples"
                "testifylint"
                "testpackage"
                "thelper"
                "tparallel"
                "unconvert"
                "unparam"
                "unqueryvet"
                "unused"
                "usestdlibvars"
                "usetesting"
                "varnamelen"
                "wastedassign"
                "whitespace"
                "wrapcheck"
                "wsl_v5"
                "zerologlint"
              ];
            in
            pkgs.runCommand "linter-allowlist"
              {
                nativeBuildInputs = [
                  pkgs.gawk
                  pkgs.diffutils
                  pkgs.coreutils
                ];
                golangciYml = ./.golangci.yml;
                allowlistFile = pkgs.writeText "linter-allowlist" (lib.concatStringsSep "\n" allowlist);
              }
              ''
                awk '/^  enable:/{f=1;next} f&&/^    - /{sub(/^    - /,"");print;next} f{exit}' "$golangciYml" | sort > actual.txt
                sort "$allowlistFile" > expected.txt
                if ! diff -u expected.txt actual.txt; then
                  echo "FAIL: .golangci.yml enabled-linter set drifted from the curated allowlist"
                  echo "The curated list is policy; buildflow's auto-configure step is skipped (BuildFlow #63 watch), so drift means a manual or tooling edit."
                  echo "Change it by editing .golangci.yml AND the allowlist in flake.nix checks.linter-allowlist together."
                  exit 1
                fi
                echo "linter-allowlist OK: enabled set matches the curated allowlist"
                touch $out
              '';

          # Dependency-pin sync: every in-house Go dependency required by go.mod
          # must have a flake input pinned to that exact version's commit. The
          # go-retry flake pin once lagged its go.mod tag for ten days while
          # every local gate stayed green — nothing compared the pair, and the
          # Nix test check vendored a different go-retry than `go test` did.
          # Hermetic here (no network); `nix run .#pin-sync` additionally
          # resolves each version tag remotely to catch a stale rev.
          checks.pin-sync =
            pkgs.runCommand "pin-sync"
              {
                nativeBuildInputs = [ pkgs.python3 ];
                src = lib.fileset.toSource {
                  root = ./.;
                  fileset = lib.fileset.unions [
                    ./go.mod
                    ./flake.nix
                    ./flake.lock
                    ./scripts/pin_sync.py
                  ];
                };
              }
              ''
                cd $src
                python3 scripts/pin_sync.py --offline --root .
                touch $out
              '';

          # Breaking-change check against the latest tagged release
          # (gorelease reports removed/changed exported API). Needs network
          # to fetch the base version, so it is an app (nix run), never a
          # sandboxed check.
          apps.apidiff =
            let
              apidiff = pkgs.writeShellApplication {
                name = "apidiff";
                runtimeInputs = [ pkgs.go ];
                text = ''
                  export GOEXPERIMENT=jsonv2
                  # Pinned pseudo-version: @latest would make apidiff results
                  # drift between runs of the same commit (non-reproducible).
                  exec go run golang.org/x/exp/cmd/gorelease@v0.0.0-20261005173118-76772065c9b0 -base=latest "$@"
                '';
              };
            in
            {
              type = "app";
              program = pkgs.lib.getExe apidiff;
              meta.description = "Compare the public API against the latest release tag (gorelease; needs network)";
            };

          # Living-docs health: link check + godoc render + count-claims
          # freshness (the drift class that let "33 methods" outlive the
          # 37-method surface). Run from the repo root: nix run .#doc-verify.
          apps.doc-verify =
            let
              doc-verify = pkgs.writeShellApplication {
                name = "doc-verify";
                runtimeInputs = [
                  pkgs.go
                  pkgs.git
                  pkgs.lychee
                ];
                text = ''
                  set -euo pipefail
                  cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
                  status=0

                  export GOEXPERIMENT=jsonv2

                  # 1. godoc must render cleanly.
                  if ! go doc -all . > /dev/null; then
                    echo "FAIL: go doc -all does not render"
                    status=1
                  fi

                  # 2. Count claims in living docs must match the real
                  # surface. A claim pattern that extracts EMPTY fails
                  # loudly — the silent-skip class (claim text drifts away,
                  # gate quietly passes) is exactly what this check exists
                  # to kill.
                  check_count() {
                    local label="$1" file="$2" pattern="$3" actual="$4" claimed
                    claimed="$(grep -oE "$pattern" "$file" 2>/dev/null | head -1 | grep -oE '^[0-9]+' || true)"
                    if [ -z "$claimed" ]; then
                      echo "FAIL: count-claim pattern extracted NOTHING from $file (pattern: $pattern)"
                      status=1
                    elif [ "$claimed" != "$actual" ]; then
                      echo "FAIL: $file claims $claimed $label, actual is $actual"
                      status=1
                    else
                      echo "ok: $file: $claimed $label"
                    fi
                  }

                  actual_methods="$(go doc -all . | grep -c '^func (c \*Client)' || true)"
                  check_count "endpoint methods" AGENTS.md '[0-9]+ endpoint methods' "$actual_methods"
                  # FEATURES phrases the same number as shipped operations;
                  # they coincide today, and if they ever diverge the gate
                  # forces a conscious doc update instead of silent drift.
                  check_count "shipped operations" FEATURES.md '[0-9]+ of those documented' "$actual_methods"
                  check_count "methods" ROADMAP.md '[0-9]+ methods' "$actual_methods"
                  check_count "Client methods" docs/reviews/2026-08-21_v1.0-api-audit.md '[0-9]+ .\*Client. methods across' "$actual_methods"

                  # Release notes: only the LATEST docs/releases file is a
                  # living claim — it must state the current method count.
                  # Historical notes stay frozen at their tags by design.
                  latest_release_notes="$(find docs/releases -name '*-release-notes.md' 2>/dev/null | sort -V | tail -1 || true)"
                  if [ -n "$latest_release_notes" ]; then
                    check_count "client methods" "$latest_release_notes" '[0-9]+ client methods' "$actual_methods"
                  fi

                  actual_examples="$(grep -c '^func Example' example_test.go || true)"
                  check_count "Example funcs" FEATURES.md '[0-9]+ .Example.. funcs' "$actual_examples"

                  # 3. No ghost relative links in the living docs.
                  if ! lychee --offline --no-progress \
                    README.md FEATURES.md ROADMAP.md TODO_LIST.md CHANGELOG.md CONTRIBUTING.md AGENTS.md docs/error-codes.md; then
                    status=1
                  fi

                  if [ "$status" -eq 0 ]; then
                    echo "doc-verify: all checks passed"
                  fi
                  exit "$status"
                '';
              };
            in
            {
              type = "app";
              program = pkgs.lib.getExe doc-verify;
              meta.description = "Check living-doc links, godoc render, and count-claim freshness";
            };
          # Release-notes pre-flight for `gh release create`: fails on the
          # two defects that shipped in the v0.12.0 release body — a code
          # span split across lines by a reflow, and repo-relative links
          # (GitHub Release bodies cannot resolve repo-relative paths).
          apps.release-notes-check =
            let
              release-notes-check = pkgs.writeShellApplication {
                name = "release-notes-check";
                text = ''
                  set -euo pipefail
                  status=0
                  shopt -s nullglob
                  files=(docs/releases/*-release-notes.md)
                  if [ "''${#files[@]}" -eq 0 ]; then
                    echo "FAIL: no docs/releases/*-release-notes.md found (run from the repo root)"
                    exit 1
                  fi

                  for f in "''${files[@]}"; do
                    while read -r line; do
                      echo "FAIL: $f:$line: code span split across lines (odd backtick count)"
                      status=1
                    done < <(awk '/^```/ { infence = !infence; next } !infence { n = gsub(/`/, "`"); if (n % 2 == 1) print FNR }' "$f")

                    while read -r target; do
                      echo "FAIL: $f: repo-relative link target '$target' (GitHub Release bodies need absolute URLs)"
                      status=1
                    done < <(awk '/^```/ { infence = !infence; next } !infence' "$f" | grep -oE '\]\([^)]+\)' | sed -E 's/^\]\(//; s/\)$//' | grep -vE '^(https?://|mailto:|#)' || true)
                  done

                  if [ "$status" -eq 0 ]; then
                    echo "release-notes-check: all release-notes files clean"
                  fi
                  exit "$status"
                '';
              };
            in
            {
              type = "app";
              program = pkgs.lib.getExe release-notes-check;
              meta.description = "Pre-flight release-notes files for split code spans and repo-relative links";
            };

          # Living-docs line-anchor verifier (scripts/anchor_check.py):
          # asserts every `Symbol` (`file.go:NN`) reference in README/
          # FEATURES/TODO_LIST/ROADMAP/AGENTS/CONTRIBUTING + docs/*.md still
          # lands. Also runs in `nix flake check` (checks.anchor-check) and
          # the pre-release chain.
          apps.anchor-check =
            let
              anchor-check = pkgs.writeShellApplication {
                name = "anchor-check";
                runtimeInputs = [
                  pkgs.python3
                  pkgs.git
                ];
                text = ''
                  python3 ${./scripts/anchor_check.py} --self-test || exit 1
                  exec python3 ${./scripts/anchor_check.py} --repo "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
                '';
              };
            in
            {
              type = "app";
              program = pkgs.lib.getExe anchor-check;
              meta.description = "Verify every file.go:NN line anchor in the living docs still lands";
            };

          # SDK surface vs live Wise API reference (scripts/spec_drift.py):
          # every (method, path) the SDK calls must still exist in the live
          # spec, and vendored-snapshot drift is reported (breaking drift for
          # the SDK = exit 1). Needs network, so it is an app (nix run),
          # never a sandboxed check. Run from the repo root:
          # nix run .#spec-drift
          apps.spec-drift =
            let
              spec-drift = pkgs.writeShellApplication {
                name = "spec-drift";
                runtimeInputs = [
                  pkgs.python3
                  pkgs.git
                ];
                text = ''
                  export PYTHONDONTWRITEBYTECODE=1
                  exec python3 ${./scripts/spec_drift.py} --repo "$(git rev-parse --show-toplevel 2>/dev/null || pwd)" "$@"
                '';
              };
            in
            {
              type = "app";
              program = pkgs.lib.getExe spec-drift;
              meta.description = "Check SDK endpoints against the live Wise API spec and report vendored-snapshot drift (needs network)";
            };

          # Full dependency-pin verification (needs network): resolves each
          # go.mod version tag on GitHub and asserts it points at the pinned
          # rev. Part of `nix run .#pre-release`.
          apps.pin-sync =
            let
              pin-sync = pkgs.writeShellApplication {
                name = "pin-sync";
                runtimeInputs = [
                  pkgs.python3
                  pkgs.git
                ];
                text = ''
                  exec python3 ${./scripts/pin_sync.py} --root "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
                '';
              };
            in
            {
              type = "app";
              program = pkgs.lib.getExe pin-sync;
              meta.description = "Verify in-house flake pins match go.mod and the remote tags";
            };

          # Apps provided by the imported flake modules; annotate them so
          # `nix flake check` stops warning about a missing meta.description.
          apps.default.meta.description = "Build the wise-go library";
          apps.test.meta.description = "Run the hermetic race + coverage test suite";
          apps.lint.meta.description = "Run golangci-lint over the module";
          apps.fmt.meta.description = "Format the repository (treefmt)";

          # One-command release gate: chains every local gate a release
          # previously relied on session discipline to run (the v0.12.0
          # cycle's broken release-body span shipped through exactly such a
          # manual sequence). Fails fast; the dirty-tree check runs first so
          # a release never cuts from uncommitted work.
          apps.pre-release =
            let
              pre-release = pkgs.writeShellApplication {
                name = "pre-release";
                runtimeInputs = with pkgs; [
                  git
                  go
                  golangci-lint
                ];
                text = ''
                  set -euo pipefail
                  cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
                  export GOEXPERIMENT=jsonv2

                  echo "==> dirty-tree check"
                  if [ -n "$(git status --porcelain)" ]; then
                    echo "FAIL: working tree is dirty; commit or stash before releasing"
                    git status --short
                    exit 1
                  fi

                  echo "==> go build"
                  go build ./...

                  echo "==> go vet"
                  go vet ./...

                  echo "==> race tests"
                  go test -race ./...

                  if [ "''${1:-}" = "--shuffle" ]; then
                    echo "==> shuffle battery (go test -shuffle=on, 3 passes)"
                    go test -shuffle=on ./...
                    go test -shuffle=on ./...
                    go test -shuffle=on ./...
                  fi

                  echo "==> golangci-lint"
                  golangci-lint run

                  echo "==> nix flake check"
                  nix flake check

                  echo "==> doc-verify"
                  nix run .#doc-verify

                  echo "==> anchor-check"
                  nix run .#anchor-check

                  echo "==> release-notes-check"
                  nix run .#release-notes-check

                  echo "==> pin-sync (needs network)"
                  nix run .#pin-sync

                  echo "==> spec-drift (needs network)"
                  nix run .#spec-drift

                  echo "==> apidiff (needs network)"
                  nix run .#apidiff

                  echo "pre-release: all gates passed"
                '';
              };
            in
            {
              type = "app";
              program = pkgs.lib.getExe pre-release;
              meta.description = "Run the full local release gate (build, vet, race, lint, flake check, docs, spec-drift, pin-sync, apidiff)";
            };
        };
    };
}
