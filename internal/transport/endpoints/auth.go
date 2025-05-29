package endpoints

import (
	"context"
	"errors"

	"github.com/yash3004/user_management_service/internal/auth"
	"github.com/yash3004/user_management_service/internal/schemas"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type AuthEndpoint struct {
	DB *gorm.DB
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

func (e *AuthEndpoint) Login(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(LoginRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	var user schemas.User
	if err := e.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		klog.Errorf("Database error: %v", err)
		return nil, errors.New("internal server error")
	}

	if !user.Active {
		return nil, errors.New("account is inactive")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	var role schemas.Role
	if err := e.DB.First(&role, "id = ?", user.RoleId).Error; err != nil {
		klog.Errorf("Error fetching role: %v", err)
		return nil, errors.New("internal server error")
	}

	token, err := auth.GenerateToken(user.ID, user.Email, role.ID, user.ProjectId, user.ExpirationTime)
	if err != nil {
		klog.Errorf("Error generating token: %v", err)
		return nil, errors.New("failed to generate authentication token")
	}

	return LoginResponse{
		Token:     token,
		UserID:    user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      role.Name,
	}, nil
}
