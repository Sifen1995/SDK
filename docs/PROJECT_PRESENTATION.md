# Skykin Platform — Project Presentation

> Use this file to build a PowerPoint / Google Slides deck.  
> Each `---` separates a **slide**. Titles are H2; keep bullets short when pasting into PPT.  
> Compatible with [Marp](https://marp.app/) (`---` slide breaks) and manual copy into slides.

---

## Slide 1 — Title

# Skykin Platform

**Intent-aware advertising & SDK infrastructure**

Privacy-preserving ads · On-device ML · Campaign ops · Geofencing · Fraud intel

---

## Slide 2 — Agenda

1. Problem & vision  
2. Who uses Skykin  
3. Product overview  
4. System architecture  
5. Backend modules  
6. Key product flows  
7. Security & privacy  
8. ML components  
9. Frontends & portals  
10. Infrastructure & demo  
11. Roadmap / status  

---

## Slide 3 — The problem

- Brands want to reach the **right user at the right moment**
- Traditional ads ignore real **intent** and context
- Apps need ads without exposing raw user identity
- Operators need control: budgets, creative review, fraud, locations
- Mobile SDKs must stay **lightweight, signed, and privacy-safe**

---

## Slide 4 — The vision

**Skykin** connects:

| Layer | Role |
|-------|------|
| Flutter SDK | Sense behavior → intents / fraud / location (on device where possible) |
| Platform API | Match campaigns, deliver creatives / SMS / store offers |
| Portals | Developers, advertisers, and operators run the business |

**Outcome:** Intent-driven, consented, measurable advertising.

---

## Slide 5 — Who uses Skykin?

| Persona | Tool | Goal |
|---------|------|------|
| **End user** | Host app + Flutter SDK | Relevant offers; privacy preserved |
| **App developer** | Developer portal | Register apps, get API keys |
| **Advertiser** | Ad portal | Buy plans, create campaigns, attach stores |
| **Operator (admin)** | Admin portal | Approve campaigns/zones, catalog, analytics, RBAC |

---

## Slide 6 — Product pillars

1. **Intent advertising** — match campaigns to predicted intents  
2. **Multi-channel delivery** — in-app banner / native / push / SMS+  
3. **Audience & billing** — segments, subscriptions, budgets  
4. **Geofencing** — store zones → enter event → local offer  
5. **Fraud intelligence** — threat reports + on-device scam models  
6. **Governance** — moderation, RBAC, analytics  

---

## Slide 7 — High-level architecture

```text
  Flutter SDK                    React portals
       │                              │
       │  X-API-Key + HMAC            │  Bearer JWT
       ▼                              ▼
              ┌─────────────────────────┐
              │     Gin API (:8081)     │
              │   /api/v1/...           │
              └───────────┬─────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        ▼                 ▼                 ▼
   PostgreSQL+PostGIS   Redis          In-process bus
   (source of truth)  cache/queues     sagas / events
```

---

## Slide 8 — Tech stack

| Area | Technology |
|------|------------|
| API | Go 1.25, Gin, GORM, Swagger |
| Database | PostgreSQL 15 + **PostGIS** |
| Cache / async | Redis 7 (lists, streams, freq caps) |
| Messaging | Internal event bus |
| Intent ML | Python / TensorFlow → TFLite |
| Fraud ML | sklearn → TFLite (`fraude_ml/`) |
| Frontends | React, Vite, TanStack Query |
| Mobile | Flutter SDK + on-device models |
| Ops | Docker Compose, Adminer |

---

## Slide 9 — Repository map

```text
cmd/api          → API entry
internal/*       → Bounded contexts (hexagonal)
portal/          → Developer UI
ad-portal/       → Advertiser UI
admin-portal/    → Operator UI
shared/          → Design system
ml/              → Intent training & serve
fraude_ml/       → Scam detector
skykin-sdk/      → On-device ML assets
docs/            → Architecture + OpenAPI
```

---

## Slide 10 — Backend design pattern

**Hexagonal / ports & adapters**

```text
HTTP handler  →  Application use case  →  Port (interface)
                                              │
                         ┌────────────────────┼────────────────┐
                         ▼                    ▼                ▼
                      Postgres              Redis           External
```

- Modules **do not** import each other’s infrastructure  
- Wiring only in `platform/bootstrap` + `route`  
- Thin handlers; business logic in application layer  

---

## Slide 11 — Backend modules (1/2)

| Module | Owns |
|--------|------|
| **auth** | Developers, apps, API keys |
| **consent** | Consent + pseudonymous mapping |
| **users** | SDK user records |
| **intents** | Ingest intent → select ad |
| **campaigns** | Campaigns, eligibility, moderation |
| **delivery** | Telemetry, SMS+, click tracking |
| **billing** | Plans, channels, subscriptions |

---

## Slide 12 — Backend modules (2/2)

| Module | Owns |
|--------|------|
| **audience** | Segments & purchases |
| **analytics** | Operator reports |
| **admin / ad_portal** | Portal HTTP surfaces |
| **permissions** | RBAC roles & permissions |
| **fraud** | Blocklists, reports, intel |
| **geofencing** | Zones, targets, visits, SDK sync |
| **rewards** | Rules & grants |
| **platform** | DB, Redis, router, workers |

---

## Slide 13 — API surfaces

Base: **`/api/v1`** · Swagger: `:8081/swagger`

| Surface | Auth | Examples |
|---------|------|----------|
| `/portal` | Developer JWT | Apps & keys |
| `/ad-portal` | Advertiser JWT | Campaigns, billing, geofences |
| `/ad-portal/admin` | Operator JWT | Approve, catalog, analytics, RBAC |
| SDK routes | API key + HMAC | Consent, intents, geofence, reports |

---

## Slide 14 — Flow: Intent → Ad

```text
App user behavior
      ↓
On-device / backend intent
      ↓
POST /intents/ingest-ad
      ↓
Eligible campaigns (active + approved + budget + freq cap)
      ↓
┌─────────────┴─────────────┐
│ In-app creative (JSON)    │  SMS+ (if channel + SMS consent)
└───────────────────────────┘
```

---

## Slide 15 — Flow: Campaign lifecycle

```text
Advertiser
  1. Subscribe to plan
  2. Create campaign  →  moderation = pending
  3. (Optional) attach geofence stores

Operator
  4. Review creative / budget / zones
  5. Approve  →  validation passed + can go live
     or Reject → stays offline

Delivery only uses approved + active campaigns
```

---

## Slide 16 — Flow: Geofencing

```text
Advertiser creates store (lat, lng, radius)
        ↓  is_active = false (draft)
Links store to campaign
        ↓
Admin approves campaign and/or activates zones
        ↓  is_active = true
SDK GET /geofences/sync  (PostGIS nearby)
        ↓
SDK POST /geofence/event  (enter)
        ↓
store_visits + ad creative (if consented)
```

---

## Slide 17 — Geofencing roles

| Role | Can do | Cannot do |
|------|--------|-----------|
| **Advertiser** | Create zones, link to campaigns | Activate zones / approve campaigns |
| **Admin** | Approve campaign, activate zones | Create advertiser stores |
| **SDK** | Sync active zones, report enter | See inactive drafts |

---

## Slide 18 — Flow: Fraud

**Platform**

- Client posts threat reports → `POST /reports`
- Redis aggregates → promotes signals
- Seeded blocklists: domains, senders, patterns

**On-device (`fraude_ml/`)**

- SMS / URL scam scoring  
- TFLite for Flutter  
- Complements server-side intelligence  

---

## Slide 19 — Privacy & identity

- End users addressed by **`pseudonymous_id`**, not raw PII in ad paths  
- Consent gates: general ads, SMS, **location ad consent** (demo cohort)  
- HMAC request signing prevents forged SDK traffic  
- Operator admin is seeded separately from advertisers  

**Principle:** Deliver relevance without leaking identity.

---

## Slide 20 — Security model

| Threat | Control |
|--------|---------|
| Spoofed SDK calls | `X-API-Key` + HMAC body signature |
| Unauthorized portal use | JWT + role checks |
| Privilege escalation | RBAC (`permissions` module) |
| Self-activating ads | Campaigns require operator approval |
| Live geofences without review | Zones start inactive |

---

## Slide 21 — ML: Intent

**Package:** `ml/`

- Input: behavioral / session features  
- Output: intent class probabilities (e.g. fashion, food, crypto)  
- Train: Keras → export TFLite for Flutter  
- Optional: host uvicorn service (`ML_SERVICE_URL`)  

Feeds campaign targeting (`target_intent`).

---

## Slide 22 — ML: Fraud / scam

**Package:** `fraude_ml/`

- Hybrid SMS + URL heuristics  
- APIs: `predict_message`, `predict_url`  
- Export TFLite for on-device use  
- Docs: `FRONTEND_MODEL_GUIDE.md`  

Protects users; informs platform threat intel.

---

## Slide 23 — Frontends

| App | Audience | Focus |
|-----|----------|-------|
| **portal** | Developers | Apps, API keys |
| **ad-portal** | Advertisers | Plans, campaigns, stores |
| **admin-portal** | Operators | Moderation, analytics, RBAC |
| **shared** | All | Design system (`@skykin/ui`) |

Stack: React + Vite + TanStack Query  
Visual identity: “Signal” design system (`DESIGN.md`)

---

## Slide 24 — Delivery channels

| Code | Purpose |
|------|---------|
| `IN_APP_BANNER` | Classic banner creative |
| `NATIVE_FEED` | Feed-style placement |
| `PUSH` | Push notification creative |
| `SMS_PLUS` | SMS with tracked click → destination |

Channel chosen per campaign; creative validated by channel rules.

---

## Slide 25 — Billing & audience

**Billing**

- Plans (Starter / Growth / Enterprise)  
- Advertiser must subscribe before creating campaigns  
- Daily / total budget caps, impression quotas  

**Audience**

- Catalog segments (intent signals, CPM)  
- Advertisers purchase segments for targeting  
- Operators approve segment candidates  

---

## Slide 26 — Async & workers

Background work (Redis + bus):

- Billing event stream → invoices / metering  
- Delivery log stream → campaign delivery logs  
- Analytics aggregate queue  
- Intent log worker  
- Targeting job (periodic)  
- Intent-consistency analysis (scheduled)  

Keeps HTTP handlers fast; write-behind to Postgres.

---

## Slide 27 — Local infrastructure

| Service | Port | Role |
|---------|------|------|
| Backend API | **8081** | Gin + Swagger |
| Postgres + PostGIS | **5435** | Data + geography |
| Redis | **6379** | Cache / queues |
| Adminer | **8082** | DB browser |

```bash
make up      # start stack
make ping    # health check
make logs    # backend logs
```

---

## Slide 28 — Demo credentials

**Operator admin (seeded)**

- Email: `admin@skykin.com`  
- Password: `Admin12345!`  

**Typical demo path**

1. Developer: register app → keys  
2. Advertiser: register → subscribe → campaign + zone  
3. Admin: approve campaign (+ zones)  
4. SDK: consent → sync → geofence event → ad  

---

## Slide 29 — Developer experience

- **Swagger UI** — live contract at `/swagger`  
- **Postman** collection under `postman/`  
- **Makefile** — up / down / test / swagger  
- **Architecture doc** — `docs/MODULE_ARCHITECTURE_EXPLANATION.md`  
- **Geofencing guide** — `docs/GEOFENCING_SWAGGER_TEST.md`  
- **README** — full project reference  

---

## Slide 30 — What makes Skykin different

1. **Intent-first** matching, not only demographics  
2. **Privacy-by-design** with pseudonymous IDs  
3. **Operator governance** on campaigns and store locations  
4. **Multi-channel** including SMS+ with tracking  
5. **Location + intent** via PostGIS geofencing  
6. **On-device ML** for intents and scam detection  
7. **Clean modular backend** ready to grow  

---

## Slide 31 — Current status

**Live today**

- Full ad portal + admin APIs  
- Intent ingest & ad selection  
- Billing / audience / analytics  
- Geofencing (create → approve → sync → event)  
- Fraud report ingestion  
- Three React portals + Docker stack  

**Retained / not mounted**

- `events` HTTP surface (package kept for later)  

---

## Slide 32 — Summary

**Skykin** is a full-stack **intent advertising platform**:

- SDK senses → API decides → portals control  
- Secure, modular Go backend  
- Postgres + PostGIS + Redis  
- Intent & fraud ML on and off device  
- Ready for demos: campaigns, SMS+, geofencing, admin ops  

---

## Slide 33 — Thank you / Q&A

# Thank you

**Questions?**

Useful links (local):

- API: `http://localhost:8081`  
- Swagger: `http://localhost:8081/swagger/index.html`  
- Adminer: `http://localhost:8082`  
- Docs: `README.md`, `docs/`  

---

## Appendix A — One-slide architecture (deep)

```text
SDK ──HMAC──► Gin
               ├ consent / users
               ├ intents ──► campaigns selector
               ├ delivery (SMS / telemetry)
               ├ geofencing (PostGIS)
               ├ fraud reports
               └ rewards
Portals ─JWT─► ad_portal / admin / auth
               ├ billing · audience · analytics
               └ permissions (RBAC)

Workers: billing stream · delivery logs · analytics · targeting
```

---

## Appendix B — Geofencing tables

| Table | Purpose |
|-------|---------|
| `geofence_zones` | Store lat/lng/radius + `is_active` |
| `campaign_geofence_targets` | M:N campaign ↔ zone |
| `store_visits` | Enter events from SDK |
| `demo_sms_recipients.location_ad_consent` | Demo location consent |

---

## Appendix C — How to turn this into PPT

**Option 1 — Marp (Markdown → slides)**

```bash
# VS Code: Marp extension, or CLI:
npx @marp-team/marp-cli docs/PROJECT_PRESENTATION.md -o docs/Skykin_Presentation.pptx
```

**Option 2 — Manual**

- Each `## Slide N` = one slide title  
- Bullets → slide bullets  
- Code fences → diagrams / architecture slides  

**Option 3 — Pandoc**

```bash
pandoc docs/PROJECT_PRESENTATION.md -o docs/Skykin_Presentation.pptx
```

Speaker tip: keep slides 3–16 for a 10–15 minute pitch; use appendices for deep dive.
