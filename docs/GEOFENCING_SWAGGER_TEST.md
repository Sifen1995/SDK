# Geofencing — Swagger step-by-step test guide

Open Swagger UI: **http://localhost:8081/swagger/index.html**

Backend must be running (`docker compose up`). Adminer (optional): **http://localhost:8082**.

Flutter / emulator implementation and Location-dashboard testing: [`GEOFENCING_FLUTTER_FRONTEND_GUIDE.md`](./GEOFENCING_FLUTTER_FRONTEND_GUIDE.md).

## Roles (important)

| Role | Who | What they do for geofencing |
|------|-----|-----------------------------|
| **Advertiser** | Ad-portal register/login | Creates **stores** (geofence zones), creates campaign, attaches stores |
| **Admin (operator)** | Seeded `admin@skykin.com` | Approves **campaign** and/or **geofence zones** if not yet activated — does **not** create zones or campaigns |
| **SDK app** | `pk_live_` + `sk_secret_` | Sync nearby **active** zones + enter events |

Until admin activates them, zones stay `is_active=false` and **do not appear** in `GET /geofences/sync`.

### Admin approval options

| Action | Endpoint | Effect |
|--------|----------|--------|
| Approve campaign (+ linked inactive zones) | `POST /ad-portal/admin/campaigns/{id}/validate` `{"action":"approve"}` | Campaign live + linked draft stores activated |
| List draft stores | `GET /ad-portal/admin/geofences/pending` | Inactive zones only |
| Activate one store | `POST /ad-portal/admin/geofences/{id}/activate` | Zone `is_active=true` if not already |
| Activate stores for a campaign | `POST /ad-portal/admin/campaigns/{id}/geofences/activate` | Linked inactive zones only |
| Activate approved campaign | `POST /ad-portal/admin/campaigns/{id}/activate` | Campaign live (if needed) + linked inactive zones |

---

## Auth cheat sheet

| Auth in Swagger | Used for |
|-----------------|----------|
| **BearerAuth** | Ad portal (advertiser **or** admin JWT — paste token only, no `Bearer `) |
| **APIKeyAuth** + **SDKSecretAuth** | SDK geofence routes (secret auto-signs POST/PATCH bodies) |

---

## Part A — SDK keys (developer portal)

### A1. `POST /portal/register`

```json
{
  "name": "Geo Tester",
  "email": "geo-dev@example.com",
  "password": "securepass123"
}
```

### A2. `POST /portal/login` → copy JWT → Authorize **BearerAuth**

### A3. `POST /portal/applications`

```json
{
  "app_name": "Geofence Demo App",
  "platform": "flutter",
  "bundle_id": "com.skykin.geofence.demo"
}
```

Save `publishable_key` (`pk_live_…`) and `secret_key` (`sk_secret_…`, once).

### A4. Authorize **APIKeyAuth** + **SDKSecretAuth** with those keys

---

## Part B — Advertiser creates stores + campaign (not admin)

### B1. Register / login as **advertiser**

`POST /ad-portal/register` (company advertiser) **or** login with an existing advertiser account.

```json
{
  "name": "Kaldi Coffee",
  "email": "advertiser@kaldi.test",
  "password": "SecurePass1!"
}
```

Then `POST /ad-portal/login` → copy JWT → Authorize **BearerAuth** (replaces any previous Bearer token).

> Do **not** use `admin@skykin.com` for creating zones/campaigns. Admin create/link is **403**.

### B2. Ensure subscription (if create campaign requires it)

1. `GET /ad-portal/plans`
2. Subscribe via billing routes if your env requires an active plan
3. `GET /ad-portal/channels` → save a `channel_id`

### B3. Create store geofence(s) — inactive draft

Tag: **Ad Portal - Geofences** → `POST /ad-portal/geofences`

```json
{
  "latitude": 9.022736,
  "longitude": 38.746799,
  "radius_metres": 150
}
```

Expect **201** with `"is_active": false`. Save `id` as `zone_id`.

Optional: `GET /ad-portal/geofences` — your stores (still inactive).

### B4. Create campaign

Tag: **Ad Portal - Campaigns** → `POST /ad-portal/campaigns`

```json
{
  "name": "Kaldi Macchiato Deal",
  "target_intent": "coffee_interest",
  "channel_id": "<channel_id>",
  "title": "Get 20% off Macchiato today!",
  "body_text": "Walk into Kaldi's and show this offer.",
  "image_url": "https://example.com/macchiato.jpg",
  "destination_url": "https://example.com/kaldi-offer",
  "daily_budget_cap": 500,
  "total_budget_cap": 5000,
  "frequency_cap_per_day": 3
}
```

Expect `moderation_status: pending`, `validation_status: pending`, `is_active: false`. Save `id` as `campaign_id`.

### B5. Link stores to campaign

