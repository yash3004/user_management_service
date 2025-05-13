package policies

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/internal/schemas"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// PolicyManager defines the interface for policy management operations
type PolicyManager interface {
	CreatePolicy(ctx context.Context, name, description, resource, action, effect string) (*schemas.Policy, error)
	GetPolicy(ctx context.Context, id uuid.UUID) (*schemas.Policy, error)
	ListPolicies(ctx context.Context) ([]schemas.Policy, error)
	UpdatePolicy(ctx context.Context, id uuid.UUID, name, description, resource, action, effect string) (*schemas.Policy, error)
	DeletePolicy(ctx context.Context, id uuid.UUID) error
}

// Manager implements the PolicyManager interface
type Manager struct {
	DB *gorm.DB
}

// NewManager creates a new policy manager
func NewManager(db *gorm.DB) PolicyManager {
	return &Manager{
		DB: db,
	}
}

// CreatePolicy creates a new policy
func (m *Manager) CreatePolicy(ctx context.Context, name, description, resource, action, effect string) (*schemas.Policy, error) {
	// Check if policy with the same name already exists
	var existingPolicy schemas.Policy
	if err := m.DB.Where("name = ?", name).First(&existingPolicy).Error; err == nil {
		return nil, errors.New("policy with this name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	// Validate effect
	if effect != "allow" && effect != "deny" {
		return nil, errors.New("effect must be either 'allow' or 'deny'")
	}

	// Create new policy
	policy := schemas.Policy{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Resource:    resource,
		Action:      action,
		Effect:      effect,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := m.DB.Create(&policy).Error; err != nil {
		klog.Errorf("Failed to create policy: %v", err)
		return nil, errors.New("failed to create policy")
	}

	return &policy, nil
}

// GetPolicy gets a policy by ID
func (m *Manager) GetPolicy(ctx context.Context, id uuid.UUID) (*schemas.Policy, error) {
	var policy schemas.Policy
	if err := m.DB.First(&policy, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("policy not found")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}
	return &policy, nil
}

// ListPolicies lists all policies
func (m *Manager) ListPolicies(ctx context.Context) ([]schemas.Policy, error) {
	var policies []schemas.Policy
	if err := m.DB.Find(&policies).Error; err != nil {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}
	return policies, nil
}

// UpdatePolicy updates a policy
func (m *Manager) UpdatePolicy(ctx context.Context, id uuid.UUID, name, description, resource, action, effect string) (*schemas.Policy, error) {
	// Check if another policy with the same name already exists
	var existingPolicy schemas.Policy
	if err := m.DB.Where("name = ? AND id != ?", name, id).First(&existingPolicy).Error; err == nil {
		return nil, errors.New("another policy with this name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	// Validate effect
	if effect != "allow" && effect != "deny" {
		return nil, errors.New("effect must be either 'allow' or 'deny'")
	}

	var policy schemas.Policy
	if err := m.DB.First(&policy, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("policy not found")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	// Update policy fields
	policy.Name = name
	policy.Description = description
	policy.Resource = resource
	policy.Action = action
	policy.Effect = effect
	policy.UpdatedAt = time.Now()

	if err := m.DB.Save(&policy).Error; err != nil {
		klog.Errorf("Failed to update policy: %v", err)
		return nil, errors.New("failed to update policy")
	}

	return &policy, nil
}

// DeletePolicy deletes a policy
func (m *Manager) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	// Check if policy exists
	var policy schemas.Policy
	if err := m.DB.First(&policy, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("policy not found")
		}
		klog.Errorf("Database error: %v", err)
		return errors.New("internal server error")
	}

	// Delete policy
	if err := m.DB.Delete(&policy).Error; err != nil {
		klog.Errorf("Failed to delete policy: %v", err)
		return errors.New("failed to delete policy")
	}

	return nil
}