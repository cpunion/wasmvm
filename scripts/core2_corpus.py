#!/usr/bin/env python3
"""Generate or verify the complete pinned Core 2.0 binary test inputs."""

import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path, PurePosixPath
import platform
import re
import subprocess
import sys
import tarfile
import tempfile
import urllib.request


ROOT = Path(__file__).resolve().parent.parent
RECORDS = ROOT / "testdata/core2"
LICENSES = {"LICENSE": "LICENSE.spec", "test/LICENSE": "LICENSE.tests"}
COMMANDS = {
    "module", "register", "action", "assert_return", "assert_trap",
    "assert_exhaustion", "assert_invalid", "assert_malformed",
    "assert_unlinkable", "assert_uninstantiable",
}


def digest(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def relative_path(name):
    path = PurePosixPath(name)
    if (not name or name == "." or path.is_absolute() or ".." in path.parts
            or "\\" in name or str(path) != name):
        raise ValueError(f"invalid corpus path: {name!r}")
    return path


def fetch(pin, cache):
    if not re.fullmatch(r"[0-9a-f]{64}", pin["sha256"]):
        raise ValueError("invalid pinned SHA-256")
    if pin["size"] <= 0 or not pin["url"].startswith("https://"):
        raise ValueError("invalid archive size or URL")
    cache.mkdir(parents=True, exist_ok=True)
    destination = cache / (pin["sha256"] + ".tar.gz")
    if not destination.exists():
        request = urllib.request.Request(pin["url"], headers={"User-Agent": "wasmvm-core2-corpus"})
        with tempfile.NamedTemporaryFile(dir=cache, delete=False) as stream:
            temporary = Path(stream.name)
            try:
                with urllib.request.urlopen(request, timeout=60) as response:
                    if not response.url.startswith("https://"):
                        raise ValueError("archive redirected away from HTTPS")
                    size = 0
                    while chunk := response.read(1024 * 1024):
                        size += len(chunk)
                        if size > pin["size"]:
                            raise ValueError("archive exceeds pinned size")
                        stream.write(chunk)
                stream.flush()
                if size != pin["size"] or digest(temporary) != pin["sha256"]:
                    raise ValueError("downloaded archive size or digest mismatch")
                temporary.replace(destination)
            finally:
                temporary.unlink(missing_ok=True)
    if destination.stat().st_size != pin["size"] or digest(destination) != pin["sha256"]:
        raise ValueError(f"cached archive size or digest mismatch: {destination}")
    return destination


def extract(archive, destination, select):
    """Extract only selected regular files; archive links are never followed."""
    selected = set()
    total = 0
    with tarfile.open(archive) as source:
        for member in source:
            output = select(member.name)
            if output is None:
                continue
            relative_path(output)
            if not member.isfile() or output in selected:
                raise ValueError(f"duplicate or non-regular selected member: {member.name}")
            total += member.size
            if member.size > 64 * 1024 * 1024 or total > 128 * 1024 * 1024:
                raise ValueError("selected archive contents exceed extraction limit")
            target = destination / output
            target.parent.mkdir(parents=True, exist_ok=True)
            with source.extractfile(member) as stream:
                target.write_bytes(stream.read())
            selected.add(output)
    return selected


def extract_source(pin, archive, destination):
    prefix = str(relative_path(pin["prefix"])) + "/"
    directory = str(relative_path(pin["directory"])) + "/"

    def select(name):
        if not name.startswith(prefix):
            return None
        name = name[len(prefix):]
        if name in LICENSES:
            return LICENSES[name]
        if name.startswith(directory) and name.endswith(".wast"):
            return "wast/" + str(relative_path(name[len(directory):]))
        return None

    selected = extract(archive, destination, select)
    scripts = sorted(name[5:] for name in selected if name.startswith("wast/"))
    if len(scripts) != pin["scripts"] or not set(LICENSES.values()) <= selected:
        raise ValueError("source archive has missing scripts or license notices")
    return {name: digest(destination / "wast" / name) for name in scripts}


def inventory(fixtures):
    if fixtures.is_symlink() or not fixtures.is_dir():
        raise ValueError("corpus fixture directory is missing or is a symlink")
    files = {}
    for path in sorted(fixtures.rglob("*")):
        if path.is_symlink() or not (path.is_file() or path.is_dir()):
            raise ValueError(f"non-regular corpus entry: {path}")
        if path.is_file():
            files[path.relative_to(fixtures).as_posix()] = digest(path)
    return files


def hash_text(files):
    return "".join(f"{sha}  {name}\n" for name, sha in sorted(files.items()))


def describe(fixtures, sources, pins_digest):
    scripts = []
    counts = Counter()
    text_assertions = 0
    expected_files = set(LICENSES.values())
    for name, source_hash in sorted(sources.items()):
        path = relative_path(name).with_suffix("") / "script.json"
        expected_files.add(str(path))
        script = json.loads((fixtures / path).read_text())
        if script["source_filename"] != name:
            raise ValueError(f"producer changed source identity: {name}")
        per_script = Counter()
        text_count = 0
        for command in script["commands"]:
            kind = command["type"]
            if kind not in COMMANDS:
                raise ValueError(f"unknown script command: {kind}")
            per_script[kind] += 1
            if "filename" in command:
                referenced = path.parent / relative_path(command["filename"])
                expected_files.add(str(referenced))
                module_type = command.get("module_type", "binary")
                if module_type == "text" and kind == "assert_malformed":
                    text_count += 1
                elif module_type != "binary":
                    raise ValueError(f"unexpected module type: {kind}/{module_type}")
        counts.update(per_script)
        text_assertions += text_count
        scripts.append({"source": name, "sourceSHA256": source_hash, "json": str(path),
                        "commands": dict(sorted(per_script.items())), "textParsingAssertions": text_count})
    files = inventory(fixtures)
    if set(files) != expected_files:
        raise ValueError(f"fixture inventory mismatch: missing={sorted(expected_files - files.keys())}, "
                         f"extra={sorted(files.keys() - expected_files)}")
    assertions = sum(count for kind, count in counts.items() if kind.startswith("assert_"))
    return {"schemaVersion": 1, "pinsSHA256": pins_digest, "scriptCount": len(scripts),
            "fileCount": len(files), "commandCount": sum(counts.values()),
            "commands": dict(sorted(counts.items())), "binaryAssertions": assertions - text_assertions,
            "textParsingAssertions": text_assertions, "scripts": scripts}, files


def verify(fixtures, records=RECORDS):
    manifest = json.loads((records / "manifest.json").read_text())
    if manifest.get("schemaVersion") != 1:
        raise ValueError("unsupported corpus manifest schema")
    if manifest["pinsSHA256"] != digest(records / "pins.json"):
        raise ValueError("pins changed without regenerating the corpus manifest")
    files = inventory(fixtures)
    if hash_text(files) != (records / "files.sha256").read_text():
        raise ValueError("generated corpus does not match the committed file hashes")
    for name in LICENSES.values():
        if digest(records / name) != files[name]:
            raise ValueError(f"committed license notice differs from the pinned input: {name}")
    sources = {script["source"]: script["sourceSHA256"] for script in manifest["scripts"]}
    actual, _ = describe(fixtures, sources, manifest["pinsSHA256"])
    if actual != manifest:
        raise ValueError("script or assertion counts differ from the committed manifest")
    return manifest


def prepare(output, records=RECORDS, update=False):
    fixtures = output / "fixtures"
    if fixtures.exists():
        if update:
            raise ValueError("manifest updates require a fresh output directory")
        return verify(fixtures, records)
    pins = json.loads((records / "pins.json").read_text())
    if pins.get("schemaVersion") != 1 or pins.get("profile") != "Core 2.0 binary":
        raise ValueError("unsupported corpus input pin format")
    host = sys.platform + "-" + platform.machine().lower()
    try:
        producer_pin = pins["producer"]["assets"][host]
    except KeyError as error:
        raise ValueError(f"no pinned producer for {host}") from error
    output.mkdir(parents=True, exist_ok=True)
    archive = fetch(pins["source"], output / "archives")
    producer_archive = fetch(producer_pin, output / "archives")
    with tempfile.TemporaryDirectory(prefix="build-", dir=output) as temporary:
        work = Path(temporary)
        source = work / "source"
        sources = extract_source(pins["source"], archive, source)
        selected = extract(producer_archive, work / "tool",
                           lambda name: "wast2json" if name == producer_pin["binary"] else None)
        if selected != {"wast2json"}:
            raise ValueError("producer archive is missing wast2json")
        producer = work / "tool/wast2json"
        producer.chmod(0o755)
        version = subprocess.check_output([str(producer), "--version"], text=True).strip()
        if version != pins["producer"]["version"]:
            raise ValueError(f"producer version mismatch: {version}")
        generated = work / "fixtures"
        generated.mkdir()
        for name in LICENSES.values():
            (generated / name).write_bytes((source / name).read_bytes())
        for name in sources:
            target = generated / PurePosixPath(name).with_suffix("") / "script.json"
            target.parent.mkdir(parents=True, exist_ok=True)
            command = [str(producer.resolve()), name, *pins["producer"]["arguments"], "-o", str(target.resolve())]
            result = subprocess.run(command, cwd=source / "wast", capture_output=True, text=True, timeout=60)
            if result.returncode:
                raise ValueError(f"corpus conversion failed for {name}:\n{result.stderr}")
        manifest, files = describe(generated, sources, digest(records / "pins.json"))
        if update:
            (records / "manifest.json").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")
            (records / "files.sha256").write_text(hash_text(files))
            for name in LICENSES.values():
                (records / name).write_bytes((source / name).read_bytes())
        else:
            verify(generated, records)
        evidence = {"sourceRevision": pins["source"]["revision"],
                    "producerRevision": pins["producer"]["revision"], "producerVersion": version,
                    "host": host, "producerBinarySHA256": digest(producer),
                    "pinsSHA256": digest(records / "pins.json"),
                    "manifestSHA256": digest(records / "manifest.json"),
                    "fileHashesSHA256": digest(records / "files.sha256"),
                    "scriptCount": manifest["scriptCount"], "commandCount": manifest["commandCount"],
                    "binaryAssertions": manifest["binaryAssertions"],
                    "textParsingAssertions": manifest["textParsingAssertions"],
                    "execution": "inputs generated and verified; no VM assertions executed"}
        generated.rename(fixtures)
        (output / "evidence.json").write_text(json.dumps(evidence, indent=2, sort_keys=True) + "\n")
    return manifest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, default=ROOT / "artifacts/core2")
    parser.add_argument("--update-manifest", action="store_true")
    parser.add_argument("--verify-only", action="store_true")
    args = parser.parse_args()
    if args.update_manifest and args.verify_only:
        parser.error("--update-manifest and --verify-only are mutually exclusive")
    try:
        if args.verify_only:
            manifest = verify(args.output / "fixtures")
        else:
            manifest = prepare(args.output.resolve(), update=args.update_manifest)
    except (OSError, ValueError, KeyError, subprocess.SubprocessError, tarfile.TarError) as error:
        parser.exit(1, f"Core 2.0 corpus preparation failed: {error}\n")
    print(f"Verified {manifest['scriptCount']} scripts, {manifest['commandCount']} commands, "
          f"{manifest['binaryAssertions']} binary assertions and "
          f"{manifest['textParsingAssertions']} text-parser assertions; no VM execution claimed.")


if __name__ == "__main__":
    main()
