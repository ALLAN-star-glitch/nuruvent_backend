// internal/modules/notification/service/email_channel.go

package service

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/resend/resend-go/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

//go:embed templates/email/*.html
var templateFS embed.FS

// EmailChannel implements the notification channel for email
type EmailChannel struct {
	client *resend.Client
	from   string
	tmpl   *template.Template
}

// EmailChannelConfig holds configuration for email channel
type EmailChannelConfig struct {
	EMAIL_API_KEY string
	EMAIL_FROM    string
}

// ============================================================
// TEMPLATE FUNCTIONS
// ============================================================

// templateFuncMap returns the custom template functions.
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		// trim strips leading/trailing whitespace
		"trim": strings.TrimSpace,

		// split splits a string by a separator
		"split": func(sep, s string) []string {
			if s == "" {
				return []string{}
			}
			return strings.Split(s, sep)
		},

		// join joins a slice of strings with a separator
		"join": func(sep string, s []string) string {
			return strings.Join(s, sep)
		},

		// contains checks if a string contains a substring
		"contains": func(substr, s string) bool {
			return strings.Contains(s, substr)
		},

		// hasPrefix checks if a string has a prefix
		"hasPrefix": func(prefix, s string) bool {
			return strings.HasPrefix(s, prefix)
		},

		// hasSuffix checks if a string has a suffix
		"hasSuffix": func(suffix, s string) bool {
			return strings.HasSuffix(s, suffix)
		},

		// role_display turns a role slug into a human-readable label.
		// "account_admin" -> "Account Admin", "trainer" -> "Trainer".
		"role_display": func(role string) string {
			switch role {
			case "account_admin":
				return "Account Admin"
			case "trainer":
				return "Trainer"
			default:
				return role
			}
		},
	}
}

// ============================================================
// CONSTRUCTOR
// ============================================================

// NewEmailChannel creates a new email channel
func NewEmailChannel(cfg EmailChannelConfig) notificationdomain.Channel {
	// Parse all templates from embedded filesystem with custom functions
	tmpl, err := template.New("").
		Funcs(templateFuncMap()).
		ParseFS(templateFS, "templates/email/*.html")
	if err != nil {
		log.Fatalf("[EmailChannel] Failed to parse email templates: %v", err)
	}

	log.Println("[EmailChannel] ✅ Email templates parsed successfully with custom functions")
	log.Println("[EmailChannel] Available template functions: split, join, contains, hasPrefix, hasSuffix, trim, role_display")

	return &EmailChannel{
		client: resend.NewClient(cfg.EMAIL_API_KEY),
		from:   cfg.EMAIL_FROM,
		tmpl:   tmpl,
	}
}

// ============================================================
// CHANNEL METADATA
// ============================================================

// GetChannel returns the channel type
func (c *EmailChannel) GetChannel() notificationdomain.NotificationChannel {
	return notificationdomain.ChannelEmail
}

// GetPriority returns the channel priority (lower = higher priority)
func (c *EmailChannel) GetPriority() int {
	return 1 // Email is primary channel
}

// ============================================================
// SEND
// ============================================================

// Send sends an email notification
func (c *EmailChannel) Send(ctx context.Context, req notificationdomain.ChannelRequest) error {
	htmlContent, textContent, err := c.renderEmail(req)
	if err != nil {
		return err
	}

	params := &resend.SendEmailRequest{
		From:    c.from,
		To:      []string{req.To},
		Subject: req.Subject,
		Html:    htmlContent,
		Text:    textContent,
	}

	if req.From != "" {
		params.From = req.From
	}

	sent, err := c.client.Emails.Send(params)
	if err != nil {
		log.Printf("[EmailChannel] Failed to send email to %s: %v", req.To, err)
		return fmt.Errorf("%w: %v", notificationdomain.ErrEmailSendFailed, err)
	}

	log.Printf("[EmailChannel] Email sent to %s (ID: %s)", req.To, sent.Id)
	return nil
}

// ============================================================
// RENDER PIPELINE
// ============================================================

