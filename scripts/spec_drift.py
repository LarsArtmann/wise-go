#!/usr/bin/env python3
"""Check the SDK's API surface against the live Wise API reference.

The SDK vendored an OpenAPI 3.1 snapshot (``docs/reviews/wise-api-openapi.json``)
that goes stale as Wise rolls spec versions forward (2026Q3 -> 2027Q1 happened
without any signal). Two questions drift independently, so this script answers
both:

1. **Does the SDK still match the live API?** Every ``(method, path)`` pair the
   SDK's client helpers call is extracted from the root ``*.go`` files and
   matched (structurally, version-prefix-stripped) against the live spec. A
   missing path or dropped method is a breaking drift for this SDK -> exit 1.

2. **How far has the vendored snapshot drifted?** Added/removed/changed paths
   between the vendored snapshot and the live spec are reported (informational
   by default, ``--fail-on-drift`` to enforce). New paths are the
   demand-gated expansion candidates reviewed against FEATURES.md.

Exit codes: 0 = SDK matches live spec; 1 = SDK endpoint missing/mismatched;
2 = ``--fail-on-drift`` and the vendored snapshot is stale; 3 = fetch/parse
failure or no SDK endpoints extracted (a silent-skip would defeat the check).
"""

from __future__ import annotations

import argparse
import json
import pathlib
import re
import sys
import urllib.error
import urllib.request

LIVE_SPEC_URL = "https://docs.wise.com/_bundle/api-reference/@preview/index.json?download"
DEFAULT_VENDORED = "docs/reviews/wise-api-openapi.json"

HTTP_VERBS = {"get", "post", "put", "patch", "delete", "head", "options", "trace"}

CLIENT_CALL_RE = re.compile(
    r"(?:\b(?:c|client)\.(?:get|getWithQuery|getWithQueryHeaders|getRaw|post|postWithHeaders|put|patch|delete)\b)"
    r"|\b(?:fetchByID|listTransfersPage)\("
)
SPRINTF_PATH_RE = re.compile(r'fmt\.Sprintf\(\s*"(/[^"]*)"')
LITERAL_PATH_RE = re.compile(r'path\s*:?=\s*"(/[^"]*)"')
CALL_WINDOW_PATH_RE = re.compile(r'"(/[^"]*)"')
FORMAT_VERB_RE = re.compile(r"%[-#0 +]*\d*(?:\.\d+)?[a-zA-Z]")
VERSION_PREFIX_RE = re.compile(r"^/(?:v\d+|\d{4}Q[1-4]|\{\})(?=/)")


def load_spec(spec_file: str | None, url: str) -> dict:
    if spec_file:
        with open(spec_file, encoding="utf-8") as f:
            return json.load(f)
    req = urllib.request.Request(url, headers={"Accept": "application/json"})
    with urllib.request.urlopen(req, timeout=60) as resp:
        return json.loads(resp.read().decode("utf-8"))


def extract_sdk_endpoints(repo: pathlib.Path) -> list[dict]:
    """Extract (method, raw path template) pairs from root *.go client files."""
    endpoints: list[dict] = []
    for go_file in sorted(repo.glob("*.go")):
        if go_file.name.endswith("_test.go"):
            continue
        lines = go_file.read_text(encoding="utf-8").splitlines()
        last_path: str | None = None
        last_path_line = -100
        for lineno, line in enumerate(lines, 1):
            m = SPRINTF_PATH_RE.search(line) or LITERAL_PATH_RE.search(line)
            if m:
                last_path = m.group(1)
                last_path_line = lineno
            method = request_method(line)
            if method is None:
                continue
            window = "\n".join(lines[lineno - 1 : lineno + 3])
            inline = CALL_WINDOW_PATH_RE.search(window)
            if inline:
                template = inline.group(1)
            elif last_path is not None and lineno - last_path_line <= 15:
                template = last_path
            else:
                template = None
            if template:
                endpoints.append(
                    {
                        "method": method,
                        "template": template,
                        "file": go_file.name,
                        "line": lineno,
                    }
                )
    return endpoints


def request_method(line: str) -> str | None:
    if CLIENT_CALL_RE.search(line) is None:
        return None
    if "fetchByID" in line or "listTransfersPage" in line:
        return "GET"
    for fragment, method in (
        ("getWithQueryHeaders", "GET"),
        ("getWithQuery", "GET"),
        ("getRaw", "GET"),
        ("get", "GET"),
        ("postWithHeaders", "POST"),
        ("post", "POST"),
        ("put", "PUT"),
        ("patch", "PATCH"),
        ("delete", "DELETE"),
    ):
        if re.search(r"\b(?:c|client)\." + fragment + r"\b", line):
            return method
    return None


def normalize_sdk_path(template: str) -> str:
    path = template.split("?", 1)[0]
    path = FORMAT_VERB_RE.sub("{}", path)
    return VERSION_PREFIX_RE.sub("", path)


def normalize_spec_path(path: str) -> str:
    return re.sub(r"\{[^}]*\}", "{}", path)


def segments_match(a: str, b: str) -> bool:
    return a == b or a == "{}" or b == "{}"


