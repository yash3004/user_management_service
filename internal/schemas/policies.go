package schemas

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Policy struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	Name        string    `gorm:"size:100;uniqueIndex"`
	Description string    `gorm:"size:255"`
	Resource    string    `gorm:"size:100;not null"` // The resource this policy applies to
	Action      string    `gorm:"size:100;not null"` // The action allowed (e.g., "read", "write")
	Effect      string    `gorm:"size:20;not null"`  // "allow" or "deny"
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relationships
	RolesId uuid.UUID `gorm:"type:char(36)not null;"`
}
