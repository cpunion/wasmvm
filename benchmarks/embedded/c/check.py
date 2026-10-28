#!/usr/bin/env python3
"""Exercise embedded C runners for correctness; do not publish these timings."""

import argparse
import json
import os
from pathlib import Path
import subprocess
import tempfile


EMBEDDED = Path(__file__).resolve().parents[1]
BACKENDS = {"wasmtime": "compiler", "wasmer": "compiler", "wasmedge": "interpreter"}


def command(runners, runtime, case, wasm=None):
    return [str(runners / (runtime + "-phase")), "--wasm", str(wasm or EMBEDDED / case["wasm"]),
            "--backend", BACKENDS[runtime], "--iterations", str(case["iterations"]),
            "--seed", str(case["seed"]), "--expected", str(case["expectedChecksum"]),
            "--samples", "7", "--warmups", "2"]


def execute(arguments, acknowledgement="run\n", failure=None):
    process = subprocess.run(arguments, input=acknowledgement, text=True,
                             capture_output=True, timeout=30)
    lines = process.stdout.splitlines()
    if not lines or lines[0] != '{"ready":true}':
        raise AssertionError(f"missing ready record: {process}")
    if failure:
        if process.returncode == 0 or len(lines) != 1 or failure not in process.stderr:
            raise AssertionError(f"expected {failure!r} rejection without result: {process}")
        return None
    if process.returncode or len(lines) != 2 or process.stderr:
        raise AssertionError(f"runner failed: {process}")
    return json.loads(lines[1])


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--runners-dir", type=Path, required=True)
    parser.add_argument("--runtimes", nargs="+", choices=tuple(BACKENDS), default=tuple(BACKENDS))
    args = parser.parse_args()
    runners = args.runners_dir.resolve()
    workloads = json.loads((EMBEDDED / "fixtures/manifest.json").read_text())["workloads"]
    # Deduplicate the two sustained rows' small oracle cases; the benchmark
    # itself executes every main workload separately after this validation.
    cases = {}
    for workload in workloads:
        for validation in workload["validationCases"]:
            case = dict(workload, **validation)
            cases[(case["wasm"], case["seed"], case["iterations"])] = case
    positive, negative = 0, 0
    for runtime in args.runtimes:
        for case in cases.values():
            result = execute(command(runners, runtime, case))
            expected = case["expectedChecksum"]
            if (result["result"] != expected or result["firstResult"] != expected or
                    result["warmResults"] != [expected] * 7 or result["warmupResults"] != [expected] * 2):
                raise AssertionError(f"unchecked result from {runtime}: {result}")
            if set(result["phasesNs"]) != {"load", "engineSetup", "compileValidate", "instantiate",
                                          "exportLookup", "firstCall", "cleanup"}:
                raise AssertionError("incomplete phase record")
            if (len(result["warmCallNs"]) != 7 or len(result["warmupCallNs"]) != 2 or
                    len(result["timerPairNs"]) != 7):
                raise AssertionError("incomplete timing arrays")
            if runtime == "wasmedge" and (sum(result["preparationPartsNs"].values()) !=
                                          result["phasesNs"]["compileValidate"]):
                raise AssertionError("WasmEdge preparation aggregation mismatch")
            positive += 1
        control = workloads[0]
        base_command = command(runners, runtime, control)
        for acknowledgements, success in (("sample\n" * 7, True), ("", False), ("wrong\n", False)):
            process = subprocess.run(base_command, input="run\n" + acknowledgements, text=True,
                                     capture_output=True, timeout=30,
                                     env=dict(os.environ, WASMVM_INTERLEAVE="1"))
            events = [json.loads(line) for line in process.stdout.splitlines()]
            if events[:2] != [{"ready": True}, {"sampleReady": 0}]:
                raise AssertionError("missing interleaved readiness")
            if success:
                if (process.returncode or len(events) != 9 or
                        events[1:8] != [{"sampleReady": i} for i in range(7)] or
                        events[-1]["warmResults"] != [control["expectedChecksum"]] * 7):
                    raise AssertionError("invalid interleaved results")
                positive += 1
            else:
                if process.returncode == 0 or len(events) != 2 or "sample acknowledgement" not in process.stderr:
                    raise AssertionError("accepted invalid sample acknowledgement")
                negative += 1
        for acknowledgement in ("", "wrong\n"):
            execute(base_command, acknowledgement, "acknowledgement")
            negative += 1
        for option, value, expected_error in [("--iterations", "0", "iterations must"),
                                               ("--seed", "4294967296", "uint32 range"),
                                               ("--samples", "65", "samples 1..64"),
                                               ("--backend", "unknown", "supported backend"),
                                               ("--expected", "0", "first call checksum mismatch")]:
            invalid = base_command.copy()
            invalid[invalid.index(option) + 1] = value
            execute(invalid, failure=expected_error)
            negative += 1
        for fixture, expected, expected_error in [("start", 0, "start sections"),
                                                    ("import", 0, "imports are excluded"),
                                                    ("wrong-signature", 0, "signature"),
                                                    ("trap", 0, "unreachable"),
                                                    ("mutating", 1, "warmup call checksum mismatch")]:
            case = dict(control, expectedChecksum=expected)
            path = EMBEDDED / "wazero/testdata" / (fixture + ".wasm")
            execute(command(runners, runtime, case, path), failure=expected_error)
            negative += 1
        with tempfile.TemporaryDirectory(prefix="wasmvm-invalid-module-") as temporary:
            malformed = Path(temporary) / "malformed.wasm"
            malformed.write_bytes(b"\x00asm\x01\x00\x00\x00\x01\xff")
            execute(command(runners, runtime, control, malformed), failure="truncated section")
            negative += 1
    print(json.dumps({"positiveCases": positive, "negativeCases": negative, "status": "passed"}))


if __name__ == "__main__":
    main()
