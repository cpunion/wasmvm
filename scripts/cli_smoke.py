#!/usr/bin/env python3
"""Execute a built CLI through Core and WASI command entry points."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("executable")
    parser.add_argument("--native", choices=("available", "unavailable"), required=True)
    parser.add_argument("--report", type=Path)
    options = parser.parse_args()
    executable = str(Path(options.executable).resolve())
    root = Path(__file__).resolve().parent.parent
    fixture = str(root / "benchmarks/embedded/fixtures/call-control.wasm")
    records = []
    commands = {
        "validate": ["validate", fixture],
        "inspect": ["inspect", fixture],
        "run": ["run", "--export", "run", "--fuel", "20", fixture,
                "i32:305419896", "i32:1"],
    }
    for backend in ("interpreter", "native"):
        for name, args in commands.items():
            # Omit the interpreter flag to verify the default selection too.
            selected = args if backend == "interpreter" else [args[0], "--backend", backend, *args[1:]]
            result = subprocess.run([executable, *selected],
                                    capture_output=True, text=True, timeout=30)
            records.append(dict(case=name, backend=backend, command=selected, exitCode=result.returncode,
                                stdout=result.stdout, stderr=result.stderr, fixtureSHA256=hashlib.sha256(Path(fixture).read_bytes()).hexdigest()))
            if backend == "native" and options.native == "unavailable":
                assert result.returncode == 1 and not result.stdout, result
                assert "unsupported" in result.stderr, result
                continue
            assert result.returncode == 0 and not result.stderr, result
            report = json.loads(result.stdout)
            assert report["backend"] == backend, report
            assert report["profile"] == "core2", report
            if name == "validate":
                assert report["valid"] is True, report
            elif name == "inspect":
                assert report["imports"] == [] and not report["hasStart"], report
                assert report["exports"][0]["params"] == ["i32", "i32"], report
            else:
                assert report["results"] == [{"type": "i32", "bits": "0x12345678",
                                               "value": "305419896"}], report
                assert 0 <= report["fuelRemaining"] < 20, report
    with tempfile.TemporaryDirectory(prefix="wasmvm-cli-") as directory:
        stdin = Path(directory) / "stdin"
        stdin.write_text("input")
        cases = [
            ("p1-stdio", "internal/cli/testdata/wasi-stdio.wasm",
             ["--wasi", "--stdin", str(stdin)], [], 0, "input", False),
            ("p2-command", "pkg/wasi/preview2/testdata/world-c-command.wasm",
             ["--wasi-preview", "2", "--fuel", "10000000", "--read-only-dir", "/data=" + directory],
             ["world"], 0, "c-command:ok\n", False),
            ("p3-command", "pkg/wasi/preview3/testdata/world-c-command.wasm",
             ["--wasi-preview", "3", "--fuel", "10000000", "--read-only-dir", "/data=" + directory],
             ["world"], 0, "c-p3-command:ok\n", False),
            ("p3-failure", "pkg/wasi/preview3/testdata/command-failure.wasm", ["--wasi-preview", "3"], [], 1, "", False),
            ("p3-exit", "pkg/wasi/preview3/testdata/command-exit.wasm", ["--wasi-preview", "3"], [], 255, "", False),
            ("p3-trap", "pkg/wasi/preview3/testdata/command-trap.wasm", ["--wasi-preview", "3"], [], 1, "", True),
        ]
        for backend in (("interpreter", "native") if options.native == "available" else ("interpreter",)):
            for name, path, flags, argv, code, output, diagnostic in cases:
                command = ["run", "--backend", backend, *flags, str(root / path), *argv]
                result = subprocess.run([executable, *command], capture_output=True, text=True, timeout=30)
                assert result.returncode == code and result.stdout == output and bool(result.stderr) == diagnostic, (name, backend, result)
                records.append(dict(case=name, backend=backend, command=command, exitCode=result.returncode,
                                    stdout=result.stdout, stderr=result.stderr, fixtureSHA256=hashlib.sha256((root / path).read_bytes()).hexdigest()))
    report = dict(passed=True, native=options.native, executableSHA256=hashlib.sha256(Path(executable).read_bytes()).hexdigest(),
                  scope="Actual CLI Core/P1/P2/P3 command smoke checks; full semantics remain separate", cases=records)
    if options.report:
        options.report.write_text(json.dumps(report, indent=2) + "\n")
    print(f"CLI Core/P1/P2/P3 checks passed ({len(records)} processes); native {options.native}; no fallback")


if __name__ == "__main__":
    main()
