package endpoints

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/internal/models"
	"github.com/yash3004/user_management_service/users"
)



type CreateUserRequest struct {
	ProjectID string `json:"project_id"`
	ID        string `json:"-"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RoleID    string `json:"role_id"`
}

type CreateUserResponse struct {
	User models.DisplayUser `json:"user"`
}

type GetUserRequest struct {
	ID string `json:"id"`
}

type GetUserResponse struct {
	User models.DisplayUser `json:"user"`
}

type ListUsersResponse struct {
	Users []models.DisplayUser `json:"users"`
}

// UpdateUserRequest represents the update user request
type UpdateUserRequest struct {
	ProjectId string `json:"project_id"`
	ID        string `json:"-"` // From URL path
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Active    bool   `json:"active"`
	RoleID    string `json:"role_id"`
}

type UpdateUserResponse struct {
	User models.DisplayUser `json:"user"`
}

type DeleteUserRequest struct {
	ProjectId string `json:"project_id"`
	ID        string `json:"id"`
}

type DeleteUserResponse struct {
	Success bool `json:"success"`
}

type ChangePasswordRequest struct {
	ProjectId       string `json:"project_id"`
	ID              string `json:"-"` // From URL path
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type ChangePasswordResponse struct {
	Success bool `json:"success"`
}

type UsersEndpoint struct {
	UserManager users.UserManager
}

func NewUsersEndpoint(manager users.UserManager) *UsersEndpoint {
	return &UsersEndpoint{
		UserManager: manager,
	}
}

func (e *UsersEndpoint) CreateUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(CreateUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role ID format")
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, errors.New("invalid project ID format")
	}

	user, err := e.UserManager.CreateUser(ctx, req.Email, req.Password, req.FirstName, req.LastName, roleID, projectID)
	if err != nil {
		return nil, err
	}

	return CreateUserResponse{
		User: models.DisplayUser{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Active:    user.Active,
			RoleID:    user.RoleId.String(),
			ProjectID: user.ProjectId.String(),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (e *UsersEndpoint) GetUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(GetUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	userID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	user, err := e.UserManager.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return GetUserResponse{
		User: models.DisplayUser{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Active:    user.Active,
			RoleID:    user.RoleId.String(),
			ProjectID: user.ProjectId.String(),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// ListUsers lists all users
func (e *UsersEndpoint) ListUsers(ctx context.Context, request interface{}) (interface{}, error) {
	usersList, err := e.UserManager.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]models.DisplayUser, len(usersList))
	for i, u := range usersList {
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

	return ListUsersResponse{
		Users: users,
	}, nil
}

func (e *UsersEndpoint) UpdateUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(UpdateUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	userID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	user, err := e.UserManager.UpdateUser(ctx, userID, req.FirstName, req.LastName, req.Active)
	if err != nil {
		return nil, err
	}

	return UpdateUserResponse{
		User: models.DisplayUser{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Active:    user.Active,
			RoleID:    user.RoleId.String(),
			ProjectID: user.ProjectId.String(),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (e *UsersEndpoint) DeleteUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(DeleteUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	userID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	err = e.UserManager.DeleteUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return DeleteUserResponse{
		Success: true,
	}, nil
}

func (e *UsersEndpoint) ChangePassword(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(ChangePasswordRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	userID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	err = e.UserManager.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}

	return ChangePasswordResponse{
		Success: true,
	}, nil
}
