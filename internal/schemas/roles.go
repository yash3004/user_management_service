package schemas

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Role struct { // Changed from Roles to Role for consistency
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	Name        string    `gorm:"size:100;uniqueIndex"`
	Description string    `gorm:"size:255"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relationships
	Users    uuid.UUID `gorm:"type:char(36);not null"`
	Policies uuid.UUID `gorm:"type:char(36);not null"`
}
