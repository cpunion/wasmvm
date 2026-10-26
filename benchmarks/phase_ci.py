#!/usr/bin/env python3
"""Measure embedding API phases directly, without subtracting CLI timings."""

import argparse
from datetime import datetime, timezone
import hashlib
import json
import math
import os
from pathlib import Path
import platform
import random
import selectors
import shutil
import statistics
import subprocess
import tempfile
import time

import ci
import own_build
import own_cli
import subsystems


ROUNDS = 3
SAMPLES = 7
WARMUPS = 2
WORKLOADS = ("call-control", "integer-loop", "integer-sustained", "function-calls",
             "memory-load-store", "memory-sustained", "float-convert", "control-multivalue",
             "indirect-reference", "simd-shuffle", "bulk-unaligned", "deep-calls",
             "image-filter-c", "packet-crc-rust")
PHASES = ("load", "engineSetup", "compileValidate", "instantiate", "exportLookup", "firstCall", "cleanup")
MAX_OUTPUT = 2 * 1024 * 1024
OWN_CONFIGURATIONS = tuple(f"wasmvm-{ref}-{compiler}-{mode}"
                           for ref in ("base", "head") for compiler in ("go", "llgo")
                           for mode in ("api", "dispatch"))
CONFIGURATIONS = OWN_CONFIGURATIONS + ci.CONFIGURATIONS


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load_workloads(head):
    root = (head / "benchmarks/embedded").resolve()
    manifest = json.loads((root / "fixtures/manifest.json").read_text())
    portfolio = json.loads((root / "fixtures/portfolio.json").read_text())
    if portfolio.get("schemaVersion") != 1 or portfolio.get("portfolioVersion") != "core2-v1":
        raise ValueError("invalid Core 2.0 portfolio version")
    for item in manifest["workloads"]:
        item["producer"] = {"version": manifest["producer"]}
    manifest["producer"] = "mixed; see per-workload producer identities"
    manifest["profile"] = portfolio["profile"]
    manifest["portfolio"] = portfolio
    manifest["workloads"].extend(portfolio["workloads"])
    entries = manifest.get("workloads", [])
    if manifest.get("schemaVersion") != 1 or not manifest.get("origin"):
        raise ValueError("invalid embedded fixture provenance")
    if len(entries) != len(WORKLOADS) or {item["id"] for item in entries} != set(WORKLOADS):
        raise ValueError("embedded fixture matrix must contain exactly the declared cases")
    for item in entries:
        for key, upper in (("seed", 2**32 - 1), ("expectedChecksum", 2**32 - 1), ("iterations", 2**31 - 1)):
            if type(item[key]) is not int or not 0 <= item[key] <= upper:
                raise ValueError(f"invalid fixture {key}")
        if item["iterations"] < 1:
            raise ValueError("workload iterations must be positive")
        expected_unit = "ns/call" if item["id"] == "call-control" else "ns/iteration"
        if item["unit"] != expected_unit or (item["id"] == "call-control" and item["iterations"] != 1):
            raise ValueError("invalid workload normalization unit")
        for kind in ("wat", "wasm"):
            path = (root / item[kind]).resolve()
            if not path.is_relative_to(root / "fixtures") or digest(path) != item[kind + "Sha256"]:
                raise ValueError("embedded fixture path or hash mismatch")
    for item in portfolio["workloads"]:
        source = (root / item["source"]).resolve()
        if not source.is_relative_to(root / "portfolio") or digest(source) != item["sourceSha256"]:
            raise ValueError("portfolio source path or hash mismatch")
    return manifest


