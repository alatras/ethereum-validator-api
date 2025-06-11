#!/bin/bash

# ETH-API Testing Examples
# Run the API server first: ./eth-api

BASE_URL="http://localhost:8080"

echo "=== ETH-API Testing Examples ==="
echo ""

# Block Reward Endpoints
echo "1. Get Block Reward - Valid Slot"
curl -X GET "$BASE_URL/blockreward/7847950" | jq .
echo -e "\n"

echo "2. Get Block Reward - Invalid Slot Format"
curl -X GET "$BASE_URL/blockreward/invalid" | jq .
echo -e "\n"

echo "3. Get Block Reward - Future Slot"
curl -X GET "$BASE_URL/blockreward/99999999" | jq .
echo -e "\n"

# Sync Duties Endpoints
echo "4. Get Sync Duties - Valid Slot"
curl -X GET "$BASE_URL/syncduties/7847950" | jq .
echo -e "\n"

echo "5. Get Sync Duties - Invalid Slot Format"
curl -X GET "$BASE_URL/syncduties/abc" | jq .
echo -e "\n"

echo "6. Get Sync Duties - Slot Too Far in Future"
curl -X GET "$BASE_URL/syncduties/99999999" | jq .
echo -e "\n"

# Test with custom port if provided
if [ -n "$1" ]; then
    echo "Testing with custom port: $1"
    curl -X GET "http://localhost:$1/blockreward/7847950" | jq .
fi
