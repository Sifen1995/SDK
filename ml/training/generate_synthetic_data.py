# skykin-ml/training/generate_synthetic_data.py

import os
import numpy as np
import pandas as pd
import json
from datetime import datetime, timedelta
import random
from feature_engineering import extract_features, INTENT_CLASSES

random.seed(42)
np.random.seed(42)

NUM_SAMPLES = 50_000

# Define behavioral archetypes for each intent
# Each archetype describes what a real person with
# that intent typically does on their phone

ARCHETYPES = {
    "fashion_interest": {
        "dominant_apps": ["fashion", "shopping", "social"],
        "time_dist": {"fashion": 0.35, "shopping": 0.25,
                      "social": 0.20, "other": 0.20},
        "ui_keywords": {"fashion": 0.40, "shopping": 0.30,
                        "other": 0.30},
        "peak_hours": [11, 12, 13, 19, 20, 21],
        "session_length_mean": 22,
    },
    "crypto_interest": {
        "dominant_apps": ["crypto", "news", "fintech"],
        "time_dist": {"crypto": 0.45, "news": 0.25,
                      "fintech": 0.15, "other": 0.15},
        "ui_keywords": {"crypto": 0.55, "fintech": 0.25,
                        "other": 0.20},
        "peak_hours": [7, 8, 9, 22, 23],
        "session_length_mean": 18,
    },
    "coffee_interest": {
        "dominant_apps": ["coffee", "food", "social"],
        "time_dist": {"coffee": 0.30, "food": 0.25,
                      "social": 0.25, "other": 0.20},
        "ui_keywords": {"coffee": 0.45, "food": 0.30,
                        "other": 0.25},
        "peak_hours": [7, 8, 9, 10, 15, 16],
        "session_length_mean": 12,
    },
    "fintech_interest": {
        "dominant_apps": ["fintech", "banking", "news"],
        "time_dist": {"fintech": 0.40, "banking": 0.30,
                      "news": 0.15, "other": 0.15},
        "ui_keywords": {"fintech": 0.50, "shopping": 0.20,
                        "other": 0.30},
        "peak_hours": [9, 10, 11, 14, 15],
        "session_length_mean": 15,
    },
    "travel_intent": {
        "dominant_apps": ["travel", "news", "social"],
        "time_dist": {"travel": 0.45, "social": 0.20,
                      "news": 0.15, "other": 0.20},
        "ui_keywords": {"travel": 0.55, "other": 0.45},
        "peak_hours": [12, 13, 20, 21, 22],
        "session_length_mean": 25,
    },
    "fitness_interest": {
        "dominant_apps": ["fitness", "food", "social"],
        "time_dist": {"fitness": 0.40, "food": 0.20,
                      "social": 0.20, "other": 0.20},
        "ui_keywords": {"fitness": 0.50, "food": 0.25,
                        "other": 0.25},
        "peak_hours": [5, 6, 7, 17, 18, 19],
        "session_length_mean": 20,
    },
    "shopping_interest": {
        "dominant_apps": ["shopping", "social", "fashion"],
        "time_dist": {"shopping": 0.45, "social": 0.25,
                      "fashion": 0.15, "other": 0.15},
        "ui_keywords": {"shopping": 0.50, "fashion": 0.25,
                        "other": 0.25},
        "peak_hours": [12, 13, 19, 20, 21],
        "session_length_mean": 28,
    },
    "food_interest": {
        "dominant_apps": ["food", "social", "shopping"],
        "time_dist": {"food": 0.40, "social": 0.30,
                      "shopping": 0.10, "other": 0.20},
        "ui_keywords": {"food": 0.55, "other": 0.45},
        "peak_hours": [11, 12, 17, 18, 19],
        "session_length_mean": 14,
    },
    "abandoned_cart": {
        "dominant_apps": ["shopping", "fashion", "fintech"],
        "time_dist": {"shopping": 0.50, "fashion": 0.20,
                      "fintech": 0.10, "other": 0.20},
        "ui_keywords": {"shopping": 0.55, "fashion": 0.25,
                        "other": 0.20},
        "peak_hours": [12, 13, 14, 19, 20, 21],
        "session_length_mean": 24,
        # abandoned_cart specifically has shopping signals
        # but no checkout completion signal
    },
    "no_clear_intent": {
        "dominant_apps": ["social", "news", "other"],
        "time_dist": {"social": 0.35, "news": 0.25,
                      "other": 0.40},
        "ui_keywords": {"other": 1.0},
        "peak_hours": list(range(9, 23)),
        "session_length_mean": 15,
    },
}

