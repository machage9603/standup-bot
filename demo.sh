#!/bin/bash

# Telex.im AI Agent Demo Script
# This script demonstrates all agent capabilities

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
TELEX_API_KEY="${TELEX_API_KEY:-test_key_123}"

echo -e "${BLUE}╔═══════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Telex.im AI Agent Demo & Testing Suite      ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════╝${NC}"
echo ""

# Function to make requests and display results
function test_endpoint() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local auth=$5
    
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}Testing: ${name}${NC}"
    echo -e "${BLUE}Endpoint: ${method} ${endpoint}${NC}"
    echo ""
    
    if [ -z "$data" ]; then
        if [ -z "$auth" ]; then
            response=$(curl -s -X ${method} "${BASE_URL}${endpoint}")
        else
            response=$(curl -s -X ${method} \
                -H "Authorization: Bearer ${TELEX_API_KEY}" \
                "${BASE_URL}${endpoint}")
        fi
    else
        if [ -z "$auth" ]; then
            response=$(curl -s -X ${method} \
                -H "Content-Type: application/json" \
                -d "${data}" \
                "${BASE_URL}${endpoint}")
        else
            response=$(curl -s -X ${method} \
                -H "Content-Type: application/json" \
                -H "Authorization: Bearer ${TELEX_API_KEY}" \
                -d "${data}" \
                "${BASE_URL}${endpoint}")
        fi
    fi
    
    echo -e "${GREEN}Response:${NC}"
    echo "$response" | jq . 2>/dev/null || echo "$response"
    echo ""
    
    # Check if successful
    if echo "$response" | jq -e '.status == "success" or .status == "healthy" or .reply' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Test passed${NC}"
    else
        echo -e "${RED}✗ Test may have issues${NC}"
    fi
    echo ""
}

# Wait for service to be ready
echo -e "${YELLOW}Checking if service is running...${NC}"
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Service is ready!${NC}"
        echo ""
        break
    fi
    
    attempt=$((attempt + 1))
    if [ $attempt -eq $max_attempts ]; then
        echo -e "${RED}✗ Service not responding after ${max_attempts} attempts${NC}"
        echo -e "${YELLOW}Make sure the agent is running with: go run main.go${NC}"
        exit 1
    fi
    
    echo -n "."
    sleep 1
done

# Test 1: Health Check
test_endpoint \
    "Health Check" \
    "GET" \
    "/health"

sleep 1

# Test 2: Agent Info
test_endpoint \
    "Agent Information" \
    "GET" \
    "/api/agent/info"

sleep 1

# Test 3: Direct Message - Greeting
test_endpoint \
    "Direct Message - Greeting" \
    "POST" \
    "/api/message" \
    '{"userId": "demo_user_001", "message": "Hello! How are you today?"}'

sleep 2

# Test 4: Direct Message - Question
test_endpoint \
    "Direct Message - Question" \
    "POST" \
    "/api/message" \
    '{"userId": "demo_user_001", "message": "What is Go programming language?"}'

sleep 2

# Test 5: Direct Message - Help Request
test_endpoint \
    "Direct Message - Help Request" \
    "POST" \
    "/api/message" \
    '{"userId": "demo_user_002", "message": "Can you help me understand AI agents?"}'

sleep 2

# Test 6: Conversation History
test_endpoint \
    "Conversation History" \
    "GET" \
    "/api/conversations/demo_user_001"

sleep 1

# Test 7: Metrics
test_endpoint \
    "System Metrics" \
    "GET" \
    "/api/metrics"

sleep 1

# Test 8: Webhook - Message Received
test_endpoint \
    "Webhook - Message Received" \
    "POST" \
    "/webhook/telex" \
    '{
        "event": "message.received",
        "message": {
            "id": "msg_demo_001",
            "from": "webhook_user_001",
            "to": "ai-agent-001",
            "content": "This is a test message from Telex.im webhook",
            "timestamp": "2025-11-05T10:30:00Z"
        }
    }' \
    "auth"

sleep 2

# Test 9: Webhook - User Joined
test_endpoint \
    "Webhook - User Joined" \
    "POST" \
    "/webhook/telex" \
    '{
        "event": "user.joined",
        "user": {
            "id": "new_user_001",
            "username": "newuser",
            "name": "New User"
        }
    }' \
    "auth"

sleep 1

# Test 10: Intent Recognition Tests
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Testing Intent Recognition${NC}"
echo ""

intents=(
    '{"userId": "intent_test", "message": "Can you help me?"}'
    '{"userId": "intent_test", "message": "What is the weather?"}'
    '{"userId": "intent_test", "message": "Thank you so much!"}'
    '{"userId": "intent_test", "message": "Hi there!"}'
)

for intent_data in "${intents[@]}"; do
    message=$(echo $intent_data | jq -r '.message')
    echo -e "${BLUE}Testing message: ${message}${NC}"
    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "${intent_data}" \
        "${BASE_URL}/api/message")
    
    detected_intent=$(echo "$response" | jq -r '.intent')
    echo -e "${GREEN}Detected intent: ${detected_intent}${NC}"
    echo ""
    sleep 1
done

# Test 11: Entity Extraction
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Testing Entity Extraction${NC}"
echo ""

test_endpoint \
    "Entity Extraction - Mentions and Hashtags" \
    "POST" \
    "/api/message" \
    '{"userId": "entity_test", "message": "Hey @john check out #golang and tell @mary"}'

sleep 1

# Test 12: Context Continuity
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Testing Context Continuity${NC}"
echo ""

echo -e "${BLUE}Message 1: Establishing context${NC}"
curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"userId": "context_test", "message": "My name is Alice"}' \
    "${BASE_URL}/api/message" | jq .
echo ""
sleep 2

echo -e "${BLUE}Message 2: Testing context memory${NC}"
curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"userId": "context_test", "message": "What is my name?"}' \
    "${BASE_URL}/api/message" | jq .
echo ""
sleep 2

echo -e "${BLUE}Checking conversation history${NC}"
curl -s "${BASE_URL}/api/conversations/context_test" | jq .
echo ""

# Test 13: Error Handling
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Testing Error Handling${NC}"
echo ""

echo -e "${BLUE}Test: Missing required field${NC}"
response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"userId": "error_test"}' \
    "${BASE_URL}/api/message")
echo "$response" | jq .
echo ""

echo -e "${BLUE}Test: Unauthorized webhook${NC}"
response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer wrong_key" \
    -d '{"event": "message.received", "message": {"from": "test", "content": "test"}}' \
    "${BASE_URL}/webhook/telex")
echo "$response" | jq .
echo ""

echo -e "${BLUE}Test: Non-existent conversation${NC}"
response=$(curl -s "${BASE_URL}/api/conversations/nonexistent_user")
echo "$response" | jq .
echo ""

# Final Summary
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Test Summary${NC}"
echo ""

# Get final metrics
echo -e "${BLUE}Final System Metrics:${NC}"
curl -s "${BASE_URL}/api/metrics" | jq .
echo ""

echo -e "${GREEN}✓ Demo completed successfully!${NC}"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Check the logs for detailed processing information"
echo "2. Try the endpoints with your own test data"
echo "3. Set up actual Telex.im webhook integration"
echo "4. Monitor metrics at ${BASE_URL}/api/metrics"
echo ""
echo -e "${BLUE}For more information, see README.md and API_DOCUMENTATION.md${NC}"
echo ""