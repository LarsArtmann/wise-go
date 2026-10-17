{
  # D5 (wise-go docs/planning/2026-10-16_decision-memos.md): the check matrix
  # is the four default systems minus x86_64-darwin, which nixpkgs-26.11 no
  # longer ships and nothing consumes. Reversible by re-adding the string.
  outputs = _: {
    systems = [
      "x86_64-linux"
      "aarch64-linux"
      "aarch64-darwin"
    ];
  };
}
