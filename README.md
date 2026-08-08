# Skykin Platform

Skykin is an **intent-aware advertising and SDK platform**. Mobile apps integrate a Flutter SDK that scores on-device behavior into intents; the backend matches eligible campaigns and delivers ads (in-app creatives, SMS+, and location-based store visits). Developers, advertisers, and operators manage the system through three React portals and a Go API.

---

## What it does

| Audience | Surface | Job |
|----------|---------|-----|
| End users (via Flutter SDK) | Consent, intent ingest, rewards, fraud reports, geofencing | Privacy-preserving identity + ad / SMS / store-visit delivery |
| Developers | `portal/` + `/api/v1/portal` | Register apps, issue API keys (`pk_live_` / `sk_secret_`) |
| Advertisers | `ad-portal/` + `/api/v1/ad-portal` | Plans, campaigns, audiences, geofence stores, billing |
| Operators | `admin-portal/` + `/api/v1/ad-portal/admin` | Moderation, analytics, catalog, RBAC, geofence activation |

### Core runtime flow

```
Flutter SDK
    │  X-API-Key + HMAC-SHA256(X-Signature of body)
    ▼
Gin API (:8081)
    │
    ├─ Consent          → provision / map pseudonymous user
    ├─ Intent ingest    → ML / rules → eligible campaigns
    ├─ Ad select        → in-app creative  OR  SMS+ dispatch
    ├─ Geofencing       → sync active zones → enter event → creative
    └─ Telemetry / rewards / threat reports
         │
         ├── Postgres + PostGIS (source of truth)
         ├── Redis (cache, queues, streams, freq caps)
         └── In-process event bus (sagas & side effects)
```

---

## Repository layout

```
SDK/
├── cmd/api/                      # Go API entrypoint + Swagger annotations
├── configs/                      # Env / config / logging
├── deployments/docker/           # api.Dockerfile, ml.Dockerfile
├── docs/                         # OpenAPI (swagger) + architecture notes
├── docker-compose.yml            # PostGIS, Redis, Adminer, backend
├── internal/                     # Backend bounded contexts (hexagonal)
│   ├── platform/                 # Bootstrap, router, DB, Redis, messaging, middleware
│   ├── auth/                     # Developers, applications, API keys
│   ├── consent/                  # SDK consent + pseudonymous mappings
│   ├── users/                    # User provisioning
│   ├── intents/                  # Intent ingest + ad fetch
│   ├── campaigns/                # Campaigns, eligibility, moderation, ad selection
│   ├── delivery/                 # SMS+ dispatch, tracking, telemetry, debug
│   ├── billing/                  # Plans, rates, subscriptions, invoices
│   ├── audience/                 # Segments & purchases
│   ├── analytics/                # Aggregates / operator reporting
│   ├── rewards/                  # Reward rules & grants
│   ├── ad_portal/ · admin/       # Advertiser & operator HTTP surfaces
│   ├── permissions/              # RBAC (roles, permissions, Redis cache)
│   ├── fraud/                    # Blocklists, patterns, threat reports
│   ├── geofencing/               # Store zones, campaign targets, SDK sync/events
│   └── events/                   # Retained (HTTP not mounted)
├── ml/                           # Intent classification (Keras → TFLite / optional uvicorn)
├── fraude_ml/                    # SMS / URL scam detector (sklearn → TFLite)
├── portal/ · ad-portal/ · admin-portal/   # React + Vite portals
├── shared/                       # Shared UI package (@skykin/ui)
├── skykin-sdk/                   # On-device ML assets for Flutter
├── postman/                      # API collection + local environment
├── scripts/ · tests/             # Helpers and tests
└── Makefile
```

Backend modules follow **ports & adapters**:

```
internal/<module>/
  domain/            # entities, repository ports
  application/       # use cases (depend on interfaces only)
  infrastructure/    # Postgres / Redis / provider adapters
  interfaces/http/   # Gin handlers + DTOs
  consumers/         # event-bus handlers (where needed)
  worker/            # Redis stream / queue workers (where needed)
  routes/            # Wire() + Register() for HTTP
```

Cross-module wiring lives only in `internal/platform/bootstrap/` and `internal/platform/route/`.  
Details: [`docs/MODULE_ARCHITECTURE_EXPLANATION.md`](docs/MODULE_ARCHITECTURE_EXPLANATION.md).

---

## Tech stack

| Layer | Stack |
|-------|-------|
| API | Go 1.25, Gin, GORM, Swagger (`swag`) |
| Data | PostgreSQL 15 + **PostGIS**, Redis 7 |
| Messaging | In-process `messaging.Bus` + Redis lists/streams |
| Intent ML | Python / TensorFlow → TFLite (`ml/`) |
| Fraud ML | Python / sklearn → distilled TFLite (`fraude_ml/`) |
| Frontends | React, Vite, TanStack Query, shared `@skykin/ui` |
| Mobile | Flutter SDK (HMAC-signed requests + on-device models) |
| Ops UI | Adminer on host port **8082** |

