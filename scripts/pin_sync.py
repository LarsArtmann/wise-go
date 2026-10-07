#!/usr/bin/env python3
"""Verify that every in-house Go dependency is pinned consistently.

A dependency version lives in two places: the `require` block of `go.mod`
(what the compiler builds against) and the `rev=`/tag of the matching
`flake.nix` input (what the hermetic Nix build vendors). When those two drift
apart, `go test` and `nix flake check` silently verify DIFFERENT versions. That
is not hypothetical: go-retry's flake pin sat at v0.6.0 for ten days while
`go.mod` named v0.7.1, and every local gate stayed green because nothing
compared the pair.

Two modes:

* ``--offline`` (hermetic, runs inside `nix flake check`): parses `flake.nix`,
  `flake.lock` and `go.mod` and asserts that each in-house dependency required
  by `go.mod` has a matching flake input, that the input is pinned to an
  immutable full-SHA rev (or a version tag), that `flake.lock`'s locked rev
  equals the `flake.nix`-declared rev, and that any version tag carries exactly
  the version `go.mod` requires.

* online (default, ``nix run .#pin-sync``): additionally resolves each version
  tag on the remote and asserts it points at the pinned rev, catching a pin
  left stale after a `go get` bump.
"""

from __future__ import annotations

import argparse
import json
import pathlib
import re
import subprocess
import sys

OWNER = "LarsArtmann"
GO_MOD_OWNER = "github.com/larsartmann"

FLAKE_INPUT_RE = re.compile(
    r"github:" + OWNER + r"/(?P<name>[A-Za-z0-9._-]+)"
    r"(?:\?rev=(?P<rev>[0-9a-f]{40}))?"
    r"(?:/(?P<tag>v[0-9][^\"\s]*))?"
)
GO_MOD_RE = re.compile(
    r"^\s*" + re.escape(GO_MOD_OWNER) + r"/(?P<name>[A-Za-z0-9._-]+)\s+(?P<version>v[0-9][^\s]*)",
    re.MULTILINE,
)


def parse_go_mod(text: str) -> dict[str, str]:
    return {m.group("name"): m.group("version") for m in GO_MOD_RE.finditer(text)}


def parse_flake_inputs(text: str) -> dict[str, dict[str, str | None]]:
    inputs: dict[str, dict[str, str | None]] = {}
    for m in FLAKE_INPUT_RE.finditer(text):
        inputs[m.group("name")] = {"rev": m.group("rev"), "tag": m.group("tag")}
    return inputs


def parse_flake_lock(text: str) -> dict[str, str]:
    locked: dict[str, str] = {}
    for node in json.loads(text).get("nodes", {}).values():
        entry = node.get("locked")
        if entry and entry.get("type") == "github" and entry.get("owner") == OWNER:
            locked[entry["repo"]] = entry.get("rev")
    return locked


def remote_tag_rev(repo: str, tag: str) -> str:
    result = subprocess.run(
        [
            "git",
            "ls-remote",
            f"https://github.com/{OWNER}/{repo}",
            f"refs/tags/{tag}",
            f"refs/tags/{tag}^{{}}",
        ],
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or "git ls-remote failed")

    plain = peeled = None
    for line in result.stdout.splitlines():
        sha, _, ref = line.partition("\t")
        if ref.endswith("^{}"):
            peeled = sha
        else:
            plain = sha

    return peeled or plain


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--offline", action="store_true", help="skip remote tag resolution")
    parser.add_argument("--root", default=".", help="repository root to inspect")
    args = parser.parse_args()
    root = pathlib.Path(args.root)

    go_mod = parse_go_mod((root / "go.mod").read_text())
    flake = parse_flake_inputs((root / "flake.nix").read_text())
    lock = parse_flake_lock((root / "flake.lock").read_text())

    errors: list[str] = []
    checked = 0
    for name, version in go_mod.items():
        inp = flake.get(name)
        if inp is None:
            errors.append(f"{name}: go.mod requires {version} but flake.nix has no matching input")
            continue

        checked += 1
        rev, tag = inp["rev"], inp["tag"]

        if rev is None and tag is None:
            errors.append(f"{name}: input is not immutably pinned (no ?rev= and no /vX.Y.Z)")

        if tag is not None and tag != version:
            errors.append(f"{name}: flake tag {tag} does not match go.mod {version}")

        if rev is not None:
            locked = lock.get(name)
            if locked is None:
                errors.append(f"{name}: flake.lock has no locked rev")
            elif locked != rev:
                errors.append(f"{name}: flake.lock rev {locked} != flake.nix rev {rev}")

        if not args.offline and rev is not None:
            try:
                actual = remote_tag_rev(name, version)
            except RuntimeError as exc:
                errors.append(f"{name}: could not resolve tag {version} on the remote: {exc}")
                continue
            if actual != rev:
                errors.append(
                    f"{name}: pinned rev {rev} != commit {actual} of tag {version}"
                    " (flake pin is stale versus go.mod)"
                )

    if errors:
        for error in errors:
            print(f"FAIL: {error}")
        sys.exit(1)

    mode = "offline" if args.offline else "online"
    print(f"pin-sync ({mode}): {checked} in-house dependency pins verified")


if __name__ == "__main__":
    main()
