// internal/modules/team/infrastructure/groq_adapter.go

package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	types "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types/emails"
)

// ============================================================
// OUTBOUND ADAPTER: GroqAIAdapter (Direct HTTP Client)
// ============================================================

// GroqAIAdapter implements the AIService interface using Groq
type GroqAIAdapter struct {
	apiKey  string
	model   string
	enabled bool
	client  *http.Client
}

// NewGroqAIAdapter creates a new Groq AI adapter from config
func NewGroqAIAdapter(cfg *config.Config) service.AIService {
	if cfg.Groq.APIKey == "" {
		log.Println("⚠️ Groq API key not provided, AI features disabled")
		return &GroqAIAdapter{enabled: false}
	}

	model := cfg.Groq.Model
	if model == "" {
		model = "llama-3.1-8b-instant" // ✅ Updated model name
	}

	log.Printf("✅ Groq AI adapter initialized with model: %s", model)
	return &GroqAIAdapter{
		apiKey:  cfg.Groq.APIKey,
		model:   model,
		enabled: true,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// GroqRequest represents the request body for Groq API
type GroqRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GroqResponse represents the response from Groq API
type GroqResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// GenerateInvitationContent generates personalized email content for invitations
func (a *GroqAIAdapter) GenerateInvitationContent(ctx context.Context, req service.GenerateInvitationRequest) (*types.PersonalizedInvitationContent, error) {
	if !a.enabled {
		return nil, fmt.Errorf("AI service not enabled")
	}

	log.Printf("[Groq] Generating invitation content for %s (existing user: %v)", req.RecipientEmail, req.IsExistingUser)

	// 1. Build prompt based on user type
	prompt := a.buildPrompt(req)

	// 2. Build messages
	messages := []Message{
		{
			Role:    "system",
			Content: `You are a professional email writer for Nuruvent, an event management platform. Generate warm, professional, and personalized invitation emails. Always return valid JSON. Do not add markdown or extra text.`,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// 3. Build request body
	requestBody := GroqRequest{
		Model:       a.model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   800,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[Groq] Request body size: %d bytes", len(jsonBody))

	// 4. Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// 5. Send request
	resp, err := a.client.Do(httpReq)
	if err != nil {
		log.Printf("[Groq] HTTP request failed: %v", err)
		return nil, fmt.Errorf("Groq API call failed: %w", err)
	}
	defer resp.Body.Close()

	// 6. Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 7. Check status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("[Groq] API returned status: %d", resp.StatusCode)
		log.Printf("[Groq] Response body: %s", string(body))
		return nil, fmt.Errorf("Groq API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 8. Parse response
	var groqResp GroqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		log.Printf("[Groq] Failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	contentText := groqResp.Choices[0].Message.Content
	log.Printf("[Groq] Raw response length: %d characters", len(contentText))

	// 9. Clean up markdown
	contentText = strings.TrimSpace(contentText)
	contentText = strings.TrimPrefix(contentText, "```json")
	contentText = strings.TrimPrefix(contentText, "```")
	contentText = strings.TrimSuffix(contentText, "```")
	contentText = strings.TrimSpace(contentText)

	// 10. Try to extract JSON if it's embedded in the response
	jsonStart := strings.Index(contentText, "{")
	jsonEnd := strings.LastIndex(contentText, "}")
	
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		contentText = contentText[jsonStart : jsonEnd+1]
	}

	// 11. Parse JSON
	var result types.PersonalizedInvitationContent
	if err := json.Unmarshal([]byte(contentText), &result); err != nil {
		log.Printf("[Groq] Failed to parse response: %v", err)
		log.Printf("[Groq] Response content: %s", contentText[:min(len(contentText), 200)])
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 12. Validate and fill missing fields
	if result.Subject == "" {
		result.Subject = fmt.Sprintf("You're Invited to Join %s", req.TeamName)
	}
	if result.Greeting == "" {
		if req.RecipientName != "" {
			result.Greeting = fmt.Sprintf("Hi %s,", req.RecipientName)
		} else {
			result.Greeting = "Hello,"
		}
	}
	if result.Body == "" {
		result.Body = fmt.Sprintf("%s has invited you to join %s on Nuruvent.", req.InviterName, req.TeamName)
	}
	if result.CallToAction == "" {
		if req.IsExistingUser {
			result.CallToAction = "Accept Invitation"
		} else {
			result.CallToAction = "Create Account & Join Team"
		}
	}
	if len(result.Benefits) == 0 {
		result.Benefits = []string{
			"Collaborate with team members",
			"Create and manage training events",
			"Issue certificates to attendees",
		}
	}
	if result.Closing == "" {
		result.Closing = "Best regards,\nThe Nuruvent Team"
	}

	log.Printf("[Groq] Successfully generated invitation content for %s", req.RecipientEmail)
	return &result, nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (a *GroqAIAdapter) buildPrompt(req service.GenerateInvitationRequest) string {
	if req.IsExistingUser {
		return a.buildExistingUserPrompt(req)
	}
	return a.buildNewUserPrompt(req)
}

func (a *GroqAIAdapter) buildExistingUserPrompt(req service.GenerateInvitationRequest) string {
	recentEvents := "No recent events"
	if len(req.RecentEventNames) > 0 {
		recentEvents = strings.Join(req.RecentEventNames, ", ")
	}

	teamMembers := "No team members listed"
	if len(req.TeamMemberNames) > 0 {
		teamMembers = strings.Join(req.TeamMemberNames, ", ")
	}

	return fmt.Sprintf(`
Generate a personalized invitation email for an EXISTING user to join a team on Nuruvent.

User Details:
- Name: %s
- Email: %s

Team Details:
- Name: %s
- Type: %s
- Members: %d
- Events: %d

Inviter: %s (Role: %s)

Recent Events: %s
Team Members: %s

Return ONLY valid JSON. No markdown, no extra text. The JSON must have these exact fields:
{
    "subject": "Email subject line",
    "greeting": "Opening greeting",
    "intro": "One sentence introduction",
    "body": "2-3 sentences explaining the invitation",
    "benefits": ["Benefit 1", "Benefit 2", "Benefit 3"],
    "call_to_action": "Button text",
    "closing": "Sign-off",
    "pss": "P.S. message"
}

Make it warm, professional, and personalized.
`,
		req.RecipientName,
		req.RecipientEmail,
		req.TeamName,
		req.TeamType,
		req.TeamMemberCount,
		req.TeamEventCount,
		req.InviterName,
		req.InviterRole,
		recentEvents,
		teamMembers,
	)
}

func (a *GroqAIAdapter) buildNewUserPrompt(req service.GenerateInvitationRequest) string {
	return fmt.Sprintf(`
Generate a welcoming invitation email for a NEW user to create an account and join a team on Nuruvent.

User: %s (new user - no account yet)

Team Details:
- Name: %s
- Type: %s
- Members: %d
- Events: %d

Inviter: %s (Role: %s)

Return ONLY valid JSON. No markdown, no extra text. The JSON must have these exact fields:
{
    "subject": "Email subject line",
    "greeting": "Opening greeting",
    "intro": "One sentence introduction",
    "body": "2-3 sentences explaining the invitation",
    "benefits": ["Benefit 1", "Benefit 2", "Benefit 3"],
    "call_to_action": "Button text",
    "closing": "Sign-off",
    "pss": "P.S. message"
}

Make it exciting and welcoming.
`,
		req.RecipientEmail,
		req.TeamName,
		req.TeamType,
		req.TeamMemberCount,
		req.TeamEventCount,
		req.InviterName,
		req.InviterRole,
	)
}

// Ensure GroqAIAdapter implements service.AIService
var _ service.AIService = (*GroqAIAdapter)(nil)