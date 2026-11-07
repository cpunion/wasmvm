"""Independent streaming SHA-256 oracle backed by Python hashlib."""
import hashlib
import struct

MAX_BLOCKS = 65536

def checksum(seed, count):
    if type(seed) is not int or not 0 <= seed < 2**32:
        raise ValueError("seed must be u32")
    if type(count) is not int or not 1 <= count <= MAX_BLOCKS:
        raise ValueError("message block count must be 1..65536")
    state = hashlib.sha256()
    x = seed
    for _ in range(count):
        words = []
        for _ in range(16):
            x = (1664525 * x + 1013904223) & 0xffffffff
            words.append(x)
        state.update(struct.pack("<16I", *words))
    return int.from_bytes(state.digest()[:4], "big")
