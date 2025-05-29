package roles

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/internal/schemas"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type RoleManager interface {
	CreateRole(ctx context.Context, name, description string,expTime time.Duration) (*schemas.Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (*schemas.Role, error)
	ListRoles(ctx context.Context) ([]schemas.Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, name, description string ,expTime time.Duration) (*schemas.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	AssignPolicyToRole(ctx context.Context, roleID, policyID uuid.UUID) error
	RemovePolicyFromRole(ctx context.Context, roleID, policyID uuid.UUID) error
	GetExpirationTime(ctx context.Context, id uuid.UUID) (time.Duration, error)
}

type Manager struct {
	DB *gorm.DB
}

func NewManager(db *gorm.DB) RoleManager {
	return &Manager{
		DB: db,
	}
}

func (m *Manager) CreateRole(ctx context.Context, name, description string, expTime time.Duration) (*schemas.Role, error) {
	var existingRole schemas.Role
	if err := m.DB.Where("name = ?", name).First(&existingRole).Error; err == nil {
		return nil, errors.New("role with this name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	role := schemas.Role{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Expiration:  expTime,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := m.DB.Create(&role).Error; err != nil {
		klog.Errorf("Failed to create role: %v", err)
		return nil, errors.New("failed to create role")
	}

	return &role, nil
}

func (m *Manager) GetRole(ctx context.Context, id uuid.UUID) (*schemas.Role, error) {
	var role schemas.Role
	if err := m.DB.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}
	return &role, nil
}

func (m *Manager) ListRoles(ctx context.Context) ([]schemas.Role, error) {
	var roles []schemas.Role
	if err := m.DB.Find(&roles).Error; err != nil {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}
	return roles, nil
}

func (m *Manager) UpdateRole(ctx context.Context, id uuid.UUID, name, description string,expirationTime time.Duration) (*schemas.Role, error) {
	var existingRole schemas.Role
	if err := m.DB.Where("name = ? AND id != ?", name, id).First(&existingRole).Error; err == nil {
		return nil, errors.New("another role with this name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	var role schemas.Role
	if err := m.DB.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	role.Name = name
	role.Description = description
	role.UpdatedAt = time.Now()
	role.Expiration= expirationTime

	if err := m.DB.Save(&role).Error; err != nil {
		klog.Errorf("Failed to update role: %v", err)
		return nil, errors.New("failed to update role")
	}

	return &role, nil
}

func (m *Manager) DeleteRole(ctx context.Context, id uuid.UUID) error {
	var role schemas.Role
	if err := m.DB.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	var count int64
	if err := m.DB.Model(&schemas.User{}).Where("role_id = ?", id).Count(&count).Error; err != nil {
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	if count > 0 {
		return errors.New("cannot delete role that is assigned to users")
	}

	if err := m.DB.Delete(&role).Error; err != nil {
		klog.Errorf("Failed to delete role: %v", err)
		return errors.New("failed to delete role")
	}

	return nil
}

func (m *Manager) AssignPolicyToRole(ctx context.Context, roleID, policyID uuid.UUID) error {
	var role schemas.Role
	if err := m.DB.First(&role, "id = ?", roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	var policy schemas.Policy
	if err := m.DB.First(&policy, "id = ?", policyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("policy not found")
		}
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	policy.RolesId = roleID
	if err := m.DB.Save(&policy).Error; err != nil {
		klog.Errorf("Failed to assign policy to role: %v", err)
		return errors.New("failed to assign policy to role")
	}

	return nil
}

func (m *Manager) RemovePolicyFromRole(ctx context.Context, roleID, policyID uuid.UUID) error {
	var policy schemas.Policy
	if err := m.DB.First(&policy, "id = ? AND roles_id = ?", policyID, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("policy not found or not assigned to this role")
		}
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	if err := m.DB.Model(&policy).Update("roles_id", nil).Error; err != nil {
		klog.Errorf("Failed to remove policy from role: %v", err)
		return errors.New("failed to remove policy from role")
	}

	return nil
}

func (m *Manager) GetExpirationTime(ctx context.Context, id uuid.UUID) (time.Duration, error) {
	var role schemas.Role
	if err := m.DB.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("role not found")
		}
		klog.Errorf("Database error: %v", err)
		return 0, errors.New("internal server error")
	}
	return role.Expiration, nil
}