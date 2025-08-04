package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Email         string `gorm:"unique;not null"`
	Username      string `gorm:"unique;not null"`
	PasswordHash  string `gorm:"not null"`
	Role          string `gorm:"type:varchar(20);default:'user'"`
	EmailVerified bool   `gorm:"default:false"`
}

type RefreshToken struct {
	gorm.Model
	UserID    uint      `gorm:"not null"`
	TokenHash string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
}

type UserProfile struct {
	gorm.Model
	UserID    uint `gorm:"unique;not null"`
	FirstName string
	LastName  string
	Bio       string `gorm:"type:text"`
	AvatarURL string
}
