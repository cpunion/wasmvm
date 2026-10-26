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
MAX_ROUNDS = 60
BASELINE_ROUNDS = 3
# Predeclared practical equivalence and stricter A/A resolution requirements.
PARITY = 0.005
CALIBRATION = 0.0025
REFERENCES = ("head-compiler", "wasmtime-cranelift", "wasmer-cranelift")


def compare(rows, rounds=ROUNDS):
    """Compare independent process blocks; require calibrated equivalence."""
    indexed = {(r["id"], r["round"]): statistics.median(r["record"]["warmCallNs"])
               for r in rows}
    names = {f"head-{c}{suffix}" for c in ("go", "llgo") for suffix in ("", "-control")} | set(REFERENCES)
    expected = {(name, n) for name in names for n in range(rounds)}
    expected |= {(f"base-{c}", n) for c in ("go", "llgo") for n in range(BASELINE_ROUNDS)}
    if set(indexed) != expected or len(rows) != len(indexed):
        raise ValueError("incomplete or duplicate native comparison matrix")
    if any(v <= 0 for v in indexed.values()):
        raise ValueError("unresolved native samples")
    result = []
    for compiler in ("go", "llgo"):
        for ref in (f"head-{compiler}-control", *REFERENCES):
            logs = [math.log(indexed[(f"head-{compiler}", n)] / indexed[(ref, n)]) for n in range(rounds)]
            rng = random.Random(20261026)
            draws = sorted(math.exp(statistics.mean(rng.choices(logs, k=rounds))) for _ in range(8000))
            # Bonferroni adjustment across eight comparisons and two planned looks.
            lower, upper = draws[12], draws[-13]
            calibration = ref.endswith("-control")
            if calibration:
                status = "calibrated" if 1-CALIBRATION <= lower <= upper <= 1+CALIBRATION else "inconclusive"
            else:
                status = "faster" if upper < 1-PARITY else "parity" if 1-PARITY <= lower <= upper <= 1+PARITY else "slower" if lower > 1+PARITY else "inconclusive"
            result.append({"compiler": compiler, "reference": ref,
                           "ratio": math.exp(statistics.mean(logs)), "lower": lower, "upper": upper, "status": status})
    return result


def run(head, builds, runtimes, fixture, invoke, timeout):
    """Build only the small native probe in already isolated source trees."""
    if fixture["id"] != "integer-sustained" or fixture["iterations"] != 5000000:
        raise ValueError("native diagnostic workload differs from its declared kernel")
    report = {"scope": "integer compiler architecture; public native VM remains unavailable",
              "rounds": 0, "plannedLooks": [ROUNDS, MAX_ROUNDS], "looks": [], "baselineRounds": BASELINE_ROUNDS, "parityMargin": PARITY, "calibrationMargin": CALIBRATION, "fixture": fixture, "builds": {}, "observations": [], "status": "running"}
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
        commands[name] = [str(binary), "--tier", tier, "--budget", "64", "--wasm", str(head / "benchmarks/embedded" / fixture["wasm"]), "--seed", str(fixture["seed"]), "--iterations", str(fixture["iterations"]), "--expected", str(fixture["expectedChecksum"])]
    for ref in REFERENCES:
        runtime = next(r for r in runtimes if r["id"] == ref)
        commands[ref] = runtime["command"] + ["--wasm", str(head / "benchmarks/embedded" / fixture["wasm"]),
            "--backend", "compiler", "--iterations", str(fixture["iterations"]), "--seed", str(fixture["seed"]),
            "--expected", str(fixture["expectedChecksum"]), "--samples", "7", "--warmups", "2"]
    for compiler in ("go", "llgo"):
        commands[f"head-{compiler}-control"] = commands[f"head-{compiler}"]
    # Linux affinity removes cross-core placement as a compiler-dependent cost.
    if hasattr(os, "sched_getaffinity"):
        cpu = min(os.sched_getaffinity(0))
        report["affinityCPU"] = cpu
        commands = {name: ["taskset", "-c", str(cpu), *command] for name, command in commands.items()}
    rng = random.Random(20261026)
    for n in range(MAX_ROUNDS):
        order = [name for name in commands if not name.startswith("base-") or n < BASELINE_ROUNDS]; rng.shuffle(order)
        for name in order:
            observation = invoke(commands[name], timeout)
            record = observation["record"]
            if len(record["warmCallNs"]) != 7 or any(type(x) is not int or x <= 0 for x in record["warmCallNs"]):
                raise ValueError("invalid native warm samples")
            if record["iterations"] != fixture["iterations"] or record["seed"] != fixture["seed"]:
                raise ValueError("workload parameters differ")
            if name not in REFERENCES:
                if record["wasmSHA256"] != fixture["wasmSha256"] or record["expected"] != fixture["expectedChecksum"]:
                    raise ValueError("native fixture mismatch")
                if record["tier"] != report["builds"][name.removesuffix("-control")]["tier"]:
                    raise ValueError("native policy mismatch")
            elif record["result"] != fixture["expectedChecksum"] or record["warmResults"] != [fixture["expectedChecksum"]] * 7:
                raise ValueError("reference checksum mismatch")
            report["observations"].append({**observation, "id": name, "round": n})
        if n+1 in (ROUNDS, MAX_ROUNDS):
            report["rounds"] = n+1
            report["comparisons"] = compare(report["observations"], n+1)
            report["looks"].append({"rounds": n+1, "comparisons": report["comparisons"]})
            statuses = {r["status"] for r in report["comparisons"]}
            if "slower" in statuses or "inconclusive" not in statuses: break
    report["status"] = "pass" if all(r["status"] in ("faster", "parity", "calibrated") for r in report["comparisons"]) else "blocked"
    return report


