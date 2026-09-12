// internal/modules/team/infrastructure/team_ai_adapter.go

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

	log.Printf("[team-ai] generating invitation for %s (existingUser=%v, existingAccountMember=%v, role=%q)",
		req.RecipientEmail, req.IsExistingUser, req.IsExistingAccountMember, req.InvitedRole)

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

// buildPrompt picks the right prompt template based on the recipient's
// state:
//
//   - Not registered            → registration prompt (mentions role)
//   - Registered, not a member  → existing-user prompt (mentions role)
//   - Registered, already a member → team-add prompt (no role mention)
func (a *TeamAIAdapter) buildPrompt(req service.GenerateInvitationRequest) string {
	switch {
	case req.IsExistingAccountMember:
		return a.buildExistingAccountMemberPrompt(req)
	case req.IsExistingUser:
		return a.buildExistingUserPrompt(req)
	default:
		return a.buildNewUserPrompt(req)
	}
}

// buildExistingAccountMemberPrompt is for recipients who already belong
// to the account. The invitation just adds them to a team; their role is
// unchanged. Do not mention a role in the email.
func (a *TeamAIAdapter) buildExistingAccountMemberPrompt(req service.GenerateInvitationRequest) string {
	recentEvents := "No recent events"
	if len(req.RecentEventNames) > 0 {
		recentEvents = strings.Join(req.RecentEventNames, ", ")
	}

	teamMembers := "No team members listed"
	if len(req.TeamMemberNames) > 0 {
		teamMembers = strings.Join(req.TeamMemberNames, ", ")
	}

	return fmt.Sprintf(`
Generate a personalized email notifying an EXISTING ACCOUNT MEMBER that they have been invited to join an additional team within the same account on Nuruvent.

IMPORTANT: The recipient is ALREADY a member of the account. Their role in the account does not change. Do NOT mention being granted a role, joining as a trainer/admin, or any role-related wording. Just invite them to join the team.

Recipient Details:
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
    "body": "2-3 sentences explaining that they have been invited to join the team",
    "benefits": ["Benefit 1", "Benefit 2", "Benefit 3"],
    "call_to_action": "Button text (e.g. 'View Invitation' or 'Join Team')",
    "closing": "Sign-off",
    "pss": "P.S. message"
}

Tone: professional and welcoming. Emphasize collaboration, not role changes.
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

// buildExistingUserPrompt is for recipients who are registered on the
// platform but not yet members of the account. The invitation grants
// them a role in the account.
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
Generate a personalized invitation email for an EXISTING Nuruvent user to join a team and become a member of its parent account.

The recipient will be granted the role "%s" in the account. Mention this role naturally in the email (e.g., "join as a Trainer" or "join as an Account Admin").

User Details:
- Name: %s
- Email: %s

Team Details:
- Name: %s
- Type: %s
- Members: %d
- Events: %d

Inviter: %s (Role: %s)
Role Being Granted: %s

Recent Events: %s
Team Members: %s

Return ONLY valid JSON. No markdown, no extra text. The JSON must have these exact fields:
{
    "subject": "Email subject line",
    "greeting": "Opening greeting",
    "intro": "One sentence introduction",
    "body": "2-3 sentences explaining the invitation, including the role they will receive",
    "benefits": ["Benefit 1", "Benefit 2", "Benefit 3"],
    "call_to_action": "Button text",
    "closing": "Sign-off",
    "pss": "P.S. message"
}

Make it warm, professional, and personalized. Reference the role naturally.
`,
		req.InvitedRole,
		req.RecipientName,
		req.RecipientEmail,
		req.TeamName,
		req.TeamType,
		req.TeamMemberCount,
		req.TeamEventCount,
		req.InviterName,
		req.InviterRole,
		req.InvitedRole,
		recentEvents,
		teamMembers,
	)
}

// buildNewUserPrompt is for recipients with no account at all. They will
// register and, in the same flow, be granted a role in the account.
func (a *TeamAIAdapter) buildNewUserPrompt(req service.GenerateInvitationRequest) string {
	return fmt.Sprintf(`
Generate a warm, friendly, and conversational email invitation for a recipient to join a team on Nuruvent.

The recipient does not have a Nuruvent account yet. When they register, they will be granted the role "%s" in the team's parent account. Mention this role naturally in the email.

Context:
- Recipient Email: %s
- Inviter Name: %s (%s)
- Team Name: %s (%s)
- Role Being Granted: %s
- Active Team Members: %d
- Total Hosted Events: %d

Guidelines:
1. Tone: Warm, approachable, human, and exciting (avoid corporate buzzwords like "elevate your event planning skills").
2. Greeting: Use "Hey there," or "Welcome!" since the recipient's first name is unknown. Never use email handles or usernames in greetings.
3. Benefits: Provide 2-3 clear, conversational value points. Do not return empty bullet items.
4. Mention the role ("%s") naturally — the recipient should know what role they're joining as.

Return ONLY valid JSON matching this schema:
{
    "subject": "Catchy, friendly subject line",
    "greeting": "Friendly greeting",
    "intro": "Warm opening line explaining why they are receiving this",
    "body": "Friendly 2-sentence paragraph inviting them to collaborate with the team as a role",
    "benefits": [
        "Collaborate directly with team members on upcoming events",
        "Manage event schedules, check-ins, and guest lists seamlessly",
        "Access shared team analytics and event resources"
    ],
    "call_to_action": "Create Account & Join Team",
    "closing": "Warm closing sign-off",
    "pss": "Optional short, friendly postscript"
}
`,
		req.InvitedRole,
		req.RecipientEmail,
		req.InviterName,
		req.InviterRole,
		req.TeamName,
		req.TeamType,
		req.InvitedRole,
		req.TeamMemberCount,
		req.TeamEventCount,
		req.InvitedRole,
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
		if req.IsExistingAccountMember {
			result.Subject = fmt.Sprintf("You've been invited to join %s", req.TeamName)
		} else {
			result.Subject = fmt.Sprintf("You're invited to join %s as a %s", req.TeamName, req.InvitedRole)
		}
	}

	if result.Greeting == "" {
		if req.RecipientName != "" {
			result.Greeting = fmt.Sprintf("Hi %s,", req.RecipientName)
		} else {
			result.Greeting = "Hello,"
		}
	}

	if result.Body == "" {
		if req.IsExistingAccountMember {
			result.Body = fmt.Sprintf("%s has invited you to join the %s team on Nuruvent.",
				req.InviterName, req.TeamName)
		} else {
			result.Body = fmt.Sprintf("%s has invited you to join %s on Nuruvent as a %s.",
				req.InviterName, req.TeamName, req.InvitedRole)
		}
	}

	if result.CallToAction == "" {
		switch {
		case req.IsExistingAccountMember:
			result.CallToAction = "View Invitation"
		case req.IsExistingUser:
			result.CallToAction = "Accept Invitation"
		default:
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