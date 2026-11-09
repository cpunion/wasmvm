#!/usr/bin/env python3
"""Build and execute the optional pinned SDK service references; no timing claims."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shlex
import shutil
import subprocess
import sys
import time

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from check_coverage import package_coverage


def digest(path):
    with path.open("rb") as source:
        return hashlib.file_digest(source, "sha256").hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--tools-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()
    tools, out = args.tools_dir.resolve(), args.output_dir.resolve()
    installation = json.loads((tools / "installation.json").read_text())
    if not installation.get("sdkEnabled"):
        parser.error("install_tools.py --sdk is required")
    entries = {tool["id"]: tool for tool in installation["tools"]}
    cc = shutil.which(os.environ.get("CC", "clang"))
    go = shutil.which("go")
    if not go or not cc:
        parser.error("pinned Go and Clang are required")
    go_version = subprocess.check_output([go, "version"], text=True)
    clang_version = subprocess.check_output([cc, "--version"], text=True)
    if "go1.27.3 " not in go_version or not re.search(r"\bclang version 22\.1\.8(?:\s|$)", clang_version):
        parser.error("expected Go 1.27.3 and Clang 22.1.8")
    out.mkdir(parents=True, exist_ok=False)
    report = dict(status="running", go=go_version, clang=clang_version,
                  goSHA256=digest(Path(go).resolve()), clangSHA256=digest(Path(cc).resolve()),
                  installation=installation, runners=[])
    paths = [HERE, ROOT / "internal/bench/service", ROOT / "benchmarks/applications",
             ROOT / "benchmarks/embedded/wazero/testdata"]
    report["sourceHashes"] = {str(p.relative_to(ROOT)):digest(p) for folder in paths
                              for p in sorted(folder.rglob("*"))
                              if p.is_file() and (p.suffix in {".go", ".c", ".h", ".wasm", ".wat", ".json", ".py"} or p.name in {"go.mod", "go.sum"})}
    env = dict(os.environ, CC=cc, CGO_ENABLED="1", GOTOOLCHAIN="local", GOWORK="off", GOFLAGS="")

    def execute(command, label, environment, cwd=HERE):
        start = time.monotonic_ns()
        with (out / (label + ".log")).open("x") as log:
            process = subprocess.run(command, cwd=cwd, env=environment, stdout=log,
                                     stderr=subprocess.STDOUT, timeout=180)
        record = dict(command=command, exitCode=process.returncode,
                      wallNs=time.monotonic_ns()-start, log=label+".log")
        if process.returncode:
            raise RuntimeError(f"{label} failed: {(out / (label+'.log')).read_text()}")
        return record

    try:
        for name, tool_id in [("wasmtime", "wasmtime-c-api"), ("wasmer", "wasmer"), ("wasmedge", "wasmedge")]:
            tool = entries[tool_id]
            prefix = tools / tool_id
            includes = [prefix / p for p in tool["includeDirectories"]]
            libraries = [prefix / p for p in tool["libraryDirectories"]]
            cflags = ["-I"+str(p) for p in includes]
            ldflags = [v for p in libraries for v in ("-L"+str(p), "-Wl,-rpath,"+str(p))]+["-l"+name]
            environment = dict(env, CGO_CFLAGS=shlex.join(cflags), CGO_LDFLAGS=shlex.join(ldflags))
            test = out / (name+".test")
            coverage = out / (name+".cover")
            row = dict(runtime=name, sdk=tool, environment={k:environment[k] for k in ["CC", "CGO_CFLAGS", "CGO_LDFLAGS"]}, steps=[])
            report["runners"].append(row)
            row["steps"].append(execute([go, "test", "-c", "-race", "-cover", "-tags="+name, "-o", str(test), "."], name+"-test-build", environment))
            row["steps"].append(execute([str(test), "-test.v", "-test.count=1", "-test.timeout=2m", "-test.coverprofile="+str(coverage)], name+"-test-execute", environment))
            cov = package_coverage(coverage.read_text())
            package = "github.com/cpunion/wasmvm/benchmarks/service"
            if set(cov) != {package} or any(c*100 < t*90 for c,t in cov.values()):
                raise RuntimeError(f"reference Go statement coverage below 90%: {cov}")
            row["coverage"] = {k:dict(covered=c,total=t,percent=c*100/t) for k,(c,t) in cov.items()}
            binary = out / (name+"-service")
            row["steps"].append(execute([go, "build", "-tags="+name, "-o", str(binary), "."], name+"-build", environment))
            row.update(binary=str(binary), binarySHA256=digest(binary),
                       sdkFiles={str(p.relative_to(prefix)):digest(p) for folder in includes+libraries for p in sorted(folder.rglob("*")) if p.is_file()})
            print(name, "compiled and executed; >=90% Go statement coverage", flush=True)
        report["status"] = "pass"
    except Exception as error:
        report.update(status="failed", error=str(error))
        raise
    finally:
        (out / "checks.json").write_text(json.dumps(report, indent=2)+"\n")


if __name__ == "__main__":
    main()