`POST /ad-portal/campaigns/{campaign_id}/geofences`

```json
{
  "zone_ids": ["<zone_id>"]
}
```

Expect **204**.

Optional: `GET /ad-portal/campaigns/{campaign_id}/geofences`.

---

## Part C — Admin approves campaign and/or zones

### C1. Login as admin

```json
{
  "email": "admin@skykin.com",
  "password": "Admin12345!"
}
```

Authorize **BearerAuth** with the admin JWT.

### C2. Review drafts

- `GET /ad-portal/admin/campaigns/pending` — pending campaigns
- `GET /ad-portal/admin/geofences/pending` — inactive stores

### C3a. One-shot: approve campaign (also activates linked inactive zones)

`POST /ad-portal/admin/campaigns/{campaign_id}/validate`

```json
{
  "action": "approve",
  "notes": "Creative and store location look good"
}
```

Expect **200** with:

- `moderation_status`: `approved`
- `validation_status`: `passed`
- `is_active`: `true`

Linked inactive `geofence_zones` become `is_active=true`.

To reject instead:

```json
{
  "action": "reject",
  "notes": "Budget or location unclear"
}
```

Zones stay inactive on reject.

### C3b. Or approve separately

1. Approve/validate campaign as above (or reject)
2. Activate stores only if still inactive:
   - `POST /ad-portal/admin/geofences/{zone_id}/activate`
   - or `POST /ad-portal/admin/campaigns/{campaign_id}/geofences/activate`

Useful when stores were linked **after** campaign approve.

---

## Part D — SDK: sync + enter event

### D1. Demo `pseudonymous_id`

Adminer → `demo_sms_recipients`, or `POST /consent` with `"sms_consented": true`.

### D2. (Optional) `PATCH /geofences/location-consent`

```json
{
  "pseudonymous_id": "<demo id>",
  "location_ad_consent": true
}
```

### D3. `GET /geofences/sync?lat=9.022736&lng=38.746799&radius_m=20000`

- **Before admin activate:** `zones` empty (or without your draft store)
- **After activate:** your zone appears for native geofence registration

### D4. `POST /geofence/event`

```json
{
  "pseudonymous_id": "<demo id>",
  "zone_id": "<zone_id>",
  "accuracy_m": 12.5
}
```

Always records a `store_visits` row when consent is granted. An ad is returned (**202** with `ad_content`) when **either**:

1. **Intent + zone:** the same `pseudonymous_id` has a current intent (Redis `user_intent:` or latest DB intent) **and** a live campaign linked to this zone has matching `target_intent`, or  
2. **Returning visitor:** the user already has **prior** `store_visits` (any zone, excluding this enter) — then any live zone-linked campaign is eligible (subject to budget / frequency cap / plan ranking).

SDK should still send intents when available so the intent path can win. **202 without `ad_content`** is normal for a first-ever store visit with no matching intent.

---

## Flow diagram (matches product)

```text
[Advertiser]                         [System]                         [Admin]
     │                                  │                                │
     ├─► 1. Creates Store(s) ───────────┼─► geofence_zones               │
     │   (lat, lng, radius)             │   is_active = FALSE            │
     │                                  │                                │
     ├─► 2. Creates Campaign & ─────────┼─► campaign_geofence_targets    │
     │   attaches Stores                │   campaign pending             │
     │                                  │                                │
     └─► 3. Submitted for review ───────┼─► moderation_status=pending    │
                                        │                                │
                                        ├───────────────────────────────►│
                                        │                                ├─► 4. Reviews creative,
                                        │                                │   budget & geofence
                                        │◄───────────────────────────────┴─► APPROVE / REJECT
                                        │
                                        ├─► If approved:
                                        │   - validation_status = passed
                                        │   - moderation_status = approved
                                        │   - campaign.is_active = TRUE (if not already)
                                        │   - geofence_zones.is_active = TRUE (if not already)
```

---

## Minimal happy-path order

1. Developer keys → Authorize SDK  
2. **Advertiser** login → create zone → create campaign → link zones  
3. **Admin** login → pending lists → `POST .../validate` with `approve`  
4. SDK `GET /geofences/sync` → `POST /geofence/event`

---

## Status codes

| Call | Success | Notes |
|------|---------|--------|
| Advertiser `POST /geofences` | 201 | `is_active=false`; admin gets 403 |
| Advertiser link zones | 204 | |
| Admin validate approve | 200 | Activates campaign + inactive linked stores |
| Admin activate zone | 200 | Idempotent if already active |
| `GET /geofences/sync` | 200 | Only **active** zones |
| `POST /geofence/event` | 202 | 403 without location consent |

## Adminer checks

```sql
SELECT id, is_active, latitude, longitude FROM geofence_zones;
SELECT * FROM campaign_geofence_targets;
SELECT id, is_active, moderation_status, validation_status FROM campaigns;
```