NOISE_LEVEL = 0.20  # 20% of sessions get noise injection


def apply_noise(value: float, scale: float, minimum: float = 0.0) -> float:
    """Add Gaussian noise while keeping values non-negative."""
    return max(value + np.random.normal(0, scale), minimum)


def generate_session(intent: str, noise: float = NOISE_LEVEL) -> dict:
    arch = ARCHETYPES[intent]

    # Session timing
    peak_hour = random.choice(arch["peak_hours"])
    hour = int(np.clip(np.random.normal(peak_hour, 1.5), 0, 23))
    day  = random.randint(0, 6)
    session_start = datetime.now().replace(
        hour=hour, minute=random.randint(0, 59))
    session_start += timedelta(minutes=int(np.random.normal(0, 30 * noise)))

    session_minutes = max(
        apply_noise(arch["session_length_mean"], 5 * noise, minimum=2), 2)
    total_switches = max(int(apply_noise(session_minutes * 0.7, 2 * noise, minimum=1)), 1)

    # Build app usage from archetype distribution
    app_usage = {}
    time_dist = arch["time_dist"]

    for cat, ratio in time_dist.items():
        # Add noise — real users do not perfectly follow patterns
        noisy_ratio = apply_noise(ratio, noise * ratio, minimum=0.0)
        minutes = noisy_ratio * session_minutes
        switches = max(int(apply_noise(noisy_ratio * total_switches, noise * total_switches, minimum=0)), 0)
        if minutes > 0.5:
            app_usage[cat] = {
                "minutes":  round(minutes, 2),
                "switches": switches,
            }

    # Build UI signals from archetype
    total_ui = random.randint(10, 80)
    ui_signals = {}
    for cat, ratio in arch["ui_keywords"].items():
        noisy_ratio = apply_noise(ratio, noise * ratio, minimum=0.0)
        count = max(int(apply_noise(noisy_ratio * total_ui, noise * total_ui, minimum=0)), 0)
        if count > 0:
            ui_signals[cat] = count

    # Inject cross-intent noise (20% of sessions)
    if random.random() < noise:
        noise_cat = random.choice(
            list(ARCHETYPES.keys()))
        noise_arch = ARCHETYPES[noise_cat]
        for cat, ratio in noise_arch["time_dist"].items():
            if cat not in app_usage:
                minutes = ratio * session_minutes * 0.2
                if minutes > 0.5:
                    app_usage[cat] = {
                        "minutes":  round(minutes, 2),
                        "switches": max(int(minutes * 0.7), 1),
                    }

    return {
        "app_usage":               app_usage,
        "ui_signals":              ui_signals,
        "session_start":           session_start,
        "session_duration_minutes": session_minutes,
        "total_switches":          total_switches,
        "is_first_session":        1.0 if random.random() < 0.3
                                   else 0.0,
    }


def generate_dataset(n: int) -> pd.DataFrame:
    records = []

    for _ in range(n):
        intent = random.choice(INTENT_CLASSES)
        session = generate_session(intent)

        # Historical data — only for ~60% of samples
        # (simulates mix of new and returning users)
        historical = None
        if random.random() < 0.6:
            days = random.randint(1, 20)
            historical = {
                "days_with_intent":   days,
                "avg_confidence":     random.uniform(0.65, 0.95),
                "last_seen_days_ago": random.randint(0, 7),
                "consistency_score":  random.uniform(0.4, 1.0),
            }

        features = extract_features(session, historical)
        label_idx = INTENT_CLASSES.index(intent)

        record = {f"f_{i}": features[i]
                  for i in range(len(features))}
        record["label"]     = intent
        record["label_idx"] = label_idx

        records.append(record)

    return pd.DataFrame(records)


if __name__ == "__main__":
    print(f"Generating {NUM_SAMPLES:,} training samples...")
    df = generate_dataset(NUM_SAMPLES)
    print(f"Shape: {df.shape}")
    print("\nLabel distribution:")
    print(df["label"].value_counts())
    out_dir = "../data/processed"
    os.makedirs(out_dir, exist_ok=True)
    df.to_csv(os.path.join(out_dir, "training_data.csv"), index=False)
    print(f"\nSaved → {out_dir}/training_data.csv")