// renderEmail renders the email template
func (c *EmailChannel) renderEmail(req notificationdomain.ChannelRequest) (string, string, error) {
	templateName := c.getTemplateName(req.Type)
	data := c.prepareTemplateData(req)

	html, err := c.renderHTML(templateName, req.Subject, data)
	if err != nil {
		return "", "", err
	}

	text := c.buildTextVersion(req, data)

	return html, text, nil
}

// getTemplateName returns the template name based on notification type
func (c *EmailChannel) getTemplateName(notifType notificationdomain.NotificationType) string {
	switch notifType {
	case notificationdomain.TypeVerificationOTP:
		return "verification-otp"
	case notificationdomain.TypeWelcome:
		return "welcome-individual"
	case notificationdomain.TypeWelcomeInstitution:
		return "welcome-institution"
	case notificationdomain.TypeWelcomeInstitutionKYC:
		return "welcome-institution-kyc"
	case notificationdomain.TypeTwoFactor:
		return "two-factor-otp"
	case notificationdomain.TypePasswordResetConfirm:
		return "password-reset-confirm"
	case notificationdomain.TypeLoginNotification:
		return "login-notification"
	case notificationdomain.TypeNewInstitutionAccountRegistration:
		return "new-institution-account"
	case notificationdomain.TypeNewPersonalAccountRegistration:
		return "new-personal-account"
	case notificationdomain.TypeTeamInviteExistingUser:
		return "team-invite-existing-user"
	case notificationdomain.TypeTeamInviteRegistration:
		return "team-invite-registration"
	case notificationdomain.TypeTeamInviteAccepted:
		return "team-invite-accepted"
	case notificationdomain.TypeTeamInviteDeclined:
		return "team-invite-declined"
	default:
		return "welcome-individual"
	}
}

// ============================================================
// TEMPLATE DATA PREPARATION
// ============================================================

