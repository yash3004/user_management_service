package projects

import (
	"fmt"
	"github.com/yash3004/user_management_service/internal/schemas"

	"gorm.io/gorm"
)

type ProjectService struct {
	DB *gorm.DB
}

func (ps *ProjectService) CreateProject(project *schemas.Project) error {
	tx := ps.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Create(project).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := ps.CreateUserProjectTable(project.UniqueID, tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (ps *ProjectService) CreateUserProjectTable(projectID string, tx *gorm.DB) error {
	db := ps.DB
	if tx != nil {
		db = tx
	}

	tableName := fmt.Sprintf("project_%s_users", projectID)

	db = db.Table(tableName)

	return db.Migrator().CreateTable(&schemas.User{})
}

func (ps *ProjectService) GetProjectByID(id uint) (*schemas.Project, error) {
	var project schemas.Project
	if err := ps.DB.First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil

}

func (ps *ProjectService) DeleteProjectByID(id uint) (bool, error) {
	var project schemas.Project
	if err := ps.DB.First(&project, id).Error; err != nil {
		return false, err
	}

	if err := ps.DB.Delete(&project).Error; err != nil {
		return false, err
	}
	return true, nil

}