def build_candidates(head, base, temporary):
    go = shutil.which("go")
    if go is None:
        raise ValueError("Go toolchain is absent")
    source = head / "benchmarks/embedded/wazero"
    inputs = {str(path.relative_to(source)): digest(path) for path in sorted(source.rglob("*"))
              if path.is_file() and path.suffix in (".go", ".mod", ".sum")}
    if "main.go" not in inputs:
        raise ValueError("embedded wazero runner source is absent")
    builds = {}
    for name, checkout in (("base", base), ("head", head)):
        selection = ci.candidate_version(checkout, go)
        target = temporary / name
        shutil.copytree(source, target)
        environment = dict(os.environ, GOWORK="off", GOFLAGS="", CGO_ENABLED="0")
        subprocess.run([go, "mod", "edit", f"-require={ci.MODULE}@{selection['version']}"],
                       cwd=target, env=environment, check=True, timeout=60)
        subprocess.run([go, "mod", "tidy"], cwd=target, env=environment, check=True, timeout=120)
        binary = target / "wazero-phase"
        subprocess.run([go, "build", "-trimpath", "-o", str(binary), "."],
                       cwd=target, env=environment, check=True, timeout=600)
        dependency = json.loads(ci.capture([go, "list", "-m", "-json", ci.MODULE], cwd=target, env=environment))
        if dependency["Version"] != selection["version"] or dependency.get("Replace"):
            raise ValueError("built embedding dependency differs from the selected pin")
        builds[name] = {**selection, "binary": str(binary), "binarySha256": digest(binary),
                        "dependency": dependency, "sourceHashes": inputs,
                        "goBuildInfo": ci.capture([go, "version", "-m", str(binary)])}
    return builds


def own_commands(builds):
    result = []
    for ref in ("base", "head"):
        for compiler in ("go", "llgo"):
            build = builds[f"{ref}-{compiler}"]
            if build["status"] == "baseline-unavailable" and ref == "base":
                continue
            if build["status"] != "available":
                raise ValueError("required own-engine build is unavailable")
            for mode in ("api", "dispatch"):
                result.append({"id": f"wasmvm-{ref}-{compiler}-{mode}", "engine": "wasmvm",
                    "backend": "interpreter", "mode": mode, "hostCompiler": compiler,
                    "compiler": "gc" if compiler == "go" else "llgo",
                    "compilerRevision": build["toolchain"]["revision"],
                    "sourceRevision": build["sourceRevision"], "sourceDigest": build["sourceDigest"],
                    "goos": build["toolchain"]["goos"], "goarch": build["toolchain"]["goarch"],
                    "binarySha256": build["binarySha256"], "command": [build["binary"], "--mode", mode]})
    return result


def runtime_commands(builds, c_directory, own_builds):
    result = own_commands(own_builds)
    for name in ("base", "head"):
        for backend in ("interpreter", "compiler"):
            result.append({"id": f"{name}-{backend}", "backend": backend,
                           "version": builds[name]["version"], "command": [builds[name]["binary"]]})
    for name, backend in (("wasmtime", "compiler"), ("wasmer", "compiler"), ("wasmedge", "interpreter")):
        binary = c_directory / (name + "-phase")
        if not binary.is_file():
            raise ValueError(f"embedded competitor is absent: {name}")
        result.append({"id": name + ("-interpreter" if name == "wasmedge" else "-cranelift"),
                       "backend": backend, "version": ci.COMPETITOR_VERSIONS[name],
                       "binarySha256": digest(binary), "command": [str(binary)]})
    return result


def invoke(command, timeout):
    """Observe process readiness before releasing the child into its API phases."""
    started = time.perf_counter_ns()
    ready_ns = None
    output = {"stdout": bytearray(), "stderr": bytearray()}
    with subprocess.Popen(command, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                          stderr=subprocess.PIPE, bufsize=0) as process:
        try:
            with selectors.DefaultSelector() as selector:
                selector.register(process.stdout, selectors.EVENT_READ, "stdout")
                selector.register(process.stderr, selectors.EVENT_READ, "stderr")
                while selector.get_map():
                    remaining = timeout - (time.perf_counter_ns() - started) / 1e9
                    if remaining <= 0:
                        raise TimeoutError("embedded runner exceeded its process deadline")
                    for key, _ in selector.select(remaining):
                        chunk = os.read(key.fileobj.fileno(), 16384)
                        if not chunk:
                            selector.unregister(key.fileobj)
                            continue
                        output[key.data].extend(chunk)
                        if sum(map(len, output.values())) > MAX_OUTPUT:
                            raise ValueError("embedded runner exceeded output limit")
                        if ready_ns is None and b"\n" in output["stdout"]:
                            ready_ns = time.perf_counter_ns() - started
                            first = bytes(output["stdout"]).split(b"\n", 1)[0]
                            if first != b'{"ready":true}':
                                raise ValueError("embedded runner readiness protocol mismatch")
                            try:
                                process.stdin.write(b"run\n")
                                process.stdin.close()
                            except BrokenPipeError as error:
                                raise ValueError("embedded runner closed before acknowledging readiness") from error
                remaining = timeout - (time.perf_counter_ns() - started) / 1e9
                process.wait(timeout=max(remaining, 0.001))
        except BaseException:
            process.kill()
            process.wait()
            raise
    wall_ns = time.perf_counter_ns() - started
    stdout = bytes(output["stdout"]).decode()
    stderr = bytes(output["stderr"]).decode(errors="replace")
    if process.returncode != 0:
        raise RuntimeError(f"embedded runner exit {process.returncode}: {stderr}")
    lines = stdout.splitlines()
    if ready_ns is None or len(lines) != 2:
        raise ValueError("embedded runner must emit readiness and exactly one result")
    return {"processReadyNs": ready_ns, "processWallNs": wall_ns, "record": json.loads(lines[1]),
            "exitCode": process.returncode, "stderr": stderr, "command": command}


