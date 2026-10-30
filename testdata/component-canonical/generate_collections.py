"""Generate collection memory oracles using the unchanged pinned ABI reference."""
import importlib.util
import json
from pathlib import Path
import random
import sys
from types import SimpleNamespace

ROOT = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("canonical_collection_reference", ROOT / "definitions.py")
m = importlib.util.module_from_spec(spec)
sys.modules[spec.name] = m
spec.loader.exec_module(m)
rng = random.Random(20261030)

class Memory:
    def __init__(self, wide):
        self.wide = wide
        self.data = bytearray(4096)
    def ptr_type(self): return "i64" if self.wide else "i32"
    def ptr_size(self): return 8 if self.wide else 4
    def __len__(self): return len(self.data)
    def __getitem__(self, key): return self.data[key]
    def __setitem__(self, key, value): self.data[key] = value

class Context:
    def __init__(self, wide):
        self.opts = SimpleNamespace(memory=Memory(wide))
        self.next = 64
    def allocate(self, align, size):
        ptr = (self.next + align - 1) & -align
        self.next = ptr + size
        assert self.next <= len(self.opts.memory)
        return ptr

def primitive(bits):
    return {"Kind": {8: 125, 16: 123, 32: 121, 64: 119}[bits]}, getattr(m, "U" + str(bits) + "Type")(), lambda: ((v := rng.randrange(1 << bits)), {"Bits": v})

def record():
    a, at, av = primitive(16)
    b, bt, bv = primitive(32)
    def value():
        x, xj = av(); y, yj = bv()
        return {"a": x, "b": y}, {"Elements": [xj, yj]}
    return {"Kind": 114, "Fields": [{"Name": "a", "Type": a}, {"Name": "b", "Type": b}]}, m.RecordType([m.FieldType("a", at), m.FieldType("b", bt)]), value

rows = []
for wide in (False, True):
    for key_bits in (8, 32):
        for ej, et, ev in (primitive(64), record()):
            kj, kt, _ = primitive(key_bits)
            pair = m.despecialize(m.TupleType([kt, et]))
            for count in (0, 1, 3, 6):
                values, host = [], []
                for i in range(count):
                    v, vj = ev(); key = i % 2
                    values.append({pair.fields[0].label: key, pair.fields[1].label: v})
                    host.append({"Elements": [{"Bits": key}, vj]})
                typ = m.MapType(kt, et); context = Context(wide)
                m.store(context, values, typ, 0)
                assert m.load(context, 0, typ) == values
                rows.append({"Type": {"Kind": 99, "Key": kj, "Element": ej}, "Memory64": wide, "Value": {"Elements": host}, "Memory": context.opts.memory.data[:max(context.next, m.elem_size(typ, context.opts.memory.ptr_type()))].hex()})
    for ej, et, ev in (primitive(8), primitive(64), record()):
        for count in (1, 3, 8, 17):
            generated = [ev() for _ in range(count)]
            values, host = [v for v, _ in generated], [v for _, v in generated]
            typ = m.ListType(et, count); context = Context(wide)
            m.store(context, values, typ, 0)
            assert m.load(context, 0, typ) == values
            rows.append({"Type": {"Kind": 103, "Element": ej, "Length": count}, "Memory64": wide, "Value": {"Elements": host}, "Memory": context.opts.memory.data[:max(context.next, m.elem_size(typ, context.opts.memory.ptr_type()))].hex()})
(ROOT / "collection-oracle.json").write_text(json.dumps(rows, separators=(",", ":")) + "\n")
print(len(rows), "independent collection memory cases")
