"""Replay hash-checked frozen ELF binaries; no compiler or source changes."""
import hashlib,json,math,os,random,statistics,subprocess
from pathlib import Path
ROOT = Path.cwd()
OUT = ROOT / "policy-replay"
OUT.mkdir()
run = lambda args, **kw: subprocess.check_output(args, text=True, **kw)
original = json.loads((ROOT / "frozen/results.json").read_text())
result = {"host": run(["lscpu"]), "compiler": original["compiler"], "roles": {},
    "frozenBuildRun": 43782740347, "frozenArtifact": "llgo-main-regression-ubuntu-22.04",
    "scope": "Separate native replay cohort, not pooled or portfolio acceptance"}
fixture = ROOT / "head/benchmarks/native/memory-fixtures/dependent-address.wasm"
for role in ["counter", "dense", "scalar"]:
    source = original["roles"][role]
    records = {"source": source["source"], "builds": source["builds"], "blocks": []}
    for label in ["base", "fix", "base-aa", "fix-aa"]:
        binary = ROOT / "frozen" / (role + "-" + label)
        assert hashlib.sha256(binary.read_bytes()).hexdigest() == source["builds"][label.split("-")[0]]["sha256"]
        binary.chmod(0o755)
    for block in range(20):
        order = ["base", "base-aa", "fix", "fix-aa"]
        random.Random(20261116 + block).shuffle(order)
        observations = {}
        for label in order:
            binary = ROOT / "frozen" / (role + "-" + label)
            command = [str(binary), "--mode", "api", "--backend", "interpreter", "--wasm", str(fixture),
                "--iterations", "100000", "--seed", "305419896", "--expected", "509818616", "--samples", "8", "--warmups", "2"]
            proc = subprocess.run(command, input="run\n", text=True, capture_output=True, check=True, timeout=30)
            receipt = json.loads(proc.stdout.splitlines()[-1])
            assert receipt["warmResults"] == [509818616] * 8 and receipt["result"] == 509818616
            assert receipt["compilerRevision"] == result["compiler"][label.split("-")[0]]["revision"]
            observations[label] = receipt
        records["blocks"].append({"order": order, "observations": observations})
    records["ratios"] = {}
    for left, right in [("fix", "base"), ("base-aa", "base"), ("fix-aa", "fix")]:
        values = [math.log(statistics.mean(block["observations"][left]["warmCallNs"]) / statistics.mean(block["observations"][right]["warmCallNs"])) for block in records["blocks"]]
        rng = random.Random(20261116)
        boot = sorted(math.exp(statistics.mean(rng.choices(values, k=len(values)))) for _ in range(10000))
        records["ratios"][left + "/" + right] = [math.exp(statistics.mean(values)), boot[250], boot[9749]]
    result["roles"][role] = records
    (OUT / "results.json").write_text(json.dumps(result, indent=2) + "\n")
    print(role, records["builds"], records["ratios"], flush=True)

