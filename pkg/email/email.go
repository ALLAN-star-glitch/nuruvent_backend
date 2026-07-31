package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/resend/resend-go/v3"
)

//go:embed templates/*.html
var templateFS embed.FS

// EmailService handles all email operations
type EmailService struct {
	client *resend.Client
	from   string
	tmpl   *template.Template
}

// NewEmailService creates a new email service
func NewEmailService(apiKey, from string) *EmailService {
	// Parse all templates from embedded filesystem
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("[EmailService] Failed to parse templates: %v", err)
	}

	return &EmailService{
		client: resend.NewClient(apiKey),
		from:   from,
		tmpl:   tmpl,
	}
}

// ================================================
// HELPER FUNCTIONS
// ================================================

func (s *EmailService) renderEmail(templateName, title string, data interface{}) (string, error) {
    var contentBuf bytes.Buffer
    if err := s.tmpl.ExecuteTemplate(&contentBuf, templateName, data); err != nil {
        return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
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
    if err := s.tmpl.ExecuteTemplate(&htmlBuf, "base", baseData); err != nil {
        // If base template fails, try rendering without it
        log.Printf("[EmailService] Base template failed, rendering without wrapper: %v", err)
        return contentBuf.String(), nil
    }

    return htmlBuf.String(), nil
}

func (s *EmailService) sendEmail(to, subject, html, text string) error {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		Html:    html,
		Text:    text,
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		log.Printf("[EmailService] Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("[EmailService] Email sent to %s (ID: %s)", to, sent.Id)
	return nil
}

// ================================================
// VERIFICATION OTP - GENERIC
// ================================================

// SendVerificationOTP sends verification OTP for multiple purposes
func (s *EmailService) SendVerificationOTP(to, name, otp, expires, purpose string, meta map[string]string) error {
	title, subtitle, description, message, extraInfo, warning := s.getVerificationContent(purpose, meta, name)

	data := struct {
		Name        string
		OTP         string
		Expires     string
		Title       string
		Subtitle    string
		Description string
		Message     string
		ExtraInfo   string
		Warning     string
	}{
		Name:        name,
		OTP:         otp,
		Expires:     expires,
		Title:       title,
		Subtitle:    subtitle,
		Description: description,
		Message:     message,
		ExtraInfo:   extraInfo,
		Warning:     warning,
	}

	html, err := s.renderEmail("verification-otp", title+" "+subtitle, data)
	if err != nil {
		return err
	}

	subject := title + " " + subtitle + " - Nuruvent"
	text := s.buildVerificationText(data)

	return s.sendEmail(to, subject, html, text)
}

func (s *EmailService) getVerificationContent(purpose string, meta map[string]string, name string) (title, subtitle, description, message, extraInfo, warning string) {
	switch purpose {
	case "registration":
		title = "Verify Your"
		subtitle = "Account"
		description = "Complete your Nuruvent registration"
		message = "Thank you for joining Nuruvent. Please use the verification code below to complete your account setup."
		warning = "If you did not create an account on Nuruvent, please ignore this email."

	case "email_change":
		title = "Verify Your"
		subtitle = "Email Change"
		description = "Confirm your new email address"
		newEmail := meta["new_email"]
		message = "You requested to change your email address to " + newEmail + ". Please use the verification code below to confirm this change."
		warning = "If you did not request this change, please contact support immediately."

	case "phone_change":
		title = "Verify Your"
		subtitle = "Phone Change"
		description = "Confirm your new phone number"
		newPhone := meta["new_phone"]
		message = "You requested to change your phone number to " + newPhone + ". Please use the verification code below to confirm this change."
		warning = "If you did not request this change, please contact support immediately."

	case "password_reset":
		title = "Reset Your"
		subtitle = "Password"
		description = "Secure access to your account"
		message = "We received a request to reset your password for your Nuruvent account. Use the verification code below to continue."
		warning = "If you did not request a password reset, please ignore this email. Your account remains secure."

	case "two_factor":
		title = "Two-Factor"
		subtitle = "Authentication"
		description = "Secure your account access"
		message = "You requested a two-factor authentication code for your Nuruvent account. Please use the verification code below to complete your login."
		warning = "If you did not attempt to log in, please reset your password immediately."

	default:
		title = "Verify Your"
		subtitle = "Account"
		description = "Complete your verification"
		message = "Please use the verification code below to complete your verification."
	}

	return
}

// SendIndividualProfessionalWelcome sends welcome email for individual professionals
func (s *EmailService) SendIndividualProfessionalWelcome(to, name string) error {
    data := struct {
        Name string
    }{
        Name: name,
    }

    html, err := s.renderEmail("individual-professional-welcome", "Welcome to Nuruvent", data)
    if err != nil {
        return err
    }

    text := fmt.Sprintf(`Welcome to Nuruvent!

Hello %s,

Welcome to Nuruvent — the platform that empowers independent trainers, coaches, and consultants to host professional training events in Kenya.

Here is what you can do as an individual professional:
- Create and publish training events (workshops, webinars, bootcamps, meetups)
- Accept M-Pesa payments instantly — no manual reconciliation
- Issue QR-verified certificates to attendees
- Track attendance automatically via Zoom or Google Meet
- Get paid every Monday — only 10%% commission
- Save 3+ hours per event with automation
- Build your personal brand as a trainer

Ready to host your first event? Log in to your dashboard and start creating.

Light Your Events. Illuminate Your Growth.

--
Nuruvent - Light Your Events. Illuminate Your Growth.
`, name)

    return s.sendEmail(to, "Welcome to Nuruvent - Your Professional Account is Ready", html, text)
}

func (s *EmailService) buildVerificationText(data struct {
	Name        string
	OTP         string
	Expires     string
	Title       string
	Subtitle    string
	Description string
	Message     string
	ExtraInfo   string
	Warning     string
}) string {
	text := "Nuruvent - " + data.Title + " " + data.Subtitle + "\n\n"
	text += data.Description + "\n\n"
	text += "Hello " + data.Name + ",\n\n"
	text += data.Message + "\n\n"
	text += "Your verification code is: " + data.OTP + "\n\n"
	text += "This code expires in " + data.Expires + ".\n\n"

	if data.ExtraInfo != "" {
		text += data.ExtraInfo + "\n\n"
	}

	if data.Warning != "" {
		text += data.Warning + "\n\n"
	}

	text += "If you did not request this, please ignore this email.\n\n"
	text += "--\nNuruvent - Light Your Events. Illuminate Your Growth."

	return text
}

// ================================================
// WELCOME EMAILS
// ================================================

// SendWelcome sends user welcome email
func (s *EmailService) SendWelcome(to, name string) error {
	data := struct {
		Name string
	}{
		Name: name,
	}

	html, err := s.renderEmail("welcome", "Welcome to Nuruvent", data)
	if err != nil {
		return err
	}

	text := fmt.Sprintf(`Welcome to Nuruvent!

Hello %s,

We are thrilled to welcome you to Nuruvent — the platform that illuminates your path to professional growth. Your account has been successfully created.

Here is what you can do on Nuruvent:
- Discover professional events near you
- Register for workshops, webinars, and bootcamps
- Pay instantly with M-Pesa
- Earn verifiable certificates with QR codes
- Access 30-day replays after events
- Track your professional development

Nuru means light. We illuminate your path to professional growth.

--
Nuruvent - Light Your Events. Illuminate Your Growth.
`, name)

	return s.sendEmail(to, "Welcome to Nuruvent", html, text)
}

// SendBusinessWelcome sends business welcome email
func (s *EmailService) SendBusinessWelcome(to, businessName, ownerName string) error {
	data := struct {
		BusinessName string
		OwnerName    string
	}{
		BusinessName: businessName,
		OwnerName:    ownerName,
	}

	html, err := s.renderEmail("business-welcome", "Welcome to Nuruvent", data)
	if err != nil {
		return err
	}

	text := fmt.Sprintf(`Welcome to Nuruvent!

Hello %s,

Congratulations! Your business %s has been successfully registered on Nuruvent. You are now ready to host professional events and grow your brand.

Here is what you can do as a Nuruvent host:
- Create and publish events (workshops, webinars, bootcamps, meetups)
- Accept payments instantly with M-Pesa
- Issue QR-verified certificates to attendees
- Track attendance automatically via Zoom or Meet
- Get paid every Monday with 10%% commission
- Save 3+ hours per event with automation

Ready to host your first event? Log in to your dashboard and start creating.

Manage Your Events. Get Paid. Build Your Brand.

--
Nuruvent - Light Your Events. Illuminate Your Growth.
`, ownerName, businessName)

	return s.sendEmail(to, "Welcome to Nuruvent - Business Registration Complete", html, text)
}

// ================================================
// SECURITY EMAILS
// ================================================

// SendTwoFactorOTP sends two-factor authentication OTP
func (s *EmailService) SendTwoFactorOTP(to, name, otp, expires string) error {
	data := struct {
		Name    string
		OTP     string
		Expires string
	}{
		Name:    name,
		OTP:     otp,
		Expires: expires,
	}

	html, err := s.renderEmail("two-factor-otp", "Two-Factor Authentication", data)
	if err != nil {
		return err
	}

	text := fmt.Sprintf(`Nuruvent - Two-Factor Authentication

Hello %s,

You requested a two-factor authentication code for your Nuruvent account. Please use the verification code below to complete your login.

Your 2FA verification code is: %s

This code expires in %s.

If you did not attempt to log in, please reset your password immediately.

For security, never share this code with anyone.

--
Nuruvent - Light Your Events. Illuminate Your Growth.
`, name, otp, expires)

	return s.sendEmail(to, "Two-Factor Authentication - Nuruvent", html, text)
}

// SendPasswordResetOTP sends password reset OTP
func (s *EmailService) SendPasswordResetOTP(to, name, otp, expires string) error {
	return s.SendVerificationOTP(to, name, otp, expires, "password_reset", nil)
}

// SendBusinessVerificationOTP sends business verification OTP
func (s *EmailService) SendBusinessVerificationOTP(to, businessName, otp, expires string) error {
	meta := map[string]string{"business_name": businessName}
	return s.SendVerificationOTP(to, businessName, otp, expires, "business_verify", meta)
}

// SendPasswordResetConfirm sends password reset confirmation
func (s *EmailService) SendPasswordResetConfirm(to, name string) error {
	data := struct {
		Name string
	}{
		Name: name,
	}

	html, err := s.renderEmail("password-reset-confirm", "Password Reset Confirmation", data)
	if err != nil {
		return err
	}

	text := fmt.Sprintf(`Nuruvent - Password Reset Confirmation

Hello %s,

Your Nuruvent password has been successfully changed.

If you did not perform this action, please contact our support team immediately at hello@nuruvent.com

Your professional growth journey continues.

--
Nuruvent - Light Your Events. Illuminate Your Growth.
`, name)

	return s.sendEmail(to, "Password Reset Confirmation - Nuruvent", html, text)
}

// SendLoginNotification sends login notification
func (s *EmailService) SendLoginNotification(to, name, time, ipAddress, userAgent string) error {
	data := struct {
		Name      string
		Time      string
		IPAddress string
		UserAgent string
	}{
		Name:      name,
		Time:      time,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	html, err := s.renderEmail("login-notification", "New Login Detected", data)
	if err != nil {
		return err
	}

	text := fmt.Sprintf(`Nuruvent - New Login Detected

Hello %s,

We detected a new login to your Nuruvent account.

Time: %s
IP Address: %s
Device: %s

If this was you, you can safely ignore this notification.
If you did not log in, please reset your password immediately.

Light Your Events. Illuminate Your Growth.

--
Nuruvent - Light Your Events. Illuminate Your Growth.
`, name, time, ipAddress, userAgent)

	return s.sendEmail(to, "New Login Notification - Nuruvent", html, text)
}