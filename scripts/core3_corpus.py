#!/usr/bin/env python3
"""Prepare or deterministically regenerate the pinned complete Core 3.0 inputs."""
import argparse
from collections import Counter
import gzip
import hashlib
import io
import json
from pathlib import Path
import shutil
import subprocess
import tarfile
import tempfile

from core2_corpus import digest, extract, fetch, inventory, relative_path

ROOT = Path(__file__).resolve().parent.parent
RECORDS = ROOT / "testdata/core3"
REVISION = "aa4dcf44b8de3f3f3270ca067422172518ecc341"
COMMANDS = {"module", "module_definition", "module_instance", "register", "action",
            "assert_return", "assert_trap", "assert_exception", "assert_exhaustion",
            "assert_invalid", "assert_malformed", "assert_unlinkable", "assert_uninstantiable"}


def describe(root):
    rows = []
    for source in sorted((root / "source").rglob("*.wast")):
        name = source.relative_to(root / "source").with_suffix("").as_posix().replace("/", "__")
        script = json.loads((root / name / "script.json").read_text())
        if script["source_filename"] != source.relative_to(root).as_posix():
            raise ValueError("source identity changed: " + name)
        counts = Counter()
        binary, text = 0, 0
        for command in script["commands"]:
            kind = command["type"]
            if kind not in COMMANDS:
                raise ValueError("unknown command: " + kind)
            counts[kind] += 1
            if command.get("module_type") == "text":
                if kind not in ("assert_malformed", "assert_invalid"):
                    raise ValueError("unconverted executable text: " + name)
                text += 1
            elif kind.startswith("assert_"):
                binary += 1
            if "filename" in command:
                filename = root / name / relative_path(command["filename"])
                if not filename.is_file():
                    raise ValueError("missing command input: " + str(filename))
        rows.append({"name": name, "source": source.relative_to(root).as_posix(),
                     "sourceSHA256": digest(source), "commands": dict(sorted(counts.items())),
                     "binaryAssertions": binary, "textParserAssertions": text})
    return rows


def verify(root, records=RECORDS):
    manifest = json.loads((records / "manifest.json").read_text())
    if manifest["revision"] != REVISION or manifest["pinsSHA256"] != digest(records / "pins.json"):
        raise ValueError("source or producer pin changed")
    if inventory(root) != manifest["sha256"]:
        raise ValueError("Core 3.0 file inventory or checksum changed")
    if describe(root) != manifest["scripts"]:
        raise ValueError("Core 3.0 command inventory changed")
    return manifest


def prepare(output, records=RECORDS):
    pins = json.loads((records / "pins.json").read_text())
    archive = records / pins["fixtures"]["file"]
    if archive.stat().st_size != pins["fixtures"]["size"] or digest(archive) != pins["fixtures"]["sha256"]:
        raise ValueError("committed Core 3.0 archive changed")
    root = output / "fixtures"
    if not root.exists():
        output.mkdir(parents=True, exist_ok=True)
        with tempfile.TemporaryDirectory(dir=output) as temporary:
            staging = Path(temporary)
            extract(archive, staging, lambda name: str(relative_path(name)))
            verify(staging, records)
            staging.rename(root)
    manifest = verify(root, records)
    evidence = {"profile": "core3", "revision": REVISION, "manifestSHA256": digest(records / "manifest.json"),
                **{k: manifest[k] for k in ("scriptCount", "commandCount", "binaryAssertions", "textParserAssertions")}}
    (output / "evidence.json").write_text(json.dumps(evidence, indent=2) + "\n")
    return manifest


