# Geofencing ads — Flutter frontend guide

How to implement store-enter ads in the Flutter SDK app, and how to verify them on the **Android Emulator Extended Controls → Location** panel (the emulator “debugging dashboard” for GPS).

Backend product rules and portal setup: [`GEOFENCING_SWAGGER_TEST.md`](./GEOFENCING_SWAGGER_TEST.md).  
Swagger: `http://localhost:8081/swagger/index.html` (API usually on **8081**).

---

## What the app must do

On every geofence **enter**, the backend:

1. Checks **location ad consent**
2. Always inserts a `store_visits` row
3. Returns an ad (**202** with `ad_content`) when **either**:
   - **Intent + zone:** same `pseudonymous_id` has a current intent **and** a live campaign linked to that zone has matching `target_intent`, **or**
   - **Returning visitor:** user already has **prior** `store_visits` (any zone) → any live zone campaign is eligible

**202 without `ad_content` is normal** for a first-ever store visit with no matching intent. Do not treat that as a client bug.

Sync (`GET /geofences/sync`) is unchanged: register native geofences from active zones only. Matching happens only on `POST /geofence/event`.

---

## Prerequisites (backend / portal)

Before Flutter testing:

1. API + PostGIS + Redis running (`docker compose up` from repo root).
2. SDK keys: `pk_live_…` + `sk_secret_…` (developer portal application).
3. Advertiser created an **active** zone + **approved** campaign linked via `campaign_geofence_targets`.
4. Campaign `target_intent` set if you want to test the intent path (e.g. `coffee_interest`).
5. A demo `pseudonymous_id` (from `POST /consent` with `sms_consented: true`, or Adminer `demo_sms_recipients`).

Note zone lat/lng/radius from portal or Swagger — you will spoof that GPS fix in the emulator.

---

## Auth (every SDK call)

| Header | Value |
|--------|--------|
| `X-API-Key` | Publishable key `pk_live_…` |
| `X-Signature` | Lowercase hex **HMAC-SHA256(`sk_secret_…`, exact raw JSON body)** |
| `Content-Type` | `application/json` |

- GET sync: API key only (no body / no signature).
- POST/PATCH (consent, intent ingest, geofence event): API key **and** signature over the **exact** body bytes you send.

Never put `sk_secret_…` in `X-Signature`. Sign the body with it.

Dart sketch:

```dart
import 'dart:convert';
import 'package:crypto/crypto.dart';

String signBody(String secret, String rawJsonBody) {
  final digest = Hmac(sha256, utf8.encode(secret)).convert(utf8.encode(rawJsonBody));
  return digest.toString(); // lowercase hex
}
```

Use one stable `baseUrl` (e.g. `http://10.0.2.2:8081` on Android emulator → host machine localhost).

---

## Implementation checklist

### 1. Stable identity

Persist one `pseudonymous_id` after consent. Use **the same id** for:

- `POST /intents/ingest-ad`
- `PATCH /geofences/location-consent`
- `POST /geofence/event`

If ids differ, intent will never unlock a geofence ad.

### 2. Location permission + location ad consent

1. Request OS location permission (and background if you use OS geofencing).
2. Call:

```http
PATCH /geofences/location-consent
{
  "pseudonymous_id": "<id>",
  "location_ad_consent": true
}
```

Without this, enter events return **403**.

### 3. Sync nearby zones

On app start / significant location change / pull-to-refresh:

```http
GET /geofences/sync?lat=<lat>&lng=<lng>&radius_m=20000
```

Response:

```json
{
  "status": "success",
  "zones": [
    {
      "id": "<zone_id>",
      "latitude": 9.022736,
      "longitude": 38.746799,
      "radius_metres": 150,
      "is_active": true
    }
  ]
}
```

Register each zone with the platform geofencing API (`geolocator` / `geofencing_api` / native `GeofencingClient`). Keep `zone.id` as the native geofence request id (or map request id → `zone_id`).

Only **active** zones appear. Empty list usually means admin has not activated the store yet.

### 4. Send intents when available

When on-device ML (or a debug picker) produces an intent:

```http
POST /intents/ingest-ad
{
  "pseudonymous_id": "<id>",
  "intent_name": "coffee_interest",
  "confidence": 0.9,
  "model_version": "1.0.0",
  "channel_code": "PUSH",
  "sms_consented": true
}
```

This caches `user_intent:{pseudonymous_id}` (and persists intent). Geofence matching reads Redis first, then latest DB intent.

Still call this when you have a prediction — the intent path is preferred when it yields zone campaigns.

### 5. On geofence enter → event

When the OS fires an **enter** for a registered zone:

```http
POST /geofence/event
{
  "pseudonymous_id": "<id>",
  "zone_id": "<zone_id from sync>",
  "accuracy_m": 12.5
}
```

Handle **202** in two ways:

| Body | UI |
|------|----|
| Has `ad_content` | Show creative (title / body / image / destination) |
| No `ad_content` | Visit recorded only — no ad UI (first cold enter is expected) |

Example with ad:

```json
{
  "status": "accepted",
  "visit_id": "...",
  "visited_at": "...",
  "campaign_id": "...",
  "campaign_name": "...",
  "channel_code": "PUSH",
  "ad_content": {
    "title": "...",
    "body_text": "...",
    "image_url": "...",
    "destination_url": "..."
  }
}
```

