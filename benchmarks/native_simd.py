"""Scoped fixed-SIMD diagnostics using existing builds, phases and barriers."""
import hashlib
import importlib.util
import json
import os
import random
import statistics
import sys

import interleaved
import phase_ci


def fixtures(head):
    sys.path.insert(0, str(head / "benchmarks/native"))
    import vector_float_fixtures
    import vector_integer_fixtures
    import runtime_fixtures
    scalar_spec = importlib.util.spec_from_file_location("wasmvm_core_fixture_oracle", head / "benchmarks/embedded/generate.py")
    scalar = importlib.util.module_from_spec(scalar_spec)
    scalar_spec.loader.exec_module(scalar)
    result = []
    groups = [("vector-integer", vector_integer_fixtures, head / "benchmarks/native/vector-integer-fixtures", None),
              ("vector-float", vector_float_fixtures, head / "benchmarks/native/vector-float-fixtures", None),
              ("regression-runtime", runtime_fixtures, head / "benchmarks/native/runtime-fixtures",
               {"indirect-calls", "simd-i32x4", "simd-memory-full", "simd-memory-widen", "simd-memory-lanes"}),
              ("regression-core", scalar, head / "benchmarks/embedded", {"call-control", "integer-sustained", "function-calls", "memory-sustained"})]
    for group, oracle, directory, selected in groups:
        manifest = phase_ci.load_workloads(head) if group == "regression-core" else json.loads((directory / "manifest.json").read_text())
        if manifest["producer"] != "wasm-tools 1.258.0":
            raise ValueError("SIMD fixture producer differs from pin")
        for original in manifest["workloads"]:
            if selected is not None and original["id"] not in selected:
                continue
            f = dict(original, id=group+"/"+original["id"], path=str(directory / original["wasm"]))
            for kind in ("wat", "wasm"):
                if hashlib.sha256((directory / f[kind]).read_bytes()).hexdigest() != f[kind+"Sha256"]:
                    raise ValueError("SIMD fixture hash mismatch: " + f["id"])
            if group in ("vector-integer", "regression-runtime"):
                checksum = lambda seed, count, name=original["id"], o=oracle: o.checksum(name, seed, count)
            elif group == "vector-float":
                checksum = lambda seed, count, f=f, o=oracle: o.checksum(f["family"], f["wide"], seed, count)
            else:
                checksum = lambda seed, count, kind=original["kind"], o=oracle: o.checksum(kind, seed, count)
            result.append((f, checksum))
    return result


def run(head, builds, references, timeout=180):
    cases = fixtures(head)
    runtimes = {name: {"command": [b["binary"], "--mode", "api"], "backend": "native",
                      "engine": "wasmvm", "sourceRevision": b["sourceRevision"],
                      "compiler": "gc" if name.endswith("go") and not name.endswith("llgo") else "llgo",
                      "compilerRevision": b["toolchain"]["revision"], "mode": "api",
                      "goos": b["toolchain"]["goos"], "goarch": b["toolchain"]["goarch"]}
                for name, b in builds.items()}
    runtimes.update({r["id"]: dict(r, backend="compiler") for r in references})
    report = {"status": "running", "acceptance": "not evaluated; diagnostics cannot authorize merging",
              "protocol": "warm-call-interleave-v1", "rounds": 3, "samples": 7, "warmups": 2,
              "iterations": 100_000, "checks": [], "cases": [], "hostLoadBefore": os.getloadavg()}
    rng = random.Random(20261104)

    def command(runtime, f, samples=7, warmups=2):
        r = runtimes[runtime]
        return r["command"] + ["--backend", r["backend"], "--wasm", f["path"],
                "--iterations", str(f["iterations"]), "--seed", str(f["seed"]),
                "--expected", str(f["expectedChecksum"]), "--samples", str(samples), "--warmups", str(warmups)]

    # Preserve partial measurements if a later case fails.
    try:
        for original, checksum in cases:
            control = original.get("kind") == "call-control"
            for seed, count in ((0, 1), (0xffffffff, 2), (7, 19), (0x80000000, 1001)):
                if control:
                    count = 1
                f = dict(original, seed=seed, iterations=count, expectedChecksum=checksum(seed, count))
                for name in runtimes:
                    row = phase_ci.invoke(command(name, f, 1, 0), timeout)
                    if row["record"]["warmResults"] != [f["expectedChecksum"]]:
                        raise ValueError("SIMD oracle mismatch: " + name + "/" + f["id"])
                    report["checks"].append(dict(id=name, fixture=f, **row))
            count = 1 if control else report["iterations"]
            f = dict(original, iterations=count, expectedChecksum=checksum(original["seed"], count))
            report["cases"].append({"fixture": f, "observations": [], "sampleOrders": []})
            print(f["id"], "oracle checks passed", flush=True)
        for block in range(report["rounds"]):
            for case in report["cases"]:
                f = case["fixture"]
                rows, order = interleaved.invoke({name: command(name, f) for name in runtimes}, timeout, rng)
                case["sampleOrders"].append(order)
                for name, row in rows.items():
                    row = dict(row, wasmSha256=f["wasmSha256"])
                    phase_ci.validate_observation(row, runtimes[name], f)
                    case["observations"].append(dict(id=name, round=block, **row))
            print("Fixed SIMD process block", block+1, "complete", flush=True)
        report["status"] = "diagnostic"
    except Exception as error:
        report["status"], report["failure"] = "failed", f"{type(error).__name__}: {error}"
    report["hostLoadAfter"] = os.getloadavg()
    return report


def markdown(report):
    lines = ["## Fixed SIMD native diagnostics", "", "**Diagnostic only; merge acceptance remains open.** "
             "Three interleaved process blocks, seven calls and two warmups. Loop workloads use 100,000 source iterations per call; "
             "call-control uses one invocation without a guest loop. "
             "All startup/compile/instantiate/first-call phases and CPU masks remain in the raw report.", "",
             "| Workload | Unit | Go base | LLGo base | Go head | LLGo head | wazero | Wasmtime | Wasmer |",
             "| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |"]
    if report["status"] != "diagnostic":
        return "\n".join(lines[:2] + ["**Measurement failed:** " + report.get("failure", "incomplete"), ""])
    for case in report["cases"]:
        values = []
        for name in ("base-go", "base-llgo", "head-go", "head-llgo", "head-compiler", "wasmtime-cranelift", "wasmer-cranelift"):
            rows = [r for r in case["observations"] if r["id"] == name]
            values.append(statistics.median(statistics.median(r["record"]["warmCallNs"])/r["record"]["iterations"] for r in rows))
        unit = "ns/call" if case["fixture"].get("kind") == "call-control" else "ns/source iteration"
        lines.append("| " + case["fixture"]["id"] + " | " + unit + " | " + " | ".join(f"{v:.3f}" for v in values) + " |")
    return "\n".join(lines + ["", "Missing A/A calibration, simultaneous intervals or target evidence blocks merging.", ""])
