#!/usr/bin/env python3
"""Regenerate or verify complete affected core-extension proposal scripts."""

import argparse
import json
from pathlib import Path
import shutil
import subprocess

from core2_corpus import digest, extract, fetch, inventory

ROOT = Path(__file__).resolve().parent.parent
PRODUCER = "wasm-tools 1.262.0 (e22db44da 2026-10-13)"
PROPOSALS = {
    "custom-page-sizes": {
        "revision": "61c873a1860acad2772539164f284cc5110f1b1b",
        "size": 15035559,
        "sha256": "79512780fbb76606fb5caed84f7811705f9467a7153be0d438b4feadc76fae2e",
        "scripts": ["custom-page-sizes/custom-page-sizes-invalid.wast",
                    "custom-page-sizes/custom-page-sizes.wast",
                    "custom-page-sizes/memory_max.wast", "custom-page-sizes/memory_max_i64.wast"],
    },
    "compact-import-section": {
        "revision": "0624de829e2663da799e708444dec80337a3fe54",
        "size": 15032531,
        "sha256": "807aa7ef0dda6e5b33aa8e5fdcc6f628d15d341c35c63426878e1b3fb0a21406",
        "scripts": ["compact-import-section/binary-compact-imports.wast",
                    "compact-import-section/imports-compact.wast"],
    },
    "wide-arithmetic": {
        "revision": "024d4ce1b28cee7fd006d9c5175e3b7ece52b429",
        "size": 15040481,
        "sha256": "7547bb29d8b3d8d395f2c67be85dc12c88404c14629eb4005f1ff17da8050820",
        "scripts": ["wide-arithmetic.wast"],
    },
}


def regenerate(name, producer, cache):
    producer = str(Path(producer).resolve())
    pin = PROPOSALS[name]
    version = subprocess.check_output([producer, "--version"], text=True).strip()
    if version != PRODUCER:
        raise ValueError("producer version differs from pin")
    source = {"url": f"https://codeload.github.com/WebAssembly/{name}/tar.gz/{pin['revision']}",
              "size": pin["size"], "sha256": pin["sha256"]}
    archive = fetch(source, cache)
    root = ROOT / "testdata" / name
    if root.exists():
        raise ValueError("remove the previous generated corpus before regeneration")
    prefix = f"{name}-{pin['revision']}/"

    def select(path):
        if path == prefix + "test/LICENSE":
            return "LICENSE.tests"
        if path in [prefix + "test/core/" + s for s in pin["scripts"]]:
            return "source/" + path[len(prefix + "test/core/"):]

    extract(archive, root, select)
    rows = {}
    for path in pin["scripts"]:
        source_path = "source/" + path
        script_name = Path(path).stem
        dest = root / script_name
        dest.mkdir()
        subprocess.run([producer, "json-from-wast", source_path, "--wasm-dir", script_name,
                        "-o", script_name + "/script.json"], cwd=root, check=True)
        script = json.loads((dest / "script.json").read_text())
        assertions = text = 0
        for cmd in script["commands"]:
            if cmd.get("module_type") == "text":
                if cmd["type"] != "assert_malformed":
                    raise ValueError("executable text was not encoded: " + source_path)
                text += 1
            elif cmd["type"].startswith("assert_"):
                assertions += 1
        rows[script_name] = {"commands": len(script["commands"]),
                             "binaryAssertions": assertions, "textAssertions": text}
    manifest = {"source": "https://github.com/WebAssembly/" + name,
                "revision": pin["revision"], "sourceArchive": source,
                "producer": version, "producerBinarySHA256": digest(Path(producer)),
                "arguments": ["json-from-wast", "--wasm-dir"],
                "scripts": rows, "sha256": inventory(root)}
    (root / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    verify(name)


def verify(name):
    root = ROOT / "testdata" / name
    manifest = json.loads((root / "manifest.json").read_text())
    actual = inventory(root)
    del actual["manifest.json"]
    pin = PROPOSALS[name]
    if (manifest["revision"] != pin["revision"] or manifest["producer"] != PRODUCER
            or manifest["sourceArchive"]["sha256"] != pin["sha256"]
            or manifest["sha256"] != actual):
        raise ValueError("source, producer or fixture inventory changed: " + name)
    print(name, "files=", len(actual), "scripts=", manifest["scripts"])


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--generate", action="store_true")
    parser.add_argument("--producer", default="wasm-tools")
    parser.add_argument("--cache", type=Path, default=ROOT / "artifacts/core-extensions/cache")
    parser.add_argument("--proposal", choices=PROPOSALS)
    args = parser.parse_args()
    producer = shutil.which(args.producer) if args.generate else None
    for name in PROPOSALS:
        if args.proposal and name != args.proposal:
            continue
        if args.generate:
            if not producer:
                raise ValueError("pinned producer is required for regeneration")
            regenerate(name, producer, args.cache)
        else:
            verify(name)


if __name__ == "__main__":
    main()
