// internal/modules/notification/service/worker.go

package service

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/hibiken/asynq"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
)

// NotificationWorker handles async notification tasks
type NotificationWorker struct {
	emailChannel notificationdomain.Channel
}

// NewNotificationWorker creates a new notification worker
func NewNotificationWorker(emailChannel notificationdomain.Channel) *NotificationWorker {
	return &NotificationWorker{
		emailChannel: emailChannel,
	}
}

// ============================================================
// UNIFIED VERIFICATION OTP HANDLER
// ============================================================

func (w *NotificationWorker) ProcessVerificationOTP(ctx context.Context, data notificationdomain.VerificationOTPTask) error {
	log.Printf("[NotificationWorker] Processing verification OTP for %s (purpose: %s)", data.To, data.Purpose)

	title, subtitle, description, message, warning := w.getVerificationContent(data)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: title + " " + subtitle + " - Nuruvent",
		Type:    notificationdomain.TypeVerificationOTP,
		Meta: map[string]string{
			"title":       title,
			"subtitle":    subtitle,
			"description": description,
			"message":     message,
			"warning":     warning,
			"name":        data.Name,
			"otp":         data.OTP,
			"expires":     data.Expires,
			"purpose":     string(data.Purpose),
		},
	}

	if data.Purpose == notificationdomain.PurposeTwoFactor {
		if ip, ok := data.Meta["ip_address"]; ok {
			channelReq.Meta["ip_address"] = ip
		}
		if ua, ok := data.Meta["user_agent"]; ok {
			channelReq.Meta["user_agent"] = ua
		}
	}

	if data.Purpose == notificationdomain.PurposeEmailChange {
		if newEmail, ok := data.Meta["new_email"]; ok {
			channelReq.Meta["new_email"] = newEmail
		}
	}

	if data.Purpose == notificationdomain.PurposePhoneChange {
		if newPhone, ok := data.Meta["new_phone"]; ok {
			channelReq.Meta["new_phone"] = newPhone
		}
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send verification OTP to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Verification OTP sent to %s (purpose: %s)", data.To, data.Purpose)
	return nil
}

func (w *NotificationWorker) getVerificationContent(data notificationdomain.VerificationOTPTask) (title, subtitle, description, message, warning string) {
	switch data.Purpose {
	case notificationdomain.PurposeRegistration:
		title = "Verify Your"
		subtitle = "Account"
		description = "Complete your Nuruvent registration"
		message = "Thank you for joining Nuruvent. Please use the verification code below to complete your account setup."
		warning = "If you did not create an account on Nuruvent, please ignore this email."

	case notificationdomain.PurposeEmailChange:
		title = "Verify Your"
		subtitle = "Email Change"
		description = "Confirm your new email address"
		newEmail := data.Meta["new_email"]
		message = "You requested to change your email address to " + newEmail + ". Please use the verification code below to confirm this change."
		warning = "If you did not request this change, please contact support immediately."

	case notificationdomain.PurposePhoneChange:
		title = "Verify Your"
		subtitle = "Phone Change"
		description = "Confirm your new phone number"
		newPhone := data.Meta["new_phone"]
		message = "You requested to change your phone number to " + newPhone + ". Please use the verification code below to confirm this change."
		warning = "If you did not request this change, please contact support immediately."

	case notificationdomain.PurposePasswordReset:
		title = "Reset Your"
		subtitle = "Password"
		description = "Secure access to your account"
		message = "We received a request to reset your password for your Nuruvent account. Use the verification code below to continue."
		warning = "If you did not request a password reset, please ignore this email. Your account remains secure."

	case notificationdomain.PurposeTwoFactor:
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

// HandleVerificationOTP is the asynq task handler
func (w *NotificationWorker) HandleVerificationOTP(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.VerificationOTPTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse verification OTP task: %v", err)
		return err
	}
	return w.ProcessVerificationOTP(ctx, data)
}

// ============================================================
// WELCOME HANDLERS
// ============================================================

// ProcessWelcomeIndividual implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessWelcomeIndividual(ctx context.Context, data notificationdomain.WelcomeIndividualTask) error {
	log.Printf("[NotificationWorker] Processing individual welcome for %s", data.To)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Welcome to Nuruvent - Your Professional Account is Ready!",
		Type:    notificationdomain.TypeWelcome,
		Meta: map[string]string{
			"name":         data.Name,
			"account_type": "individual",
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send individual welcome to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Individual welcome sent to %s", data.To)
	return nil
}

// HandleWelcomeIndividual is the asynq task handler
func (w *NotificationWorker) HandleWelcomeIndividual(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.WelcomeIndividualTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse individual welcome task: %v", err)
		return err
	}
	return w.ProcessWelcomeIndividual(ctx, data)
}

