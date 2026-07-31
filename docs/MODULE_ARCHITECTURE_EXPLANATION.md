# Skykin Platform — Full Architecture Explanation

This document explains how the backend is structured: every module, the hexagonal (ports/adapters) pattern, the in-process event bus, Redis usage, background workers, composition root wiring, and how database migrations should be organized.

**Entry point:** `cmd/api/main.go`  
**Module path:** `skykin-platform`  
**Composition root:** `internal/platform/bootstrap/` + `internal/platform/route/router.go`

---

## 1. High-level picture

```
Flutter SDK / Portals
        │
        ▼
   Gin HTTP (cmd/api)
        │
        ├── SDK auth middleware (X-API-Key + signature)
        ├── Ad-portal JWT (Bearer)
        └── Handlers (thin) → Application services → Ports
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
              Postgres           Redis            Messaging Bus
           (source of truth)  (cache/queues/     (in-process
                               streams)           pub/sub)
                    ▲                 ▲                 │
                    │                 │                 ▼
               Workers ◄──────────────┘           Consumers
          (BRPop / XReadGroup)              (saga + CampaignMatched)
```

**Architectural style**

| Concern | Approach |
|---------|----------|
| Bounded contexts | Folders under `internal/` (consent, users, intents, campaigns, delivery, billing, …) |
| Layering | `domain` → `application` → `infrastructure` / `interfaces` |
| Cross-module calls | Prefer **ports** at the application boundary; wire real adapters in **bootstrap** only |
| Async side effects | In-process `messaging.Bus` and Redis streams/lists |
| Persistence | Postgres via GORM repositories; some tables owned by explicit SQL migrations |

Modules must not import each other’s infrastructure packages. The composition root (`bootstrap`) is allowed to import multiple modules and glue them together.

---

## 2. Layering (ports & adapters)

A mature module typically looks like:

```
internal/<module>/
  domain/           # entities, value objects, repository ports, event names
  application/      # use cases / services (depend on interfaces, not GORM)
  infrastructure/   # Postgres/Redis/Twilio adapters implementing ports
  interfaces/http/  # Gin handlers + DTOs (or http/)
  consumers/        # Bus event handlers
  worker/           # Redis stream/list consumers (long-running)
  routes/           # Wire() mounts HTTP under a Gin group
```

**Adapter examples**

| Port (owned by) | Adapter (wired in bootstrap) |
|-----------------|------------------------------|
| `intents/application.AdSelector` | `campaigns/application.IntentAdSelector` |
| `intents/application.SMSAdDispatcher` | thin wrapper over `delivery/application.SMSDispatchService` |
| `intents/application.ActiveIntentCache` | Redis `user_intent:{id}` |
| `admin` `SegmentPublisher` | audience `ProcessApprovedCandidateUseCase` |
| `delivery` `SMSProvider` | `MockSMSProvider` or `TwilioSMSProvider` |
| Campaign eligibility | `CachedCampaignRepository` (Redis + Postgres) |

Handlers stay thin: bind JSON → call one use case → map errors to HTTP status.

---

## 3. Process bootstrap & composition root

### `cmd/api/main.go`

1. Load config (`configs.LoadConfig`)
2. Connect Postgres (`database.ConnectDB`)
3. Run `database.Migrate` (SQL embeds + AutoMigrate + seeds)
4. Seed operator admin
5. Create `messaging.NewBus()`
6. Optional Redis client
7. `bootstrap.NewPermissionSystem`
8. `route.InitRouter` (HTTP + many consumers/workers)
9. Start more workers/jobs: intent log, analytics aggregate, billing stream, delivery log stream, targeting job (5m), intent-consistency jobs (24h)

### `route.InitRouter`

| Step | What |
|------|------|
| Middleware | CORS, logger, recovery |
| `auth.RegisterRoutes` | Developer portal + SDK auth middleware |
| `ad_portal.Register` | Billing, audience, campaigns, admin, analytics + admin bus consumers |
| `NewConsentSystem` | Consent HTTP + consent/users saga consumers |
| `NewDeliverySDKSystem` | Anonymous campaigns, telemetry, SMS click/dispatch/debug |
| `NewIntentSystem(..., SMSDispatch)` | Ingest-ad / ingest-aggregate |
| `RegisterDeliveryEventConsumers` | `SMSPlusConsumer` on `CampaignMatched` |
| Start stream/queue workers | Billing, delivery logs, analytics aggregate, intent logs |
| Mount `/api/v1` SDK routes | consent, intents, delivery |

