"""Load complete compiled Core fixtures and verify their source/oracle identity."""
import hashlib
import json
from pathlib import Path
from .oracle import checksum


def load(directory):
    directory = Path(directory)
    manifest = json.loads((directory / "manifest.json").read_text())
    if manifest.get("schemaVersion") != 1 or manifest.get("complete") is not True or manifest.get("profile") != "Core 2.0":
        raise ValueError("incomplete compiled-application manifest")
    names = {"clang", "rust"}
    if set(manifest["producers"]) != names or len(manifest["workloads"]) != 2 or {row["producer"] for row in manifest["workloads"]} != names:
        raise ValueError("compiled-application producer matrix differs")
    result = []
    def verified(name, expected):
        path = Path(name)
        if path.is_absolute() or ".." in path.parts:
            raise ValueError("fixture path escapes its directory")
        path = directory / path
        if hashlib.sha256(path.read_bytes()).hexdigest() != expected:
            raise ValueError("compiled-application input hash mismatch")
        return path
    for row in manifest["workloads"]:
        producer = manifest["producers"][row["producer"]]
        if row["id"] != "applications/sha256-" + row["producer"] or row["family"] != "application":
            raise ValueError("compiled-application identity differs")
        if row["maxIterations"] != 65536 or checksum(row["seed"], row["iterations"]) != row["expectedChecksum"]:
            raise ValueError("compiled-application oracle differs")
        verified(producer["source"], producer["sourceSha256"])
        path = verified(row["wasm"], row["wasmSha256"])
        if path.stat().st_size != row["bytes"] or producer["validation"]["exit"] != 0:
            raise ValueError("compiled-application validation differs")
        result.append(dict(row, path=str(path), producer=producer))
    return result