func (c *EmailChannel) prepareTemplateData(req notificationdomain.ChannelRequest) map[string]string {
	data := make(map[string]string)

	switch req.Type {
	case notificationdomain.TypeVerificationOTP:
		data["name"] = req.Meta["name"]
		data["otp"] = req.Meta["otp"]
		data["expires"] = req.Meta["expires"]
		data["title"] = req.Meta["title"]
		data["subtitle"] = req.Meta["subtitle"]
		data["description"] = req.Meta["description"]
		data["message"] = req.Meta["message"]
		data["extra_info"] = req.Meta["extra_info"]
		data["warning"] = req.Meta["warning"]

	case notificationdomain.TypeWelcomeInstitution:
		data["admin_name"] = req.Meta["admin_name"]
		data["institution_name"] = req.Meta["institution_name"]
		data["account_type"] = "institution"

	case notificationdomain.TypeWelcome:
		data["name"] = req.Meta["name"]
		data["account_type"] = "individual"

	case notificationdomain.TypeWelcomeInstitutionKYC:
		data["admin_name"] = req.Meta["admin_name"]
		data["institution_name"] = req.Meta["institution_name"]
		data["kyc_required"] = req.Meta["kyc_required"]

	case notificationdomain.TypeTwoFactor:
		data["name"] = req.Meta["name"]
		data["otp"] = req.Meta["otp"]
		data["expires"] = req.Meta["expires"]
		data["ip_address"] = req.Meta["ip_address"]
		data["user_agent"] = req.Meta["user_agent"]

	case notificationdomain.TypePasswordResetConfirm:
		data["name"] = req.Meta["name"]

	case notificationdomain.TypeLoginNotification:
		data["name"] = req.Meta["name"]
		data["time"] = req.Meta["time"]
		data["ip_address"] = req.Meta["ip_address"]
		data["user_agent"] = req.Meta["user_agent"]

	case notificationdomain.TypeNewInstitutionAccountRegistration:
		data["admin_name"] = req.Meta["admin_name"]
		data["institution_name"] = req.Meta["institution_name"]

	case notificationdomain.TypeNewPersonalAccountRegistration:
		data["name"] = req.Meta["name"]

	// ============================================================
	// TEAM INVITATION — EXISTING USER
	// ============================================================
	case notificationdomain.TypeTeamInviteExistingUser:
		data["user_name"] = req.Meta["user_name"]
		data["invited_by"] = req.Meta["invited_by"]
		data["team_name"] = req.Meta["team_name"]
		data["team_id"] = req.Meta["team_id"]
		data["accept_link"] = req.Meta["accept_link"]
		data["expires_in"] = req.Meta["expires_in"]
		data["role"] = req.Meta["role"] // may be empty for existing account members

		// AI content
		if req.Meta["ai_subject"] != "" {
			data["ai_subject"] = req.Meta["ai_subject"]
		}
		if req.Meta["ai_greeting"] != "" {
			data["ai_greeting"] = req.Meta["ai_greeting"]
		}
		if req.Meta["ai_intro"] != "" {
			data["ai_intro"] = req.Meta["ai_intro"]
		}
		if req.Meta["ai_body"] != "" {
			data["ai_body"] = req.Meta["ai_body"]
		}
		if req.Meta["ai_benefits"] != "" {
			data["ai_benefits"] = req.Meta["ai_benefits"]
		}
		if req.Meta["ai_call_to_action"] != "" {
			data["ai_call_to_action"] = req.Meta["ai_call_to_action"]
		}
		if req.Meta["ai_closing"] != "" {
			data["ai_closing"] = req.Meta["ai_closing"]
		}
		if req.Meta["ai_pss"] != "" {
			data["ai_pss"] = req.Meta["ai_pss"]
		}

	// ============================================================
	// TEAM INVITATION — NEW USER
	// ============================================================
	case notificationdomain.TypeTeamInviteRegistration:
		data["name"] = req.Meta["name"]
		data["invited_by"] = req.Meta["invited_by"]
		data["team_name"] = req.Meta["team_name"]
		data["team_id"] = req.Meta["team_id"]
		data["registration_link"] = req.Meta["registration_link"]
		data["expires_in"] = req.Meta["expires_in"]
		data["role"] = req.Meta["role"] // always populated for new users

		// AI content
		if req.Meta["ai_subject"] != "" {
			data["ai_subject"] = req.Meta["ai_subject"]
		}
		if req.Meta["ai_greeting"] != "" {
			data["ai_greeting"] = req.Meta["ai_greeting"]
		}
		if req.Meta["ai_intro"] != "" {
			data["ai_intro"] = req.Meta["ai_intro"]
		}
		if req.Meta["ai_body"] != "" {
			data["ai_body"] = req.Meta["ai_body"]
		}
		if req.Meta["ai_benefits"] != "" {
			data["ai_benefits"] = req.Meta["ai_benefits"]
		}
		if req.Meta["ai_call_to_action"] != "" {
			data["ai_call_to_action"] = req.Meta["ai_call_to_action"]
		}
		if req.Meta["ai_closing"] != "" {
			data["ai_closing"] = req.Meta["ai_closing"]
		}
		if req.Meta["ai_pss"] != "" {
			data["ai_pss"] = req.Meta["ai_pss"]
		}

	// ============================================================
	// TEAM INVITATION — ACCEPTED
	// ============================================================
	case notificationdomain.TypeTeamInviteAccepted:
		data["admin_name"] = req.Meta["admin_name"]
		data["user_name"] = req.Meta["user_name"]
		data["user_email"] = req.Meta["user_email"]
		data["team_name"] = req.Meta["team_name"]
		data["team_id"] = req.Meta["team_id"]
		data["role"] = req.Meta["role"]

		// AI content
		if req.Meta["ai_subject"] != "" {
			data["ai_subject"] = req.Meta["ai_subject"]
		}
		if req.Meta["ai_greeting"] != "" {
			data["ai_greeting"] = req.Meta["ai_greeting"]
		}
		if req.Meta["ai_intro"] != "" {
			data["ai_intro"] = req.Meta["ai_intro"]
		}
		if req.Meta["ai_body"] != "" {
			data["ai_body"] = req.Meta["ai_body"]
		}
		if req.Meta["ai_closing"] != "" {
			data["ai_closing"] = req.Meta["ai_closing"]
		}

	// ============================================================
	// TEAM INVITATION — DECLINED
	// ============================================================
	case notificationdomain.TypeTeamInviteDeclined:
		data["admin_name"] = req.Meta["admin_name"]
		data["user_name"] = req.Meta["user_name"]
		data["user_email"] = req.Meta["user_email"]
		data["team_name"] = req.Meta["team_name"]
		data["team_id"] = req.Meta["team_id"]
		data["role"] = req.Meta["role"]

		// AI content
		if req.Meta["ai_subject"] != "" {
			data["ai_subject"] = req.Meta["ai_subject"]
		}
		if req.Meta["ai_greeting"] != "" {
			data["ai_greeting"] = req.Meta["ai_greeting"]
		}
		if req.Meta["ai_intro"] != "" {
			data["ai_intro"] = req.Meta["ai_intro"]
		}
		if req.Meta["ai_body"] != "" {
			data["ai_body"] = req.Meta["ai_body"]
		}
		if req.Meta["ai_closing"] != "" {
			data["ai_closing"] = req.Meta["ai_closing"]
		}
	}

	return data
}

