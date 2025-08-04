package handlers

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Username string `json:"username" binding:"required,min=3" example:"johndoe"`
	Password string `json:"password" binding:"required,min=8" example:"strongpassword123"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Login    string `json:"login" binding:"required" example:"user@example.com"` // email or username
	Password string `json:"password" binding:"required" example:"strongpassword123"`
}

// TokenResponse represents the response containing tokens
type TokenResponse struct {
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User         UserResponse `json:"user"`
}

// UserResponse represents the user information in responses
type UserResponse struct {
	ID       uint   `json:"id" example:"1"`
	Email    string `json:"email" example:"user@example.com"`
	Username string `json:"username" example:"johndoe"`
	Role     string `json:"role" example:"user"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// LogoutRequest represents the logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// TokenPairResponse represents the response containing a new token pair
type TokenPairResponse struct {
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User         UserResponse `json:"user,omitempty"`
}

// UpdateProfileRequest represents the profile update request
type UpdateProfileRequest struct {
	FirstName string `json:"firstName" example:"John"`
	LastName  string `json:"lastName" example:"Doe"`
	Bio       string `json:"bio" example:"Software Developer"`
	AvatarURL string `json:"avatarURL" example:"https://example.com/avatar.jpg"`
}

// ProfileResponse represents the profile information in responses
type ProfileResponse struct {
	FirstName string `json:"firstName" example:"John"`
	LastName  string `json:"lastName" example:"Doe"`
	Bio       string `json:"bio" example:"Software Developer"`
	AvatarURL string `json:"avatarURL" example:"https://example.com/avatar.jpg"`
}

// UserProfileResponse represents the complete user profile response
type UserProfileResponse struct {
	User    UserResponse    `json:"user"`
	Profile ProfileResponse `json:"profile"`
}

// ChangePasswordRequest represents the password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required" example:"oldpassword123"`
	NewPassword     string `json:"newPassword" binding:"required,min=8" example:"newpassword123"`
}

// ChangeRoleRequest represents the role change request
type ChangeRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user admin" example:"admin"`
}

// UserRoleResponse represents the response after changing a user's role
type UserRoleResponse struct {
	Message string       `json:"message" example:"User role updated successfully"`
	User    UserResponse `json:"user"`
}

// UsersListResponse represents the response for listing all users
type UsersListResponse struct {
	Users []struct {
		ID        uint   `json:"id" example:"1"`
		Email     string `json:"email" example:"user@example.com"`
		Username  string `json:"username" example:"johndoe"`
		Role      string `json:"role" example:"user"`
		Verified  bool   `json:"verified" example:"true"`
		CreatedAt string `json:"createdAt" example:"2025-08-04T12:00:00Z"`
		Profile   struct {
			FirstName string `json:"firstName" example:"John"`
			LastName  string `json:"lastName" example:"Doe"`
		} `json:"profile"`
	} `json:"users"`
}
