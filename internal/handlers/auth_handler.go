package handlers

import (
	"api/internal/auth"
	"api/internal/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	db     *gorm.DB
	logger *logrus.Logger
	config *struct {
		AccessSecret  string
		RefreshSecret string
		AccessExpiry  int
		RefreshExpiry int
	}
}

func NewAuthHandler(db *gorm.DB, logger *logrus.Logger, config *struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  int
	RefreshExpiry int
}) *AuthHandler {
	return &AuthHandler{
		db:     db,
		logger: logger,
		config: config,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email, username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param registration body RegisterRequest true "Registration Details"
// @Success 201 {object} map[string]string "message: Registration successful"
// @Failure 400 {object} map[string]string "error: Validation error message"
// @Failure 409 {object} map[string]string "error: Email or username already exists"
// @Failure 500 {object} map[string]string "error: Internal server error message"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Username string `json:"username" binding:"required,min=3"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email or username already exists
	var existingUser models.User
	if err := h.db.Where("email = ? OR username = ?", input.Email, input.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email or username already exists"})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		h.logger.WithError(err).Error("Failed to hash password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process registration"})
		return
	}

	user := models.User{
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: hashedPassword,
		Role:         "user",
	}

	if err := h.db.Create(&user).Error; err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Simulate email verification
	h.logger.WithFields(logrus.Fields{
		"email": user.Email,
		"id":    user.ID,
	}).Info("Verification email would be sent here")

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. Please check your email for verification.",
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email/username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login Credentials"
// @Success 200 {object} TokenResponse "Returns access_token, refresh_token and user details"
// @Failure 400 {object} map[string]string "error: Validation error message"
// @Failure 401 {object} map[string]string "error: Invalid credentials"
// @Failure 500 {object} map[string]string "error: Internal server error message"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Login    string `json:"login" binding:"required"` // Can be email or username
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if strings.Contains(input.Login, "@") {
		if err := h.db.Where("email = ?", input.Login).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
	} else {
		if err := h.db.Where("username = ?", input.Login).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
	}

	if err := auth.ComparePasswords(user.PasswordHash, input.Password); err != nil {
		h.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err,
		}).Warn("Failed login attempt")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	tokens, err := auth.GenerateTokenPair(
		user.ID,
		user.Role,
		h.config.AccessSecret,
		h.config.RefreshSecret,
		h.config.AccessExpiry,
		h.config.RefreshExpiry,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate tokens")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Store refresh token
	refreshToken := models.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokens.RefreshToken, // In production, hash this before storing
		ExpiresAt: time.Now().Add(time.Hour * 24 * time.Duration(h.config.RefreshExpiry)),
	}

	if err := h.db.Create(&refreshToken).Error; err != nil {
		h.logger.WithError(err).Error("Failed to store refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete login"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
	}).Info("Successful login")

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} TokenPairResponse
// @Failure 400 {object} map[string]string "error: Validation error message"
// @Failure 401 {object} map[string]string "error: Invalid refresh token"
// @Failure 500 {object} map[string]string "error: Internal server error message"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate refresh token
	userID, err := auth.ValidateRefreshToken(input.RefreshToken, h.config.RefreshSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Check if token exists in database
	var storedToken models.RefreshToken
	if err := h.db.Where("token_hash = ? AND user_id = ?", input.RefreshToken, userID).First(&storedToken).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Get user details
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Generate new token pair
	tokens, err := auth.GenerateTokenPair(
		user.ID,
		user.Role,
		h.config.AccessSecret,
		h.config.RefreshSecret,
		h.config.AccessExpiry,
		h.config.RefreshExpiry,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate new tokens")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh tokens"})
		return
	}

	// Update refresh token in database
	if err := h.db.Delete(&storedToken).Error; err != nil {
		h.logger.WithError(err).Error("Failed to delete old refresh token")
	}

	newRefreshToken := models.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokens.RefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * time.Duration(h.config.RefreshExpiry)),
	}

	if err := h.db.Create(&newRefreshToken).Error; err != nil {
		h.logger.WithError(err).Error("Failed to store new refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate the refresh token and logout the user
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Param logout body LogoutRequest true "Refresh Token"
// @Success 200 {object} map[string]string "message: Successfully logged out"
// @Failure 400 {object} map[string]string "error: Validation error message"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Internal server error message"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delete refresh token from database
	if err := h.db.Where("token_hash = ?", input.RefreshToken).Delete(&models.RefreshToken{}).Error; err != nil {
		h.logger.WithError(err).Error("Failed to delete refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
