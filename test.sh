#!/bin/bash

# =============================================
# Тестовая сюита MVP Posiflora Telegram
# Формат номеров заказов: A-XXXX (X = 0–9)
# Конфигурация берется из .env
# =============================================

set -e

# ------------------------------------------------
# Load .env
# ------------------------------------------------
if [ ! -f .env ]; then
  echo "❌ .env file not found"
  exit 1
fi

export $(grep -v '^#' .env | xargs)

if [ -z "$TELEGRAM_BOT_TOKEN" ] || [ -z "$TELEGRAM_CHAT_ID" ]; then
  echo "❌ TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID is missing in .env"
  exit 1
fi

# ------------------------------------------------
# Config
# ------------------------------------------------
SHOP_ID=1
PORT="${PORT:-8080}"
BASE_URL="http://localhost:${PORT}"

echo "============================================="
echo " POSIFLORA TELEGRAM E2E TEST (A-XXXX)"
echo "============================================="
echo "Shop ID : $SHOP_ID"
echo "Base URL: $BASE_URL"
echo "Chat ID : ****${TELEGRAM_CHAT_ID: -4}"
echo "============================================="

# ------------------------------------------------
# Utils
# ------------------------------------------------
generate_order_number() {
  echo "A-$(shuf -i 0-9 -n 4 | tr -d '\n')"
}

# ------------------------------------------------
# Generate order numbers
# ------------------------------------------------
ORDER_NUMBERS=()
for i in {1..5}; do
  ORDER_NUMBERS+=("$(generate_order_number)")
done

echo "Generated orders:"
for o in "${ORDER_NUMBERS[@]}"; do
  echo " - $o"
done

# ------------------------------------------------
# Wait for server
# ------------------------------------------------
echo -e "\n⏳ Waiting for server..."
for i in {1..30}; do
  if curl -s "$BASE_URL/shops/$SHOP_ID/growth/telegram" > /dev/null; then
    echo "✅ Server ready"
    break
  fi
  sleep 1
done

# ------------------------------------------------
# Connect Telegram integration
# ------------------------------------------------
echo -e "\n=== CONNECT TELEGRAM ==="

HTTP_CODE=$(curl -s -o response.json -w "%{http_code}" -X POST "$BASE_URL/shops/$SHOP_ID/telegram/connect" \
  -H "Content-Type: application/json" \
  -d "{
    \"botToken\": \"${TELEGRAM_BOT_TOKEN}\",
    \"chatId\": \"${TELEGRAM_CHAT_ID}\",
    \"enabled\": true
  }")

if [ "$HTTP_CODE" -ne 200 ]; then
  echo "❌ Telegram connect failed with HTTP code $HTTP_CODE"
  cat response.json
  exit 1
fi

cat response.json | jq

# ------------------------------------------------
# Send generated orders
# ------------------------------------------------
echo -e "\n=== SEND ORDERS ==="
for order in "${ORDER_NUMBERS[@]}"; do
  echo "=== ORDER $order ==="
  HTTP_CODE=$(curl -s -o order_response.json -w "%{http_code}" -X POST "$BASE_URL/shops/$SHOP_ID/orders" \
    -H "Content-Type: application/json" \
    -d "{
      \"number\": \"$order\",
      \"total\": 100.50,
      \"customer_name\": \"Test User\"
    }")

  if [ "$HTTP_CODE" -ne 200 ]; then
    echo "❌ Order $order failed with HTTP code $HTTP_CODE"
    cat order_response.json
    continue
  fi

  cat order_response.json | jq
done


# ------------------------------------------------
# Attempt to send orders while Telegram is disabled
# ------------------------------------------------
echo -e "\n=== SEND ORDERS WITH TELEGRAM DISABLED ==="
for order in "${ORDER_NUMBERS[@]}"; do
  echo "=== ORDER $order ==="
  HTTP_CODE=$(curl -s -o order_response_disabled.json -w "%{http_code}" -X POST "$BASE_URL/shops/$SHOP_ID/orders" \
    -H "Content-Type: application/json" \
    -d "{
      \"number\": \"$order\",
      \"total\": 100.50,
      \"customer_name\": \"Test User\"
    }")

  if [ "$HTTP_CODE" -ne 200 ]; then
    echo "❌ Order $order failed with HTTP code $HTTP_CODE (expected if Telegram is disabled)"
    cat order_response_disabled.json
    continue
  fi

  cat order_response_disabled.json | jq
done

echo -e "\n✅ E2E test finished "
