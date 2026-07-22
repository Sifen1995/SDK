# skykin-ml/training/feature_engineering.py

"""
Feature vector: 71 features total

GROUP 1 — App usage time (12 features: 0–11)
GROUP 2 — App switch frequency (12 features: 12–23)
GROUP 3 — UI text signals (8 features: 24–31)
GROUP 4 — Temporal features (6 features: 32–37)
GROUP 5 — Session features (5 features: 38–42)
GROUP 6 — Historical signals (4 features: 43–46)

GROUP 7 — In-App Funnel & Behavioral Events (24 features: 47–70)
  Generic event tracking abstraction that seamlessly translates between
  E-commerce app flows (Add to Cart, Checkout) and Telebirr Super App flows 
  (Biller Amount Set, PIN Confirmation Prompt).

  Action Counts Normalized by Total Behavioral Actions (5 features)
  [47] browse_category_action_ratio
  [48] view_item_action_ratio
  [49] stage_transaction_ratio    (e.g., Add to Cart OR Telebirr Amount Set)
  [50] initiate_checkout_ratio    (e.g., Proceed to Checkout OR Telebirr PIN Modal)
  [51] abandon_transaction_ratio  (e.g., Cart Abandoned OR Telebirr PIN Cancelled)

  Funnel Velocity & Intent Ratios (4 features)
  [52] stage_to_browse_ratio      (stage_transaction / browse_category)
  [53] checkout_to_stage_ratio    (initiate_checkout / stage_transaction)
  [54] abandon_to_stage_ratio     (abandon_transaction / stage_transaction)
  [55] high_intent_funnel_depth   ((stage + 2*checkout + 3*abandon) / total_actions)

  Behavioral Category Action Counts Normalized by Total Behavioral Actions (8 features)
  [56] coffee_event_ratio
  [57] fashion_event_ratio
  [58] crypto_event_ratio
  [59] fintech_event_ratio
  [60] travel_event_ratio
  [61] fitness_event_ratio
  [62] shopping_event_ratio
  [63] food_event_ratio

  Cross-Dimensional Interaction Signals (6 features)
  [64] shopping_stage_intent_signal  (shopping_app_time * stage_transaction_ratio)
  [65] fintech_stage_intent_signal   (fintech_app_time * stage_transaction_ratio)
  [66] travel_stage_intent_signal    (travel_app_time * stage_transaction_ratio)
  [67] shopping_abandonment_signal   (shopping_ui_signal * abandon_transaction_ratio)
  [68] fintech_abandonment_signal    (fintech_ui_signal * abandon_transaction_ratio)
  [69] high_intent_conversion_rate   ((stage + checkout) / max(total_actions, 1))

  Metadata Indicator (1 feature)
  [70] has_behavioral_data_flag      (1.0 if host app emits in-app events, else 0.0)
"""

import numpy as np
from datetime import datetime

FEATURE_SIZE = 71

CATEGORIES = [
    "fashion", "shopping", "crypto", "fintech", "coffee",
    "food", "news", "social", "travel", "fitness",
    "banking", "other"
]

BEHAVIORAL_CATEGORIES = [
    "coffee", "fashion", "crypto", "fintech",
    "travel", "fitness", "shopping", "food"
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
    "no_clear_intent",
]

