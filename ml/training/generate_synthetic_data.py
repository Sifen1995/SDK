# skykin-ml/training/generate_synthetic_data.py

import numpy as np
import pandas as pd
import json
from datetime import datetime, timedelta
import random
from feature_engineering import extract_features, INTENT_CLASSES

random.seed(42)
np.random.seed(42)

NUM_SAMPLES = 50_000

# Expanded behavioral archetypes to include:
# 1. Standard app usage / UI keyword patterns
# 2. In-App Behavioral Events (Funnel stages & Categories)
ARCHETYPES = {
    "fashion_interest": {
        "dominant_apps": ["fashion", "shopping", "social"],
        "time_dist": {"fashion": 0.35, "shopping": 0.25, "social": 0.20, "other": 0.20},
        "ui_keywords": {"fashion": 0.40, "shopping": 0.30, "other": 0.30},
        "peak_hours": [11, 12, 13, 19, 20, 21],
        "session_length_mean": 22,
        # Behavioral profile
        "has_behavioral": True,
        "primary_category": "fashion",
        "action_weights": {"browse": 12, "view": 8, "stage": 3, "checkout": 1, "abandon": 0},
    },
    "crypto_interest": {
        "dominant_apps": ["crypto", "news", "fintech"],
        "time_dist": {"crypto": 0.45, "news": 0.25, "fintech": 0.15, "other": 0.15},
        "ui_keywords": {"crypto": 0.55, "fintech": 0.25, "other": 0.20},
        "peak_hours": [7, 8, 9, 22, 23],
        "session_length_mean": 18,
        "has_behavioral": True,
        "primary_category": "crypto",
        "action_weights": {"browse": 10, "view": 15, "stage": 2, "checkout": 1, "abandon": 0},
    },
    "coffee_interest": {
        "dominant_apps": ["coffee", "food", "social"],
        "time_dist": {"coffee": 0.30, "food": 0.25, "social": 0.25, "other": 0.20},
        "ui_keywords": {"coffee": 0.45, "food": 0.30, "other": 0.25},
        "peak_hours": [7, 8, 9, 10, 15, 16],
        "session_length_mean": 12,
        "has_behavioral": True,
        "primary_category": "coffee",
        "action_weights": {"browse": 5, "view": 6, "stage": 2, "checkout": 1, "abandon": 0},
    },
    "fintech_interest": {
        "dominant_apps": ["fintech", "banking", "news"],
        "time_dist": {"fintech": 0.40, "banking": 0.30, "news": 0.15, "other": 0.15},
        "ui_keywords": {"fintech": 0.50, "shopping": 0.20, "other": 0.30},
        "peak_hours": [9, 10, 11, 14, 15],
        "session_length_mean": 15,
        "has_behavioral": True,
        "primary_category": "fintech",
        "action_weights": {"browse": 8, "view": 10, "stage": 4, "checkout": 2, "abandon": 0},
    },
    "travel_intent": {
        "dominant_apps": ["travel", "news", "social"],
        "time_dist": {"travel": 0.45, "social": 0.20, "news": 0.15, "other": 0.20},
        "ui_keywords": {"travel": 0.55, "other": 0.45},
        "peak_hours": [12, 13, 20, 21, 22],
        "session_length_mean": 25,
        "has_behavioral": True,
        "primary_category": "travel",
        "action_weights": {"browse": 15, "view": 12, "stage": 5, "checkout": 2, "abandon": 0},
    },
    "fitness_interest": {
        "dominant_apps": ["fitness", "food", "social"],
        "time_dist": {"fitness": 0.40, "food": 0.20, "social": 0.20, "other": 0.20},
        "ui_keywords": {"fitness": 0.50, "food": 0.25, "other": 0.25},
        "peak_hours": [5, 6, 7, 17, 18, 19],
        "session_length_mean": 20,
        "has_behavioral": True,
        "primary_category": "fitness",
        "action_weights": {"browse": 7, "view": 8, "stage": 2, "checkout": 1, "abandon": 0},
    },
    "shopping_interest": {
        "dominant_apps": ["shopping", "social", "fashion"],
        "time_dist": {"shopping": 0.45, "social": 0.25, "fashion": 0.15, "other": 0.15},
        "ui_keywords": {"shopping": 0.50, "fashion": 0.25, "other": 0.25},
        "peak_hours": [12, 13, 19, 20, 21],
        "session_length_mean": 28,
        "has_behavioral": True,
        "primary_category": "shopping",
        "action_weights": {"browse": 18, "view": 14, "stage": 6, "checkout": 3, "abandon": 0},
    },
    "food_interest": {
        "dominant_apps": ["food", "social", "shopping"],
        "time_dist": {"food": 0.40, "social": 0.30, "shopping": 0.10, "other": 0.20},
        "ui_keywords": {"food": 0.55, "other": 0.45},
        "peak_hours": [11, 12, 17, 18, 19],
        "session_length_mean": 14,
        "has_behavioral": True,
        "primary_category": "food",
        "action_weights": {"browse": 8, "view": 10, "stage": 3, "checkout": 2, "abandon": 0},
    },
    "no_clear_intent": {
        "dominant_apps": ["social", "news", "other"],
        "time_dist": {"social": 0.35, "news": 0.25, "other": 0.40},
        "ui_keywords": {"other": 1.0},
        "peak_hours": list(range(9, 23)),
        "session_length_mean": 15,
        "has_behavioral": False,  # Apps with no tracking integration fallback to False
        "primary_category": None,
        "action_weights": {"browse": 0, "view": 0, "stage": 0, "checkout": 0, "abandon": 0},
    },
}

