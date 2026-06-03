"""Load trained behavioral intent model artifact."""

from __future__ import annotations

import os
from pathlib import Path

import joblib

DEFAULT_MODEL_PATH = Path(__file__).resolve().parent.parent / "models" / "intent_model.pkl"


class ModelArtifact:
    def __init__(self, data: dict):
        self.model = data["model"]
        self.feature_columns: list[str] = data["feature_columns"]
        self.threshold: float = float(data.get("threshold", 0.7))
        self.intents: list[str] = data.get("intents", [])
        self.model_version: str = data.get("model_version", "unknown")


def load_model(path: str | Path | None = None) -> ModelArtifact:
    model_path = Path(path or os.getenv("MODEL_PATH", DEFAULT_MODEL_PATH))
    if not model_path.exists():
        raise FileNotFoundError(
            f"Model not found at '{model_path}'. Run: python -m training.train"
        )
    data = joblib.load(model_path)
    # Backward compat with TF-IDF pipeline artifact
    if "pipeline" in data and "model" not in data:
        raise RuntimeError(
            "Legacy text-based model detected. Retrain with: python -m training.train"
        )
    return ModelArtifact(data)
