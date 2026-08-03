"""Unit tests for fraude_ml cleaner, features, and ScamDetector inference.

Run from ``fraude_ml/``:

    pytest tests/ -q
"""

from __future__ import annotations

import os
import sys

import numpy as np
import pytest

# Ensure package imports resolve when pytest is launched from fraude_ml/
ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
if ROOT not in sys.path:
    sys.path.insert(0, ROOT)

from src.data.cleaner import TextCleaner
from src.data.sanitize_threat_feed import REPLACEMENT, sanitize_text
from src.features.build_features import FeatureExtractor
from src.models.predict import ScamDetector


ARTIFACT_DIR = os.path.join(ROOT, "models", "artifacts")
HAS_ARTIFACTS = all(
    os.path.exists(os.path.join(ARTIFACT_DIR, name))
    for name in (
        "scam_detector.joblib",
        "tfidf_vectorizer.joblib",
        "feature_config.joblib",
    )
)

requires_artifacts = pytest.mark.skipif(
    not HAS_ARTIFACTS,
    reason="Train the model first (python -m src.models.train)",
)


def test_sanitize_threat_feed_redacts_telegram_token():
    token = "123456789:" + "A" * 35
    sanitized, count = sanitize_text(
        f"https://api.telegram.org/bot{token}/getChat?chat_id=123"
    )

    assert count == 1
    assert token not in sanitized
    assert REPLACEMENT in sanitized


class TestTextCleaner:
    def setup_method(self):
        self.cleaner = TextCleaner()

    def test_empty_returns_empty(self):
        assert self.cleaner.clean_text("") == ""
        assert self.cleaner.clean_text("   ") == ""

    def test_lowercases_and_collapses_whitespace(self):
        assert self.cleaner.clean_text("Hello   WORLD") == "hello world"

    def test_masks_url_and_phone_by_default(self):
        text = "Call 0911234567 or visit https://evil.example/path now"
        cleaned = self.cleaner.clean_text(text, preserve_urls=False)
        assert "[URL]" in cleaned
        assert "[PHONE]" in cleaned
        assert "https://" not in cleaned

    def test_preserve_urls_keeps_link(self):
        text = "visit https://evil.example/path now"
        cleaned = self.cleaner.clean_text(text, preserve_urls=True)
        assert "https://evil.example/path" in cleaned

    def test_process_sample_flags_amharic(self):
        sample = self.cleaner.process_sample("ማስጠንቀቂያ Telebirr")
        assert sample["has_amharic"] == 1
        assert sample["word_count"] >= 1


class TestFeatureExtractor:
    def setup_method(self):
        self.extractor = FeatureExtractor()

    def test_url_features_high_risk_tld(self):
        feats = self.extractor.extract_url_features(
            "http://telebirr-verify-kyc.xyz/login.php"
        )
        assert feats["has_url"] == 1
        assert feats["is_high_risk_tld"] == 1
        assert feats["has_hyphen_in_domain"] == 1
        assert feats["domain_entropy"] > 0

    def test_url_features_ip_address(self):
        feats = self.extractor.extract_url_features("http://192.168.1.105/cbe-sec")
        assert feats["ip_in_url"] == 1
        assert feats["has_url"] == 1

    def test_url_features_shortener(self):
        feats = self.extractor.extract_url_features("http://bit.ly/ethio-bonus")
        assert feats["is_url_shortener"] == 1

    def test_empty_url_returns_zeros(self):
        feats = self.extractor.extract_url_features("")
        assert feats["has_url"] == 0
        assert feats["url_length"] == 0

    def test_message_features_include_url_and_urgency(self):
        sms = "Urgent Telebirr Alert: verify now http://telebirr-kyc.xyz/login"
        feats = self.extractor.extract_message_features(sms)
        assert feats["has_url"] == 1
        assert feats["has_urgency_word"] == 1
        assert feats["has_brand_mention"] == 1
        assert "text_length" in feats
        assert "domain_entropy" in feats

    def test_shannon_entropy_known_value(self):
        # "aa" → entropy 0; "ab" → 1.0
        assert FeatureExtractor.calculate_shannon_entropy("aa") == 0.0
        assert FeatureExtractor.calculate_shannon_entropy("ab") == 1.0


@pytest.fixture(scope="module")
def detector():
    if not HAS_ARTIFACTS:
        pytest.skip("Train the model first (python -m src.models.train)")
    return ScamDetector(artifact_dir=ARTIFACT_DIR)


@requires_artifacts
class TestScamDetector:

    def test_predict_message_spam_like(self, detector):
        result = detector.predict_message(
            "Urgent Telebirr Alert: Your wallet is suspended. Verify now: "
            "http://telebirr-verify-kyc.xyz/login.php"
        )
        assert result["mode"] == "message"
        assert "spam_probability" in result
        assert 0.0 <= result["spam_probability"] <= 1.0
        assert result["label"] in ("spam", "ham")
        # Localized phishing template should score as spam
        assert result["is_spam"] is True

    def test_predict_message_ham_like(self, detector):
        result = detector.predict_message(
            "Telebirr: You have successfully bought 2GB 7-Days Data Package. "
            "Transaction ID: TX12345678. Thank you for using Ethio Telecom."
        )
        assert result["mode"] == "message"
        assert result["is_spam"] is False

    def test_predict_url_phishing(self, detector):
        result = detector.predict_url("http://telebirr-verify-kyc.xyz/login.php")
        assert result["mode"] == "url"
        assert result["url_features"]["is_high_risk_tld"] == 1
        assert 0.0 <= result["spam_probability"] <= 1.0

    def test_predict_url_normalizes_bare_domain(self, detector):
        result = detector.predict_url("example.com/path")
        assert result["normalized_url"].startswith("http://")

    def test_batch_predict_message(self, detector):
        results = detector.predict_message(
            [
                "win free money http://scam.xyz/claim",
                "Your Ethio Telecom package will expire in 2 days.",
            ]
        )
        assert isinstance(results, list)
        assert len(results) == 2

    def test_empty_input_raises(self, detector):
        with pytest.raises(ValueError):
            detector.predict_message("   ")
        with pytest.raises(ValueError):
            detector.predict_url("")

    def test_feature_vector_dense_shape(self, detector):
        vec = detector.feature_vector_dense("hello http://example.com")
        assert isinstance(vec, np.ndarray)
        assert vec.dtype == np.float32
        assert vec.ndim == 1
        assert vec.shape[0] == detector.input_size
