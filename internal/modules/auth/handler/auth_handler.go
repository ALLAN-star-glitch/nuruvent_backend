package handler

import (
	"errors"
	"log"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/constants"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/service"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/response"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/validation"
	"github.com/gofiber/fiber/v3"
)

var validator = validation.New()

// Handler handles HTTP requests for auth
type Handler struct {
	service service.Service
	config  *config.Config
}

// NewHandler creates a new auth handler
func NewHandler(
	service service.Service,
	cfg *config.Config,
) *Handler {
	return &Handler{
		service: service,
		config:  cfg,
	}
}

// ================================================
// COOKIE HELPERS
// ================================================

func (h *Handler) setAccessTokenCookie(c fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(h.config.JWT.AccessExpiration),
		HTTPOnly: true,
		Secure:   h.config.Environment == "production",
		SameSite: "Lax",
		Path:     "/",
	})
}

func (h *Handler) setRefreshTokenCookie(c fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Expires:  time.Now().Add(h.config.JWT.RefreshExpiration),
		HTTPOnly: true,
		Secure:   h.config.Environment == "production",
		SameSite: "Lax",
		Path:     "/auth/refresh",
	})
}

func (h *Handler) clearAuthCookies(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   h.config.Environment == "production",
		SameSite: "Lax",
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   h.config.Environment == "production",
		SameSite: "Lax",
		Path:     "/auth/refresh",
	})
}

func (h *Handler) getRefreshTokenFromCookie(c fiber.Ctx) (string, error) {
	token := c.Cookies("refresh_token")
	if token == "" {
		return "", errors.New("refresh token not found")
	}
	return token, nil
}

// ================================================
// REGISTER HANDLER
// ================================================

// Register handles user registration
// @Summary Register a new account
// @Description Register a new account (personal or institution)
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 200 {object} response.BaseResponse{data=OTPResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 409 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	// Validate based on account type
	if err := h.validateRegisterRequest(&req); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	// Prepare service request
	svcReq := service.RegisterRequest{
		Email:       req.Email,
		Password:    req.Password,
		Name:        req.Name,
		Phone:       req.Phone,
		AccountType: req.AccountType,
	}

	// Add institution fields if provided
	if req.AccountType == "institution" {
		svcReq.InstitutionName = req.InstitutionName
		svcReq.InstitutionEmail = req.InstitutionEmail
		svcReq.InstitutionPhone = req.InstitutionPhone
		svcReq.InstitutionType = req.InstitutionType
	}

	if err := h.service.RegisterAccount(c.Context(), svcReq); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	return response.Success(c, "OTP sent successfully", OTPResponse{
		Email:     req.Email,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Message:   "OTP sent to your email. Verify to complete registration.",
	})
}

// validateRegisterRequest validates the registration request
func (h *Handler) validateRegisterRequest(req *RegisterRequest) error {
	// Validate email
	if req.Email == "" {
		return errors.New("email is required")
	}
	if !validator.Email.Validate(req.Email) {
		return errors.New("invalid email format")
	}

	// Validate password
	if valid, msg := validator.Password.Validate(req.Password); !valid {
		return errors.New(msg)
	}

	// Validate name
	if req.Name == "" {
		return errors.New("name is required")
	}

	// Validate account type
	if req.AccountType != "personal" && req.AccountType != "institution" {
		return errors.New("account_type must be 'personal' or 'institution'")
	}

	// Validate phone
	if req.Phone == "" {
		return errors.New("phone is required")
	}
	if !validator.Phone.Validate(req.Phone) {
		return errors.New("invalid Kenyan phone number. Use format: 254XXXXXXXXX, +254XXXXXXXXX, 07XXXXXXXX, 01XXXXXXXX, or 02XXXXXXXX")
	}
	req.Phone = validator.Phone.Normalize(req.Phone)

	// Validate institution fields
	if req.AccountType == "institution" {
		if req.InstitutionName == "" {
			return errors.New("institution_name is required for institution accounts")
		}
		if req.InstitutionEmail == "" {
			return errors.New("institution_email is required for institution accounts")
		}
		if !validator.Email.Validate(req.InstitutionEmail) {
			return errors.New("invalid institution email format")
		}
		if req.InstitutionPhone == "" {
			return errors.New("institution_phone is required for institution accounts")
		}
		if !validator.Phone.Validate(req.InstitutionPhone) {
			return errors.New("invalid institution phone number")
		}
		req.InstitutionPhone = validator.Phone.Normalize(req.InstitutionPhone)
		if req.InstitutionType == "" {
			return errors.New("institution_type is required for institution accounts")
		}
		if !constants.IsValidInstitutionType(req.InstitutionType) {
			return errors.New("invalid institution type")
		}
	}

	return nil
}

