#!/usr/bin/env python3
"""Execute the compiler's actual CLI binary against an authored pinned fixture."""
import argparse
import json
from pathlib import Path
import subprocess


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("executable")
    parser.add_argument("--native", choices=("available", "unavailable"), required=True)
    options = parser.parse_args()
    executable = str(Path(options.executable).resolve())
    root = Path(__file__).resolve().parent.parent
    fixture = str(root / "benchmarks/embedded/fixtures/call-control.wasm")
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
    print(f"CLI validate/inspect/run passed; native {options.native}; no fallback")


if __name__ == "__main__":
    main()
