#!/usr/bin/env python3
"""Build direct C API phase runners against the checksum-pinned SDK install."""

import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess


HERE = Path(__file__).resolve().parent


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--tools-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--runtimes", nargs="+", choices=("wasmtime", "wasmer", "wasmedge"), default=("wasmtime", "wasmer", "wasmedge"))
    args = parser.parse_args()
    tools_dir, output_dir = args.tools_dir.resolve(), args.output_dir.resolve()
    installation = json.loads((tools_dir / "installation.json").read_text())
    if not installation.get("sdkEnabled"):
        parser.error("install_tools.py --sdk is required")
    tools = {tool["id"]: tool for tool in installation["tools"]}
    compiler = shutil.which("cc")
    if not compiler:
        parser.error("a C11 compiler named cc is required")
    output_dir.mkdir(parents=True, exist_ok=True)
    records = []
    for runtime, tool_id, library in [("wasmtime", "wasmtime-c-api", "wasmtime"),
                                       ("wasmer", "wasmer", "wasmer"),
                                       ("wasmedge", "wasmedge", "wasmedge")]:
        if runtime not in args.runtimes:
            continue
        tool = tools[tool_id]
        prefix = tools_dir / tool_id
        executable = output_dir / (runtime + "-phase")
        command = [compiler, "-std=c11", "-D_POSIX_C_SOURCE=200809L", "-O2", "-Wall", "-Wextra", "-Werror"]
        for include in tool["includeDirectories"]:
            command.extend(["-I", str(prefix / include)])
        command.extend([str(HERE / (runtime + ".c")), "-o", str(executable)])
        for library_dir in tool["libraryDirectories"]:
            directory = str(prefix / library_dir)
            command.extend(["-L", directory, "-Wl,-rpath," + directory])
        command.append("-l" + library)
        subprocess.run(command, check=True)
        with executable.open("rb") as source:
            binary_hash = hashlib.file_digest(source, "sha256").hexdigest()
        records.append({"runtime": runtime, "sdk": tool, "command": command,
                        "executable": str(executable),
                        "sha256": binary_hash})
    sources = {path.name: hashlib.sha256(path.read_bytes()).hexdigest()
               for path in sorted(HERE.iterdir()) if path.suffix in {".c", ".h", ".py"}}
    metadata = {"schemaVersion": 1, "compiler": subprocess.check_output([compiler, "--version"], text=True),
                "sources": sources, "runners": records}
    (output_dir / "build.json").write_text(json.dumps(metadata, indent=2) + "\n")


if __name__ == "__main__":
    main()
