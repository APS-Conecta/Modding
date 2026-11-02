#!/usr/bin/env python3
"""Validate authored Markdown links and formatting; imported source stays byte-exact."""
import json
import re
from pathlib import Path
from urllib.parse import unquote, urlsplit

root = Path(__file__).resolve().parents[1]
imported = {item["path"] for item in json.loads((root / "third_party/gestion/sources.json").read_text())["files"]}
count = links = 0
for path in root.rglob("*.md"):
    relative = path.relative_to(root)
    if str(relative) in imported or any(part.startswith(".") and part != ".github" for part in relative.parts):
        continue
    count += 1
    content = path.read_text()
    assert content.endswith("\n"), f"{relative}: missing final newline"
    assert all(line == line.rstrip() for line in content.splitlines()), f"{relative}: trailing whitespace"
    assert sum(line.startswith("```") for line in content.splitlines()) % 2 == 0, f"{relative}: unclosed code fence"
    for target in re.findall(r"\[[^\]]*\]\(([^\s)]+)\)", content):
        parsed = urlsplit(target)
        if parsed.scheme or parsed.netloc or not parsed.path:
            continue
        links += 1
        resolved = (path.parent / unquote(parsed.path)).resolve()
        assert resolved.is_relative_to(root) and resolved.exists(), f"{relative}: broken link {target}"
print(f"PASS: {count} authored Markdown files and {links} relative links.")
