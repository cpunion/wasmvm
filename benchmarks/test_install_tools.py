"""Integrity and extraction checks for pinned competitor installation."""

import hashlib
import io
import json
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile
import unittest
from unittest.mock import patch

import install_tools


class InstallerTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.archive = self.root / "fixture.tar.gz"
        self.tool = {"id": "example", "version": "1.2.3", "url": "https://example.org/release",
                     "files": ["bin/example", "lib/libexample.so.1", "lib/libexample.so"],
                     "executable": "bin/example", "libraryDirectories": ["lib"],
                     "versionPattern": r"^example 1\.2\.3$"}

    def make_archive(self, extra=None):
        with tarfile.open(self.archive, "w:gz") as archive:
            for name, contents in [("bin/example", b"#!/bin/sh\necho example 1.2.3\n"),
                                   ("lib/libexample.so.1", b"library"), ("unused.a", b"omit")]:
                member = tarfile.TarInfo(name)
                member.size = len(contents)
                member.mode = 0o755 if name == "bin/example" else 0o644
                archive.addfile(member, io.BytesIO(contents))
            member = tarfile.TarInfo("lib/libexample.so")
            member.type = tarfile.SYMTYPE
            member.linkname = "libexample.so.1"
            archive.addfile(member)
            if extra:
                archive.addfile(extra)
        self.tool["sha256"] = hashlib.sha256(self.archive.read_bytes()).hexdigest()
        self.tool["size"] = self.archive.stat().st_size

    def test_verified_selected_files_and_internal_symlink(self):
        self.make_archive()
        destination = self.root / "installed"
        install_tools.extract_verified(self.tool, self.archive, destination)
        self.assertEqual((destination / "lib/libexample.so").read_bytes(), b"library")
        self.assertFalse((destination / "unused.a").exists())
        self.assertEqual(install_tools.verify_version(self.tool, destination, os.environ.copy()), "example 1.2.3")

    def test_hash_failure_happens_before_extraction(self):
        self.make_archive()
        self.tool["sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "SHA256 mismatch"):
            install_tools.extract_verified(self.tool, self.archive, self.root / "installed")
        self.assertFalse((self.root / "installed").exists())

    def test_reject_traversal_special_files_and_duplicate_paths(self):
        members = [tarfile.TarInfo("../escaped"), tarfile.TarInfo("/absolute"),
                   tarfile.TarInfo("bin/example"), tarfile.TarInfo("device")]
        members[-1].type = tarfile.CHRTYPE
        for member in members:
            with self.subTest(name=member.name):
                self.make_archive(member)
                with self.assertRaises(ValueError):
                    install_tools.extract_verified(self.tool, self.archive, self.root / "installed")
                self.assertFalse((self.root / "installed").exists())

    def test_reject_symlink_escape_before_extraction(self):
        for target in ("/outside", "../../outside", "..\\outside"):
            member = tarfile.TarInfo("lib/escape")
            member.type = tarfile.SYMTYPE
            member.linkname = target
            self.make_archive(member)
            with self.subTest(target=target), self.assertRaisesRegex(ValueError, "unsafe archive symlink"):
                install_tools.extract_verified(self.tool, self.archive, self.root / "installed")

    def test_reject_missing_selected_file(self):
        self.make_archive()
        self.tool["files"].append("missing")
        with self.assertRaisesRegex(ValueError, "missing selected"):
            install_tools.extract_verified(self.tool, self.archive, self.root / "installed")

    def test_reject_unexpected_version(self):
        result = subprocess.CompletedProcess([], 0, "example 9.0.0\n")
        with patch("install_tools.subprocess.run", return_value=result), self.assertRaisesRegex(ValueError, "unexpected"):
            install_tools.verify_version(self.tool, self.root, {})

    def test_install_is_atomic_and_preserves_provenance(self):
        self.make_archive()
        manifest = {"schemaVersion": 1, "tools": [self.tool]}
        destination = self.root / "installed"

        def cached_download(tool, output):
            output.write_bytes(self.archive.read_bytes())

        with patch("install_tools.download", side_effect=cached_download):
            report = install_tools.install(destination, manifest)
        self.assertEqual(report["observedVersions"], {"example": "example 1.2.3"})
        self.assertEqual(json.loads((destination / "installation.json").read_text()), report)
        self.assertEqual(report["tools"][0]["sha256"], self.tool["sha256"])
        with self.assertRaisesRegex(ValueError, "already exists"):
            install_tools.install(destination, manifest)

    def test_failure_cleans_temporary_installation(self):
        self.make_archive()
        destination = self.root / "installed"
        with patch("install_tools.download", side_effect=RuntimeError("network failure")):
            with self.assertRaisesRegex(RuntimeError, "network failure"):
                install_tools.install(destination, {"tools": [self.tool]})
        self.assertFalse(destination.exists())
        self.assertFalse(list(self.root.glob(".wasmvm-tools-*")))

    def test_workflow_paths_are_concrete_and_reject_line_injection(self):
        path_file, env_file = self.root / "path", self.root / "env"
        with patch.dict(os.environ, {"GITHUB_PATH": str(path_file), "GITHUB_ENV": str(env_file)}):
            install_tools.github_environment(self.root, [self.tool])
            self.assertEqual(path_file.read_text(), str(self.root / "example/bin") + "\n")
            self.assertEqual(env_file.read_text(), "LD_LIBRARY_PATH=" + str(self.root / "example/lib") + "\n")
            with self.assertRaises(ValueError):
                install_tools.github_environment(self.root / "bad\nPATH=injected", [self.tool])

    def test_download_checks_transport_and_size(self):
        self.make_archive()
        payload = self.archive.read_bytes()
        for actual, url, expected in [(payload, self.tool["url"], None),
                                      (payload + b"extra", self.tool["url"], "exceeds"),
                                      (payload[:-1], self.tool["url"], "does not match"),
                                      (payload, "http://example.org/release", "redirected")]:
            response = io.BytesIO(actual)
            response.url = url
            with self.subTest(url=url, size=len(actual)), patch("install_tools.urllib.request.urlopen", return_value=response):
                if expected:
                    with self.assertRaisesRegex(ValueError, expected):
                        install_tools.download(self.tool, self.root / "download")
                else:
                    install_tools.download(self.tool, self.root / "download")
                    self.assertEqual((self.root / "download").read_bytes(), payload)

    def test_manifest_pins_versions_and_exact_release_assets(self):
        manifest = json.loads(install_tools.MANIFEST.read_text())
        self.assertEqual(manifest["target"], "linux-amd64")
        self.assertEqual({t["id"]: t["version"] for t in manifest["tools"]},
                         {"wasmtime": "48.0.1", "wasmer": "7.3.0", "wasmedge": "0.17.1"})
        for tool in manifest["tools"]:
            self.assertRegex(tool["sha256"], r"^[0-9a-f]{64}$")
            self.assertIn(tool["executable"], tool["files"])
            self.assertTrue(tool["url"].startswith("https://github.com/"))
            self.assertIn("/releases/download/", tool["url"])
            self.assertTrue(tool["licenseUrl"].startswith("https://github.com/"))


if __name__ == "__main__":
    unittest.main()