def extract_features(session_data: dict,
                     historical_data: dict = None) -> np.ndarray:
    """
    Convert raw session data into a 71-feature vector.
    """
    features = np.zeros(FEATURE_SIZE, dtype=np.float32)

    app_usage = session_data.get("app_usage", {})
    ui_signals = session_data.get("ui_signals", {})
    session_start = session_data.get("session_start", datetime.now())
    session_minutes = session_data.get("session_duration_minutes", 1.0)
    total_switches = max(session_data.get("total_switches", 1), 1)

    total_minutes = sum(v.get("minutes", 0) for v in app_usage.values()) or 1.0
    total_ui = sum(ui_signals.values()) or 1

    # =========================================================================
    # LEGACY FEATURES (0–46) — UNCHANGED
    # =========================================================================

    # Group 1 — App usage time ratios (0–11)
    for i, cat in enumerate(CATEGORIES):
        usage = app_usage.get(cat, {})
        features[i] = usage.get("minutes", 0) / total_minutes

    # Group 2 — App switch frequency ratios (12–23)
    for i, cat in enumerate(CATEGORIES):
        usage = app_usage.get(cat, {})
        features[12 + i] = usage.get("switches", 0) / total_switches

    # Group 3 — UI text signal ratios (24–31)
    ui_cats = [
        "fashion", "crypto", "coffee", "fintech",
        "travel", "fitness", "shopping", "food"
    ]
    for i, cat in enumerate(ui_cats):
        features[24 + i] = ui_signals.get(cat, 0) / total_ui

    # Group 4 — Temporal features (32–37)
    hour = session_start.hour
    day  = session_start.weekday()
    features[32] = np.sin(2 * np.pi * hour / 24)
    features[33] = np.cos(2 * np.pi * hour / 24)
    features[34] = np.sin(2 * np.pi * day  / 7)
    features[35] = np.cos(2 * np.pi * day  / 7)
    features[36] = 1.0 if day >= 5 else 0.0
    features[37] = 1.0 if 6 <= hour <= 11 else 0.0

    # Group 5 — Session features (38–42)
    features[38] = min(session_minutes / 120.0, 1.0)
    features[39] = min(total_switches / 50.0, 1.0)

    cat_times = [app_usage.get(c, {}).get("minutes", 0) for c in CATEGORIES]
    features[40] = max(cat_times) / total_minutes if cat_times else 0.0

    # Category diversity
    cat_ratios = np.array(cat_times) / total_minutes
    cat_ratios = cat_ratios[cat_ratios > 0]
    features[41] = float(-np.sum(
        cat_ratios * np.log(cat_ratios + 1e-9)
    ) / np.log(len(CATEGORIES)))

    features[42] = session_data.get("is_first_session", 0.0)

    # Group 6 — Historical signals (43–46)
    if historical_data:
        features[43] = min(historical_data.get("days_with_intent", 0) / 30.0, 1.0)
        features[44] = historical_data.get("avg_confidence", 0.0)
        days_ago = historical_data.get("last_seen_days_ago", 30)
        features[45] = np.exp(-days_ago / 7.0)
        features[46] = historical_data.get("consistency_score", 0.0)

    # =========================================================================
    # NEW GROUP 7 — IN-APP FUNNEL & BEHAVIORAL EVENTS (47–70)
    # =========================================================================
    behavioral = session_data.get("behavioral_events", {})
    has_data = behavioral.get("has_data", 0.0)
    features[70] = float(has_data)

    if has_data > 0.0:
        actions = behavioral.get("actions", {})
        browse_cnt   = actions.get("browseCategory", 0)
        view_cnt     = actions.get("viewItem", 0)
        stage_cnt    = actions.get("stageTransaction", 0)
        checkout_cnt = actions.get("initiateCheckout", 0)
        abandon_cnt  = actions.get("abandonTransaction", 0)

        total_actions = sum([browse_cnt, view_cnt, stage_cnt, checkout_cnt, abandon_cnt])
        denom_actions = max(total_actions, 1)

        # Action Ratios (47–51)
        features[47] = browse_cnt / denom_actions
        features[48] = view_cnt / denom_actions
        features[49] = stage_cnt / denom_actions
        features[50] = checkout_cnt / denom_actions
        features[51] = abandon_cnt / denom_actions

        # Funnel Velocity & Ratios (52–55)
        features[52] = stage_cnt / max(browse_cnt, 1)
        features[53] = checkout_cnt / max(stage_cnt, 1)
        features[54] = abandon_cnt / max(stage_cnt, 1)
        features[55] = min((stage_cnt + (2 * checkout_cnt) + (3 * abandon_cnt)) / denom_actions, 1.0)

        # Category Event Distributions (56–63)
        b_categories = behavioral.get("categories", {})
        total_cat_events = sum(b_categories.get(cat, 0) for cat in BEHAVIORAL_CATEGORIES) or 1
        for idx, cat in enumerate(BEHAVIORAL_CATEGORIES):
            features[56 + idx] = b_categories.get(cat, 0) / total_cat_events

        # Cross-Dimensional Interactions (64–69)
        # Combines App Time (0–11) or UI Text (24–31) with Funnel Progression
        shopping_time = features[1]   # shopping_time_ratio
        fintech_time  = features[3]   # fintech_time_ratio
        travel_time   = features[8]   # travel_time_ratio
        shopping_ui   = features[30]  # shopping_ui_signal_ratio
        fintech_ui    = features[27]  # fintech_ui_signal_ratio

        features[64] = shopping_time * features[49]  # shopping_time * stage_transaction
        features[65] = fintech_time * features[49]   # fintech_time * stage_transaction
        features[66] = travel_time * features[49]    # travel_time * stage_transaction
        features[67] = shopping_ui * features[51]    # shopping_ui * abandon_transaction
        features[68] = fintech_ui * features[51]     # fintech_ui * abandon_transaction
        features[69] = (stage_cnt + checkout_cnt) / denom_actions

    return features