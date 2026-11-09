import copy
import json
import re
import unittest
from pathlib import Path
import subprocess
import sys
import tempfile

from ci_shards import ROOT, validate, validate_matrix


class ShardsTest(unittest.TestCase):
    def setUp(self):
        self.shards = json.loads((ROOT / '.github/llgo-shards.json').read_text())

    def test_complete_repository_plan(self):
        self.assertEqual(validate(self.shards), 32)

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

    def test_duplicate_names_inputs_and_shell_fragments(self):
        for mode in ('name', 'inputs', 'path'):
            with self.subTest(mode=mode):
                shards = copy.deepcopy(self.shards)
                if mode == 'name':
                    shards[1]['name'] = shards[0]['name']
                elif mode == 'inputs':
                    shards[0]['inputs'] = 'untrusted-command'
                else:
                    shards[0]['packages'] += ' ./pkg/wasm;echo'
                with self.assertRaises(ValueError):
                    validate(shards)


    def test_independent_matrix_preserves_every_shard(self):
        workflow = (ROOT / '.github/workflows/ci.yml').read_text()
        validate_matrix(self.shards, workflow)
        for replacement in ('[]', '["core", "core", "vm", "component", "http", "async"]', '["missing"]'):
            changed = re.sub(r'^        shard: \[[^\n]+\]$', '        shard: ' + replacement, workflow, flags=re.MULTILINE)
            with self.assertRaises(ValueError):
                validate_matrix(self.shards, changed)

    def test_selected_job_uses_the_authoritative_package_and_input_lists(self):
        with tempfile.TemporaryDirectory() as directory:
            for shard in self.shards:
                target = Path(directory) / shard['name']
                subprocess.run([sys.executable, str(ROOT / 'scripts/ci_shards.py'), '--shard', shard['name'], '--github-output', str(target)], check=True, capture_output=True)
                self.assertEqual(target.read_text(), 'packages=' + shard['packages'] + '\ninputs=' + shard['inputs'] + '\n')
            target = Path(directory) / 'invalid'
            failed = subprocess.run([sys.executable, str(ROOT / 'scripts/ci_shards.py'), '--shard', 'missing', '--github-output', str(target)], capture_output=True)
            self.assertNotEqual(failed.returncode, 0)
            self.assertFalse(target.exists())


if __name__ == '__main__':
    unittest.main()
