package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	// Remove gorm.Model and use explicit fields to avoid duplication
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	Name        string    `gorm:"size:255;not null"`
	Description string    `gorm:"size:1000"`
	UniqueID    string    `gorm:"size:50;uniqueIndex;not null"` // This will be used for table naming
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relationships
}
