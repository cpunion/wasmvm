#!/usr/bin/env python3
"""Verify and unpack unchanged, pinned upstream WASI consumers offline."""
import hashlib
import io
import json
from pathlib import Path, PurePosixPath
import tarfile

ROOT = Path(__file__).resolve().parents[1]


def prepare(destination=None):
    inputs = ROOT / "testdata/wasi-preview3/upstream"
    manifest = json.loads((inputs / "manifest.json").read_text())
    sha = lambda data: hashlib.sha256(data).hexdigest()
    archive = (inputs / "fixtures.tar.gz").read_bytes()
    if sha(archive) != manifest["archiveSHA256"]:
        raise ValueError("changed upstream WASI archive")
    if sha((ROOT / "scripts/build_wasi_upstream_fixtures.py").read_bytes()) != manifest["generatorSHA256"]:
        raise ValueError("changed upstream WASI producer")
    if len(manifest["inventory"]) != 61 or len(manifest["cases"]) != 60 or len(manifest["excluded"]) != 1:
        raise ValueError("changed upstream WASI inventory")
    if set(manifest["inventory"]) != set(manifest["cases"]) | set(manifest["excluded"]):
        raise ValueError("incomplete upstream WASI classification")
    destination = Path(destination or ROOT / "artifacts/components/wasi-upstream")
    destination.mkdir(parents=True, exist_ok=True)
    expected = manifest["sha256"]
    with tarfile.open(fileobj=io.BytesIO(archive)) as source:
        members = source.getmembers()
        if len(members) != len(expected) or {member.name for member in members} != set(expected):
            raise ValueError("changed upstream WASI archive inventory")
        for member in members:
            path = PurePosixPath(member.name)
            if not member.isfile() or path.is_absolute() or ".." in path.parts:
                raise ValueError("unsafe upstream WASI archive member")
            if sha(source.extractfile(member).read()) != expected[member.name]:
                raise ValueError("changed upstream WASI asset: " + member.name)
        source.extractall(destination, filter="data")
    for name, want in expected.items():
        if sha((destination / name).read_bytes()) != want:
            raise ValueError("changed extracted upstream WASI asset: " + name)
    print("Verified 60 upstream WASI consumers, one composed service and original source/licenses offline")
    return manifest


if __name__ == "__main__":
    prepare()