NOISE_LEVEL = 0.20  # 20% noise injection

def generate_session(intent: str, noise: float = NOISE_LEVEL) -> dict:
    arch = ARCHETYPES[intent]

    # 1. Legacy Session Timing & Switches
    peak_hour = random.choice(arch["peak_hours"])
    hour = int(np.clip(np.random.normal(peak_hour, 1.5), 0, 23))
    session_start = datetime.now().replace(hour=hour, minute=random.randint(0, 59))
    session_minutes = max(np.random.normal(arch["session_length_mean"], 5), 2)
    total_switches  = max(int(session_minutes * 0.7), 1)

    # 2. Legacy App Usage
    app_usage = {}
    for cat, ratio in arch["time_dist"].items():
        noisy_ratio = max(ratio + np.random.normal(0, noise * ratio), 0)
        minutes  = noisy_ratio * session_minutes
        switches = max(int(noisy_ratio * total_switches), 0)
        if minutes > 0.5:
            app_usage[cat] = {"minutes": round(minutes, 2), "switches": switches}

    # 3. Legacy UI Signals
    total_ui = random.randint(10, 80)
    ui_signals = {}
    for cat, ratio in arch["ui_keywords"].items():
        noisy_ratio = max(ratio + np.random.normal(0, noise * ratio), 0)
        count = int(noisy_ratio * total_ui)
        if count > 0:
            ui_signals[cat] = count

    # 4. In-App Behavioral Events (Features 47–70)
    # Simulate a ~25% chance that even intent apps don't emit behavioral logs (fallback test)
    has_behavioral_data = arch["has_behavioral"] and (random.random() > 0.25)

    behavioral_events = {
        "has_data": 1.0 if has_behavioral_data else 0.0,
        "actions": {
            "browseCategory": 0,
            "viewItem": 0,
            "stageTransaction": 0,
            "initiateCheckout": 0,
            "abandonTransaction": 0,
        },
        "categories": {
            "coffee": 0,
            "fashion": 0,
            "crypto": 0,
            "fintech": 0,
            "travel": 0,
            "fitness": 0,
            "shopping": 0,
            "food": 0,
        }
    }

    if has_behavioral_data and arch["primary_category"]:
        weights = arch["action_weights"]
        p_cat = arch["primary_category"]

        # Add primary category counts with noise
        for act, base in weights.items():
            if base > 0:
                count = max(0, int(np.random.normal(base, base * noise)))
                if act == "browse": behavioral_events["actions"]["browseCategory"] += count
                elif act == "view": behavioral_events["actions"]["viewItem"] += count
                elif act == "stage": behavioral_events["actions"]["stageTransaction"] += count
                elif act == "checkout": behavioral_events["actions"]["initiateCheckout"] += count
                elif act == "abandon": behavioral_events["actions"]["abandonTransaction"] += count

                # Increment category tally
                if p_cat in behavioral_events["categories"]:
                    behavioral_events["categories"][p_cat] += count

    return {
        "app_usage":               app_usage,
        "ui_signals":              ui_signals,
        "session_start":           session_start,
        "session_duration_minutes": session_minutes,
        "total_switches":          total_switches,
        "is_first_session":        1.0 if random.random() < 0.3 else 0.0,
        "behavioral_events":       behavioral_events,  # Added payload
    }


def generate_dataset(n: int) -> pd.DataFrame:
    records = []

    for _ in range(n):
        intent = random.choice(INTENT_CLASSES)
        session = generate_session(intent)

        historical = None
        if random.random() < 0.6:
            days = random.randint(1, 20)
            historical = {
                "days_with_intent":   days,
                "avg_confidence":     random.uniform(0.65, 0.95),
                "last_seen_days_ago": random.randint(0, 7),
                "consistency_score":  random.uniform(0.4, 1.0),
            }

        # Extracts full 71-element float array
        features = extract_features(session, historical)
        label_idx = INTENT_CLASSES.index(intent)

        record = {f"f_{i}": features[i] for i in range(len(features))}
        record["label"]     = intent
        record["label_idx"] = label_idx

        records.append(record)

    return pd.DataFrame(records)


if __name__ == "__main__":
    print(f"Generating {NUM_SAMPLES:,} training samples with 71 feature dimensions...")
    df = generate_dataset(NUM_SAMPLES)
    print(f"Dataset Shape: {df.shape}")
    print("\nLabel distribution:")
    print(df["label"].value_counts())
    df.to_csv("../data/processed/training_data.csv", index=False)
    print("\nSaved → data/processed/training_data.csv")