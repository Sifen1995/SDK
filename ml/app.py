"""
Skykin — Behavioral Intent Prediction API (stateless microservice).

Receives events from the Go backend over HTTP, engineers features, predicts intent.
No database connection — all data comes from the backend.

Run: uvicorn app:app --host 0.0.0.0 --port 8000
"""

from __future__ import annotations

import sys
from pathlib import Path

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field, field_validator

ML_ROOT = Path(__file__).resolve().parent
sys.path.insert(0, str(ML_ROOT))

from inference.model_loader import load_model
from inference.predictor import predict_from_events

artifact = load_model()

VALID_EVENT_TYPES = {
    "session_started",
    "screen_viewed",
    "content_viewed",
    "search_performed",
    "interaction_received",
    "scroll_activity",
    "notification_opened",
    "campaign_impression",
    "campaign_clicked",
    "conversion_completed",
    "transaction_completed",
    "reward_claimed",
}

app = FastAPI(
    title="Skykin Behavioral Intent API",
    description="Stateless intent prediction from behavioral events supplied by the backend.",
    version="3.0.0",
)


class EventPayload(BaseModel):
    event_type: str
    domain: str = ""
    screen_name: str = ""
    metadata: dict = Field(default_factory=dict)
    session_id: str = ""
    created_at: str | None = None

    @field_validator("event_type")
    @classmethod
    def validate_event_type(cls, v: str) -> str:
        if v not in VALID_EVENT_TYPES:
            raise ValueError(
                f"'{v}' is not a valid event_type. "
                f"Choose from: {sorted(VALID_EVENT_TYPES)}"
            )
        return v


class PredictRequest(BaseModel):
    user_id: str = Field(..., description="SDK external user id")
    events: list[EventPayload] = Field(..., min_length=1)


class PredictResponse(BaseModel):
    user_id: str
    intent: str | None
    confidence: float | None
    threshold: float
    reward_triggered: bool
    top_signals: list[str] = []


@app.get("/ping")
def ping():
    return {"status": "ok", "model_version": artifact.model_version}


@app.get("/health")
def health():
    return {
        "status": "ok",
        "model_version": artifact.model_version,
        "intents": artifact.intents,
        "threshold": artifact.threshold,
        "feature_count": len(artifact.feature_columns),
    }


@app.post("/predict-intent", response_model=PredictResponse)
def predict_intent(body: PredictRequest):
    if len(body.events) < 3:
        return PredictResponse(
            user_id=body.user_id,
            intent=None,
            confidence=None,
            threshold=artifact.threshold,
            reward_triggered=False,
            top_signals=["insufficient_history"],
        )

    events = [e.model_dump() for e in body.events]
    result = predict_from_events(artifact, body.user_id, events)
    return PredictResponse(**result)


@app.get("/intents")
def list_intents():
    return {"intents": artifact.intents}