def duration(value):
    if type(value) is not int or value < 0:
        raise ValueError("durations must be nonnegative integer nanoseconds")


def validate_observation(observation, runtime, workload):
    record = observation["record"]
    if observation["exitCode"] != 0 or observation["wasmSha256"] != workload["wasmSha256"]:
        raise ValueError("phase observation exit or fixture identity mismatch")
    duration(observation["processReadyNs"])
    duration(observation["processWallNs"])
    if observation["processReadyNs"] > observation["processWallNs"]:
        raise ValueError("process readiness follows process completion")
    for key, value in (("schemaVersion", 1), ("backend", runtime["backend"]),
                       ("iterations", workload["iterations"]), ("seed", workload["seed"]),
                       ("samples", SAMPLES), ("warmups", WARMUPS), ("result", workload["expectedChecksum"])):
        if type(record[key]) is not type(value) or record[key] != value:
            raise ValueError(f"phase runner {key} mismatch")
    own = runtime.get("engine") == "wasmvm"
    if own:
        for key in ("engine", "sourceRevision", "compiler", "compilerRevision", "mode", "goos", "goarch"):
            if record.get(key) != runtime[key]:
                raise ValueError(f"own-engine {key} mismatch")
        if record.get("runtimeVersion") != own_build.GO_VERSION or record.get("policy") != "unmetered; bounded cancellation; checked Core 2.0":
            raise ValueError("own-engine runtime or execution policy mismatch")
    if set(record["phasesNs"]) != set(PHASES):
        raise ValueError("phase runner omitted or added a canonical timing boundary")
    for value in record["phasesNs"].values():
        duration(value)
    for key, count in (("warmCallNs", SAMPLES), ("warmupCallNs", WARMUPS), ("timerPairNs", SAMPLES)):
        if len(record[key]) != count:
            raise ValueError("phase runner sample count mismatch")
        for value in record[key]:
            duration(value)
    for key, count in (("warmResults", SAMPLES), ("warmupResults", WARMUPS)):
        if len(record[key]) != count or any(type(v) is not int or v != workload["expectedChecksum"] for v in record[key]):
            raise ValueError("phase runner returned an incorrect checksum")
    if "preparationPartsNs" in record:
        parts = record["preparationPartsNs"]
        expected_parts = {"decode", "validate", "lower", "slots"} if own else {"parse", "validate"}
        if set(parts) != expected_parts:
            raise ValueError("unexpected preparation detail")
        for value in parts.values():
            duration(value)
        aggregate = record["phasesNs"]["compileValidate"]
        if sum(parts.values()) > aggregate or not own and sum(parts.values()) != aggregate:
            raise ValueError("preparation parts disagree with their aggregate")
    if own and ("preparationPartsNs" in record) != (runtime["mode"] == "dispatch"):
        raise ValueError("own-engine preparation detail belongs only to dispatch diagnostics")
    # These API intervals are disjoint. Startup notification and JSON/bookkeeping
    # are deliberately not subtracted to invent a 'pure guest' measurement.
    measured = sum(record["phasesNs"].values()) + sum(record["warmCallNs"]) + sum(record["warmupCallNs"])
    if observation["processReadyNs"] + measured > observation["processWallNs"]:
        raise ValueError("disjoint readiness and phase intervals exceed process wall time")


