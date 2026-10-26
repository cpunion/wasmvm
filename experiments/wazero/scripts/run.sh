#!/usr/bin/env bash
set -euo pipefail

usage() {
  printf 'usage: %s <go|llgo> <upstream|interpreter-only> [go test args...]\n' "$0" >&2
  exit 2
}

[[ $# -ge 2 ]] || usage
compiler=$1
mode=$2
shift 2
case "$compiler" in go|llgo) ;; *) usage ;; esac
case "$mode" in upstream|interpreter-only) ;; *) usage ;; esac

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
eval_dir=$(cd -- "$script_dir/.." && pwd)
project_dir=$(cd -- "$eval_dir/../.." && pwd)
artifact_dir=${WASMVM_EVAL_ARTIFACTS:-"$project_dir/artifacts/wazero/$compiler-$mode"}
mkdir -p -- "$artifact_dir"
artifact_dir=$(cd -- "$artifact_dir" && pwd)

compiler_path=$(command -v "$compiler")
go_path=$(command -v go)
command -v python3 >/dev/null
command -v patch >/dev/null
export GOWORK=off

work_dir=$(mktemp -d "${TMPDIR:-/tmp}/wasmvm-wazero-eval.XXXXXX")
trap 'rm -rf -- "$work_dir"' EXIT
cp -R "$eval_dir/." "$work_dir/evaluation"

{
  printf 'mode=%s\ncompiler=%s\ncompiler_path=%s\n' "$mode" "$compiler" "$compiler_path"
  "$go_path" version
  if [[ "$compiler" == llgo ]]; then
    "$compiler_path" version
    "$compiler_path" env LLGO_ROOT GOROOT
  fi
} >"$artifact_dir/toolchains.txt"

"$go_path" mod download -json github.com/tetratelabs/wazero@v1.12.0 >"$artifact_dir/dependency.json"
dependency_dir=$(python3 - "$artifact_dir/dependency.json" <<'PY'
import json, sys
with open(sys.argv[1]) as source:
    dependency = json.load(source)
if dependency.get("Path") != "github.com/tetratelabs/wazero" or dependency.get("Version") != "v1.12.0":
    raise SystemExit("unexpected dependency identity")
print(dependency["Dir"])
PY
)

if [[ "$mode" == interpreter-only ]]; then
  cp -R "$dependency_dir/." "$work_dir/wazero"
  chmod -R u+w "$work_dir/wazero"
  patch --batch --fuzz=0 -p1 -d "$work_dir/wazero" \
    <"$eval_dir/patches/interpreter-only-v1.12.0.patch" >"$artifact_dir/patch.log"
  "$go_path" -C "$work_dir/evaluation" mod edit \
    "-replace=github.com/tetratelabs/wazero=$work_dir/wazero"
fi

# Resolve dependencies only in the temporary copy; never edit the repository's
# module files or the downloaded module cache.
"$go_path" -C "$work_dir/evaluation" mod download
test_args=(-v -count=1 -timeout=10m)
if [[ "$compiler" == go ]]; then
  test_args+=("-coverprofile=$artifact_dir/coverage.out")
fi

printf 'Running %s evaluator (%s); artifacts: %s\n' "$compiler" "$mode" "$artifact_dir"
set +e
(
  cd -- "$work_dir/evaluation"
  "$compiler_path" test "${test_args[@]}" "$@" ./...
) >"$artifact_dir/test.log" 2>&1
result=$?
set -e
printf '%s\n' "$result" >"$artifact_dir/exit-status.txt"
cat "$artifact_dir/test.log"
exit "$result"
