"""Narrow compiler-feature comparison, reusing the phase job's pinned builds.

This is the integer/register/polling gate, not full native backend acceptance.
New feature families must extend the declared workloads before implementation.
"""
import math
import os
from pathlib import Path
import random
import statistics
import subprocess

import own_build

ROUNDS = 20
REFERENCES = ("head-compiler", "wasmtime-cranelift", "wasmer-cranelift")


def compare(rows):
    """Bootstrap paired process medians; calls inside a process are not units."""
    indexed = {(r["id"], r["round"]): statistics.median(r["record"]["warmCallNs"])
               for r in rows}
    expected = {f"{ref}-{c}" for ref in ("base", "head") for c in ("go", "llgo")} | set(REFERENCES)
    if set(indexed) != {(name, n) for name in expected for n in range(ROUNDS)}:
        raise ValueError("incomplete or duplicate native comparison matrix")
    if len(rows) != len(indexed) or any(v <= 0 for v in indexed.values()):
        raise ValueError("duplicate or unresolved native samples")
    result = []
    for compiler in ("go", "llgo"):
        for ref in (*REFERENCES, f"base-{compiler}"):
            ratios = [indexed[(f"head-{compiler}", n)] / indexed[(ref, n)] for n in range(ROUNDS)]
            logs = [math.log(x) for x in ratios]
            rng = random.Random(20261026)
            draws = sorted(math.exp(statistics.mean(rng.choices(logs, k=ROUNDS))) for _ in range(4000))
            # Bonferroni adjustment across both compilers and four comparisons.
            lower, upper = draws[12], draws[-13]
            result.append({"compiler": compiler, "reference": ref,
                           "ratio": math.exp(statistics.mean(logs)), "lower": lower, "upper": upper,
                           "status": "pass" if upper <= 1 else "slower" if lower > 1 else "inconclusive"})
    return result


def run(head, builds, runtimes, fixture, invoke, timeout):
    """Build only the small native probe in already isolated source trees."""
    if fixture["id"] != "integer-loop" or fixture["iterations"] != 250000 or fixture["expectedChecksum"] != 956690092:
        raise ValueError("native diagnostic workload differs from its declared kernel")
    report = {"scope": "integer compiler architecture; public native VM remains unavailable",
              "rounds": ROUNDS, "fixture": fixture, "builds": {}, "observations": [], "status": "running"}
    commands = {}
    source = (head / "benchmarks/native/main.go").read_text()
    for name in ("base-go", "base-llgo", "head-go", "head-llgo"):
        build = builds[name]
        target = Path(build["binary"]).parent
        before = own_build.hashes(target, own_build.ENGINE_PATHS, own_build.HARNESS)
        compiler_source = (target / "internal/compiler/codegen/function.go").read_text()
        optimized = "func CompileUnmeteredFunction(" in compiler_source
        if name.startswith("head-") and not optimized:
            raise ValueError("head lacks required unmetered compiler")
        tier = "unmetered" if optimized else "sliced"
        probe = target / "benchmarks/native/main.go"
        probe.parent.mkdir(parents=True, exist_ok=True)
        # Bootstrap only the measurement adapter, never any baseline engine file.
        probe.write_text(source if optimized else source.replace("codegen.CompileUnmeteredFunction", "codegen.CompileFunction"))
        binary = target / "wasmvm-native-phase"
        environment = dict(os.environ, GOWORK="off", GOFLAGS="", GOTOOLCHAIN="local", CGO_ENABLED="1")
        command = [build["toolchain"]["command"], "build", "-o", str(binary), "./benchmarks/native"]
        subprocess.run(command, cwd=target, env=environment, check=True, timeout=600)
        if own_build.hashes(target, own_build.ENGINE_PATHS, own_build.HARNESS) != before:
            raise ValueError("native probe build changed archived engine sources")
        report["builds"][name] = {"sourceRevision": build["sourceRevision"], "sourceDigest": build["sourceDigest"],
                                  "toolchain": build["toolchain"], "binarySha256": own_build.digest(binary),
                                  "probeSha256": own_build.digest(probe), "tier": tier, "command": command}
        commands[name] = [str(binary), "--tier", tier, "--budget", "64", "--wasm", str(head / "benchmarks/embedded" / fixture["wasm"])]
    for ref in REFERENCES:
        runtime = next(r for r in runtimes if r["id"] == ref)
        commands[ref] = runtime["command"] + ["--wasm", str(head / "benchmarks/embedded" / fixture["wasm"]),
            "--backend", "compiler", "--iterations", str(fixture["iterations"]), "--seed", str(fixture["seed"]),
            "--expected", str(fixture["expectedChecksum"]), "--samples", "7", "--warmups", "2"]
    rng = random.Random(20261026)
    for n in range(ROUNDS):
        order = list(commands); rng.shuffle(order)
        for name in order:
            observation = invoke(commands[name], timeout)
            record = observation["record"]
            if len(record["warmCallNs"]) != 7 or any(type(x) is not int or x <= 0 for x in record["warmCallNs"]):
                raise ValueError("invalid native warm samples")
            if name not in REFERENCES:
                if record["wasmSHA256"] != fixture["wasmSha256"] or record["expected"] != fixture["expectedChecksum"]:
                    raise ValueError("native fixture mismatch")
                if record["tier"] != report["builds"][name]["tier"]:
                    raise ValueError("native policy mismatch")
            elif record["result"] != fixture["expectedChecksum"] or record["warmResults"] != [fixture["expectedChecksum"]] * 7:
                raise ValueError("reference checksum mismatch")
            report["observations"].append({**observation, "id": name, "round": n})
    report["comparisons"] = compare(report["observations"])
    report["status"] = "pass" if all(r["status"] == "pass" for r in report["comparisons"] if r["reference"] in REFERENCES) else "blocked"
    return report


def markdown(report):
    lines = ["### Native integer feature gate", "",
             "Compiler-only integer/register/polling comparison; **not public native API or full-profile acceptance**. "
             "Fuel is disabled. Own code retains bounded cancellation polls; references use their pinned default interruption settings. "
             "Twenty interleaved process blocks; seven warm calls per process. Shared-runner evidence is provisional. "
             "The hard gate covers all three competitors for each host compiler; own baseline ratios are reported separately.", "",
             "| Runtime | Warm ns/iteration | Preparation ms | First call ms |", "| --- | ---: | ---: | ---: |"]
    for name in sorted({r["id"] for r in report["observations"]}):
        records = [r["record"] for r in report["observations"] if r["id"] == name]
        warm = statistics.median(statistics.median(r["warmCallNs"]) for r in records) / report["fixture"]["iterations"]
        prep = statistics.median(r["phasesNs"]["compileValidate"] if "phasesNs" in r else r["frontendNs"] + r["backendNs"] for r in records) / 1e6
        first = statistics.median(r["phasesNs"]["firstCall"] if "phasesNs" in r else r["firstCallNs"] for r in records) / 1e6
        lines.append(f"| {name} | {warm:.3f} | {prep:.3f} | {first:.3f} |")
    lines += ["", "| Host | Reference | Warm ratio | Simultaneous 95% interval | Gate |", "| --- | --- | ---: | --- | --- |"]
    for r in report["comparisons"]:
        lines.append(f"| {r['compiler']} | {r['reference']} | {r['ratio']:.3f} | {r['lower']:.3f}–{r['upper']:.3f} | {r['status']} |")
    lines += ["", f"Feature merge gate: **{report['status']}**. Ratios above 1 are slower. Missing/inconclusive evidence cannot authorize merging.", ""]
    return "\n".join(lines)
