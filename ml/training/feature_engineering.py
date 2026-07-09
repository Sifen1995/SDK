# skykin-ml/training/feature_engineering.py

"""
Feature vector: 47 features total

GROUP 1 — App usage time (12 features)
  One feature per app category, normalized to 0-1
  Value = minutes_in_category / total_session_minutes

  [0]  fashion_time_ratio
  [1]  shopping_time_ratio
  [2]  crypto_time_ratio
  [3]  fintech_time_ratio
  [4]  coffee_time_ratio
  [5]  food_time_ratio
  [6]  news_time_ratio
  [7]  social_time_ratio
  [8]  travel_time_ratio
  [9]  fitness_time_ratio
  [10] banking_time_ratio
  [11] other_time_ratio

GROUP 2 — App switch frequency (12 features)
  How many times user switched to each category
  Normalized by total switches

  [12] fashion_switch_count_ratio
  [13] shopping_switch_count_ratio
  ... (same categories as Group 1)
  [23] other_switch_count_ratio

GROUP 3 — UI text signals (8 features)
  How many UI elements matched each keyword category
  Normalized by total UI events in session

  [24] fashion_ui_signal_ratio
  [25] crypto_ui_signal_ratio
  [26] coffee_ui_signal_ratio
  [27] fintech_ui_signal_ratio
  [28] travel_ui_signal_ratio
  [29] fitness_ui_signal_ratio
  [30] shopping_ui_signal_ratio
  [31] food_ui_signal_ratio

GROUP 4 — Temporal features (6 features)
  When the session happened — affects intent probabilities

  [32] hour_sin    = sin(2π × hour/24)   — cyclical encoding
  [33] hour_cos    = cos(2π × hour/24)
  [34] day_sin     = sin(2π × day/7)
  [35] day_cos     = cos(2π × day/7)
  [36] is_weekend  = 0 or 1
  [37] is_morning  = 0 or 1 (6am-11am)

GROUP 5 — Session features (5 features)
  Context about the current session

  [38] session_duration_normalized  (minutes/120, capped at 1.0)
  [39] total_app_switches_normalized (switches/50, capped at 1.0)
  [40] dominant_category_ratio      (time in top category / total)
  [41] category_diversity_score     (how spread across categories)
  [42] is_first_session_of_day      = 0 or 1

GROUP 6 — Historical signals (4 features)
  From stored pseudonymous profile (consented users only)
  For non-consented users these are all 0.0

  [43] days_with_this_intent_last_30  (normalized /30)
  [44] avg_confidence_last_30_days
  [45] intent_recency_score           (1.0 if yesterday, decays)
  [46] intent_consistency_score       (how stable across sessions)
"""

import numpy as np
from datetime import datetime

FEATURE_SIZE = 47

CATEGORIES = [
    "fashion", "shopping", "crypto", "fintech", "coffee",
    "food", "news", "social", "travel", "fitness",
    "banking", "other"
]

INTENT_CLASSES = [
    "fashion_interest",
    "crypto_interest",
    "coffee_interest",
    "fintech_interest",
    "travel_intent",
    "fitness_interest",
    "shopping_interest",
    "food_interest",
    "abandoned_cart",
    "no_clear_intent",
]

def extract_features(session_data: dict,
                     historical_data: dict = None) -> np.ndarray:
    """
    Convert raw session data into a 47-feature vector.

    session_data = {
        "app_usage": {
            "fashion": {"minutes": 12.5, "switches": 4},
            "crypto":  {"minutes": 3.2,  "switches": 2},
            ...
        },
        "ui_signals": {
            "fashion": 8,
            "crypto":  2,
            ...
        },
        "session_start": datetime,
        "session_duration_minutes": 18.3,
        "total_switches": 12,
    }

    historical_data = {
        "days_with_intent": 7,
        "avg_confidence": 0.82,
        "last_seen_days_ago": 1,
        "consistency_score": 0.75,
    }
    """
    features = np.zeros(FEATURE_SIZE, dtype=np.float32)

    app_usage = session_data.get("app_usage", {})
    ui_signals = session_data.get("ui_signals", {})
    session_start = session_data.get("session_start", datetime.now())
    session_minutes = session_data.get("session_duration_minutes", 1.0)
    total_switches = max(session_data.get("total_switches", 1), 1)

    total_minutes = sum(
        v.get("minutes", 0) for v in app_usage.values()
    ) or 1.0
    total_ui = sum(ui_signals.values()) or 1

    # Group 1 — App usage time ratios
    for i, cat in enumerate(CATEGORIES):
        usage = app_usage.get(cat, {})
        features[i] = usage.get("minutes", 0) / total_minutes

    # Group 2 — App switch frequency ratios
    for i, cat in enumerate(CATEGORIES):
        usage = app_usage.get(cat, {})
        features[12 + i] = usage.get("switches", 0) / total_switches

    # Group 3 — UI text signal ratios
    ui_cats = [
        "fashion", "crypto", "coffee", "fintech",
        "travel", "fitness", "shopping", "food"
    ]
    for i, cat in enumerate(ui_cats):
        features[24 + i] = ui_signals.get(cat, 0) / total_ui

    # Group 4 — Temporal features (cyclical encoding)
    hour = session_start.hour
    day  = session_start.weekday()
    features[32] = np.sin(2 * np.pi * hour / 24)
    features[33] = np.cos(2 * np.pi * hour / 24)
    features[34] = np.sin(2 * np.pi * day  / 7)
    features[35] = np.cos(2 * np.pi * day  / 7)
    features[36] = 1.0 if day >= 5 else 0.0
    features[37] = 1.0 if 6 <= hour <= 11 else 0.0

    # Group 5 — Session features
    features[38] = min(session_minutes / 120.0, 1.0)
    features[39] = min(total_switches / 50.0, 1.0)

    cat_times = [
        app_usage.get(c, {}).get("minutes", 0) for c in CATEGORIES
    ]
    features[40] = max(cat_times) / total_minutes if cat_times else 0.0

    # Category diversity — higher means more spread across categories
    cat_ratios = np.array(cat_times) / total_minutes
    cat_ratios = cat_ratios[cat_ratios > 0]
    features[41] = float(-np.sum(
        cat_ratios * np.log(cat_ratios + 1e-9)
    ) / np.log(len(CATEGORIES)))

    features[42] = session_data.get("is_first_session", 0.0)

    # Group 6 — Historical signals
    if historical_data:
        features[43] = min(
            historical_data.get("days_with_intent", 0) / 30.0, 1.0)
        features[44] = historical_data.get("avg_confidence", 0.0)
        days_ago = historical_data.get("last_seen_days_ago", 30)
        features[45] = np.exp(-days_ago / 7.0)  # exponential decay
        features[46] = historical_data.get("consistency_score", 0.0)
    # If no historical data, features[43:47] stay 0.0

    return features