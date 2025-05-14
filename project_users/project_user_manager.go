package projectusers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/auth/oauth"
	"github.com/yash3004/user_management_service/internal/models"
	"github.com/yash3004/user_management_service/internal/schemas"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// ProjectUserManager defines the interface for project-specific user management operations
type ProjectUserManager interface {
	CreateProjectUser(ctx context.Context, projectID string, email, password, firstName, lastName string, roleID uuid.UUID) (*models.DisplayUser, error)
	GetProjectUser(ctx context.Context, projectID string, userID uuid.UUID) (*models.DisplayUser, error)
	GetProjectUserByEmail(ctx context.Context, projectID string, email string) (*models.DisplayUser, error)
	ListProjectUsers(ctx context.Context, projectID string) ([]models.DisplayUser, error)
	UpdateProjectUser(ctx context.Context, projectID string, userID uuid.UUID, firstName, lastName string, active bool) (*models.DisplayUser, error)
	DeleteProjectUser(ctx context.Context, projectID string, userID uuid.UUID) error
	CreateOrUpdateOAuthProjectUser(ctx context.Context, projectID string, userInfo *oauth.UserInfo, roleID uuid.UUID) (*models.DisplayUser, error)
}

// ProjectUserManagerImpl implements the ProjectUserManager interface
type ProjectUserManagerImpl struct {
	DB *gorm.DB
}

func NewManager(db *gorm.DB) ProjectUserManager {
	return &ProjectUserManagerImpl{
		DB: db,
	}
}

// getProjectUserTableName returns the table name for a specific project
func getProjectUserTableName(projectID string) string {
	return fmt.Sprintf("project_%s_users", projectID)
}