def regenerate(output, producer, source_archive=None):
    pins = json.loads((RECORDS / "pins.json").read_text())
    if subprocess.check_output([producer, "--version"], text=True).strip() != pins["producer"]["version"]:
        raise ValueError("wasm-tools version differs from pin")
    archive = Path(source_archive) if source_archive else fetch(pins["source"], output / "cache")
    if archive.stat().st_size != pins["source"]["size"] or digest(archive) != pins["source"]["sha256"]:
        raise ValueError("specification archive differs from pin")
    prefix = "spec-" + REVISION + "/"
    root = output / "regenerated"
    if root.exists():
        raise ValueError("regeneration destination already exists")
    def select(name):
        if name == prefix + "test/LICENSE":
            return "LICENSE.tests"
        if name.startswith(prefix + "test/core/") and name.endswith(".wast"):
            return "source/" + name[len(prefix + "test/core/"):]
    extract(archive, root, select)
    corrections, transcodes = [], []
    for source in sorted((root / "source").rglob("*.wast")):
        name = source.relative_to(root / "source").with_suffix("").as_posix().replace("/", "__")
        dest = root / name
        dest.mkdir()
        original = source.read_bytes()
        if name == "gc__type-subtyping":
            needle = b"(sub $parent1 $parent2 (struct))"
            if original.count(needle) != 1:
                raise ValueError("multiple-supertype producer input changed")
            source.write_bytes(original.replace(needle, b"(sub $parent1 (struct))"))
        try:
            subprocess.run([producer, "json-from-wast", source.relative_to(root).as_posix(),
                            "--wasm-dir", name, "-o", name + "/script.json"], cwd=root, check=True)
        finally:
            source.write_bytes(original)
        script = json.loads((dest / "script.json").read_text())
        for command in script["commands"]:
            if name == "gc__type-subtyping" and command["line"] == 955 and command["type"] == "assert_invalid":
                p = dest / command["filename"]
                before = p.read_bytes()
                if hashlib.sha256(before).hexdigest() != "06389b1ca09c0b64d1073d9b23819affdbd3c68f37884b0bc55d0d11ba2318ce":
                    raise ValueError("multiple-supertype producer binary changed")
                after = bytearray(before.replace(bytes.fromhex("5001005f00"), bytes.fromhex("500200015f00")))
                after[9] += 1  # Type-section payload grows by one parent index.
                if hashlib.sha256(after).hexdigest() != "c4572c3e640e1e2d4047d20040d77e08613a951ae113ea06ee9b822de7d93a22":
                    raise ValueError("multiple-supertype correction changed")
                p.write_bytes(after)
                corrections.append({"source": "source/gc/type-subtyping.wast", "line": 955,
                                    "fixture": name + "/" + p.name, "beforeSHA256": hashlib.sha256(before).hexdigest(),
                                    "afterSHA256": digest(p), "reason": "Restore both original supertype indices after the producer rejects their text form."})
            if command.get("module_type") == "text" and command["type"] == "module":
                p = dest / command["filename"]
                binary = p.with_suffix(".transcoded.wasm")
                subprocess.run([producer, "parse", str(p), "-o", str(binary)], check=True)
                transcodes.append({"source": name + "/" + p.name, "binary": name + "/" + binary.name,
                                   "line": command["line"]})
                command["filename"], command["module_type"] = binary.name, "binary"
        (dest / "script.json").write_text(json.dumps(script, separators=(",", ":")) + "\n")
    rows = describe(root)
    counts = {"scriptCount": len(rows), "commandCount": sum(sum(r["commands"].values()) for r in rows),
              "binaryAssertions": sum(r["binaryAssertions"] for r in rows),
              "textParserAssertions": sum(r["textParserAssertions"] for r in rows)}
    if list(counts.values()) != [259, 65344, 61395, 1239] or len(corrections) != 1 or len(transcodes) != 10:
        raise ValueError("pinned Core 3.0 inventory changed: " + str(counts))
    bundle = output / "inputs.tar.gz"
    with bundle.open("wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as compressed:
        with tarfile.open(fileobj=compressed, mode="w") as tar:
            for p in sorted(root.rglob("*")):
                if p.is_file():
                    data = p.read_bytes()
                    info = tarfile.TarInfo(p.relative_to(root).as_posix())
                    info.size, info.mode = len(data), 0o644
                    tar.addfile(info, io.BytesIO(data))
    pins["fixtures"] = {"file": "inputs.tar.gz", "size": bundle.stat().st_size, "sha256": digest(bundle)}
    (RECORDS / "pins.json").write_text(json.dumps(pins, indent=2) + "\n")
    manifest = {"revision": REVISION, "pinsSHA256": digest(RECORDS / "pins.json"), **counts,
                "scripts": rows, "producerCorrections": corrections, "textTranscodes": transcodes, "sha256": inventory(root)}
    (RECORDS / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    shutil.copyfile(bundle, RECORDS / "inputs.tar.gz")
    shutil.copyfile(root / "LICENSE.tests", RECORDS / "LICENSE.tests")
    (output / "producer.json").write_text(json.dumps({"version": pins["producer"]["version"], "binarySHA256": digest(Path(producer))}, indent=2) + "\n")
    verify(root)
    return manifest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, default=ROOT / "artifacts/core3")
    parser.add_argument("--generate", action="store_true")
    parser.add_argument("--producer", default="wasm-tools")
    parser.add_argument("--source-archive")
    args = parser.parse_args()
    if args.generate:
        producer = shutil.which(args.producer)
        if not producer:
            raise ValueError("pinned wasm-tools is required for regeneration")
        manifest = regenerate(args.output, producer, args.source_archive)
    else:
        manifest = prepare(args.output)
    print(f"Core 3.0: {manifest['scriptCount']} scripts, {manifest['binaryAssertions']} binary assertions, "
          f"{manifest['textParserAssertions']} text-parser assertions; input verification is separate from execution.")


if __name__ == "__main__":
    main()