// ============================================================
// HTML RENDERING
// ============================================================

func (c *EmailChannel) renderHTML(templateName, title string, data map[string]string) (string, error) {
	var contentBuf bytes.Buffer

	if err := c.tmpl.ExecuteTemplate(&contentBuf, templateName, data); err != nil {
		return "", fmt.Errorf("failed to execute email template %s: %w", templateName, err)
	}

	baseData := struct {
		Title   string
		Content template.HTML
		Year    int
	}{
		Title:   title,
		Content: template.HTML(contentBuf.String()),
		Year:    time.Now().Year(),
	}

	var htmlBuf bytes.Buffer
	if err := c.tmpl.ExecuteTemplate(&htmlBuf, "base", baseData); err != nil {
		log.Printf("[EmailChannel] Base template failed, rendering without wrapper: %v", err)
		return contentBuf.String(), nil
	}

	return htmlBuf.String(), nil
}

// ============================================================
// PLAIN-TEXT RENDERING
// ============================================================

// roleDisplayPlain is the non-template equivalent of the role_display
// template function. Used by the plain-text builders.
func roleDisplayPlain(role string) string {
	switch role {
	case "account_admin":
		return "Account Admin"
	case "trainer":
		return "Trainer"
	default:
		return role
	}
}

func (c *EmailChannel) buildTextVersion(req notificationdomain.ChannelRequest, data map[string]string) string {
	text := "Nuruvent - " + req.Subject + "\n\n"

	switch req.Type {
	case notificationdomain.TypeVerificationOTP:
		text += data["description"] + "\n\n"
		text += "Hello " + data["name"] + ",\n\n"
		text += data["message"] + "\n\n"
		text += "Your verification code is: " + data["otp"] + "\n\n"
		text += "This code expires in " + data["expires"] + ".\n\n"
		if data["warning"] != "" {
			text += data["warning"] + "\n\n"
		}

	case notificationdomain.TypeWelcomeInstitution:
		text += "Hello " + data["admin_name"] + ",\n\n"
		text += "Welcome to Nuruvent! Your institution " + data["institution_name"] + " has been successfully registered.\n\n"
		text += "Your role: Account Admin\n"
		text += "You have full control to manage events, members, and settings for your institution.\n\n"
		text += "What you can do:\n"
		text += "- Create and publish events under your institution's brand\n"
		text += "- Invite team members to manage events with you\n"
		text += "- Accept M-Pesa payments — 3.5% commission\n"
		text += "- Issue QR-verified certificates to attendees\n\n"
		text += "Ready to start? Log in to your dashboard and create your first event.\n\n"
		text += "Note: Complete your KYC verification to receive payouts.\n\n"

	case notificationdomain.TypeWelcome:
		text += "Hello " + data["name"] + ",\n\n"
		text += "Welcome to Nuruvent — the platform that empowers independent trainers, coaches, and consultants to host professional training events in Kenya.\n\n"
		text += "Your Professional Account is Ready\n"
		text += "Start hosting workshops, webinars, and bootcamps today\n\n"
		text += "Here is what you can do as an individual professional:\n"
		text += "- Create and publish training events (workshops, webinars, bootcamps, meetups)\n"
		text += "- Accept M-Pesa payments instantly — no manual reconciliation\n"
		text += "- Issue QR-verified certificates to attendees\n"
		text += "- Track attendance automatically via Zoom or Google Meet\n"
		text += "- Get paid every Monday — We take only 3.5% commission\n"
		text += "- Save 3+ hours per event with automation\n"
		text += "- Build your personal brand as a trainer\n\n"
		text += "Ready to host your first event? Log in to your dashboard and start creating.\n\n"
		text += "Note: Complete your KYC verification to start receiving payouts. Check your dashboard for details.\n\n"

	case notificationdomain.TypeTwoFactor:
		text += "Hello " + data["name"] + ",\n\n"
		text += "You requested a two-factor authentication code for your Nuruvent account.\n\n"
		text += "Your 2FA verification code is: " + data["otp"] + "\n\n"
		text += "This code expires in " + data["expires"] + ".\n\n"
		text += "If you did not attempt to log in, please reset your password immediately.\n\n"
		text += "For security, never share this code with anyone.\n\n"

	case notificationdomain.TypePasswordResetConfirm:
		text += "Hello " + data["name"] + ",\n\n"
		text += "Your Nuruvent password has been successfully changed.\n\n"
		text += "If you did not perform this action, please contact our support team immediately at hello@nuruvent.com\n\n"

	case notificationdomain.TypeLoginNotification:
		text += "Hello " + data["name"] + ",\n\n"
		text += "We detected a new login to your Nuruvent account.\n\n"
		text += "Time: " + data["time"] + "\n"
		text += "IP Address: " + data["ip_address"] + "\n"
		text += "Device: " + data["user_agent"] + "\n\n"
		text += "If this was you, you can safely ignore this notification.\n"
		text += "If you did not log in, please reset your password immediately.\n\n"

	case notificationdomain.TypeWelcomeInstitutionKYC:
		text += "Hello " + data["admin_name"] + ",\n\n"
		text += "Congratulations! " + data["institution_name"] + " has been successfully registered on Nuruvent.\n\n"
		text += "Action Required: Complete Your KYC\n\n"
		text += "To start receiving payouts and unlock all features, please complete your Know Your Customer (KYC) verification within the next 7 days.\n\n"
		text += "As an Institution Host, you can:\n"
		text += "- Create and publish events under your institution's brand\n"
		text += "- Accept payments instantly with M-Pesa\n"
		text += "- Issue QR-verified certificates to attendees\n"
		text += "- Track attendance automatically via Zoom or Google Meet\n"
		text += "- Get paid every Monday — We take only 3.5% commission\n"
		text += "- Invite team members to manage events\n"
		text += "- Build your institution's professional brand\n\n"
		text += "Complete your KYC within 7 days to:\n"
		text += "- Receive payments directly to your M-Pesa or bank account\n"
		text += "- Get featured in our 'Verified Institutions' directory\n"
		text += "- Access premium event management tools\n"
		text += "- Build trust with attendees\n\n"
		text += "Ready to get started? Login to the dashboard and complete KYC.\n\n"

	case notificationdomain.TypeNewInstitutionAccountRegistration:
		text += "Hello " + data["admin_name"] + ",\n\n"
		text += "Success! " + data["institution_name"] + " has been successfully registered on Nuruvent.\n"
		text += "The account has been created with Host privileges, allowing the account admin to manage events for the institution.\n\n"
		text += "As an Institution Host, they can:\n"
		text += "- Create and publish events under their institution's brand\n"
		text += "- Accept payments instantly with M-Pesa\n"
		text += "- Issue QR-verified certificates to attendees\n"
		text += "- Track attendance automatically via Zoom or Google Meet\n"
		text += "- Get paid every Monday — We take only 3.5% commission\n"
		text += "- Invite team members to manage events\n"
		text += "- Build their institution's professional brand\n\n"
		text += "Important: Please follow up with the institution to help them get more acquainted with Nuruvent.\n\n"

	case notificationdomain.TypeNewPersonalAccountRegistration:
		text += "Hello " + data["name"] + ",\n\n"
		text += "Welcome to Nuruvent — the platform that empowers independent trainers, coaches, and consultants to host professional training events in Kenya.\n"
		text += "Your individual account has been successfully created.\n\n"
		text += "Here is what you can do with your personal account:\n"
		text += "- Create and publish training events (workshops, webinars, bootcamps, meetups)\n"
		text += "- Accept M-Pesa payments instantly — no manual reconciliation\n"
		text += "- Issue/Accept QR-verified certificates to attendees\n"
		text += "- Track attendance automatically via Zoom or Google Meet\n"
		text += "- Get paid every Monday — We take only 3.5% commission\n"
		text += "- Save 3+ hours per event with automation\n"
		text += "- Build your personal brand as a trainer/attendee\n\n"
		text += "Please welcome them to Nuruvent.\n\n"

	// ============================================================
	// TEAM INVITATION — EXISTING USER
	// ============================================================
	case notificationdomain.TypeTeamInviteExistingUser:
		text += "Hello " + data["user_name"] + ",\n\n"
		text += data["invited_by"] + " has invited you to join the team " + data["team_name"]
		if data["role"] != "" {
			text += " as a " + roleDisplayPlain(data["role"])
		}
		text += " on Nuruvent.\n\n"
		text += "To accept this invitation, click the link below:\n"
		text += data["accept_link"] + "\n\n"
		text += "This invitation will expire in " + data["expires_in"] + ".\n\n"
		text += "If you already have a Nuruvent account, simply log in and you'll be automatically added to the team.\n\n"
		text += "If you did not expect this invitation, please ignore this email.\n\n"

	// ============================================================
	// TEAM INVITATION — NEW USER
	// ============================================================
	case notificationdomain.TypeTeamInviteRegistration:
		text += "Hello " + data["name"] + ",\n\n"
		text += data["invited_by"] + " has invited you to join the team " + data["team_name"]
		if data["role"] != "" {
			text += " as a " + roleDisplayPlain(data["role"])
		}
		text += " on Nuruvent.\n\n"
		text += "To accept this invitation, click the link below to create your account:\n"
		text += data["registration_link"] + "\n\n"
		text += "This invitation will expire in " + data["expires_in"] + ".\n\n"
		text += "If you already have a Nuruvent account, simply log in and you'll be automatically added to the team.\n\n"
		text += "If you did not expect this invitation, please ignore this email.\n\n"

	// ============================================================
	// TEAM INVITATION — ACCEPTED
	// ============================================================
	case notificationdomain.TypeTeamInviteAccepted:
		text += "Hello " + data["admin_name"] + ",\n\n"
		text += data["user_name"] + " has accepted your invitation and joined " + data["team_name"]
		if data["role"] != "" {
			text += " as a " + roleDisplayPlain(data["role"])
		}
		text += " on Nuruvent.\n\n"
		text += "Team Member Details:\n"
		text += "- Name: " + data["user_name"] + "\n"
		text += "- Email: " + data["user_email"] + "\n\n"
		text += "You can view your updated team members list in your dashboard.\n\n"

	// ============================================================
	// TEAM INVITATION — DECLINED
	// ============================================================
	case notificationdomain.TypeTeamInviteDeclined:
		text += "Hello " + data["admin_name"] + ",\n\n"
		text += data["user_name"] + " has declined your invitation to join " + data["team_name"]
		if data["role"] != "" {
			text += " as a " + roleDisplayPlain(data["role"])
		}
		text += " on Nuruvent.\n\n"
		text += "You can invite other users to join your team at any time.\n\n"
	}

	text += "\n--\nNuruvent - Light Your Events. Illuminate Your Growth."
	return text
}

// ============================================================
// INTERFACE ASSERTION
// ============================================================

// Ensure EmailChannel implements notificationdomain.Channel
var _ notificationdomain.Channel = (*EmailChannel)(nil)