**Note:** Stream/queue workers are started in **both** `InitRouter` and `main.go`. That means duplicate consumers if Redis is up. Prefer starting them in one place only.

### Call graph

```
main
 ├─ database.Migrate
 ├─ messaging.NewBus
 ├─ NewPermissionSystem
 └─ InitRouter
      ├─ auth.RegisterRoutes
      ├─ ad_portal.Register
      │    ├─ billing / audience / campaigns / admin / analytics Wire
      │    ├─ RegisterAdminEventConsumers
      │    └─ SetupIntentConsistency
      ├─ NewConsentSystem
      ├─ NewDeliverySDKSystem
      ├─ NewIntentSystem
      ├─ RegisterDeliveryEventConsumers
      └─ Start*Worker (streams/queues)
 └─ StartTargetingJob / StartIntentConsistencyJobs / (duplicate Start*Workers)
```

---

## 4. Event-driven architecture (in-process bus)

### Bus implementation — `internal/platform/messaging/`

| API | Behavior |
|-----|----------|
| `Subscribe` / `Register(bus, name, handler)` | Register handlers per event name |
| `Publish` | Each handler runs in a **new goroutine** (fire-and-forget) |
| `PublishSync` | Handlers run on the **caller goroutine**, in registration order |

`PublishSync` is used for **HTTP sagas** that must finish before the response (consent registration).  
`Publish` is used for **side effects** that can lag (moderation notifications, SMS after targeting).

This is **not** Kafka/NATS. It is an in-memory bus inside one API process. If you scale to multiple replicas, bus events do not cross machines — Redis streams/lists already do for telemetry write-behind.

### Active event catalog

| Event name | Published by | Consumed by | Sync? |
|------------|--------------|-------------|-------|
| `ConsentRegistrationRequested` | `CreateConsentUseCase` | users `ConsentRegistrationConsumer` | Sync |
| `UserProvisionedForConsent` | `ProvisionUserForConsentUseCase` | consent `UserProvisionedConsumer` | Sync |
| `ConsentCreated` | `CompleteConsentRegistrationUseCase` | *(none today)* | Async |
| `admin.subscription_plan_created` | billing `PlanService` | billing `PlanConsumer` (seed rates) | Async |
| `admin.campaign_moderation_passed` / `_rejected` / `_activated` | campaigns `ModerationService` | campaigns `ModerationConsumer` (logging today) | Async |
| `CampaignMatched` | campaigns `TargetingJob` | delivery `SMSPlusConsumer` | Async |

**Defined but mostly unused / not wired:** permissions role events, `events.event_*`, rewards `evaluation.requested`, analytics finding topic (findings are processed via direct adapter, not bus).

### Consent saga (canonical bus example)

```
POST /consent (sms_consented=false)
  → CreateConsentUseCase
  → PublishSync(ConsentRegistrationRequested)
       → users: create users row
       → PublishSync(UserProvisionedForConsent)
            → consent: create pseudonymous_mappings + consents
            → Publish(ConsentCreated)  // async, no consumer yet
  → HTTP 201 { pseudonymous_id, ... }
```

When `sms_consented=true`, the saga is **skipped**: the API returns an existing demo SMS recipient’s `pseudonymous_id` (no new user).

### CampaignMatched → SMS (async bus example)

```
TargetingJob (every 5 minutes)
  → find eligible campaign + segment members with matching intent
  → Publish(CampaignMatched{campaign_id, pseudonymous_id, channel})
       → SMSPlusConsumer (if channel is SMS_PLUS)
            → SMSDispatchService.DispatchCampaignMatch
```

Ingest-ad can also dispatch SMS **directly** (port call, not bus) when the selected channel is `SMS_PLUS`.

---

## 5. Redis — how it is used

Client wrapper: `internal/platform/redis/redis.go`  
(`Set`, `SetNX`, `Get`, `Incr`, `IncrByFloat`, `RPush`, `BRPop`, `LPush`, `LTrim`, `Expire`, `XAdd`, `XReadGroup`, `XAck`, …)

### Caches (cache-aside)

| Key | Module | Purpose | TTL |
|-----|--------|---------|-----|
| `user_intent:{pseudonymous_id}` | intents | Latest active intent for a user | ~30m |
| `eligible_campaigns:{intent}:{channel\|all}` | campaigns | Eligible campaign list for ingest-ad | 5m (**empty lists are not cached**) |
| `cache:active_campaigns_master` | campaigns | Anonymous campaign master list | 5m |
| `budget_exhausted:{campaign_id}` | campaigns / billing | Skip campaigns that hit budget | until cleared |
| `freq:{pseudonymous_id}:{campaign_id}` | campaigns | Server-side frequency cap counter | 24h on first incr |
| `perm_cache:{roleName}` | permissions | RBAC permission set | cache layer |
| `budget:spent:{campaign_id}:{YYYY-MM-DD}` | billing tracker | Daily spend accumulator | day-scoped |

