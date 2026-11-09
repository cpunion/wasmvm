#!/usr/bin/env python3
"""Build and actually execute the pinned, independent P3 application reference."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
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
        return hashlib.file_digest(source,"sha256").hexdigest()


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir",type=Path,required=True)
    parser.add_argument("--rust-target-dir",type=Path,required=True)
    args=parser.parse_args()
    out,target=args.output_dir.resolve(),args.rust_target_dir.resolve()
    go,cargo,cc=shutil.which("go"),shutil.which("cargo"),shutil.which(os.environ.get("CC","clang"))
    if not all([go,cargo,cc]): parser.error("pinned Go, Rust and Clang are required")
    version=subprocess.check_output([go,"version"],text=True)
    clang=subprocess.check_output([cc,"--version"],text=True)
    rust=subprocess.check_output(["rustc","+1.98.0","-vV"],text=True)
    if "go1.27.3 " not in version or "clang version 22.1.8" not in clang or "release: 1.98.0\n" not in rust or "LLVM version: 22.1.8" not in rust:
        parser.error("expected Go1.27.3, Clang22.1.8 and Rust1.98.0/LLVM22.1.8")
    out.mkdir(parents=True,exist_ok=False)
    report=dict(status="running",platform=platform.platform(),go=version,clang=clang,rust=rust,
                goSHA256=digest(Path(go).resolve()),clangSHA256=digest(Path(cc).resolve()),steps=[])
    env=dict(os.environ,CC=cc,CGO_ENABLED="1",GOTOOLCHAIN="local",GOWORK="off",GOFLAGS="",CARGO_TARGET_DIR=str(target))

    def execute(command,label,environment=env,timeout=180):
        start=time.monotonic_ns()
        with (out/(label+".log")).open("x") as log:
            result=subprocess.run(command,cwd=HERE,env=environment,stdout=log,stderr=subprocess.STDOUT,timeout=timeout)
        report["steps"].append(dict(command=command,exitCode=result.returncode,wallNs=time.monotonic_ns()-start,log=label+".log"))
        if result.returncode: raise RuntimeError((out/(label+".log")).read_text())

    try:
        execute([cargo,"+1.98.0","rustc","--locked","--release","--lib","--manifest-path",str(HERE/"rust/Cargo.toml"),"--","--print","native-static-libs"],"rust-build",timeout=540)
        lines=(out/"rust-build.log").read_text().splitlines()
        flags=next(line.split("native-static-libs:",1)[1].strip() for line in lines if "native-static-libs:" in line)
        library=target/"release/libwasmvm_p3_reference.a"
        ldflags=shlex.join([str(library),*shlex.split(flags)])
        environment=dict(env,CGO_LDFLAGS=ldflags)
        report.update(nativeStaticFlags=flags,librarySHA256=digest(library),linkEnvironment=dict(CC=cc,CGO_LDFLAGS=ldflags),
            sourceHashes={str(p.relative_to(ROOT)):digest(p) for directory in [HERE,ROOT/"internal/bench/service"] for p in sorted(directory.rglob("*")) if p.is_file() and p.suffix in {".go",".rs",".h",".py",".toml",".lock",".mod",".sum"}})
        test=out/"p3-reference.test";coverage=out/"reference.cover"
        execute([go,"test","-c","-race","-cover","-tags=p3reference","-o",str(test),"."],"test-build",environment)
        execute([str(test),"-test.v","-test.count=1","-test.timeout=2m","-test.coverprofile="+str(coverage)],"test-execute",environment)
        cov=package_coverage(coverage.read_text());package="github.com/cpunion/wasmvm/benchmarks/p3io"
        if set(cov)!={package} or any(c*100<t*90 for c,t in cov.values()): raise RuntimeError(f"P3 Go adapter coverage below90%: {cov}")
        report["coverage"]={k:dict(covered=c,total=t,percent=c*100/t) for k,(c,t) in cov.items()}
        binary=out/"wasmtime-p3-reference"
        execute([go,"vet","-tags=p3reference","./..."],"vet",environment)
        execute([go,"build","-tags=p3reference","-o",str(binary),"."],"build",environment)
        deps=subprocess.check_output([go,"list","-tags=p3reference","-deps","-f","{{.ImportPath}}","."],cwd=HERE,env=environment,text=True).splitlines()
        owned=[p for p in deps if p.startswith("github.com/cpunion/wasmvm/") and not p.startswith("github.com/cpunion/wasmvm/benchmarks/")]
        if owned!=["github.com/cpunion/wasmvm/internal/bench/service"]: raise RuntimeError(f"reference imports owned engine: {owned}")
        report.update(binary=str(binary),binarySHA256=digest(binary),ownedDependencies=owned,status="pass")
        print("P3 reference compiled and executed; >=90% Go adapter coverage; no owned VM dependency",flush=True)
    except Exception as error:
        report.update(status="failed",error=str(error));raise
    finally:
        (out/"checks.json").write_text(json.dumps(report,indent=2)+"\n")


if __name__=="__main__": main()
