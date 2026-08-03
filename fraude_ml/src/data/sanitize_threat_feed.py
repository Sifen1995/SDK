"""Remove credential-shaped values from downloaded threat-intelligence feeds.

Threat feeds can contain live credentials because malicious URLs sometimes embed
API tokens. GitHub secret scanning correctly detects those values even though
they belong to third parties. Run this sanitizer before committing a feed.
"""

from __future__ import annotations

import argparse
import re
from pathlib import Path


TELEGRAM_BOT_TOKEN = re.compile(
    r"(?<!\d)\d{8,10}:[A-Za-z0-9_-]{30,}(?![A-Za-z0-9_-])"
)
REPLACEMENT = "REDACTED_TELEGRAM_BOT_TOKEN"


def sanitize_text(text: str) -> tuple[str, int]:
    """Return sanitized text and the number of removed token values."""
    return TELEGRAM_BOT_TOKEN.subn(REPLACEMENT, text)


def sanitize_file(path: Path) -> int:
    """Sanitize a UTF-8-compatible text/CSV file without changing newlines."""
    original_bytes = path.read_bytes()
    original = original_bytes.decode("utf-8", errors="replace")
    sanitized, replacements = sanitize_text(original)
    if replacements:
        path.write_bytes(sanitized.encode("utf-8"))
    return replacements


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Redact embedded credentials from a threat-feed file."
    )
    parser.add_argument("path", type=Path)
    args = parser.parse_args()

    replacements = sanitize_file(args.path)
    print(f"Sanitized {replacements} credential value(s) in {args.path}")


if __name__ == "__main__":
    main()
