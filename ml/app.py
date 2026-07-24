"""ML inference HTTP service — FastAPI entrypoint for uvicorn app:app.

Request contract matches the 71-feature pipeline in training/feature_engineering.py:
  app usage + UI signals + in-app behavioral events (embedding host) → extract_features → model.
"""

from __future__ import annotations

import logging
from typing import Any

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field, model_validator

from inference.model_loader import load_model
from inference.predictor import predict_from_features, predict_from_session
from training.feature_engineering import CATEGORIES, FEATURE_SIZE

logger = logging.getLogger("ml-service")
logging.basicConfig(level=logging.INFO)

app = FastAPI(title="Skykin ML Service", docs_url=None, redoc_url=None, openapi_url=None)

_artifact = None
_load_error: str | None = None


@app.on_event("startup")
def _startup() -> None:
    global _artifact, _load_error
    try:
        _artifact = load_model()
        _load_error = None
        logger.info("model loaded version=%s intents=%d", _artifact.model_version, len(_artifact.intents))
    except Exception as exc:  # noqa: BLE001 — keep service up for contract testing
        _artifact = None
        _load_error = str(exc)
        logger.warning("model not loaded (%s); /predict-intent will use heuristic fallback", exc)


class CategoryUsage(BaseModel):
    minutes: float = 0.0
    switches: int = 0


class BehavioralActions(BaseModel):
    browseCategory: int = 0
    viewItem: int = 0
    stageTransaction: int = 0
    initiateCheckout: int = 0
    abandonTransaction: int = 0


class BehavioralEvents(BaseModel):
    """In-app funnel events from the embedding host (e-commerce / Telebirr-style flows)."""

    has_data: float = 0.0
    actions: BehavioralActions = Field(default_factory=BehavioralActions)
    categories: dict[str, int] = Field(default_factory=dict)


class SessionPayload(BaseModel):
    """Raw session from Flutter accessibility + app-usage + host behavioral embedding."""

    app_usage: dict[str, CategoryUsage] = Field(default_factory=dict)
    ui_signals: dict[str, int] = Field(default_factory=dict)
    behavioral_events: BehavioralEvents | None = None
    session_start: str | None = None
    session_duration_minutes: float = 1.0
    total_switches: int = 1
    is_first_session: float = 0.0


class HistoricalPayload(BaseModel):
    """Optional pseudonymous history (consented users only; zeros otherwise)."""

    days_with_intent: int = 0
    avg_confidence: float = 0.0
    last_seen_days_ago: int = 30
    consistency_score: float = 0.0


class PredictRequest(BaseModel):
    user_id: str
    session: SessionPayload | None = None
    historical: HistoricalPayload | None = None
    features: list[float] | None = None

    @model_validator(mode="after")
    def require_session_or_features(self) -> PredictRequest:
        if self.session is None and self.features is None:
            raise ValueError("provide either session or features")
        if self.features is not None and len(self.features) != FEATURE_SIZE:
            raise ValueError(f"features must have length {FEATURE_SIZE}, got {len(self.features)}")
        return self


class PredictResponse(BaseModel):
    user_id: str | None = None
    intent: str
    confidence: float
    reward_triggered: bool
    top_signals: list[str] = Field(default_factory=list)
    threshold: float | None = None
    source: str = "model"


def _session_dict(session: SessionPayload) -> dict:
    out: dict[str, Any] = {
        "app_usage": {
            cat: {"minutes": usage.minutes, "switches": usage.switches}
            for cat, usage in session.app_usage.items()
        },
        "ui_signals": dict(session.ui_signals),
        "session_start": session.session_start,
        "session_duration_minutes": session.session_duration_minutes,
        "total_switches": session.total_switches,
        "is_first_session": session.is_first_session,
    }
    if session.behavioral_events is not None:
        be = session.behavioral_events
        out["behavioral_events"] = {
            "has_data": be.has_data,
            "actions": be.actions.model_dump(),
            "categories": dict(be.categories),
        }
    return out


