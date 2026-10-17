#!/usr/bin/env python3
"""Parity gate for the spec_drift endpoint extractor: the regex extractor
(legacy) and the go/parser extractor must produce the SAME unique
(method, normalized template) set on this repo. Run from the repo root:

    python3 scripts/spec_drift_parity.py

Wired into nothing automated yet; run it after changing either extractor or
the SDK's call shapes. Exit 0 = parity, 1 = drift between extractors.
"""

from __future__ import annotations

import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

import spec_drift


def main() -> int:
    repo = pathlib.Path(__file__).resolve().parent.parent
    old_eps = spec_drift.extract_sdk_endpoints(repo)
    old_set = {
        (e["method"], spec_drift.normalize_sdk_path(e["template"])) for e in old_eps
    }
    new_eps = spec_drift.extract_sdk_endpoints_go(repo)
    new_set = {
        (e["method"], spec_drift.normalize_sdk_path(e["template"])) for e in new_eps
    }

    print(f"regex extractor: {len(old_set)} unique endpoints")
    print(f"go/parser extractor: {len(new_set)} unique endpoints")

    only_old = sorted(old_set - new_set)
    only_new = sorted(new_set - old_set)
    for entry in only_old:
        print(f"ONLY-REGEX: {entry[0]} {entry[1]}")
    for entry in only_new:
        print(f"ONLY-GO: {entry[0]} {entry[1]}")

    if only_old or only_new:
        print("FAIL: extractor parity broken")
        return 1
    print("parity: EQUAL")
    return 0


if __name__ == "__main__":
    sys.exit(main())
