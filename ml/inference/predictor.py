from __future__ import annotations

from datetime import datetime
from typing import Any

import numpy as np

from training.feature_engineering import CATEGORIES, FEATURE_SIZE, extract_features


def _parse_session_start(value: Any) -> datetime:
    if isinstance(value, datetime):
        return value
    if isinstance(value, str) and value.strip():
        return datetime.fromisoformat(value.replace("Z", "+00:00")).replace(tzinfo=None)
    return datetime.now()


def _normalize_session(session: dict) -> dict:
    data = dict(session)
    data["session_start"] = _parse_session_start(session.get("session_start"))
    data.setdefault("app_usage", {})
    data.setdefault("ui_signals", {})
    data.setdefault("behavioral_events", {"has_data": 0.0, "actions": {}, "categories": {}})
    data.setdefault("session_duration_minutes", 1.0)
    data.setdefault("total_switches", 1)
    data.setdefault("is_first_session", 0.0)
    return data


def _top_signals(session: dict) -> list[str]:
    app_usage = session.get("app_usage", {})
    ui_signals = session.get("ui_signals", {})
    behavioral = session.get("behavioral_events") or {}
    cats = behavioral.get("categories") or {}

    ranked: list[tuple[float, str]] = []
    for cat in CATEGORIES:
        minutes = float(app_usage.get(cat, {}).get("minutes", 0) or 0)
        switches = float(app_usage.get(cat, {}).get("switches", 0) or 0)
        ui = float(ui_signals.get(cat, 0) or 0)
        events = float(cats.get(cat, 0) or 0)
        score = minutes + switches * 0.5 + ui + events * 0.8
        if score > 0:
            ranked.append((score, cat))

    ranked.sort(reverse=True)
    return [name for _, name in ranked[:3]]


def predict_from_session(
    artifact: Any,
    user_id: str,
    session: dict,
    historical: dict | None = None,
) -> dict:
    """Build the 71-feature vector (incl. behavioral events), then predict."""
    session_data = _normalize_session(session)
    feature_vector = extract_features(session_data, historical)
    probabilities = artifact.model.predict(feature_vector.reshape(1, -1), verbose=0)[0]
    class_idx = int(np.argmax(probabilities))
    confidence = float(probabilities[class_idx])

    return {
        "user_id": user_id,
        "intent": artifact.intents[class_idx],
        "confidence": confidence,
        "threshold": artifact.threshold,
        "reward_triggered": confidence >= artifact.threshold,
        "top_signals": _top_signals(session_data),
    }


def predict_from_features(artifact: Any, user_id: str, features: list[float]) -> dict:
    """Direct feature-vector path (debug / training-parity tests)."""
    if len(features) != FEATURE_SIZE:
        raise ValueError(f"features must have length {FEATURE_SIZE}, got {len(features)}")

    feature_vector = np.asarray(features, dtype=np.float32).reshape(1, -1)
    probabilities = artifact.model.predict(feature_vector, verbose=0)[0]
    class_idx = int(np.argmax(probabilities))
    confidence = float(probabilities[class_idx])

    return {
        "user_id": user_id,
        "intent": artifact.intents[class_idx],
        "confidence": confidence,
        "threshold": artifact.threshold,
        "reward_triggered": confidence >= artifact.threshold,
        "top_signals": [],
    }