// ================================================
// VERIFY OTP HANDLER
// ================================================

// VerifyOTP verifies OTP and completes registration
// @Summary Verify OTP and complete registration
// @Description Verify the OTP sent to your email and complete account creation
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body VerifyOTPRequest true "OTP verification details"
// @Success 201 {object} response.BaseResponse{data=AuthResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/verify-otp [post]
func (h *Handler) VerifyOTP(c fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	account, result, err := h.service.VerifyOTPAndCreateAccount(c.Context(), req.Email, req.OTP)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	// Set cookies
	if accessToken, ok := result["access_token"].(string); ok && accessToken != "" {
		h.setAccessTokenCookie(c, accessToken)
	}
	if refreshToken, ok := result["refresh_token"].(string); ok && refreshToken != "" {
		h.setRefreshTokenCookie(c, refreshToken)
	}

	// Build response
	authResp := AuthResponse{
		TokenResponse: TokenResponse{
			AccessToken:  result["access_token"].(string),
			RefreshToken: result["refresh_token"].(string),
			TokenType:    "Bearer",
			ExpiresIn:    int64(h.config.JWT.AccessExpiration.Seconds()),
		},
		Account: NewAccountResponse(account),
	}

	if institution, ok := result["institution"].(*models.Institution); ok && institution != nil {
		authResp.Institution = NewInstitutionResponse(institution)
	}

	return response.Created(c, "Account created successfully", authResp)
}

// ================================================
// RESEND OTP HANDLER
// ================================================

// ResendOTP resends OTP to the user's email
// @Summary Resend OTP
// @Description Resend the verification OTP to the user's email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body ResendOTPRequest true "Email for OTP resend"
// @Success 200 {object} response.BaseResponse{data=OTPResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 409 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/resend-otp [post]
func (h *Handler) ResendOTP(c fiber.Ctx) error {
	var req ResendOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	otp := h.service.GenerateOTP()
	if err := h.service.StoreOTP(req.Email, otp); err != nil {
		return response.InternalError(c, "Failed to resend OTP", fiber.Map{"error": err.Error()})
	}

	// Send OTP email
	if err := h.service.SendOTPEmail(req.Email, "User", otp); err != nil {
		log.Printf("Failed to send OTP email: %v", err)
	}

	return response.Success(c, "OTP resent successfully", OTPResponse{
		Email:     req.Email,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Message:   "OTP resent to your email",
	})
}

// ================================================
// LOGIN HANDLER
// ================================================

// Login authenticates a user
// @Summary Login user
// @Description Authenticate a user with email and password, initiates 2FA
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} response.BaseResponse{data=TwoFactorResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	account, _, err := h.service.LoginAccount(c.Context(), req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		return response.Unauthorized(c, "Invalid credentials", fiber.Map{"error": err.Error()})
	}

	return response.Success(c, "2FA verification required", TwoFactorResponse{
		Requires2FA: true,
		Email:       account.Email,
		ExpiresIn:   300,
	})
}

// ================================================
// VERIFY TWO-FACTOR OTP HANDLER
// ================================================

// VerifyTwoFactorOTP verifies 2FA OTP and completes login
// @Summary Verify 2FA OTP and complete login
// @Description Verify the two-factor authentication OTP and complete the login process
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body VerifyTwoFactorOTPRequest true "2FA OTP verification"
// @Success 200 {object} response.BaseResponse{data=AuthResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/verify-2fa [post]
func (h *Handler) VerifyTwoFactorOTP(c fiber.Ctx) error {
	var req VerifyTwoFactorOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	account, accessToken, refreshToken, err := h.service.VerifyTwoFactorAndLogin(
		c.Context(),
		req.Email,
		req.OTP,
		ipAddress,
		userAgent,
	)
	if err != nil {
		return response.Unauthorized(c, err.Error(), nil)
	}

	h.setAccessTokenCookie(c, accessToken)
	h.setRefreshTokenCookie(c, refreshToken)

	return response.Success(c, "Login successful", AuthResponse{
		TokenResponse: TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(h.config.JWT.AccessExpiration.Seconds()),
		},
		Account: NewAccountResponse(account),
	})
}

