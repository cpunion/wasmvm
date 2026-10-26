"""Measure complete production CLI wall time independently of API phases."""
import json
import random
import statistics
import subprocess
import time

ROUNDS = 3


def configurations(builds):
    result = []
    for ref in ("base", "head"):
        for compiler in ("go", "llgo"):
            build = builds[f"{ref}-{compiler}"]
            if build["status"] == "baseline-unavailable":
                continue
            cli = build["cli"]
            if cli["status"] == "baseline-unavailable" and ref == "base":
                continue
            if cli["status"] != "available":
                raise ValueError("head production CLI is unavailable")
            result.append({"id": f"{ref}-{compiler}", "binary": cli["binary"],
                           "binarySha256": cli["binarySha256"], "sourceRevision": build["sourceRevision"]})
    return result


def run(head, builds, fixtures, timeout, seed, result=None):
    if result is None:
        result = {"rounds": ROUNDS, "runtimes": configurations(builds), "observations": []}
    cases = [(runtime, workload) for runtime in result["runtimes"] for workload in fixtures["workloads"]]
    randomizer = random.Random(seed)
    for number in range(ROUNDS):
        randomizer.shuffle(cases)
        for runtime, workload in cases:
            command = [runtime["binary"], "run", "--backend", "interpreter", "--profile", "core2",
                       "--export", "run", "--timeout", "0", str(head / "benchmarks/embedded" / workload["wasm"]),
                       f"i32:{workload['seed']}", f"i32:{workload['iterations']}"]
            start = time.perf_counter_ns()
            process = subprocess.run(command, capture_output=True, text=True, timeout=timeout, check=True)
            elapsed = time.perf_counter_ns() - start
            result["observations"].append({"runtime": runtime["id"], "workload": workload["id"],
                "round": number, "processWallNs": elapsed, "result": json.loads(process.stdout),
                "wasmSha256": workload["wasmSha256"], "exitCode": process.returncode,
                "stderr": process.stderr, "command": command})
        print(f"Owned production CLI round {number+1}/{ROUNDS} complete", flush=True)
    summarize(result, builds, fixtures)
    return result


def summarize(series, builds, fixtures):
    runtimes = configurations(builds)
    if type(series["rounds"]) is not int or series["rounds"] != ROUNDS or series["runtimes"] != runtimes:
        raise ValueError("production CLI build identity mismatch")
    workloads = {w["id"]: w for w in fixtures["workloads"]}
    expected = {(r["id"], w, n) for r in runtimes for w in workloads for n in range(ROUNDS)}
    seen = set()
    for item in series["observations"]:
        key = item["runtime"], item["workload"], item["round"]
        if key not in expected or key in seen or type(item["round"]) is not int:
            raise ValueError("duplicate or unexpected production CLI process")
        seen.add(key)
        workload = workloads[item["workload"]]
        result = item["result"]
        checksum = workload["expectedChecksum"]
        signed = checksum if checksum < 2**31 else checksum - 2**32
        wanted = {"backend": "interpreter", "profile": "core2", "export": "run",
                  "results": [{"type": "i32", "bits": f"0x{checksum:08x}", "value": str(signed)}]}
        if type(item["exitCode"]) is not int or item["exitCode"] != 0 or result != wanted or item["wasmSha256"] != workload["wasmSha256"]:
            raise ValueError("production CLI correctness or fixture identity failure")
        if type(item["processWallNs"]) is not int or item["processWallNs"] <= 0:
            raise ValueError("invalid production CLI wall time")
    if seen != expected:
        raise ValueError("production CLI process matrix is incomplete")
    return [{"runtime": r["id"], "workload": w,
             "medianNs": statistics.median(item["processWallNs"] for item in series["observations"]
                                            if item["runtime"] == r["id"] and item["workload"] == w)}
            for r in runtimes for w in workloads]


def markdown(series, builds, fixtures):
    indexed = {(r["runtime"], r["workload"]): r["medianNs"] for r in summarize(series, builds, fixtures)}
    lines = ["<details><summary>Owned production CLI latency, separately measured</summary>", "",
             "Three fresh CLI processes per case. Each invokes the actual revision's cmd/wasmvm binary once; "
             "wall time includes startup, signal/context setup, file I/O, preparation, execution, JSON and exit. "
             "It is not subtracted from any API/dispatch timing. Values are milliseconds.", "",
             "| Workload | Go base | Go head | LLGo base | LLGo head |",
             "| --- | ---: | ---: | ---: | ---: |"]
    for workload in fixtures["workloads"]:
        cells = []
        for runtime in ("base-go", "head-go", "base-llgo", "head-llgo"):
            number = indexed.get((runtime, workload["id"]))
            cells.append("unavailable" if number is None else f"{number / 1e6:.4f}")
        lines.append(f"| {workload['id']} | " + " | ".join(cells) + " |")
    return "\n".join(lines + ["", "</details>", ""])
