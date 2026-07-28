#!/bin/bash

# Test script for telemetry endpoints
# Make sure your API is running and you have valid credentials

API_URL="http://localhost:8080"
API_KEY="your-api-key"
SDK_SECRET="your-sdk-secret"

# Get a valid campaign ID from your database
# You'll need to update this with a real campaign ID
CAMPAIGN_ID="550e8400-e29b-41d4-a716-446655440000"

echo "=== Testing /telemetry/track-anonymous endpoint ==="
echo ""

# Test 1: Anonymous impression
echo "Test 1: Posting anonymous impression..."
curl -X POST "$API_URL/telemetry/track-anonymous" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -H "X-SDK-Secret: $SDK_SECRET" \
  -d "{
    \"campaign_id\": \"$CAMPAIGN_ID\",
    \"event_type\": \"impression\"
  }" \
  -w "\nHTTP Status: %{http_code}\n\n"

echo "Waiting 3 seconds for consumer to process..."
sleep 3

echo "=== Testing /telemetry/track endpoint ==="
echo ""

# Test 2: Consented track (requires pseudonymous_id)
echo "Test 2: Posting consented click..."
curl -X POST "$API_URL/telemetry/track" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -H "X-SDK-Secret: $SDK_SECRET" \
  -d "{
    \"campaign_id\": \"$CAMPAIGN_ID\",
    \"event_type\": \"click\",
    \"pseudonymous_id\": \"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d\"
  }" \
  -w "\nHTTP Status: %{http_code}\n\n"

echo "Waiting 3 seconds for consumer to process..."
sleep 3

echo "=== Database verification queries ==="
echo ""
echo "Check billing_events table:"
echo "SELECT COUNT(*), event_type FROM billing_events WHERE campaign_id = '$CAMPAIGN_ID' GROUP BY event_type;"
echo ""
echo "Check campaign_delivery_logs table:"
echo "SELECT COUNT(*), delivery_status FROM campaign_delivery_logs WHERE campaign_id = '$CAMPAIGN_ID' GROUP BY delivery_status;"
echo ""
echo "To verify, run these queries in your PostgreSQL database:"
echo "psql -U your_user -d your_db -c 'SELECT id, campaign_id, event_type, charge_etb, occurred_at FROM billing_events WHERE campaign_id = '"'"'$CAMPAIGN_ID'"'"' ORDER BY created_at DESC LIMIT 5;'"
echo "psql -U your_user -d your_db -c 'SELECT id, campaign_id, user_id, delivery_status, logged_at FROM campaign_delivery_logs WHERE campaign_id = '"'"'$CAMPAIGN_ID'"'"' ORDER BY created_at DESC LIMIT 5;'"