def _heuristic_from_session(user_id: str, session: dict) -> dict:
    """Fallback when keras weights are missing — score from app/UI/behavioral dominance."""
    scores: dict[str, float] = {
        "fashion_interest": 0.12,
        "crypto_interest": 0.12,
        "coffee_interest": 0.10,
        "fintech_interest": 0.12,
        "travel_intent": 0.10,
        "fitness_interest": 0.10,
        "shopping_interest": 0.14,
        "food_interest": 0.12,
        "no_clear_intent": 0.20,
    }
    intent_for_cat = {
        "fashion": "fashion_interest",
        "shopping": "shopping_interest",
        "crypto": "crypto_interest",
        "fintech": "fintech_interest",
        "coffee": "coffee_interest",
        "food": "food_interest",
        "travel": "travel_intent",
        "fitness": "fitness_interest",
        "banking": "fintech_interest",
        "news": "no_clear_intent",
        "social": "no_clear_intent",
        "other": "no_clear_intent",
    }

    app_usage = session.get("app_usage", {})
    ui_signals = session.get("ui_signals", {})
    for cat in CATEGORIES:
        minutes = float(app_usage.get(cat, {}).get("minutes", 0) or 0)
        switches = float(app_usage.get(cat, {}).get("switches", 0) or 0)
        ui = float(ui_signals.get(cat, 0) or 0)
        bump = minutes * 0.04 + switches * 0.03 + ui * 0.02
        intent = intent_for_cat.get(cat, "no_clear_intent")
        scores[intent] += bump

    behavioral = session.get("behavioral_events") or {}
    if float(behavioral.get("has_data", 0) or 0) > 0:
        actions = behavioral.get("actions") or {}
        stage = float(actions.get("stageTransaction", 0) or 0)
        checkout = float(actions.get("initiateCheckout", 0) or 0)
        cats = behavioral.get("categories") or {}
        for cat, count in cats.items():
            intent = intent_for_cat.get(cat, "no_clear_intent")
            scores[intent] += float(count or 0) * 0.03
        if stage + checkout >= 3:
            scores["shopping_interest"] += 0.25

    shopping_m = float(app_usage.get("shopping", {}).get("minutes", 0) or 0)
    fashion_m = float(app_usage.get("fashion", {}).get("minutes", 0) or 0)
    shopping_ui = float(ui_signals.get("shopping", 0) or 0)
    if shopping_m + fashion_m > 8 and shopping_ui >= 5:
        scores["shopping_interest"] += 0.2

    intent = max(scores, key=scores.get)
    confidence = min(0.99, round(scores[intent] / (scores[intent] + 0.5), 3))
    threshold = 0.70

    ranked = sorted(
        (
            (
                float(app_usage.get(c, {}).get("minutes", 0) or 0)
                + float(ui_signals.get(c, 0) or 0),
                c,
            )
            for c in CATEGORIES
        ),
        reverse=True,
    )
    top = [name for score, name in ranked if score > 0][:3]

    return {
        "user_id": user_id,
        "intent": intent,
        "confidence": confidence,
        "threshold": threshold,
        "reward_triggered": confidence >= threshold,
        "top_signals": top,
        "source": "heuristic_fallback",
    }


@app.get("/health")
def health() -> dict:
    return {
        "status": "ok",
        "model_loaded": _artifact is not None,
        "model_error": _load_error,
        "feature_size": FEATURE_SIZE,
    }


@app.post("/predict-intent", response_model=PredictResponse)
def predict_intent(req: PredictRequest) -> PredictResponse:
    if not req.user_id.strip():
        raise HTTPException(status_code=400, detail="user_id is required")

    historical = req.historical.model_dump() if req.historical else None

    if req.features is not None:
        if _artifact is None:
            raise HTTPException(
                status_code=503,
                detail="model not loaded; features path requires best_model.keras",
            )
        try:
            result = predict_from_features(_artifact, req.user_id, req.features)
            result["source"] = "model"
            return PredictResponse(**result)
        except Exception as exc:  # noqa: BLE001
            raise HTTPException(status_code=500, detail=f"prediction failed: {exc}") from exc

    assert req.session is not None
    session = _session_dict(req.session)

    if _artifact is not None:
        try:
            result = predict_from_session(_artifact, req.user_id, session, historical)
            result["source"] = "model"
            return PredictResponse(**result)
        except Exception as exc:  # noqa: BLE001
            logger.exception("model prediction failed; falling back to heuristic")
            result = _heuristic_from_session(req.user_id, session)
            result["source"] = f"heuristic_after_error:{exc}"
            return PredictResponse(**result)

    return PredictResponse(**_heuristic_from_session(req.user_id, session))
