# go module vendor hash for buildGoModule.
#
# Extracted from flake.nix so dependency bumps produce a one-line diff here
# instead of touching (and reformatting) the whole flake. Update with:
#   nix run nixpkgs#nix-update -- wise-go-test
# or set vendorHash = lib.fakeHash, run `nix build`, and paste the got: hash.
"sha256-IeIRBgChUgX5pqLNFOTfWHVOpzsN5TsXBmDwUoYSCFE="