def summarize(report):
    runtimes = {r["id"]: r for r in report["runtimes"]}
    workloads = {w["id"]: w for w in report["fixtures"]["workloads"]}
    required_builds = {f"{ref}-{compiler}" for ref in ("base", "head") for compiler in ("go", "llgo")}
    if set(report.get("ownBuilds", {})) != required_builds:
        raise ValueError("own-engine build provenance is incomplete")
    if report["ownBuilds"]["base-go"]["status"] != report["ownBuilds"]["base-llgo"]["status"]:
        raise ValueError("a compiler failure cannot be labeled as an absent baseline engine")
    for name, build in report["ownBuilds"].items():
        expected_sha = report["baseSha"] if name.startswith("base-") else report["headSha"]
        if build["sourceRevision"] != expected_sha:
            raise ValueError("own-engine build source does not match the selected commit")
        if build["sourceDigest"] != own_build.source_digest(build["sourceHashes"]):
            raise ValueError("own-engine source digest mismatch")
        if build["status"] == "baseline-unavailable" and (not name.startswith("base-") or not build.get("reason")):
            raise ValueError("invalid baseline-unavailable provenance")
    expected_own = {r["id"]: r for r in own_commands(report["ownBuilds"])}
    expected_ids = set(ci.CONFIGURATIONS) | set(expected_own)
    if len(report["runtimes"]) != len(expected_ids) or set(runtimes) != expected_ids or len(workloads) != len(WORKLOADS) or set(workloads) != set(WORKLOADS):
        raise ValueError("incomplete embedded runtime/workload configuration")
    for name, expected in expected_own.items():
        if runtimes[name] != expected:
            raise ValueError("own-engine runtime does not match build provenance")
    own_cli.summarize(report["productionCLI"], report["ownBuilds"], report["fixtures"])
    subsystems.summarize(report["subsystems"], report["ownBuilds"])
    expected = {(r, w, n) for r in runtimes for w in workloads for n in range(ROUNDS)}
    seen = set()
    for item in report["observations"]:
        key = (item["runtime"], item["workload"], item["round"])
        if type(item["round"]) is not int or key not in expected or key in seen:
            raise ValueError("duplicate or unexpected embedded process observation")
        seen.add(key)
        validate_observation(item, runtimes[item["runtime"]], workloads[item["workload"]])
    if seen != expected:
        raise ValueError("incomplete embedded process matrix")
    result = []
    for runtime in CONFIGURATIONS:
        if runtime not in runtimes:
            continue
        for workload in WORKLOADS:
            selected = [item for item in report["observations"] if item["runtime"] == runtime and item["workload"] == workload]
            row = {"runtime": runtime, "workload": workload, "processRounds": ROUNDS}
            for phase in PHASES:
                row[phase + "Ns"] = statistics.median(item["record"]["phasesNs"][phase] for item in selected)
            for key in ("processReadyNs", "processWallNs"):
                row[key] = statistics.median(item[key] for item in selected)
            row["warmCallNs"] = statistics.median(statistics.median(item["record"]["warmCallNs"]) for item in selected)
            row["timerPairNs"] = statistics.median(statistics.median(item["record"]["timerPairNs"]) for item in selected)
            row["normalizedNs"] = row["warmCallNs"] / workloads[workload]["iterations"]
            row["unit"] = workloads[workload]["unit"]
            if "preparationPartsNs" in selected[0]["record"]:
                if any(set(item["record"].get("preparationPartsNs", {})) != set(selected[0]["record"]["preparationPartsNs"]) for item in selected):
                    raise ValueError("incomplete preparation detail across process rounds")
                row["preparationPartsNs"] = {part: statistics.median(
                    item["record"]["preparationPartsNs"][part] for item in selected)
                    for part in selected[0]["record"]["preparationPartsNs"]}
            result.append(row)
    return result


