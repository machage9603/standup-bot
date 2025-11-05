# Telex.im AI Agent - API Documentation

## Base URL
```
http://localhost:8080
```

For production, replace with your deployed URL.

---

## Authentication

### Webhook Endpoints
Webhook endpoints require Bearer token authentication:
```
Authorization: Bearer YOUR_TELEX_API_KEY
```

### Public Endpoints
Health check and info endpoints are public and don't require authentication.

---

## Endpoints

### 1. Health Check

Check if the agent is running and healthy.

**Endpoint:** `GET /health`

**Authentication:** None required

**Response:**
```json
{
  "status": "healthy",
  "service": "telex-ai-agent",
  "timestamp": "2025-11-05T10:30:00Z",
  "version": "1.0.0"
}
```

**Status Codes:**
- `200 OK` - Service is healthy

**Example:**
```bash
curl http://localhost:8080/health
```

---

### 2. Telex.im Webhook Handler

Receives and processes events from Telex.im platform.

**Endpoint:** `POST /webhook/telex`

**Authentication:** Required (Bearer token)

**Headers:**
```
Content-Type: application/json
Authorization: Bearer YOUR_TELEX_API_KEY
```

**Request Body:**
```json
{
  "event": "message.received",
  "message": {
    "id": "msg_123456",
    "from": "user_789",
    "to": "ai-agent-001",
    "content": "Hello, AI assistant!",
    "timestamp": "2025-11-05T10:30:00Z",
    "type": "text"
  }
}
```

**Supported Events:**

#### message.received
User sends a message to the agent.

```json
{
  "event": "message.received",
  "message": {
    "id": "string",
    "from": "string",
    "to": "string",
    "content": "string",
    "timestamp": "ISO 8601 datetime",
    "type": "text"
  }
}
```

#### user.joined
New user joins the conversation.

```json
{
  "event": "user.joined",
  "user": {
    "id": "string",
    "username": "string",
    "name": "string"
  }
}
```

#### user.typing
User is typing (acknowledged but not processed).

```json
{
  "event": "user.typing",
  "message": {
    "from": "string"
  }
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Message processed"
}
```

**Status Codes:**
- `200 OK` - Event processed successfully
- `401 Unauthorized` - Invalid or missing authentication
- `400 Bad Request` - Invalid payload
- `500 Internal Server Error` - Processing error

**Example:**
```bash
curl -X POST http://localhost:8080/webhook/telex \
  -H "Authorization: Bearer YOUR_TELEX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message.received",
    "message": {
      "id": "msg_001",
      "from": "user_123",
      "to": "ai-agent-001",
      "content": "What is the weather today?"
    }
  }'
```

---

### 3. Direct Message (Testing)

Send a message directly to the agent for testing purposes.

**Endpoint:** `POST /api/message`

**Authentication:** None required (for testing only)

**Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "userId": "user_123",
  "message": "Hello, how can you help me?"
}
```

**Response:**
```json
{
  "reply": "Hello! I'm an AI assistant here to help you. I can answer questions, provide information, help with tasks, and have conversations. What would you like assistance with today?",
  "intent": "greeting",
  "entities": {},
  "confidence": 0.85
}
```

**Fields:**
- `reply` (string) - The AI's response
- `intent` (string) - Detected intent (greeting, question, help_request, gratitude, general)
- `entities` (object) - Extracted entities (mentions, hashtags, etc.)
- `confidence` (float) - Confidence score (0.0 - 1.0)

**Status Codes:**
- `200 OK` - Message processed successfully
- `400 Bad Request` - Missing required fields
- `500 Internal Server Error` - AI processing error

**Example:**
```bash
curl -X POST http://localhost:8080/api/message \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "test_user_456",
    "message": "Can you help me learn about Go programming?"
  }'
