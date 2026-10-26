from collections import Counter
import hashlib
import io
import json
from pathlib import Path
import tarfile
import tempfile
import unittest
from unittest.mock import patch

import core2_corpus as corpus


class Response(io.BytesIO):
    url = "https://example.test/archive"


class CorpusTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)

    def test_paths(self):
        self.assertEqual(str(corpus.relative_path("simd/i8x16.wast")), "simd/i8x16.wast")
        for name in ["", ".", "../escape", "/absolute", "a/../b", "a\\b", "a//b", "./a", "a\nb", "a\x00b"]:
            with self.subTest(name=name), self.assertRaises(ValueError):
                corpus.relative_path(name)

    def pin(self, payload):
        return {"sha256": hashlib.sha256(payload).hexdigest(), "size": len(payload),
                "url": "https://example.test/archive"}

    def test_download_and_cached_integrity(self):
        payload = b"pinned archive bytes"
        pin = self.pin(payload)
        with patch.object(corpus.urllib.request, "urlopen", return_value=Response(payload)) as request:
            path = corpus.fetch(pin, self.root / "cache")
            self.assertEqual(path.read_bytes(), payload)
            self.assertEqual(corpus.fetch(pin, self.root / "cache"), path)
            request.assert_called_once()
        path.write_bytes(b"x" * len(payload))
        with self.assertRaisesRegex(ValueError, "cached archive"):
            corpus.fetch(pin, self.root / "cache")

    def test_download_size_hash_and_redirect_failures(self):
        for payload, pin, response_url in [
            (b"long", self.pin(b"x"), "https://example.test/archive"),
            (b"short", self.pin(b"longer"), "https://example.test/archive"),
            (b"wrong", self.pin(b"right"), "https://example.test/archive"),
            (b"x", self.pin(b"x"), "http://example.test/archive"),
        ]:
            response = Response(payload)
            response.url = response_url
            with self.subTest(payload=payload), patch.object(corpus.urllib.request, "urlopen", return_value=response):
                with self.assertRaises(ValueError):
                    corpus.fetch(pin, self.root / "cache")
                self.assertEqual(list((self.root / "cache").iterdir()), [])
        for pin in [{**self.pin(b"x"), "sha256": "short"},
                    {**self.pin(b"x"), "size": 0},
                    {**self.pin(b"x"), "url": "file:///archive"}]:
            with self.assertRaises(ValueError):
                corpus.fetch(pin, self.root / "cache")

    def archive(self, entries):
        path = self.root / "source.tar.gz"
        with tarfile.open(path, "w:gz") as archive:
            for name, content in entries:
                info = tarfile.TarInfo(name)
                if content is None:
                    info.type = tarfile.SYMTYPE
                    info.linkname = "../../escape"
                    archive.addfile(info)
                else:
                    info.size = len(content)
                    archive.addfile(info, io.BytesIO(content))
        return path

    def test_complete_source_selection_includes_simd_and_licenses(self):
        pin = {"prefix": "spec-rev", "directory": "test/core", "scripts": 2}
        archive = self.archive([
            ("spec-rev/LICENSE", b"directory notice"),
            ("spec-rev/test/LICENSE", b"test license"),
            ("spec-rev/test/core/base.wast", b"(module)"),
            ("spec-rev/test/core/simd/vector.wast", b"(module (func))"),
            ("spec-rev/test/other/extra.wast", b"not part of core"),
        ])
        destination = self.root / "source"
        sources = corpus.extract_source(pin, archive, destination)
        self.assertEqual(set(sources), {"base.wast", "simd/vector.wast"})
        self.assertEqual((destination / "LICENSE.tests").read_bytes(), b"test license")
        self.assertFalse((destination / "wast/extra.wast").exists())
        with self.assertRaisesRegex(ValueError, "missing scripts"):
            corpus.extract_source({**pin, "scripts": 3}, archive, self.root / "missing")

    def test_selected_archive_links_duplicates_and_escape_fail(self):
        for entries in [[("chosen", None)], [("chosen", b"a"), ("chosen", b"b")], [("../escape", b"a")]]:
            with self.subTest(entries=entries), self.assertRaises(ValueError):
                corpus.extract(self.archive(entries), self.root / "selected", lambda name: name)

    def fixture(self):
        root = self.root / "output/fixtures"
        (root / "simple").mkdir(parents=True)
        for name in corpus.LICENSES.values():
            (root / name).write_text("preserved notice")
        commands = [
            {"type": "module", "filename": "one.wasm"},
            {"type": "assert_return"},
            {"type": "assert_invalid", "filename": "invalid.wasm", "module_type": "binary"},
            {"type": "assert_malformed", "filename": "bad.wasm", "module_type": "binary"},
            {"type": "assert_malformed", "filename": "text.wat", "module_type": "text"},
            {"type": "register"}, {"type": "action"},
        ]
        script = {"source_filename": "simple.wast", "commands": commands}
        for command in commands:
            if "filename" in command:
                (root / "simple" / command["filename"]).write_bytes(b"fixture")
        (root / "simple/script.json").write_text(json.dumps(script))
        records = self.root / "records"
        records.mkdir()
        (records / "pins.json").write_text(json.dumps(
            {"schemaVersion": 1, "profile": "Core 2.0 binary", "source": {"scripts": 1}}))
        for name in corpus.LICENSES.values():
            (records / name).write_bytes((root / name).read_bytes())
        sources = {"simple.wast": "0" * 64}
        manifest, files = corpus.describe(root, sources, corpus.digest(records / "pins.json"))
        (records / "manifest.json").write_text(json.dumps(manifest))
        (records / "files.sha256").write_text(corpus.hash_text(files))
        return root, records, sources, manifest

    def test_classification_and_verified_reuse(self):
        root, records, _, manifest = self.fixture()
        self.assertEqual(manifest["commandCount"], 7)
        self.assertEqual(manifest["binaryAssertions"], 3)
        self.assertEqual(manifest["textParsingAssertions"], 1)
        self.assertEqual(manifest["fileCount"], 7)
        self.assertEqual(manifest["commands"], dict(Counter(
            ["module", "assert_return", "assert_invalid", "assert_malformed", "assert_malformed", "register", "action"])))
        self.assertEqual(corpus.verify(root, records), manifest)
        self.assertEqual(corpus.prepare(root.parent, records), manifest)
        with self.assertRaisesRegex(ValueError, "fresh output"):
            corpus.prepare(root.parent, records, update=True)

    def test_changed_fixture_extra_file_and_counts_fail(self):
        root, records, _, manifest = self.fixture()
        file = root / "simple/one.wasm"
        file.write_bytes(b"changed")
        with self.assertRaisesRegex(ValueError, "file hashes"):
            corpus.verify(root, records)
        file.write_bytes(b"fixture")
        extra = root / "extra.wasm"
        extra.write_bytes(b"unexpected")
        with self.assertRaisesRegex(ValueError, "file hashes"):
            corpus.verify(root, records)
        extra.unlink()
        manifest["binaryAssertions"] += 1
        (records / "manifest.json").write_text(json.dumps(manifest))
        with self.assertRaisesRegex(ValueError, "assertion counts"):
            corpus.verify(root, records)

    def test_missing_files_unknown_commands_and_paths_fail(self):
        root, _, sources, _ = self.fixture()
        script_file = root / "simple/script.json"
        original = script_file.read_text()
        for modification, reason in [
            (lambda s: s.update(source_filename="other.wast"), "source identity"),
            (lambda s: s["commands"][0].update(type="future_command"), "unknown script"),
            (lambda s: s["commands"][0].update(filename="../escape.wasm"), "invalid corpus path"),
            (lambda s: s["commands"][0].update(module_type="unknown"), "module type"),
        ]:
            script = json.loads(original)
            modification(script)
            script_file.write_text(json.dumps(script))
            with self.subTest(reason=reason), self.assertRaisesRegex(ValueError, reason):
                corpus.describe(root, sources, "pins")
        script_file.write_text(original)
        (root / "simple/one.wasm").unlink()
        with self.assertRaisesRegex(ValueError, "inventory mismatch"):
            corpus.describe(root, sources, "pins")

    def test_changed_pins_and_symlinks_fail(self):
        root, records, _, _ = self.fixture()
        (records / "pins.json").write_text('{"schemaVersion":2}')
        with self.assertRaisesRegex(ValueError, "pins changed"):
            corpus.verify(root, records)
        (root / "outside").symlink_to(self.root, target_is_directory=True)
        with self.assertRaisesRegex(ValueError, "non-regular corpus entry"):
            corpus.inventory(root)

    def test_changed_license_record_fails(self):
        root, records, _, _ = self.fixture()
        (records / "LICENSE.tests").write_text("changed notice")
        with self.assertRaisesRegex(ValueError, "committed license notice"):
            corpus.verify(root, records)

    def test_pinned_producer_corrections(self):
        original = b"\x00asm\x01\x00\x00\x00\x0a\x01\x00"
        corrected = original[:8] + b"\x0c\x01\x00" + original[8:]
        fix = {"beforeSHA256": hashlib.sha256(original).hexdigest(),
               "afterSHA256": hashlib.sha256(corrected).hexdigest(),
               "offset": 8, "insertHex": "0c0100"}
        path = self.root / "input.wasm"
        path.write_bytes(original)
        corpus.correct_producer_binaries(self.root, {path.name: fix})
        self.assertEqual(path.read_bytes(), corrected)
        with self.assertRaisesRegex(ValueError, "correction input changed"):
            corpus.correct_producer_binaries(self.root, {path.name: fix})
        for change in [{"offset": 7}, {"offset": 100}, {"insertHex": "00"}]:
            path.write_bytes(original)
            with self.assertRaises(ValueError):
                corpus.correct_producer_binaries(self.root, {path.name: {**fix, **change}})
            self.assertEqual(path.read_bytes(), original)

    def test_reduced_script_inventory_fails_even_with_matching_pin_hash(self):
        root, records, _, manifest = self.fixture()
        pins = json.loads((records / "pins.json").read_text())
        pins["source"]["scripts"] = 2
        (records / "pins.json").write_text(json.dumps(pins))
        manifest["pinsSHA256"] = corpus.digest(records / "pins.json")
        (records / "manifest.json").write_text(json.dumps(manifest))
        with self.assertRaisesRegex(ValueError, "source script inventory"):
            corpus.verify(root, records)


if __name__ == "__main__":
    unittest.main()
