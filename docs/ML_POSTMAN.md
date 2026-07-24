# ML service — 71-feature behavioral pipeline

Base URL (docker-compose): `http://localhost:8000`

Flutter / embedding host builds a **session** (app usage + UI text + optional
in-app behavioral events). The service runs `extract_features` → **71 floats**,
then predicts an intent.

| Index | Group | Source |
|------:|-------|--------|
| 0–11 | App time ratios | App usage scan |
| 12–23 | App switch ratios | App usage scan |
| 24–31 | UI keyword ratios | Accessibility text |
| 32–37 | Temporal | Session clock |
| 38–42 | Session context | Session |
| 43–46 | Historical (consented only) | Profile |
| 47–70 | In-app funnel / behavioral | Embedding host events |

Categories: `fashion`, `shopping`, `crypto`, `fintech`, `coffee`, `food`,
`news`, `social`, `travel`, `fitness`, `banking`, `other`
(see `ml/data/app_category_map.py`, `ml/data/ui_keyword_map.py`).

---

## Retrain (local)

```bash
cd ml
python3 -m venv .venv
source .venv/bin/activate   # Windows: .venv\Scripts\activate
pip install -r requirements.txt

# One-shot: generate 71-feature data → train keras → TFLite → SDK assets
python main.py
```

Or step-by-step:

```bash
cd ml/training
python generate_synthetic_data.py   # writes data/processed/training_data.csv
python train_model.py               # writes models/best_model.keras + label_map.json
python convert_to_tflite.py         # writes TFLite + skykin-sdk/lib/ml/assets/*
```

Then restart the ML container so it loads the new keras weights:

```bash
docker compose up --build -d ml-service
# or locally:
cd ml && uvicorn app:app --host 0.0.0.0 --port 8000
```

Confirm:

```bash
curl http://localhost:8000/health
# expect: "feature_size": 71, "model_loaded": true
```

---

## Postman

### 1) Health

`GET http://localhost:8000/health`

### 2) Predict with behavioral events (shopping / funnel)

`POST http://localhost:8000/predict-intent`  
`Content-Type: application/json`

```json
{
  "user_id": "test-user-001",
  "session": {
    "app_usage": {
      "shopping": { "minutes": 12.0, "switches": 5 },
      "fashion":  { "minutes": 6.0,  "switches": 3 },
      "social":   { "minutes": 3.0,  "switches": 2 },
      "other":    { "minutes": 2.0,  "switches": 1 }
    },
    "ui_signals": {
      "shopping": 10,
      "fashion":  6
    },
    "behavioral_events": {
      "has_data": 1.0,
      "actions": {
        "browseCategory": 10,
        "viewItem": 12,
        "stageTransaction": 8,
        "initiateCheckout": 5,
        "abandonTransaction": 0
      },
      "categories": {
        "shopping": 20,
        "fashion": 10,
        "fintech": 4,
        "coffee": 0,
        "crypto": 0,
        "travel": 0,
        "fitness": 0,
        "food": 0
      }
    },
    "session_start": "2026-07-22T19:30:00",
    "session_duration_minutes": 24.0,
    "total_switches": 11,
    "is_first_session": 0.0
  },
  "historical": {
    "days_with_intent": 4,
    "avg_confidence": 0.82,
    "last_seen_days_ago": 1,
    "consistency_score": 0.7
  }
}
```

Expected: intent near `shopping_interest` (or fashion), `source: "model"` after retrain.

### 3) Fashion interest (behavioral browse/view heavy)

```json
{
  "user_id": "test-user-002",
  "session": {
    "app_usage": {
      "fashion":  { "minutes": 10.0, "switches": 4 },
      "shopping": { "minutes": 5.0,  "switches": 2 },
      "social":   { "minutes": 3.0,  "switches": 2 }
    },
    "ui_signals": { "fashion": 14, "shopping": 5 },
    "behavioral_events": {
      "has_data": 1.0,
      "actions": {
        "browseCategory": 12,
        "viewItem": 8,
        "stageTransaction": 3,
        "initiateCheckout": 1,
        "abandonTransaction": 0
      },
      "categories": {
        "fashion": 18,
        "shopping": 6,
        "coffee": 0,
        "crypto": 0,
        "fintech": 0,
        "travel": 0,
        "fitness": 0,
        "food": 0
      }
    },
    "session_start": "2026-07-22T10:30:00",
    "session_duration_minutes": 22.0,
    "total_switches": 8,
    "is_first_session": 1.0
  }
}
```

### 4) No behavioral data (legacy accessibility-only host)

Omit `behavioral_events` or set `"has_data": 0.0` — features 47–70 stay zero; model still runs on groups 0–46.

### 5) Raw 71-float debug path

```json
{
  "user_id": "debug",
  "features": [0.0]
}
```

Must be **exactly 71** floats. Requires `model_loaded: true`.