### Lists (queues)

| Key | Producer | Consumer worker | Writes to |
|-----|----------|-----------------|-----------|
| `queue:intent_logs` | Intent profile save path | `StartIntentLogWorker` (BRPop) | `intents` table batch insert |
| `queue:analytics_aggregate` | `POST /intents/ingest-aggregate` | `StartAnalyticsAggregateWorker` | upsert `intent_aggregate_counts` |

### Streams (write-behind)

| Stream | Producers | Consumer groups | Postgres |
|--------|-----------|-----------------|----------|
| `stream:billing_events` | Telemetry track, CPC click, SMS click | `billing_processor_group` → billing worker; `delivery_log_processor_group` → delivery log worker | `billing_events` and `campaign_delivery_logs` (+ delivery jobs) |

Dedup for telemetry: `lock:telemetry:{pseudo}:{campaign}:{event}` via `SETNX` (impression ~5m, click ~1h).

### Dormant Redis keys (events module not mounted)

- `event:dedup:{event_id}`
- `user_events:{pseudonymous_id}` (LPush, trimmed)

---

## 6. Workers & scheduled jobs

| Starter | Consumes | Effect |
|---------|----------|--------|
| `StartBillingStreamWorker` | `stream:billing_events` (group `billing_processor_group`) | Insert billing rows, update daily spend, mark budget exhausted |
| `StartDeliveryLogStreamWorker` | same stream (group `delivery_log_processor_group`) | Insert `campaign_delivery_logs` / jobs |
| `StartIntentLogWorker` | `queue:intent_logs` | Batch insert intent profiles |
| `StartAnalyticsAggregateWorker` | `queue:analytics_aggregate` | Upsert anonymous aggregate counts |
| `StartTargetingJob` | Timer (default 5m) | Match campaigns → `Publish(CampaignMatched)` |
| `StartIntentConsistencyJobs` | Timer (24h + quick startup) | Scan sustained intents → audience candidates/segments |

Workers are **adapters at the edge**: they call application services (`EventProcessor`, etc.) and must not contain business rules duplicated from domain.

---

## 7. Module-by-module explanation

### 7.1 `platform` — shared kernel & composition

| Area | Role |
|------|------|
| `bootstrap/` | Wire systems: consent, intents, delivery SDK, permissions, campaigns targeting, analytics jobs, admin ports |
| `database/` | Connect, Migrate, seeds (demo users, SMS recipients, billing catalog, …) |
| `messaging/` | In-process bus |
| `redis/` | Redis client wrapper |
| `route/` | Top-level Gin router |
| `middleware/` | CORS, recovery, etc. |
| `http/` | Shared `APIError` / JSON helpers |
| `security/` | Signature helpers for SDK |

### 7.2 `auth` — developer portal & SDK keys

Classic controller/service/repository (less hexagonal than newer modules).

- Register/login developers
- Applications + API keys (`pk_live_…`, `sk_secret_…`)
- Returns **SDK auth middleware** used on `/api/v1` Flutter routes

HTTP under `/api/v1/portal`.

### 7.3 `ad_portal` — advertiser / operator portal auth

- Portal users, advertisers, roles
- JWT login (`BearerAuth` in Swagger)
- `routes.Register` is the hub that mounts billing, audience, campaigns, admin, analytics under `/api/v1/ad-portal`

### 7.4 `permissions` — RBAC

- Roles, permissions, assignments
- Checker used by admin routes
- Redis/memory cache for role permission sets
- Can publish `PermissionAssigned` / `RoleCreated` (consumers not critical today)

### 7.5 `users` — identity rows for consent

No public HTTP. Exists so consent does not write to `users` directly (bounded-context split).

- Consumes `ConsentRegistrationRequested`
- Creates `users` row
- PublishSync `UserProvisionedForConsent`

### 7.6 `consent` — SDK registration

- `POST /api/v1/consent`
- Persists `consents` + `pseudonymous_mappings` after user provision
- Demo path: `sms_consented=true` → return mapping linked to `demo_sms_recipients`
- Domain events: registration requested / created

### 7.7 `intents` — on-device ML → ads

