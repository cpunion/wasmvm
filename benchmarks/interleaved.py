"""Serialize warm calls across fresh processes without timing their barriers."""
import json
import os
import selectors
import subprocess
import time

MAX_OUTPUT = 2 * 1024 * 1024


class Child:
    def __init__(self, command, deadline):
        self.command, self.deadline = command, deadline
        self.started = time.perf_counter_ns()
        self.output = {"stdout": bytearray(), "stderr": bytearray()}
        self.pending = bytearray()
        self.process = subprocess.Popen(command, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                        stderr=subprocess.PIPE, bufsize=0,
                                        env=dict(os.environ, WASMVM_INTERLEAVE="1"))
        self.selector = selectors.DefaultSelector()
        self.selector.register(self.process.stdout, selectors.EVENT_READ, "stdout")
        self.selector.register(self.process.stderr, selectors.EVENT_READ, "stderr")

    def remaining(self):
        remaining = self.deadline - time.monotonic()
        if remaining <= 0:
            raise TimeoutError("interleaved block exceeded its deadline")
        return remaining

    def read(self):
        for key, _ in self.selector.select(self.remaining()):
            chunk = os.read(key.fileobj.fileno(), 16384)
            if not chunk:
                self.selector.unregister(key.fileobj)
                continue
            self.output[key.data].extend(chunk)
            if sum(map(len, self.output.values())) > MAX_OUTPUT:
                raise ValueError("interleaved runner exceeded output limit")
            if key.data == "stdout":
                self.pending.extend(chunk)

    def event(self):
        while b"\n" not in self.pending:
            if not self.selector.get_map():
                raise ValueError("interleaved runner ended before expected event: " +
                                 self.output["stderr"].decode(errors="replace"))
            self.read()
        line, _, rest = self.pending.partition(b"\n")
        self.pending = bytearray(rest)
        return json.loads(line)

    def expect(self, expected):
        if self.event() != expected:
            raise ValueError(f"interleaved runner expected {expected}")

    def send(self, token):
        try:
            self.process.stdin.write(token)
        except BrokenPipeError as error:
            raise ValueError("interleaved runner closed before acknowledgement") from error

    def finish(self, ready_ns, record):
        self.process.stdin.close()
        while self.selector.get_map():
            self.read()
        self.process.wait(timeout=self.remaining())
        stderr = self.output["stderr"].decode(errors="replace")
        if self.process.returncode:
            raise RuntimeError(f"interleaved runner exit {self.process.returncode}: {stderr}")
        if self.pending:
            raise ValueError("interleaved runner emitted extra output")
        return {"processReadyNs": ready_ns, "processWallNs": time.perf_counter_ns()-self.started,
                "record": record, "exitCode": self.process.returncode, "stderr": stderr,
                "command": self.command, "protocolEvents": self.output["stdout"].decode()}

    def close(self):
        if self.process.poll() is None:
            self.process.kill()
        self.process.wait()
        self.selector.close()
        for stream in (self.process.stdin, self.process.stdout, self.process.stderr):
            stream.close()


def invoke(commands, timeout, rng, samples=7):
    """Return one process block; individual calls are not independent blocks."""
    children, readiness, observations, orders = {}, {}, {}, []
    deadline = time.monotonic() + timeout
    names = list(commands)
    rng.shuffle(names)
    try:
        # Preparation and warmups also run serially, before any measured call.
        for name in names:
            child = children[name] = Child(commands[name], deadline)
            child.expect({"ready": True})
            readiness[name] = time.perf_counter_ns()-child.started
            child.send(b"run\n")
            child.expect({"sampleReady": 0})
        for sample in range(samples):
            rng.shuffle(names)
            orders.append(list(names))
            for name in names:
                child = children[name]
                child.send(b"sample\n")
                if sample+1 < samples:
                    child.expect({"sampleReady": sample+1})
                else:
                    record = child.event()
                    if not isinstance(record, dict) or "warmCallNs" not in record:
                        raise ValueError("interleaved runner omitted final record")
                    observations[name] = child.finish(readiness[name], record)
        return observations, orders
    finally:
        for child in children.values():
            child.close()
