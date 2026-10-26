#!/usr/bin/env python3
"""Install checksum-pinned Linux amd64 competitors without running installers."""

import argparse
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import platform
import posixpath
import re
import shutil
import subprocess
import tarfile
import tempfile
import urllib.request


HERE = Path(__file__).resolve().parent
MANIFEST = HERE / "tools-linux-amd64.json"
DEFAULT_DESTINATION = HERE.parent / "artifacts/benchmark-tools/linux"


def safe_path(name):
    """Require portable archive-relative names, without traversal components."""
    path = PurePosixPath(name)
    if not name or path.is_absolute() or ".." in path.parts or "\\" in name:
        raise ValueError(f"unsafe archive path: {name!r}")
    return path


def download(tool, destination):
    """Bound the download by the pinned release size; hashing happens next."""
    if not tool["url"].startswith("https://"):
        raise ValueError("release downloads require HTTPS")
    request = urllib.request.Request(tool["url"], headers={"User-Agent": "wasmvm-benchmarks"})
    with urllib.request.urlopen(request, timeout=60) as response, destination.open("wb") as output:
        if not response.url.startswith("https://"):
            raise ValueError("release download redirected away from HTTPS")
        size = 0
        while chunk := response.read(1024 * 1024):
            size += len(chunk)
            if size > tool["size"]:
                raise ValueError("release download exceeds its pinned size")
            output.write(chunk)
    if size != tool["size"]:
        raise ValueError("release download does not match its pinned size")


def extract_verified(tool, archive, destination):
    """Verify all bytes, validate archive paths, then extract selected files."""
    with archive.open("rb") as source:
        digest = hashlib.file_digest(source, "sha256").hexdigest()
    if digest != tool["sha256"]:
        raise ValueError(f"SHA256 mismatch for {tool['id']}: {digest}")
    expected = {str(safe_path(name)) for name in tool["files"]}
    prefixes = [str(safe_path(name)).rstrip("/") + "/" for name in tool.get("filePrefixes", [])]
    if len(expected) != len(tool["files"]):
        raise ValueError("duplicate selected archive paths")
    with tarfile.open(archive) as source:
        members = source.getmembers()
        seen = set()
        for member in members:
            name = str(safe_path(member.name))
            if name in seen:
                raise ValueError(f"duplicate archive path: {name}")
            seen.add(name)
            if not (member.isfile() or member.isdir() or member.issym()):
                raise ValueError(f"unsupported archive member: {name}")
            if member.issym():
                target = PurePosixPath(member.linkname)
                resolved = posixpath.normpath(str(PurePosixPath(name).parent / target))
                if target.is_absolute() or "\\" in member.linkname or resolved == ".." or resolved.startswith("../"):
                    raise ValueError(f"unsafe archive symlink: {name}")
        if not expected.issubset(seen):
            raise ValueError(f"missing selected archive files: {sorted(expected - seen)}")
        for prefix in prefixes:
            if not any(name.startswith(prefix) for name in seen):
                raise ValueError(f"missing selected archive directory: {prefix}")
        # Python's data filter also prevents symlink pivots and special modes.
        selected = [m for m in members if str(PurePosixPath(m.name)) in expected or
                    any(str(PurePosixPath(m.name)).startswith(prefix) for prefix in prefixes)]
        source.extractall(destination, members=selected, filter="data")
    if "executable" in tool:
        executable = destination / safe_path(tool["executable"])
        if executable.is_symlink() or not executable.is_file() or not os.access(executable, os.X_OK):
            raise ValueError(f"missing regular executable: {executable}")
    for directory in tool["libraryDirectories"] + tool.get("includeDirectories", []):
        if not (destination / safe_path(directory)).is_dir():
            raise ValueError(f"missing library directory: {directory}")


def paths_for(tools, destination):
    bins, libraries = [], []
    for tool in tools:
        directory = destination / tool["id"]
        if "executable" in tool:
            bins.append(str((directory / tool["executable"]).parent))
        libraries.extend(str(directory / path) for path in tool["libraryDirectories"])
    return bins, libraries


