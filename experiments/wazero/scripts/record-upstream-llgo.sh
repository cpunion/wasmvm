#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
project_dir=$(cd -- "$script_dir/../../.." && pwd)
artifact_dir=${WASMVM_EVAL_ARTIFACTS:-"$project_dir/artifacts/wazero/llgo-upstream"}
mkdir -p -- "$artifact_dir"
artifact_dir=$(cd -- "$artifact_dir" && pwd)
export WASMVM_EVAL_ARTIFACTS=$artifact_dir

set +e
"$script_dir/run.sh" llgo upstream "$@"
result=$?
set -e

# A known upstream link limitation is an assessment result, not a passing VM
# test. Any other failure, or newly successful upstream execution, requires a
# deliberate review of this expectation.
python3 - "$artifact_dir/test.log" "$result" <<'PY'
import pathlib, re, sys

log_path = pathlib.Path(sys.argv[1])
status = int(sys.argv[2])
if status == 0:
    raise SystemExit("upstream LLGo unexpectedly passed; remove/review the baseline limitation")
if status != 1 or not log_path.is_file():
    raise SystemExit("upstream LLGo failed outside the expected linker stage")
log = log_path.read_text(errors="replace")
symbols = set(re.findall(r"undefined symbol: ([^\s]+)", log))
symbols.update(re.findall(r"undefined reference to [`'\"]?([^`'\"\s]+)", log))
allowed = {"runtime.memmove", "runtime.getAuxv"}
if not symbols or not symbols <= allowed or "linker command failed" not in log:
    raise SystemExit("unexpected upstream failure; inspect the complete test.log")
for line in log.splitlines():
    if "error:" in line and not any(message in line for message in (
        "undefined symbol:", "undefined reference to", "linker command failed",
    )):
        raise SystemExit("additional unexpected error; inspect the complete test.log")
if "=== RUN" in log or "--- FAIL" in log:
    raise SystemExit("guest tests ran or failed; this is not the expected pre-execution linker failure")
print("Recorded upstream LLGo limitation: missing " + ", ".join(sorted(symbols)))
print("No guest executed in this baseline. The interpreter-only mode must pass independently.")
PY
