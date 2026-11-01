import copy
import json
from pathlib import Path
import unittest

from ci_shards import ROOT, validate


class ShardsTest(unittest.TestCase):
    def setUp(self):
        self.shards = json.loads((ROOT / '.github/llgo-shards.json').read_text())

    def test_complete_repository_plan(self):
        self.assertEqual(validate(self.shards), 31)

    def test_missing_duplicate_and_unknown_packages(self):
        for mode in ('missing', 'duplicate', 'unknown'):
            with self.subTest(mode=mode):
                shards = copy.deepcopy(self.shards)
                paths = shards[0]['packages'].split()
                if mode == 'missing':
                    paths.pop()
                elif mode == 'duplicate':
                    paths.append(paths[0])
                else:
                    paths.append('./internal/unlisted')
                shards[0]['packages'] = ' '.join(paths)
                with self.assertRaises(ValueError):
                    validate(shards)

    def test_duplicate_names_bad_inputs_and_missing_cli(self):
        for mode in ('name', 'inputs', 'cli'):
            with self.subTest(mode=mode):
                shards = copy.deepcopy(self.shards)
                if mode == 'name':
                    shards[1]['name'] = shards[0]['name']
                elif mode == 'inputs':
                    shards[0]['inputs'] = 'untrusted-command'
                else:
                    for shard in shards:
                        shard['cli'] = False
                with self.assertRaises(ValueError):
                    validate(shards)


if __name__ == '__main__':
    unittest.main()
