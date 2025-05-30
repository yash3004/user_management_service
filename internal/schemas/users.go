package schemas

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	Email     string    `gorm:"uniqueIndex"`
	Password  string    `gorm:"size:255"` // Hashed password for local auth
	FirstName string    `gorm:"size:100"`
	LastName  string    `gorm:"size:100"`
	Active    bool      `gorm:"default:true"`

	// OAuth related fields
	OAuthID      string `gorm:"size:100;index"` // ID from OAuth provider
	OAuthType    string `gorm:"size:50"`        // "google", "github", etc.
	AccessToken  string `gorm:"size:4000"`      // OAuth access token
	RefreshToken string `gorm:"size:4000"`      // OAuth refresh token
	TokenExpiry  time.Time
	ExpirationTime time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relationships
	RoleId    uuid.UUID `gorm:"type:char(36)not null;"` // Changed from Roles to Role
	ProjectId uuid.UUID `gorm:"type:char(36);not null"` // Corrected relationship table name
}
