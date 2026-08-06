#!/usr/bin/env bash
set -euo pipefail

BASE="http://localhost:8081/api/v1"
OUT="/home/sifen/Desktop/my git/SDK/.geo_e2e"
mkdir -p "$OUT"
TS=$(date +%s)
ADV_EMAIL="geo-e2e-${TS}@kaldi.test"
ADV_PASS="SecurePass1!"
ADMIN_EMAIL="admin@skykin.com"
ADMIN_PASS="Admin12345!"

json_body() {
  python3 - <<'PY' "$1"
import json, sys
raw=open(sys.argv[1]).read().strip().splitlines()
body="\n".join(l for l in raw if not l.startswith("HTTP:"))
print(body)
PY
}

http_code() {
  python3 - <<'PY' "$1"
import sys
for l in open(sys.argv[1]).read().splitlines():
    if l.startswith("HTTP:"):
        print(l.split(":",1)[1].strip()); break
PY
}

hmac_sig() {
  python3 - <<'PY' "$1" "$2"
import hmac, hashlib, sys
secret, body = sys.argv[1], sys.argv[2]
print(hmac.new(secret.encode(), body.encode(), hashlib.sha256).hexdigest())
PY
}

echo "========== 1. ADVERTISER REGISTER + LOGIN =========="
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/ad-portal/register" \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"Kaldi Coffee E2E\",\"email\":\"${ADV_EMAIL}\",\"password\":\"${ADV_PASS}\"}" \
  | tee "$OUT/01_register.json"
echo "register HTTP: $(http_code "$OUT/01_register.json")"

curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/ad-portal/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"${ADV_EMAIL}\",\"password\":\"${ADV_PASS}\"}" \
  | tee "$OUT/02_adv_login.json"
ADV_TOKEN=$(python3 -c "import json; print(json.loads('''$(json_body "$OUT/02_adv_login.json")''')['token'])")
ADV_AUTH="Authorization: Bearer $ADV_TOKEN"
echo "ADV_EMAIL=$ADV_EMAIL"

