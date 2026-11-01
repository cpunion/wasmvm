"""Descriptive, exact-revision survey of the complete existing Core portfolio.

This is a gap inventory, not a controlled performance acceptance gate. Reuse the
phase runners and serial warm-call interleaving; never subtract setup or controls.
"""
import argparse
from datetime import datetime, timezone
import hashlib
import json
import os
from pathlib import Path
import platform
import random
import statistics
import subprocess

import ci
import interleaved
import own_build
import phase_ci


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def fixtures(root):
    groups = [("core", root / "benchmarks/embedded")]
    groups += [(p.name.removesuffix("-fixtures"), p)
               for p in sorted((root / "benchmarks/native").glob("*-fixtures"))]
    cases = []
    for family, directory in groups:
        manifest_path = directory / ("fixtures/manifest.json" if family == "core" else "manifest.json")
        manifest = json.loads(manifest_path.read_text())
        for item in manifest["workloads"]:
            for suffix, key in (("wasm", "wasmSha256"), ("wat", "watSha256")):
                path = directory / item.get(suffix, item["id"] + "." + suffix)
                expected = item.get(key, item.get("sourceSha256") if suffix == "wat" else None)
                if expected is None or digest(path) != expected:
                    raise ValueError("fixture hash mismatch: " + str(path))
            cases.append(dict(item, id=family + "/" + item["id"], family=family,
                              path=str(directory / item["wasm"]), producer=manifest["producer"],
                              unit="ns/call" if item["id"] in ("call-control", "growth-first-call")
                              else "ns/source iteration"))
    return cases


def build(root, output, sdks):
    revision = ci.capture(["git", "-C", str(root), "rev-parse", "HEAD"])
    toolchains = own_build.toolchains()
    builds, runtimes = {}, {}
    for compiler, toolchain in toolchains.items():
        target = output / compiler
        identity = own_build.prepare(root, root, revision, target)
        generated = target / "internal/bench/phase/provenance_gen.go"
        generated.write_text("package phase\nfunc init(){sourceRevision=" + json.dumps(revision)
                             + ";compilerRevision=" + json.dumps(toolchain["revision"]) + "}\n")
        binary = target / "wasmvm-phase"
        environment = dict(os.environ, CGO_ENABLED="1", GOWORK="off", GOFLAGS="", GOTOOLCHAIN="local")
        command = [toolchain["command"], "build", *toolchain["buildFlags"], "-o", str(binary),
                   "./benchmarks/embedded/wasmvm"]
        subprocess.run(command, cwd=target, env=environment, check=True, timeout=600)
        if own_build.hashes(target, own_build.ENGINE_PATHS, own_build.HARNESS) != identity["sourceHashes"]:
            raise ValueError("build changed archived engine sources")
        builds[compiler] = dict(identity, toolchain=toolchain, command=command, binary=str(binary),
                                binarySha256=digest(binary), harnessSha256=own_build.hashes(target, own_build.HARNESS))
        for backend in ("interpreter", "native"):
            name = "wasmvm-" + compiler + "-" + backend
            runtimes[name] = dict(command=[str(binary), "--mode", "api"], backend=backend, engine="wasmvm",
                                  compiler="gc" if compiler == "go" else "llgo", compilerRevision=toolchain["revision"],
                                  sourceRevision=revision, mode="api", goos=toolchain["goos"], goarch=toolchain["goarch"])
        print(compiler, "exact-revision phase runner built", flush=True)
    target = output / "wazero"
    own_build.archive(root, revision, target, ("benchmarks/embedded/wazero",))
    target = target / "benchmarks/embedded/wazero"
    binary = target / "wazero-phase"
    subprocess.run(["go", "build", "-trimpath", "-o", str(binary), "."], cwd=target, check=True, timeout=600)
    builds["wazero"] = dict(binary=str(binary), binarySha256=digest(binary),
                             dependency=json.loads(ci.capture(["go", "list", "-m", "-json", ci.MODULE], cwd=target)))
    if builds["wazero"]["dependency"]["Version"] != "v1.12.0" or builds["wazero"]["dependency"].get("Replace"):
        raise ValueError("wazero dependency differs from the survey pin")
    for backend in ("interpreter", "compiler"):
        runtimes["wazero-" + backend] = dict(command=[str(binary)], backend=backend, version="1.12.0")
    for name, sdk in sdks.items():
        binary = output / (name + "-phase")
        source = root / "benchmarks/embedded/c" / (name + ".c")
        command = ["cc", "-std=c11", "-D_POSIX_C_SOURCE=200809L", "-O2", "-Wall", "-Wextra", "-Werror",
                   "-I", str(sdk / "include"), str(source), "-o", str(binary),
                   "-L", str(sdk / "lib"), "-Wl,-rpath," + str(sdk / "lib"), "-l" + name]
        subprocess.run(command, check=True, timeout=180)
        suffix = ".dylib" if platform.system() == "Darwin" else ".so"
        builds[name] = dict(command=command, binary=str(binary), binarySha256=digest(binary),
                            sourceSha256=digest(source), sdkLibrarySha256=digest(sdk / "lib" / ("lib" + name + suffix)))
        runtimes[name + ("-interpreter" if name == "wasmedge" else "-cranelift")] = dict(
            command=[str(binary)], backend="interpreter" if name == "wasmedge" else "compiler",
            version={"wasmtime": "48.0.1", "wasmer": "7.3.0", "wasmedge": "0.17.1"}[name])
    return dict(revision=revision, builds=builds, runtimes=runtimes)


