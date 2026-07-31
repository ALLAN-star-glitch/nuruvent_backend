package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/constants"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/response"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/validation"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var validator = validation.New()

// ================================================
// REQUEST MODELS
// ================================================

type RegisterRequest struct {
	// Personal fields (for all users)
	Phone    string `json:"phone,omitempty" example:"+254712345678"`
	Email    string `json:"email,omitempty" example:"john@example.com"` // Changed: optional for organizations
	Name     string `json:"name,omitempty" example:"John Doe"` // Changed: optional for organizations
	Password string `json:"password" binding:"required,min=8" example:"SecurePass123!"`
	Role     string `json:"role" binding:"omitempty,oneof=business_admin attendee" example:"attendee"`

	// Business fields (for business_admin - Formal Individual & Organization)
	BusinessEmail  string `json:"business_email" example:"info@business.com"`
	BusinessPhone  string `json:"business_phone" example:"+254700000000"`
	BusinessTypeID string `json:"business_type_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BusinessType   string `json:"business_type" example:"training_institute"`
	BusinessName   string `json:"business_name" example:"Nuruvent Training Institute"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
	OTP   string `json:"otp" binding:"required,len=6" example:"123456"`
}

type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
}

type VerifyTwoFactorOTPRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
	OTP   string `json:"otp" binding:"required,len=6" example:"123456"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"abc123xyz789..."`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" example:"abc123xyz789..."`
}

type ForgotPasswordRequest struct {
	Email       string `json:"email" binding:"required,email" example:"john@example.com"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"SecurePass123!"`
}

type VerifyResetOTPRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
	OTP   string `json:"otp" binding:"required,len=6" example:"123456"`
}

// ================================================
// RESPONSE MODELS
// ================================================

