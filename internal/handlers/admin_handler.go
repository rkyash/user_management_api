package handlers

import (
	"api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

type AdminHandler struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewAdminHandler(db *gorm.DB, logger *logrus.Logger) *AdminHandler {
	return &AdminHandler{
		db:     db,
		logger: logger,
	}
}

// ListUsers godoc
// @Summary List all users
// @Description Get a list of all users (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} UsersListResponse
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Forbidden - Admin access required"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var users []models.User
	if err := h.db.Find(&users).Error; err != nil {
		h.logger.WithError(err).Error("Failed to fetch users list")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	var usersList []gin.H
	for _, user := range users {
		var profile models.UserProfile
		h.db.Where("user_id = ?", user.ID).First(&profile)

		usersList = append(usersList, gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"username":  user.Username,
			"role":      user.Role,
			"verified":  user.EmailVerified,
			"createdAt": user.CreatedAt,
			"profile": gin.H{
				"firstName": profile.FirstName,
				"lastName":  profile.LastName,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": usersList})
}

// ChangeUserRole godoc
// @Summary Change user role
// @Description Change the role of a specific user (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Param role body ChangeRoleRequest true "New Role"
// @Success 200 {object} UserRoleResponse
// @Failure 400 {object} map[string]string "error: Validation error"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Forbidden - Admin access required"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /admin/users/{id}/role [put]
func (h *AdminHandler) ChangeUserRole(c *gin.Context) {
	userID := c.Param("id")

	var input struct {
		Role string `json:"role" binding:"required,oneof=user admin"`
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

	user.Role = input.Role
	if err := h.db.Save(&user).Error; err != nil {
		h.logger.WithError(err).Error("Failed to update user role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"new_role": input.Role,
	}).Info("User role updated")

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}
