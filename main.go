package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Message structures for Telex.im
type TelexMessage struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type,omitempty"`
}

type TelexWebhook struct {
	Event   string       `json:"event"`
	Message TelexMessage `json:"message"`
	User    TelexUser    `json:"user,omitempty"`
}

type TelexUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name,omitempty"`
}

type TelexResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// AI Agent structures
type AgentResponse struct {
	Reply      string            `json:"reply"`
	Intent     string            `json:"intent,omitempty"`
	Entities   map[string]string `json:"entities,omitempty"`
	Confidence float64           `json:"confidence,omitempty"`
}

type ConversationContext struct {
	UserID       string
	LastMessage  string
	MessageCount int
	Topics       []string
	Timestamp    time.Time
}

// Agent state management
var conversationHistory = make(map[string]*ConversationContext)

// Groq AI configuration (FREE API!)
var (
	groqAPIKey   string
	telexAPIKey  string
	telexBaseURL string
	agentID      string
)

func main() {
	// Load environment variables
	godotenv.Load()

	groqAPIKey = os.Getenv("GROQ_API_KEY")
	telexAPIKey = os.Getenv("TELEX_API_KEY")
	telexBaseURL = os.Getenv("TELEX_BASE_URL")
	agentID = os.Getenv("AGENT_ID")

	if groqAPIKey == "" || telexAPIKey == "" {
		log.Fatal("Missing required API keys in environment")
	}

	if telexBaseURL == "" {
		telexBaseURL = "https://api.telex.im/v1"
	}

	if agentID == "" {
		agentID = "ai-agent-001"
	}

	// Initialize Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", healthCheck)

	// Telex.im webhook endpoint
	r.POST("/webhook/telex", handleTelexWebhook)

	// Direct message endpoint (for testing)
	r.POST("/api/message", handleDirectMessage)

	// Agent info endpoint
	r.GET("/api/agent/info", getAgentInfo)

	// Conversation history endpoint
	r.GET("/api/conversations/:userId", getConversationHistory)

	// Metrics endpoint
	r.GET("/api/metrics", getMetrics)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🤖 AI Agent starting on port %s", port)
	log.Printf("📡 Telex.im integration ready")
	log.Printf("⚡ Using Groq AI (FREE - Llama 3.3 70B)")
	r.Run(":" + port)
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"service":     "telex-ai-agent",
		"timestamp":   time.Now(),
		"version":     "2.0.0",
		"ai_provider": "Groq (Free)",
	})
}

