"""Ensure common instrumentation cannot replace a baseline's engine code."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import ci
import own_build


class IsolatedSourceTests(unittest.TestCase):
    def commit(self, root):
        subprocess.run(["git", "-C", str(root), "add", "."], check=True, capture_output=True)
        subprocess.run(["git", "-C", str(root), "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
                        "commit", "-qm", "fixture"], check=True, capture_output=True)
        return ci.capture(["git", "-C", str(root), "rev-parse", "HEAD"])

    def test_archive_and_overlay_keep_engine_revision_exact(self):
        with tempfile.TemporaryDirectory() as directory:
            temporary = Path(directory)
            root = temporary / "source"
            root.mkdir()
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            (root / "go.mod").write_text("module github.com/cpunion/wasmvm\ngo 1.24.0\n")
            (root / "pkg/vm").mkdir(parents=True)
            engine = root / "pkg/vm/engine.go"
            engine.write_text("package vm\nconst Revision = 1\n")
            for path in own_build.HARNESS:
                (root / path).mkdir(parents=True)
                (root / path / "main.go").write_text("package main\n// old harness\n")
            adapter = root / "internal/platform/native/adapter.go"
            adapter.parent.mkdir(parents=True)
            adapter.write_text("package native\n")
            base_sha = self.commit(root)
            base = temporary / "base"
            subprocess.run(["git", "clone", "-q", str(root), str(base)], check=True)
            engine.write_text("package vm\nconst Revision = 2\n")
            for path in own_build.HARNESS:
                (root / path / "main.go").write_text("package main\n// new harness\n")
            adapter.write_text("package native\nfunc CPUFeatures() uint32 { return 3 }\n")
            phase = root / "internal/bench/phase/phase.go"
            phase.write_text('package phase\nimport (\n"github.com/cpunion/wasmvm/internal/platform/native"\n)\n'
                             'func metadata() { out.NativeCPUFeatures = native.CPUFeatures() }\n')
            head_sha = self.commit(root)
            # An untracked file must not accidentally become baseline/head code.
            (root / "pkg/vm/untracked.go").write_text("package vm\nconst Untracked = true\n")
            baseline = temporary / "baseline-build"
            candidate = temporary / "candidate-build"
            before = own_build.prepare(root, base, base_sha, baseline)
            after = own_build.prepare(root, root, head_sha, candidate)
            self.assertEqual((baseline / "pkg/vm/engine.go").read_text(), "package vm\nconst Revision = 1\n")
            self.assertEqual((candidate / "pkg/vm/engine.go").read_text(), "package vm\nconst Revision = 2\n")
            self.assertFalse((candidate / "pkg/vm/untracked.go").exists())
            self.assertNotEqual(before["sourceDigest"], after["sourceDigest"])
            phase_path = "internal/bench/phase/phase.go"
            self.assertEqual({k: v for k, v in before["harnessHashes"].items() if k != phase_path},
                             {k: v for k, v in after["harnessHashes"].items() if k != phase_path})
            self.assertEqual((baseline / "internal/platform/native/adapter.go").read_text(), "package native\n")
            self.assertIn("out.NativeCPUFeatures = 0", (baseline / phase_path).read_text())
            self.assertNotIn('"github.com/cpunion/wasmvm/internal/platform/native"', (baseline / phase_path).read_text())
            self.assertIn("native.CPUFeatures()", (candidate / phase_path).read_text())
            engine.write_text("dirty tracked engine")
            with self.assertRaises(subprocess.CalledProcessError):
                own_build.prepare(root, root, head_sha, temporary / "dirty-build")
            with self.assertRaises(ValueError):
                own_build.prepare(root, base, head_sha, temporary / "wrong-revision")

    def test_old_engine_is_explicitly_unavailable_and_symlinks_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory) / "source"
            root.mkdir()
            subprocess.run(["git", "init", "-q", str(root)], check=True)
            (root / "go.mod").write_text("module github.com/cpunion/wasmvm\n")
            sha = self.commit(root)
            result = own_build.prepare(root, root, sha, Path(directory) / "archive")
            self.assertEqual(result["status"], "baseline-unavailable")
            self.assertEqual(result["sourceRevision"], sha)
            (root / "link").symlink_to(root / "go.mod")
            with self.assertRaisesRegex(ValueError, "symlink"):
                own_build.hashes(root, ("link",))

    def test_missing_or_wrong_toolchain_is_not_silently_replaced(self):
        with patch("own_build.shutil.which", return_value=None), patch.dict(os.environ, {}, clear=True):
            with self.assertRaisesRegex(ValueError, "both Go"):
                own_build.toolchains()
        with patch("own_build.shutil.which", return_value="tool"), patch("own_build.ci.capture", return_value="go1.0"):
            with self.assertRaisesRegex(ValueError, "differs from pin"):
                own_build.toolchains()
        with patch("own_build.shutil.which", return_value="tool"), patch("own_build.ci.capture", return_value=own_build.GO_VERSION), patch.dict(os.environ, {}, clear=True):
            with self.assertRaisesRegex(ValueError, "WASMVM_LLGO_REVISION"):
                own_build.toolchains()

    def test_environment_pin_does_not_override_the_executable_version(self):
        with tempfile.TemporaryDirectory() as directory:
            go, llgo = (Path(directory) / name for name in ("go", "llgo"))
            go.write_bytes(b"go fixture")
            llgo.write_bytes(b"llgo fixture")
            version = "llgo (devel) linux/amd64"

            def capture(command):
                if command[1:] == ["env", "GOVERSION"]:
                    return own_build.GO_VERSION
                if command == [str(llgo), "version"]:
                    return version
                return "fixture build information"

            with patch("own_build.shutil.which", return_value=str(go)), \
                    patch("own_build.ci.capture", side_effect=capture), \
                    patch.dict(os.environ, {"WASMVM_LLGO": str(llgo), "WASMVM_LLGO_REVISION": own_build.LLGO_REVISION}), \
                    patch("own_build.platform.system", return_value="Linux"), \
                    patch("own_build.platform.machine", return_value="x86_64"):
                for version in ("llgo (devel) linux/amd64", "llgo v0.0.0-bc7c991d8fe9 linux/amd64"):
                    with self.assertRaisesRegex(ValueError, "executable version"):
                        own_build.toolchains()
                version = "llgo v0.0.0-" + own_build.LLGO_REVISION[:12] + " linux/amd64"
                identity = own_build.toolchains()["llgo"]
                self.assertEqual(identity["version"], version)
                self.assertEqual(identity["goBuildInfo"], "fixture build information")


if __name__ == "__main__":
    unittest.main()