def markdown(report):
    lines = ["## wasmvm implementation and direct embedding phases", "",
             f"Base `{report['baseSha']}` → head `{report['headSha']}`.", ""]
    if report.get("status") != "complete":
        return "\n".join(lines + ["**Phase measurement failed; no phase comparison is accepted.** See raw artifacts and job logs.", ""])
    rows = summarize(report)
    indexed = {(row["runtime"], row["workload"]): row for row in rows}
    lines += ["Owned engines are built from the exact commits above with identical head instrumentation; "
              "Go and LLGo are separate measurements. The primary rows use the public VM API. "
              "The interpreter retains cancellation polling, bounds/trap checks and quotas; optional fuel is off.", "",
              "| Workload | Unit | Own Go base | Own Go head | Delta | Own LLGo base | Own LLGo head | Delta |",
              "| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |"]
    def own_cells(workload, compiler, mode):
        base = indexed.get((f"wasmvm-base-{compiler}-{mode}", workload))
        head = indexed[(f"wasmvm-head-{compiler}-{mode}", workload)]["normalizedNs"]
        if base is None:
            return f"unavailable | {head:.2f} | unavailable"
        baseline = base["normalizedNs"]
        delta = f"{ci.median_delta(baseline, head):+.1f}%" if baseline > 0 and head > 0 else "unresolved"
        return f"{baseline:.2f} | {head:.2f} | {delta}"
    for workload in WORKLOADS:
        unit = indexed[("wasmvm-head-go-api", workload)]["unit"]
        lines.append(f"| {workload} | {unit} | {own_cells(workload, 'go', 'api')} | {own_cells(workload, 'llgo', 'api')} |")
    for compiler in ("go", "llgo"):
        base = report["ownBuilds"][f"base-{compiler}"]
        head = report["ownBuilds"][f"head-{compiler}"]
        if base["status"] == "baseline-unavailable":
            lines.append(f"\nOwn {compiler} baseline unavailable: the selected commit predates the Engine/Store API; no revision-regression claim is accepted.")
        elif base["sourceDigest"] == head["sourceDigest"]:
            lines.append(f"\nOwn {compiler} engine inputs are identical: this is an A/A baseline, so timing differences are variation rather than an implementation optimization.")
    lines += ["", "| Workload | Unit | Own Go API | Own LLGo API | wazero interpreter | wazero native | Wasmtime Cranelift | Wasmer Cranelift | WasmEdge interpreter |",
              "| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |"]
    compared = ("wasmvm-head-go-api", "wasmvm-head-llgo-api", "head-interpreter", "head-compiler", "wasmtime-cranelift", "wasmer-cranelift", "wasmedge-interpreter")
    for workload in WORKLOADS:
        values = " | ".join(f"{indexed[(runtime, workload)]['normalizedNs']:.2f}" for runtime in compared)
        lines.append(f"| {workload} | {indexed[(compared[0], workload)]['unit']} | {values} |")
    lines += ["", "<details><summary>Internal dispatch diagnostics and compiler phases</summary>", "",
              "These diagnostic rows invoke the same checked interpreter with reused frames and cancellable contexts, "
              "without the public root/marshalling/admission driver. They do not replace the public API comparison. "
              "No medians are subtracted to manufacture a pure instruction or boundary cost.", "",
              "| Workload | Unit | Go base dispatch | Go head dispatch | Delta | LLGo base dispatch | LLGo head dispatch | Delta |",
              "| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |"]
    for workload in WORKLOADS:
        lines.append(f"| {workload} | {indexed[('wasmvm-head-go-api', workload)]['unit']} | {own_cells(workload, 'go', 'dispatch')} | {own_cells(workload, 'llgo', 'dispatch')} |")
    lines += ["", "Head owned preparation details (milliseconds); all base/head values remain in JSON; the outer preparation interval also includes phase transitions. "
              "These produce interpreter slot code, not native code. Parts are not added a second time to preparation.", "",
              "| Runtime | Workload | Decode | Validate | IR lowering | Slot allocation |",
              "| --- | --- | ---: | ---: | ---: | ---: |"]
    for row in rows:
        parts = row.get("preparationPartsNs", {})
        if "decode" in parts and row["runtime"].startswith("wasmvm-head-"):
            values = " | ".join(f"{parts[key] / 1e6:.6f}" for key in ("decode", "validate", "lower", "slots"))
            lines.append(f"| {row['runtime']} | {row['workload']} | {values} |")
    lines += ["", "</details>", "", own_cli.markdown(report["productionCLI"], report["ownBuilds"], report["fixtures"]),
              subsystems.markdown(report["subsystems"], report["ownBuilds"]),
              "### Wazero dependency selection (reference only)", ""]
    versions = report["builds"]
    lines += [f"wazero base `{versions['base']['version']}`, head `{versions['head']['version']}`; "
              "identical head instrumentation and fixture bytes measure each selected dependency."]
    if versions["base"]["version"] == versions["head"]["version"]:
        lines.append("**Both revisions select the same engine version; differences represent measurement variation, not a code optimization.**")
    lines += ["", "**Warm execution is timed directly around the embedding API call**, after compile, instantiate, "
              "export lookup, first call, and two extra warmup calls. There is no guest WASI, output, or host callback. "
              "It still includes API dispatch and result marshalling. The empty-call control exposes that boundary; "
              "no startup, empty-call, or clock median is subtracted.", "",
              "Three fresh processes per case; seven measured calls in each. Values are medians of the three process medians. "
              "`ns/iteration` means one **workload loop iteration**, not one Wasm instruction. Memory reset is included. "
              "Native compilers may inline source-level calls.", "",
              "| Workload | Unit | Wazero base interpreter | Wazero head interpreter | Delta | Wazero base compiler | Wazero head compiler | Delta |",
              "| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |"]
    for workload in WORKLOADS:
        values = [indexed[(f"{ref}-{mode}", workload)]["normalizedNs"] for mode in ("interpreter", "compiler") for ref in ("base", "head")]
        deltas = [f"{ci.median_delta(values[n], values[n+1]):+.1f}%" if values[n] > 0 and values[n+1] > 0 else "unresolved" for n in (0, 2)]
        lines.append(f"| {workload} | {indexed[('head-compiler', workload)]['unit']} | {values[0]:.2f} | {values[1]:.2f} | {deltas[0]} | {values[2]:.2f} | {values[3]:.2f} | {deltas[1]} |")
    lines += ["", "| Workload | Unit | Head interpreter | Head compiler | Wasmtime Cranelift | Wasmer Cranelift | WasmEdge interpreter |",
              "| --- | --- | ---: | ---: | ---: | ---: | ---: |"]
    for workload in WORKLOADS:
        values = " | ".join(f"{indexed[(runtime, workload)]['normalizedNs']:.2f}" for runtime in ("head-interpreter", "head-compiler", "wasmtime-cranelift", "wasmer-cranelift", "wasmedge-interpreter"))
        lines.append(f"| {workload} | {indexed[('head-compiler', workload)]['unit']} | {values} |")
    lines += ["", "<details><summary>Head public API phases (milliseconds); full base/head and dispatch phases in JSON</summary>", "",
              "| Runtime | Workload | Process ready | Read | Engine/store | Decode+validate/compile | Instantiate | Lookup | First call | Warm call | Cleanup | Clock pair |",
              "| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |"]
    keys = ("processReadyNs", "loadNs", "engineSetupNs", "compileValidateNs", "instantiateNs", "exportLookupNs", "firstCallNs", "warmCallNs", "cleanupNs", "timerPairNs")
    for row in rows:
        if row["runtime"] not in compared:
            continue
        values = " | ".join(f"{row[key] / 1e6:.6f}" for key in keys)
        lines.append(f"| {row['runtime']} | {row['workload']} | {values} |")
    lines += ["", "</details>", "",
              "Process readiness includes OS launch, native/Go initialization, and pipe notification. A parent acknowledgement "
              "releases the child into measured API work only after readiness is observed. Process wall time (all calls, "
              "bookkeeping, serialization, IPC, and shutdown) is retained separately in JSON.", "",
              "Preparation boundaries follow actual APIs: wazero/Wasmtime/Wasmer combine decode, validation, and compilation. "
              "WasmEdge interpreter records separate parse/validate details and their sum; it does not compile native code. "
              "Calls and assertions are outside preparation timers. Separate standalone validation would repeat work and is not added.", "",
              "Zero-duration or near-clock-cost controls are resolution-limited. Clock-pair samples are controls, not corrections. "
              "No disk compilation cache is enabled. Timings are advisory, with only three process repetitions and shared-runner noise; "
              "incorrect results, incomplete matrices, or broken protocols fail the job. These Go/LLGo/C embedding paths have different "
              "host-call overhead, and the report does not claim to remove it. Native wasmvm is not implemented; "
              "hosted measurements do not pass the M2 controlled performance gates.", ""]
    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("base-dir", "head-dir", "output-dir", "tools-dir"):
        parser.add_argument(f"--{name}", type=Path, required=True)
    for name in ("base-sha", "head-sha"):
        parser.add_argument(f"--{name}", required=True)
    parser.add_argument("--timeout", type=float, default=60)
    args = parser.parse_args()
    if not math.isfinite(args.timeout) or args.timeout <= 0:
        parser.error("timeout must be positive and finite")
    output = args.output_dir.resolve()
    output.mkdir(parents=True, exist_ok=True)
    report = {"schemaVersion": 1, "status": "running", "baseSha": args.base_sha, "headSha": args.head_sha,
              "processRounds": ROUNDS, "samplesPerProcess": SAMPLES, "warmupsPerProcess": WARMUPS,
              "processDeadlineSeconds": args.timeout,
              "startedAt": datetime.now(timezone.utc).isoformat(), "observations": [],
              "host": {"platform": platform.platform(), "machine": platform.machine(), "logicalCPUs": os.cpu_count()},
              "seed": 20261026, "metric": "direct embedding API phase nanoseconds"}
    try:
        base, head = args.base_dir.resolve(), args.head_dir.resolve()
        ci.verify_checkout(base, args.base_sha)
        ci.verify_checkout(head, args.head_sha)
        report["fixtures"] = load_workloads(head)
        report["host"]["go"] = ci.capture(["go", "version"])
        report["host"]["cpu"] = ci.capture(["lscpu"]) if platform.system() == "Linux" else ci.capture(["sysctl", "-n", "machdep.cpu.brand_string"])
        report["host"]["loadAverageBefore"] = os.getloadavg()
        report["host"]["cachePolicy"] = "fresh child processes; no disk compilation cache; filesystem cache not flushed"
        report["installation"] = json.loads((args.tools_dir / "installation.json").read_text())
        with tempfile.TemporaryDirectory(prefix="wasmvm-phase-") as temporary:
            temporary = Path(temporary)
            report["builds"] = build_candidates(head, base, temporary)
            report["ownBuilds"] = own_build.build(head, base, temporary, args.base_sha, args.head_sha)
            c_directory = temporary / "c"
            subprocess.run(["python3", str(head / "benchmarks/embedded/c/build.py"), "--tools-dir", str(args.tools_dir.resolve()),
                            "--output-dir", str(c_directory)], check=True, timeout=180)
            report["runtimes"] = runtime_commands(report["builds"], c_directory, report["ownBuilds"])
            report["cBuild"] = json.loads((c_directory / "build.json").read_text())
            subprocess.run(["python3", str(head / "benchmarks/embedded/c/check.py"),
                            "--runners-dir", str(c_directory)], check=True, timeout=180)
            cases = [(runtime, workload) for runtime in report["runtimes"] for workload in report["fixtures"]["workloads"]]
            randomizer = random.Random(report["seed"])
            for number in range(ROUNDS):
                randomizer.shuffle(cases)
                for runtime, workload in cases:
                    command = runtime["command"] + ["--wasm", str(head / "benchmarks/embedded" / workload["wasm"]),
                        "--backend", runtime["backend"], "--iterations", str(workload["iterations"]),
                        "--seed", str(workload["seed"]), "--expected", str(workload["expectedChecksum"]),
                        "--samples", str(SAMPLES), "--warmups", str(WARMUPS)]
                    report["activeCase"] = {"runtime": runtime["id"], "workload": workload["id"], "round": number, "command": command}
                    item = {**invoke(command, args.timeout), "runtime": runtime["id"], "workload": workload["id"],
                            "round": number, "wasmSha256": workload["wasmSha256"]}
                    validate_observation(item, runtime, workload)
                    report["observations"].append(item)
                    del report["activeCase"]
                print(f"Embedded phase round {number+1}/{ROUNDS} complete", flush=True)
            report["productionCLI"] = {"rounds": own_cli.ROUNDS, "runtimes": own_cli.configurations(report["ownBuilds"]), "observations": []}
            own_cli.run(head, report["ownBuilds"], report["fixtures"], args.timeout, report["seed"], report["productionCLI"])
            report["subsystems"] = {}
            subsystems.run(head, report["ownBuilds"], args.timeout, report["seed"], invoke, report["subsystems"])
            report["summary"] = summarize(report)
            report["status"] = "complete"
    except Exception as error:
        report["status"] = "failed"
        report["failure"] = f"{type(error).__name__}: {error}"
        raise
    finally:
        report["host"]["loadAverageAfter"] = os.getloadavg()
        report["finishedAt"] = datetime.now(timezone.utc).isoformat()
        (output / "results.json").write_text(json.dumps(report, indent=2) + "\n")
        (output / "report.md").write_text(markdown(report))


if __name__ == "__main__":
    main()
