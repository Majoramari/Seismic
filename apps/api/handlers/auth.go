package handlers

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/models"
	"github.com/majoramari/seismic/apps/api/services"
)

type AuthHandler struct {
	Pool          *pgxpool.Pool
	EmailCfg      services.EmailConfig
	JWTSecret     string
	CloudinaryCfg services.CloudinaryConfig
}

type magicLinkRequest struct {
	Email string `json:"email"`
}

type requestEmailChangeInput struct {
	NewEmail string `json:"newEmail"`
}

type updateProfileRequest struct {
	Username     string  `json:"username"`
	DisplayName  string  `json:"displayName"`
	Bio          string  `json:"bio"`
	Website      string  `json:"website"`
	ContactEmail string  `json:"contactEmail"`
	ProfileImage *string `json:"profileImage"`
}

var reservedUsernames = map[string]bool{
	"p":           true,
	"login":       true,
	"verify":      true,
	"settings":    true,
	"dashboard":   true,
	"leaderboard": true,
	"api":         true,
	"admin":       true,
	"auth":        true,
	"about":       true,
	"help":        true,
	"support":     true,
}

// isReservedUsername checks if a username is reserved for
// application routes/pages and can't be used by real accounts.
func isReservedUsername(username string) bool {
	return reservedUsernames[username]
}

// RequestMagicLink godoc
// @Summary      Request a magic link
// @Description  Sends a login link to the given email. Works for both new and existing users.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body magicLinkRequest true "Email address"
// @Success      200 {object} helpers.APIResponse
// @Failure      400 {object} helpers.APIResponse
// @Router       /api/auth/magic-link [post]
func (h *AuthHandler) RequestMagicLink(c *fiber.Ctx) error {
	var body magicLinkRequest
	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	email := helpers.NormalizeEmail(body.Email)
	if email == "" || !strings.Contains(email, "@") {
		return helpers.Error(c, fiber.StatusBadRequest, "Please provide a valid email address")
	}

	ctx := c.Context()

	link, err := models.CreateMagicLink(ctx, h.Pool, email)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to create login link")
	}

	err = services.SendMagicLinkEmail(h.EmailCfg, email, link.Token)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to send login email")
	}

	return helpers.Success(c, "Check your email for a login link", nil)
}

// VerifyMagicLink godoc
// @Summary      Verify a magic link token
// @Description  Validates the token from the login email. Logs in existing users, or returns a signup token for new users.
// @Tags         auth
// @Produce      json
// @Param        token query string true "Magic link token"
// @Success      200 {object} helpers.APIResponse
// @Failure      400 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Router       /api/auth/verify [get]
func (h *AuthHandler) VerifyMagicLink(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return helpers.Error(c, fiber.StatusBadRequest, "Missing token")
	}

	ctx := c.Context()

	link, err := models.FindMagicLinkByToken(ctx, h.Pool, token)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
	if link == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "Invalid login link")
	}
	if link.Used {
		return helpers.Error(c, fiber.StatusUnauthorized, "This login link has already been used")
	}
	if link.IsExpired() {
		return helpers.Error(c, fiber.StatusUnauthorized, "This login link has expired")
	}

	if err := models.MarkMagicLinkUsed(ctx, h.Pool, link.ID); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}

	user, err := models.FindUserByEmail(ctx, h.Pool, link.Email)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}

	// New user, needs to pick a username before account is created
	if user == nil {
		signupToken, err := generateSignupToken(link.Email, h.JWTSecret)
		if err != nil {
			return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
		}
		return helpers.Success(c, "New account, complete signup", fiber.Map{
			"newUser":     true,
			"signupToken": signupToken,
			"email":       link.Email,
		})
	}

	// Existing user, log in normally
	accessToken, err := generateJWT(user.ID, h.JWTSecret)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to create session")
	}

	refreshToken, err := models.CreateRefreshToken(ctx, h.Pool, user.ID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to create session")
	}
	setRefreshTokenCookie(c, refreshToken, strings.HasPrefix(h.EmailCfg.AppURL, "https"))

	return helpers.Success(c, "Logged in successfully", fiber.Map{
		"newUser":     false,
		"accessToken": accessToken,
		"user":        user,
	})
}