def markdown(report):
    lines = ["### Native integer feature gate", "",
             "Compiler-only integer/register/polling comparison; **not public native API or full-profile acceptance**. "
             "Fuel is disabled. Own code retains bounded cancellation polls; references use their pinned default interruption settings. "
             f"{report['rounds']} interleaved sustained-work blocks plus three baseline blocks; seven warm calls per process. "
             "The planned looks are 20 and, only if inconclusive, 60 blocks; intervals adjust for both looks. "
             "Parity requires the entire simultaneous 95% interval within ±0.5%; A/A controls must resolve ±0.25%. Shared-runner evidence is provisional. "
             "The hard gate covers all three competitors for each host compiler; own baseline ratios are reported separately.", "",
             "| Runtime | Warm ns/iteration | Preparation ms | First call ms |", "| --- | ---: | ---: | ---: |"]
    for name in sorted({r["id"] for r in report["observations"]}):
        records = [r["record"] for r in report["observations"] if r["id"] == name]
        warm = statistics.median(statistics.median(r["warmCallNs"]) for r in records) / report["fixture"]["iterations"]
        prep = statistics.median(r["phasesNs"]["compileValidate"] if "phasesNs" in r else r["frontendNs"] + r["backendNs"] for r in records) / 1e6
        first = statistics.median(r["phasesNs"]["firstCall"] if "phasesNs" in r else r["firstCallNs"] for r in records) / 1e6
        lines.append(f"| {name} | {warm:.3f} | {prep:.3f} | {first:.3f} |")
    for compiler in ("go", "llgo"):
        medians = {ref: statistics.median(statistics.median(r["record"]["warmCallNs"]) for r in report["observations"] if r["id"] == f"{ref}-{compiler}") for ref in ("base", "head")}
        lines += [f"Own {compiler} head/base median ratio: {medians['head']/medians['base']:.4f} (three baseline processes; descriptive)."]
    lines += ["", "| Host | Reference | Warm ratio | Simultaneous 95% interval | Gate |", "| --- | --- | ---: | --- | --- |"]
    for r in report["comparisons"]:
        lines.append(f"| {r['compiler']} | {r['reference']} | {r['ratio']:.3f} | {r['lower']:.3f}–{r['upper']:.3f} | {r['status']} |")
    lines += ["", f"Feature merge gate: **{report['status']}**. Ratios above 1 take longer; practical parity is limited to the declared 0.5% equivalence band. Missing/inconclusive evidence cannot authorize merging.", ""]
    return "\n".join(lines)