// ProcessWelcomeInstitution implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessWelcomeInstitution(ctx context.Context, data notificationdomain.WelcomeInstitutionTask) error {
	log.Printf("[NotificationWorker] Processing institution welcome for %s - Institution: %s", data.To, data.InstitutionName)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Welcome to Nuruvent - Your Institution is Live!",
		Type:    notificationdomain.TypeWelcomeInstitution,
		Meta: map[string]string{
			"admin_name":       data.AdminName,
			"institution_name": data.InstitutionName,
			"account_type":     "institution",
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send institution welcome to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Institution welcome sent to %s", data.To)
	return nil
}

// HandleWelcomeInstitution is the asynq task handler
func (w *NotificationWorker) HandleWelcomeInstitution(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.WelcomeInstitutionTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse institution welcome task: %v", err)
		return err
	}
	return w.ProcessWelcomeInstitution(ctx, data)
}

// ============================================================
// PASSWORD RESET CONFIRM HANDLER
// ============================================================

// ProcessPasswordResetConfirm implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessPasswordResetConfirm(ctx context.Context, data notificationdomain.PasswordResetConfirmTask) error {
	log.Printf("[NotificationWorker] Processing password reset confirmation for %s", data.To)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Password Reset Confirmation - Nuruvent",
		Type:    notificationdomain.TypePasswordResetConfirm,
		Meta: map[string]string{
			"name": data.Name,
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send password reset confirmation to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Password reset confirmation sent to %s", data.To)
	return nil
}

// HandlePasswordResetConfirm is the asynq task handler
func (w *NotificationWorker) HandlePasswordResetConfirm(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.PasswordResetConfirmTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse password reset confirm task: %v", err)
		return err
	}
	return w.ProcessPasswordResetConfirm(ctx, data)
}

// ============================================================
// LOGIN NOTIFICATION HANDLER
// ============================================================

// ProcessLoginNotification implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessLoginNotification(ctx context.Context, data notificationdomain.LoginNotificationTask) error {
	log.Printf("[NotificationWorker] Processing login notification for %s", data.To)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "New Login Notification - Nuruvent",
		Type:    notificationdomain.TypeLoginNotification,
		Meta: map[string]string{
			"name":       data.Name,
			"time":       data.Time,
			"ip_address": data.IPAddress,
			"user_agent": data.UserAgent,
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send login notification to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Login notification sent to %s", data.To)
	return nil
}

// HandleLoginNotification is the asynq task handler
func (w *NotificationWorker) HandleLoginNotification(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.LoginNotificationTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse login notification task: %v", err)
		return err
	}
	return w.ProcessLoginNotification(ctx, data)
}

// ============================================================
// INSTITUTION KYC WELCOME HANDLER
// ============================================================

// ProcessWelcomeInstitutionKYC implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessWelcomeInstitutionKYC(ctx context.Context, data notificationdomain.WelcomeInstitutionKYCTask) error {
	log.Printf("[NotificationWorker] Processing institution KYC welcome for %s - Institution: %s", data.To, data.InstitutionName)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Welcome to Nuruvent - Complete Your KYC Verification",
		Type:    notificationdomain.TypeWelcomeInstitutionKYC,
		Meta: map[string]string{
			"admin_name":       data.AdminName,
			"institution_name": data.InstitutionName,
			"institution_type": data.InstitutionType,
			"account_type":     "institution",
			"kyc_required":     "true",
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send institution KYC welcome to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Institution KYC welcome sent to %s", data.To)
	return nil
}

