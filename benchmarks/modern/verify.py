"""Reproduce owned modern fixtures and verify independent integer oracles."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile


def checksum(name, seed, count):
    mask = 0x7fffffff if name == "gc/i31-roundtrip" else 0xffffffff
    if name.startswith("gc/"):
        x, cells = seed, [seed, seed ^ 0xdeadbeef]
        for i in range(count):
            updated = (x * 1664525 + 1013904223) & mask
            if name == "gc/struct-indirect-access":
                cells[x & 1] = updated
                x = cells[(x >> 1) & 1] ^ i
            else:
                x = updated
        return x
    if name == "threads/notify-empty":
        return seed
    multiplier = 2 if name == "threads/wait-timeout" else 1
    return (seed + multiplier * count) & mask


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--producer", default="wasm-tools")
    args = parser.parse_args()
    root = Path(__file__).resolve().parent
    manifest = json.loads((root / "manifest.json").read_text())
    version = subprocess.check_output([args.producer, "--version"], text=True).strip()
    with tempfile.TemporaryDirectory(prefix="wasmvm-modern-reproduce-") as directory:
        for f in manifest["workloads"]:
            for key in ("wat", "wasm"):
                if hashlib.sha256((root / f[key]).read_bytes()).hexdigest() != f[key + "Sha256"]:
                    raise ValueError("fixture digest mismatch: " + f["id"])
            for seed, count, expected in ((f["seed"], f["iterations"], f["expectedChecksum"]),
                                          (0xffffffff, 17, f["alternateExpectedChecksum"])):
                if checksum(f["id"], seed, count) != expected:
                    raise ValueError("independent oracle mismatch: " + f["id"])
            binary = Path(directory) / f["wasm"]
            subprocess.run([args.producer, "parse", str(root / f["wat"]), "-o", str(binary)], check=True)
            if binary.read_bytes() != (root / f["wasm"]).read_bytes():
                raise ValueError("producer output changed: " + f["id"])
    print(version + ": 10 byte-identical fixtures; 20 independent oracle vectors verified")


if __name__ == "__main__":
    main()
