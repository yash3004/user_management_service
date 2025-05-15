package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/auth/oauth"
	"github.com/yash3004/user_management_service/internal/models"
	"github.com/yash3004/user_management_service/internal/schemas"
	"k8s.io/klog/v2"
)

// CreateOrUpdateOAuthUser creates or updates a user from OAuth provider information
func (m *Manager) CreateOrUpdateOAuthUser(ctx context.Context, userInfo *oauth.UserInfo, projectID uuid.UUID, roleID uuid.UUID) (*models.DisplayUser, error) {
	// Check if user with the same email already exists
	var existingUser schemas.User
	if err := m.DB.Where("email = ?", userInfo.Email).First(&existingUser).Error; err == nil {
		// User exists, update OAuth information
		existingUser.FirstName = userInfo.FirstName
		existingUser.LastName = userInfo.LastName
		existingUser.UpdatedAt = time.Now()

		if err := m.DB.Save(&existingUser).Error; err != nil {
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

	// Check if project exists
	var project schemas.Project
	if err := m.DB.First(&project, "id = ?", projectID).Error; err != nil {
		klog.Errorf("Project not found: %v", err)
		return nil, errors.New("project not found")
	}

	// Check if role exists
	var role schemas.Role
	if err := m.DB.First(&role, "id = ?", roleID).Error; err != nil {
		klog.Errorf("Role not found: %v", err)
		return nil, errors.New("role not found")
	}

	// Create new user
	newUser := schemas.User{
		ID:        uuid.New(),
		Email:     userInfo.Email,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
		Active:    true,
		RoleId:    roleID,
		ProjectId: projectID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := m.DB.Create(&newUser).Error; err != nil {
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