def paths_match(sdk_norm: str, spec_norm: str) -> bool:
    a, b = sdk_norm.split("/"), spec_norm.split("/")
    return len(a) == len(b) and all(segments_match(x, y) for x, y in zip(a, b))


def build_spec_index(spec: dict) -> dict[str, set[str]]:
    index: dict[str, set[str]] = {}
    for path, item in spec.get("paths", {}).items():
        if not isinstance(item, dict):
            continue
        index[path] = {m.upper() for m in item if m in HTTP_VERBS}
    return index


def match_endpoint(index: dict[str, set[str]], method: str, sdk_norm: str) -> tuple[str, list[str]]:
    """Return (verdict, matched spec paths): 'ok' | 'method' | 'missing'."""
    exact = index.get(sdk_norm)
    if exact is not None:
        return ("ok", [sdk_norm]) if method in exact else ("method", [sdk_norm])
    structural = [
        path for path, methods in index.items() if paths_match(sdk_norm, normalize_spec_path(path))
    ]
    if not structural:
        return "missing", []
    if any(method in index[path] for path in structural):
        return "ok", structural
    return "method", structural


def spec_label(spec: dict) -> str:
    info = spec.get("info", {})
    version = info.get("version") or "(no version)"
    return f"OpenAPI {spec.get('openapi', '?')}, {version}, {len(spec.get('paths', {}))} paths"


def spec_drift(vendored: dict, live: dict) -> dict:
    v = build_spec_index(vendored)
    l = build_spec_index(live)
    return {
        "added": sorted(set(l) - set(v)),
        "removed": sorted(set(v) - set(l)),
        "changed": sorted(p for p in set(v) & set(l) if v[p] != l[p]),
        "vendored_version": vendored.get("info", {}).get("version") or "(no version)",
        "live_version": live.get("info", {}).get("version") or "(no version)",
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--url", default=LIVE_SPEC_URL, help="live spec URL to fetch")
    parser.add_argument("--spec-file", help="use a local spec JSON instead of downloading")
    parser.add_argument(
        "--repo",
        help="repo root (default: the git toplevel containing this script)",
    )
    parser.add_argument("--fail-on-drift", action="store_true", help="exit 2 when the vendored snapshot is stale")
    parser.add_argument("--verbose", action="store_true", help="list every extracted endpoint and its match")
    args = parser.parse_args()

    repo = pathlib.Path(args.repo) if args.repo else pathlib.Path(__file__).resolve().parent.parent
    try:
        live = load_spec(args.spec_file, args.url)
    except (OSError, json.JSONDecodeError, urllib.error.URLError) as err:
        print(f"FAIL: could not load live spec: {err}", file=sys.stderr)
        return 3
    if not isinstance(live.get("paths"), dict) or not live["paths"]:
        print("FAIL: live spec has no paths; refusing to judge on a broken document", file=sys.stderr)
        return 3
    index = build_spec_index(live)

    endpoints = extract_sdk_endpoints(repo)
    unique = {(e["method"], normalize_sdk_path(e["template"])) for e in endpoints}
    if not unique:
        print(f"FAIL: extracted 0 SDK endpoints from {repo}/*.go; extraction is broken", file=sys.stderr)
        return 3

    print(f"Live spec:  {spec_label(live)}")
    print(f"SDK surface: {len(unique)} unique (method, path) endpoints extracted from root *.go files")

    failures: list[str] = []
    for method, norm in sorted(unique):
        verdict, matched = match_endpoint(index, method, norm)
        if args.verbose:
            tag = "OK" if verdict == "ok" else "PROBLEM"
            print(f"  [{tag}] {method} {norm}" + (f"  ->  {', '.join(matched)}" if verdict != "ok" else ""))
        if verdict == "missing":
            failures.append(f"MISSING: {method} {norm} matches no path in the live spec")
        elif verdict == "method":
            offered = sorted(set().union(*(index[p] for p in matched))) if matched else []
            failures.append(f"METHOD GONE: {method} {norm} - live spec offers {offered} on {matched}")
    if failures:
        print(f"\nSDK vs live spec: {len(failures)} problem(s)")
        for failure in failures:
            print(f"  {failure}")
        status = 1
    else:
        print("SDK vs live spec: all endpoints match")
        status = 0

    vendored_file = repo / DEFAULT_VENDORED
    if vendored_file.exists():
        vendored = load_spec(str(vendored_file), args.url)
        drift = spec_drift(vendored, live)
        print(f"\nVendored snapshot: {spec_label(vendored)}")
        print(f"Drift vs live: +{len(drift['added'])} added / -{len(drift['removed'])} removed / ~{len(drift['changed'])} changed paths")
        if args.verbose:
            for label, entries in (("added", drift["added"]), ("removed", drift["removed"]), ("changed methods", drift["changed"])):
                for entry in entries:
                    print(f"  {label}: {entry}")
        if drift["added"] or drift["removed"] or drift["changed"]:
            print("Re-vendor docs/reviews/wise-api-openapi.json when adopting; new paths are expansion candidates for FEATURES.md")
            if args.fail_on_drift and status == 0:
                status = 2
    else:
        print(f"\nVendored snapshot not found at {vendored_file}; skipping drift report")

    return status


if __name__ == "__main__":
    sys.exit(main())
