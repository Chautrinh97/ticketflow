#!/usr/bin/env bash
# End-to-end Phase 1 smoke test, driven entirely through the API Gateway:
# login -> create+publish event -> add ticket types -> multi-item booking ->
# mock payment -> confirm paid with tickets issued. See the implementation
# plan's verification section.
#
# Usage: ./scripts/smoke-test.sh
# Requires: curl, jq, the full stack running (docker compose up) and
# scripts/seed-dev-users.sh already run (uses organizer@ticketflow.dev).
set -euo pipefail

GATEWAY="${GATEWAY_URL:-http://localhost:8080/api/v1}"

mock_token() {
  local email="$1" name="$2"
  printf '{"uid":"mock-%s","email":"%s","name":"%s"}' "$email" "$email" "$name" | base64 -w0 2>/dev/null \
    || printf '{"uid":"mock-%s","email":"%s","name":"%s"}' "$email" "$email" "$name" | base64
}

echo "== Login as organizer =="
ORGANIZER_TOKEN=$(mock_token "organizer@ticketflow.dev" "Demo Organizer")
LOGIN_RESP=$(curl -sS -X POST "$GATEWAY/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"firebase_id_token\":\"$ORGANIZER_TOKEN\"}")
ACCESS_TOKEN=$(echo "$LOGIN_RESP" | jq -r .access_token)
[ "$ACCESS_TOKEN" != "null" ] || { echo "login failed: $LOGIN_RESP"; exit 1; }
echo "OK"

echo "== Create event =="
START_TIME=$(date -u -d '+7 days' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v+7d +%Y-%m-%dT%H:%M:%SZ)
EVENT_RESP=$(curl -sS -X POST "$GATEWAY/organizer/events" \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"title\":\"Smoke Test Concert\",\"category\":\"concert\",\"start_time\":\"$START_TIME\"}")
EVENT_ID=$(echo "$EVENT_RESP" | jq -r .id)
[ "$EVENT_ID" != "null" ] || { echo "create event failed: $EVENT_RESP"; exit 1; }
echo "OK ($EVENT_ID)"

echo "== Add ticket types (VIP + Standard) =="
VIP_RESP=$(curl -sS -X POST "$GATEWAY/organizer/events/$EVENT_ID/ticket-types" \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"VIP","price":500000,"quota":5}')
VIP_ID=$(echo "$VIP_RESP" | jq -r .id)
STD_RESP=$(curl -sS -X POST "$GATEWAY/organizer/events/$EVENT_ID/ticket-types" \
  -H "Authorization: Bearer $ACCESS_TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Standard","price":200000,"quota":10}')
STD_ID=$(echo "$STD_RESP" | jq -r .id)
[ "$VIP_ID" != "null" ] && [ "$STD_ID" != "null" ] || { echo "add ticket types failed"; exit 1; }
echo "OK (VIP=$VIP_ID, Standard=$STD_ID)"

echo "== Publish event =="
curl -sS -X POST "$GATEWAY/organizer/events/$EVENT_ID/publish" -H "Authorization: Bearer $ACCESS_TOKEN" >/dev/null
echo "OK"

echo "== Login as buyer =="
BUYER_TOKEN=$(mock_token "buyer@ticketflow.dev" "Demo Buyer")
BUYER_LOGIN=$(curl -sS -X POST "$GATEWAY/auth/login" -H 'Content-Type: application/json' \
  -d "{\"firebase_id_token\":\"$BUYER_TOKEN\"}")
BUYER_ACCESS=$(echo "$BUYER_LOGIN" | jq -r .access_token)
[ "$BUYER_ACCESS" != "null" ] || { echo "buyer login failed: $BUYER_LOGIN"; exit 1; }
echo "OK"

echo "== Book 2 VIP + 3 Standard in one multi-item order =="
ORDER_RESP=$(curl -sS -X POST "$GATEWAY/bookings" \
  -H "Authorization: Bearer $BUYER_ACCESS" -H 'Content-Type: application/json' \
  -d "{\"items\":[{\"ticket_type_id\":\"$VIP_ID\",\"quantity\":2},{\"ticket_type_id\":\"$STD_ID\",\"quantity\":3}]}")
ORDER_ID=$(echo "$ORDER_RESP" | jq -r .id)
[ "$ORDER_ID" != "null" ] || { echo "booking failed: $ORDER_RESP"; exit 1; }
echo "OK ($ORDER_ID), status=$(echo "$ORDER_RESP" | jq -r .status)"

echo "== Mock checkout (Phase 1: confirms payment synchronously) =="
curl -sS -X POST "$GATEWAY/payments/$ORDER_ID/checkout" -H "Authorization: Bearer $BUYER_ACCESS" >/dev/null
echo "OK"

echo "== Confirm order is paid with tickets issued =="
FINAL_ORDER=$(curl -sS "$GATEWAY/bookings/$ORDER_ID" -H "Authorization: Bearer $BUYER_ACCESS")
STATUS=$(echo "$FINAL_ORDER" | jq -r .status)
TICKET_COUNT=$(echo "$FINAL_ORDER" | jq '.tickets | length')
if [ "$STATUS" = "paid" ] && [ "$TICKET_COUNT" = "5" ]; then
  echo "SUCCESS: order paid with $TICKET_COUNT tickets issued"
else
  echo "FAILED: status=$STATUS tickets=$TICKET_COUNT"
  echo "$FINAL_ORDER"
  exit 1
fi
