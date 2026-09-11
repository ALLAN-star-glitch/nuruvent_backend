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
// OUTBOUND ADAPTER: OpenRouterAIAdapter (Direct HTTP Client)
// ============================================================

// OpenRouterAIAdapter implements the AIService interface using OpenRouter's OpenAI-compatible API
type OpenRouterAIAdapter struct {
	apiKey  string
	model   string
	enabled bool
	client  *http.Client
}

// NewOpenRouterAIAdapter creates a new OpenRouter AI adapter from config
func NewOpenRouterAIAdapter(cfg *config.Config) service.AIService {
	apiKey := strings.TrimSpace(cfg.OpenRouter.APIKey)
	if apiKey == "" {
		log.Println("⚠️ OpenRouter API key not provided, AI features disabled")
		return &OpenRouterAIAdapter{enabled: false}
	}

	model := strings.TrimSpace(cfg.OpenRouter.Model)
	if model == "" {
		model = "meta-llama/llama-3.1-8b-instruct:free"
	}

	log.Printf("✅ OpenRouter AI adapter initialized with model: %s", model)
	return &OpenRouterAIAdapter{
		apiKey:  apiKey,
		model:   model,
		enabled: true,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// OpenRouterRequest represents the request body for OpenRouter Chat Completions API
type OpenRouterRequest struct {
	Model          string            `json:"model"`
	Messages       []OpenRouterMsg   `json:"messages"`
	Temperature    float64           `json:"temperature,omitempty"`
	MaxTokens      int               `json:"max_tokens,omitempty"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
}

// OpenRouterMsg represents a chat message
type OpenRouterMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterResponse represents the JSON response structure from OpenRouter
type OpenRouterResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

// GenerateInvitationContent generates personalized email content using OpenRouter
func (a *OpenRouterAIAdapter) GenerateInvitationContent(ctx context.Context, req service.GenerateInvitationRequest) (*types.PersonalizedInvitationContent, error) {
	if !a.enabled {
		return nil, fmt.Errorf("AI service not enabled")
	}

	log.Printf("[OpenRouter] Generating invitation content for %s (existing user: %v)", req.RecipientEmail, req.IsExistingUser)

	prompt := a.buildPrompt(req)

	messages := []OpenRouterMsg{
		{
			Role:    "system",
			Content: `You are a professional email writer for Nuruvent, an event management platform. Generate warm, professional, and personalized invitation emails. Always return valid JSON only. Do not add markdown backticks, explanations, or extra text.`,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	requestBody := OpenRouterRequest{
		Model:          a.model,
		Messages:       messages,
		Temperature:    0.7,
		MaxTokens:      800,
		ResponseFormat: map[string]string{"type": "json_object"},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("HTTP-Referer", "https://nuruvent.com") // Optional header for OpenRouter rankings
	httpReq.Header.Set("X-Title", "Nuruvent Engine")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		log.Printf("[OpenRouter] HTTP request failed: %v", err)
		return nil, fmt.Errorf("OpenRouter API call failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[OpenRouter] API returned status: %d", resp.StatusCode)
		log.Printf("[OpenRouter] Response body: %s", string(body))
		return nil, fmt.Errorf("OpenRouter API returned status %d: %s", resp.StatusCode, string(body))
	}

	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		log.Printf("[OpenRouter] Failed to parse API response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if openRouterResp.Error != nil {
		return nil, fmt.Errorf("OpenRouter API error: %s", openRouterResp.Error.Message)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenRouter API")
	}

	contentText := openRouterResp.Choices[0].Message.Content
	log.Printf("[OpenRouter] Raw response length: %d characters", len(contentText))

	// Clean markdown wrappers if returned by AI
	contentText = strings.TrimSpace(contentText)
	contentText = strings.TrimPrefix(contentText, "```json")
	contentText = strings.TrimPrefix(contentText, "```")
	contentText = strings.TrimSuffix(contentText, "```")
	contentText = strings.TrimSpace(contentText)

	jsonStart := strings.Index(contentText, "{")
	jsonEnd := strings.LastIndex(contentText, "}")
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		contentText = contentText[jsonStart : jsonEnd+1]
	}

	var result types.PersonalizedInvitationContent
	if err := json.Unmarshal([]byte(contentText), &result); err != nil {
		log.Printf("[OpenRouter] Failed to parse JSON content: %v", err)
		log.Printf("[OpenRouter] Response content: %s", contentText[:minIntVal(len(contentText), 200)])
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	a.applyDefaults(&result, req)

	log.Printf("[OpenRouter] Successfully generated invitation content for %s", req.RecipientEmail)
	return &result, nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

func minIntVal(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (a *OpenRouterAIAdapter) buildPrompt(req service.GenerateInvitationRequest) string {
	if req.IsExistingUser {
		return a.buildExistingUserPrompt(req)
	}
	return a.buildNewUserPrompt(req)
}

func (a *OpenRouterAIAdapter) buildExistingUserPrompt(req service.GenerateInvitationRequest) string {
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

// In buildExistingUserPrompt and buildNewUserPrompt inside openrouter_adapter.go

func (a *OpenRouterAIAdapter) buildNewUserPrompt(req service.GenerateInvitationRequest) string {
	return fmt.Sprintf(`
Generate a warm, friendly, and conversational email invitation for a recipient to join a team on Nuruvent.

Context:
- Recipient Email: %s
- Inviter Name: %s (%s)
- Team Name: %s (%s)
- Active Team Members: %d
- Total Hosted Events: %d

Guidelines:
1. Tone: Warm, approachable, human, and exciting (avoid corporate buzzwords like "elevate your event planning skills").
2. Greeting: Use "Hey there," or "Welcome!" since the recipient's first name is unknown. Never use email handles or usernames in greetings.
3. Benefits: Provide 2-3 clear, conversational value points. Do not return empty bullet items.

Return ONLY valid JSON matching this schema:
{
    "subject": "Catchy, friendly subject line",
    "greeting": "Friendly greeting",
    "intro": "Warm opening line explaining why they are receiving this",
    "body": "Friendly 2-sentence paragraph inviting them to collaborate with the team",
    "benefits": [
        "Collaborate directly with team members on upcoming events",
        "Manage event schedules, check-ins, and guest lists seamlessly",
        "Access shared team analytics and event resources"
    ],
    "call_to_action": "Join Team",
    "closing": "Warm closing sign-off",
    "pss": "Optional short, friendly postscript"
}
`,
		req.RecipientEmail,
		req.InviterName,
		req.InviterRole,
		req.TeamName,
		req.TeamType,
		req.TeamMemberCount,
		req.TeamEventCount,
	)
}

func (a *OpenRouterAIAdapter) applyDefaults(result *types.PersonalizedInvitationContent, req service.GenerateInvitationRequest) {
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
}

// Ensure OpenRouterAIAdapter implements service.AIService
var _ service.AIService = (*OpenRouterAIAdapter)(nil)