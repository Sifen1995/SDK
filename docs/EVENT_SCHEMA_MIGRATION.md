# Event Schema & ML Pipeline Migration Guide

## Overview

The platform moved from product-specific events (`product_view`, `add_to_cart`) to **domain-agnostic behavioral events** (`content_viewed`, `search_performed`, etc.).

Intent prediction now uses **engineered behavioral features** (dwell time, category counters, campaign CTR) instead of raw event text (TF-IDF).

---

## Do I need to delete tables in pgAdmin?

**No — you do not need to delete existing data.**

1. Run the SQL migration (or let GORM `AutoMigrate` create the new `events` table).
2. The migration **copies** rows from `sdk_events` → `events` when `sdk_events` exists.
3. Legacy table `sdk_events` is **left in place** for audit/rollback.

**Optional fresh start (dev only):** Drop `sdk_events` and `events` in pgAdmin, then `docker compose up --build`. GORM will recreate `events`; reward rules re-seed on startup.

---

## Apply database migration

### Option A — GORM AutoMigrate (default on backend start)

```bash
docker compose up --build
```

Creates/updates the `events` table from `EventRecord`.

### Option B — SQL migration file (recommended for production)

File: `internal/platform/database/migrations/20260601120000_refactor_events_schema.sql`

Run manually in pgAdmin Query Tool or:

```bash
psql "postgres://skykin_user:password@localhost:5435/skykin_db?sslmode=disable" \
  -f internal/platform/database/migrations/20260601120000_refactor_events_schema.sql
```

---

## New `events` table schema

| Column          | Type         | Notes                          |
|-----------------|--------------|--------------------------------|
| id              | UUID PK      |                                |
| event_id        | UUID UNIQUE  | Client-generated UUID          |
| user_id         | UUID         | Internal `users.id`            |
| application_id  | UUID         | Optional                       |
| session_id      | UUID         | Optional                       |
| event_type      | VARCHAR(100) | See supported types below      |
| domain          | VARCHAR(100) | e.g. crypto, fashion, fintech  |
| screen_name     | VARCHAR(255) |                                |
| metadata        | JSONB        | Flexible context               |
| device_type     | VARCHAR(50)  |                                |
| platform        | VARCHAR(50)  |                                |
| app_version     | VARCHAR(50)  |                                |
| created_at      | TIMESTAMPTZ  |                                |

**Indexes:** `user_id`, `session_id`, `event_type`, `domain`, `created_at`, GIN on `metadata`.

---

## Supported event types

- `session_started`
- `screen_viewed`
- `content_viewed`
- `search_performed`
- `interaction_received`
- `scroll_activity`
- `notification_opened`
- `campaign_impression`
- `campaign_clicked`
- `conversion_completed`
- `transaction_completed`
- `reward_claimed`

---

## SDK example payload

```http
POST /api/v1/events
X-API-Key: pk_live_...
Content-Type: application/json
```

```json
{
  "user_id": "user_abc_123",
  "events": [
    {
      "event_id": "550e8400-e29b-41d4-a716-446655440001",
      "event_type": "content_viewed",
      "domain": "crypto",
      "session_id": "660e8400-e29b-41d4-a716-446655440002",
      "screen_name": "asset_details",
      "metadata": {
        "category": "crypto",
        "asset": "bitcoin",
        "dwell_time": 90
      },
      "device_type": "mobile",
      "platform": "android",
      "app_version": "1.2.0",
      "created_at": "2026-06-01T12:00:00Z"
    }
  ]
}
```

Response: `202 Accepted` with per-event status. Intent/reward runs asynchronously.

---

## Retrain ML model

```bash
cd ml
pip install -r requirements.txt
python -m training.train
```

Artifact: `ml/models/intent_model.pkl`

Rebuild ML container:

```bash
docker compose up --build ml-service
```

---

## ML architecture

The ML service is **stateless** — it has no database connection. The Go backend loads events from PostgreSQL and sends them to ML over HTTP.

```
ml/
├── training/
│   ├── feature_engineering.py   # Behavioral features
│   ├── label_generation.py      # Rule-based labels
│   ├── dataset_builder.py       # Synthetic + export datasets
│   └── train.py                 # RandomForest → .pkl
├── inference/
│   ├── model_loader.py
│   └── predictor.py
└── app.py                       # POST /predict-intent { user_id, events }
```

**Inference flow:**

1. Backend receives SDK events → stores in `events` table
2. Backend loads recent user events from DB
3. Backend `POST`s `{ user_id, events }` to ML `/predict-intent`
4. ML builds feature vector, predicts intent + confidence + `top_signals`
5. Backend persists intent and triggers rewards

---

## Intent labels

- `fashion_interest`
- `crypto_interest`
- `food_interest`
- `education_interest`
- `gaming_interest`
- `fintech_interest`
- `general_interest`

---

## Breaking changes for Flutter SDK

1. Use **UUID** for `event_id` and `session_id`.
2. Send **batch** to `POST /api/v1/events` (not single legacy types).
3. Put product-specific fields in `metadata` only (e.g. `category`, `asset`, `dwell_time`).
4. Include `domain` on every event.
