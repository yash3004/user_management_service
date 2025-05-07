package schemas

import (
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Name        string `gorm:"size:255;not null"`
	Description string `gorm:"size:1000"`
	UniqueID    string `gorm:"size:50;uniqueIndex;not null"` // This will be used for table naming
}
