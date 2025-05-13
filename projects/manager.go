package projects

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/internal/schemas"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// ProjectManager defines the interface for project management operations
type ProjectManager interface {
	CreateProject(ctx context.Context, name, description, uniqueID string) (*schemas.Project, error)
	GetProject(ctx context.Context, id uuid.UUID) (*schemas.Project, error)
	ListProjects(ctx context.Context) ([]schemas.Project, error)
	UpdateProject(ctx context.Context, id uuid.UUID, name, description string) (*schemas.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
}

// Manager implements the ProjectManager interface
type Manager struct {
	DB *gorm.DB
}

// NewManager creates a new project manager
func NewManager(db *gorm.DB) ProjectManager {
	return &Manager{
		DB: db,
	}
}

// CreateProject creates a new project
func (m *Manager) CreateProject(ctx context.Context, name, description, uniqueID string) (*schemas.Project, error) {
	// Check if project with the same unique ID already exists
	var existingProject schemas.Project
	if err := m.DB.Where("unique_id = ?", uniqueID).First(&existingProject).Error; err == nil {
		return nil, errors.New("project with this unique ID already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	// Create new project
	project := schemas.Project{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		UniqueID:    uniqueID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Start a transaction
	tx := m.DB.Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}

	// Create the project
	if err := tx.Create(&project).Error; err != nil {
		tx.Rollback()
		klog.Errorf("Failed to create project: %v", err)
		return nil, errors.New("failed to create project")
	}

	// Create project-specific user table
	tableName := "project_" + uniqueID + "_users"
	if err := tx.Table(tableName).Migrator().CreateTable(&schemas.ProjectUser{}); err != nil {
		tx.Rollback()
		klog.Errorf("Failed to create project user table: %v", err)
		return nil, errors.New("failed to create project resources")
	}
	
	// Add indexes to the project-specific user table
	if err := tx.Table(tableName).Migrator().CreateIndex(&schemas.ProjectUser{}, "Email"); err != nil {
		tx.Rollback()
		klog.Errorf("Failed to create email index on project user table: %v", err)
		return nil, errors.New("failed to create project resources")
	}
	
	if err := tx.Table(tableName).Migrator().CreateIndex(&schemas.ProjectUser{}, "OAuthID"); err != nil {
		tx.Rollback()
		klog.Errorf("Failed to create oauth_id index on project user table: %v", err)
		return nil, errors.New("failed to create project resources")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		klog.Errorf("Failed to commit transaction: %v", err)
		return nil, errors.New("failed to create project")
	}

	return &project, nil
}

// GetProject gets a project by ID
func (m *Manager) GetProject(ctx context.Context, id uuid.UUID) (*schemas.Project, error) {
	var project schemas.Project
	if err := m.DB.First(&project, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}
	return &project, nil
}

// ListProjects lists all projects
func (m *Manager) ListProjects(ctx context.Context) ([]schemas.Project, error) {
	var projects []schemas.Project
	if err := m.DB.Find(&projects).Error; err != nil {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}
	return projects, nil
}

// UpdateProject updates a project
func (m *Manager) UpdateProject(ctx context.Context, id uuid.UUID, name, description string) (*schemas.Project, error) {
	var project schemas.Project
	if err := m.DB.First(&project, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	// Update project fields
	project.Name = name
	project.Description = description
	project.UpdatedAt = time.Now()

	if err := m.DB.Save(&project).Error; err != nil {
		klog.Errorf("Failed to update project: %v", err)
		return nil, errors.New("failed to update project")
	}

	return &project, nil
}

// DeleteProject deletes a project
func (m *Manager) DeleteProject(ctx context.Context, id uuid.UUID) error {
	// Start a transaction
	tx := m.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	// Get the project to get the uniqueID
	var project schemas.Project
	if err := tx.First(&project, "id = ?", id).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("project not found")
		}
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	// Delete the project
	if err := tx.Delete(&project).Error; err != nil {
		tx.Rollback()
		klog.Errorf("Failed to delete project: %v", err)
		return errors.New("failed to delete project")
	}

	// Drop the project-specific user table
	tableName := "project_" + project.UniqueID + "_users"
	if err := tx.Table(tableName).Migrator().DropTable(&schemas.ProjectUser{}); err != nil {
		tx.Rollback()
		klog.Errorf("Failed to drop project user table: %v", err)
		return errors.New("failed to delete project resources")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		klog.Errorf("Failed to commit transaction: %v", err)
		return errors.New("failed to delete project")
	}

	return nil
}