def summarize(report):
    rows = []
    for case in report["cases"]:
        for name in report["runtimes"]:
            observations = [x for x in case["observations"] if x["runtime"] == name]
            if not observations:
                continue
            denominator = 1 if case["fixture"]["unit"] == "ns/call" else case["fixture"]["iterations"]
            rows.append(dict(runtime=name, workload=case["fixture"]["id"], family=case["fixture"]["family"],
                             unit=case["fixture"]["unit"], processes=len(observations),
                             warm=statistics.median(statistics.median(x["record"]["warmCallNs"])
                                                    for x in observations) / denominator,
                             processReadyNs=statistics.median(x["processReadyNs"] for x in observations),
                             phasesNs={key: statistics.median(x["record"]["phasesNs"][key] for x in observations)
                                       for key in observations[0]["record"]["phasesNs"]}))
    return rows


def markdown(report):
    lines = ["## Owned runtime performance survey", "",
             "**Descriptive only; performance acceptance remains open.** "
             "Three fresh process blocks, seven serially interleaved warm calls, two warmups. "
             "Setup, first calls and cleanup remain separate in the raw artifact. "
             "Each call checks its checksum; no control/setup time is subtracted.", "",
             "Engine source: `" + report["sourceRevision"] + "`.", ""]
    if report["status"] != "complete":
        return "\n".join(lines + ["**Incomplete:** " + report.get("failure", report["status"]), ""])
    names = list(report["runtimes"])
    rows = {(r["runtime"], r["workload"]): r for r in summarize(report)}
    lines += ["| Workload | Unit | " + " | ".join(names) + " |",
              "| --- | --- | " + " | ".join("---:" for _ in names) + " |"]
    for case in report["cases"]:
        f = case["fixture"]
        values = [f"{rows[(name, f['id'])]['warm']:.3f}" if (name, f["id"]) in rows else "unsupported"
                  for name in names]
        lines.append("| " + f["id"] + " | " + f["unit"] + " | " + " | ".join(values) + " |")
    return "\n".join(lines + ["", "Unknown/missing features are not successful zero-cost measurements. "
                            "Short controls are clock-resolution limited. No calibrated parity verdict is inferred.", ""])


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    for name in ("wasmtime", "wasmer", "wasmedge"):
        parser.add_argument("--" + name + "-sdk", type=Path, required=True)
    parser.add_argument("--build-only", action="store_true")
    parser.add_argument("--reuse-build", action="store_true")
    args = parser.parse_args()
    root = Path(__file__).resolve().parent.parent
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=True)
    build_file = output / "build.json"
    if args.reuse_build:
        built = json.loads(build_file.read_text())
        for value in built["builds"].values():
            if digest(Path(value["binary"])) != value["binarySha256"]:
                raise ValueError("built runner hash mismatch")
    else:
        built = build(root, output / "build", {name: getattr(args, name + "_sdk").resolve()
                                             for name in ("wasmtime", "wasmer", "wasmedge")})
        build_file.write_text(json.dumps(built, indent=2) + "\n")
    if args.build_only:
        return
    report = dict(schemaVersion=1, status="running", acceptance="not evaluated; descriptive shared-host survey",
                  sourceRevision=built["revision"], runtimes=built["runtimes"],
                  startedAt=datetime.now(timezone.utc).isoformat(), rounds=3, samples=7, warmups=2, seed=20261101,
                  host=dict(platform=platform.platform(), machine=platform.machine(), loadBefore=os.getloadavg(),
                            cpu=ci.capture(["sysctl", "-n", "machdep.cpu.brand_string"])
                            if platform.system() == "Darwin" else ci.capture(["lscpu"])),
                  cases=[dict(fixture=f, observations=[], sampleOrders=[]) for f in fixtures(root)])
    rng = random.Random(report["seed"])
    try:
        for block in range(report["rounds"]):
            cases = list(report["cases"])
            rng.shuffle(cases)
            for case in cases:
                f = case["fixture"]
                commands = {name: r["command"] + ["--backend", r["backend"], "--wasm", f["path"],
                            "--iterations", str(f["iterations"]), "--seed", str(f["seed"]),
                            "--expected", str(f["expectedChecksum"]), "--samples", "7", "--warmups", "2"]
                            for name, r in report["runtimes"].items()}
                rows, order = interleaved.invoke(commands, 180, rng)
                case["sampleOrders"].append(order)
                for name, row in rows.items():
                    row = dict(row, wasmSha256=f["wasmSha256"])
                    phase_ci.validate_observation(row, report["runtimes"][name], f)
                    case["observations"].append(dict(row, runtime=name, block=block))
                (output / "results.json").write_text(json.dumps(report, indent=2) + "\n")
                print("block", block + 1, f["id"], "checked", flush=True)
        report["status"] = "complete"
    except BaseException as error:
        report["status"], report["failure"] = "failed", str(error)
        raise
    finally:
        report["host"]["loadAfter"] = os.getloadavg()
        report["finishedAt"] = datetime.now(timezone.utc).isoformat()
        (output / "results.json").write_text(json.dumps(report, indent=2) + "\n")
        (output / "summary.json").write_text(json.dumps(summarize(report), indent=2) + "\n")
        (output / "report.md").write_text(markdown(report))


if __name__ == "__main__":
    main()