| Endpoint | Behavior |
|----------|----------|
| `POST /intents/ingest-ad` | Validate profile → cache active intent → enqueue/save → select campaign → maybe SMS dispatch → 200 ad body or 202 SMS accepted |
| `POST /intents/ingest-aggregate` | Enqueue anonymous counters to Redis list → 202 |

**Ports:** `ProfileRepository`, `ActiveIntentCache`, `AdSelector`, `SMSAdDispatcher`.

**Ad selection rules (via campaigns `IntentAdSelector`):**

- `sms_consented=true`, empty channel → try `SMS_PLUS` first, else non-SMS channels
- `sms_consented=false` → never `SMS_PLUS`
- Explicit `channel_code` (normalized: `SMS` → `SMS_PLUS`)

Eligibility is not “campaign exists”: must be `is_active`, `validation_status=passed`, `moderation_status=approved`, and advertiser subscription **in current period**, plus Redis budget/frequency filters.

### 7.8 `campaigns` — creative inventory & targeting

| Piece | Role |
|--------|------|
| `CampaignService` | Advertiser CRUD / create with segment purchase |
| `ModerationService` | Operator validate/activate; publishes admin campaign events |
| `IntentAdSelector` | Rank eligible campaigns by plan fee for ingest-ad |
| `AnonymousCampaignService` | Master list + click tokens for non-consented surfaces |
| `TargetingJob` | Periodic segment/intent matching → `CampaignMatched` |
| `CachedCampaignRepository` | Redis + Postgres eligibility |

HTTP under ad-portal `/campaigns`. Admin moderation under `/ad-portal/admin/campaigns/...`.

### 7.9 `audience` — Audiencemart

| Concept | Meaning |
|---------|---------|
| `AudienceSegment` | Purchasable cohort (intent + CPM) |
| `SegmentCandidate` | Pending cohort from intent-consistency scan |
| `SegmentPurchase` | Advertiser entitlement window |
| Memberships | Users in a segment |
| `TargetingResolver` | Who may be targeted for a campaign |
| `FindingProcessorAdapter` | Analytics findings → candidates/segments |

Purchases are confirmed inside campaign creation (`ConfirmPurchaseTx`), not a separate Flutter route.

### 7.10 `analytics` — operator reads + consistency + aggregates

- Read models: overview, campaigns, delivery, revenue, advertisers
- `AnalyzeIntentConsistencyUseCase` — sustained intent scan (scheduled)
- `AggregateIngestService` — enqueue device batches for non-consented users
- Aggregate worker upserts weighted counts

Findings currently flow through a **direct adapter** into audience (not bus).

### 7.11 `billing` — money & entitlements

| Piece | Role |
|-------|------|
| Plans / channels / rates | Catalog |
| Subscriptions | Advertiser plan periods (eligibility join) |
| `EventProcessor` | Stream consumer logic for impression/click |
| `PlanConsumer` | On plan created → seed rates |
| Admin catalog handlers | Via admin module ports |

HTTP: list plans/channels, get/create subscription.

### 7.12 `delivery` — SDK delivery & SMS+

| Piece | Role |
|-------|------|
| Anonymous campaigns | `GET /campaigns/anonymous` |
| Telemetry ingest | `POST /telemetry/track` → billing stream |
| CPC anonymous click | Tokenized click → stream |
| `SMSDispatchService` | Resolve recipient by pseudo → write `sms_send_attempts` → mock/Twilio send → delivery log |
| `SMSClickService` | Public `GET /telemetry/sms/click?token=` → mark clicked + stream click |
| Twilio webhook | Status updates when `SMS_PROVIDER=twilio` |
| Debug sends | `GET /telemetry/sms/debug/sends` (masked phones) |
| `SMSPlusConsumer` | Bus `CampaignMatched` → dispatch |

Mock SMS does not hit a real phone; opening the tracking URL (Swagger, browser, crawler) still marks `clicked`.

### 7.13 `admin` — operator use cases

Thin BC that owns **operator orchestration** ports:

- Approve/reject segment candidates
- List SDK users with intents
- HTTP for plans, segments, campaign moderation, intent-consistency run

Does not own campaign/billing tables; calls into those modules via Wire/bootstrap.

### 7.14 `events` — **not mounted**

Package retained for later. Would ingest raw SDK events, dedupe in Redis, publish bus events. Router explicitly does not mount it.

### 7.15 `rewards` — **dormant**

Tables AutoMigrated and seeded; `ImpressionRewardConsumer` exists but is **not registered** in bootstrap.

