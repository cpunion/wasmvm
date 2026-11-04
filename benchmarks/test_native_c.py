"""Check ordinary C references against independent byte/arithmetic oracles."""
import importlib.util
import json
from pathlib import Path
import shutil
import subprocess
import tempfile
import time
import unittest

from native.memory_fixtures import checksum as memory_checksum
from native.runtime_fixtures import checksum as vector_checksum

ROOT = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("native_c_core_oracle", ROOT / "embedded/generate.py")
CORE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(CORE)


class NativeCReferenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.directory = tempfile.TemporaryDirectory()
        cls.addClassCleanup(cls.directory.cleanup)
        cls.binary = Path(cls.directory.name) / "native-reference"
        compiler = shutil.which("clang") or shutil.which("cc")
        if compiler is None:
            raise RuntimeError("native C reference checks require a C compiler")
        subprocess.run([compiler, "-std=c11", "-D_POSIX_C_SOURCE=200809L", "-O3",
                        "-Wall", "-Wextra", "-Werror", "-o", str(cls.binary),
                        str(ROOT / "native/c/scalar_baseline.c")], check=True,
                       capture_output=True, text=True)

    def check(self, kind, seed, count, expected):
        result = subprocess.run([str(self.binary), kind, str(seed), str(count), str(expected)],
                                check=True, capture_output=True, text=True, timeout=10)
        record = json.loads(result.stdout)
        self.assertEqual((record["checksum"], record["seed"], record["iterations"]),
                         (expected, seed, count))
        self.assertEqual(len(record["warmCallNs"]), 7)

    def test_reference_clocks_share_the_observer_time_domain(self):
        compiler = shutil.which("clang") or shutil.which("cc")
        sources = [ROOT / "embedded/c/common.h", *(ROOT / "native/c").glob("*_baseline.c")]
        self.assertEqual(len(sources), 4)
        for source in sources:
            with self.subTest(source=source.name):
                path = Path(self.directory.name) / "clock.c"
                binary = path.with_suffix("")
                path.write_text('#define main reference_main\n#include ' + json.dumps(str(source)) +
                                '\n#undef main\nint main(void) { printf("%" PRIu64 "\\n", now_ns()); }\n')
                subprocess.run([compiler, "-std=c11", "-D_POSIX_C_SOURCE=200809L", "-O2",
                                "-Wall", "-Wextra", "-Werror", "-Wno-unused-function",
                                str(path), "-o", str(binary)],
                               check=True, capture_output=True, text=True)
                before = time.perf_counter_ns()
                sampled = int(subprocess.check_output([str(binary)], text=True))
                after = time.perf_counter_ns()
                self.assertLessEqual(before, sampled)
                self.assertLessEqual(sampled, after)

    def test_simd_byte_overlap_sign_extension_and_wrap(self):
        for kind in ("simd-memory-full", "simd-memory-widen", "simd-memory-lanes"):
            for seed in (0, 1, 0x12345678, 0x7fffffff, 0x80000000, 0xffffffff):
                for count in (1, 2, 15, 16, 17, 127, 128, 129, 257, 4097):
                    with self.subTest(kind=kind, seed=seed, count=count):
                        self.check(kind, seed, count, vector_checksum(kind, seed, count))

    def test_existing_scalar_references(self):
        for kind in ("integer-loop", "function-calls", "memory-load-store",
                     "narrow-unaligned", "dependent-address", "float64-unaligned"):
            for seed in (0, 0x12345678, 0xffffffff):
                for count in (1, 17, 8193):
                    with self.subTest(kind=kind, seed=seed, count=count):
                        if kind in ("integer-loop", "function-calls", "memory-load-store"):
                            expected = CORE.checksum("integer-mix" if kind == "integer-loop" else kind,
                                                     seed, count)
                        else:
                            expected = memory_checksum(kind, seed, count)
                        self.check(kind, seed, count, expected)

    def test_checksum_failure_and_iteration_limits(self):
        wrong = subprocess.run([str(self.binary), "simd-memory-widen", "1", "17", "1"],
                               capture_output=True, text=True, timeout=10)
        self.assertEqual(wrong.returncode, 1)
        self.assertIn("checksum:", wrong.stderr)
        for count in ("0", "-1", "2147483648", "4294967297", "17junk", ""):
            with self.subTest(count=count):
                result = subprocess.run([str(self.binary), "simd-memory-full", "0", count, "0"],
                                        capture_output=True, text=True, timeout=10)
                self.assertEqual(result.returncode, 2)
        result = subprocess.run([str(self.binary), "float64-unaligned", "0", "1000001", "0"],
                                capture_output=True, text=True, timeout=10)
        self.assertEqual(result.returncode, 2)