def verify_version(tool, directory, environment):
    executable = directory / tool["executable"]
    result = subprocess.run([str(executable), "--version"], env=environment, check=True,
                            stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, timeout=30)
    first_line = result.stdout.splitlines()[0] if result.stdout else ""
    if not re.search(tool["versionPattern"], first_line):
        raise ValueError(f"unexpected {tool['id']} version: {result.stdout!r}")
    return result.stdout.strip()


def install(destination, manifest):
    """Publish a complete installation atomically; refuse existing destinations."""
    if destination.exists() or destination.is_symlink():
        raise ValueError(f"destination already exists: {destination}; choose a fresh directory")
    destination.parent.mkdir(parents=True, exist_ok=True)
    tools = manifest["tools"]
    with tempfile.TemporaryDirectory(prefix=".wasmvm-tools-", dir=destination.parent) as temporary:
        staging = Path(temporary) / "installation"
        staging.mkdir()
        for tool in tools:
            archive = Path(temporary) / (tool["id"] + ".archive")
            print(f"Downloading {tool['id']} {tool['version']} ({tool['size']} bytes)", flush=True)
            download(tool, archive)
            extract_verified(tool, archive, staging / tool["id"])
            archive.unlink()
        _, libraries = paths_for(tools, staging)
        environment = os.environ.copy()
        environment["LD_LIBRARY_PATH"] = os.pathsep.join(libraries)
        observed = {tool["id"]: verify_version(tool, staging / tool["id"], environment)
                    for tool in tools if "executable" in tool}
        # Keep the pinned provenance next to binaries; avoid ephemeral staging paths.
        report = dict(manifest, observedVersions={key: value.replace(str(staging), str(destination))
                                                for key, value in observed.items()})
        (staging / "installation.json").write_text(json.dumps(report, indent=2) + "\n")
        staging.rename(destination)
    return report


def github_environment(destination, tools):
    bins, libraries = paths_for(tools, destination)
    # Reject line separators rather than creating extra workflow environment keys.
    for value in bins + libraries:
        if "\n" in value or "\r" in value or os.pathsep in value:
            raise ValueError("installation path cannot contain line or path separators")
    with Path(os.environ["GITHUB_PATH"]).open("a") as output:
        output.write("\n".join(bins) + "\n")
    with Path(os.environ["GITHUB_ENV"]).open("a") as output:
        output.write("LD_LIBRARY_PATH=" + os.pathsep.join(libraries) + "\n")


def load_manifest(with_sdk=False):
    manifest = json.loads(MANIFEST.read_text())
    manifest["sdkEnabled"] = with_sdk
    sdks = manifest.pop("sdks", [])
    for tool in manifest["tools"]:
        sdk = tool.pop("sdk", {})
        if with_sdk:
            tool["files"].extend(sdk.get("files", []))
            tool["filePrefixes"] = sdk.get("filePrefixes", [])
            tool["includeDirectories"] = sdk.get("includeDirectories", [])
            tool["libraryDirectories"].extend(sdk.get("libraryDirectories", []))
    if with_sdk:
        manifest["tools"].extend(sdks)
    return manifest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--destination", type=Path, default=DEFAULT_DESTINATION)
    parser.add_argument("--github-env", action="store_true", help="append paths to GITHUB_PATH and GITHUB_ENV")
    parser.add_argument("--sdk", action="store_true", help="also install pinned headers and shared libraries for embedded runners")
    args = parser.parse_args()
    if platform.system() != "Linux" or platform.machine() not in {"x86_64", "amd64"}:
        parser.error("these release assets require Linux amd64")
    if args.github_env and not all(os.environ.get(name) for name in ("GITHUB_PATH", "GITHUB_ENV")):
        parser.error("--github-env requires GITHUB_PATH and GITHUB_ENV")
    # Check before resolve(), which would otherwise hide a symlink destination.
    if args.destination.is_symlink():
        parser.error("destination must not be a symlink")
    destination = args.destination.resolve()
    manifest = load_manifest(args.sdk)
    report = install(destination, manifest)
    if args.github_env:
        github_environment(destination, manifest["tools"])
    bins, libraries = paths_for(manifest["tools"], destination)
    print(json.dumps({"installation": str(destination / "installation.json"), "path": bins,
                      "LD_LIBRARY_PATH": os.pathsep.join(libraries),
                      "versions": report["observedVersions"]}, indent=2))


if __name__ == "__main__":
    main()
