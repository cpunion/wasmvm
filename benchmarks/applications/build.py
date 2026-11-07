#!/usr/bin/env python3
"""Build pinned compiled compute guests; validation never depends on a VM timer."""
import argparse
import hashlib
import json
import re
from pathlib import Path
import subprocess
import time

from oracle import checksum

HERE = Path(__file__).resolve().parent

def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ["clang", "rustc", "wasm-tools", "linker", "evidence"]:
        parser.add_argument("--"+name, type=Path, required=True)
    parser.add_argument("--only", choices=["clang", "rust"], help="Rebuild one producer, preserving verified existing rows")
    args = parser.parse_args()
    args.evidence.mkdir(parents=True, exist_ok=False)
    fixtures = HERE / "fixtures"
    fixtures.mkdir(exist_ok=True)
    versions = {name: subprocess.check_output([str(path), "--version"], text=True).strip()
                for name, path in [("clang", args.clang), ("rust", args.rustc), ("validator", args.wasm_tools)]}
    if not re.search(r"\bclang version 22\.1\.8\b", versions["clang"].splitlines()[0]) or not versions["rust"].startswith("rustc 1.98.0 (88d9e12ae") or not versions["validator"].startswith("wasm-tools 1.262.0"):
        raise ValueError("producer/validator differs from the declared pins")
    commands = {
        "clang": [str(args.clang), "--no-default-config", "--target=wasm32-unknown-unknown", "-O3", "-ffreestanding", "-nostdlib", "-mbulk-memory", "-fuse-ld="+str(args.linker), "-Wl,--no-entry", "-Wl,--export=run", "-Wl,--export-memory", "-Wl,-z,stack-size=65536", "-Wl,--initial-memory=131072", "-Wl,--max-memory=131072", "sha256.c", "-o", "fixtures/sha256-clang.wasm"],
        "rust": [str(args.rustc), "--edition=2021", "--crate-type=cdylib", "-C", "linker="+str(args.linker), "--target=wasm32-unknown-unknown", "-C", "opt-level=3", "-C", "panic=abort", "-C", "lto=fat", "-C", "codegen-units=1", "-C", "link-arg=--strip-debug", "--remap-path-prefix="+str(HERE)+"=/wasmvm/applications", "sha256.rs", "-o", "fixtures/sha256-rust.wasm"],
    }
    metadata = {"schemaVersion":1,"profile":"Core 2.0","algorithm":"SHA-256 over LCG-generated 64-byte blocks, including standard padding","checksum":"First 32 digest bits, big-endian","api":{"export":"run","parameters":["i32","i32"],"results":["i32"]},"producers":{},"workloads":[]}
    if args.only:
        metadata = json.loads((HERE/"manifest.json").read_text())
        for row in metadata["workloads"]:
            if row["producer"] != args.only:
                prior = metadata["producers"][row["producer"]]
                if digest(HERE/row["wasm"]) != row["wasmSha256"] or digest(HERE/prior["source"]) != prior["sourceSha256"]:
                    raise ValueError("cannot reuse changed guest/source bytes")
    linker_version = subprocess.check_output([str(args.linker), "--version"], text=True).strip()
    if not re.search(r"\bLLD 23\.1\.4\b", linker_version): raise ValueError("linker differs from pin")
    for name, path in [("clang",args.clang),("rust",args.rustc)]:
        if args.only and args.only != name: continue
        command=commands[name];start=time.monotonic()
        with (args.evidence/(name+".log")).open("x") as log:
            result=subprocess.run(command,cwd=HERE,stdout=log,stderr=subprocess.STDOUT)
        record=dict(command=command,exit=result.returncode,elapsedSeconds=time.monotonic()-start,version=versions[name],compilerBinarySha256=digest(path))
        (args.evidence/(name+".json")).write_text(json.dumps(record,indent=2)+"\n")
        if result.returncode:raise RuntimeError("guest compiler failed: "+name)
        wasm=fixtures/("sha256-"+name+".wasm")
        wasm.chmod(0o644)
        validate=[str(args.wasm_tools),"validate","--features=-all,wasm2",str(wasm)]
        with (args.evidence/(name+"-validation.log")).open("x") as log:
            valid=subprocess.run(validate,stdout=log,stderr=subprocess.STDOUT)
        record["validation"]={"command":validate,"exit":valid.returncode,"version":versions["validator"],"binarySha256":digest(args.wasm_tools)}
        if valid.returncode:raise RuntimeError("Core2 validator rejected "+name)
        source="sha256."+("c" if name=="clang" else "rs")
        record.update(source=source,sourceSha256=digest(HERE/source))
        record["linker"] = {"version":linker_version,"binarySha256":digest(args.linker)}
        metadata["producers"][name]=record
        metadata["workloads"]=[row for row in metadata["workloads"] if row["producer"] != name]
        metadata["workloads"].append(dict(id="applications/sha256-"+name,producer=name,family="application",wasm=str(wasm.relative_to(HERE)),wasmSha256=digest(wasm),bytes=wasm.stat().st_size,seed=305419896,iterations=5000,expectedChecksum=checksum(305419896,5000),unit="ns/message block (64 bytes)",maxIterations=65536))
    if len(metadata["workloads"]) != 2: raise ValueError("both producers are required")
    metadata["complete"] = True
    sysroot=Path(subprocess.check_output([str(args.rustc),"--print","sysroot"],text=True).strip())
    metadata["producers"]["rust"]["targetLibraries"]={p.name:digest(p) for p in sorted((sysroot/"lib/rustlib/wasm32-unknown-unknown/lib").glob("*.rlib"))}
    if not metadata["producers"]["rust"]["targetLibraries"]:raise ValueError("missing target library identities")
    (HERE/"manifest.json").write_text(json.dumps(metadata,indent=2)+"\n")
    print("Built and validated",len(metadata["workloads"]),"compiled application guests")

if __name__ == "__main__":
    main()
