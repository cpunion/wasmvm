#!/usr/bin/env python3
"""Execute the compiler's actual CLI binary against an authored pinned fixture."""
import json
from pathlib import Path
import subprocess
import sys


def main():
    executable = str(Path(sys.argv[1]).resolve())
    root = Path(__file__).resolve().parent.parent
    fixture = str(root / "benchmarks/embedded/fixtures/call-control.wasm")
    commands = {
        "validate": ["validate", fixture],
        "inspect": ["inspect", fixture],
        "run": ["run", "--export", "run", "--fuel", "20", fixture,
                "i32:305419896", "i32:1"],
    }
    for name, args in commands.items():
        result = subprocess.run([executable, *args], check=True,
                                capture_output=True, text=True, timeout=30)
        report = json.loads(result.stdout)
        assert report["backend"] == "interpreter", report
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
    rejected = subprocess.run(
        [executable, "validate", "--backend", "native", fixture],
        capture_output=True, text=True, timeout=30)
    assert rejected.returncode == 1 and not rejected.stdout, rejected
    assert "unsupported" in rejected.stderr, rejected
    print("CLI validate/inspect/run passed; no native fallback")


if __name__ == "__main__":
    main()
