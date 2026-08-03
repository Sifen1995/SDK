"""Inference API for SMS spam and browsing URL scam detection.

Both paths share the same hybrid model trained in ``train.py``:
  [TF-IDF text features] + [structural / URL heuristic features]

- ``predict_message`` — inspect an SMS / notification body
- ``predict_url``     — inspect a URL during browser navigation
  (URL is scored as text so TF-IDF + URL heuristics both fire)
"""

from __future__ import annotations

import os
from typing import Any, Dict, List, Optional, Union

import joblib
import numpy as np
import pandas as pd
from scipy.sparse import hstack

from src.features.build_features import FeatureExtractor


class ScamDetector:
    """Loads serialized artifacts and scores messages or URLs."""

    DEFAULT_THRESHOLD = 0.50

    def __init__(
        self,
        artifact_dir: str = "models/artifacts",
        threshold: float = DEFAULT_THRESHOLD,
    ):
        self.artifact_dir = artifact_dir
        self.threshold = float(threshold)
        self.feature_extractor = FeatureExtractor()

        model_path = os.path.join(artifact_dir, "scam_detector.joblib")
        vectorizer_path = os.path.join(artifact_dir, "tfidf_vectorizer.joblib")
        config_path = os.path.join(artifact_dir, "feature_config.joblib")

        for path in (model_path, vectorizer_path, config_path):
            if not os.path.exists(path):
                raise FileNotFoundError(
                    f"Missing artifact: {path}. Run `python -m src.models.train` first."
                )

        self.model = joblib.load(model_path)
        self.vectorizer = joblib.load(vectorizer_path)
        self.config = joblib.load(config_path)
        self.tabular_feature_names: List[str] = list(
            self.config.get("tabular_features", [])
        )

    # ------------------------------------------------------------------
    # Feature building (must mirror ModelTrainer.prepare_feature_matrix)
    # ------------------------------------------------------------------
    def build_feature_matrix(self, texts: List[str]):
        """Build the hybrid sparse matrix for one or more raw strings."""
        cleaned = [
            self.feature_extractor.cleaner.clean_text(t, preserve_urls=True)
            for t in texts
        ]
        tfidf_matrix = self.vectorizer.transform(cleaned)

        df = pd.DataFrame({"text": texts})
        tabular_df = self.feature_extractor.transform_dataframe(df, text_column="text")

        # Keep column order identical to training when config is present
        if self.tabular_feature_names:
            missing = [c for c in self.tabular_feature_names if c not in tabular_df.columns]
            for col in missing:
                tabular_df[col] = 0
            tabular_df = tabular_df[self.tabular_feature_names]

        return hstack([tfidf_matrix, tabular_df.values]).tocsr()

    def feature_vector_dense(self, text: str) -> np.ndarray:
        """Dense float32 row used by the TFLite export path."""
        matrix = self.build_feature_matrix([text])
        return np.asarray(matrix.toarray(), dtype=np.float32)[0]

    @property
    def input_size(self) -> int:
        n_tfidf = len(self.vectorizer.get_feature_names_out())
        n_tab = len(self.tabular_feature_names) or 16
        return n_tfidf + n_tab

    # ------------------------------------------------------------------
    # Prediction helpers
    # ------------------------------------------------------------------
    def _score(self, texts: List[str]) -> List[Dict[str, Any]]:
        X = self.build_feature_matrix(texts)
        probs = self.model.predict_proba(X)[:, 1]
        preds = (probs >= self.threshold).astype(int)

        results: List[Dict[str, Any]] = []
        for text, prob, pred in zip(texts, probs, preds):
            results.append(
                {
                    "input": text,
                    "is_spam": bool(pred),
                    "label": "spam" if pred else "ham",
                    "spam_probability": float(round(prob, 6)),
                    "threshold": self.threshold,
                }
            )
        return results

    def predict_message(
        self, text: Union[str, List[str]], threshold: Optional[float] = None
    ) -> Union[Dict[str, Any], List[Dict[str, Any]]]:
        """Score one or more SMS / notification messages."""
        if threshold is not None:
            prev = self.threshold
            self.threshold = float(threshold)
            try:
                return self._predict_message(text)
            finally:
                self.threshold = prev
        return self._predict_message(text)

    def _predict_message(
        self, text: Union[str, List[str]]
    ) -> Union[Dict[str, Any], List[Dict[str, Any]]]:
        single = isinstance(text, str)
        texts = [text] if single else list(text)
        if not texts or any(not isinstance(t, str) or not t.strip() for t in texts):
            raise ValueError("text must be a non-empty string or list of strings")

        scored = self._score(texts)
        for item in scored:
            item["mode"] = "message"
            item["url_features"] = self.feature_extractor.extract_url_features(
                self._first_url(item["input"]) or ""
            )
        return scored[0] if single else scored

    def predict_url(
        self, url: Union[str, List[str]], threshold: Optional[float] = None
    ) -> Union[Dict[str, Any], List[Dict[str, Any]]]:
        """Score one or more URLs (browser navigation / link taps).

        The URL is fed through the same hybrid pipeline as SMS text so both
        TF-IDF tokens and URL heuristics contribute to the score.
        """
        if threshold is not None:
            prev = self.threshold
            self.threshold = float(threshold)
            try:
                return self._predict_url(url)
            finally:
                self.threshold = prev
        return self._predict_url(url)

    def _predict_url(
        self, url: Union[str, List[str]]
    ) -> Union[Dict[str, Any], List[Dict[str, Any]]]:
        single = isinstance(url, str)
        urls = [url] if single else list(url)
        if not urls or any(not isinstance(u, str) or not u.strip() for u in urls):
            raise ValueError("url must be a non-empty string or list of strings")

        # Normalize bare domains so TF-IDF / urlparse behave consistently
        normalized = [
            u if u.startswith(("http://", "https://")) else f"http://{u}" for u in urls
        ]
        scored = self._score(normalized)
        for raw, item in zip(urls, scored):
            item["mode"] = "url"
            item["input"] = raw
            item["normalized_url"] = (
                raw if raw.startswith(("http://", "https://")) else f"http://{raw}"
            )
            item["url_features"] = self.feature_extractor.extract_url_features(
                item["normalized_url"]
            )
        return scored[0] if single else scored

    @staticmethod
    def _first_url(text: str) -> Optional[str]:
        extractor = FeatureExtractor()
        matches = extractor.cleaner.url_pattern.findall(text)
        if not matches:
            return None
        first = matches[0]
        return first[0] if isinstance(first, tuple) else first


def load_detector(artifact_dir: str = "models/artifacts") -> ScamDetector:
    """Convenience factory used by tests and the TFLite exporter."""
    return ScamDetector(artifact_dir=artifact_dir)


if __name__ == "__main__":
    detector = load_detector()

    sms = "Urgent Telebirr Alert: wallet suspended. Verify now http://telebirr-kyc.xyz/login"
    print("message →", detector.predict_message(sms))

    url = "http://192.168.1.105/cbe-sec-login"
    print("url     →", detector.predict_url(url))