// ================================================
// REFRESH TOKEN HANDLER
// ================================================

// RefreshToken refreshes access token
// @Summary Refresh access token
// @Description Refresh the access token using a valid refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.BaseResponse{data=TokenResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/refresh [post]
func (h *Handler) RefreshToken(c fiber.Ctx) error {
	var refreshToken string

	var req RefreshTokenRequest
	if err := c.Bind().Body(&req); err == nil && req.RefreshToken != "" {
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		cookieToken, err := h.getRefreshTokenFromCookie(c)
		if err == nil && cookieToken != "" {
			refreshToken = cookieToken
		}
	}

	if refreshToken == "" {
		return response.Unauthorized(c, "Refresh token required", nil)
	}

	newAccessToken, newRefreshToken, err := h.service.RefreshTokens(
		c.Context(),
		refreshToken,
		c.Get("User-Agent"),
		c.IP(),
	)
	if err != nil {
		return response.Unauthorized(c, "Invalid refresh token", fiber.Map{"error": err.Error()})
	}

	h.setAccessTokenCookie(c, newAccessToken)
	h.setRefreshTokenCookie(c, newRefreshToken)

	return response.Success(c, "Token refreshed successfully", TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.config.JWT.AccessExpiration.Seconds()),
	})
}

// ================================================
// LOGOUT HANDLER
// ================================================

// Logout logs out a user
// @Summary Logout user
// @Description Logout the current user and revoke the refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body LogoutRequest true "Logout request"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(c fiber.Ctx) error {
	var refreshToken string

	var req LogoutRequest
	if err := c.Bind().Body(&req); err == nil && req.RefreshToken != "" {
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		cookieToken, err := h.getRefreshTokenFromCookie(c)
		if err == nil && cookieToken != "" {
			refreshToken = cookieToken
		}
	}

	if refreshToken != "" {
		if err := h.service.RevokeToken(c.Context(), refreshToken); err != nil {
			log.Printf("Failed to revoke refresh token: %v", err)
		}
	}

	h.clearAuthCookies(c)
	return response.Success(c, "Logged out successfully", nil)
}

// ================================================
// FORGOT PASSWORD HANDLER
// ================================================

// ForgotPassword initiates password reset
// @Summary Initiate password reset
// @Description Send OTP to reset password for a user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Password reset request"
// @Success 200 {object} response.BaseResponse{data=PasswordResetResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/forgot-password [post]
func (h *Handler) ForgotPassword(c fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	if valid, msg := validator.Password.Validate(req.NewPassword); !valid {
		return response.BadRequest(c, "Weak password", fiber.Map{
			"field":    "new_password",
			"error":    msg,
			"strength": validator.Password.StrengthLabel(validator.Password.Score(req.NewPassword)),
		})
	}

	if err := h.service.InitiatePasswordReset(c.Context(), req.Email, req.NewPassword); err != nil {
		return response.InternalError(c, "Failed to process request", fiber.Map{"error": err.Error()})
	}

	return response.Success(c, "OTP sent to your email", PasswordResetResponse{
		Message:   "Check your email for the OTP",
		ExpiresIn: 300,
	})
}

// ================================================
// VERIFY RESET OTP HANDLER
// ================================================

// VerifyResetOTP verifies reset OTP and resets password
// @Summary Verify reset OTP and reset password
// @Description Verify the OTP and complete password reset
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body VerifyResetOTPRequest true "OTP verification"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/verify-reset-otp [post]
func (h *Handler) VerifyResetOTP(c fiber.Ctx) error {
	var req VerifyResetOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	if err := h.service.VerifyResetOTPAndResetPassword(c.Context(), req.Email, req.OTP); err != nil {
		return response.Unauthorized(c, err.Error(), nil)
	}

	return response.Success(c, "Password reset successfully", fiber.Map{
		"message": "Password reset successfully",
	})
}