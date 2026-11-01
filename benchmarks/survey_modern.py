"""Measure GC and uncontended Core Threads with explicit feature configurations.

Uses the Core survey's exact archived sources and records every instrumentation
patch. Missing GC in wazero/Wasmer is reported separately from successful calls.
"""
import argparse
import json
import os
from pathlib import Path
import random
import shutil
import subprocess

import interleaved
import own_build
import phase_ci
import survey

POLICY = "unmetered; bounded cancellation; checked Core 3.0 + Threads"


def expected_gc_rejection(name, error):
    """A crash, timeout or unrelated compiler failure is not missing GC support."""
    message = str(error)
    if not message.startswith("embedded runner exit 1:"):
        return False
    if name.startswith("wazero-"):
        return "compile/validate:" in message and any(text in message for text in (
            "invalid instruction 0xfb", "invalid byte: 0x5f != 0x60"))
    return name == "wasmer-cranelift" and "Validation error:" in message and any(
        text in message for text in ("gc support is not enabled", "not supported without the gc feature"))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--core-output", type=Path, required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parent.parent
    output = args.core_output.resolve() / "modern"
    output.mkdir(exist_ok=True)
    core = json.loads((args.core_output / "build.json").read_text())
    runtimes, builds = {}, {}
    for compiler in ("go", "llgo"):
        old = core["builds"][compiler]
        target = Path(old["binary"]).parent
        path = target / "internal/bench/phase/phase.go"
        original = path.read_text()
        patched = original.replace("config.Backend = args.backend", "config.Backend = args.backend\n config.Features = wasm.SupportedFeatures", 1)
        patched = patched.replace("unmetered; bounded cancellation; checked Core 2.0", POLICY)
        if patched == original:
            raise ValueError("missing feature instrumentation boundary")
        path.write_text(patched)
        binary = output / ("wasmvm-" + compiler)
        command = [*old["command"]]
        command[command.index("-o") + 1] = str(binary)
        try:
            subprocess.run(command, cwd=target, env=dict(os.environ, CGO_ENABLED="1"), check=True, timeout=600)
        finally:
            path.write_text(original)
        if own_build.hashes(target, own_build.ENGINE_PATHS, own_build.HARNESS) != old["sourceHashes"]:
            raise ValueError("modern instrumentation changed engine sources")
        builds[compiler] = dict(command=command, binary=str(binary), binarySha256=survey.digest(binary),
                                instrumentation=patched, features="wasm.SupportedFeatures")
        for backend in ("interpreter", "native"):
            name = "wasmvm-" + compiler + "-" + backend
            runtimes[name] = dict(core["runtimes"][name], command=[str(binary), "--mode", "api"], policy=POLICY)
    target = output / "wazero"
    shutil.copytree(Path(core["builds"]["wazero"]["binary"]).parent, target, dirs_exist_ok=True)
    path = target / "main.go"
    source = path.read_text().replace('"github.com/tetratelabs/wazero/api"',
                                     '"github.com/tetratelabs/wazero/api"\n"github.com/tetratelabs/wazero/experimental"')
    source = source.replace("runtime := wazero.NewRuntimeWithConfig(ctx, config)",
                            "config = config.WithCoreFeatures(api.CoreFeaturesV2 | experimental.CoreFeaturesThreads)\n runtime := wazero.NewRuntimeWithConfig(ctx, config)")
    path.write_text(source)
    binary = target / "wazero-modern"
    subprocess.run(["go", "build", "-trimpath", "-o", str(binary), "."], cwd=target, check=True, timeout=600)
    builds["wazero"] = dict(binary=str(binary), binarySha256=survey.digest(binary), sourceSha256=survey.digest(path),
                             features="CoreFeaturesV2 | CoreFeaturesThreads")
    for backend in ("interpreter", "compiler"):
        runtimes["wazero-" + backend] = dict(command=[str(binary)], backend=backend)
    shutil.copyfile(root / "benchmarks/embedded/c/common.h", output / "common.h")
    for name in ("wasmtime", "wasmer", "wasmedge"):
        source = (root / "benchmarks/embedded/c" / (name + ".c")).read_text()
        if name == "wasmtime":
            source = source.replace("wasmtime_config_strategy_set(config, WASMTIME_STRATEGY_CRANELIFT);",
                                    "wasmtime_config_strategy_set(config, WASMTIME_STRATEGY_CRANELIFT);\n wasmtime_config_wasm_threads_set(config,true);\n wasmtime_config_shared_memory_set(config,true);")
        elif name == "wasmer":
            source = source.replace("wasm_config_set_backend(config, CRANELIFT);",
                                    'wasm_config_set_backend(config, CRANELIFT);\n wasmer_features_t *features=wasmer_features_new();\n if(!wasmer_features_threads(features,true))last_error("enable threads");\n wasm_config_set_features(config,features);')
        else:
            source = source.replace("WasmEdge_ConfigureSetRunMode(config, WasmEdge_RunMode_Interpreter);",
                                    "WasmEdge_ConfigureSetRunMode(config, WasmEdge_RunMode_Interpreter);\n WasmEdge_ConfigureAddProposal(config,WasmEdge_Proposal_Threads);\n WasmEdge_ConfigureAddProposal(config,WasmEdge_Proposal_FunctionReferences);\n WasmEdge_ConfigureAddProposal(config,WasmEdge_Proposal_GC);")
        path, binary = output / (name + ".c"), output / (name + "-modern")
        path.write_text(source)
        command = list(core["builds"][name]["command"])
        command[command.index("-o") - 1] = str(path)
        command[command.index("-o") + 1] = str(binary)
        subprocess.run(command, check=True, timeout=180)
        builds[name] = dict(command=command, binary=str(binary), binarySha256=survey.digest(binary), sourceSha256=survey.digest(path))
        name_id = name + ("-interpreter" if name == "wasmedge" else "-cranelift")
        runtimes[name_id] = dict(core["runtimes"][name_id], command=[str(binary)])
    (output / "build.json").write_text(json.dumps(dict(builds=builds, runtimes=runtimes), indent=2) + "\n")
    manifest = json.loads((root / "benchmarks/modern/manifest.json").read_text())
    fixtures = []
    for f in manifest["workloads"]:
        for suffix in ("wasm", "wat"):
            if survey.digest(root / "benchmarks/modern" / f[suffix]) != f[suffix + "Sha256"]:
                raise ValueError("modern fixture hash mismatch")
        fixtures.append(dict(f, path=str(root / "benchmarks/modern" / f["wasm"]), unit="ns/source iteration"))
    report = dict(status="running", acceptance="not evaluated; descriptive shared-host survey",
                  sourceRevision=core["revision"], hostLoadBefore=os.getloadavg(),
                  runtimes=runtimes, cases=[dict(fixture=f, observations=[], sampleOrders=[]) for f in fixtures],
                  unsupported=[], checks=[])
    rng = random.Random(20261101)
    try:
        for case in report["cases"]:
            f = case["fixture"]
            commands = {name: r["command"] + ["--backend", r["backend"], "--wasm", f["path"],
                        "--iterations", str(f["iterations"]), "--seed", str(f["seed"]),
                        "--expected", str(f["expectedChecksum"]), "--samples", "7", "--warmups", "2"]
                        for name, r in runtimes.items()}
            if f["requiresGC"]:
                for name in ("wazero-interpreter", "wazero-compiler", "wasmer-cranelift"):
                    try:
                        phase_ci.invoke(commands.pop(name), 180)
                    except RuntimeError as error:
                        if not expected_gc_rejection(name, error):
                            raise
                        report["unsupported"].append(dict(runtime=name, workload=f["id"], error=str(error)))
                    else:
                        raise ValueError("unexpected GC support; update the declared comparison matrix")
            for name, command in commands.items():
                alternate = list(command)
                alternate[alternate.index("--iterations") + 1] = "17"
                alternate[alternate.index("--seed") + 1] = "4294967295"
                alternate[alternate.index("--expected") + 1] = str(f["alternateExpectedChecksum"])
                check = phase_ci.invoke(alternate, 180)
                if check["record"]["result"] != f["alternateExpectedChecksum"]:
                    raise ValueError("alternate-seed checksum mismatch")
                report["checks"].append(dict(runtime=name, workload=f["id"], **check))
            for block in range(3):
                rows, order = interleaved.invoke(commands, 180, rng)
                case["sampleOrders"].append(order)
                for name, row in rows.items():
                    row = dict(row, wasmSha256=f["wasmSha256"])
                    phase_ci.validate_observation(row, runtimes[name], f)
                    case["observations"].append(dict(row, runtime=name, block=block))
            print("modern", f["id"], "checked", flush=True)
            (output / "results.json").write_text(json.dumps(report, indent=2) + "\n")
        report["status"] = "complete"
    except BaseException as error:
        report["status"], report["failure"] = "failed", str(error)
        raise
    finally:
        report["hostLoadAfter"] = os.getloadavg()
        (output / "results.json").write_text(json.dumps(report, indent=2) + "\n")
        (output / "summary.json").write_text(json.dumps(survey.summarize(report), indent=2) + "\n")
        (output / "report.md").write_text(survey.markdown(report))
        with (args.core_output / "report.md").open("a") as stream:
            stream.write("\n" + survey.markdown(report))


if __name__ == "__main__":
    main()
