#!/usr/bin/env python3
"""Compare pinned candidate dependencies from two revisions on one CI runner."""

import argparse
from datetime import datetime, timezone
import hashlib
import json
import math
import os
from pathlib import Path
import platform
import random
import re
import shutil
import subprocess
import tempfile
import time

from run import summary


MODULE = "github.com/tetratelabs/wazero"
WARMUPS = 2
SAMPLES = 7
CONFIGURATIONS = (
    "base-interpreter", "head-interpreter", "base-compiler", "head-compiler",
    "wasmtime-cranelift", "wasmer-cranelift", "wasmedge-interpreter",
)
WORKLOADS = (
    "startup", "integer-loop", "integer-sustained", "function-calls",
    "memory-load-store", "memory-sustained",
)
COMPETITOR_VERSIONS = {"wasmtime": "48.0.1", "wasmer": "7.3.0", "wasmedge": "0.17.1"}


def capture(command, *, cwd=None, env=None, timeout=60):
    result = subprocess.run(command, cwd=cwd, env=env, capture_output=True,
                            text=True, check=True, timeout=timeout)
    return (result.stdout + result.stderr).strip()


def sha256(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify_checkout(directory, revision):
    if not re.fullmatch(r"[0-9a-f]{40}", revision):
        raise ValueError("checkout revisions must be full 40-character commit IDs")
    actual = capture(["git", "-C", str(directory), "rev-parse", "HEAD"])
    if actual != revision:
        raise ValueError(f"checkout {directory} is {actual}, expected {revision}")
    # The report attributes these input files to the commit, so reject local
    # tracked modifications instead of silently benchmarking different inputs.
    subprocess.run(["git", "-C", str(directory), "diff", "--quiet", revision, "--",
                    "experiments/wazero/go.mod", "experiments/wazero/go.sum", "benchmarks"],
                   check=True, timeout=60)


def candidate_version(checkout, go):
    module_file = checkout / "experiments/wazero/go.mod"
    if not module_file.is_file():
        raise ValueError(f"candidate module is absent in {checkout}; no implicit baseline fallback")
    metadata = json.loads(capture([go, "mod", "edit", "-json", str(module_file)],
                                 env=dict(os.environ, GOWORK="off")))
    replacements = metadata.get("Replace") or []
    if any(item["Old"]["Path"] == MODULE for item in replacements):
        raise ValueError("candidate replacements cannot be compared as official release CLIs")
    selected = [item["Version"] for item in metadata.get("Require", []) if item["Path"] == MODULE]
    if len(selected) != 1 or not re.fullmatch(r"v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?", selected[0]):
        raise ValueError("candidate must select exactly one immutable wazero module version")
    return {"version": selected[0], "moduleFileSha256": sha256(module_file)}


def load_workloads(head):
    root = (head / "benchmarks").resolve()
    manifest = json.loads((root / "fixtures/manifest.json").read_text())
    entries = manifest.get("workloads", [])
    if manifest.get("schemaVersion") != 1 or not manifest.get("origin"):
        raise ValueError("invalid fixture manifest or missing provenance")
    identifiers = [item["id"] for item in entries]
    if len(identifiers) != len(WORKLOADS) or set(identifiers) != set(WORKLOADS):
        raise ValueError("fixture matrix must contain each of the six declared workloads once")
    for item in entries:
        if not isinstance(item.get("expectedStdout"), str) or not item["expectedStdout"].endswith("\n"):
            raise ValueError("fixture stdout must be an exact newline-terminated string")
        for kind in ("wat", "wasm"):
            fixture = (root / item[kind]).resolve()
            if not fixture.is_relative_to(root / "fixtures"):
                raise ValueError("fixture path escapes the head fixtures directory")
            if sha256(fixture) != item[kind + "Sha256"]:
                raise ValueError(f"fixture hash mismatch: {item['id']} {kind}")
    return manifest


def build_candidate(go, selection, directory):
    directory.mkdir(parents=True)
    environment = dict(os.environ, GOBIN=str(directory), GOWORK="off", GOFLAGS="", CGO_ENABLED="0")
    # Each revision chooses an official CLI release. No wasmvm VM/CLI exists
    # yet, and no source-code performance change is inferred from this build.
    subprocess.run([go, "install", f"{MODULE}/cmd/wazero@{selection['version']}"],
                   cwd=directory, env=environment, check=True, timeout=600)
    binary = directory / "wazero"
    build = capture([go, "version", "-m", str(binary)])
    if f"\tmod\t{MODULE}\t{selection['version']}\t" not in build:
        raise ValueError("built wazero CLI does not match the selected dependency")
    return {**selection, "binary": str(binary), "binarySha256": sha256(binary),
            "cliVersion": capture([str(binary), "version"]), "goBuildInfo": build}


def runtime_commands(builds):
    runtimes = []
    for revision in ("base", "head"):
        build = builds[revision]
        for backend in ("interpreter", "compiler"):
            command = [build["binary"], "run"]
            if backend == "interpreter":
                command.append("-interpreter")
            command.append("-cachedir=")
            runtimes.append({"id": f"{revision}-{backend}", "project": "wazero",
                             "backend": backend, "version": build["version"], "command": command})
    flags = {
        "wasmtime": ["run", "-C", "cache=n", "-C", "compiler=cranelift", "-O", "opt-level=2"],
        "wasmer": ["run", "--cranelift", "--disable-cache", "--offline", "--no-tty"],
        "wasmedge": ["--run-mode", "interpreter"],
    }
    for name, version in COMPETITOR_VERSIONS.items():
        binary = shutil.which(name)
        if binary is None:
            raise ValueError(f"required competitor is absent: {name}")
        observed = capture([binary, "--version"])
        first_line = observed.splitlines()[0] if observed else ""
        # WasmEdge prints argv[0], which can be an absolute executable path.
        pattern = rf"(?:^|[/\\]){name}(?:-cli)?(?: version)? {re.escape(version)}(?:\s|$)"
        if not re.search(pattern, first_line, re.IGNORECASE):
            raise ValueError(f"competitor version mismatch: {name}: {observed}")
        backend = "interpreter" if name == "wasmedge" else "cranelift"
        runtimes.append({"id": f"{name}-{backend}", "project": name,
                         "backend": backend, "version": observed,
                         "binarySha256": sha256(Path(binary)), "command": [binary] + flags[name]})
    return runtimes


def measure(runtime, workload, head, warmup, round_number, timeout):
    command = runtime["command"] + [str(head / "benchmarks" / workload["wasm"])]
    started = time.perf_counter_ns()
    result = subprocess.run(command, capture_output=True, timeout=timeout)
    elapsed = (time.perf_counter_ns() - started) / 1_000_000
    if result.returncode != 0 or result.stdout != workload["expectedStdout"].encode():
        raise RuntimeError(f"{runtime['id']}/{workload['id']} correctness failure: "
                           f"exit={result.returncode}, stdout={result.stdout!r}, stderr={result.stderr!r}")
    return {"runtime": runtime["id"], "workload": workload["id"], "warmup": warmup,
            "round": round_number, "elapsedMs": elapsed, "exitCode": result.returncode,
            "stdout": result.stdout.decode(), "stderr": result.stderr.decode(errors="replace"),
            "wasmSha256": workload["wasmSha256"], "command": command}


def validate_matrix(report):
    """Reject partial, duplicated, or incorrect runs before publishing comparisons."""
    runtimes = [item["id"] for item in report["runtimes"]]
    workloads = report["fixtures"]["workloads"]
    identifiers = [item["id"] for item in workloads]
    if len(runtimes) != 7 or set(runtimes) != set(CONFIGURATIONS):
        raise ValueError("runtime matrix must contain each of seven configurations once")
    if len(identifiers) != 6 or set(identifiers) != set(WORKLOADS):
        raise ValueError("workload matrix must contain each of six fixtures once")
    expected = {(runtime, workload, warmup, round_number)
                for runtime in runtimes for workload in identifiers
                for warmup, rounds in ((True, WARMUPS), (False, SAMPLES))
                for round_number in range(rounds)}
    seen = set()
    by_id = {item["id"]: item for item in workloads}
    for item in report["observations"]:
        if type(item["warmup"]) is not bool or type(item["round"]) is not int:
            raise ValueError("invalid warmup or round identity")
        key = (item["runtime"], item["workload"], item["warmup"], item["round"])
        if key not in expected or key in seen:
            raise ValueError("unexpected or duplicate benchmark observation")
        seen.add(key)
        workload = by_id[item["workload"]]
        if (item["exitCode"] != 0 or item["stdout"] != workload["expectedStdout"]
                or item["wasmSha256"] != workload["wasmSha256"]):
            raise ValueError("observation failed correctness or fixture identity")
        elapsed = item["elapsedMs"]
        if isinstance(elapsed, bool) or not isinstance(elapsed, (int, float)) or not math.isfinite(elapsed) or elapsed <= 0:
            raise ValueError("timings must be positive and finite")
    if seen != expected:
        raise ValueError(f"incomplete benchmark matrix: got {len(seen)}, expected {len(expected)}")


def summarize(report):
    validate_matrix(report)
    result = []
    for runtime in CONFIGURATIONS:
        for workload in WORKLOADS:
            values = [item["elapsedMs"] for item in report["observations"]
                      if not item["warmup"] and item["runtime"] == runtime and item["workload"] == workload]
            result.append({"runtime": runtime, "workload": workload, **summary(values)})
    return result


def median_delta(base, head):
    if not all(math.isfinite(value) and value > 0 for value in (base, head)):
        raise ValueError("comparison medians must be finite and positive")
    return (head / base - 1) * 100


def markdown(report):
    lines = ["## Candidate runtime benchmark", "",
             f"Base `{report['baseSha']}` → head `{report['headSha']}`.", ""]
    if report.get("status") != "complete":
        # Failed logs remain in the artifact. Do not present partial medians as
        # a comparison or interpolate arbitrary child-process output as Markdown.
        return "\n".join(lines + ["**Benchmark failed; no performance comparison is accepted.**",
                                    "See the raw JSON artifact for the failure and completed observations.", ""])
    calculated = summarize(report)
    indexed = {(item["runtime"], item["workload"]): item for item in calculated}
    builds = report["builds"]
    versions_equal = builds["base"]["version"] == builds["head"]["version"]
    lines += [f"Official wazero CLI dependency selection: base `{builds['base']['version']}`, "
              f"head `{builds['head']['version']}`; both built with the same Go toolchain.", "",
              "This measures the candidate dependency and backend, not production wasmvm code. The phase report separately measures the owned wasmvm API and production CLI."]
    if versions_equal:
        lines.append("**The selected wazero version is unchanged. Differences are repeated-run measurements, not evidence of a head engine-code change.**")
    lines += ["", "Same runner and head fixtures for all seven configurations; randomized interleaving, "
              "two warmups and seven measured samples per case (378 successful executions). "
              "Median milliseconds include process startup, validation/compilation, instantiation, execution, and shutdown.", "",
              "| Workload | Backend | Base ms | Head ms | Head delta | Advisory |",
              "| --- | --- | ---: | ---: | ---: | --- |"]
    for workload in WORKLOADS:
        for backend in ("interpreter", "compiler"):
            base = indexed[(f"base-{backend}", workload)]["medianMs"]
            head = indexed[(f"head-{backend}", workload)]["medianMs"]
            delta = median_delta(base, head)
            advisory = "investigate" if abs(delta) >= 10 and abs(head - base) >= 1 else "within advisory band"
            lines.append(f"| {workload} | {backend} | {base:.3f} | {head:.3f} | {delta:+.1f}% | {advisory} |")
    lines += ["", "Negative delta means lower latency. The advisory requires both ≥10% and ≥1 ms change; "
              "it is not a statistical significance test or a speed gate. Shared-runner noise can exceed it. "
              "Correctness failures and incomplete runs fail the job.", "",
              "| Workload | Head interpreter ms | Head compiler ms | Wasmtime ms | Wasmer ms | WasmEdge interpreter ms |",
              "| --- | ---: | ---: | ---: | ---: | ---: |"]
    competitors = ("head-interpreter", "head-compiler", "wasmtime-cranelift", "wasmer-cranelift", "wasmedge-interpreter")
    for workload in WORKLOADS:
        values = " | ".join(f"{indexed[(runtime, workload)]['medianMs']:.3f}" for runtime in competitors)
        lines.append(f"| {workload} | {values} |")
    lines += ["", "Competitors: Wasmtime 48.0.1 / Cranelift opt-level 2; Wasmer 7.3.0 / Cranelift; "
              "WasmEdge 0.17.1 / interpreter. Compilation caches are disabled; OS file/executable caches remain warm. "
              "Native compilers may inline the function-call workload. These small synthetic CLI measurements "
              "are not steady-state throughput, memory usage, or full application benchmarks.", "",
              "Raw JSON retains exact revisions, compiler/build metadata, tool versions, fixture hashes, command lines, "
              "all samples, minimum/median/p95, and host information. With seven samples, nearest-rank p95 is the maximum.", ""]
    return "\n".join(lines)


def execute(report, head, timeout, seed):
    cases = [(runtime, workload) for runtime in report["runtimes"] for workload in report["fixtures"]["workloads"]]
    randomizer = random.Random(seed)
    for warmup, rounds in ((True, WARMUPS), (False, SAMPLES)):
        for round_number in range(rounds):
            order = list(cases)
            randomizer.shuffle(order)
            for runtime, workload in order:
                report["observations"].append(measure(runtime, workload, head, warmup, round_number, timeout))
            print(f"{'warmup' if warmup else 'measured'} round {round_number + 1}/{rounds} complete", flush=True)
    report["summary"] = summarize(report)
    report["status"] = "complete"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("base-dir", "head-dir", "output-dir"):
        parser.add_argument(f"--{name}", required=True, type=Path)
    for name in ("base-sha", "head-sha"):
        parser.add_argument(f"--{name}", required=True)
    parser.add_argument("--seed", type=int, default=20261026)
    parser.add_argument("--timeout", type=float, default=30)
    args = parser.parse_args()
    if not math.isfinite(args.timeout) or args.timeout <= 0:
        parser.error("timeout must be positive and finite")
    output = args.output_dir.resolve()
    output.mkdir(parents=True, exist_ok=True)
    report = {"schemaVersion": 1, "startedAt": datetime.now(timezone.utc).isoformat(),
              "baseSha": args.base_sha, "headSha": args.head_sha, "status": "running",
              "metric": "separate-process CLI wall time", "warmupsPerCase": WARMUPS,
              "samplesPerCase": SAMPLES, "seed": args.seed, "observations": [],
              "scope": "official candidate CLI dependency/backend comparison; no production wasmvm implementation",
              "cachePolicy": "compilation caches disabled; OS executable/filesystem cache warmed, not flushed",
              "host": {"platform": platform.platform(), "machine": platform.machine(),
                       "logicalCPUs": os.cpu_count(), "python": platform.python_version()}}
    try:
        base, head = args.base_dir.resolve(), args.head_dir.resolve()
        verify_checkout(base, args.base_sha)
        verify_checkout(head, args.head_sha)
        go = shutil.which("go")
        if go is None:
            raise ValueError("Go toolchain is absent")
        report["host"]["go"] = capture([go, "version"])
        report["fixtures"] = load_workloads(head)
        selections = {"base": candidate_version(base, go), "head": candidate_version(head, go)}
        with tempfile.TemporaryDirectory(prefix="wasmvm-ci-bench-") as temporary:
            report["builds"] = {name: build_candidate(go, selection, Path(temporary) / name)
                                for name, selection in selections.items()}
            report["runtimes"] = runtime_commands(report["builds"])
            if hasattr(os, "getloadavg"):
                report["host"]["loadAverageBefore"] = list(os.getloadavg())
            execute(report, head, args.timeout, args.seed)
    except Exception as error:
        report["status"] = "failed"
        report["failure"] = f"{type(error).__name__}: {error}"
        raise
    finally:
        report["finishedAt"] = datetime.now(timezone.utc).isoformat()
        if hasattr(os, "getloadavg"):
            report["host"]["loadAverageAfter"] = list(os.getloadavg())
        (output / "results.json").write_text(json.dumps(report, indent=2) + "\n")
        (output / "report.md").write_text(markdown(report))
    print(f"Benchmark results: {output / 'results.json'}", flush=True)


if __name__ == "__main__":
    main()
