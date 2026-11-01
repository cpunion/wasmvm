#!/usr/bin/env python3
"""Build the complete pinned P2 consumer inventory without changing its sources."""
import argparse
import difflib
import gzip
import io
import json
import os
from pathlib import Path
import subprocess
import tarfile

from build_wasi_upstream_fixtures import ARCHIVE_SHA, REVISION, ROOT, run, sha

EXCLUDED = {
    "p2_tls_sample_application": "Draft wasi:tls is outside the stable WASI profile.",
    "p2_cli_serve_config": "wasi:config/store@0.2.0-rc.1 is outside the declared worlds.",
    "p2_cli_serve_keyvalue": "wasi:keyvalue/store@0.2.0-draft is outside the declared worlds.",
}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-archive", type=Path, required=True)
    parser.add_argument("--work", type=Path, required=True)
    parser.add_argument("--wasm-tools", type=Path, required=True)
    parser.add_argument("--output", type=Path, default=ROOT / "testdata/wasi-preview2/upstream")
    args = parser.parse_args()
    archive = args.source_archive.read_bytes()
    if sha(archive) != ARCHIVE_SHA:
        raise ValueError("changed upstream source archive")
    if not run("rustc", "+1.98.0", "--version", text=True).startswith("rustc 1.98.0 "):
        raise ValueError("changed Rust compiler")
    version = run(args.wasm_tools, "--version", text=True).strip()
    if version != "wasm-tools 1.262.0 (e22db44da 2026-10-13)":
        raise ValueError("changed component producer")
    base = ROOT / "testdata/wasi-preview3/upstream"
    base_manifest = json.loads((base / "manifest.json").read_text())
    if sha((base / "fixtures.tar.gz").read_bytes()) != base_manifest["archiveSHA256"]:
        raise ValueError("changed shared sources and notices")
    work = args.work.resolve()
    work.mkdir(parents=True, exist_ok=True)
    with tarfile.open(fileobj=io.BytesIO(archive)) as source:
        source.extractall(work, filter="data")
    source = work / ("wasmtime-" + REVISION)
    programs = source / "crates/test-programs"
    inventory = sorted(path.stem for path in (programs / "src/bin").glob("p2_*.rs"))
    if len(inventory) != 95 or not set(EXCLUDED) < set(inventory):
        raise ValueError("changed upstream P2 inventory")
    names = [name for name in inventory if name != "p2_tls_sample_application"]
    env = os.environ.copy()
    cargo_home = Path(env.get("CARGO_HOME", Path.home() / ".cargo")).resolve()
    env["RUSTFLAGS"] = (f"--remap-path-prefix={source}=/wasmtime "
                        f"--remap-path-prefix={cargo_home}=/cargo -C panic=abort")
    command = ["cargo", "+1.98.0", "build", "--locked", "--release", "--target",
               "wasm32-wasip2", "-p", "test-programs", "--manifest-path", source / "Cargo.toml"]
    for name in names:
        command += ["--bin", name]
    subprocess.run([str(arg) for arg in command], env=env, check=True)
    files, cases, drafts = {}, {}, {}
    binaries = source / "target/wasm32-wasip2/release"
    for name in names:
        output = work / (name + ".wasm")
        run(args.wasm_tools, "strip", "--all", binaries / (name + ".wasm"), "-o", output)
        run(args.wasm_tools, "validate", "--features", "all", output)
        files["components/" + name + ".wasm"] = output.read_bytes()
        files["components/" + name + ".wit"] = run(args.wasm_tools, "component", "wit", output)
        target = drafts if name in EXCLUDED else cases
        target[name] = {"sourceSHA256": sha((programs / ("src/bin/" + name + ".rs")).read_bytes()),
                        "binarySHA256": sha(output.read_bytes())}
    # Retain a separate stress variant with explicit sample-executor repairs.
    # The unchanged binary above remains the consumer-inventory entry.
    name = "p2_api_proxy_streaming"
    path = programs / ("src/bin/" + name + ".rs")
    original = path.read_text()
    start = original.index("                    let mut new_wakers = Vec::new();")
    end = original.index("                    *WAKERS.lock().unwrap() = new_wakers;", start)
    end += len("                    *WAKERS.lock().unwrap() = new_wakers;")
    replacement = """                    let wakers = mem::take::<Vec<_>>(&mut WAKERS.lock().unwrap());
                    assert!(!wakers.is_empty());
                    let pollables = wakers.iter().map(|(pollable, _)| pollable).collect::<Vec<_>>();
                    io::poll::poll(&pollables);
                    drop(pollables);
                    // Repoll all futures after retiring every old subscription.
                    for (_pollable, waker) in wakers {
                        waker.wake();
                    }"""
    repaired = original[:start] + replacement + original[end:]
    patch = "".join(difflib.unified_diff(original.splitlines(True), repaired.splitlines(True),
                                         fromfile="a/"+str(path.relative_to(source)),
                                         tofile="b/"+str(path.relative_to(source))))
    files["patches/p2-streaming-executor.patch"] = patch.encode()
    try:
        path.write_text(repaired)
        stress_command = command[:command.index("--bin")] + ["--bin", name]
        subprocess.run([str(arg) for arg in stress_command], env=env, check=True)
        output = work / (name + "-executor-lifetimes.wasm")
        run(args.wasm_tools, "strip", "--all", binaries / (name + ".wasm"), "-o", output)
        run(args.wasm_tools, "validate", "--features", "all", output)
        files["components/" + output.name] = output.read_bytes()
        files["components/" + output.stem + ".wit"] = run(args.wasm_tools, "component", "wit", output)
        variant = {"original": name, "patchSHA256": sha(patch.encode()),
                   "repairedSourceSHA256": sha(repaired.encode()), "binarySHA256": sha(output.read_bytes())}
    finally:
        path.write_text(original)
    # The existing P3 archive already preserves all test-program sources,
    # Cargo.lock and dependency/Rust notices. Add only missing P2 contracts.
    for prefix in ["crates/wasi/src/p2/wit", "crates/wasi-http/src/p2/wit",
                   "crates/wasi/tests/all/p2", "crates/wasi-http/tests/all/p2"]:
        for path in (source / prefix).rglob("*"):
            if path.is_file():
                files["source/" + str(path.relative_to(source))] = path.read_bytes()
    path = source / "crates/wasi-http/tests/all/p2.rs"
    files["source/" + str(path.relative_to(source))] = path.read_bytes()
    raw = io.BytesIO()
    with tarfile.open(fileobj=raw, mode="w") as out:
        for name, data in sorted(files.items()):
            entry = tarfile.TarInfo(name)
            entry.mode, entry.size = 0o644, len(data)
            out.addfile(entry, io.BytesIO(data))
    packed = gzip.compress(raw.getvalue(), mtime=0)
    manifest = {"upstream": base_manifest["upstream"], "version": "48.0.1", "revision": REVISION,
                "sourceArchiveSHA256": ARCHIVE_SHA, "sourceArchiveURL": base_manifest["sourceArchiveURL"],
                "rust": "1.98.0", "target": "wasm32-wasip2", "bindings": "wit-bindgen 0.60.0",
                "producer": version, "producerExecutableSHA256": sha(args.wasm_tools.read_bytes()),
                "generatorSHA256": sha(Path(__file__).read_bytes()),
                "sharedArchiveSHA256": base_manifest["archiveSHA256"],
                "sharedGeneratorSHA256": base_manifest["generatorSHA256"],
                "inventory": inventory, "excluded": EXCLUDED, "cases": cases,
                "retainedDraftCases": drafts, "variants": {"p2_api_proxy_streaming-executor-lifetimes": variant},
                "archiveSHA256": sha(packed),
                "sha256": {name: sha(data) for name, data in sorted(files.items())}}
    args.output.mkdir(parents=True, exist_ok=True)
    (args.output / "fixtures.tar.gz").write_bytes(packed)
    (args.output / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"Built {len(cases)} stable and {len(drafts)} retained draft P2 consumers; {len(files)} assets")


if __name__ == "__main__":
    main()
