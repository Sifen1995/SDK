# Skykin API — Postman guide (Ad Portal, Developer Portal, SDK)

This document is for **frontend (React/Next.js campaign portal)** and **Flutter/mobile SDK** developers. Import the Postman assets from the repo root:

| File | Purpose |
|------|---------|
| [`postman/Skykin-API.postman_collection.json`](../postman/Skykin-API.postman_collection.json) | All HTTP endpoints + WebSocket notes |
| [`postman/Skykin-Local.postman_environment.json`](../postman/Skykin-Local.postman_environment.json) | Local defaults (`baseUrl`, tokens, IDs) |

**Import in Postman:** *File → Import* → select both files → choose environment **Skykin — Local**.

**Swagger UI:** [http://localhost:8081/swagger/index.html](http://localhost:8081/swagger/index.html) — tags **Ad Portal - Auth**, **Ad Portal - Campaigns**, **Ad Portal - Admin**. Regenerate: `make swagger`.

**Architecture:** See [`docs/AD_CAMPAIGN_ARCHITECTURE.md`](AD_CAMPAIGN_ARCHITECTURE.md) for module split (`advertisers` vs `campaigns`) and delivery flow.

**DB migration (required):** Run [`internal/platform/database/migrations/20260603120000_advertisers_campaigns.sql`](../internal/platform/database/migrations/20260603120000_advertisers_campaigns.sql) against Postgres before using ad features.

---

## Base URL & health

| Variable | Default | Description |
|----------|---------|-------------|
| `baseUrl` | `http://localhost:8081` | API host (Docker maps `8081`) |
| `adPortalToken` | *(set by Login)* | JWT for campaign portal |
| `portalToken` | *(set by Login)* | JWT for developer portal |
| `publishableKey` | `pk_live_...` | SDK `X-API-Key` |
| `secretKey` | `sk_...` | HMAC secret (never ship in client apps) |
| `userId` | `user-demo-001` | SDK end-user id |
| `campaignId` | *(set by Create campaign)* | Active campaign UUID |
| `creativeId` | *(set by Create creative)* | Creative UUID |

**Health check:** `GET {{baseUrl}}/ping` → `{ "status": "ok" }`

**Run stack:**
```bash
docker compose up --build
```
Ensure `.env` includes `JWT_SECRET` (required for all JWT flows).

---

## API surfaces (three separate prefixes)

```
┌─────────────────────────────────────────────────────────────────┐
│  /api/v1/ad-portal     Campaign portal (advertisers) — JWT      │
│  /api/v1/portal        Developer portal (SDK keys) — JWT        │
│  /api/v1               SDK (events, predict, ws) — API key+HMAC │
└─────────────────────────────────────────────────────────────────┘
```

Do **not** mix tokens: `adPortalToken` ≠ `portalToken`.

---

## 1. Ad Portal — `/api/v1/ad-portal`

Used by the **campaign management UI** (React/Next.js). Authentication is **Bearer JWT** after login.

### Roles

| Role | `role` value | Write (create/edit) | Read |
|------|----------------|---------------------|------|
| Operator Admin | `operator_admin` | Yes + user management | All campaigns |
| Advertiser | `advertiser` | Own account | Own campaigns |
| Read-Only Analyst | `read_only_analyst` | No | Own account |

Default seeded admin (override via env):

- `ADMIN_EMAIL` → default `admin@skykin.com`
- `ADMIN_PASSWORD` → default `Admin12345!`

### Auth (public)

#### Register advertiser or analyst
`POST {{baseUrl}}/api/v1/ad-portal/register`

```json
{
  "name": "Acme Ads",
  "email": "advertiser@test.com",
  "password": "SecurePass1!",
  "company_name": "Acme Inc",
  "role": "advertiser"
}
```

- `role`: `advertiser` (default) or `read_only_analyst`
- Cannot self-register as `operator_admin`

**201 response:**
```json
{
  "user": {
    "id": "uuid",
    "name": "Acme Ads",
    "email": "advertiser@test.com",
    "company_name": "Acme Inc",
    "role": "advertiser"
  }
}
```

#### Login
`POST {{baseUrl}}/api/v1/ad-portal/login`

```json
{
  "email": "advertiser@test.com",
  "password": "SecurePass1!"
}
```

**200 response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { "id": "...", "name": "...", "email": "...", "role": "advertiser" }
}
```

Store `token` as `adPortalToken`. Header for protected routes:

```
Authorization: Bearer {{adPortalToken}}
```

### Profile

#### Me
`GET {{baseUrl}}/api/v1/ad-portal/me`

### Campaigns

#### Create campaign
`POST {{baseUrl}}/api/v1/ad-portal/campaigns`  
Requires: **advertiser** or **operator_admin**

```json
{
  "name": "Crypto Promo",
  "target_intent": "crypto_interest",
  "channel": "banner",
  "budget": 1000
}
```

| Field | Values |
|-------|--------|
| `target_intent` | ML intent slug, e.g. `crypto_interest`, `fashion_interest`, `food_interest`, `gaming_interest`, `fintech_interest`, `education_interest`, `general_interest` |
| `channel` | `banner` \| `sms_plus` \| `push` |

**201:** Campaign object with `id`, `status: "draft"`, etc.

#### List campaigns
`GET {{baseUrl}}/api/v1/ad-portal/campaigns`

**200:**
```json
{ "campaigns": [ { "id": "...", "name": "...", "target_intent": "...", "channel": "banner", "status": "draft", ... } ] }
```

#### Get campaign
`GET {{baseUrl}}/api/v1/ad-portal/campaigns/{{campaignId}}`

#### Activate campaign
`POST {{baseUrl}}/api/v1/ad-portal/campaigns/{{campaignId}}/activate`

- Requires at least one creative with `validation_status: "passed"`
- Sets `status` to `active`

### Creatives

Creative `format` must match campaign `channel`.

| Channel | Format | Payload fields |
|---------|--------|----------------|
| `banner` | `banner` | `image_url`, `click_url` |
| `push` | `push` | `title` (1–50), `body` (1–120), `icon_url` (optional), `deep_link` (optional) |
| `sms_plus` | `sms_plus` | `image_url`, `title` (1–40), `description` (1–160), `cta_url` |

#### Create creative
`POST {{baseUrl}}/api/v1/ad-portal/campaigns/{{campaignId}}/creatives`

**Banner example:**
```json
{
  "format": "banner",
  "payload": {
    "image_url": "https://via.placeholder.com/320x50.png",
    "click_url": "https://example.com"
  }
}
```

**Push example:**
```json
{
  "format": "push",
  "payload": {
    "title": "Flash sale",
    "body": "20% off today only",
    "icon_url": "https://via.placeholder.com/64",
    "deep_link": "https://example.com/deals"
  }
}
```

**SMS+ example:**
```json
{
  "format": "sms_plus",
  "payload": {
    "image_url": "https://via.placeholder.com/150",
    "title": "Deal",
    "description": "Save on crypto fees",
    "cta_url": "https://example.com"
  }
}
```

**201 response:**
```json
{
  "id": "uuid",
  "campaign_id": "uuid",
  "format": "banner",
  "validation_status": "passed",
  "validation_notes": "",
  "payload": { ... },
  "is_active": true,
  "created_at": "..."
}
```

| `validation_status` | Meaning |
|---------------------|---------|
| `passed` | OK; counts toward activation |
| `warning` | Active but URL check failed; alone cannot activate |
| `failed` | Not active; fix payload |

Use reachable URLs for `click_url`, `cta_url`, `deep_link` (server does HEAD/GET check).

#### List creatives
`GET {{baseUrl}}/api/v1/ad-portal/campaigns/{{campaignId}}/creatives`

#### Preview (simulator)
`GET {{baseUrl}}/api/v1/ad-portal/campaigns/{{campaignId}}/creatives/{{creativeId}}/preview`

**200:**
```json
{
  "format": "banner",
  "campaign_name": "Crypto Promo",
  "simulator": true,
  "channel_label": "In-App Banner",
  "preview": { "image_url": "...", "click_url": "..." }
}
```

### Admin (operator only)

#### Create portal user
`POST {{baseUrl}}/api/v1/ad-portal/admin/users`  
Login as `operator_admin`.

```json
{
  "name": "Analyst Jane",
  "email": "analyst@acme.com",
  "password": "SecurePass1!",
  "role": "read_only_analyst"
}
```

`role`: `advertiser` \| `read_only_analyst` \| `operator_admin`

### Error shape (Ad Portal)

```json
{
  "status": "error",
  "code": 400,
  "message": "activation failed",
  "details": "campaign cannot be activated without..."
}
```

---

## 2. Developer Portal — `/api/v1/portal`

Used once per app to obtain **SDK API keys** (backend/CI holds `secret_key`; mobile app only gets `publishable_key`).

#### Register developer
`POST {{baseUrl}}/api/v1/portal/register`

```json
{
  "name": "Mobile Team",
  "email": "dev@company.com",
  "password": "securepass123"
}
```

#### Login developer
`POST {{baseUrl}}/api/v1/portal/login` → save `token` as `portalToken`

#### Create application (get SDK keys)
`POST {{baseUrl}}/api/v1/portal/applications`  
`Authorization: Bearer {{portalToken}}`

```json
{
  "app_name": "My Flutter App",
  "platform": "flutter",
  "bundle_id": "com.company.myapp"
}
```

**201** — save `credentials.publishable_key` and `credentials.secret_key` (shown once).

#### List applications
`GET {{baseUrl}}/api/v1/portal/applications`

---

## 3. SDK — `/api/v1` (Flutter / native)

All SDK routes require:

| Header | Value |
|--------|--------|
| `X-API-Key` | `{{publishableKey}}` |
| `X-Signature` | `HMAC-SHA256(secretKey, rawBody)` as **lowercase hex** |
| `Content-Type` | `application/json` |

The Postman collection runs the HMAC script automatically on the **SDK** folder.

> **Flutter:** Compute HMAC on the exact JSON bytes you send. Use `secret_key` only on a secure backend proxy, or use a dev build with a test key — never embed production secrets in the app binary.

### Ingest events
`POST {{baseUrl}}/api/v1/events`

```json
{
  "user_id": "user-demo-001",
  "events": [
    {
      "event_id": "550e8400-e29b-41d4-a716-446655440001",
      "event_type": "content_viewed",
      "domain": "crypto",
      "session_id": "660e8400-e29b-41d4-a716-446655440002",
      "screen_name": "asset_details",
      "metadata": { "asset": "BTC", "duration_ms": 4200 },
      "device_type": "mobile",
      "platform": "flutter",
      "app_version": "1.0.0",
      "created_at": "2026-06-02T12:00:00Z"
    }
  ]
}
```

**202 response:**
```json
{
  "accepted": true,
  "prediction_queued": true,
  "results": [ { "event_id": "...", "status": "accepted" } ]
}
```

Use a **new UUID** per `event_id` (repeats return `duplicate`).

Supported `event_type` values include: `session_started`, `screen_viewed`, `content_viewed`, `search_performed`, `interaction_received`, `scroll_activity`, `notification_opened`, `campaign_impression`, `campaign_clicked`, `conversion_completed`, `transaction_completed`, `reward_claimed`.

Send several events in the `crypto` domain before predicting.

### Predict intent (sync)
`POST {{baseUrl}}/api/v1/intents/predict`

```json
{
  "user_id": "user-demo-001"
}
```

**200** (when ML returns an intent):
```json
{
  "user_id": "user-demo-001",
  "status": "predicted",
  "intent": "crypto_interest",
  "confidence": 0.82,
  "reward_triggered": true,
  "top_signals": ["..."]
}
```

After prediction, if an **active** campaign matches `intent` + channel, a **campaign ad** is pushed on the WebSocket (see below).

### WebSocket — rewards & campaign ads
`GET {{baseUrl}}/api/v1/ws/rewards/{{userId}}` (upgrade to WebSocket)

**Postman:** New → WebSocket Request → URL above (no HMAC on WS in current build).

**Flutter:** use `web_socket_channel` or similar; connect before calling predict.

**Message — campaign ad:**
```json
{
  "type": "campaign_ad",
  "intent": "crypto_interest",
  "channel": "banner",
  "campaign_id": "uuid",
  "campaign_name": "Crypto Promo",
  "creative_id": "uuid",
  "creative_format": "banner",
  "content": {
    "image_url": "https://...",
    "click_url": "https://..."
  }
}
```

**Message — reward:**
```json
{
  "type": "reward_earned",
  "reward_id": "uuid",
  "reward_type": "coins",
  "amount": 50,
  "currency": "COINS",
  "message": "Crypto enthusiast! You earned 50 Coins!",
  "intent": "crypto_interest",
  "confidence": 0.82
}
```

Render by `type`: `campaign_ad` → banner / SMS+ / push UI; `reward_earned` → reward modal.

---

## End-to-end test order (Postman)

### A — Campaign portal (frontend dev)

1. **Register** or **Login** → `adPortalToken` saved  
2. **Create campaign** → `campaignId` saved (`target_intent: crypto_interest`, `channel: banner`)  
3. **Create creative** (passed validation) → `creativeId` saved  
4. **Activate campaign**  
5. **Preview creative** (optional UI check)

### B — SDK + ad delivery (Flutter dev)

1. **Developer Login** → `portalToken`  
2. **Create application** → `publishableKey`, `secretKey`  
3. Open **WebSocket** `ws://localhost:8081/api/v1/ws/rewards/{{userId}}`  
4. **Ingest events** (3–5 crypto events, unique `event_id`s)  
5. **Predict intent**  
6. Confirm WebSocket receives `campaign_ad` (and optionally `reward_earned`)

