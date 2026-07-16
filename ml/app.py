"""ML inference HTTP service — FastAPI entrypoint for uvicorn app:app.

Request contract matches the 47-feature pipeline in training/feature_engineering.py:
  app usage (accessibility + usage scan) + UI text signals → extract_features → model.
"""

from __future__ import annotations

import logging
from typing import Any

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field, model_validator

from inference.model_loader import load_model
from inference.predictor import predict_from_features, predict_from_session
from training.feature_engineering import CATEGORIES

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


class SessionPayload(BaseModel):
    """Raw session from Flutter accessibility + app-usage scan (before feature extraction)."""

    app_usage: dict[str, CategoryUsage] = Field(default_factory=dict)
    ui_signals: dict[str, int] = Field(default_factory=dict)
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
    # Primary path: session dict → 47 features inside the service
    session: SessionPayload | None = None
    historical: HistoricalPayload | None = None
    # Debug path: send the 47-float vector already built
    features: list[float] | None = None

    @model_validator(mode="after")
    def require_session_or_features(self) -> PredictRequest:
        if self.session is None and self.features is None:
            raise ValueError("provide either session (app_usage + ui_signals) or features (length 47)")
        if self.features is not None and len(self.features) != 47:
            raise ValueError(f"features must have length 47, got {len(self.features)}")
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
    return {
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


def _heuristic_from_session(user_id: str, session: dict) -> dict:
    """Fallback when keras weights are missing — score from dominant app/UI categories."""
    scores: dict[str, float] = {
        "fashion_interest": 0.12,
        "crypto_interest": 0.12,
        "coffee_interest": 0.10,
        "fintech_interest": 0.12,
        "travel_intent": 0.10,
        "fitness_interest": 0.10,
        "shopping_interest": 0.14,
        "food_interest": 0.12,
        "abandoned_cart": 0.08,
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

    shopping_m = float(app_usage.get("shopping", {}).get("minutes", 0) or 0)
    fashion_m = float(app_usage.get("fashion", {}).get("minutes", 0) or 0)
    shopping_ui = float(ui_signals.get("shopping", 0) or 0)
    if shopping_m + fashion_m > 8 and shopping_ui >= 5:
        scores["abandoned_cart"] += 0.35

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
        "feature_size": 47,
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