---

## Backend modules (what each owns)

| Module | Responsibility |
|--------|----------------|
| **auth** | Developer register/login, applications, publishable + secret keys |
| **consent** | SDK consent levels, SMS consent, pseudonymous ID mapping |
| **users** | SDK user rows used by consent / intents / admin listings |
| **intents** | `ingest-ad` / `ingest-aggregate`, intent cache, campaign match |
| **campaigns** | CRUD, creative validation, eligibility, moderation hooks |
| **delivery** | Anonymous campaigns, telemetry, SMS+ mock/Twilio, click redirects |
| **billing** | Subscription plans, channels, rates, advertiser subscribe |
| **audience** | Segment catalog, purchases, candidates for operator review |
| **analytics** | Operator overview / revenue / delivery / intent-consistency |
| **admin** | Operator campaign validate/activate, plans, segments, SDK users |
| **ad_portal** | Advertiser auth + mounts billing/campaigns/audience/geofences |
| **permissions** | RBAC permissions & roles (`geofences:manage`, etc.) |
| **fraud** | Intelligence seed + `POST /reports` threat ingestion |
| **geofencing** | Advertiser stores, campaign links, admin activate, SDK sync/event |
| **rewards** | Rules and grants tied to consented behavior |
| **platform** | DB migrate/seed, Redis, middleware, composition root |

---

## API surfaces

Base path: **`/api/v1`**  
Swagger UI: **http://localhost:8081/swagger/index.html**

### Auth models

| Surface | Auth |
|---------|------|
| Developer portal | Bearer JWT (developer session) |
| Ad portal / admin | Bearer JWT (`advertiser` or `operator_admin`) |
| SDK routes | `X-API-Key` (`pk_live_…`) + `X-Signature` = lowercase hex **HMAC-SHA256(secret_key, raw body)** on POST/PATCH |

### Notable route groups

**Developer portal**

- `POST /portal/register`, `/portal/login`
- `POST|GET /portal/applications` → `publishable_key` + `secret_key`

**Advertiser (ad portal)**

- Auth: `/ad-portal/register`, `/login`, `/me`
- Billing: `/plans`, `/subscription`, `/channels`
- Campaigns: CRUD, preview, activate (only after operator approval)
- Audience: list/purchase segments
- Geofences: create/list stores, link/list campaign targets

**Operator admin** (`/ad-portal/admin`, role `operator_admin`)

- Campaigns: pending list, validate (approve/reject), activate
- Geofences: pending inactive zones, activate zone, activate campaign zones
- Plans, billing rates, audience segments & candidates
- Analytics, SDK users, RBAC permissions/roles

**SDK (Flutter)**

- `POST /consent`
- `POST /intents/ingest-ad`, `/intents/ingest-aggregate`
- `GET /campaigns/anonymous`, telemetry + SMS click/debug
- `PATCH /geofences/location-consent`, `GET /geofences/sync`, `POST /geofence/event`
- `POST /reports` (fraud threat reports), `/sync`

**Postman:** import files under [`postman/`](postman/).

---

## Important product flows

### 1. Intent → ad (or SMS+)

1. App registers → keys issued.  
2. User consents → `pseudonymous_id`.  
3. SDK posts intent (`ingest-ad`).  
4. Backend selects an eligible **active + approved** campaign.  
5. Returns in-app creative, or queues **SMS+** when channel and consent allow.

### 2. Advertiser campaign lifecycle

1. Advertiser registers, **subscribes to a plan**, then creates a campaign (`moderation_status=pending`).  
2. Operator lists pending campaigns and **approves or rejects**.  
3. Approve sets `validation_status=passed`, `moderation_status=approved`, and can set `is_active=true`.  
4. Delivery / geofence enter only use approved active campaigns.

### 3. Geofencing (store visits)

1. **Advertiser** creates store zones (`is_active=false` drafts) and links them to a campaign.  
2. **Admin** approves the campaign and/or activates zones (`is_active=true` if not already).  
3. SDK syncs nearby **active** zones via PostGIS `ST_DWithin`.  
4. On enter: consent check → `store_visits` row → eligible campaign creative.

Step-by-step Swagger guide: [`docs/GEOFENCING_SWAGGER_TEST.md`](docs/GEOFENCING_SWAGGER_TEST.md).

### 4. Fraud

- On-device / client reports → `POST /reports` with aggregation in Redis.  
- Backend catalog: domains, senders, patterns (seeded).  
- Separate Python package `fraude_ml/` trains/export TFLite for SMS/URL scam detection on device.

---

## Prerequisites

- Go **1.25+**
- Docker & Docker Compose
- Node.js **18+** (portals)
- Python **3.10+** (optional — ML training / export)
- `jq` (optional — `make ping` / `make register-test`)

