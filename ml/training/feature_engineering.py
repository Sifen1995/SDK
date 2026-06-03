"""
Behavioral feature engineering from raw event rows.

Each row is a dict with keys:
  event_type, domain, screen_name, metadata (dict), session_id, created_at (optional)
"""

from __future__ import annotations

from collections import Counter, defaultdict
from datetime import datetime
from typing import Any

import pandas as pd

CATEGORY_KEYS = (
    "fashion",
    "crypto",
    "food",
    "education",
    "gaming",
    "fintech",
)

CONTENT_VIEW_TYPES = {"content_viewed", "screen_viewed"}
SEARCH_TYPES = {"search_performed"}
INTERACTION_TYPES = {"interaction_received", "scroll_activity"}
CAMPAIGN_IMPRESSION_TYPES = {"campaign_impression"}
CAMPAIGN_CLICK_TYPES = {"campaign_clicked"}
CONVERSION_TYPES = {"conversion_completed"}
TRANSACTION_TYPES = {"transaction_completed"}
REWARD_TYPES = {"reward_claimed"}
SESSION_TYPES = {"session_started"}


def _meta_category(metadata: dict[str, Any] | None) -> str | None:
    if not metadata:
        return None
    cat = metadata.get("category")
    if cat is None:
        return None
    return str(cat).strip().lower()


def _num(metadata: dict[str, Any] | None, key: str, default: float = 0.0) -> float:
    if not metadata:
        return default
    val = metadata.get(key)
    if val is None:
        return default
    try:
        return float(val)
    except (TypeError, ValueError):
        return default


def _parse_ts(value: Any) -> datetime | None:
    if value is None:
        return None
    if isinstance(value, datetime):
        return value
    try:
        return pd.to_datetime(value).to_pydatetime()
    except Exception:
        return None


def build_feature_vector(events: list[dict[str, Any]]) -> dict[str, float]:
    """Compute behavioral features for one user/session event window."""
    if not events:
        return _empty_features()

    sessions: set[str] = set()
    session_durations: list[float] = []
    session_event_counts: Counter[str] = Counter()

    dwell_times: list[float] = []
    scroll_depths: list[float] = []

    content_views = 0
    searches = 0
    interactions = 0
    campaign_impressions = 0
    campaign_clicks = 0
    conversions = 0
    transactions = 0
    rewards_claimed = 0

    category_counter: Counter[str] = Counter()
    domain_counter: Counter[str] = Counter()

    category_views = {k: 0 for k in CATEGORY_KEYS}
    category_searches = {k: 0 for k in CATEGORY_KEYS}

    session_times: dict[str, list[datetime]] = defaultdict(list)

    for ev in events:
        et = str(ev.get("event_type", "")).lower()
        domain = str(ev.get("domain", "") or "").lower()
        meta = ev.get("metadata") or {}
        if not isinstance(meta, dict):
            meta = {}

        sid = str(ev.get("session_id") or "")
        if sid:
            sessions.add(sid)
            session_event_counts[sid] += 1

        ts = _parse_ts(ev.get("created_at"))
        if sid and ts:
            session_times[sid].append(ts)

        if domain:
            domain_counter[domain] += 1

        cat = _meta_category(meta)
        if cat:
            category_counter[cat] += 1
            if cat in category_views:
                category_views[cat] += 1

        if et in CONTENT_VIEW_TYPES:
            content_views += 1
            dwell = _num(meta, "dwell_time")
            if dwell > 0:
                dwell_times.append(dwell)
            scroll = _num(meta, "scroll_depth")
            if scroll > 0:
                scroll_depths.append(scroll)
            if cat and cat in category_views:
                category_views[cat] += 0  # already counted

        if et in SEARCH_TYPES:
            searches += 1
            if cat and cat in category_searches:
                category_searches[cat] += 1

        if et in INTERACTION_TYPES:
            interactions += 1

        if et in CAMPAIGN_IMPRESSION_TYPES:
            campaign_impressions += 1
        if et in CAMPAIGN_CLICK_TYPES:
            campaign_clicks += 1
        if et in CONVERSION_TYPES:
            conversions += 1
        if et in TRANSACTION_TYPES:
            transactions += 1
        if et in REWARD_TYPES:
            rewards_claimed += 1

    for sid, times in session_times.items():
        if len(times) >= 2:
            duration = (max(times) - min(times)).total_seconds()
            if duration > 0:
                session_durations.append(duration)

    session_count = max(len(sessions), 1)
    events_per_session = len(events) / session_count

    features: dict[str, float] = {
        "session_count": float(len(sessions)),
        "average_session_duration": _mean(session_durations),
        "events_per_session": float(events_per_session),
        "average_dwell_time": _mean(dwell_times),
        "average_scroll_depth": _mean(scroll_depths),
        "content_views_count": float(content_views),
        "search_count": float(searches),
        "interaction_count": float(interactions),
        "campaign_impressions": float(campaign_impressions),
        "campaign_clicks": float(campaign_clicks),
        "campaign_ctr": _safe_div(campaign_clicks, campaign_impressions),
        "reward_claim_count": float(rewards_claimed),
        "conversion_count": float(conversions),
        "transaction_count": float(transactions),
        "transaction_frequency": _safe_div(transactions, len(events)),
    }

    for cat in CATEGORY_KEYS:
        features[f"{cat}_views"] = float(category_views.get(cat, 0))
        features[f"{cat}_searches"] = float(category_searches.get(cat, 0))

    top_cat, top_cat_n = _top_counter(category_counter)
    top_dom, top_dom_n = _top_counter(domain_counter)
    features["top_category_count"] = float(top_cat_n)
    features["top_domain_count"] = float(top_dom_n)
    features[f"top_category_is_{top_cat}"] = 1.0 if top_cat else 0.0
    features[f"top_domain_is_{top_dom}"] = 1.0 if top_dom else 0.0

    return features


def features_to_dataframe(rows: list[dict[str, float]]) -> pd.DataFrame:
    df = pd.DataFrame(rows).fillna(0.0)
    return df


def feature_column_order() -> list[str]:
    """Stable column order for training and inference."""
    base = [
        "session_count",
        "average_session_duration",
        "events_per_session",
        "average_dwell_time",
        "average_scroll_depth",
        "content_views_count",
        "search_count",
        "interaction_count",
        "campaign_impressions",
        "campaign_clicks",
        "campaign_ctr",
        "reward_claim_count",
        "conversion_count",
        "transaction_count",
        "transaction_frequency",
        "top_category_count",
        "top_domain_count",
    ]
    cats = []
    for c in CATEGORY_KEYS:
        cats.extend([f"{c}_views", f"{c}_searches"])
    return base + cats


def _empty_features() -> dict[str, float]:
    f = {k: 0.0 for k in feature_column_order()}
    for c in CATEGORY_KEYS:
        f[f"{c}_views"] = 0.0
        f[f"{c}_searches"] = 0.0
    return f


def _mean(values: list[float]) -> float:
    return float(sum(values) / len(values)) if values else 0.0


def _safe_div(num: float, den: float) -> float:
    return float(num / den) if den > 0 else 0.0


def _top_counter(counter: Counter[str]) -> tuple[str, int]:
    if not counter:
        return "", 0
    key, count = counter.most_common(1)[0]
    return key, count