---

## Frontend integration notes (React/Next.js)

```typescript
const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8081";

async function login(email: string, password: string) {
  const res = await fetch(`${API}/api/v1/ad-portal/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  const data = await res.json();
  localStorage.setItem("adPortalToken", data.token);
  return data;
}

function authHeaders(): HeadersInit {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${localStorage.getItem("adPortalToken")}`,
  };
}
```

Role-gate UI: hide create/activate buttons when `user.role === "read_only_analyst"`.

---

## Flutter integration notes

```dart
import 'dart:convert';
import 'package:crypto/crypto.dart';
import 'package:http/http.dart' as http;

String signBody(String secret, String body) {
  final h = Hmac(sha256, utf8.encode(secret));
  return h.convert(utf8.encode(body)).toString();
}

Future<void> ingestEvents({
  required String baseUrl,
  required String publishableKey,
  required String secretKey,
  required String userId,
  required List<Map<String, dynamic>> events,
}) async {
  final body = jsonEncode({"user_id": userId, "events": events});
  final res = await http.post(
    Uri.parse("$baseUrl/api/v1/events"),
    headers: {
      "Content-Type": "application/json",
      "X-API-Key": publishableKey,
      "X-Signature": signBody(secretKey, body),
    },
    body: body,
  );
  // expect 202
}
```

WebSocket URL: `ws://<host>:8081/api/v1/ws/rewards/$userId` — parse JSON by `type` field.

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| `401 login failed` | Check email/password; ensure `JWT_SECRET` is set |
| `403 read-only access` | Analyst role cannot POST campaigns/creatives |
| `activation failed` | Add a creative with `validation_status: passed` |
| `401 cryptographic payload signature mismatch` | Sign exact raw body; check `secretKey` |
| `duplicate` on events | Generate new `event_id` UUIDs |
| No WebSocket `campaign_ad` | WS connected before predict? Campaign `active`? Intent matches `target_intent`? |
| URL validation `failed` | Use `https://example.com` or other 2xx URLs |

---

## Collection maintenance

When new Ad Portal routes are added, update:

1. `internal/advertisers/interfaces/http/routes.go`
2. `postman/Skykin-API.postman_collection.json`
3. This file