type AuthResponse struct {
	AccessToken  string   `json:"access_token,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	User         UserInfo `json:"user"`
}

// ================================================
// HANDLER STRUCT
// ================================================

type Handler struct {
	service      *Service
	repo         *Repository
	config       *config.Config
	tokenService *TokenService
}

func NewHandler(
	service *Service,
	repo *Repository,
	cfg *config.Config,
	tokenService *TokenService,
) *Handler {
	return &Handler{
		service:      service,
		repo:         repo,
		config:       cfg,
		tokenService: tokenService,
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
// @Summary Register a new user
// @Description Register a new user account with optional business creation for business_admin role
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 409 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	// Determine user type
	isFormalIndividual := req.BusinessType == "individual_formal"
	isInformalIndividual := req.BusinessType == "individual_informal"
	isOrganization := req.Role == string(authorization.RoleBusinessAdmin) && !isFormalIndividual && !isInformalIndividual

	// Validate email based on user type
	if req.Role != string(authorization.RoleBusinessAdmin) || isInformalIndividual {
		// Attendee or Informal Individual - personal email is required
		if req.Email == "" {
			return response.BadRequest(c, "Email is required", fiber.Map{"field": "email"})
		}
		if !validator.Email.Validate(req.Email) {
			return response.BadRequest(c, "Invalid email format", fiber.Map{"field": "email"})
		}
	} else if isFormalIndividual || isOrganization {
		// Formal Individual or Organization - personal email is optional
		if req.Email != "" && !validator.Email.Validate(req.Email) {
			return response.BadRequest(c, "Invalid email format", fiber.Map{"field": "email"})
		}
	}

	// Validate phone based on user type
	if req.Role != string(authorization.RoleBusinessAdmin) || isInformalIndividual {
		// Attendee or Informal Individual - phone is required
		if req.Phone == "" {
			return response.BadRequest(c, "Phone number is required", fiber.Map{"field": "phone"})
		}
		if !validator.Phone.Validate(req.Phone) {
			return response.BadRequest(c, "Invalid Kenyan phone number", fiber.Map{
				"field":  "phone",
				"format": "Use format: 254XXXXXXXXX, +254XXXXXXXXX, 07XXXXXXXX, 01XXXXXXXX, or 02XXXXXXXX",
			})
		}
		req.Phone = validator.Phone.Normalize(req.Phone)
	} else if req.Phone != "" && !validator.Phone.Validate(req.Phone) {
		// Validate phone if provided for other roles
		return response.BadRequest(c, "Invalid Kenyan phone number", fiber.Map{
			"field":  "phone",
			"format": "Use format: 254XXXXXXXXX, +254XXXXXXXXX, 07XXXXXXXX, 01XXXXXXXX, or 02XXXXXXXX",
		})
	}
	if req.Phone != "" {
		req.Phone = validator.Phone.Normalize(req.Phone)
	}

	// Validate password
	if valid, msg := validator.Password.Validate(req.Password); !valid {
		return response.BadRequest(c, "Weak password", fiber.Map{
			"field":    "password",
			"error":    msg,
			"strength": validator.Password.StrengthLabel(validator.Password.Score(req.Password)),
		})
	}

	// Set default role
	if req.Role == "" {
		req.Role = string(authorization.RoleAttendee)
	}

	// Validate role using authorization package
	if !authorization.IsValidRole(req.Role) {
		return response.BadRequest(c, "Invalid role", fiber.Map{
			"allowed_roles": authorization.GetAllRoles(),
			"provided":      req.Role,
		})
	}

	// Validate business fields if role is business_admin
	if req.Role == string(authorization.RoleBusinessAdmin) {
		if isOrganization {
			// Organization - full business details required
			if req.BusinessName == "" {
				return response.BadRequest(c, "Business name is required for organizations", fiber.Map{"field": "business_name"})
			}
			if req.BusinessEmail == "" {
				return response.BadRequest(c, "Business email is required for organizations", fiber.Map{"field": "business_email"})
			}
			if !validator.Email.Validate(req.BusinessEmail) {
				return response.BadRequest(c, "Invalid business email", fiber.Map{"field": "business_email"})
			}
			if req.BusinessPhone == "" {
				return response.BadRequest(c, "Business phone is required for organizations", fiber.Map{"field": "business_phone"})
			}
			if !validator.Phone.Validate(req.BusinessPhone) {
				return response.BadRequest(c, "Invalid business phone number", fiber.Map{
					"field":  "business_phone",
					"format": "Use format: 254XXXXXXXXX, +254XXXXXXXXX, 07XXXXXXXX, 01XXXXXXXX, or 02XXXXXXXX",
				})
			}
			req.BusinessPhone = validator.Phone.Normalize(req.BusinessPhone)

			// Validate business type (either ID or name)
			if req.BusinessTypeID == "" && req.BusinessType == "" {
				return response.BadRequest(c, "Business type is required for organizations", fiber.Map{"field": "business_type"})
			}

			if req.BusinessTypeID != "" {
				if _, err := uuid.Parse(req.BusinessTypeID); err != nil {
					return response.BadRequest(c, "Invalid business type ID", fiber.Map{"field": "business_type_id"})
				}
			}

			if req.BusinessType != "" && req.BusinessTypeID == "" {
				if !constants.IsValidBusinessType(req.BusinessType) {
					return response.BadRequest(c, "Invalid business type", fiber.Map{
						"field":         "business_type",
						"allowed_types": constants.AllBusinessTypeValues(),
					})
				}
			}
		} else if isFormalIndividual {
			// Formal Individual - business name and email required
			if req.BusinessName == "" {
				return response.BadRequest(c, "Business name is required for formal individual professionals", fiber.Map{"field": "business_name"})
			}
			if req.BusinessEmail == "" {
				return response.BadRequest(c, "Business email is required for formal individual professionals", fiber.Map{"field": "business_email"})
			}
			if !validator.Email.Validate(req.BusinessEmail) {
				return response.BadRequest(c, "Invalid business email", fiber.Map{"field": "business_email"})
			}
			if req.BusinessPhone == "" {
				return response.BadRequest(c, "Business phone is required for formal individual professionals", fiber.Map{"field": "business_phone"})
			}
			if !validator.Phone.Validate(req.BusinessPhone) {
				return response.BadRequest(c, "Invalid business phone number", fiber.Map{
					"field":  "business_phone",
					"format": "Use format: 254XXXXXXXXX, +254XXXXXXXXX, 07XXXXXXXX, 01XXXXXXXX, or 02XXXXXXXX",
				})
			}
			req.BusinessPhone = validator.Phone.Normalize(req.BusinessPhone)

			if req.BusinessType == "" {
				return response.BadRequest(c, "Business type is required", fiber.Map{"field": "business_type"})
			}
			if !constants.IsValidBusinessType(req.BusinessType) {
				return response.BadRequest(c, "Invalid business type", fiber.Map{
					"field":         "business_type",
					"allowed_types": constants.AllBusinessTypeValues(),
				})
			}
		} else if isInformalIndividual {
			// Informal Individual - personal email and phone required (already validated above)
			if req.BusinessType == "" {
				return response.BadRequest(c, "Business type is required", fiber.Map{"field": "business_type"})
			}
			if !constants.IsValidBusinessType(req.BusinessType) {
				return response.BadRequest(c, "Invalid business type", fiber.Map{
					"field":         "business_type",
					"allowed_types": constants.AllBusinessTypeValues(),
				})
			}
			log.Printf("Informal Individual registration: %s", req.Email)
		} else {
			return response.BadRequest(c, "Invalid business type", fiber.Map{
				"field":         "business_type",
				"allowed_types": constants.AllBusinessTypeValues(),
			})
		}
	} else {
		// Attendee - personal email and phone required (already validated above)
		if req.BusinessType != "" {
			return response.BadRequest(c, "Business type is not allowed for attendees", fiber.Map{"field": "business_type"})
		}
	}

	// Prepare user data
	userData := map[string]string{
		"email":    req.Email,
		"phone":    req.Phone,
		"name":     req.Name,
		"password": req.Password,
		"role":     req.Role,
	}

	if req.Role == string(authorization.RoleBusinessAdmin) {
		if req.BusinessTypeID != "" {
			userData["business_type_id"] = req.BusinessTypeID
		} else if req.BusinessType != "" {
			userData["business_type"] = req.BusinessType
		}
		userData["business_name"] = req.BusinessName
		userData["business_email"] = req.BusinessEmail
		userData["business_phone"] = req.BusinessPhone
	}

	log.Printf("DEBUG userData being sent to service: %+v", userData)

	ctx := context.Background()
	if err := h.service.RegisterUser(ctx, userData); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	respData := fiber.Map{
		"email":      req.Email,
		"phone":      req.Phone,
		"role":       req.Role,
		"expires_at": time.Now().Add(5 * time.Minute),
		"message":    "OTP sent successfully. Verify to complete registration.",
	}

	if req.Role == string(authorization.RoleBusinessAdmin) {
		if isOrganization {
			respData["message"] = "OTP sent to your business email. Verify to complete registration."
			respData["business_email"] = req.BusinessEmail
		} else if isFormalIndividual {
			respData["message"] = "OTP sent to your business email. Verify to complete your professional registration."
			respData["business_email"] = req.BusinessEmail
			respData["is_individual"] = true
			respData["is_formal"] = true
		} else if isInformalIndividual {
			respData["message"] = "OTP sent to your personal email. Verify to complete your professional registration."
			respData["is_individual"] = true
			respData["is_formal"] = false
		}
	}

	return response.Success(c, "OTP sent successfully", respData)
}

// ================================================
// VERIFY OTP HANDLER
// ================================================

// VerifyOTP verifies OTP and creates account
// @Summary Verify OTP and complete registration
// @Description Verify the OTP sent to your email and complete account creation
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body VerifyOTPRequest true "OTP verification details"
// @Success 201 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/verify-otp [post]
func (h *Handler) VerifyOTP(c fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	ctx := context.Background()
	user, result, err := h.service.VerifyOTPAndCreateUser(ctx, req.Email, req.OTP)
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
	responseData := fiber.Map{
		"token_type": "Bearer",
		"expires_in": int64(h.config.JWT.AccessExpiration.Seconds()),
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"phone": user.Phone,
			"name":  user.Name,
			"role":  user.Role,
		},
	}

	// Merge result data
	for key, value := range result {
		if key != "user" && key != "business" {
			responseData[key] = value
		}
	}

	// Check if this is an individual professional
	if business, ok := result["business"].(*models.Business); ok && business != nil {
		if business.BusinessType != nil {
			businessTypeName := business.BusinessType.Name
			if businessTypeName == "individual_formal" ||
				businessTypeName == "individual_informal" ||
				businessTypeName == "individual" {
				responseData["is_individual"] = true
			}
			if businessTypeName == "individual_formal" {
				responseData["is_formal"] = true
			}
			if businessTypeName == "individual_informal" {
				responseData["is_formal"] = false
			}
		}
	}

	return response.Created(c, "Account created successfully", responseData)
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
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 409 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/resend-otp [post]
func (h *Handler) ResendOTP(c fiber.Ctx) error {
	var req ResendOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	// Check if user exists
	existingUser, _ := h.repo.GetUserByEmail(req.Email)
	if existingUser != nil {
		return response.Conflict(c, "Email already registered", fiber.Map{"field": "email", "value": req.Email})
	}

	newOTP := h.service.GenerateOTP()
	if err := h.service.StoreOTP(req.Email, newOTP); err != nil {
		return response.InternalError(c, "Failed to resend OTP", fiber.Map{"error": err.Error()})
	}

	if err := h.service.StoreUserData(req.Email, map[string]interface{}{"refresh": time.Now().Unix()}); err != nil {
		log.Printf("Failed to refresh user data TTL: %v", err)
	}

	// Send OTP email
	if err := h.service.SendOTPEmail(req.Email, "User", newOTP); err != nil {
		log.Printf("Failed to resend OTP email: %v", err)
	}

	return response.Success(c, "OTP resent successfully", fiber.Map{
		"email":      req.Email,
		"expires_at": time.Now().Add(5 * time.Minute),
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
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	ctx := context.Background()

	// Capture IP and User-Agent for login notification
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	user, _, err := h.service.LoginUser(ctx, req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		return response.Unauthorized(c, "Invalid credentials", fiber.Map{"error": err.Error()})
	}

	return response.Success(c, "2FA verification required", fiber.Map{
		"requires_2fa": true,
		"email":        user.Email,
		"expires_in":   300,
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

	ctx := context.Background()

	// Capture IP and User-Agent for login notification
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	user, accessToken, refreshToken, err := h.service.VerifyTwoFactorAndLogin(ctx, req.Email, req.OTP, ipAddress, userAgent)
	if err != nil {
		return response.Unauthorized(c, err.Error(), nil)
	}

	h.setAccessTokenCookie(c, accessToken)
	h.setRefreshTokenCookie(c, refreshToken)

	userInfo := h.service.BuildUserInfo(user)

	return response.Success(c, "Login successful", fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    int64(h.config.JWT.AccessExpiration.Seconds()),
		"user":          userInfo,
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
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
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

	ctx := context.Background()
	newAccessToken, newRefreshToken, err := h.service.RefreshTokens(ctx, refreshToken, c.Get("User-Agent"), c.IP())
	if err != nil {
		return response.Unauthorized(c, "Invalid refresh token", fiber.Map{"error": err.Error()})
	}

	h.setAccessTokenCookie(c, newAccessToken)
	h.setRefreshTokenCookie(c, newRefreshToken)

	return response.Success(c, "Token refreshed successfully", fiber.Map{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    int64(h.config.JWT.AccessExpiration.Seconds()),
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
		ctx := context.Background()
		if err := h.service.RevokeToken(ctx, refreshToken); err != nil {
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
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
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

	ctx := context.Background()
	if err := h.service.InitiatePasswordReset(ctx, req.Email, req.NewPassword); err != nil {
		return response.InternalError(c, "Failed to process request", fiber.Map{"error": err.Error()})
	}

	return response.Success(c, "OTP sent to your email", fiber.Map{
		"message":    "Check your email for the OTP",
		"expires_in": 300,
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
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/auth/verify-reset-otp [post]
func (h *Handler) VerifyResetOTP(c fiber.Ctx) error {
	var req VerifyResetOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{"error": err.Error()})
	}

	ctx := context.Background()
	if err := h.service.VerifyResetOTPAndResetPassword(ctx, req.Email, req.OTP); err != nil {
		return response.Unauthorized(c, err.Error(), nil)
	}

	return response.Success(c, "Password reset successfully", fiber.Map{
		"message": "Password reset successfully",
	})
}