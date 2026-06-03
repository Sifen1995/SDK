"""
Rule-based intent labeling from behavioral features.
Used when real labeled data is not yet available.
"""

from __future__ import annotations

INTENT_LABELS = [
    "fashion_interest",
    "crypto_interest",
    "food_interest",
    "education_interest",
    "gaming_interest",
    "fintech_interest",
    "general_interest",
]


def generate_label(features: dict[str, float]) -> str:
    if features.get("crypto_views", 0) > 5 and features.get("crypto_searches", 0) >= 2:
        return "crypto_interest"
    if features.get("fashion_views", 0) > 5:
        return "fashion_interest"
    if features.get("food_views", 0) > 5 and features.get("search_count", 0) >= 1:
        return "food_interest"
    if features.get("education_views", 0) > 5:
        return "education_interest"
    if features.get("gaming_views", 0) > 5 and features.get("interaction_count", 0) >= 3:
        return "gaming_interest"
    if features.get("fintech_views", 0) > 4 and (
        features.get("transaction_count", 0) >= 1 or features.get("conversion_count", 0) >= 1
    ):
        return "fintech_interest"
    return "general_interest"


def generate_labels_batch(feature_rows: list[dict[str, float]]) -> list[str]:
    return [generate_label(row) for row in feature_rows]
