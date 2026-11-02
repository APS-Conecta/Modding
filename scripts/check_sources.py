#!/usr/bin/env python3
"""Check imported bytes against the recorded gestion source manifest."""
import hashlib
import json
from pathlib import Path

root = Path(__file__).resolve().parents[1]
manifest = json.loads((root / "third_party/gestion/sources.json").read_text())
for item in manifest["files"]:
    actual = hashlib.sha256((root / item["path"]).read_bytes()).hexdigest()
    if actual != item["sha256"]:
        raise SystemExit(f"Imported source changed without provenance review: {item['path']}")
print(f"PASS: {len(manifest['files'])} imported files match their recorded hashes.")
