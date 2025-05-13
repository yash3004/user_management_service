package endpoints

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/users"
)

// User represents a user in the response
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Active    bool      `json:"active"`
	RoleID    string    `json:"role_id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest represents the create user request
type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RoleID    string `json:"role_id"`
	ProjectID string `json:"project_id"`
}

// CreateUserResponse represents the create user response
type CreateUserResponse struct {
	User User `json:"user"`
}

// GetUserRequest represents the get user request
type GetUserRequest struct {
	ID string `json:"id"`
}

// GetUserResponse represents the get user response
type GetUserResponse struct {
	User User `json:"user"`
}

// ListUsersRequest represents the list users request
type ListUsersRequest struct {
	// Add pagination parameters if needed
}

// ListUsersResponse represents the list users response
type ListUsersResponse struct {
	Users []User `json:"users"`
}

// UpdateUserRequest represents the update user request
type UpdateUserRequest struct {
	ID        string `json:"-"` // From URL path
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Active    bool   `json:"active"`
}

// UpdateUserResponse represents the update user response
type UpdateUserResponse struct {
	User User `json:"user"`
}

// DeleteUserRequest represents the delete user request
type DeleteUserRequest struct {
	ID string `json:"id"`
}

// DeleteUserResponse represents the delete user response
type DeleteUserResponse struct {
	Success bool `json:"success"`
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	ID              string `json:"-"` // From URL path
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePasswordResponse represents the change password response
type ChangePasswordResponse struct {
	Success bool `json:"success"`
}

// UsersEndpoint handles user-related endpoints
type UsersEndpoint struct {
	UserManager users.UserManager
}

// NewUsersEndpoint creates a new users endpoint
func NewUsersEndpoint(manager users.UserManager) *UsersEndpoint {
	return &UsersEndpoint{
		UserManager: manager,
	}
}

// CreateUser creates a new user
func (e *UsersEndpoint) CreateUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(CreateUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUIDs
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role ID format")
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, errors.New("invalid project ID format")
	}

	// Delegate to the user manager
	user, err := e.UserManager.CreateUser(ctx, req.Email, req.Password, req.FirstName, req.LastName, roleID, projectID)
	if err != nil {
		return nil, err
	}

	return CreateUserResponse{
		User: User{
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

// GetUser gets a user by ID
func (e *UsersEndpoint) GetUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(GetUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	userID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Delegate to the user manager
	user, err := e.UserManager.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return GetUserResponse{
		User: User{
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
	// Delegate to the user manager
	usersList, err := e.UserManager.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	users := make([]User, len(usersList))
	for i, u := range usersList {
		users[i] = User{
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

// UpdateUser updates a user
func (e *UsersEndpoint) UpdateUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(UpdateUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	userID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Delegate to the user manager
	user, err := e.UserManager.UpdateUser(ctx, userID, req.FirstName, req.LastName, req.Active)
	if err != nil {
		return nil, err
	}

	return UpdateUserResponse{
		User: User{
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

// DeleteUser deletes a user
func (e *UsersEndpoint) DeleteUser(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(DeleteUserRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	userID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Delegate to the user manager
	err = e.UserManager.DeleteUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return DeleteUserResponse{
		Success: true,
	}, nil
}

// ChangePassword changes a user's password
func (e *UsersEndpoint) ChangePassword(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(ChangePasswordRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	userID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Delegate to the user manager
	err = e.UserManager.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}

	return ChangePasswordResponse{
		Success: true,
	}, nil
}