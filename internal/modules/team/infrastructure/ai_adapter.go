package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/ai"
	types "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types/emails"
)

// ============================================================
// TEAM AI ADAPTER
// ============================================================
//
// Thin domain wrapper around the shared ai.Client. Owns team-specific
// prompt construction, response parsing, and default injection.
//
// Provider selection (OpenRouter, OpenAI, Groq, Together, Gemini via
// OpenAI-compat shim) is handled entirely by the shared client — this
// file never touches HTTP.

type TeamAIAdapter struct {
	client *ai.Client
}

// NewTeamAIAdapter wires a shared ai.Client into the team AIService.
//
// Callers should construct exactly one *ai.Client at app startup and
// inject it into every module adapter.
func NewTeamAIAdapter(client *ai.Client) service.AIService {
	return &TeamAIAdapter{client: client}
}

// GenerateInvitationContent generates a personalized team-invitation
// email body using the shared AI client.
func (a *TeamAIAdapter) GenerateInvitationContent(
	ctx context.Context,
	req service.GenerateInvitationRequest,
) (*types.PersonalizedInvitationContent, error) {

	if !a.client.Enabled() {
		return nil, ai.ErrDisabled
	}

	log.Printf("[team-ai] generating invitation for %s (existing=%v)",
		req.RecipientEmail, req.IsExistingUser)

	const system = `You are a professional email writer for Nuruvent, an event management platform.
Generate warm, professional, and personalized invitation emails.
Always return valid JSON only. Do not add markdown backticks, explanations, or extra text.`

	user := a.buildPrompt(req)

	raw, err := a.client.Chat(ctx, system, user)
	if err != nil {
		return nil, fmt.Errorf("team-ai chat failed: %w", err)
	}

	clean := ai.CleanJSONResponse(raw)

	var result types.PersonalizedInvitationContent
	if err := json.Unmarshal([]byte(clean), &result); err != nil {
		log.Printf("[team-ai] JSON parse failed. clean=%q", truncate(clean, 200))
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	a.applyDefaults(&result, req)

	log.Printf("[team-ai] invitation generated for %s", req.RecipientEmail)
	return &result, nil
}

// ============================================================
// PROMPT BUILDERS
// ============================================================

func (a *TeamAIAdapter) buildPrompt(req service.GenerateInvitationRequest) string {
	if req.IsExistingUser {
		return a.buildExistingUserPrompt(req)
	}
	return a.buildNewUserPrompt(req)
}

func (a *TeamAIAdapter) buildExistingUserPrompt(req service.GenerateInvitationRequest) string {
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

func (a *TeamAIAdapter) buildNewUserPrompt(req service.GenerateInvitationRequest) string {
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

// ============================================================
// DEFAULTS
// ============================================================

func (a *TeamAIAdapter) applyDefaults(
	result *types.PersonalizedInvitationContent,
	req service.GenerateInvitationRequest,
) {
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
		result.Body = fmt.Sprintf("%s has invited you to join %s on Nuruvent.",
			req.InviterName, req.TeamName)
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

// ============================================================
// HELPERS
// ============================================================

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// Ensure TeamAIAdapter implements service.AIService.
var _ service.AIService = (*TeamAIAdapter)(nil)