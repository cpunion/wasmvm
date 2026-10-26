"""Build a common measurement harness over exact, isolated wasmvm revisions."""
import hashlib
import json
import os
from pathlib import Path
import platform
import shutil
import subprocess
import tarfile

import ci

MODULE = "github.com/cpunion/wasmvm"
LLGO_REVISION = "8b1389c627c77a02e56af03f08d9cdc19edffe77"
GO_VERSION = "go1.27.3"
HARNESS = ("internal/bench/phase", "benchmarks/embedded/wasmvm")
ENGINE_PATHS = ("go.mod", "go.sum", "internal", "pkg", "cmd")


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def archive(checkout, revision, destination, paths=()):
    """Archive committed files only; untracked Go files cannot enter a build."""
    destination.mkdir(parents=True, exist_ok=True)
    archive_file = destination.parent / (destination.name + "-archive.tar")
    with archive_file.open("wb") as stream:
        subprocess.run(["git", "-C", str(checkout), "archive", "--format=tar", revision, *paths],
                       stdout=stream, check=True, timeout=60)
    with tarfile.open(archive_file) as source:
        source.extractall(destination, filter="data")
    archive_file.unlink()


def hashes(root, paths, exclude=()):
    result = {}
    for relative in paths:
        path = root / relative
        candidates = [path] if path.is_file() else sorted(path.rglob("*"))
        for candidate in candidates:
            name = str(candidate.relative_to(root))
            if any(name == item or name.startswith(item + "/") for item in exclude):
                continue
            if candidate.is_symlink():
                raise ValueError("symlink in measured source: " + name)
            if candidate.is_file():
                result[name] = digest(candidate)
    return result


def source_digest(sources):
    return hashlib.sha256(json.dumps(sources, sort_keys=True, separators=(",", ":")).encode()).hexdigest()


def toolchains():
    go = shutil.which("go")
    llgo = os.environ.get("WASMVM_LLGO") or shutil.which("llgo")
    if not go or not llgo:
        raise ValueError("both Go and pinned LLGo are required for own-engine measurements")
    version = ci.capture([go, "env", "GOVERSION"])
    if version != GO_VERSION:
        raise ValueError("own-engine Go toolchain differs from pin: " + version)
    if os.environ.get("WASMVM_LLGO_REVISION") != LLGO_REVISION:
        raise ValueError("verified WASMVM_LLGO_REVISION must match the pinned setup revision")
    common = {"goos": platform.system().lower(), "goarch": {"x86_64": "amd64", "aarch64": "arm64", "arm64": "arm64"}.get(platform.machine())}
    if common["goos"] not in ("linux", "darwin") or common["goarch"] is None:
        raise ValueError("own-engine benchmark host is outside the accepted platform matrix")
    result = {
        "go": {**common, "command": str(Path(go).resolve()), "revision": version,
               "version": ci.capture([go, "version"]), "binarySha256": digest(Path(go))},
        "llgo": {**common, "command": str(Path(llgo).resolve()), "revision": LLGO_REVISION,
                 "version": ci.capture([llgo, "version"]), "binarySha256": digest(Path(llgo)),
                 "environment": ci.capture([llgo, "env", "LLGO_ROOT", "GOROOT"])},
    }
    return result


def prepare(head, checkout, revision, target):
    # No tracked implementation change may be silently attributed to a commit.
    ci.verify_checkout(checkout, revision)
    subprocess.run(["git", "-C", str(checkout), "diff", "--quiet", revision, "--", "."],
                   check=True, timeout=60)
    archive(checkout, revision, target)
    sources = hashes(target, ENGINE_PATHS, HARNESS)
    if not (target / "pkg/vm/engine.go").is_file():
        return {"status": "baseline-unavailable", "sourceRevision": revision,
                "reason": "the selected revision predates the owned Engine/Store API",
                "sourceHashes": sources, "sourceDigest": source_digest(sources)}
    for path in HARNESS:
        if (target / path).exists():
            shutil.rmtree(target / path)
    # Only these two instrumentation directories come from head. No core or
    # public VM implementation file is ever copied over the baseline.
    archive(head, ci.capture(["git", "-C", str(head), "rev-parse", "HEAD"]), target, HARNESS)
    after = hashes(target, ENGINE_PATHS, HARNESS)
    if after != sources:
        raise ValueError("harness overlay changed engine sources")
    return {"status": "available", "sourceRevision": revision,
            "sourceHashes": sources, "sourceDigest": source_digest(sources),
            "harnessHashes": hashes(target, HARNESS)}


def build(head, base, temporary, base_sha, head_sha):
    selected = toolchains()
    result = {}
    for name, checkout, revision in (("base", base, base_sha), ("head", head, head_sha)):
        for compiler in ("go", "llgo"):
            target = temporary / f"own-{name}-{compiler}"
            identity = prepare(head, checkout, revision, target)
            if identity["status"] != "available":
                if name != "base":
                    raise ValueError("head does not contain an owned engine")
                result[f"{name}-{compiler}"] = identity
                continue
            toolchain = selected[compiler]
            generated = target / "internal/bench/phase/provenance_gen.go"
            generated.write_text("package phase\nfunc init() { sourceRevision = " + json.dumps(revision)
                                 + "; compilerRevision = " + json.dumps(toolchain["revision"]) + " }\n")
            environment = dict(os.environ, GOWORK="off", GOFLAGS="", GOTOOLCHAIN="local")
            if compiler == "go":
                environment["CGO_ENABLED"] = "0"
            executable = target / "wasmvm-phase"
            command = [toolchain["command"], "build", "-o", str(executable), "./benchmarks/embedded/wasmvm"]
            subprocess.run(command, cwd=target, env=environment, check=True, timeout=600)
            dependencies = ci.capture([selected["go"]["command"], "list", "-deps", "./benchmarks/embedded/wasmvm"],
                                      cwd=target, env=environment).splitlines()
            if MODULE + "/pkg/vm" not in dependencies or any(
                    item.startswith(("github.com/tetratelabs/wazero", "github.com/bytecodealliance/wasmtime",
                                     "github.com/wasmerio/wasmer", "github.com/wasmedge")) for item in dependencies):
                raise ValueError("own-engine executable contains an unexpected engine dependency")
            cli = {"status": "baseline-unavailable", "reason": "the selected revision predates cmd/wasmvm"}
            if (target / "cmd/wasmvm/main.go").is_file():
                cli_binary = target / "wasmvm-cli"
                cli_command = [toolchain["command"], "build", "-o", str(cli_binary), "./cmd/wasmvm"]
                subprocess.run(cli_command, cwd=target, env=environment, check=True, timeout=600)
                cli = {"status": "available", "binary": str(cli_binary),
                       "binarySha256": digest(cli_binary), "buildCommand": cli_command}
            elif name == "head":
                raise ValueError("head does not contain the production CLI")
            if hashes(target, ENGINE_PATHS, HARNESS) != identity["sourceHashes"]:
                raise ValueError("build modified archived engine inputs")
            result[f"{name}-{compiler}"] = {**identity, "binary": str(executable),
                "binarySha256": digest(executable), "compiler": compiler, "toolchain": toolchain,
                "buildCommand": command, "buildEnvironment": {k: environment.get(k) for k in
                    ("GOWORK", "GOFLAGS", "GOTOOLCHAIN", "CGO_ENABLED", "LLGO_ROOT")},
                "generatedProvenanceSha256": digest(generated), "dependencies": dependencies, "cli": cli}
    return result
