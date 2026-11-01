#!/usr/bin/env python3
"""Verify and unpack unchanged, pinned upstream WASI consumers offline."""
import hashlib
import io
import json
from pathlib import Path, PurePosixPath
import tarfile

ROOT = Path(__file__).resolve().parents[1]


def unpack(inputs, destination, counts, generator):
    manifest = json.loads((inputs / "manifest.json").read_text())
    sha = lambda data: hashlib.sha256(data).hexdigest()
    archive = (inputs / "fixtures.tar.gz").read_bytes()
    if sha(archive) != manifest["archiveSHA256"]:
        raise ValueError("changed upstream WASI archive")
    if sha((ROOT / "scripts" / generator).read_bytes()) != manifest["generatorSHA256"]:
        raise ValueError("changed upstream WASI producer")
    if tuple(len(manifest[key]) for key in ("inventory", "cases", "excluded")) != counts:
        raise ValueError("changed upstream WASI inventory")
    if set(manifest["inventory"]) != set(manifest["cases"]) | set(manifest["excluded"]):
        raise ValueError("incomplete upstream WASI classification")
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
    return manifest


def prepare(destination=None):
    destination = Path(destination or ROOT / "artifacts/components/wasi-upstream")
    base = unpack(ROOT / "testdata/wasi-preview3/upstream", destination, (61, 60, 1),
                  "build_wasi_upstream_fixtures.py")
    p2 = unpack(ROOT / "testdata/wasi-preview2/upstream", destination, (95, 92, 3),
                "build_wasi_p2_world_fixtures.py")
    if (p2["sharedArchiveSHA256"] != base["archiveSHA256"] or
            p2["sharedGeneratorSHA256"] != base["generatorSHA256"] or
            p2["sourceArchiveSHA256"] != base["sourceArchiveSHA256"] or
            len(p2["retainedDraftCases"]) != 2 or
            not set(p2["retainedDraftCases"]) < set(p2["excluded"]) or
            set(p2["variants"]) != {"p2_api_proxy_streaming-executor-lifetimes"}):
        raise ValueError("changed P2 shared provenance or draft inventory")
    print("Verified 60 P3 and 92 stable P2 consumers, two retained P2 drafts, "
          "one composed service, one explicit executor stress variant and shared sources/notices offline")
    return base


if __name__ == "__main__":
    prepare()
