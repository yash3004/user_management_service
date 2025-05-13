package endpoints

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/internal/models"
	"github.com/yash3004/user_management_service/project_users"
)

// CreateProjectUserRequest represents the create project user request
type CreateProjectUserRequest struct {
	ProjectID string `json:"project_id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RoleID    string `json:"role_id"`
}

// CreateProjectUserResponse represents the create project user response
type CreateProjectUserResponse struct {
	User models.DisplayUser `json:"user"`
}

// GetProjectUserRequest represents the get project user request
type GetProjectUserRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}

// GetProjectUserResponse represents the get project user response
type GetProjectUserResponse struct {
	User models.DisplayUser `json:"user"`
}

// ListProjectUsersRequest represents the list project users request
type ListProjectUsersRequest struct {
	ProjectID string `json:"project_id"`
}

// ListProjectUsersResponse represents the list project users response
type ListProjectUsersResponse struct {
	Users []models.DisplayUser `json:"users"`
}

// UpdateProjectUserRequest represents the update project user request
type UpdateProjectUserRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Active    bool   `json:"active"`
}

// UpdateProjectUserResponse represents the update project user response
type UpdateProjectUserResponse struct {
	User models.DisplayUser `json:"user"`
}

// DeleteProjectUserRequest represents the delete project user request
type DeleteProjectUserRequest struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}

// DeleteProjectUserResponse represents the delete project user response
type DeleteProjectUserResponse struct {
	Success bool `json:"success"`
}

// ProjectUsersEndpoint handles project-specific user-related endpoints
type ProjectUsersEndpoint struct {
	ProjectUserManager projectusers.ProjectUserManager
}

// NewProjectUsersEndpoint creates a new project users endpoint
func NewProjectUsersEndpoint(manager projectusers.ProjectUserManager) *ProjectUsersEndpoint {
	return &ProjectUsersEndpoint{
		ProjectUserManager: manager,
	}
}

// CreateProjectUser creates a new user in a project-specific user table
func (e *ProjectUsersEndpoint) CreateProjectUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(CreateProjectUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse role ID
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role ID format")
	}

	// Delegate to the project user manager
	user, err := e.ProjectUserManager.CreateProjectUser(ctx, req.ProjectID, req.Email, req.Password, req.FirstName, req.LastName, roleID)
	if err != nil {
		return nil, err
	}

	return CreateProjectUserResponse{
		User: *user,
	}, nil
}

// GetProjectUser gets a user from a project-specific user table by ID
func (e *ProjectUsersEndpoint) GetProjectUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(GetProjectUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Delegate to the project user manager
	user, err := e.ProjectUserManager.GetProjectUser(ctx, req.ProjectID, userID)
	if err != nil {
		return nil, err
	}

	return GetProjectUserResponse{
		User: *user,
	}, nil
}

// ListProjectUsers lists all users in a project-specific user table
func (e *ProjectUsersEndpoint) ListProjectUsers(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(ListProjectUsersRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Delegate to the project user manager
	users, err := e.ProjectUserManager.ListProjectUsers(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}

	return ListProjectUsersResponse{
		Users: users,
	}, nil
}

// UpdateProjectUser updates a user in a project-specific user table
func (e *ProjectUsersEndpoint) UpdateProjectUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(UpdateProjectUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Delegate to the project user manager
	user, err := e.ProjectUserManager.UpdateProjectUser(ctx, req.ProjectID, userID, req.FirstName, req.LastName, req.Active)
	if err != nil {
		return nil, err
	}

	return UpdateProjectUserResponse{
		User: *user,
	}, nil
}

// DeleteProjectUser deletes a user from a project-specific user table
func (e *ProjectUsersEndpoint) DeleteProjectUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(DeleteProjectUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Delegate to the project user manager
	err = e.ProjectUserManager.DeleteProjectUser(ctx, req.ProjectID, userID)
	if err != nil {
		return nil, err
	}

	return DeleteProjectUserResponse{
		Success: true,
	}, nil
}
