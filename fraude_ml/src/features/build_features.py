import math
import re
import sys
from pathlib import Path
from urllib.parse import urlparse
from typing import Dict, Any, List, Union
import pandas as pd

try:
    from ..data.cleaner import TextCleaner
except ImportError:
    # Allow: python3 build_features.py (no parent package)
    sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
    from data.cleaner import TextCleaner


class FeatureExtractor:
    """
    Dual-Purpose Feature Extractor:
    1. Extracts structural metrics and entropy from text messages.
    2. Extracts URL heuristics directly for live browser URL checking.
    """

    def __init__(self):
        self.cleaner = TextCleaner()
        # High-risk TLD list commonly exploited in phishing
        self.high_risk_tlds = {
            "xyz", "top", "sbs", "cfd", "buzz", "cc", "info", "work", "tk", "ga", "ml", "online"
        }
        # Common brand targets in local market
        self.brand_keywords = [
            "telebirr", "cbe", "ethiotel", "ethio telecom", "commercial bank", "dashen", "boa"
        ]
        # Urgent trigger words
        self.urgency_keywords = [
            "urgent", "suspended", "locked", "blocked", "verify", "update", "bonus", "reward", "claim",
            "ማስጠንቀቂያ", "ታግዷል", "ይረጋገጡ", "ስጦታ", "አሸንፈዋል"
        ]

    # --- Domain Entropy Calculation ---
    @staticmethod
    def calculate_shannon_entropy(text: str) -> float:
        """
        Calculates Shannon Entropy (randomness score) of a given string.
        High entropy indicates random bot-generated domains (e.g., 'a7x-9z-claim.xyz').
        """
        if not text:
            return 0.0

        text = text.lower()
        length = len(text)
        frequency = {}

        for char in text:
            frequency[char] = frequency.get(char, 0) + 1

        entropy = 0.0
        for count in frequency.values():
            p = count / length
            entropy -= p * math.log2(p)

        return round(entropy, 4)

    # --- Standalone URL Feature Extractor (Used by Browser Shield & SMS Inspector) ---
    def extract_url_features(self, url: str) -> Dict[str, Any]:
        """
        Inspects a raw URL string and generates domain security features.
        Can be called directly for live browser navigation checks.
        """
        if not isinstance(url, str) or not url.strip():
            return {
                "has_url": 0,
                "url_count": 0,
                "url_length": 0,
                "domain_entropy": 0.0,
                "is_high_risk_tld": 0,
                "is_url_shortener": 0,
                "ip_in_url": 0,
                "subdomain_count": 0,
                "has_hyphen_in_domain": 0,
            }

        url = url.strip()
        parsed = urlparse(url if url.startswith(("http://", "https://")) else f"http://{url}")
        domain = parsed.netloc or parsed.path.split("/")[0]

        # Clean port if present
        domain_name = domain.split(":")[0]

        # Extract TLD
        domain_parts = domain_name.split(".")
        tld = domain_parts[-1].lower() if len(domain_parts) > 1 else ""

        # Check for IP address in URL
        ip_pattern = re.compile(r"^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$")
        is_ip = 1 if ip_pattern.match(domain_name) else 0

        # Subdomain count
        subdomain_count = max(0, len(domain_parts) - 2)

        # Check domain entropy (excluding standard TLD)
        main_domain = domain_parts[-2] if len(domain_parts) >= 2 else domain_name
        domain_entropy = self.calculate_shannon_entropy(main_domain)

        return {
            "has_url": 1,
            "url_count": 1,
            "url_length": len(url),
            "domain_entropy": domain_entropy,
            "is_high_risk_tld": 1 if tld in self.high_risk_tlds else 0,
            "is_url_shortener": 1 if domain_name in ["bit.ly", "tinyurl.com", "t.co", "cutt.ly"] else 0,
            "ip_in_url": is_ip,
            "subdomain_count": subdomain_count,
            "has_hyphen_in_domain": 1 if "-" in main_domain else 0,
        }

    # --- Combined SMS Feature Extractor ---
    def extract_message_features(self, raw_text: str) -> Dict[str, Any]:
        """Extracts complete feature set from an incoming SMS message."""
        cleaned_text = self.cleaner.clean_text(raw_text, preserve_urls=True)

        # 1. Text Structural Metrics
        text_length = len(raw_text)
        word_count = len(raw_text.split())
        digit_count = sum(c.isdigit() for c in raw_text)
        uppercase_count = sum(c.isupper() for c in raw_text)

        # 2. Extract Embedded URLs
        urls = self.cleaner.url_pattern.findall(raw_text)
        if urls:
            # Flatten extracted URL regex tuple/matches if any
            first_url = urls[0][0] if isinstance(urls[0], tuple) else urls[0]
            url_features = self.extract_url_features(first_url)
            url_features["url_count"] = len(urls)
        else:
            url_features = self.extract_url_features("")

        # 3. Domain/Brand & Urgency Keywords
        has_brand = 1 if any(b in cleaned_text for b in self.brand_keywords) else 0
        has_urgency = 1 if any(u in cleaned_text for u in self.urgency_keywords) else 0

        # 4. Amharic Script Flag
        has_amharic = 1 if self.cleaner.amharic_pattern.search(raw_text) else 0

        # Combine into flat feature dictionary
        features = {
            "text_length": text_length,
            "word_count": word_count,
            "digit_count": digit_count,
            "uppercase_count": uppercase_count,
            "has_brand_mention": has_brand,
            "has_urgency_word": has_urgency,
            "has_amharic": has_amharic,
            **url_features,
        }
        return features

    def transform_dataframe(self, df: pd.DataFrame, text_column: str = "text") -> pd.DataFrame:
        """Transforms a DataFrame containing raw text into a complete numerical feature matrix."""
        feature_list = [self.extract_message_features(text) for text in df[text_column]]
        return pd.DataFrame(feature_list)


if __name__ == "__main__":
    extractor = FeatureExtractor()

    # Test 1: SMS Message Feature Extraction
    sample_sms = "ማስጠንቀቂያ: የ Telebirr መለያዎ ታግዷል። verify now http://telebirr-kyc-verify.xyz/login"
    print("--- Test 1: SMS Features ---")
    print(extractor.extract_message_features(sample_sms))

    # Test 2: Live Browser URL Feature Extraction
    sample_url = "http://192.168.1.105/cbe-sec-login"
    print("\n--- Test 2: Live Browser URL Features ---")
    print(extractor.extract_url_features(sample_url))