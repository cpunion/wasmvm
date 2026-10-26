#!/usr/bin/env python3
"""Measure separate-process CLI latency; this is not steady-state throughput."""

import argparse
from datetime import datetime, timezone
import hashlib
import json
import math
import os
from pathlib import Path
import platform
import random
import shutil
import statistics
import subprocess
import time


ROOT = Path(__file__).resolve().parent
PROJECT = ROOT.parent
WAZERO_VERSION = "v1.12.0"


def capture(command):
    result = subprocess.run(command, capture_output=True, text=True, timeout=60, check=True)
    return (result.stdout + result.stderr).strip()


def summary(values):
    if not values or any(value < 0 or not math.isfinite(value) for value in values):
        raise ValueError("samples must be nonempty, finite, and nonnegative")
    ordered = sorted(values)
    return {
        "samples": len(values), "minMs": min(values), "medianMs": statistics.median(values),
        "p95Ms": ordered[math.ceil(len(ordered) * 0.95) - 1],
    }


def load_workloads():
    manifest = json.loads((ROOT / "fixtures" / "manifest.json").read_text())
    if manifest.get("schemaVersion") != 1 or not manifest.get("workloads"):
        raise ValueError("invalid fixture manifest")
    for workload in manifest["workloads"]:
        for kind in ("wat", "wasm"):
            path = (ROOT / workload[kind]).resolve()
            if not path.is_relative_to(ROOT / "fixtures"):
                raise ValueError("fixture path escapes fixtures directory")
            if hashlib.sha256(path.read_bytes()).hexdigest() != workload[kind + "Sha256"]:
                raise ValueError(f"fixture hash mismatch: {workload['id']} {kind}")
        if not workload["expectedStdout"].endswith("\n"):
            raise ValueError("expected stdout must include its trailing newline")
    return manifest


def runtime_commands(binary):
    return [
        {"id": "wazero-interpreter", "backend": "interpreter", "project": "wazero", "command": [str(binary), "run", "-interpreter", "-cachedir="], "version": capture([str(binary), "version"])},
        {"id": "wazero-compiler", "backend": "wazevo native compiler", "project": "wazero", "command": [str(binary), "run", "-cachedir="], "version": capture([str(binary), "version"])},
        {"id": "wasmtime-cranelift", "backend": "Cranelift, opt-level=2", "project": "Wasmtime", "command": [shutil.which("wasmtime"), "run", "-C", "cache=n", "-C", "compiler=cranelift", "-O", "opt-level=2"], "version": capture(["wasmtime", "--version"])},
        {"id": "wasmer-cranelift", "backend": "Cranelift", "project": "Wasmer", "command": [shutil.which("wasmer"), "run", "--cranelift", "--disable-cache", "--offline", "--no-tty"], "version": capture(["wasmer", "--version"])},
        {"id": "wasmedge-interpreter", "backend": "interpreter (explicit)", "project": "WasmEdge", "command": [shutil.which("wasmedge"), "--run-mode", "interpreter"], "version": capture(["wasmedge", "--version"])},
    ]


