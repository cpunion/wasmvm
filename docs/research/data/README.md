# Research data

Small records are ordinary JSON. Large historical records use a JSON pointer to
an XZ archive in `../evidence`. The pointer records both compressed and original
SHA-256 digests and sizes. Decompression restores the complete original bytes,
including unsuccessful and inconclusive observations; no samples are discarded.

Read a pointer from the repository root:

```python
import hashlib, json, lzma
from pathlib import Path

pointer = Path("docs/research/data/loop-value-coalescing-2026-11-05.json")
info = json.loads(pointer.read_text())
packed = (pointer.parent / info["archive"]).read_bytes()
assert hashlib.sha256(packed).hexdigest() == info["archiveSHA256"]
raw = lzma.decompress(packed)
assert hashlib.sha256(raw).hexdigest() == info["originalSHA256"]
record = json.loads(raw)
```

Source revisions, targets, compiler profiles and measurement events remain
independent. Dated research evidence does not imply current milestone acceptance.
