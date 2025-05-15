package endpoints

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/auth/oauth"
	"github.com/yash3004/user_management_service/internal/models"
	projectusers "github.com/yash3004/user_management_service/project_users"
)

// OAuthLoginRequest represents the OAuth login request
type OAuthLoginRequest struct {
	Provider  string `json:"provider"`
	ProjectID string `json:"project_id"`
	RoleID    string `json:"role_id"`
	State     string `json:"state"`
}

// OAuthLoginResponse represents the OAuth login response
type OAuthLoginResponse struct {
	RedirectURL string `json:"redirect_url"`
}

// OAuthCallbackRequest represents the OAuth callback request
type OAuthCallbackRequest struct {
	Provider  string `json:"provider"`
	ProjectID string `json:"project_id"`
	Code      string `json:"code"`
	State     string `json:"state"`
	RoleID    string `json:"role_id"`
}

// OAuthCallbackResponse represents the OAuth callback response
type OAuthCallbackResponse struct {
	Token     string             `json:"token"`
	User      models.DisplayUser `json:"user"`
	ExpiresIn int64              `json:"expires_in"`
}

// OAuthEndpoint handles OAuth-related endpoints
type OAuthEndpoint struct {
	ProjectUser     projectusers.ProjectUserManager
	ProviderFactory *oauth.ProviderFactory
}

func NewOAuthEndpoint(userManager projectusers.ProjectUserManager, providerFactory *oauth.ProviderFactory) *OAuthEndpoint {
	return &OAuthEndpoint{
		ProjectUser:     userManager,
		ProviderFactory: providerFactory,
	}
}

// Login initiates the OAuth login flow
func (e *OAuthEndpoint) Login(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(OAuthLoginRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	provider, err := e.ProviderFactory.GetProvider(req.Provider)
	if err != nil {
		return nil, err
	}

	redirectURL := provider.GetAuthURL(req.State)

	return OAuthLoginResponse{
		RedirectURL: redirectURL,
	}, nil
}

func (e *OAuthEndpoint) Callback(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(OAuthCallbackRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	provider, err := e.ProviderFactory.GetProvider(req.Provider)
	if err != nil {
		return nil, err
	}

	// Exchange the code for a token
	token, err := provider.Exchange(ctx, req.Code)
	if err != nil {
		return nil, errors.New("failed to exchange code for token")
	}

	userInfo, err := provider.GetUserInfo(ctx, token)
	if err != nil {
		return nil, errors.New("failed to get user info")
	}

	projectID := req.ProjectID
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role ID format")
	}

	// Create or update the user in our system
	user, err := e.ProjectUser.CreateOrUpdateOAuthProjectUser(ctx, projectID, userInfo, roleID)
	if err != nil {
		return nil, err
	}

	// Generate a token for the user
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	jwtToken, expiresAt, err := e.ProjectUser.GenerateToken(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	return OAuthCallbackResponse{
		Token:     jwtToken,
		User:      *user,
		ExpiresIn: expiresAt.Unix() - time.Now().Unix(),
	}, nil
}
