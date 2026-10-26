"""Integrity and independent-oracle checks for parameterized embedded fixtures."""

import hashlib
import importlib.util
import json
from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parent / "embedded"
SPEC = importlib.util.spec_from_file_location("embedded_fixture_generator", ROOT / "generate.py")
GENERATOR = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(GENERATOR)


def unsigned_leb(data, offset):
    value = 0
    for shift in range(0, 35, 7):
        byte = data[offset]
        offset += 1
        value |= (byte & 0x7f) << shift
        if byte < 0x80:
            return value, offset
    raise ValueError("invalid u32 LEB128")


def binary_sections(data):
    if data[:8] != b"\x00asm\x01\x00\x00\x00":
        raise ValueError("not a Core version-1 module")
    sections = {}
    offset = 8
    while offset < len(data):
        identifier = data[offset]
        length, start = unsigned_leb(data, offset + 1)
        end = start + length
        if end > len(data) or (identifier != 0 and identifier in sections):
            raise ValueError("invalid section framing")
        if identifier:
            sections[identifier] = data[start:end]
        offset = end
    return sections


class EmbeddedFixtureTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.manifest = json.loads((ROOT / "fixtures/manifest.json").read_text())

    def test_complete_parameterized_case_matrix(self):
        manifest = self.manifest
        self.assertEqual(manifest["schemaVersion"], 1)
        self.assertEqual(manifest["producer"], "wasm-tools 1.258.0")
        self.assertTrue(manifest["origin"].startswith("authored in wasmvm"))
        self.assertEqual(manifest["api"]["parameters"], ["i32", "i32"])
        cases = manifest["workloads"]
        self.assertEqual([(x["id"], x["kind"], x["iterations"]) for x in cases], list(GENERATOR.CASES))
        self.assertEqual(len({x["wasm"] for x in cases}), 4)
        for case in cases:
            with self.subTest(case=case["id"]):
                control = case["kind"] == "call-control"
                self.assertEqual(case["unit"], "ns/call" if control else "ns/iteration")
                self.assertEqual(case["memoryResetIncluded"], case["kind"] == "memory-load-store")
                self.assertEqual(case["memoryResetBytes"], 32768 if case["memoryResetIncluded"] else 0)
                self.assertEqual(case["expectedChecksum"], GENERATOR.checksum(case["kind"], case["seed"], case["iterations"]))

    def test_committed_hashes_source_and_minimal_binary_abi(self):
        for case in self.manifest["workloads"]:
            with self.subTest(case=case["id"]):
                for kind in ("wat", "wasm"):
                    path = (ROOT / case[kind]).resolve()
                    self.assertTrue(path.is_relative_to(ROOT / "fixtures"))
                    self.assertEqual(hashlib.sha256(path.read_bytes()).hexdigest(), case[kind + "Sha256"])
                self.assertEqual((ROOT / case["wat"]).read_text(), GENERATOR.module(case["kind"]))
                sections = binary_sections((ROOT / case["wasm"]).read_bytes())
                # No imports, start function, or initialized data. Instantiation
                # therefore cannot silently execute the benchmark workload.
                for forbidden in (2, 8, 11):
                    self.assertNotIn(forbidden, sections)
                self.assertEqual(sections[1], b"\x01\x60\x02\x7f\x7f\x01\x7f")
                exports = sections[7]
                self.assertEqual(exports[:6], b"\x01\x03run\x00")
                function_index, end = unsigned_leb(exports, 6)
                self.assertEqual(end, len(exports))
                self.assertEqual(function_index, 1 if case["kind"] == "function-calls" else 0)
                if case["kind"] == "memory-load-store":
                    self.assertEqual(sections[5], b"\x01\x01\x01\x01")  # one fixed 64 KiB memory
                else:
                    self.assertNotIn(5, sections)

    def test_oracle_small_values_wraparound_and_exact_iteration_count(self):
        # Literal results from the first arithmetic step independently pin the
        # initial state, wrapping behavior, and one-iteration loop semantics.
        expected = {
            "integer-mix": (1013904223, 1067169023, 1012239698),
            "function-calls": (2654435769, 2654435641, 1640531526),
            "memory-load-store": (2654435769, 2654435771, 2654435767),
        }
        for kind, values in expected.items():
            self.assertEqual(tuple(GENERATOR.checksum(kind, seed, 1) for seed in (0, 1, 0xffffffff)), values)
        self.assertEqual(GENERATOR.checksum("memory-load-store", 0, 3), 3668340014)
        self.assertEqual(GENERATOR.checksum("call-control", 0xffffffff, 1), 0xffffffff)
        for case in self.manifest["workloads"]:
            for vector in case["validationCases"]:
                self.assertEqual(vector["expectedChecksum"], GENERATOR.checksum(case["kind"], vector["seed"], vector["iterations"]))

    def test_invalid_input_cannot_masquerade_as_a_workload_count(self):
        for count in (0, -1, 1 << 31, True, 1.5):
            for kind in GENERATOR.KINDS:
                with self.subTest(kind=kind, count=count), self.assertRaises(ValueError):
                    GENERATOR.checksum(kind, 0, count)
        for seed in (-1, 1 << 32, True, 1.5):
            with self.assertRaises(ValueError):
                GENERATOR.checksum("integer-mix", seed, 1)
        with self.assertRaises(ValueError):
            GENERATOR.checksum("call-control", 0, 2)
        with self.assertRaises(ValueError):
            GENERATOR.checksum("unknown", 0, 1)


if __name__ == "__main__":
    unittest.main()
