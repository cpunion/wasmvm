import tempfile
import unittest
from pathlib import Path

import preview1_corpus


class Preview1Inputs(unittest.TestCase):
    def test_offline_inventory_and_tamper_rejection(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary) / 'fixtures'
            manifest = preview1_corpus.prepare(root)
            self.assertEqual(len(manifest['tests']), 72)
            program = root / manifest['tests'][0]['path']
            program.write_bytes(program.read_bytes() + b'changed')
            with self.assertRaisesRegex(ValueError, 'checksum'):
                preview1_corpus.verify(root)


if __name__ == '__main__':
    unittest.main()
