# ML service — current contract (47-feature vector)

Base URL (docker-compose): `http://localhost:8000`

The model does **not** take SDK event lists. Flutter (accessibility + app-usage scan)
builds a **session**, the service runs `extract_features` → **47 floats**, then predicts.

Feature groups (see `ml/training/feature_engineering.py`):

| Index | Group | Source |
|------:|-------|--------|
| 0–11 | App time ratios (12 categories) | App usage scan |
| 12–23 | App switch ratios | App usage scan |
| 24–31 | UI keyword signal ratios | Accessibility text |
| 32–37 | Temporal (hour/day cyclical, weekend, morning) | Session clock |
| 38–42 | Session context (duration, switches, diversity, …) | Session |
| 43–46 | Historical intent (consented only; else 0) | Profile |

App categories: `fashion`, `shopping`, `crypto`, `fintech`, `coffee`, `food`,
`news`, `social`, `travel`, `fitness`, `banking`, `other`
(package → category via `ml/data/app_category_map.py`).

UI signal keys: `fashion`, `crypto`, `coffee`, `fintech`, `travel`, `fitness`,
`shopping`, `food` (keyword hits via `ml/data/ui_keyword_map.py`).

---

## 1) Health

`GET http://localhost:8000/health`

```json
{
  "status": "ok",
  "model_loaded": false,
  "model_error": "...",
  "feature_size": 47
}
```

---

## 2) Predict from session (normal path)

`POST http://localhost:8000/predict-intent`  
`Content-Type: application/json`

Example: fashion-heavy usage + fashion/shopping UI text (morning weekday).

```json
{
  "user_id": "test-user-001",
  "session": {
    "app_usage": {
      "fashion":  { "minutes": 9.5,  "switches": 4 },
      "shopping": { "minutes": 6.0,  "switches": 3 },
      "social":   { "minutes": 4.0,  "switches": 2 },
      "other":    { "minutes": 2.5,  "switches": 1 }
    },
    "ui_signals": {
      "fashion":  12,
      "shopping": 7,
      "food":     1
    },
    "session_start": "2026-07-14T10:30:00",
    "session_duration_minutes": 22.0,
    "total_switches": 10,
    "is_first_session": 1.0
  },
  "historical": {
    "days_with_intent": 5,
    "avg_confidence": 0.81,
    "last_seen_days_ago": 1,
    "consistency_score": 0.72
  }
}
```

Omit `historical` (or send all zeros) for non-consented / first-time sessions.

Crypto-heavy example:

```json
{
  "user_id": "test-user-002",
  "session": {
    "app_usage": {
      "crypto":  { "minutes": 11.0, "switches": 5 },
      "news":    { "minutes": 5.0,  "switches": 3 },
      "fintech": { "minutes": 3.0,  "switches": 2 },
      "other":   { "minutes": 2.0,  "switches": 1 }
    },
    "ui_signals": {
      "crypto":  18,
      "fintech": 6
    },
    "session_start": "2026-07-14T08:15:00",
    "session_duration_minutes": 21.0,
    "total_switches": 11,
    "is_first_session": 0.0
  }
}
```

Expected shape:

```json
{
  "user_id": "test-user-001",
  "intent": "fashion_interest",
  "confidence": 0.xx,
  "reward_triggered": true,
  "top_signals": ["fashion", "shopping", "social"],
  "threshold": 0.7,
  "source": "heuristic_fallback"
}
```

`source` is `"model"` once `ml/models/best_model.keras` is present.

---

## 3) Predict from raw 47-float vector (debug only)

```json
{
  "user_id": "test-user-001",
  "features": [0.0, 0.0, /* … exactly 47 floats … */]
}
```

Requires the keras model loaded (`model_loaded: true`).
