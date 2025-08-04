package handlers

import (
	"api/internal/auth"
	"api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewUserHandler(db *gorm.DB, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		db:     db,
		logger: logger,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the profile information of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} UserProfileResponse
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("userID")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var profile models.UserProfile
	if err := h.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			profile = models.UserProfile{
				UserID: userID,
			}
		} else {
			h.logger.WithError(err).Error("Failed to fetch user profile")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
			"role":     user.Role,
		},
		"profile": gin.H{
			"firstName": profile.FirstName,
			"lastName":  profile.LastName,
			"bio":       profile.Bio,
			"avatarURL": profile.AvatarURL,
		},
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the profile information of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param profile body UpdateProfileRequest true "Profile Information"
// @Success 200 {object} ProfileResponse
// @Failure 400 {object} map[string]string "error: Validation error"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("userID")

	var input struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Bio       string `json:"bio"`
		AvatarURL string `json:"avatarURL"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var profile models.UserProfile
	result := h.db.Where("user_id = ?", userID).First(&profile)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			profile = models.UserProfile{
				UserID:    userID,
				FirstName: input.FirstName,
				LastName:  input.LastName,
				Bio:       input.Bio,
				AvatarURL: input.AvatarURL,
			}
			if err := h.db.Create(&profile).Error; err != nil {
				h.logger.WithError(err).Error("Failed to create user profile")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
				return
			}
		} else {
			h.logger.WithError(result.Error).Error("Failed to fetch user profile")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}
	} else {
		profile.FirstName = input.FirstName
		profile.LastName = input.LastName
		profile.Bio = input.Bio
		profile.AvatarURL = input.AvatarURL

		if err := h.db.Save(&profile).Error; err != nil {
			h.logger.WithError(err).Error("Failed to update user profile")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": gin.H{
			"firstName": profile.FirstName,
			"lastName":  profile.LastName,
			"bio":       profile.Bio,
			"avatarURL": profile.AvatarURL,
		},
	})
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change the password of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param passwords body ChangePasswordRequest true "Password Information"
// @Success 200 {object} map[string]string "message: Password changed successfully"
// @Failure 400 {object} map[string]string "error: Validation error"
// @Failure 401 {object} map[string]string "error: Current password is incorrect"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /users/change-password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("userID")

	var input struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := auth.ComparePasswords(user.PasswordHash, input.CurrentPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	hashedPassword, err := auth.HashPassword(input.NewPassword)
	if err != nil {
		h.logger.WithError(err).Error("Failed to hash new password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	user.PasswordHash = hashedPassword
	if err := h.db.Save(&user).Error; err != nil {
		h.logger.WithError(err).Error("Failed to update password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	h.logger.WithField("user_id", userID).Info("Password changed successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// DeleteAccount godoc
// @Summary Delete user account
// @Description Soft delete the authenticated user's account
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]string "message: Account deleted successfully"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /users/account [delete]
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID := c.GetUint("userID")

	// Start a transaction
	tx := h.db.Begin()

	// Delete refresh tokens
	if err := tx.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error; err != nil {
		tx.Rollback()
		h.logger.WithError(err).Error("Failed to delete refresh tokens")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	// Delete user profile
	if err := tx.Where("user_id = ?", userID).Delete(&models.UserProfile{}).Error; err != nil {
		tx.Rollback()
		h.logger.WithError(err).Error("Failed to delete user profile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	// Soft delete user
	if err := tx.Delete(&models.User{}, userID).Error; err != nil {
		tx.Rollback()
		h.logger.WithError(err).Error("Failed to delete user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		h.logger.WithError(err).Error("Failed to commit account deletion transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	h.logger.WithField("user_id", userID).Info("Account deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
