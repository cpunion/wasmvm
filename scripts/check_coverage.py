#!/usr/bin/env python3
"""Enforce statement coverage per executable package, never by aggregate."""

import argparse
from pathlib import Path
import re
import sys


BLOCK = re.compile(r"(.+):([0-9]+\.[0-9]+,[0-9]+\.[0-9]+) ([0-9]+) ([0-9]+)")


def package_coverage(text):
    lines = text.splitlines()
    if not lines or lines[0] not in ("mode: set", "mode: count", "mode: atomic"):
        raise ValueError("missing or invalid Go coverage mode")
    blocks = {}
    for line in lines[1:]:
        match = BLOCK.fullmatch(line)
        if match is None:
            raise ValueError(f"invalid coverage block: {line!r}")
        filename, span, statements, count = match.groups()
        if "/" not in filename:
            raise ValueError(f"coverage filename lacks package path: {filename!r}")
        key = filename, span
        statements, count = int(statements), int(count)
        if key in blocks:
            old_statements, old_count = blocks[key]
            if statements != old_statements:
                raise ValueError(f"inconsistent duplicate coverage block: {filename}:{span}")
            count = max(count, old_count)
        blocks[key] = statements, count
    packages = {}
    for (filename, _), (statements, count) in blocks.items():
        if statements == 0:
            continue
        package = filename.rsplit("/", 1)[0]
        covered, total = packages.get(package, (0, 0))
        packages[package] = covered + (statements if count else 0), total + statements
    if not packages:
        raise ValueError("coverage profile contains no executable statements")
    return packages


def check(text, minimum, required=()):
    packages = package_coverage(text)
    missing = set(required) - packages.keys()
    if missing:
        raise ValueError("missing required coverage: " + ", ".join(sorted(missing)))
    failed = False
    for package, (covered, total) in sorted(packages.items()):
        percentage = covered * 100 / total
        print(f"{package}: {percentage:.2f}% ({covered}/{total} statements)")
        if covered * 100 < minimum * total:
            failed = True
    return not failed


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("profile", type=Path)
    parser.add_argument("--minimum", type=float, default=90)
    parser.add_argument("--require", action="append", default=[])
    args = parser.parse_args()
    if not 0 <= args.minimum <= 100:
        parser.error("--minimum must be between 0 and 100")
    try:
        passed = check(args.profile.read_text(), args.minimum, args.require)
    except (OSError, ValueError) as error:
        print(error, file=sys.stderr)
        return 2
    return 0 if passed else 1


if __name__ == "__main__":
    sys.exit(main())
