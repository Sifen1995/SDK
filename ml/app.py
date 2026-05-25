from fastapi import FastAPI



"""
Skykin — Intent Prediction API
File:    app.py
Run:     uvicorn app:app --reload
Docs:    http://127.0.0.1:8000/docs
"""

import os
import joblib
from pathlib import Path
from fastapi import FastAPI, HTTPException, Header, Depends
from pydantic import BaseModel, field_validator

# ─────────────────────────────────────────────────────────────
#  CONFIGURATION
# ─────────────────────────────────────────────────────────────

   # move to .env / secret manager in production
MODEL_PATH = Path(__file__).parent / "models" / "intent_model.pkl"

VALID_EVENT_TYPES = {
    "search",
    "product_view",
    "category_view",
    "add_to_cart",
    "remove_from_cart",
    "signup_complete",
    "checkout_started",
}


# ─────────────────────────────────────────────────────────────
#  LOAD MODEL — once at startup, not on every request
# ─────────────────────────────────────────────────────────────

if not MODEL_PATH.exists():
    raise RuntimeError(
        f"Model not found at '{MODEL_PATH}'.\n"
        "Run  python model.py  first to train and save the model."
    )

_artifact  = joblib.load(MODEL_PATH)
_pipeline  = _artifact["pipeline"]
_threshold = _artifact["threshold"]
_intents   = _artifact["intents"]
_version   = _artifact.get("model_version", "unknown")

print(f"Model loaded — version {_version} | intents: {_intents} | threshold: {_threshold}")


# ─────────────────────────────────────────────────────────────
#  FASTAPI APP
# ─────────────────────────────────────────────────────────────

app = FastAPI(
    title="Skykin Intent Prediction API",
    description="Predicts user intent from a session of events and triggers rewards.",
    version="1.0.0",
)


# ─────────────────────────────────────────────────────────────
#  AUTHENTICATION
# ─────────────────────────────────────────────────────────────

# def verify_api_key(x_api_key: str = Header(..., description="Your Skykin API key")):
#     """Dependency — validates the X-API-Key header on every protected route."""
#     if x_api_key != API_KEY:
#         raise HTTPException(status_code=401, detail="Invalid or missing API key.")
#     return x_api_key


# ─────────────────────────────────────────────────────────────
#  REQUEST / RESPONSE SCHEMAS
# ─────────────────────────────────────────────────────────────

class Event(BaseModel):
    event_type: str
    metadata:   dict = {}

    @field_validator("event_type")
    @classmethod
    def validate_event_type(cls, v):
        if v not in VALID_EVENT_TYPES:
            raise ValueError(
                f"'{v}' is not a valid event_type. "
                f"Choose from: {sorted(VALID_EVENT_TYPES)}"
            )
        return v


class PredictRequest(BaseModel):
    user_id: str
    events:  list[Event]





class PredictResponse(BaseModel):
    user_id:         str
    intent:          str | None
    confidence:      float | None
    threshold:       float
    reward_triggered: bool
    





# ─────────────────────────────────────────────────────────────
#  HELPERS
# ─────────────────────────────────────────────────────────────

def build_session_text(events: list[Event]) -> str:
    """
    Converts a list of Event objects into a single text string
    for the TF-IDF vectorizer — the same transformation used
    during training in generate_data.py.
    """
    tokens = []
    for event in events:
        tokens.append(event.event_type)
        for value in event.metadata.values():
            cleaned = str(value).lower().replace("_", " ").replace("&", "")
            tokens.append(cleaned)
    return " ".join(tokens)


def run_prediction(events: list[Event]) -> dict:
    """
    Runs the ML pipeline on a session and returns intent,
    confidence score, and whether the threshold is met.
    """
    session_text = build_session_text(events)
    proba        = _pipeline.predict_proba([session_text])[0]
    confidence   = float(proba.max())
    intent       = _pipeline.classes_[proba.argmax()]

    return {
        "intent":            intent,
        "confidence":        round(confidence, 4),
        "reward_triggered":  confidence >= _threshold,
    }


# ─────────────────────────────────────────────────────────────
#  ROUTES
# ─────────────────────────────────────────────────────────────


@app.get("/ping")
def ping() -> dict[str, str]:
    return {"status": "ok", "message": "ML service is running"}

@app.get("/health", tags=["System"])
def health_check():
    """
    Public health check — no API key required.
    Returns model version and loaded intents.
    """
    return {
        "status":          "ok",
        "model_version":   _version,
        "intents":         _intents,
        "threshold":       _threshold,
    }


@app.post(
    "/predict-intent",
    response_model=PredictResponse,
    tags=["Prediction"],
    # dependencies=[Depends(verify_api_key)],
)
def predict_intent(body: PredictRequest):
    """
    Accepts a user session (list of events), runs intent
    prediction, and returns the intent + reward if the
    confidence meets the threshold.

    Requires header:  X-API-Key: <your key>
    """
    if not body.events:
        raise HTTPException(
            status_code=422,
            detail="The 'events' list cannot be empty."
        )

    result = run_prediction(body.events)
   

    

    return PredictResponse(
        user_id=          body.user_id,
        intent=           result["intent"],
        confidence=       result["confidence"],
        threshold=        _threshold,
        reward_triggered= result["reward_triggered"],
        
    )


@app.get(
    "/intents",
    tags=["System"],
    # dependencies=[Depends(verify_api_key)],
)
def list_intents():
    """Returns the list of intent labels the model can predict."""
    return {"intents": _intents}


