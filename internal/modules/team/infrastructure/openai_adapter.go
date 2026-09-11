// internal/modules/team/infrastructure/openai_adapter.go

package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	types "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types/emails"
)

// ============================================================
// OUTBOUND ADAPTER: OpenAIAIAdapter
// ============================================================

// OpenAIAIAdapter implements the AIService interface using OpenAI
type OpenAIAIAdapter struct {
	client  openai.Client // ✅ Use value, not pointer
	enabled bool
	model   string        // ✅ Store the model to use
}

// NewOpenAIAIAdapter creates a new OpenAI AI adapter from config
func NewOpenAIAIAdapter(cfg *config.Config) service.AIService {
	if cfg.OpenAI.APIKey == "" {
		log.Println("⚠️ OpenAI API key not provided, AI features disabled")
		return &OpenAIAIAdapter{enabled: false}
	}

	model := cfg.OpenAI.Model
	if model == "" {
		model = "openai/gpt-4o" // fallback default
	}

	log.Printf("✅ OpenAI AI adapter initialized (official SDK) with model: %s", model)
	return &OpenAIAIAdapter{
		client:  openai.NewClient(option.WithAPIKey(cfg.OpenAI.APIKey)),
		enabled: true,
		model:   model,
	}
}


// GenerateInvitationContent generates personalized email content for invitations
func (a *OpenAIAIAdapter) GenerateInvitationContent(ctx context.Context, req service.GenerateInvitationRequest) (*types.PersonalizedInvitationContent, error) {
	if !a.enabled {
		return nil, fmt.Errorf("AI service not enabled")
	}

	log.Printf("[OpenAI] Generating invitation content for %s (existing user: %v)", req.RecipientEmail, req.IsExistingUser)

	// 1. Build prompt based on user type
	prompt := a.buildPrompt(req)

	// 2. Call OpenAI API using official SDK
	resp, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(a.model), // ✅ Use the configured model
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(`You are a professional email writer for Nuruvent, an event management platform.
Generate warm, professional, and personalized invitation emails.
Return valid JSON matching the following structure:
{
    "subject": "Email subject line (max 60 chars)",
    "greeting": "Opening greeting (e.g., 'Hi Sarah,')",
    "intro": "One sentence introduction",
    "body": "2-3 sentences explaining the invitation",
    "benefits": ["Benefit 1", "Benefit 2", "Benefit 3"],
    "call_to_action": "Button text (max 30 chars)",
    "closing": "Sign-off",
    "pss": "P.S. message"
}`),
			openai.UserMessage(prompt),
		},
		MaxTokens:   openai.Int(500),
		Temperature: openai.Float(0.7),
	})

	if err != nil {
		log.Printf("[OpenAI] API call failed: %v", err)
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	// 3. Extract response content
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	contentText := resp.Choices[0].Message.Content

	// 4. Parse response
	var content types.PersonalizedInvitationContent
	if err := json.Unmarshal([]byte(contentText), &content); err != nil {
		log.Printf("[OpenAI] Failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	log.Printf("[OpenAI] Successfully generated invitation content for %s", req.RecipientEmail)
	return &content, nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

func (a *OpenAIAIAdapter) buildPrompt(req service.GenerateInvitationRequest) string {
	if req.IsExistingUser {
		return a.buildExistingUserPrompt(req)
	}
	return a.buildNewUserPrompt(req)
}

func (a *OpenAIAIAdapter) buildExistingUserPrompt(req service.GenerateInvitationRequest) string {
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

Make the email warm, professional, and personalized.
Highlight why they'd be a great fit for the team.
Mention the team's activity and achievements.
The tone should be exciting and welcoming.
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

func (a *OpenAIAIAdapter) buildNewUserPrompt(req service.GenerateInvitationRequest) string {
	return fmt.Sprintf(`
Generate a welcoming invitation email for a NEW user to create an account and join a team on Nuruvent.

User: %s (new user - no account yet)

Team Details:
- Name: %s
- Type: %s
- Members: %d
- Events: %d

Inviter: %s (Role: %s)

Make the email exciting and welcoming.
Focus on the opportunities and benefits of joining the team.
Highlight what they'll be able to do once they join.
The tone should be warm and encouraging.
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

// Ensure OpenAIAIAdapter implements service.AIService
var _ service.AIService = (*OpenAIAIAdapter)(nil)