type completeSignupRequest struct {
	SignupToken string `json:"signupToken"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

// CompleteSignup godoc
// @Summary      Complete signup for a new user
// @Description  Creates the account using the signup token from VerifyMagicLink, once a username is chosen.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body completeSignupRequest true "Signup details"
// @Success      200 {object} helpers.APIResponse
// @Failure      400 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Failure      409 {object} helpers.APIResponse
// @Router       /api/auth/complete-signup [post]
func (h *AuthHandler) CompleteSignup(c *fiber.Ctx) error {
	var body completeSignupRequest
	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	email, err := verifySignupToken(body.SignupToken, h.JWTSecret)
	if err != nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "Invalid or expired signup token")
	}

	username := strings.TrimSpace(strings.ToLower(body.Username))
	if username == "" {
		return helpers.Error(c, fiber.StatusBadRequest, "Username is required")
	}
	if !isValidUsername(username) {
		return helpers.Error(c, fiber.StatusBadRequest, "Username must be 3-20 characters, start with a letter, and contain only lowercase letters, numbers, underscore, or hyphen")
	}
	if isReservedUsername(username) {
		return helpers.Error(c, fiber.StatusBadRequest, "This username is reserved")
	}

	ctx := c.Context()

	existing, _ := models.FindUserByUsername(ctx, h.Pool, username)
	if existing != nil {
		return helpers.Error(c, fiber.StatusConflict, "Username is already taken")
	}

	user, err := models.CreateUser(ctx, h.Pool, email, username, body.DisplayName)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to create account")
	}

	accessToken, err := generateJWT(user.ID, h.JWTSecret)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to create session")
	}

	refreshToken, err := models.CreateRefreshToken(ctx, h.Pool, user.ID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to create session")
	}
	setRefreshTokenCookie(c, refreshToken, strings.HasPrefix(h.EmailCfg.AppURL, "https"))

	return helpers.Success(c, "Account created", fiber.Map{
		"accessToken": accessToken,
		"user":        user,
	})
}

// RefreshAccessToken godoc
// @Summary      Refresh access token
// @Description  Uses the refresh token cookie to issue a new access token and rotate the refresh token.
// @Tags         auth
// @Produce      json
// @Success      200 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Router       /api/auth/refresh [post]
func (h *AuthHandler) RefreshAccessToken(c *fiber.Ctx) error {
	rawToken := c.Cookies("refresh_token")
	if rawToken == "" {
		return helpers.Error(c, fiber.StatusUnauthorized, "No refresh token provided")
	}

	ctx := c.Context()

	stored, err := models.FindValidRefreshToken(ctx, h.Pool, rawToken)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
	if stored == nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "Invalid or expired refresh token")
	}

	if err := models.RevokeRefreshToken(ctx, h.Pool, stored.ID); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}

	newRawToken, err := models.CreateRefreshToken(ctx, h.Pool, stored.UserID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
	setRefreshTokenCookie(c, newRawToken, strings.HasPrefix(h.EmailCfg.AppURL, "https"))

	accessToken, err := generateJWT(stored.UserID, h.JWTSecret)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}

	return helpers.Success(c, "Token refreshed", fiber.Map{
		"accessToken": accessToken,
	})
}

// setRefreshTokenCookie sets the refresh token as an httpOnly
// cookie so client-side JavaScript can never read it directly.
// Secure is disabled for local http testing, enabled in
// production where everything runs over https.
func setRefreshTokenCookie(c *fiber.Ctx, token string, secure bool) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "None",
		Path:     "/api/auth",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	})
}

// GetAPIKey godoc
// @Summary      Get API key
// @Description  Returns the logged-in user's API key, used to configure editor plugins.
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Router       /api/auth/apikey [get]
func (h *AuthHandler) GetAPIKey(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	ctx := c.Context()
	user, err := models.FindUserByID(ctx, h.Pool, userID)
	if err != nil || user == nil {
		return helpers.Error(c, fiber.StatusNotFound, "User not found")
	}

	return helpers.Success(c, "API key retrieved", fiber.Map{
		"apiKey": user.APIKey,
	})
}

// RegenerateAPIKey godoc
// @Summary      Regenerate API key
// @Description  Replaces the user's API key, invalidating any editor plugins still using the old one.
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Router       /api/auth/apikey/regenerate [post]
func (h *AuthHandler) RegenerateAPIKey(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	ctx := c.Context()
	newKey, err := models.RegenerateAPIKey(ctx, h.Pool, userID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to regenerate api key")
	}

	return helpers.Success(c, "API key regenerated", fiber.Map{
		"apiKey": newKey,
	})
}

var usernameRegex = regexp.MustCompile(`^[a-z][a-z0-9_-]{2,19}$`)

// isValidUsername checks the username matches our rules:
// starts with a letter, 3-20 chars, lowercase letters,
// numbers, underscore, and hyphen only.
func isValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

// CheckUsername godoc
// @Summary      Check username availability
// @Description  Checks if a username is valid and not already taken.
// @Tags         auth
// @Produce      json
// @Param        username query string true "Username to check"
// @Success      200 {object} helpers.APIResponse
// @Router       /api/auth/check-username [get]
func (h *AuthHandler) CheckUsername(c *fiber.Ctx) error {
	username := strings.TrimSpace(strings.ToLower(c.Query("username")))

	if !isValidUsername(username) {
		return helpers.Success(c, "Invalid format", fiber.Map{
			"available": false,
			"reason":    "invalid_format",
		})
	}
	if isReservedUsername(username) {
		return helpers.Success(c, "Reserved", fiber.Map{
			"available": false,
			"reason":    "reserved",
		})
	}

	ctx := c.Context()
	existing, err := models.FindUserByUsername(ctx, h.Pool, username)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}

	if existing != nil {
		return helpers.Success(c, "Username taken", fiber.Map{
			"available": false,
			"reason":    "taken",
		})
	}

	return helpers.Success(c, "Username available", fiber.Map{
		"available": true,
	})
}

// GetMe godoc
// @Summary      Get current user
// @Description  Returns the logged-in user's profile information.
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Router       /api/auth/me [get]
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	ctx := c.Context()
	user, err := models.FindUserByID(ctx, h.Pool, userID)
	if err != nil || user == nil {
		return helpers.Error(c, fiber.StatusNotFound, "User not found")
	}

	return helpers.Success(c, "User retrieved", user)
}

// UpdateProfile godoc
// @Summary      Update profile
// @Description  Updates the logged-in user's profile fields.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body updateProfileRequest true "Profile fields"
// @Success      200 {object} helpers.APIResponse
// @Failure      400 {object} helpers.APIResponse
// @Failure      409 {object} helpers.APIResponse
// @Router       /api/auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var body updateProfileRequest
	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	username := strings.TrimSpace(strings.ToLower(body.Username))
	if !isValidUsername(username) {
		return helpers.Error(c, fiber.StatusBadRequest, "Username must be 3-20 characters, start with a letter, and contain only lowercase letters, numbers, underscore, or hyphen")
	}
	if isReservedUsername(username) {
		return helpers.Error(c, fiber.StatusBadRequest, "This username is reserved")
	}

	displayName := strings.TrimSpace(body.DisplayName)
	if len(displayName) > 50 {
		return helpers.Error(c, fiber.StatusBadRequest, "Display name must be 50 characters or less")
	}

	bio := strings.TrimSpace(body.Bio)
	if len(bio) > 280 {
		return helpers.Error(c, fiber.StatusBadRequest, "Bio must be 280 characters or less")
	}

	ctx := c.Context()

	currentUser, err := models.FindUserByID(ctx, h.Pool, userID)
	if err != nil || currentUser == nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}

	usernameOwner, err := models.FindUserByUsername(ctx, h.Pool, username)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
	if usernameOwner != nil && usernameOwner.ID != userID {
		return helpers.Error(c, fiber.StatusConflict, "Username is already taken")
	}

	var avatarURL, avatarPublicID *string

	switch {
	case body.ProfileImage == nil:
		// Explicit removal — clean up the old Cloudinary asset if one exists.
		if currentUser.AvatarPublicID != nil {
			if err := services.DeleteAvatar(h.CloudinaryCfg, *currentUser.AvatarPublicID); err != nil {
				return helpers.Error(c, fiber.StatusInternalServerError, "Failed to remove image")
			}
		}
		avatarURL = nil
		avatarPublicID = nil

	case strings.HasPrefix(*body.ProfileImage, "data:image"):
		uploaded, err := services.UploadAvatar(h.CloudinaryCfg, userID, *body.ProfileImage)
		if err != nil {
			log.Printf("UploadAvatar error for user %s: %v", userID, err) // TEMPORARY
			return helpers.Error(c, fiber.StatusInternalServerError, "Failed to upload image")
		}
		avatarURL = &uploaded.URL
		avatarPublicID = &uploaded.PublicID

	default:
		// Unchanged — the frontend sent back the existing hosted URL as-is.
		avatarURL = currentUser.AvatarURL
		avatarPublicID = currentUser.AvatarPublicID
	}

	err = models.UpdateProfile(ctx, h.Pool, userID, models.UpdateProfileInput{
		Username:       username,
		DisplayName:    displayName,
		Bio:            bio,
		Website:        strings.TrimSpace(body.Website),
		ContactEmail:   strings.TrimSpace(body.ContactEmail),
		AvatarURL:      avatarURL,
		AvatarPublicID: avatarPublicID,
	})
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to update profile")
	}

	profile, err := models.GetPublicProfile(ctx, h.Pool, username, userID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}

	return helpers.Success(c, "Profile updated", profile)
}

// RequestEmailChange godoc
// @Summary      Request email change
// @Description  Sends a verification link to the new email address.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/auth/change-email [post]
func (h *AuthHandler) RequestEmailChange(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var body requestEmailChangeInput
	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	newEmail := helpers.NormalizeEmail(body.NewEmail)
	if newEmail == "" || !strings.Contains(newEmail, "@") {
		return helpers.Error(c, fiber.StatusBadRequest, "Please provide a valid email address")
	}

	ctx := c.Context()

	existing, err := models.FindUserByEmail(ctx, h.Pool, newEmail)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
	if existing != nil {
		return helpers.Error(c, fiber.StatusConflict, "That email is already in use")
	}

	token, err := generateEmailChangeToken(userID, newEmail, h.JWTSecret)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}

	err = services.SendEmailChangeConfirmation(h.EmailCfg, newEmail, token)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to send confirmation email")
	}

	return helpers.Success(c, "Check your new email to confirm the change", nil)
}

// ConfirmEmailChange godoc
// @Summary      Confirm email change
// @Description  Validates the token and applies the new email address.
// @Tags         auth
// @Produce      json
// @Param        token query string true "Email change token"
// @Success      200 {object} helpers.APIResponse
// @Router       /api/auth/confirm-email-change [get]
func (h *AuthHandler) ConfirmEmailChange(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return helpers.Error(c, fiber.StatusBadRequest, "Missing token")
	}

	userID, newEmail, err := verifyEmailChangeToken(token, h.JWTSecret)
	if err != nil {
		return helpers.Error(c, fiber.StatusUnauthorized, "Invalid or expired confirmation link")
	}

	ctx := c.Context()
	if err := models.UpdateUserEmail(ctx, h.Pool, userID, newEmail); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to update email")
	}

	return helpers.Success(c, "Email updated successfully", nil)
}
