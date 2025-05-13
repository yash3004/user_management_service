package endpoints

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/roles"
)

// Role represents a role in the response
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateRoleRequest represents the create role request
type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateRoleResponse represents the create role response
type CreateRoleResponse struct {
	Role Role `json:"role"`
}

// GetRoleRequest represents the get role request
type GetRoleRequest struct {
	ID string `json:"id"`
}

// GetRoleResponse represents the get role response
type GetRoleResponse struct {
	Role Role `json:"role"`
}

// ListRolesRequest represents the list roles request
type ListRolesRequest struct {
	// Add pagination parameters if needed
}

// ListRolesResponse represents the list roles response
type ListRolesResponse struct {
	Roles []Role `json:"roles"`
}

// UpdateRoleRequest represents the update role request
type UpdateRoleRequest struct {
	ID          string `json:"-"` // From URL path
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateRoleResponse represents the update role response
type UpdateRoleResponse struct {
	Role Role `json:"role"`
}

// DeleteRoleRequest represents the delete role request
type DeleteRoleRequest struct {
	ID string `json:"id"`
}

// DeleteRoleResponse represents the delete role response
type DeleteRoleResponse struct {
	Success bool `json:"success"`
}

// RolesEndpoint handles role-related endpoints
type RolesEndpoint struct {
	RoleManager roles.RoleManager
}

// NewRolesEndpoint creates a new roles endpoint
func NewRolesEndpoint(manager roles.RoleManager) *RolesEndpoint {
	return &RolesEndpoint{
		RoleManager: manager,
	}
}

// CreateRole creates a new role
func (e *RolesEndpoint) CreateRole(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(CreateRoleRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Delegate to the role manager
	role, err := e.RoleManager.CreateRole(ctx, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	return CreateRoleResponse{
		Role: Role{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		},
	}, nil
}

// GetRole gets a role by ID
func (e *RolesEndpoint) GetRole(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(GetRoleRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	roleID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid role ID format")
	}

	// Delegate to the role manager
	role, err := e.RoleManager.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return GetRoleResponse{
		Role: Role{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		},
	}, nil
}

// ListRoles lists all roles
func (e *RolesEndpoint) ListRoles(ctx context.Context, request interface{}) (interface{}, error) {
	// Delegate to the role manager
	rolesList, err := e.RoleManager.ListRoles(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	roles := make([]Role, len(rolesList))
	for i, r := range rolesList {
		roles[i] = Role{
			ID:          r.ID.String(),
			Name:        r.Name,
			Description: r.Description,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		}
	}

	return ListRolesResponse{
		Roles: roles,
	}, nil
}

// UpdateRole updates a role
func (e *RolesEndpoint) UpdateRole(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(UpdateRoleRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	roleID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid role ID format")
	}

	// Delegate to the role manager
	role, err := e.RoleManager.UpdateRole(ctx, roleID, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	return UpdateRoleResponse{
		Role: Role{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		},
	}, nil
}

// DeleteRole deletes a role
func (e *RolesEndpoint) DeleteRole(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(DeleteRoleRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	roleID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid role ID format")
	}

	// Delegate to the role manager
	err = e.RoleManager.DeleteRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return DeleteRoleResponse{
		Success: true,
	}, nil
}