```

---

### 4. Agent Information

Get information about the agent's configuration and status.

**Endpoint:** `GET /api/agent/info`

**Authentication:** None required

**Response:**
```json
{
  "agentId": "ai-agent-001",
  "name": "Telex AI Assistant",
  "version": "1.0.0",
  "capabilities": [
    "conversation",
    "context-awareness",
    "multi-turn-dialogue"
  ],
  "status": "online",
  "uptime": "2025-11-05T10:30:00Z"
}
```

**Status Codes:**
- `200 OK` - Information retrieved successfully

**Example:**
```bash
curl http://localhost:8080/api/agent/info
```

---

### 5. Conversation History

Retrieve conversation history for a specific user.

**Endpoint:** `GET /api/conversations/:userId`

**Authentication:** None required

**Parameters:**
- `userId` (path parameter) - The user ID to retrieve history for

**Response:**
```json
{
  "userId": "user_123",
  "messageCount": 15,
  "topics": ["programming", "weather", "support"],
  "lastMessage": "Thanks for your help!",
  "timestamp": "2025-11-05T10:25:00Z"
}
```

**Fields:**
- `userId` (string) - User identifier
- `messageCount` (integer) - Total messages in conversation
- `topics` (array) - Topics discussed
- `lastMessage` (string) - Most recent message content
- `timestamp` (datetime) - Last activity timestamp

**Status Codes:**
- `200 OK` - History retrieved successfully
- `404 Not Found` - No conversation found for user

**Example:**
```bash
curl http://localhost:8080/api/conversations/user_123
```

---

### 6. Metrics

Get system metrics and statistics.

**Endpoint:** `GET /api/metrics`

**Authentication:** None required

**Response:**
```json
{
  "totalConversations": 42,
  "totalMessages": 156,
  "activeUsers": 42,
  "timestamp": "2025-11-05T10:30:00Z"
}
```

**Fields:**
- `totalConversations` (integer) - Total unique conversations
- `totalMessages` (integer) - Total messages processed
- `activeUsers` (integer) - Number of users with active conversations
- `timestamp` (datetime) - Metrics snapshot time

**Status Codes:**
- `200 OK` - Metrics retrieved successfully

**Example:**
```bash
curl http://localhost:8080/api/metrics
```

---

## Data Models

### TelexMessage
```typescript
{
  id: string;           // Unique message ID
  from: string;         // Sender user ID
  to: string;           // Recipient ID (agent)
  content: string;      // Message text
  timestamp: DateTime;  // ISO 8601 format
  type?: string;        // Message type (default: "text")
}
```

### TelexWebhook
```typescript
{
  event: string;        // Event type
  message?: TelexMessage;
  user?: TelexUser;
}
```

### TelexUser
```typescript
{
  id: string;           // User ID
  username: string;     // Username
  name?: string;        // Display name
}
```

### AgentResponse
```typescript
{
  reply: string;        // AI response text
  intent?: string;      // Detected intent
  entities?: {          // Extracted entities
    [key: string]: string;
  };
  confidence?: number;  // Confidence score (0-1)
}
```

### ConversationContext
```typescript
{
  userId: string;       // User identifier
  lastMessage: string;  // Last message content
  messageCount: number; // Total messages
  topics: string[];     // Discussion topics
  timestamp: DateTime;  // Last activity
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "status": "error",
  "message": "Description of the error"
}
```

Or for API endpoints:

```json
{
  "error": "Detailed error message"
}
```

### Common Error Codes

| Status Code | Meaning |
|-------------|---------|
| 400 | Bad Request - Invalid or missing parameters |
| 401 | Unauthorized - Authentication failed |
| 404 | Not Found - Resource doesn't exist |
| 500 | Internal Server Error - Server-side error |

---

## Rate Limits

Currently, no rate limits are enforced on the agent side. However:

- **Claude AI**: Subject to Anthropic's API rate limits
- **Telex.im**: Subject to platform rate limits

**Recommendations:**
- Don't send more than 100 requests/second
- Implement exponential backoff for failed requests
- Monitor response times

---

## Webhook Setup with Telex.im

### 1. Register Your Webhook

```bash
curl -X POST https://api.telex.im/v1/webhooks \
  -H "Authorization: Bearer YOUR_TELEX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-domain.com/webhook/telex",
    "events": ["message.received", "user.joined"]
  }'
```

### 2. Verify Webhook

```bash
curl -X GET https://api.telex.im/v1/webhooks \
  -H "Authorization: Bearer YOUR_TELEX_API_KEY"
```

### 3. Test Webhook

Send a test event from Telex.im dashboard or API.

---

## Best Practices

### 1. Error Handling
Always wrap API calls in try-catch blocks and handle errors gracefully.

### 2. Timeouts
Set appropriate timeouts for all HTTP requests (recommended: 30s).

### 3. Logging
Log all webhook events and errors for debugging.

### 4. Security
- Never expose API keys in code or logs
- Use HTTPS in production
- Validate all incoming webhook requests
- Implement rate limiting if needed

### 5. Testing
- Test webhook authentication
- Test all event types
- Test error scenarios
- Load test before production

---

## Examples

### Complete Conversation Flow

```bash
# 1. User sends message via Telex.im
# Telex.im sends webhook to agent

# 2. Agent receives and processes
curl -X POST http://localhost:8080/webhook/telex \
  -H "Authorization: Bearer TELEX_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message.received",
    "message": {
      "from": "user_123",
      "content": "What is Go programming?"
    }
  }'

# 3. Agent generates response via Claude AI
# 4. Agent sends response to Telex.im
# 5. User receives message

# Check conversation history
curl http://localhost:8080/api/conversations/user_123
```

### Testing Direct Message

```bash
# Send a test message
curl -X POST http://localhost:8080/api/message \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "test_user",
    "message": "Tell me a joke"
  }' | jq .

# Expected response:
# {
#   "reply": "Why do programmers prefer dark mode?...",
#   "intent": "general",
#   "entities": {},
#   "confidence": 0.85
# }
```

---

## Support

For issues or questions:
- GitHub Issues: [github.com/yourusername/telex-ai-agent/issues](https://github.com/yourusername/telex-ai-agent/issues)
- Email: support@yourproject.com
- Documentation: Full README at repository

---

## Changelog

### Version 1.0.0 (2025-11-05)
- Initial release
- Full Telex.im webhook integration
- Claude AI powered responses
- Context-aware conversations
- REST API endpoints
- Docker support

---

