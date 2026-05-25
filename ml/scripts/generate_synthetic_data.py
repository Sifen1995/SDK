"""
Skykin — Synthetic Session Data Generator
File:    scripts/generate_data.py
Run:     python scripts/generate_data.py
Output:  data/synthetic_sessions.csv
"""

import json
import random
import os
import numpy as np
import pandas as pd

# ─────────────────────────────────────────────────────────────
#  CONFIGURATION
# ─────────────────────────────────────────────────────────────

NUM_SESSIONS      = 10_000
NOISE_PROBABILITY = 0.20   # 20% of sessions get a distractor event injected
RANDOM_STATE      = 42
OUTPUT_DIR        = os.path.join(os.path.dirname(__file__), "..", "data")
OUTPUT_FILE       = os.path.join(OUTPUT_DIR, "synthetic_sessions.csv")

INTENTS = [
    "coffee_interest",
    "crypto_interest",
    "fashion_interest",
    "abandoned_cart",
    "signup_intent",
]

# ─────────────────────────────────────────────────────────────
#  EVENT TEMPLATES PER INTENT
#  Each intent has a pool of (event_type, metadata) tuples.
#  Sessions are built by sampling 2-5 events from the pool.
# ─────────────────────────────────────────────────────────────

EVENT_TEMPLATES = {
    "coffee_interest": [
        ("search",        {"query": "best arabica coffee"}),
        ("search",        {"query": "coffee near me"}),
        ("search",        {"query": "espresso vs americano"}),
        ("search",        {"query": "where to buy coffee beans"}),
        ("product_view",  {"product_name": "Espresso Machine",      "category": "Appliances"}),
        ("product_view",  {"product_name": "Ethiopian Yirgacheffe",  "category": "Coffee"}),
        ("product_view",  {"product_name": "French Press 1L",        "category": "Coffee"}),
        ("product_view",  {"product_name": "Latte Blend 500g",       "category": "Coffee"}),
        ("category_view", {"category": "Coffee & Beverages"}),
        ("category_view", {"category": "Cafe Equipment"}),
        ("add_to_cart",   {"product_name": "Latte Blend 500g",       "price": 120.0}),
        ("add_to_cart",   {"product_name": "Espresso Machine",       "price": 4500.0}),
    ],
    "crypto_interest": [
        ("search",        {"query": "bitcoin price today"}),
        ("search",        {"query": "how to buy ethereum"}),
        ("search",        {"query": "best crypto wallet 2026"}),
        ("search",        {"query": "crypto trading Ethiopia"}),
        ("search",        {"query": "what is defi"}),
        ("product_view",  {"product_name": "Ledger Nano Wallet",     "category": "Crypto Hardware"}),
        ("product_view",  {"product_name": "Trezor Model T",         "category": "Crypto Hardware"}),
        ("category_view", {"category": "Blockchain & Crypto"}),
        ("category_view", {"category": "DeFi Products"}),
        ("category_view", {"category": "Web3 Tools"}),
    ],
    "fashion_interest": [
        ("category_view", {"category": "Mens Shoes"}),
        ("category_view", {"category": "Summer Dresses"}),
        ("category_view", {"category": "Luxury Watches"}),
        ("category_view", {"category": "Streetwear"}),
        ("category_view", {"category": "Womens Bags"}),
        ("product_view",  {"product_name": "Air Max 2026",           "category": "Mens Shoes"}),
        ("product_view",  {"product_name": "Gold Chronograph",       "category": "Luxury Watches"}),
        ("product_view",  {"product_name": "Linen Blazer",           "category": "Clothing"}),
        ("search",        {"query": "summer fashion trends"}),
        ("search",        {"query": "luxury watches Ethiopia"}),
        ("add_to_cart",   {"product_name": "Linen Blazer",           "price": 850.0}),
    ],
    "abandoned_cart": [
        ("add_to_cart",      {"product_id": "p_101", "product_name": "Sony Headphones",    "price": 3200.0}),
        ("add_to_cart",      {"product_id": "p_202", "product_name": "Nike Running Shoes", "price": 1500.0}),
        ("add_to_cart",      {"product_id": "p_303", "product_name": "Coffee Grinder",     "price": 780.0}),
        ("add_to_cart",      {"product_id": "p_404", "product_name": "Leather Wallet",     "price": 450.0}),
        ("remove_from_cart", {"product_id": "p_101"}),
        ("remove_from_cart", {"product_id": "p_202"}),
        ("product_view",     {"product_name": "Sony Headphones",     "category": "Electronics"}),
        ("checkout_started", {"cart_total": 3200.0, "item_count": 1}),
        # intentionally no checkout_complete — that's the abandoned signal
    ],
    "signup_intent": [
        ("signup_complete",  {"method": "email"}),
        ("signup_complete",  {"method": "google"}),
        ("signup_complete",  {"method": "phone"}),
        ("product_view",     {"product_name": "Welcome Bundle",      "category": "Offers"}),
        ("category_view",    {"category": "New User Deals"}),
        ("search",           {"query": "how to get started"}),
        ("search",           {"query": "new user offers"}),
    ],
}

