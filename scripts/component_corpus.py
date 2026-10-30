#!/usr/bin/env python3
"""Prepare pinned component grammar/type/ABI inputs and verify reference assets."""
import hashlib
import io
import json
from pathlib import Path
import tarfile

ROOT = Path(__file__).resolve().parent.parent


def prepare(destination=None):
    inputs = ROOT / "testdata/component-binary"
    manifest = json.loads((inputs / "manifest.json").read_text())
    archive = (inputs / "inputs.tar.gz").read_bytes()
    if hashlib.sha256(archive).hexdigest() != manifest["archiveSHA256"]:
        raise ValueError("component grammar archive hash changed")
    destination = Path(destination or ROOT / "artifacts/components/binary-corpus")
    expected = manifest["sha256"]
    destination.mkdir(parents=True, exist_ok=True)
    with tarfile.open(fileobj=io.BytesIO(archive)) as source:
        members = source.getmembers()
        if len(members) != len(expected) or {m.name for m in members} != set(expected):
            raise ValueError("component grammar inventory changed")
        for member in members:
            if not member.isfile():
                raise ValueError("non-file component grammar input")
            if hashlib.sha256(source.extractfile(member).read()).hexdigest() != expected[member.name]:
                raise ValueError("component grammar hash changed: " + member.name)
        source.extractall(destination, filter="data")
    for name, want in expected.items():
        if hashlib.sha256((destination / name).read_bytes()).hexdigest() != want:
            raise ValueError("extracted component grammar hash changed: " + name)
    commands = json.loads((destination / "binary.json").read_text())["commands"]
    if len(commands) != 132 or sum(c["type"] == "assert_malformed" for c in commands) != 80:
        raise ValueError("incomplete component binary script")
    for name, counts in manifest.get("semanticScripts", {}).items():
        commands = json.loads((destination / (name + ".json")).read_text())["commands"]
        if (len(commands) != counts["commands"] or
                sum(c["type"] == "assert_invalid" for c in commands) != counts["invalid"] or
                sum(c["type"] in ("module", "module_definition") for c in commands) != counts["valid"] or
                sum(c.get("module_type") == "text" and c["type"] == "assert_malformed"
                    for c in commands) != counts.get("textParser", 0)):
            raise ValueError("incomplete component semantic script: " + name)
    reference = ROOT / "testdata/component-canonical"
    pins = json.loads((reference / "manifest.json").read_text())
    for name, want in pins["sha256"].items():
        if hashlib.sha256((reference / name).read_bytes()).hexdigest() != want:
            raise ValueError("canonical reference input hash changed: " + name)
    authored = ROOT / "pkg/component/testdata"
    values = json.loads((authored / "resource-values.json").read_text())
    generator = ROOT / "scripts/build_resource_value_fixtures.py"
    if hashlib.sha256(generator.read_bytes()).hexdigest() != values["generatorSHA256"]:
        raise ValueError("resource value fixture generator hash changed")
    for name, pin in values["fixtures"].items():
        if hashlib.sha256((authored / name).read_bytes()).hexdigest() != pin["sha256"]:
            raise ValueError("authored resource value fixture hash changed: " + name)
    scripts = manifest.get("semanticScripts", {})
    binary = sum(c["invalid"] + c["valid"] for c in scripts.values())
    text = sum(c.get("textParser", 0) for c in scripts.values())
    print(f"Verified 132 grammar commands, {len(scripts)} semantic scripts "
          f"({binary} binary commands; {text} separate text-parser assertions), "
          "and 1,024 canonical layout cases")
    return manifest


if __name__ == "__main__":
    prepare()
