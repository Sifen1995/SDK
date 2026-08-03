# Skykin Platform

Skykin is an intent-aware advertising and SDK platform. Mobile apps integrate the Flutter SDK; the backend scores on-device behavior into intents, matches eligible campaigns, and delivers ads (including SMS+). Advertisers, developers, and operators manage everything through three web portals.

---

## What it does

| Audience | Product surface | Job |
|---|---|---|
| End users (via Flutter SDK) | Consent, intent ingest, rewards, fraud signals | Privacy-preserving identity + ad / SMS delivery |
| Developers | `portal/` | Register apps, issue API keys, manage SDK access |
| Advertisers | `ad-portal/` | Campaigns, budgets, audiences, billing |
| Operators | `admin-portal/` | Moderation, analytics, catalog, RBAC |

**Core runtime flow**

```
Flutter SDK
    │  X-API-Key + HMAC signature
    ▼
Gin API (:8081)
    │
    ├─ Consent → provision / map pseudonymous user
    ├─ Intent ingest → ML / rules → eligible campaigns
    ├─ Ad select → in-app creative  OR  SMS+ dispatch
    └─ Telemetry / rewards / threat reports
         │
         ├── Postgres (source of truth)
         ├── Redis (cache, queues, streams)
         └── In-process event bus (sagas & side effects)
```

---

## Repository layout

```
SDK/
├── cmd/api/                 # Go API entrypoint + Swagger
├── configs/                 # Env / config / logging
├── deployments/             # Dockerfiles, compose, k8s
├── docs/                    # Architecture + OpenAPI artifacts
├── internal/                # Backend bounded contexts (hexagonal)
│   ├── platform/            # Bootstrap, router, DB, Redis, messaging
│   ├── auth/                # Developers, apps, API keys
│   ├── consent/             # SDK consent + pseudonymous mappings
│   ├── users/               # User provisioning
│   ├── intents/             # Intent ingest + ad fetch
│   ├── campaigns/           # Campaigns, eligibility, ad selection
│   ├── delivery/            # SMS+ dispatch, tracking, debug
│   ├── billing/             # Plans, rates, subscriptions, invoices
│   ├── audience/            # Segments & purchases
│   ├── analytics/           # Aggregates / reporting
│   ├── rewards/             # Reward rules & grants
│   ├── ad_portal/ · admin/  # Portal & operator APIs
│   ├── permissions/         # RBAC
│   ├── fraud/               # Blocklists, patterns, threat reports
│   ├── events/ · geofencing/# Present; some routes not mounted yet
│   └── …
├── ml/                      # Intent classification (Keras → TFLite)
├── fraude_ml/               # SMS / URL scam detector (sklearn → TFLite)
├── portal/ · ad-portal/ · admin-portal/   # React portals
├── shared/                  # Shared UI package (@skykin/ui)
├── skykin-sdk/              # On-device ML assets for Flutter
├── postman/                 # API collection + local environment
└── tests/                   # unit / integration / e2e
```

Backend modules follow **ports & adapters**:

```
internal/<module>/
  domain/            # entities, repository ports, events
  application/       # use cases (depend on interfaces only)
  infrastructure/    # Postgres / Redis / provider adapters
  interfaces/http/   # Gin handlers + DTOs
  consumers/         # event-bus handlers
  worker/            # Redis stream / queue workers
```

Cross-module wiring lives only in `internal/platform/bootstrap/` and `route/`. Details: [`docs/MODULE_ARCHITECTURE_EXPLANATION.md`](docs/MODULE_ARCHITECTURE_EXPLANATION.md).

---

## Tech stack

| Layer | Stack |
|---|---|
| API | Go 1.25, Gin, GORM, Swagger (swag) |
| Data | PostgreSQL 15, Redis 7 |
| Messaging | In-process bus + Redis lists/streams |
| Intent ML | Python / TensorFlow → TFLite (`ml/`) |
| Fraud ML | Python / sklearn → distilled TFLite (`fraude_ml/`) |
| Frontends | React, Vite, TanStack Query, shared `@skykin/ui` |
| Mobile | Flutter SDK (HMAC-signed requests + on-device models) |

---

## Prerequisites

- Go **1.25+**
- Docker & Docker Compose
- Node.js **18+** (portals)
- Python **3.10+** (optional — ML training / export)
- `jq` (optional — `make ping` / `make register-test`)

---

## Quick start (Docker — recommended)

API listens on **http://localhost:8081**. Do not run `make run` while containers are up (port conflict).

1. Copy / edit `.env` (must include a real `JWT_SECRET` for non-dev use).
2. Start the stack:

```bash
make up
```

