"""Exercise the actual pinned bundle's inventory and tamper checks."""
import json
from pathlib import Path
import tempfile
import unittest

import core3_corpus as corpus


class Core3CorpusTests(unittest.TestCase):
    def test_complete_archive_roundtrip_and_tamper(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            manifest = corpus.prepare(output)
            self.assertEqual([manifest[k] for k in ("scriptCount", "commandCount", "binaryAssertions", "textParserAssertions")],
                             [259, 65344, 61395, 1239])
            self.assertEqual(len(manifest["producerCorrections"]), 1)
            self.assertEqual(len(manifest["textTranscodes"]), 10)
            evidence = json.loads((output / "evidence.json").read_text())
            self.assertEqual(evidence["manifestSHA256"], corpus.digest(corpus.RECORDS / "manifest.json"))
            fixture = output / "fixtures" / "gc__type-subtyping" / "script.88.wasm"
            fixture.write_bytes(fixture.read_bytes() + b"\0")
            with self.assertRaisesRegex(ValueError, "checksum"):
                corpus.verify(output / "fixtures")


if __name__ == "__main__":
    unittest.main()
