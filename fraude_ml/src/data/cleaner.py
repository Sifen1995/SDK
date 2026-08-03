import re
import unicodedata
from typing import Dict, Any


class TextCleaner:
    """Standardizes, normalizes, and cleans raw SMS strings for ML processing."""

    def __init__(self):
        # Amharic Unicode block range: \u1200-\u137F
        self.amharic_pattern = re.compile(r"[\u1200-\u137F]")
        self.url_pattern = re.compile(
            r"(https?://[^\s]+|www\.[^\s]+|[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}/[^\s]*)"
        )
        self.phone_pattern = re.compile(
            r"(\+?251|0)?[97]\d{8}"
        )  # Ethiopian phone pattern

    def normalize_unicode(self, text: str) -> str:
        """Normalizes Unicode characters (NFC form) to collapse script variations."""
        if not text:
            return ""
        return unicodedata.normalize("NFC", text)

    def clean_text(self, text: str, preserve_urls: bool = False) -> str:
        """
        Applies standard NLP cleaning pipeline:
        1. Unicode normalization
        2. Lowercasing
        3. URL & Phone placeholder masking (optional)
        4. Whitespace collapsing
        """
        if not isinstance(text, str) or not text.strip():
            return ""

        # Step 1: Normalize Unicode
        text = self.normalize_unicode(text)

        # Step 2: Lowercase
        text = text.lower()

        # Step 3: Replace URLs and Phone numbers with generic tokens if requested
        if not preserve_urls:
            text = self.url_pattern.sub(" [URL] ", text)
            text = self.phone_pattern.sub(" [PHONE] ", text)

        # Step 4: Remove non-printable control characters
        text = "".join(ch for ch in text if unicodedata.category(ch)[0] != "C")

        # Step 5: Collapse multiple spaces and newlines
        text = re.sub(r"\s+", " ", text).strip()

        return text

    def process_sample(self, raw_text: str) -> Dict[str, Any]:
        """Cleans input and extracts quick metadata flags."""
        cleaned = self.clean_text(raw_text, preserve_urls=False)
        return {
            "cleaned_text": cleaned,
            "has_amharic": 1 if self.amharic_pattern.search(raw_text) else 0,
            "char_count": len(cleaned),
            "word_count": len(cleaned.split()),
        }