This brings up **Postgres** (`5435`), **Redis** (`6379`), **pgAdmin** (`5055`), and the **backend** (`8081`).

3. Verify:

```bash
make ping
make register-test
```

4. Explore the API:

- Swagger UI → http://localhost:8081/swagger/index.html  
- Health → `GET /ping`

Stop / logs:

```bash
make down
make logs
```

### Local API (optional)

Only after `make down`:

```bash
make run          # go run cmd/api/main.go on :8081
make build        # writes bin/api
make test
make swagger      # regenerates docs/
```

---

## Frontends

npm workspaces — design system documented in [`DESIGN.md`](DESIGN.md).

| App | Path | Audience |
|---|---|---|
| Developer portal | `portal/` | SDK apps & API keys |
| Ad portal | `ad-portal/` | Advertisers |
| Admin portal | `admin-portal/` | Operators |

```bash
npm install
npm run build --workspaces --if-present
# then start the app you need, e.g. from ad-portal/
```

Identity accents (wayfinding only): developer blue, advertiser sky, admin navy — same component system in `shared/`.

---

## API surfaces

Base path: `/api/v1`

| Area | Auth | Examples |
|---|---|---|
| Developer / SDK keys | Portal session / JWT | App registration, API key lifecycle |
| Ad portal | Bearer JWT | Campaigns, billing, audience |
| Admin | Bearer JWT + RBAC | Moderation, analytics, roles |
| SDK (Flutter) | `X-API-Key` + `X-Signature` (HMAC of body) | Consent, intent ingest, rewards, telemetry |

**Postman**

- Import files under `postman/`
- Collection: [`postman/Skykin-API.postman_collection.json`](postman/Skykin-API.postman_collection.json)
- Environment: [`postman/Skykin-Local.postman_environment.json`](postman/Skykin-Local.postman_environment.json)

In Postman: **File → Import** both JSON files → activate **Skykin — Local**.

---

## ML components

### Intent model (`ml/`)

Session / behavioral features → intent class probabilities. Trained as Keras, exported to TFLite for the Flutter SDK (`skykin-sdk/lib/ml/assets/`).

### Fraud / scam model (`fraude_ml/`)

Hybrid SMS + URL scam detector (TF-IDF + URL heuristics). Supports:

- `predict_message` — SMS / notification body  
- `predict_url` — browser / link inspection  

```bash
cd fraude_ml
python -m src.models.train
python -m src.export.convert_to_tflite
pytest tests/ -q
```

Frontend integration notes: [`fraude_ml/FRONTEND_MODEL_GUIDE.md`](fraude_ml/FRONTEND_MODEL_GUIDE.md).

Backend fraud persistence (blocklists, patterns, threat reports) lives under `internal/fraud/` with SQL in `internal/platform/database/migrations/fraud.sql`.

---

## Configuration & ports

| Service | Host port |
|---|---|
| API | `8081` |
| Postgres | `5435` → container `5432` |
| Redis | `6379` |
| pgAdmin | `5055` |
| Intent ML (optional host uvicorn) | `8000` (`ML_SERVICE_URL`) |

Typical env vars (see `.env` / compose): `DB_*`, `DATABASE_URL`, `REDIS_ADDR`, `JWT_SECRET`, `ML_SERVICE_URL`, `PORT`.

---

## Useful Make targets

| Target | Purpose |
|---|---|
| `make up` | Build & start Docker stack (detached) |
| `make down` | Stop stack |
| `make logs` | Tail backend logs |
| `make ping` | `GET /ping` |
| `make register-test` | Smoke ad-portal register |
| `make run` | Local Go API (conflicts with Docker) |
| `make test` | `go test ./...` |
| `make swagger` | Regenerate OpenAPI under `docs/` |
| `make clean` | Remove `bin/` |

---

## Further reading

| Doc | Contents |
|---|---|
| [`docs/MODULE_ARCHITECTURE_EXPLANATION.md`](docs/MODULE_ARCHITECTURE_EXPLANATION.md) | Modules, bus, Redis, workers, migrations |
| [`DESIGN.md`](DESIGN.md) | Shared frontend design system (“Signal”) |
| [`fraude_ml/README.md`](fraude_ml/README.md) | Scam-detector training & inference |
| [`fraude_ml/FRONTEND_MODEL_GUIDE.md`](fraude_ml/FRONTEND_MODEL_GUIDE.md) | On-device TFLite usage for Flutter |

---

## Status notes

Internal platform monorepo. Some domains (`events`, `geofencing`) exist in code but are not fully mounted in the HTTP router yet — prefer the architecture doc and Swagger for the live surface.