// HandleWelcomeInstitutionKYC is the asynq task handler
func (w *NotificationWorker) HandleWelcomeInstitutionKYC(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.WelcomeInstitutionKYCTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse institution KYC welcome task: %v", err)
		return err
	}
	return w.ProcessWelcomeInstitutionKYC(ctx, data)
}

// ============================================================
// ✅ NEW: NEW INSTITUTION ACCOUNT REGISTRATION HANDLER
// ============================================================

// ProcessNewInstitutionAccountRegistration implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessNewInstitutionAccountRegistration(ctx context.Context, data notificationdomain.NewInstitutionAccountRegistrationNotice) error {
	log.Printf("[NotificationWorker] Processing new institution account registration for admin: %s, institution: %s", 
		data.NewAccountAdminName, data.InstitutionName)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "New Institution Account Registration - Nuruvent",
		Type:    notificationdomain.TypeNewInstitutionAccountRegistration,
		Meta: map[string]string{
			"admin_name":       data.NewAccountAdminName,
			"institution_name": data.InstitutionName,
			"institution_type": data.InstitutionType,
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send new institution account registration to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] New institution account registration sent to %s", data.To)
	return nil
}

// HandleNewInstitutionAccountRegistration is the asynq task handler
func (w *NotificationWorker) HandleNewInstitutionAccountRegistration(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.NewInstitutionAccountRegistrationNotice
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse new institution account registration task: %v", err)
		return err
	}
	return w.ProcessNewInstitutionAccountRegistration(ctx, data)
}

// ============================================================
// ✅ NEW: NEW PERSONAL ACCOUNT REGISTRATION HANDLER
// ============================================================

// ProcessNewPersonalAccountRegistration implements notificationdomain.TaskProcessor
func (w *NotificationWorker) ProcessNewPersonalAccountRegistration(ctx context.Context, data notificationdomain.NewPersonalAccountRegistrationTask) error {
	log.Printf("[NotificationWorker] Processing new personal account registration for admin: %s", 
		data.NewAccountAdminName)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "New Personal Account Registration - Nuruvent",
		Type:    notificationdomain.TypeNewPersonalAccountRegistration,
		Meta: map[string]string{
			"name": data.NewAccountAdminName,
		},
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send new personal account registration to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] New personal account registration sent to %s", data.To)
	return nil
}

// HandleNewPersonalAccountRegistration is the asynq task handler
func (w *NotificationWorker) HandleNewPersonalAccountRegistration(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.NewPersonalAccountRegistrationTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse new personal account registration task: %v", err)
		return err
	}
	return w.ProcessNewPersonalAccountRegistration(ctx, data)
}

// ============================================================
// ✅ TEAM INVITATION HANDLERS (UPDATED - NO ROLE, NO OTP, AI-READY)
// ============================================================

