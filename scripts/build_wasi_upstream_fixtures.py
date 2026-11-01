#!/usr/bin/env python3
"""Build unchanged Wasmtime P3 consumers with pinned Rust and component tools."""
import argparse
import gzip
import hashlib
import io
import json
import os
from pathlib import Path
import re
import subprocess
import tarfile

ROOT = Path(__file__).resolve().parents[1]
REVISION = "7bac2c2775808aaec5d4aa5627a5e447b51102cf"
ARCHIVE_SHA = "25fe25e2b1494cf8d3f1f74d112a59fca98fee97fec978207ae140525d075761"
EXCLUDED = {"p3_tls_sample_application": "Draft wasi:tls is outside the stable WASI 0.3 profile."}


def sha(data):
    return hashlib.sha256(data).hexdigest()


def run(*args, **kwargs):
    return subprocess.check_output([str(arg) for arg in args], **kwargs)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-archive", type=Path, required=True)
    parser.add_argument("--work", type=Path, required=True)
    parser.add_argument("--wasm-tools", type=Path, required=True)
    parser.add_argument("--output", type=Path, default=ROOT / "testdata/wasi-preview3/upstream")
    args = parser.parse_args()
    archive = args.source_archive.read_bytes()
    if sha(archive) != ARCHIVE_SHA:
        raise ValueError("changed upstream source archive")
    if not run("rustc", "+1.98.0", "--version", text=True).startswith("rustc 1.98.0 "):
        raise ValueError("changed Rust compiler")
    version = run(args.wasm_tools, "--version", text=True).strip()
    if version != "wasm-tools 1.262.0 (e22db44da 2026-10-13)":
        raise ValueError("changed component producer")
    work = args.work.resolve()
    work.mkdir(parents=True, exist_ok=True)
    with tarfile.open(fileobj=io.BytesIO(archive)) as source:
        source.extractall(work, filter="data")
    source = work / ("wasmtime-" + REVISION)
    programs = source / "crates/test-programs"
    inventory = sorted(path.stem for path in (programs / "src/bin").glob("p3_*.rs"))
    if len(inventory) != 61 or not set(EXCLUDED) < set(inventory):
        raise ValueError("changed upstream P3 inventory")
    names = [name for name in inventory if name not in EXCLUDED]
    env = os.environ.copy()
    cargo_home = Path(env.get("CARGO_HOME", Path.home() / ".cargo")).resolve()
    env["RUSTFLAGS"] = (f"--remap-path-prefix={source}=/wasmtime "
                        f"--remap-path-prefix={cargo_home}=/cargo -C panic=abort")
    command = ["cargo", "+1.98.0", "build", "--locked", "--release", "--target",
               "wasm32-wasip2", "-p", "test-programs", "--manifest-path", source / "Cargo.toml"]
    for name in names:
        command += ["--bin", name]
    run(*command, env=env)
    files = {}
    cases = {}
    binaries = source / "target/wasm32-wasip2/release"
    for name in names:
        output = work / (name + ".wasm")
        run(args.wasm_tools, "strip", "--all", binaries / (name + ".wasm"), "-o", output)
        run(args.wasm_tools, "validate", "--features", "all", output)
        files["components/" + name + ".wasm"] = output.read_bytes()
        files["components/" + name + ".wit"] = run(args.wasm_tools, "component", "wit", output)
        cases[name] = {"sourceSHA256": sha((programs / ("src/bin/" + name + ".rs")).read_bytes()),
                       "binarySHA256": sha(output.read_bytes())}
    config = work / "chain.yml"
    config.write_text("dependencies:\n  chain-http:\n    path: " + str(work / "p3_http_echo.wasm") +
                      "\ninstantiations:\n  root:\n    arguments:\n      local:local/chain-http:\n"
                      "        instance: chain-http\n        export: wasi:http/handler@0.3.0\n")
    composed = work / "p3_http_middleware_with_chain-composed.wasm"
    run(args.wasm_tools, "compose", work / "p3_http_middleware_with_chain.wasm", "-c", config, "-o", composed)
    run(args.wasm_tools, "validate", "--features", "all", composed)
    files["components/" + composed.name] = composed.read_bytes()
    cases["p3_http_middleware_with_chain"]["composedSHA256"] = sha(composed.read_bytes())
    # Retain the original consumers, shared support, WIT, lockfile and reference
    # harness. Rebuilding uses the complete, independently hashed source archive.
    paths = [source / "Cargo.toml", source / "Cargo.lock"]
    for prefix in ["crates/test-programs", "crates/wasi/src/p3/wit", "crates/wasi-http/src/p3/wit"]:
        paths += [path for path in (source / prefix).rglob("*") if path.is_file()]
    paths += list(source.glob("LICENSE*"))
    paths += [source / "crates/wasi-http/tests/all/p3/mod.rs",
              source / "crates/wasi-http/tests/all/http_server.rs",
              source / "crates/wasi/tests/all/p3/mod.rs", source / "tests/all/cli_tests.rs"]
    for path in paths:
        files["source/" + str(path.relative_to(source))] = path.read_bytes()
    tree = run("cargo", "+1.98.0", "tree", "--locked", "--offline", "--target", "wasm32-wasip2",
               "-p", "test-programs", "--manifest-path", source / "Cargo.toml", "--edges", "normal,build",
               "--prefix", "none", "--format", "{p}", text=True)
    packages = sorted(set(re.findall(r"^([\w-]+) v([^\s]+)", tree, re.M)) - {("test-programs", "0.0.0")})
    notices = []
    for name, version_number in packages:
        roots = list((cargo_home / "registry/src").glob("*/" + name + "-" + version_number))
        if len(roots) != 1:
            raise ValueError("missing dependency source: " + name)
        crate = roots[0]
        selected = [p for p in crate.rglob("*") if p.is_file() and
                    (p.name.lower().startswith(("license", "copying", "notice")) or
                     p.parent == crate and p.name.lower().startswith(("cargo.toml", "readme")))]
        if not selected:
            raise ValueError("missing dependency notice: " + name)
        for path in selected:
            files[f"notices/{name}-{version_number}/" + str(path.relative_to(crate))] = path.read_bytes()
        notices.append({"name": name, "version": version_number})
    for name in ["Rust-MIT.txt", "Rust-COPYRIGHT-library.html.gz"]:
        files["notices/rust/" + name] = (ROOT / "testdata/wasi-preview3/guests/notices" / name).read_bytes()
    args.output.mkdir(parents=True, exist_ok=True)
    raw = io.BytesIO()
    with tarfile.open(fileobj=raw, mode="w") as archive_out:
        for name, data in sorted(files.items()):
            entry = tarfile.TarInfo(name)
            entry.mode, entry.size = 0o644, len(data)
            archive_out.addfile(entry, io.BytesIO(data))
    packed = gzip.compress(raw.getvalue(), mtime=0)
    (args.output / "fixtures.tar.gz").write_bytes(packed)
    manifest = {"upstream": "https://github.com/bytecodealliance/wasmtime", "version": "48.0.1",
                "revision": REVISION, "sourceArchiveSHA256": ARCHIVE_SHA,
                "sourceArchiveURL": "https://github.com/bytecodealliance/wasmtime/archive/" + REVISION + ".tar.gz",
                "rust": "1.98.0", "target": "wasm32-wasip2", "bindings": "wit-bindgen 0.60.0",
                "producer": version, "producerExecutableSHA256": sha(args.wasm_tools.read_bytes()),
                "generatorSHA256": sha(Path(__file__).read_bytes()), "inventory": inventory,
                "excluded": EXCLUDED, "cases": cases, "dependencies": notices,
                "archiveSHA256": sha(packed), "sha256": {name: sha(data) for name, data in sorted(files.items())}}
    (args.output / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"Built {len(cases)} unchanged upstream consumers, one composed service, and {len(files)} pinned assets")


if __name__ == "__main__":
    main()