def measure(runtime, workload, warmup, round_number, timeout):
    command = runtime["command"] + [str(ROOT / workload["wasm"])]
    start = time.perf_counter_ns()
    result = subprocess.run(command, capture_output=True, timeout=timeout)
    elapsed = (time.perf_counter_ns() - start) / 1_000_000
    expected = workload["expectedStdout"].encode()
    if result.returncode != 0 or result.stdout != expected:
        raise RuntimeError(
            f"{runtime['id']} / {workload['id']} failed correctness: exit={result.returncode}, "
            f"stdout={result.stdout!r}, expected={expected!r}, stderr={result.stderr!r}"
        )
    return {
        "runtime": runtime["id"], "workload": workload["id"], "warmup": warmup,
        "round": round_number, "elapsedMs": elapsed, "exitCode": result.returncode,
        "stdout": result.stdout.decode(), "stderr": result.stderr.decode(errors="replace"),
        "command": command,
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--samples", type=int, default=7)
    parser.add_argument("--warmups", type=int, default=2)
    parser.add_argument("--seed", type=int, default=20261026)
    parser.add_argument("--timeout", type=float, default=30)
    parser.add_argument("--environment-note", default="Uncontrolled local development host")
    parser.add_argument("--output", type=Path, default=ROOT / "results" / "latest.json")
    args = parser.parse_args()
    if args.samples < 7 or args.warmups < 2 or args.timeout <= 0:
        parser.error("use at least seven measured samples, two warmups, and a positive timeout")
    for tool in ("go", "wasmtime", "wasmer", "wasmedge"):
        if shutil.which(tool) is None:
            parser.error(f"required local runtime/tool is absent: {tool}")

    manifest = load_workloads()
    binary_dir = PROJECT / "artifacts" / "benchmark-tools"
    binary_dir.mkdir(parents=True, exist_ok=True)
    environment = dict(os.environ, GOBIN=str(binary_dir), GOWORK="off")
    subprocess.run(["go", "install", f"github.com/tetratelabs/wazero/cmd/wazero@{WAZERO_VERSION}"], env=environment, check=True)
    binary = binary_dir / "wazero"
    runtimes = runtime_commands(binary)
    report = {
        "schemaVersion": 1, "startedAt": datetime.now(timezone.utc).isoformat(),
        "metric": "separate-process CLI wall time including startup, validation/compilation, instantiation, guest execution, and shutdown",
        "cachePolicy": "wazero no compilation-cache directory; Wasmtime cache=n; Wasmer --disable-cache; WasmEdge raw-Wasm interpreter; OS filesystem/executable pages are warmed, not flushed",
        "warmupsPerCase": args.warmups, "samplesPerCase": args.samples, "seed": args.seed,
        "environmentNote": args.environment_note,
        "p95Method": "nearest rank; with seven samples p95 is the sample maximum",
        "host": {"platform": platform.platform(), "machine": platform.machine(), "logicalCPUs": os.cpu_count(), "go": capture(["go", "version"]), "python": platform.python_version()},
        "wazeroBuild": capture(["go", "version", "-m", str(binary)]),
        "runtimes": runtimes, "fixtures": manifest, "observations": [], "summary": [],
        "limitations": ["Single local machine, no CPU affinity or isolation", "Small synthetic workloads, not an application benchmark suite", "No peak memory or energy measurement", "CLI timing is not warmed in-process guest throughput", "All modes use their actual backend and CLI; runtime startup/host overhead differs"],
    }
    if platform.system() == "Darwin":
        report["host"].update({"cpu": capture(["sysctl", "-n", "machdep.cpu.brand_string"]), "memoryBytes": int(capture(["sysctl", "-n", "hw.memsize"])), "osVersion": capture(["sw_vers"])})
    if hasattr(os, "getloadavg"):
        report["host"]["loadAverageBefore"] = list(os.getloadavg())
    args.output.parent.mkdir(parents=True, exist_ok=True)
    randomizer = random.Random(args.seed)
    cases = [(runtime, workload) for runtime in runtimes for workload in manifest["workloads"]]
    try:
        for warmup, rounds in ((True, args.warmups), (False, args.samples)):
            for round_number in range(rounds):
                order = list(cases)
                randomizer.shuffle(order)
                for runtime, workload in order:
                    observation = measure(runtime, workload, warmup, round_number, args.timeout)
                    report["observations"].append(observation)
                print(f"{'warmup' if warmup else 'measured'} round {round_number + 1}/{rounds} complete", flush=True)
        for runtime, workload in cases:
            values = [item["elapsedMs"] for item in report["observations"] if not item["warmup"] and item["runtime"] == runtime["id"] and item["workload"] == workload["id"]]
            report["summary"].append({"runtime": runtime["id"], "workload": workload["id"], **summary(values)})
        report["status"] = "complete"
    except Exception as error:
        report["status"] = "failed"
        report["failure"] = str(error)
        raise
    finally:
        report["finishedAt"] = datetime.now(timezone.utc).isoformat()
        if hasattr(os, "getloadavg"):
            report["host"]["loadAverageAfter"] = list(os.getloadavg())
        args.output.write_text(json.dumps(report, indent=2) + "\n")
    for item in report["summary"]:
        print(f"{item['runtime']:23} {item['workload']:18} median={item['medianMs']:.3f} ms min={item['minMs']:.3f} p95={item['p95Ms']:.3f}")
    print(f"Raw results: {args.output}")


if __name__ == "__main__":
    main()