Suggested client flow:

```text
OS enter → POST /geofence/event
         → if ad_content != null → show in-app / local notification
         → else → log "visit only" (optional debug snackbar)
```

Do **not** invent client-side “always show an ad on first enter.” Matching is server-side.

### 6. What not to change

- Do not filter campaigns by intent on the client for geofence ads.
- Do not skip `POST /geofence/event` when you think “no ad will come.”
- Leave `/geofences/sync` as “register fences only.”

---

## Expected scenarios (for QA / debug UI)

| Scenario | Setup | Expected `POST /geofence/event` |
|----------|--------|----------------------------------|
| First visit + matching intent | Fresh user, ingest `intent_name` = campaign `target_intent`, then enter | **202 + `ad_content`** |
| First visit + no intent | Fresh user, no ingest, enter | **202, no `ad_content`** |
| Returning visitor + no intent | Enter once (visit only), enter again | Second enter: **202 + `ad_content`** (if live zone campaigns exist) |
| Returning + intent match | Prior visits + matching ingest | Ad from **intent-filtered** set |
| Returning + intent mismatch | Prior visits + unrelated intent | Still ad via **visit-history** branch |

Wipe visits / use a new `pseudonymous_id` when retesting “first visit.”

---

## Test on Android Emulator Location dashboard

Use the emulator’s **Extended Controls** location panel to fake GPS (this is the usual “debugging dashboard” for geofence testing).

### Open the panel

1. Start an AVD and run the Flutter app.
2. On the emulator toolbar, click **⋯** (Extended controls).
3. Open **Location**.
4. Use **Single points** (one fix) or **Routes** (walk into / out of the fence).

### Point A — outside the store

1. Pick a lat/lng **outside** the zone radius (e.g. ~500 m away from the store center used in portal).
2. Click **Set location** (or **Send** on older UI).
3. In the app: grant location permission → ensure location consent → trigger sync (or wait for your refresh).
4. Confirm sync returned your zone and native geofences were registered (log `zone_id`s).

### Point B — enter the store (trigger enter)

1. In Extended Controls → Location, set lat/lng to the **zone center** from sync/portal  
   (docs sample: `9.022736`, `38.746799` — only if your zone was created there).
2. **Set location** again.
3. OS should fire a geofence **enter** → app calls `POST /geofence/event`.

Tips:

- If enter does not fire, move **outside** again, wait a few seconds, then set the center again.
- Prefer accuracy ≤ zone radius; pass `accuracy_m` from the location fix when available.
- Background geofencing on emulator can be flaky; a **debug “Simulate enter”** button that posts `/geofence/event` with the synced `zone_id` is strongly recommended.

### Optional: Routes tab

1. Save an outside point and the store center.
2. Build a short route outside → inside.
3. **Play route** to watch enter/exit over time.

### Optional: ADB (same idea)

```bash
# Note: longitude then latitude
adb emu geo fix 38.746799 9.022736
```

### Flutter / DevTools logging

While testing, log:

- current mock lat/lng
- synced zone ids + radii
- enter callbacks
- raw `/geofence/event` status + whether `ad_content` was present

DevTools **Logging** / your debug overlay is enough; GPS spoofing still comes from Extended Controls → Location.

---

## Suggested debug screen (emulator-friendly)

Add a temporary debug page (behind a flavor flag):

1. Show `pseudonymous_id`
2. Toggle / call location consent
3. Intent dropdown → `POST /intents/ingest-ad`
4. Button **Sync now** (uses last known / mock location)
5. List synced zones
6. Button **Simulate enter** per zone → `POST /geofence/event`
7. Show last response JSON (highlight missing `ad_content`)

That lets you validate the OR rules without fighting OS geofence timing, then re-test once with Extended Controls Location for end-to-end GPS.

---

## Host URL on emulator

| Runtime | Base URL |
|---------|----------|
| Android emulator → API on host | `http://10.0.2.2:8081` |
| iOS simulator → API on host | `http://127.0.0.1:8081` |
| Physical device | Your LAN IP / tunnel (same ports) |

Cleartext HTTP may need `android:usesCleartextTraffic` / network security config for local debug builds.

---

## Failure cheat sheet

| Symptom | Likely cause |
|---------|----------------|
| 401 | Bad API key or HMAC (body bytes ≠ signed bytes) |
| 403 on event | `location_ad_consent` false / not set |
| 404 demo recipient | Unknown `pseudonymous_id` |
| Sync empty | Zone still `is_active=false` or wrong lat/lng/radius |
| 202 no ad, first enter | Expected without matching intent |
| 202 no ad, returning user | No live linked campaigns, budget, or frequency cap |
| Intent never helps | Different `pseudonymous_id` than geofence event, or `target_intent` mismatch |

---

## Minimal client sequence

```text
1. Consent → save pseudonymous_id
2. PATCH location-consent = true
3. (Optional) POST intents/ingest-ad with same id
4. GET geofences/sync → register native fences
5. On enter (or debug Simulate enter):
      POST geofence/event
6. If ad_content present → show creative
```

Backend setup and campaign approval remain in [`GEOFENCING_SWAGGER_TEST.md`](./GEOFENCING_SWAGGER_TEST.md).
