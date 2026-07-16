from __future__ import annotations

from datetime import datetime
from typing import Any

import numpy as np

from training.feature_engineering import CATEGORIES, extract_features


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
    data.setdefault("session_duration_minutes", 1.0)
    data.setdefault("total_switches", 1)
    data.setdefault("is_first_session", 0.0)
    return data


def _top_signals(session: dict) -> list[str]:
    app_usage = session.get("app_usage", {})
    ui_signals = session.get("ui_signals", {})

    ranked: list[tuple[float, str]] = []
    for cat in CATEGORIES:
        minutes = float(app_usage.get(cat, {}).get("minutes", 0) or 0)
        switches = float(app_usage.get(cat, {}).get("switches", 0) or 0)
        ui = float(ui_signals.get(cat, 0) or 0)
        score = minutes + switches * 0.5 + ui
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
    """Build the 47-feature vector from accessibility + app-usage session data, then predict."""
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
    """Direct 47-float vector path (debug / training-parity tests)."""
    if len(features) != 47:
        raise ValueError(f"features must have length 47, got {len(features)}")

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
