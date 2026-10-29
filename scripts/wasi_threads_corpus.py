#!/usr/bin/env python3
"""Verify and prepare the complete pinned legacy thread corpus offline."""
import hashlib
import io
import json
from pathlib import Path
import tarfile

ROOT = Path(__file__).resolve().parent.parent


def prepare(destination=None):
    inputs = ROOT / "testdata/wasi-threads"
    manifest = json.loads((inputs / "manifest.json").read_text())
    archive = (inputs / "inputs.tar.gz").read_bytes()
    if hashlib.sha256(archive).hexdigest() != manifest["archiveSHA256"]:
        raise ValueError("legacy thread archive hash changed")
    expected = manifest["sha256"]
    destination = Path(destination or ROOT / "artifacts/wasi-threads")
    destination.mkdir(parents=True, exist_ok=True)
    with tarfile.open(fileobj=io.BytesIO(archive)) as source:
        members = source.getmembers()
        if {m.name for m in members} != set(expected) or len(members) != len(expected):
            raise ValueError("legacy thread archive inventory changed")
        for member in members:
            if not member.isfile():
                raise ValueError("non-file thread input")
            data = source.extractfile(member).read()
            if hashlib.sha256(data).hexdigest() != expected[member.name]:
                raise ValueError("legacy thread input hash changed: " + member.name)
        source.extractall(destination, filter="data")
    for name, expected_hash in expected.items():
        if hashlib.sha256((destination / name).read_bytes()).hexdigest() != expected_hash:
            raise ValueError("extracted legacy thread input hash changed: " + name)
    actual = {"official/" + name for name in manifest["official"]}
    if len(actual) != 15 or actual != {p for p in expected if p.startswith("official/") and p.endswith(".wasm")}:
        raise ValueError("incomplete official thread inventory")
    for name, expected_hash in manifest["producers"]["sources"].items():
        if hashlib.sha256((inputs / "guest-llgo" / name).read_bytes()).hexdigest() != expected_hash:
            raise ValueError("LLGo guest source hash changed: " + name)
    print(f"Verified 15 legacy thread programs and {len(expected)} assets")
    return manifest


if __name__ == "__main__":
    prepare()