func handleTelexWebhook(c *gin.Context) {
	// Verify webhook authenticity
	authHeader := c.GetHeader("Authorization")
	if authHeader != "Bearer "+telexAPIKey {
		log.Printf("⚠️  Unauthorized webhook attempt")
		c.JSON(http.StatusUnauthorized, TelexResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	var webhook TelexWebhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		log.Printf("❌ Error parsing webhook: %v", err)
		c.JSON(http.StatusBadRequest, TelexResponse{
			Status:  "error",
			Message: "Invalid payload",
		})
		return
	}

	log.Printf("📨 Received webhook event: %s from %s", webhook.Event, webhook.Message.From)

	// Handle different event types
	switch webhook.Event {
	case "message.received", "message":
		handleIncomingMessage(c, webhook.Message)
	case "user.joined":
		handleUserJoined(c, webhook.User)
	case "user.typing":
		// Acknowledge typing events
		c.JSON(http.StatusOK, TelexResponse{Status: "acknowledged"})
	default:
		log.Printf("⚠️  Unknown event type: %s", webhook.Event)
		c.JSON(http.StatusOK, TelexResponse{Status: "acknowledged"})
	}
}

func handleIncomingMessage(c *gin.Context, msg TelexMessage) {
	// Ignore messages from ourselves
	if msg.From == agentID {
		c.JSON(http.StatusOK, TelexResponse{Status: "ignored"})
		return
	}

	// Update conversation context
	updateConversationContext(msg.From, msg.Content)

	// Generate AI response
	agentReply, err := generateAIResponse(msg.From, msg.Content)
	if err != nil {
		log.Printf("❌ Error generating AI response: %v", err)
		agentReply = &AgentResponse{
			Reply:      "I apologize, but I'm having trouble processing your message right now. Please try again.",
			Confidence: 0.0,
		}
	}

	// Send response back to Telex.im
	err = sendTelexMessage(msg.From, agentReply.Reply)
	if err != nil {
		log.Printf("❌ Error sending message to Telex: %v", err)
		c.JSON(http.StatusInternalServerError, TelexResponse{
			Status:  "error",
			Message: "Failed to send response",
		})
		return
	}

	log.Printf("✅ Sent reply to %s (confidence: %.2f)", msg.From, agentReply.Confidence)

	c.JSON(http.StatusOK, TelexResponse{
		Status:  "success",
		Message: "Message processed",
	})
}

func handleUserJoined(c *gin.Context, user TelexUser) {
	welcomeMsg := fmt.Sprintf("Welcome %s! 👋 I'm an AI assistant here to help. Feel free to ask me anything!", user.Name)

	err := sendTelexMessage(user.ID, welcomeMsg)
	if err != nil {
		log.Printf("❌ Error sending welcome message: %v", err)
		c.JSON(http.StatusInternalServerError, TelexResponse{
			Status:  "error",
			Message: "Failed to send welcome message",
		})
		return
	}

	c.JSON(http.StatusOK, TelexResponse{Status: "success"})
}

func handleDirectMessage(c *gin.Context) {
	var req struct {
		UserID  string `json:"userId" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateConversationContext(req.UserID, req.Message)

	agentReply, err := generateAIResponse(req.UserID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, agentReply)
}

func generateAIResponse(userID, message string) (*AgentResponse, error) {
	// Get conversation context
	ctx := conversationHistory[userID]

	// Build context-aware prompt
	contextPrompt := buildContextPrompt(ctx, message)

	// Call Groq API (FREE!)
	groqResp, err := callGroqAPI(contextPrompt)
	if err != nil {
		return nil, err
	}

	// Extract intent and entities (simple keyword-based for now)
	intent := extractIntent(message)
	entities := extractEntities(message)

	return &AgentResponse{
		Reply:      groqResp,
		Intent:     intent,
		Entities:   entities,
		Confidence: 0.90,
	}, nil
}

func buildContextPrompt(ctx *ConversationContext, message string) string {
	prompt := "You are a helpful AI assistant integrated with Telex.im messaging platform. "

	if ctx != nil && ctx.MessageCount > 0 {
		prompt += fmt.Sprintf("This is message #%d in the conversation. ", ctx.MessageCount)
		if len(ctx.Topics) > 0 {
			prompt += fmt.Sprintf("Previous topics discussed: %s. ", strings.Join(ctx.Topics, ", "))
		}
	}

	prompt += "Respond naturally and helpfully to the following message:\n\n" + message

	return prompt
}

func callGroqAPI(prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": "llama-3.3-70b-versatile", // FREE model - 70B parameters!
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  1024,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+groqAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq API error: %s", string(body))
	}

	var groqResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &groqResp); err != nil {
		return "", err
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from Groq")
	}

	return groqResp.Choices[0].Message.Content, nil
}

func sendTelexMessage(toUserID, content string) error {
	msg := TelexMessage{
		From:      agentID,
		To:        toUserID,
		Content:   content,
		Timestamp: time.Now(),
		Type:      "text",
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", telexBaseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+telexAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telex API error: %s", string(body))
	}

	return nil
}

func updateConversationContext(userID, message string) {
	ctx, exists := conversationHistory[userID]
	if !exists {
		ctx = &ConversationContext{
			UserID:       userID,
			MessageCount: 0,
			Topics:       []string{},
		}
		conversationHistory[userID] = ctx
	}

	ctx.LastMessage = message
	ctx.MessageCount++
	ctx.Timestamp = time.Now()

	// Extract and add topics
	topics := extractTopics(message)
	for _, topic := range topics {
		if !contains(ctx.Topics, topic) {
			ctx.Topics = append(ctx.Topics, topic)
		}
	}
}

func extractIntent(message string) string {
	lower := strings.ToLower(message)

	if strings.Contains(lower, "help") || strings.Contains(lower, "assist") {
		return "help_request"
	}
	if strings.Contains(lower, "?") {
		return "question"
	}
	if strings.Contains(lower, "thank") {
		return "gratitude"
	}
	if strings.Contains(lower, "hello") || strings.Contains(lower, "hi") {
		return "greeting"
	}

	return "general"
}

func extractEntities(message string) map[string]string {
	entities := make(map[string]string)

	// Simple entity extraction (can be enhanced with NLP)
	words := strings.Fields(strings.ToLower(message))

	for _, word := range words {
		if strings.HasPrefix(word, "@") {
			entities["mention"] = word
		}
		if strings.HasPrefix(word, "#") {
			entities["hashtag"] = word
		}
	}

	return entities
}

func extractTopics(message string) []string {
	topics := []string{}
	lower := strings.ToLower(message)

	keywords := map[string]string{
		"code":    "programming",
		"python":  "programming",
		"go":      "programming",
		"weather": "weather",
		"help":    "support",
		"how":     "tutorial",
	}

	for keyword, topic := range keywords {
		if strings.Contains(lower, keyword) {
			topics = append(topics, topic)
		}
	}

	return topics
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getAgentInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"agentId":      agentID,
		"name":         "Telex AI Assistant",
		"version":      "2.0.0",
		"ai_provider":  "Groq (FREE)",
		"model":        "Llama 3.3 70B Versatile",
		"capabilities": []string{"conversation", "context-awareness", "multi-turn-dialogue", "ultra-fast-inference"},
		"status":       "online",
		"uptime":       time.Now().Format(time.RFC3339),
	})
}

func getConversationHistory(c *gin.Context) {
	userID := c.Param("userId")

	ctx, exists := conversationHistory[userID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "No conversation found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":       ctx.UserID,
		"messageCount": ctx.MessageCount,
		"topics":       ctx.Topics,
		"lastMessage":  ctx.LastMessage,
		"timestamp":    ctx.Timestamp,
	})
}

func getMetrics(c *gin.Context) {
	totalConversations := len(conversationHistory)
	totalMessages := 0

	for _, ctx := range conversationHistory {
		totalMessages += ctx.MessageCount
	}

	c.JSON(http.StatusOK, gin.H{
		"totalConversations": totalConversations,
		"totalMessages":      totalMessages,
		"activeUsers":        totalConversations,
		"aiProvider":         "Groq (FREE)",
		"model":              "Llama 3.3 70B",
		"timestamp":          time.Now(),
	})
}