// ProcessTeamInviteExistingUser implements notificationdomain.TaskProcessor
// Sent when: User already has a Nuruvent account
// ✅ NO ROLE - Roles are inherited from account level
// ✅ NO OTP - User clicks accept link to join
// ✅ AI-READY - PersonalizedContent field is available for future AI integration
func (w *NotificationWorker) ProcessTeamInviteExistingUser(ctx context.Context, data notificationdomain.TeamInviteExistingUserTask) error {
	log.Printf("[NotificationWorker] Processing team invite for existing user %s to join %s", data.To, data.TeamName)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "You've Been Invited to Join " + data.TeamName + " on Nuruvent",
		Type:    notificationdomain.TypeTeamInviteExistingUser,
		Meta: map[string]string{
			"user_name":   data.UserName,
			"invited_by":  data.InvitedBy,
			"team_name":   data.TeamName,
			"team_id":     data.TeamID,
			"accept_link": data.AcceptLink,
			"expires_in":  data.ExpiresIn,
		},
	}

	// ✅ AI-Ready: If personalized content exists, add it to the channel request
	if data.PersonalizedContent != nil {
		channelReq.Meta["ai_subject"] = data.PersonalizedContent.Subject
		channelReq.Meta["ai_greeting"] = data.PersonalizedContent.Greeting
		channelReq.Meta["ai_intro"] = data.PersonalizedContent.Intro
		channelReq.Meta["ai_body"] = data.PersonalizedContent.Body
		channelReq.Meta["ai_benefits"] = strings.Join(data.PersonalizedContent.Benefits, "|")
		channelReq.Meta["ai_call_to_action"] = data.PersonalizedContent.CallToAction
		channelReq.Meta["ai_closing"] = data.PersonalizedContent.Closing
		channelReq.Meta["ai_pss"] = data.PersonalizedContent.PSS
		log.Printf("[NotificationWorker] ✅ AI content included for %s", data.To)
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send team invite to existing user %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Team invite sent to existing user %s", data.To)
	return nil
}

// HandleTeamInviteExistingUser is the asynq task handler
func (w *NotificationWorker) HandleTeamInviteExistingUser(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.TeamInviteExistingUserTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse team invite existing user task: %v", err)
		return err
	}
	return w.ProcessTeamInviteExistingUser(ctx, data)
}

// ProcessTeamInviteRegistration implements notificationdomain.TaskProcessor
// Sent when: User does NOT have a Nuruvent account
// ✅ NO ROLE - Roles are inherited from account level
// ✅ NO OTP - User clicks registration link with token embedded
// ✅ AI-READY - PersonalizedContent field is available for future AI integration
func (w *NotificationWorker) ProcessTeamInviteRegistration(ctx context.Context, data notificationdomain.TeamInviteRegistrationTask) error {
	log.Printf("[NotificationWorker] Processing team invite registration for new user %s to join %s", data.To, data.TeamName)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "You're Invited to Join " + data.TeamName + " on Nuruvent",
		Type:    notificationdomain.TypeTeamInviteRegistration,
		Meta: map[string]string{
			"name":              data.Name,
			"invited_by":        data.InvitedBy,
			"team_name":         data.TeamName,
			"team_id":           data.TeamID,
			"registration_link": data.RegistrationLink,
			"expires_in":        data.ExpiresIn,
		},
	}

	// ✅ AI-Ready: If personalized content exists, add it to the channel request
	if data.PersonalizedContent != nil {
		channelReq.Meta["ai_subject"] = data.PersonalizedContent.Subject
		channelReq.Meta["ai_greeting"] = data.PersonalizedContent.Greeting
		channelReq.Meta["ai_intro"] = data.PersonalizedContent.Intro
		channelReq.Meta["ai_body"] = data.PersonalizedContent.Body
	    channelReq.Meta["ai_benefits"] = strings.Join(data.PersonalizedContent.Benefits, "|")
		channelReq.Meta["ai_call_to_action"] = data.PersonalizedContent.CallToAction
		channelReq.Meta["ai_closing"] = data.PersonalizedContent.Closing
		channelReq.Meta["ai_pss"] = data.PersonalizedContent.PSS
		log.Printf("[NotificationWorker] ✅ AI content included for new user %s", data.To)
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send team invite registration to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Team invite registration sent to %s", data.To)
	return nil
}

// HandleTeamInviteRegistration is the asynq task handler
func (w *NotificationWorker) HandleTeamInviteRegistration(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.TeamInviteRegistrationTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse team invite registration task: %v", err)
		return err
	}
	return w.ProcessTeamInviteRegistration(ctx, data)
}

