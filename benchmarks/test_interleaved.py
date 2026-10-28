import json
from pathlib import Path
import random
import sys
import tempfile
import unittest

import interleaved


class InterleavedTests(unittest.TestCase):
    def test_calls_are_serial_and_every_process_has_every_sample(self):
        with tempfile.TemporaryDirectory() as directory:
            lock = str(Path(directory) / "executing")
            script = '''
import json, os, sys, time
assert os.environ["WASMVM_INTERLEAVE"] == "1"
print('{"ready":true}', flush=True)
assert input() == "run"
for n in range(7):
    print(json.dumps({"sampleReady": n}), flush=True)
    assert input() == "sample"
    fd = os.open(sys.argv[1], os.O_CREAT | os.O_EXCL | os.O_WRONLY)
    time.sleep(.001)
    os.close(fd)
    os.unlink(sys.argv[1])
print(json.dumps({"warmCallNs": list(range(1, 8)), "identity": sys.argv[2]}))
'''
            commands = {name: [sys.executable, "-c", script, lock, name] for name in ("candidate", "control", "reference")}
            rows, orders = interleaved.invoke(commands, 10, random.Random(42))
            self.assertEqual(set(rows), set(commands))
            self.assertEqual(len(orders), 7)
            self.assertGreater(len(set(map(tuple, orders))), 1)
            for order in orders:
                self.assertEqual(set(order), set(commands))
            for name, row in rows.items():
                self.assertEqual(row["record"], {"warmCallNs": list(range(1, 8)), "identity": name})
                self.assertEqual(len(row["protocolEvents"].splitlines()), 9)
                self.assertGreater(row["processWallNs"], row["processReadyNs"])

    def test_failures_never_produce_a_successful_block(self):
        prefix = 'print(\'{"ready":true}\', flush=True); input(); '
        samples = '''
for i in range(7):
    print(json.dumps({"sampleReady": i}), flush=True)
    input()
'''
        cases = [
            (ValueError, "print('{}')"),
            (ValueError, prefix + "print('{}')"),
            (ValueError, prefix + "pass"),
            (ValueError, prefix + "print('x' * (3 * 1024 * 1024))"),
            (ValueError, prefix + "print('{invalid}')"),
            (ValueError, prefix + "\n" + samples + "print('{}')"),
            (ValueError, prefix + "\n" + samples + 'print(\'{"warmCallNs":[]}\\nextra\')'),
            (RuntimeError, prefix + "\n" + samples + 'print(\'{"warmCallNs":[]}\'); sys.exit(2)'),
        ]
        for exception, script in cases:
            with self.subTest(script=script), self.assertRaises(exception):
                interleaved.invoke({"broken": [sys.executable, "-c", "import json, sys\n" + script]}, 10, random.Random(42))

    def test_deadline_kills_waiting_processes(self):
        with self.assertRaises(TimeoutError):
            interleaved.invoke({"stalled": [sys.executable, "-c", "import time; time.sleep(60)"]}, .1, random.Random(42))