---

## Quick start (Docker — recommended)

API: **http://localhost:8081**. Do not run `make run` while containers are up (port conflict).

1. Ensure `.env` exists (compose loads it). Use a real `JWT_SECRET` outside local dev.
2. Start the stack:

```bash
make up
```

Brings up:

| Service | Host port |
|---------|-----------|
| Backend API | `8081` |
| Postgres + PostGIS | `5435` → `5432` |
| Redis | `6379` |
| Adminer (DB UI) | `8082` |

3. Verify:

```bash
make ping
make register-test
```

4. Explore:

- Swagger → http://localhost:8081/swagger/index.html  
- Health → `GET /ping`  
- Adminer → http://localhost:8082  
  - System: **PostgreSQL**, Server: **`db`**, User: `skykin_user`, Password: `password`, Database: `skykin_db`

Stop / logs:

```bash
make down
make logs
```

### Seeded operator admin

| Field | Value |
|-------|--------|
| Email | `admin@skykin.com` |
| Password | `Admin12345!` |
| Role | `operator_admin` |

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

npm workspaces — design system in [`DESIGN.md`](DESIGN.md).

| App | Path | Audience |
|-----|------|----------|
| Developer portal | `portal/` | SDK apps & API keys |
| Ad portal | `ad-portal/` | Advertisers |
| Admin portal | `admin-portal/` | Operators |

```bash
npm install
npm run build --workspaces --if-present
# then start the app you need from its folder (Vite)
```

Identity accents (wayfinding): developer blue, advertiser sky, admin navy — shared components in `shared/`.

---

## ML components

### Intent model (`ml/`)

Behavioral / session features → intent class probabilities. Trained as Keras, exported to TFLite for Flutter (`skykin-sdk/`). Optional host inference:

```bash
cd ml && uvicorn app:app --port 8000
# backend: ML_SERVICE_URL=http://host.docker.internal:8000
```

### Fraud / scam model (`fraude_ml/`)

Hybrid SMS + URL scam detector. Supports `predict_message` and `predict_url`.

```bash
cd fraude_ml
python -m src.models.train
python -m src.export.convert_to_tflite
pytest tests/ -q
```

See [`fraude_ml/README.md`](fraude_ml/README.md) and [`fraude_ml/FRONTEND_MODEL_GUIDE.md`](fraude_ml/FRONTEND_MODEL_GUIDE.md).

---

## Configuration

Typical env vars (`.env` / compose):

| Variable | Purpose |
|----------|---------|
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | Postgres |
| `DATABASE_URL` | Full DSN (used inside Docker) |
| `REDIS_ADDR` | e.g. `redis:6379` |
| `JWT_SECRET` | Ad-portal / portal JWT signing |
| `ML_SERVICE_URL` | Optional intent HTTP service |
| `PORT` | API listen port (default `8081`) |
| `GIN_MODE` | `release` in Docker |

Migrations and seeds run on API boot (`internal/platform/database/`), including PostGIS geofencing SQL, billing catalog, audience segments, fraud intelligence, demo SMS recipients, and RBAC permissions.

---

## Make targets

| Target | Purpose |
|--------|---------|
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

## Architecture principles (short)

1. **No cross-repo table access** — each module’s repositories touch their own tables.  
2. **Composition root only** — `bootstrap` + `route` wire adapters across modules.  
3. **Thin HTTP** — bind → one use case → map errors.  
4. **Async via bus / Redis** — billing events, delivery logs, analytics aggregates, intent logs.  
5. **Privacy** — SDK users are addressed by `pseudonymous_id`, not raw device identity in ad paths.

Full explanation: [`docs/MODULE_ARCHITECTURE_EXPLANATION.md`](docs/MODULE_ARCHITECTURE_EXPLANATION.md).

---

## Further reading

| Doc | Contents |
|-----|----------|
| [`docs/MODULE_ARCHITECTURE_EXPLANATION.md`](docs/MODULE_ARCHITECTURE_EXPLANATION.md) | Modules, bus, Redis, workers, migrations |
| [`docs/GEOFENCING_SWAGGER_TEST.md`](docs/GEOFENCING_SWAGGER_TEST.md) | End-to-end geofencing Swagger walkthrough |
| [`DESIGN.md`](DESIGN.md) | Shared frontend design system (“Signal”) |
| [`fraude_ml/README.md`](fraude_ml/README.md) | Scam-detector training & inference |
| [`fraude_ml/FRONTEND_MODEL_GUIDE.md`](fraude_ml/FRONTEND_MODEL_GUIDE.md) | On-device TFLite usage for Flutter |
| `docs/swagger.yaml` / Swagger UI | Live HTTP contract |

---

## License / status

Internal platform monorepo. Prefer **Swagger** and the architecture doc as the source of truth for mounted routes; the `events` package exists in the tree but its HTTP surface is intentionally not mounted.
