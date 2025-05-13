package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/auth/oauth"
	"github.com/yash3004/user_management_service/internal/models"
	"github.com/yash3004/user_management_service/internal/schemas"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// UserManager defines the interface for user management operations
type UserManager interface {
	CreateUser(ctx context.Context, email, password, firstName, lastName string, roleID, projectID uuid.UUID) (*schemas.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*schemas.User, error)
	GetUserByEmail(ctx context.Context, email string) (*schemas.User, error)
	ListUsers(ctx context.Context) ([]schemas.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, firstName, lastName string, active bool) (*schemas.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ChangePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error
	AssignRole(ctx context.Context, userID, roleID uuid.UUID) error
	CreateOrUpdateOAuthUser(ctx context.Context, userInfo *oauth.UserInfo, projectID uuid.UUID, roleID uuid.UUID) (*models.DisplayUser, error)
	GenerateToken(ctx context.Context, userID uuid.UUID) (string, time.Time, error)
}


// Manager implements the UserManager interface
type Manager struct {
	DB *gorm.DB
}

// NewManager creates a new user manager
func NewManager(db *gorm.DB) UserManager {
	return &Manager{
		DB: db,
	}
}


// CreateUser creates a new user
func (m *Manager) CreateUser(ctx context.Context, email, password, firstName, lastName string, roleID, projectID uuid.UUID) (*schemas.User, error) {
	// Check if user with the same email already exists
	var existingUser schemas.User
	if err := m.DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	// Check if role exists
	var role schemas.Role
	if err := m.DB.First(&role, "id = ?", roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	// Check if project exists
	var project schemas.Project
	if err := m.DB.First(&project, "id = ?", projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		klog.Errorf("Failed to hash password: %v", err)
		return nil, errors.New("failed to process password")
	}

	// Create new user
	user := schemas.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		Active:    true,
		RoleId:    roleID,
		ProjectId: projectID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := m.DB.Create(&user).Error; err != nil {
		klog.Errorf("Failed to create user: %v", err)
		return nil, errors.New("failed to create user")
	}

	return &user, nil
}

// GetUser gets a user by ID
func (m *Manager) GetUser(ctx context.Context, id uuid.UUID) (*schemas.User, error) {
	var user schemas.User
	if err := m.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}
	return &user, nil
}

// GetUserByEmail gets a user by email
func (m *Manager) GetUserByEmail(ctx context.Context, email string) (*schemas.User, error) {
	var user schemas.User
	if err := m.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}
	return &user, nil
}

// ListUsers lists all users
func (m *Manager) ListUsers(ctx context.Context) ([]schemas.User, error) {
	var users []schemas.User
	if err := m.DB.Find(&users).Error; err != nil {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}
	return users, nil
}

// UpdateUser updates a user
func (m *Manager) UpdateUser(ctx context.Context, id uuid.UUID, firstName, lastName string, active bool) (*schemas.User, error) {
	var user schemas.User
	if err := m.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	// Update user fields
	user.FirstName = firstName
	user.LastName = lastName
	user.Active = active
	user.UpdatedAt = time.Now()

	if err := m.DB.Save(&user).Error; err != nil {
		klog.Errorf("Failed to update user: %v", err)
		return nil, errors.New("failed to update user")
	}

	return &user, nil
}

// DeleteUser deletes a user
func (m *Manager) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Check if user exists
	var user schemas.User
	if err := m.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	// Delete user (soft delete with gorm)
	if err := m.DB.Delete(&user).Error; err != nil {
		klog.Errorf("Failed to delete user: %v", err)
		return errors.New("failed to delete user")
	}

	return nil
}

// ChangePassword changes a user's password
func (m *Manager) ChangePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error {
	var user schemas.User
	if err := m.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		klog.Errorf("Failed to hash password: %v", err)
		return errors.New("failed to process password")
	}

	// Update password
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := m.DB.Save(&user).Error; err != nil {
		klog.Errorf("Failed to update password: %v", err)
		return errors.New("failed to update password")
	}

	return nil
}

// AssignRole assigns a role to a user
func (m *Manager) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	// Check if user exists
	var user schemas.User
	if err := m.DB.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	// Check if role exists
	var role schemas.Role
	if err := m.DB.First(&role, "id = ?", roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	// Update user's role
	user.RoleId = roleID
	user.UpdatedAt = time.Now()

	if err := m.DB.Save(&user).Error; err != nil {
		klog.Errorf("Failed to assign role to user: %v", err)
		return errors.New("failed to assign role to user")
	}

	return nil
}
