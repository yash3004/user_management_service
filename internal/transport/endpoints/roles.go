package endpoints

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/roles"
)

type Role struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Expiration  time.Duration `json:"expiration"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Expiration  int    `json:"expiration"`
}

type CreateRoleResponse struct {
	Role Role `json:"role"`
}

type GetRoleRequest struct {
	ID string `json:"id"`
}

type GetRoleResponse struct {
	Role Role `json:"role"`
}

type ListRolesResponse struct {
	Roles []Role `json:"roles"`
}

type UpdateRoleRequest struct {
	ID          string `json:"-"` // From URL path
	Name        string `json:"name"`
	Description string `json:"description"`
	Expiration  int    `json:"expiration"`
}

type UpdateRoleResponse struct {
	Role Role `json:"role"`
}

type DeleteRoleRequest struct {
	ID string `json:"id"`
}

type DeleteRoleResponse struct {
	Success bool `json:"success"`
}

type RolesEndpoint struct {
	RoleManager roles.RoleManager
}

func NewRolesEndpoint(manager roles.RoleManager) *RolesEndpoint {
	return &RolesEndpoint{
		RoleManager: manager,
	}
}

func (e *RolesEndpoint) CreateRole(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(CreateRoleRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	role, err := e.RoleManager.CreateRole(ctx, req.Name, req.Description, addHours(req.Expiration))
	if err != nil {
		return nil, err
	}

	return CreateRoleResponse{
		Role: Role{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			Expiration:  role.Expiration,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		},
	}, nil
}

func (e *RolesEndpoint) GetRole(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(GetRoleRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	roleID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid role ID format")
	}

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

func (e *RolesEndpoint) ListRoles(ctx context.Context, request interface{}) (interface{}, error) {
	rolesList, err := e.RoleManager.ListRoles(ctx)
	if err != nil {
		return nil, err
	}

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

func (e *RolesEndpoint) UpdateRole(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(UpdateRoleRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	roleID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid role ID format")
	}

	role, err := e.RoleManager.UpdateRole(ctx, roleID, req.Name, req.Description, addHours(req.Expiration))
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

func (e *RolesEndpoint) DeleteRole(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(DeleteRoleRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	roleID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid role ID format")
	}

	err = e.RoleManager.DeleteRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return DeleteRoleResponse{
		Success: true,
	}, nil
}

func addHours(hours int) time.Duration {
	return time.Duration(hours) * time.Hour
}
