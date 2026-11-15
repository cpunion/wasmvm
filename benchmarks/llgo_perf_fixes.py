"""Frozen-source compiler diagnostic; paired native processes, no GC overrides."""
import hashlib
import json
import math
import os
from pathlib import Path
import random
import shutil
import statistics
import subprocess
import sys
import tarfile
import time

ROOT = Path.cwd()
OUT = ROOT / "main-regression"
OUT.mkdir()
sys.path.insert(0, str(ROOT / "control/benchmarks"))
import own_build

SYMBOL = "github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop"
BASE = "113daaad50547e5ac8f994d63909e70f2b635e0c"
FIX = "b96415f25a0ec7067ac4e4ab3765ce38ba057d4b"
compilers = {
    "base": (shutil.which("llgo"), os.environ["LLGO_ROOT"], BASE),
    "fix": (str(ROOT / "compiler/llgo-diag"), str(ROOT / "compiler"), FIX),
}
assert subprocess.check_output(["git", "-C", "compiler", "rev-parse", "HEAD"], text=True).strip() == FIX
run = lambda args, **kw: subprocess.check_output(args, text=True, **kw)
result = {"host": run(["lscpu"]), "compiler": {}, "roles": {}, "scope": "Focused frozen-source diagnostic, not portfolio acceptance"}
for label, (cli, sdk, revision) in compilers.items():
    result["compiler"][label] = {"revision": revision, "cliSha256": own_build.digest(Path(cli)), "version": run([cli, "version"], env=dict(os.environ, LLGO_ROOT=sdk))}
fixture = ROOT / "head/benchmarks/native/memory-fixtures/dependent-address.wasm"
for role in ["counter", "dense", "scalar"]:
    folder = ROOT / ("source-" + role)
    folder.mkdir()
    with tarfile.open(ROOT / "downloads" / role / "poll-llgo.tar.gz") as archive:
        archive.extractall(folder, filter="data")
    build = folder / "routing-build"
    manifest = json.loads((build / "poll-llgo.json").read_text())
    source = build / "poll-llgo"
    hashes = manifest["builds"]["poll-llgo"]["sourceHashes"]
    assert own_build.hashes(source, own_build.ENGINE_PATHS, own_build.HARNESS) == hashes
    records = {"source": manifest["candidateRevision"], "builds": {}, "blocks": []}
    for label, (cli, sdk, revision) in compilers.items():
        (source / "internal/bench/phase/provenance_gen.go").write_text(
            "package phase\nfunc init(){sourceRevision=" + json.dumps(manifest["candidateRevision"]) + ";compilerRevision=" + json.dumps(revision) + "}\n")
        binary = OUT / (role + "-" + label)
        command = [cli, "build", "-O3", "-lto=full", "-o", str(binary), "./benchmarks/embedded/wasmvm"]
        before = time.monotonic()
        subprocess.run(command, cwd=source, env=dict(os.environ, LLGO_ROOT=sdk), check=True, timeout=240)
        assert own_build.hashes(source, own_build.ENGINE_PATHS, own_build.HARNESS) == hashes
        assembly = run(["llvm-objdump", "--disassemble-symbols=" + SYMBOL, "--no-show-raw-insn", "--x86-asm-syntax=intel", str(binary)])
        (OUT / (role + "-" + label + ".asm")).write_text(assembly)
        records["builds"][label] = {"sha256": own_build.digest(binary), "buildSeconds": time.monotonic() - before,
            "symbol": [line for line in run(["llvm-nm", "--print-size", str(binary)]).splitlines() if line.endswith(" " + SYMBOL)]}
        shutil.copy2(binary, str(binary) + "-aa")
    for block in range(20):
        order = ["base", "base-aa", "fix", "fix-aa"]
        random.Random(20261116 + block).shuffle(order)
        observations = {}
        for label in order:
            binary = OUT / (role + "-" + label)
            command = [str(binary), "--mode", "api", "--backend", "interpreter", "--wasm", str(fixture),
                "--iterations", "100000", "--seed", "305419896", "--expected", "509818616", "--samples", "8", "--warmups", "2"]
            proc = subprocess.run(command, input="run\n", text=True, capture_output=True, check=True, timeout=30)
            receipt = json.loads(proc.stdout.splitlines()[-1])
            assert receipt["warmResults"] == [509818616] * 8 and receipt["result"] == 509818616
            assert receipt["compilerRevision"] == compilers[label.split("-")[0]][2]
            observations[label] = receipt
        records["blocks"].append({"order": order, "observations": observations})
    records["ratios"] = {}
    for left, right in [("fix", "base"), ("base-aa", "base"), ("fix-aa", "fix")]:
        values = [math.log(statistics.mean(block["observations"][left]["warmCallNs"]) / statistics.mean(block["observations"][right]["warmCallNs"])) for block in records["blocks"]]
        rng = random.Random(20261116)
        boot = sorted(math.exp(statistics.mean(rng.choices(values, k=len(values)))) for _ in range(10000))
        records["ratios"][left + "/" + right] = [math.exp(statistics.mean(values)), boot[250], boot[9749]]
    result["roles"][role] = records
    (OUT / "results.json").write_text(json.dumps(result, indent=2) + "\n")
    print(role, records["builds"], records["ratios"], flush=True)
