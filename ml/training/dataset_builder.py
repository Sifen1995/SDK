"""
Build training dataset from synthetic or exported event sessions.
"""

from __future__ import annotations

import json
import os
import random
import uuid
from datetime import datetime, timedelta

import pandas as pd

from training.feature_engineering import build_feature_vector, feature_column_order
from training.label_generation import generate_label

RANDOM_STATE = 42
OUTPUT_DIR = os.path.join(os.path.dirname(__file__), "..", "data")
OUTPUT_FILE = os.path.join(OUTPUT_DIR, "behavioral_features.csv")

DOMAINS = ["crypto", "fashion", "food", "education", "gaming", "fintech", "general"]

SESSION_TEMPLATES: dict[str, list[dict]] = {
    "crypto": [
        ("content_viewed", {"category": "crypto", "asset": "bitcoin", "dwell_time": 120}),
        ("search_performed", {"category": "crypto", "query": "eth price"}),
        ("screen_viewed", {"category": "crypto", "screen": "portfolio"}),
        ("campaign_impression", {"category": "crypto", "campaign_id": "c1"}),
        ("transaction_completed", {"category": "crypto", "amount": 500}),
    ],
    "fashion": [
        ("content_viewed", {"category": "fashion", "brand": "nike", "dwell_time": 90}),
        ("scroll_activity", {"category": "fashion", "scroll_depth": 80}),
        ("search_performed", {"category": "fashion", "query": "summer shoes"}),
        ("campaign_clicked", {"category": "fashion", "campaign_id": "f2"}),
    ],
    "food": [
        ("content_viewed", {"category": "food", "item": "burger", "dwell_time": 45}),
        ("search_performed", {"category": "food", "query": "delivery near me"}),
        ("conversion_completed", {"category": "food"}),
    ],
    "education": [
        ("content_viewed", {"category": "education", "content_type": "course", "dwell_time": 200}),
        ("screen_viewed", {"category": "education", "screen": "lesson"}),
        ("interaction_received", {"category": "education", "action": "quiz_start"}),
    ],
    "gaming": [
        ("content_viewed", {"category": "gaming", "game": "rpg", "dwell_time": 300}),
        ("interaction_received", {"category": "gaming", "action": "level_up"}),
        ("interaction_received", {"category": "gaming", "action": "purchase_skin"}),
        ("reward_claimed", {"category": "gaming"}),
    ],
    "fintech": [
        ("content_viewed", {"category": "fintech", "product": "savings", "dwell_time": 60}),
        ("transaction_completed", {"category": "fintech", "amount": 1000}),
        ("conversion_completed", {"category": "fintech"}),
    ],
    "general": [
        ("session_started", {"category": "general"}),
        ("screen_viewed", {"category": "general", "screen": "home"}),
        ("notification_opened", {"category": "general"}),
    ],
}


def _synthetic_session(domain: str, rng: random.Random) -> list[dict]:
    templates = SESSION_TEMPLATES.get(domain, SESSION_TEMPLATES["general"])
    n = rng.randint(4, 10)
    base = datetime.utcnow() - timedelta(hours=rng.randint(1, 72))
    session_id = str(uuid.uuid4())
    events = []
    for i in range(n):
        et, meta = rng.choice(templates)
        meta = dict(meta)
        events.append(
            {
                "event_type": et,
                "domain": domain,
                "screen_name": meta.get("screen", "main"),
                "metadata": meta,
                "session_id": session_id,
                "created_at": base + timedelta(seconds=i * rng.randint(5, 60)),
            }
        )
    return events


def build_synthetic_dataset(num_users: int = 2000) -> pd.DataFrame:
    rng = random.Random(RANDOM_STATE)
    rows = []
    for _ in range(num_users):
        domain = rng.choice(DOMAINS)
        # repeat events to satisfy labeling thresholds
        events = _synthetic_session(domain, rng)
        if domain != "general":
            events = events + _synthetic_session(domain, rng)
        features = build_feature_vector(events)
        label = generate_label(features)
        row = {**features, "label": label, "domain_hint": domain}
        rows.append(row)
    df = pd.DataFrame(rows)
    for col in feature_column_order():
        if col not in df.columns:
            df[col] = 0.0
    return df


def build_from_event_export(export_path: str) -> pd.DataFrame:
    """
    export_path: JSON lines file, each line {"user_id": "...", "events": [...]}
    """
    rows = []
    with open(export_path, encoding="utf-8") as f:
        for line in f:
            record = json.loads(line)
            features = build_feature_vector(record.get("events", []))
            label = generate_label(features)
            rows.append({**features, "label": label})
    return pd.DataFrame(rows)


def save_dataset(df: pd.DataFrame, path: str = OUTPUT_FILE) -> str:
    os.makedirs(os.path.dirname(path), exist_ok=True)
    df.to_csv(path, index=False)
    return path


if __name__ == "__main__":
    df = build_synthetic_dataset()
    out = save_dataset(df)
    print(f"Dataset saved: {out} ({len(df)} rows)")
    print(df["label"].value_counts())
