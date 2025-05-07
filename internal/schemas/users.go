package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// internal/schemas/users.go
type User struct {
	gorm.Model
    ID           uuid.UUID `gorm:"type:uuid;primary_key"`
    Email        string    `gorm:"uniqueIndex"`
    Password     string    // Hashed password for local auth
    OAuthID      string    // ID from OAuth provider
    OAuthType    string    // "google", "github", etc.
    AccessToken  string    // OAuth access token
    RefreshToken string    // OAuth refresh token
    TokenExpiry  time.Time // When the token expires
    // ... other fields
}