// CreateProjectUser creates a new user in a project-specific user table
func (m *ProjectUserManagerImpl) CreateProjectUser(ctx context.Context, projectID string, email, password, firstName, lastName string, roleID uuid.UUID) (*models.DisplayUser, error) {
	tableName := getProjectUserTableName(projectID)

	// Check if user with the same email already exists
	var existingUser schemas.ProjectUser
	if err := m.DB.Table(tableName).Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email already exists in this project")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		klog.Errorf("Failed to hash password: %v", err)
		return nil, errors.New("failed to process password")
	}

	// Parse project ID
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project ID format")
	}

	// Create new user
	user := schemas.ProjectUser{
		ID:          uuid.New(),
		Email:       email,
		Password:    string(hashedPassword),
		FirstName:   firstName,
		LastName:    lastName,
		Active:      true,
		RoleId:      roleID,
		ProjectId:   projectUUID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		TokenExpiry: time.Now().Add(24 * time.Hour), // Set token expiry to 24 hours
	}

	if err := m.DB.Table(tableName).Create(&user).Error; err != nil {
		klog.Errorf("Failed to create user: %v", err)
		return nil, errors.New("failed to create user")
	}

	return &models.DisplayUser{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Active:    user.Active,
		RoleID:    user.RoleId.String(),
		ProjectID: user.ProjectId.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// GetProjectUser gets a user from a project-specific user table by ID
func (m *ProjectUserManagerImpl) GetProjectUser(ctx context.Context, projectID string, userID uuid.UUID) (*models.DisplayUser, error) {
	tableName := getProjectUserTableName(projectID)

	var user schemas.ProjectUser
	if err := m.DB.Table(tableName).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found in this project")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	return &models.DisplayUser{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Active:    user.Active,
		RoleID:    user.RoleId.String(),
		ProjectID: user.ProjectId.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// GetProjectUserByEmail gets a user from a project-specific user table by email
func (m *ProjectUserManagerImpl) GetProjectUserByEmail(ctx context.Context, projectID string, email string) (*models.DisplayUser, error) {
	tableName := getProjectUserTableName(projectID)

	var user schemas.ProjectUser
	if err := m.DB.Table(tableName).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found in this project")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	return &models.DisplayUser{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Active:    user.Active,
		RoleID:    user.RoleId.String(),
		ProjectID: user.ProjectId.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// ListProjectUsers lists all users in a project-specific user table
func (m *ProjectUserManagerImpl) ListProjectUsers(ctx context.Context, projectID string) ([]models.DisplayUser, error) {
	tableName := getProjectUserTableName(projectID)

	var projectUsers []schemas.ProjectUser
	if err := m.DB.Table(tableName).Find(&projectUsers).Error; err != nil {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	users := make([]models.DisplayUser, len(projectUsers))
	for i, u := range projectUsers {
		users[i] = models.DisplayUser{
			ID:        u.ID.String(),
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Active:    u.Active,
			RoleID:    u.RoleId.String(),
			ProjectID: u.ProjectId.String(),
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}

	return users, nil
}

// UpdateProjectUser updates a user in a project-specific user table
func (m *ProjectUserManagerImpl) UpdateProjectUser(ctx context.Context, projectID string, userID uuid.UUID, firstName, lastName string, active bool) (*models.DisplayUser, error) {
	tableName := getProjectUserTableName(projectID)

	var user schemas.ProjectUser
	if err := m.DB.Table(tableName).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found in this project")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	// Update user fields
	user.FirstName = firstName
	user.LastName = lastName
	user.Active = active
	user.UpdatedAt = time.Now()

	if err := m.DB.Table(tableName).Save(&user).Error; err != nil {
		klog.Errorf("Failed to update user: %v", err)
		return nil, errors.New("failed to update user")
	}

	return &models.DisplayUser{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Active:    user.Active,
		RoleID:    user.RoleId.String(),
		ProjectID: user.ProjectId.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// DeleteProjectUser deletes a user from a project-specific user table
func (m *ProjectUserManagerImpl) DeleteProjectUser(ctx context.Context, projectID string, userID uuid.UUID) error {
	tableName := getProjectUserTableName(projectID)

	// Check if user exists
	var user schemas.ProjectUser
	if err := m.DB.Table(tableName).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found in this project")
		}
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	// Delete user (soft delete with gorm)
	if err := m.DB.Table(tableName).Delete(&user).Error; err != nil {
		klog.Errorf("Failed to delete user: %v", err)
		return errors.New("failed to delete user")
	}

	return nil
}

// CreateOrUpdateOAuthProjectUser creates or updates a user from OAuth provider information in a project-specific user table
func (m *ProjectUserManagerImpl) CreateOrUpdateOAuthProjectUser(ctx context.Context, projectID string, userInfo *oauth.UserInfo, roleID uuid.UUID) (*models.DisplayUser, error) {
	tableName := getProjectUserTableName(projectID)

	// Check if user with the same email already exists
	var existingUser schemas.ProjectUser
	if err := m.DB.Table(tableName).Where("email = ?", userInfo.Email).First(&existingUser).Error; err == nil {
		// User exists, update OAuth information
		existingUser.FirstName = userInfo.FirstName
		existingUser.LastName = userInfo.LastName
		existingUser.OAuthID = userInfo.ID
		existingUser.OAuthType = userInfo.Provider
		existingUser.UpdatedAt = time.Now()

		if err := m.DB.Table(tableName).Save(&existingUser).Error; err != nil {
			klog.Errorf("Failed to update user: %v", err)
			return nil, errors.New("failed to update user")
		}

		// Return the updated user
		return &models.DisplayUser{
			ID:        existingUser.ID.String(),
			Email:     existingUser.Email,
			FirstName: existingUser.FirstName,
			LastName:  existingUser.LastName,
			Active:    existingUser.Active,
			RoleID:    existingUser.RoleId.String(),
			ProjectID: existingUser.ProjectId.String(),
			CreatedAt: existingUser.CreatedAt,
			UpdatedAt: existingUser.UpdatedAt,
		}, nil
	}

	// Parse project ID
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, errors.New("invalid project ID format")
	}

	// Create new user
	newUser := schemas.ProjectUser{
		ID:        uuid.New(),
		Email:     userInfo.Email,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
		Active:    true,
		OAuthID:   userInfo.ID,
		OAuthType: userInfo.Provider,
		RoleId:    roleID,
		ProjectId: projectUUID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := m.DB.Table(tableName).Create(&newUser).Error; err != nil {
		klog.Errorf("Failed to create user: %v", err)
		return nil, errors.New("failed to create user")
	}

	// Return the created user
	return &models.DisplayUser{
		ID:        newUser.ID.String(),
		Email:     newUser.Email,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Active:    newUser.Active,
		RoleID:    newUser.RoleId.String(),
		ProjectID: newUser.ProjectId.String(),
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}, nil
}
