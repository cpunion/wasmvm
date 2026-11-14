"""Whole-process PMU diagnostics, separate from the fixed-block timing protocol."""

import argparse
import glob
import json
import os
from pathlib import Path
import random
import shutil
import signal
import subprocess


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("results", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    data = json.loads(args.results.read_text())
    metadata = dict(scope=__doc__, samples=1000, repetitions=3, runs=[], probes=[])

    def save():
        (args.output / "metadata.json").write_text(json.dumps(metadata, indent=2) + "\n")

    if not hasattr(os, "sched_getaffinity"):
        metadata["status"] = "unsupported-host"
        save()
        return
    core = min(os.sched_getaffinity(0))
    os.sched_setaffinity(0, {core})
    metadata["affinity"] = [core]
    sysctl = subprocess.run(["sudo", "sysctl", "-w", "kernel.perf_event_paranoid=-1"],
                            capture_output=True, text=True, timeout=15)
    metadata["sysctl"] = dict(exitCode=sysctl.returncode, stdout=sysctl.stdout, stderr=sysctl.stderr)
    candidates = sorted(glob.glob("/usr/lib/linux-tools/*/perf"))
    installed = shutil.which("perf")
    if installed:
        candidates.append(installed)
    events = "cycles,instructions,branches,branch-misses"
    tool = None
    for index, candidate in enumerate(candidates):
        csv = args.output / f"probe-{index}.csv"
        probe = subprocess.run([candidate, "stat", "-x,", "-e", events, "-o", str(csv), "--", "true"],
                               capture_output=True, text=True, timeout=15)
        text = csv.read_text() if csv.exists() else ""
        metadata["probes"].append(dict(tool=candidate, exitCode=probe.returncode,
                                       stderr=probe.stderr, counters=text))
        if probe.returncode == 0 and "<not supported>" not in text and "<not counted>" not in text:
            tool = candidate
            break
    if tool is None:
        metadata["status"] = "pmu-unavailable" if candidates else "perf-tool-missing"
        save()
        print("PMU diagnostic:", metadata["status"], flush=True)
        return

    commands = {}
    for row in data["cases"][0]["observations"]:
        name = row["runtime"]
        if name.startswith("aa-") or not name.endswith(("go-interpreter", "llgo-interpreter")):
            continue
        command = list(row["command"])
        command[command.index("--samples") + 1] = str(metadata["samples"])
        commands[name] = command
    metadata["status"] = "running"
    metadata["tool"] = tool
    save()
    env = dict(os.environ, GOMAXPROCS="1")
    rng = random.Random(3094)
    for repetition in range(metadata["repetitions"]):
        order = sorted(commands)
        rng.shuffle(order)
        for name in order:
            label = f"{repetition + 1}-{name}"
            command = [tool, "stat", "-x,", "-e", events, "-o", str(args.output / (label + ".csv")),
                       "--", *commands[name]]
            with (args.output / (label + ".stdout")).open("wb") as stdout, \
                    (args.output / (label + ".stderr")).open("wb") as stderr:
                process = subprocess.Popen(command, env=env, stdout=stdout, stderr=stderr, start_new_session=True)
                try:
                    code = process.wait(timeout=60)
                except subprocess.TimeoutExpired:
                    os.killpg(process.pid, signal.SIGKILL)
                    process.wait(timeout=10)
                    metadata["status"] = "timeout"
                    save()
                    raise
            metadata["runs"].append(dict(name=name, repetition=repetition + 1, command=command, exitCode=code))
            save()
            if code:
                raise RuntimeError("PMU workload failed: " + label)
    metadata["status"] = "complete"
    save()


if __name__ == "__main__":
    main()
