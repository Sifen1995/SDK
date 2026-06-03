"""Behavioral intent prediction from engineered features."""

from __future__ import annotations

from typing import Any

import pandas as pd

from inference.model_loader import ModelArtifact
from training.feature_engineering import build_feature_vector


def top_signals(features: dict[str, float], intent: str) -> list[str]:
    signals: list[tuple[str, float]] = []

    prefix = intent.replace("_interest", "")
    view_key = f"{prefix}_views"
    search_key = f"{prefix}_searches"
    if features.get(view_key, 0) > 0:
        signals.append((view_key, features[view_key]))
    if features.get(search_key, 0) > 0:
        signals.append((search_key, features[search_key]))
    if features.get("average_dwell_time", 0) > 30:
        signals.append(("high_dwell_time", features["average_dwell_time"]))
    if features.get("search_count", 0) > 0:
        signals.append(("search_count", features["search_count"]))
    if features.get("transaction_count", 0) > 0:
        signals.append(("transaction_count", features["transaction_count"]))

    signals.sort(key=lambda x: x[1], reverse=True)
    return [s[0] for s in signals[:5]]


def predict_from_events(artifact: ModelArtifact, user_id: str, events: list[dict[str, Any]]) -> dict:
    features = build_feature_vector(events)
    return predict_from_features(artifact, user_id, features)


def predict_from_features(
    artifact: ModelArtifact, user_id: str, features: dict[str, float]
) -> dict:
    cols = artifact.feature_columns
    row = {c: float(features.get(c, 0.0)) for c in cols}
    X = pd.DataFrame([row])[cols]

    proba = artifact.model.predict_proba(X)[0]
    idx = int(proba.argmax())
    intent = artifact.model.classes_[idx]
    confidence = float(proba[idx])

    return {
        "user_id": user_id,
        "intent": intent,
        "confidence": round(confidence, 4),
        "threshold": artifact.threshold,
        "reward_triggered": confidence >= artifact.threshold,
        "top_signals": top_signals(features, intent),
    }