### 7.16 Stubs

`advertisers/`, `fraud/`, `geofencing/`, `payments/` (Telebirr stub), `utils/`, `websocket/` — placeholders or incomplete.

---

## 8. End-to-end flows

### A. Consented in-app / SMS ingest

```
POST /consent { sms_consented }
  true  → demo pseudonymous_id (SMS recipient)
  false → saga creates user + mapping + consent

POST /intents/ingest-ad { pseudonymous_id, intent_name, sms_consented, channel_code? }
  → persist/cache intent
  → SelectBestCampaign (subscription + moderation + Redis filters)
  → if SMS_PLUS: DispatchCampaignMatch → 202 {status:accepted}
  → else: 200 { ad_content, campaign_id, ... }
```

### B. Telemetry write-behind

```
POST /telemetry/track | CPC click | SMS click
  → (optional SETNX dedup)
  → XAdd stream:billing_events
       ├─ billing worker → billing_events + spend
       └─ delivery worker → campaign_delivery_logs
```

### C. Audience discovery loop

```
IntentConsistencyJobs (daily)
  → findings → ProcessIntentFindingUseCase
  → pending SegmentCandidate
  → admin approve → AudienceSegment + memberships
  → advertiser buys via campaign create → SegmentPurchase
  → TargetingJob matches → CampaignMatched → SMS+
```

---

## 9. Database migrations — recommendation

### What you have today (hybrid)

`database.Migrate` combines:

1. **Selectively embedded SQL** (`//go:embed` a subset of `migrations/*.sql`)
2. **GORM AutoMigrate** for many tables
3. **Ad-hoc ensure/align helpers** + **Go seeds**
4. **Orphan SQL files** in `migrations/` that are **not** applied by current Migrate

Sensitive tables (`consents`, `pseudonymous_mappings`, SMS demo tables) intentionally skip AutoMigrate because GORM fights explicit unique constraint names.

### Should everything be one SQL file?

**No — not recommended as the long-term architecture.**

A single mega-file is hard to review, hard to roll forward safely on existing environments, and fights team collaboration. What *is* recommended:

| Do | Don’t |
|----|--------|
| One **migration strategy** (e.g. golang-migrate / goose / embed-all with a `schema_migrations` version table) | Mix AutoMigrate + selective embeds + forgotten SQL files |
| Many **small, timestamped** SQL files applied in order | One giant forever-growing `schema.sql` as the only history |
| Seeds as a separate, idempotent job | Hide destructive `DROP TABLE` in everyday migrate without guards |
| Explicit SQL for constraint-sensitive tables | Rely on GORM to rename constraints in production |

**Practical target for this repo**

1. Keep **versioned SQL files** under `migrations/` (as you already name them `YYYYMMDDHHMMSS_*.sql`).
2. Embed **all** active migrations and apply them in order, recording versions in `schema_migrations`.
3. Turn off (or gate behind `DEV_AUTOMIGRATE=true`) GORM AutoMigrate for production.
4. Leave unused historical SQL archived or delete after confirming content is covered by the versioned chain.
5. Keep demo seeds in Go (`demo_user_seed.go`) or a `seed/` job — not mixed into every schema migration blindly.

So: **one ledger and one runner**, not necessarily one physical file.

---

## 10. API surface map (current)

| Area | Prefix | Auth |
|------|--------|------|
| Health | `/ping` | none |
| Swagger | `/swagger` | none |
| Developer portal | `/api/v1/portal` | developer JWT / register |
| Ad portal | `/api/v1/ad-portal/...` | portal JWT (+ RBAC on admin) |
| SDK consent / intents / delivery | `/api/v1/...` | X-API-Key (+ signature on mutating) |
| SMS click / Twilio | `/api/v1/telemetry/sms/...` | public (token/signature verified) |

---

## 11. Design principles to preserve

1. **Handlers thin; use cases own rules.**
2. **Bootstrap owns cross-module wiring** — no `intents` → `delivery/infrastructure` imports.
3. **Sync bus for request sagas; async bus for side effects; Redis for multi-worker durability.**
4. **Postgres is source of truth;** Redis is cache/queue/stream.
5. **Eligibility is subscription + moderation + budget/frequency**, not merely `is_active`.
6. **Migrations: one strategy, many versioned files** — do not collapse into a single mega SQL file.

---

## 12. Related docs

- OpenAPI: `docs/swagger.yaml` / Swagger UI `/swagger/index.html`
- This file is the full-system architecture reference.
