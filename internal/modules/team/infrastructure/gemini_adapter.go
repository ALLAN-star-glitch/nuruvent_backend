package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	types "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types/emails"
)

// ============================================================
// OUTBOUND ADAPTER: GeminiAIAdapter
// ============================================================

// GeminiAIAdapter implements the AIService interface using Google Gemini
type GeminiAIAdapter struct {
	client  *genai.Client
	enabled bool
	model   string
}

// NewGeminiAIAdapter creates a new Gemini AI adapter from config
func NewGeminiAIAdapter(cfg *config.Config) service.AIService {
	if cfg.Gemini.APIKey == "" {
		log.Println("⚠️ Gemini API key not provided, AI features disabled")
		return &GeminiAIAdapter{enabled: false}
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.Gemini.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Printf("⚠️ Failed to create Gemini client: %v, AI features disabled", err)
		return &GeminiAIAdapter{enabled: false}
	}

	model := cfg.Gemini.Model
	if model == "" {
		model = "gemini-3.6-flash"
	}

	log.Printf("✅ Gemini AI adapter initialized (official SDK) with model: %s", model)
	return &GeminiAIAdapter{
		client:  client,
		enabled: true,
		model:   model,
	}
}

// GenerateInvitationContent generates personalized email content for invitations
func (a *GeminiAIAdapter) GenerateInvitationContent(ctx context.Context, req service.GenerateInvitationRequest) (*types.PersonalizedInvitationContent, error) {
	if !a.enabled {
		return nil, fmt.Errorf("AI service not enabled")
	}

	log.Printf("[Gemini] Generating invitation content for %s (existing user: %v)", req.RecipientEmail, req.IsExistingUser)

	prompt := a.buildPrompt(req)

	temp := float32(0.7)
	genConfig := &genai.GenerateContentConfig{
		Temperature:     &temp,
		MaxOutputTokens: 1024,
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"subject":        {Type: genai.TypeString},
				"greeting":       {Type: genai.TypeString},
				"intro":          {Type: genai.TypeString},
				"body":           {Type: genai.TypeString},
				"benefits":       {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
				"call_to_action": {Type: genai.TypeString},
				"closing":        {Type: genai.TypeString},
				"pss":            {Type: genai.TypeString},
			},
			Required: []string{"subject", "greeting", "body", "call_to_action"},
		},
	}

	content := genai.NewContentFromText(prompt, genai.RoleUser)

	// Exponential backoff and retry loop for transient failures (e.g., 503 UNAVAILABLE, 429)
	var resp *genai.GenerateContentResponse
	var err error
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
		resp, err = a.client.Models.GenerateContent(reqCtx, a.model, []*genai.Content{content}, genConfig)
		cancel()

		if err == nil {
			break
		}

		if isTransientError(err) && attempt < maxRetries {
			backoff := time.Duration(1<<(attempt-1)) * 500 * time.Millisecond
			log.Printf("[Gemini] Transient error on attempt %d/%d: %v. Retrying in %v...", attempt, maxRetries, err, backoff)

			select {
			case <-time.After(backoff):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		} else {
			break
		}
	}

	if err != nil {
		log.Printf("[Gemini] API call failed after retries: %v", err)
		return nil, fmt.Errorf("Gemini API call failed: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidate from Gemini")
	}

	contentText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			contentText += part.Text
		}
	}

	if contentText == "" {
		return nil, fmt.Errorf("empty response text from Gemini")
	}

	var result types.PersonalizedInvitationContent
	if err := json.Unmarshal([]byte(contentText), &result); err != nil {
		log.Printf("[Gemini] Failed to parse structured JSON: %v. Raw text: %s", err, contentText)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Validate required fields - fallback defaults if empty
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

	log.Printf("[Gemini] Successfully generated invitation content for %s", req.RecipientEmail)
	return &result, nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Correct assertion using *genai.APIError from the root SDK package
	if apiErr, ok := err.(*genai.APIError); ok {
		if apiErr.Code == http.StatusServiceUnavailable || apiErr.Code == http.StatusTooManyRequests {
			return true
		}
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "unavailable") ||
		strings.Contains(errStr, "high demand") ||
		strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "resource_exhausted")
}

func (a *GeminiAIAdapter) buildPrompt(req service.GenerateInvitationRequest) string {
	if req.IsExistingUser {
		return a.buildExistingUserPrompt(req)
	}
	return a.buildNewUserPrompt(req)
}

func (a *GeminiAIAdapter) buildExistingUserPrompt(req service.GenerateInvitationRequest) string {
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

func (a *GeminiAIAdapter) buildNewUserPrompt(req service.GenerateInvitationRequest) string {
	return fmt.Sprintf(`
Generate a welcoming invitation email for a NEW user to create an account and join a team on Nuruvent.

User: %s (new user - no account yet)

Team Details:
- Name: %s
- Type: %s
- Members: %d
- Events: %d

Inviter: %s (Role: %s)
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


// Ensure GeminiAIAdapter implements service.AIService
var _ service.AIService = (*GeminiAIAdapter)(nil)