// ProcessTeamInviteAccepted implements notificationdomain.TaskProcessor
// Sent to: Admin who sent the invitation
// ✅ AI-READY - PersonalizedContent field is available for future AI integration
func (w *NotificationWorker) ProcessTeamInviteAccepted(ctx context.Context, data notificationdomain.TeamInviteAcceptedTask) error {
	log.Printf("[NotificationWorker] Processing team invite accepted notification for admin: %s, user: %s joined %s", 
		data.AdminName, data.UserName, data.TeamName)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Team Invitation Accepted - " + data.UserName + " joined " + data.TeamName,
		Type:    notificationdomain.TypeTeamInviteAccepted,
		Meta: map[string]string{
			"admin_name": data.AdminName,
			"user_name":  data.UserName,
			"user_email": data.UserEmail,
			"team_name":  data.TeamName,
			"team_id":    data.TeamID,
		},
	}

	// ✅ AI-Ready: If personalized content exists, add it to the channel request
	if data.PersonalizedContent != nil {
		channelReq.Meta["ai_subject"] = data.PersonalizedContent.Subject
		channelReq.Meta["ai_greeting"] = data.PersonalizedContent.Greeting
		channelReq.Meta["ai_body"] = data.PersonalizedContent.Body
		channelReq.Meta["ai_closing"] = data.PersonalizedContent.Closing
		log.Printf("[NotificationWorker] ✅ AI content included for admin notification")
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send team invite accepted notification to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Team invite accepted notification sent to %s", data.To)
	return nil
}

// HandleTeamInviteAccepted is the asynq task handler
func (w *NotificationWorker) HandleTeamInviteAccepted(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.TeamInviteAcceptedTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse team invite accepted task: %v", err)
		return err
	}
	return w.ProcessTeamInviteAccepted(ctx, data)
}

// ProcessTeamInviteDeclined implements notificationdomain.TaskProcessor
// Sent to: Admin who sent the invitation
// ✅ AI-READY - PersonalizedContent field is available for future AI integration
func (w *NotificationWorker) ProcessTeamInviteDeclined(ctx context.Context, data notificationdomain.TeamInviteDeclinedTask) error {
	log.Printf("[NotificationWorker] Processing team invite declined notification for admin: %s, user: %s declined %s", 
		data.AdminName, data.UserName, data.TeamName)

	channelReq := notificationdomain.ChannelRequest{
		To:      data.To,
		Subject: "Team Invitation Declined - " + data.UserName + " declined " + data.TeamName,
		Type:    notificationdomain.TypeTeamInviteDeclined,
		Meta: map[string]string{
			"admin_name": data.AdminName,
			"user_name":  data.UserName,
			"user_email": data.UserEmail,
			"team_name":  data.TeamName,
			"team_id":    data.TeamID,
		},
	}

	// ✅ AI-Ready: If personalized content exists, add it to the channel request
	if data.PersonalizedContent != nil {
		channelReq.Meta["ai_subject"] = data.PersonalizedContent.Subject
		channelReq.Meta["ai_greeting"] = data.PersonalizedContent.Greeting
		channelReq.Meta["ai_body"] = data.PersonalizedContent.Body
		channelReq.Meta["ai_closing"] = data.PersonalizedContent.Closing
		log.Printf("[NotificationWorker] ✅ AI content included for admin notification")
	}

	if err := w.emailChannel.Send(ctx, channelReq); err != nil {
		log.Printf("[NotificationWorker] Failed to send team invite declined notification to %s: %v", data.To, err)
		return err
	}

	log.Printf("[NotificationWorker] Team invite declined notification sent to %s", data.To)
	return nil
}

// HandleTeamInviteDeclined is the asynq task handler
func (w *NotificationWorker) HandleTeamInviteDeclined(ctx context.Context, task *asynq.Task) error {
	var data notificationdomain.TeamInviteDeclinedTask
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		log.Printf("[NotificationWorker] Failed to parse team invite declined task: %v", err)
		return err
	}
	return w.ProcessTeamInviteDeclined(ctx, data)
}

// Ensure NotificationWorker implements notificationdomain.TaskProcessor
var _ notificationdomain.TaskProcessor = (*NotificationWorker)(nil)