echo
echo "========== 2. SUBSCRIBE + CHANNEL =========="
curl -s -H "$ADV_AUTH" "$BASE/ad-portal/plans" | tee "$OUT/03_plans.json"
PLAN_ID=$(python3 -c "import json; print(json.load(open('$OUT/03_plans.json'))['plans'][0]['id'])")
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/ad-portal/subscription" \
  -H "$ADV_AUTH" -H 'Content-Type: application/json' \
  -d "{\"plan_id\":\"$PLAN_ID\"}" | tee "$OUT/04_subscribe.json"
echo "subscribe HTTP: $(http_code "$OUT/04_subscribe.json")"

curl -s -H "$ADV_AUTH" "$BASE/ad-portal/channels" | tee "$OUT/05_channels.json"
CHANNEL_ID=$(python3 -c "import json; ch=json.load(open('$OUT/05_channels.json'))['channels']; print(next(c['id'] for c in ch if c['code']=='IN_APP_BANNER'))")
echo "CHANNEL_ID=$CHANNEL_ID"

echo
echo "========== 3. CREATE ZONE (draft) =========="
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/ad-portal/geofences" \
  -H "$ADV_AUTH" -H 'Content-Type: application/json' \
  -d '{"latitude":9.022736,"longitude":38.746799,"radius_metres":150}' \
  | tee "$OUT/06_zone.json"
ZONE_ID=$(python3 -c "import json; print(json.loads('''$(json_body "$OUT/06_zone.json")''')['id'])")
ZONE_ACTIVE=$(python3 -c "import json; print(json.loads('''$(json_body "$OUT/06_zone.json")''')['is_active'])")
echo "ZONE_ID=$ZONE_ID is_active=$ZONE_ACTIVE HTTP: $(http_code "$OUT/06_zone.json")"

echo
echo "========== 4. CREATE CAMPAIGN =========="
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/ad-portal/campaigns" \
  -H "$ADV_AUTH" -H 'Content-Type: application/json' \
  -d "{\"name\":\"Kaldi Macchiato E2E ${TS}\",\"target_intent\":\"fashion_interest\",\"channel_id\":\"$CHANNEL_ID\",\"title\":\"Get 20% off Macchiato today!\",\"body_text\":\"Walk into Kaldi and show this offer.\",\"image_url\":\"https://example.com/macchiato.jpg\",\"destination_url\":\"https://example.com/kaldi-offer\",\"daily_budget_cap\":500,\"total_budget_cap\":5000,\"frequency_cap_per_day\":3}" \
  | tee "$OUT/07_campaign.json"
CAMP_BODY=$(json_body "$OUT/07_campaign.json")
CAMPAIGN_ID=$(python3 -c "import json; print(json.loads('''$CAMP_BODY''')['id'])")
echo "CAMPAIGN_ID=$CAMPAIGN_ID HTTP: $(http_code "$OUT/07_campaign.json")"
python3 -c "import json; d=json.loads('''$CAMP_BODY'''); print('moderation=',d.get('moderation_status'),'validation=',d.get('validation_status'),'active=',d.get('is_active'))"

echo
echo "========== 5. LINK ZONE TO CAMPAIGN =========="
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/ad-portal/campaigns/$CAMPAIGN_ID/geofences" \
  -H "$ADV_AUTH" -H 'Content-Type: application/json' \
  -d "{\"zone_ids\":[\"$ZONE_ID\"]}" | tee "$OUT/08_link.json"
echo "link HTTP: $(http_code "$OUT/08_link.json")"

echo
echo "========== 6. ADMIN LOGIN =========="
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/ad-portal/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASS}\"}" \
  | tee "$OUT/09_admin_login.json"
ADMIN_TOKEN=$(python3 -c "import json; print(json.loads('''$(json_body "$OUT/09_admin_login.json")''')['token'])")
ADMIN_AUTH="Authorization: Bearer $ADMIN_TOKEN"

echo
echo "========== 7. ADMIN PENDING LISTS =========="
curl -s -H "$ADMIN_AUTH" "$BASE/ad-portal/admin/campaigns/pending" | tee "$OUT/10_pending_campaigns.json"
curl -s -H "$ADMIN_AUTH" "$BASE/ad-portal/admin/geofences/pending" | tee "$OUT/11_pending_zones.json"
python3 -c "import json; d=json.load(open('$OUT/10_pending_campaigns.json')); print('pending campaigns:', d.get('count'))"
python3 -c "import json; d=json.load(open('$OUT/11_pending_zones.json')); print('pending zones:', d.get('count'))"

echo
echo "========== 8. ADMIN APPROVE CAMPAIGN =========="
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/ad-portal/admin/campaigns/$CAMPAIGN_ID/validate" \
  -H "$ADMIN_AUTH" -H 'Content-Type: application/json' \
  -d '{"action":"approve","notes":"E2E geofence test approve"}' \
  | tee "$OUT/12_approve_campaign.json"
APPROVE_BODY=$(json_body "$OUT/12_approve_campaign.json")
echo "approve HTTP: $(http_code "$OUT/12_approve_campaign.json")"
python3 -c "import json; d=json.loads('''$APPROVE_BODY'''); print('moderation=',d.get('moderation_status'),'validation=',d.get('validation_status'),'active=',d.get('is_active'))"

echo
echo "========== 9. ADMIN ACTIVATE ZONE (idempotent) =========="
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/ad-portal/admin/geofences/$ZONE_ID/activate" \
  -H "$ADMIN_AUTH" | tee "$OUT/13_activate_zone.json"
ZONE_ACT_BODY=$(json_body "$OUT/13_activate_zone.json")
echo "activate zone HTTP: $(http_code "$OUT/13_activate_zone.json")"
python3 -c "import json; d=json.loads('''$ZONE_ACT_BODY'''); print('zone is_active=', d.get('is_active'))"

echo
echo "========== 10. SDK APP KEYS =========="
DEV_EMAIL="geo-dev-e2e-${TS}@example.com"
curl -s -X POST "$BASE/portal/register" -H 'Content-Type: application/json' \
  -d "{\"name\":\"Geo Dev E2E\",\"email\":\"${DEV_EMAIL}\",\"password\":\"securepass123\"}" > "$OUT/14_dev_register.json" || true
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/portal/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"${DEV_EMAIL}\",\"password\":\"securepass123\"}" \
  | tee "$OUT/15_dev_login.json"
DEV_TOKEN=$(python3 -c "import json; print(json.loads('''$(json_body "$OUT/15_dev_login.json")''')['token'])")
DEV_AUTH="Authorization: Bearer $DEV_TOKEN"

curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/portal/applications" \
  -H "$DEV_AUTH" -H 'Content-Type: application/json' \
  -d '{"app_name":"Geofence E2E App","platform":"flutter","bundle_id":"com.skykin.geofence.e2e"}' \
  | tee "$OUT/16_app.json"
APP_BODY=$(json_body "$OUT/16_app.json")
PK=$(python3 -c "import json; print(json.loads('''$APP_BODY''')['publishable_key'])")
SK=$(python3 -c "import json; print(json.loads('''$APP_BODY''')['secret_key'])")
echo "PK=$PK"

echo
echo "========== 11. DEMO PSEUDONYMOUS ID =========="
PSEUDO=$(docker compose -f "/home/sifen/Desktop/my git/SDK/docker-compose.yml" exec -T db \
  psql -U skykin_user -d skykin_db -tAc \
  "SELECT pseudonymous_id::text FROM demo_sms_recipients WHERE pseudonymous_id IS NOT NULL LIMIT 1;" | tr -d '[:space:]')
echo "PSEUDO=$PSEUDO" | tee "$OUT/17_pseudo.txt"

echo
echo "========== 12. LOCATION CONSENT =========="
CONSENT_BODY="{\"pseudonymous_id\":\"$PSEUDO\",\"location_ad_consent\":true}"
CONSENT_SIG=$(hmac_sig "$SK" "$CONSENT_BODY")
curl -s -w "\nHTTP:%{http_code}\n" -X PATCH "$BASE/geofences/location-consent" \
  -H "X-API-Key: $PK" -H "X-Signature: $CONSENT_SIG" -H 'Content-Type: application/json' \
  -d "$CONSENT_BODY" | tee "$OUT/18_consent.json"
echo "consent HTTP: $(http_code "$OUT/18_consent.json")"

echo
echo "========== 13. SYNC NEARBY ZONES =========="
curl -s -w "\nHTTP:%{http_code}\n" -H "X-API-Key: $PK" \
  "$BASE/geofences/sync?lat=9.022736&lng=38.746799&radius_m=20000" \
  | tee "$OUT/19_sync.json"
echo "sync HTTP: $(http_code "$OUT/19_sync.json")"
python3 -c "import json; d=json.loads('''$(json_body "$OUT/19_sync.json")'''); zs=d.get('zones',[]); print('zones in sync:', len(zs)); print([(z['id'], z['is_active']) for z in zs if z['id']=='$ZONE_ID'])"

echo
echo "========== 14. GEOFENCE ENTER EVENT =========="
EVENT_BODY="{\"pseudonymous_id\":\"$PSEUDO\",\"zone_id\":\"$ZONE_ID\",\"accuracy_m\":12.5}"
EVENT_SIG=$(hmac_sig "$SK" "$EVENT_BODY")
curl -s -w "\nHTTP:%{http_code}\n" -X POST "$BASE/geofence/event" \
  -H "X-API-Key: $PK" -H "X-Signature: $EVENT_SIG" -H 'Content-Type: application/json' \
  -d "$EVENT_BODY" | tee "$OUT/20_event.json"
echo "event HTTP: $(http_code "$OUT/20_event.json")"
python3 -m json.tool <(json_body "$OUT/20_event.json") 2>/dev/null | head -40

echo
echo "========== 15. SAVE IDS =========="
cat > "$OUT/ids.txt" <<EOF
ADV_EMAIL=$ADV_EMAIL
ZONE_ID=$ZONE_ID
CAMPAIGN_ID=$CAMPAIGN_ID
CHANNEL_ID=$CHANNEL_ID
PSEUDO=$PSEUDO
PK=$PK
EOF

echo
echo "========== 16. DB SNAPSHOT =========="
docker compose -f "/home/sifen/Desktop/my git/SDK/docker-compose.yml" exec -T db \
  psql -U skykin_user -d skykin_db -x -c \
  "SELECT id, advertiser_id, latitude, longitude, radius_metres, is_active, created_at FROM geofence_zones WHERE id = '$ZONE_ID';"

docker compose -f "/home/sifen/Desktop/my git/SDK/docker-compose.yml" exec -T db \
  psql -U skykin_user -d skykin_db -x -c \
  "SELECT * FROM campaign_geofence_targets WHERE campaign_id = '$CAMPAIGN_ID';"

docker compose -f "/home/sifen/Desktop/my git/SDK/docker-compose.yml" exec -T db \
  psql -U skykin_user -d skykin_db -x -c \
  "SELECT id, name, is_active, moderation_status, validation_status, target_intent FROM campaigns WHERE id = '$CAMPAIGN_ID';"

docker compose -f "/home/sifen/Desktop/my git/SDK/docker-compose.yml" exec -T db \
  psql -U skykin_user -d skykin_db -x -c \
  "SELECT id, zone_id, pseudonymous_id, accuracy_m, visited_at FROM store_visits WHERE zone_id = '$ZONE_ID' ORDER BY visited_at DESC LIMIT 3;"

docker compose -f "/home/sifen/Desktop/my git/SDK/docker-compose.yml" exec -T db \
  psql -U skykin_user -d skykin_db -tAc \
  "SELECT location_ad_consent FROM demo_sms_recipients WHERE pseudonymous_id = '$PSEUDO'::uuid;"

echo
echo "E2E COMPLETE"
