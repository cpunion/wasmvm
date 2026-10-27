"""Run only the native feature gate on a separate host; reuse all phase tools."""
import argparse
import json
import os
from pathlib import Path
import platform
import subprocess
import tempfile

import ci
import native_compare
import native_simd
import own_build
import phase_ci


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("head-dir", "base-dir", "tools-dir", "output-dir"):
        parser.add_argument("--"+name, type=Path, required=True)
    parser.add_argument("--scope", choices=("integer", "fixed-simd"), default="integer")
    args = parser.parse_args()
    head, base, output = args.head_dir.resolve(), args.base_dir.resolve(), args.output_dir.resolve()
    output.mkdir(parents=True, exist_ok=True)
    report = {"status": "failed", "host": {"platform": platform.platform(), "loadBefore": os.getloadavg()}}
    try:
        toolchains = own_build.toolchains()
        with tempfile.TemporaryDirectory(prefix="wasmvm-native-ci-") as directory:
            temporary = Path(directory)
            builds = {}
            for ref, checkout in (("base", base), ("head", head)):
                revision = ci.capture(["git", "-C", str(checkout), "rev-parse", "HEAD"])
                report[ref+"Sha"] = revision
                for compiler in ("go", "llgo"):
                    target = temporary / (ref+"-"+compiler)
                    identity = own_build.prepare(head, checkout, revision, target)
                    builds[ref+"-"+compiler] = {**identity, "binary": str(target / "wasmvm-native-phase"), "toolchain": toolchains[compiler]}
            refs = phase_ci.build_candidates(head, base, temporary / "references")
            c_dir = temporary / "c"
            subprocess.run(["python3", str(head / "benchmarks/embedded/c/build.py"), "--tools-dir", str(args.tools_dir.resolve()),
                            "--output-dir", str(c_dir), "--runtimes", "wasmtime", "wasmer"], check=True)
            subprocess.run(["python3", str(head / "benchmarks/embedded/c/check.py"), "--runners-dir", str(c_dir),
                            "--runtimes", "wasmtime", "wasmer"], check=True)
            report["references"] = refs
            report["cBuild"] = json.loads((c_dir / "build.json").read_text())
            runtimes = [{"id": "head-compiler", "command": [refs["head"]["binary"]]}]
            runtimes += [{"id": name+"-cranelift", "command": [str(c_dir / (name+"-phase"))]} for name in ("wasmtime", "wasmer")]
            if args.scope == "fixed-simd":
                builds = own_build.build(head, base, temporary / "api", report["baseSha"], report["headSha"], native=True)
                report["ownBuilds"] = builds
                report["fixedSIMD"] = native_simd.run(head, builds, runtimes)
                report["status"] = report["fixedSIMD"]["status"]
            else:
                fixture = next(w for w in phase_ci.load_workloads(head)["workloads"] if w["id"] == "integer-sustained")
                report["nativeFeature"] = native_compare.run(head, builds, runtimes, fixture, phase_ci.invoke, 180)
                report["status"] = report["nativeFeature"]["status"]
    except Exception as error:
        report["failure"] = f"{type(error).__name__}: {error}"
        raise
    finally:
        report["host"]["loadAfter"] = os.getloadavg()
        (output / "results.json").write_text(json.dumps(report, indent=2)+"\n")
        text = native_compare.markdown(report["nativeFeature"]) if "nativeFeature" in report else "Native feature measurements unavailable; see logs and raw report.\n"
        if "fixedSIMD" in report:
            text = native_simd.markdown(report["fixedSIMD"])
        (output / "report.md").write_text(text)
    if report["status"] not in ("pass", "diagnostic"):
        raise SystemExit("native feature gate blocks merging; see report")


if __name__ == "__main__":
    main()
