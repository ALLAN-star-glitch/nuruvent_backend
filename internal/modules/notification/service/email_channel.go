package service

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
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

// NewEmailChannel creates a new email channel
func NewEmailChannel(cfg EmailChannelConfig) notificationdomain.Channel {
	// Parse all templates from embedded filesystem
	tmpl, err := template.ParseFS(templateFS, "templates/email/*.html")
	if err != nil {
		log.Fatalf("[EmailChannel] Failed to parse email templates: %v", err)
	}

	return &EmailChannel{
		client: resend.NewClient(cfg.EMAIL_API_KEY),
		from:   cfg.EMAIL_FROM,
		tmpl:   tmpl,
	}
}

// GetChannel returns the channel type
func (c *EmailChannel) GetChannel() notificationdomain.NotificationChannel {
	return notificationdomain.ChannelEmail
}

// GetPriority returns the channel priority (lower = higher priority)
func (c *EmailChannel) GetPriority() int {
	return 1 // Email is primary channel
}

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

// renderEmail renders the email template
func (c *EmailChannel) renderEmail(req notificationdomain.ChannelRequest) (string, string, error) {
	// Determine template name based on notification type
	templateName := c.getTemplateName(req.Type)

	// Prepare template data
	data := c.prepareTemplateData(req)

	// Render HTML
	html, err := c.renderHTML(templateName, req.Subject, data)
	if err != nil {
		return "", "", err
	}

	// Build plain text version
	text := c.buildTextVersion(req, data)

	return html, text, nil
}

// ✅ UPDATED: All template names use hyphens to match template definitions
func (c *EmailChannel) getTemplateName(notifType notificationdomain.NotificationType) string {
	switch notifType {
	case notificationdomain.TypeVerificationOTP:
		return "verification-otp"        // ← hyphen
	case notificationdomain.TypeWelcome:
		return "welcome-individual"      // ← hyphen
    case notificationdomain.TypeWelcomeInstitutionKYC: // ✅ NEW
        return "welcome-institution-kyc"
	case notificationdomain.TypeTwoFactor:
		return "two-factor-otp"          // ← hyphen
	case notificationdomain.TypePasswordResetConfirm:
		return "password-reset-confirm"  // ← hyphen
	case notificationdomain.TypeLoginNotification:
		return "login-notification"      // ← hyphen
	default:
		return "welcome-individual"      // ← hyphen
	}
}

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

	case notificationdomain.TypeWelcome:
		if req.Meta["account_type"] == "institution" {
			data["admin_name"] = req.Meta["admin_name"]
			data["institution_name"] = req.Meta["institution_name"]
			data["account_type"] = "institution"
		} else {
			data["name"] = req.Meta["name"]
			data["account_type"] = "individual"
		}

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
	}

	return data
}

func (c *EmailChannel) renderHTML(templateName, title string, data map[string]string) (string, error) {
    var contentBuf bytes.Buffer

    // Determine the actual template name
    actualTemplateName := templateName
    
    // For welcome emails, check if we need KYC version
    if templateName == "welcome-institution" {
        if data["kyc_required"] == "true" {
            actualTemplateName = "welcome-institution-kyc"
        }
    }

    // Execute the specific template
    if err := c.tmpl.ExecuteTemplate(&contentBuf, actualTemplateName, data); err != nil {
        return "", fmt.Errorf("failed to execute email template %s: %w", actualTemplateName, err)
    }

    // Wrap with base template
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

	case notificationdomain.TypeWelcome:
		if data["account_type"] == "institution" {
			text += "Hello " + data["admin_name"] + ",\n\n"
			text += "Congratulations! Your institution " + data["institution_name"] + " has been successfully registered on Nuruvent.\n\n"
			text += "Here is what you can do as an institution account owner:\n"
			text += "- Create and publish events under your institution's brand\n"
			text += "- Accept payments instantly with M-Pesa\n"
			text += "- Issue QR-verified certificates to attendees\n"
			text += "- Track attendance automatically via Zoom or Google Meet\n"
			text += "- Get paid every Monday — only 10% commission\n"
			text += "- Invite team members to manage events\n"
			text += "- Build your institution's professional brand\n\n"
		} else {
			text += "Hello " + data["name"] + ",\n\n"
			text += "Welcome to Nuruvent — the platform that empowers independent trainers, coaches, and consultants to host professional training events in Kenya.\n\n"
			text += "Here is what you can do as an individual professional:\n"
			text += "- Create and publish training events (workshops, webinars, bootcamps, meetups)\n"
			text += "- Accept M-Pesa payments instantly — no manual reconciliation\n"
			text += "- Issue QR-verified certificates to attendees\n"
			text += "- Track attendance automatically via Zoom or Google Meet\n"
			text += "- Get paid every Monday — only 10% commission\n"
			text += "- Save 3+ hours per event with automation\n"
			text += "- Build your personal brand as a trainer\n\n"
		}
		text += "Ready to host your first event? Log in to your dashboard and start creating.\n\n"

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
        text += "Welcome to Nuruvent!\n\n"
        text += "Hello ,\n\n"
        text += "Congratulations! " + data["institution_name"] + " an account has been successfully registered on Nuruvent, by admin: " +  data["admin_name"] +  "\n\n"
        text += "Action Required: Complete Your KYC\n\n"
        text += "To start receiving payouts and unlock all features, please complete your Know Your Customer (KYC) verification within the next 7 days.\n\n"
        text += "The system allows you to:\n"
        text += "- Create and publish events under your institution's brand\n"
        text += "- Accept payments instantly with M-Pesa\n"
        text += "- Issue QR-verified certificates to attendees\n"
        text += "- Track attendance automatically via Zoom or Google Meet\n"
        text += "- Get paid every Monday — only 10% commission\n"
        text += "- Invite team members to manage events\n"
        text += "- Build your institution's professional brand\n\n"
        text += "Complete your KYC within 7 days to:\n"
        text += "- Receive payments directly to your M-Pesa or bank account\n"
        text += "- Get featured in our 'Verified Institutions' directory\n"
        text += "- Access premium event management tools\n"
        text += "- Build trust with attendees\n\n"
        text += "Ready to get started? The admin should login to the dashboard and complete KYC.\n\n"
	}



	text += "--\nNuruvent - Light Your Events. Illuminate Your Growth."
	return text
}

// Ensure EmailChannel implements notificationdomain.Channel
var _ notificationdomain.Channel = (*EmailChannel)(nil)