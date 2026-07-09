from __future__ import annotations

from typing import Any

import numpy as np


def predict_from_events(artifact: Any, user_id: str, events: list[dict]) -> dict:
    model = artifact.model

    # Build a lightweight feature vector from the event payloads.
    # This keeps the service runnable even before a richer feature extractor is wired up.
    feature_vector = np.zeros(47, dtype=np.float32)

    for event in events:
        event_type = event.get("event_type", "")
        if event_type in {"screen_viewed", "content_viewed", "search_performed"}:
            feature_vector[0] += 0.05
        if event_type in {"campaign_impression", "campaign_clicked", "conversion_completed"}:
            feature_vector[1] += 0.05
        if event_type in {"interaction_received", "scroll_activity"}:
            feature_vector[2] += 0.03
        if event_type in {"notification_opened", "reward_claimed"}:
            feature_vector[3] += 0.02

    feature_vector = feature_vector.reshape(1, -1)
    probabilities = model.predict(feature_vector, verbose=0)[0]
    class_idx = int(np.argmax(probabilities))
    confidence = float(probabilities[class_idx])

    return {
        "user_id": user_id,
        "intent": artifact.intents[class_idx],
        "confidence": confidence,
        "threshold": artifact.threshold,
        "reward_triggered": confidence >= artifact.threshold,
        "top_signals": [event.get("event_type", "") for event in events[:3] if event.get("event_type")],
    }