# Random distractor events injected as noise
DISTRACTOR_EVENTS = [
    ("search",        {"query": "deals today"}),
    ("search",        {"query": "trending products"}),
    ("category_view", {"category": "Electronics"}),
    ("category_view", {"category": "Home & Garden"}),
    ("product_view",  {"product_name": "Random Item",  "category": "General"}),
]


# ─────────────────────────────────────────────────────────────
#  HELPERS
# ─────────────────────────────────────────────────────────────

def build_session_text(events: list[tuple]) -> str:
    """
    Flattens a list of (event_type, metadata_dict) tuples into
    a single whitespace-separated string for TF-IDF input.

    Example:
      [("search", {"query": "coffee near me"}),
       ("category_view", {"category": "Coffee & Beverages"})]
      →  "search coffee near me category_view coffee beverages"
    """
    tokens = []
    for event_type, metadata in events:
        tokens.append(event_type)
        for value in metadata.values():
            cleaned = str(value).lower().replace("_", " ").replace("&", "")
            tokens.append(cleaned)
    return " ".join(tokens)


def generate_sessions(n: int) -> pd.DataFrame:
    """
    Generates n synthetic user sessions.

    Each row represents one user session with:
      - session_length  : how many events occurred
      - features        : flattened text of all events (used by TF-IDF)
      - label           : ground-truth intent
    """
    records = []

    for _ in range(n):
        intent = random.choice(INTENTS)
        pool   = EVENT_TEMPLATES[intent]

        # Sample 2–5 events from the intent pool
        session_length = random.randint(2, 5)
        session_events = random.choices(pool, k=session_length)

        # Inject noise: 20% chance of one random distractor event
        if random.random() < NOISE_PROBABILITY:
            insert_pos = random.randint(0, len(session_events))
            session_events.insert(insert_pos, random.choice(DISTRACTOR_EVENTS))

        records.append({
            "session_length": len(session_events),
            "features":       build_session_text(session_events),
            "label":          intent,
        })

    return pd.DataFrame(records)


# ─────────────────────────────────────────────────────────────
#  MAIN
# ─────────────────────────────────────────────────────────────

if __name__ == "__main__":
    random.seed(RANDOM_STATE)
    np.random.seed(RANDOM_STATE)

    print("=" * 55)
    print("  Skykin — Synthetic Data Generator")
    print("=" * 55)
    print(f"\nGenerating {NUM_SESSIONS:,} sessions...")

    df = generate_sessions(NUM_SESSIONS)

    print(f"Done. Shape: {df.shape}")
    print("\nLabel distribution:")
    print(df["label"].value_counts().to_string())

    # Save
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    df.to_csv(OUTPUT_FILE, index=False)
    print(f"\nSaved → {OUTPUT_FILE}")
    print("\nRun model.py next to train the classifier.")