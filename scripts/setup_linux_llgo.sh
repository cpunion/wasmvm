#!/usr/bin/env bash
set -euo pipefail

# Match the pinned setup-llgo dependencies, with bounded network retries before
# compilation. Keep compiler/test failures outside the installation retry path.
# https://github.com/xgo-dev/setup-llgo/blob/4b2f2c9aed3fd28f139615fe636dee5e2ce5bc91/action.yml
llvm_major=22
llvm_release=22.1.8
key_file=$(mktemp)
trap 'rm -f "$key_file"' EXIT

curl --fail --silent --show-error --location \
  --retry 3 --retry-all-errors --retry-delay 2 --retry-max-time 45 \
  --connect-timeout 10 --max-time 15 \
  https://apt.llvm.org/llvm-snapshot.gpg.key --output "$key_file"
sudo apt-key add "$key_file"
printf 'deb https://apt.llvm.org/%s/ llvm-toolchain-%s-%s main\n' \
  "$(lsb_release -cs)" "$(lsb_release -cs)" "$llvm_major" |
  sudo tee /etc/apt/sources.list.d/llvm.list

apt_options=(-o Acquire::Retries=2 -o Acquire::http::Timeout=20 -o Acquire::https::Timeout=20)
sudo apt-get "${apt_options[@]}" update
sudo apt-get "${apt_options[@]}" install -y \
  "clang-$llvm_major" "libclang-$llvm_major-dev" "lld-$llvm_major" \
  pkg-config libgc-dev libssl-dev zlib1g-dev libffi-dev libuv1-dev

llvm_bin="/usr/lib/llvm-$llvm_major/bin"
test "$("$llvm_bin/llvm-config" --version)" = "$llvm_release"
printf '%s\n' "$llvm_bin" >> "$GITHUB_PATH"
