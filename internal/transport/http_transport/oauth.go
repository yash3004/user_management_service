package http_transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/yash3004/user_management_service/auth"
	"github.com/yash3004/user_management_service/auth/oauth"
	"github.com/yash3004/user_management_service/internal/schemas"
	"github.com/yash3004/user_management_service/users"
)

// OAuthHandler handles OAuth authentication
type OAuthHandler struct {
	providerFactory *oauth.ProviderFactory
	sessionManager  *auth.SessionManager
	tokenManager    *auth.TokenManager
	userManager     *users.Manager
	baseURL         string
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	providerFactory *oauth.ProviderFactory,
	sessionManager *auth.SessionManager,
	tokenManager *auth.TokenManager,
	userManager *users.Manager,
	baseURL string,
) *OAuthHandler {
	return &OAuthHandler{
		providerFactory: providerFactory,
		sessionManager:  sessionManager,
		tokenManager:    tokenManager,
		userManager:     userManager,
		baseURL:         baseURL,
	}
}

// RegisterRoutes registers the OAuth routes
func (h *OAuthHandler) RegisterRoutes(r *mux.Router) {
	// Get available providers
	r.HandleFunc("/api/auth/providers", h.GetProviders).Methods("GET")

	// Start OAuth flow
	r.HandleFunc("/api/auth/oauth/{provider}", h.StartOAuth).Methods("GET")

	// OAuth callback
	r.HandleFunc("/api/auth/oauth/{provider}/callback", h.OAuthCallback).Methods("GET")

	// Refresh token
	r.HandleFunc("/api/auth/refresh", h.RefreshToken).Methods("POST")

	// Logout
	r.HandleFunc("/api/auth/logout", h.Logout).Methods("POST")
}

// GetProviders returns all available OAuth providers
func (h *OAuthHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.providerFactory.GetAllProviders()

	type ProviderInfo struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	providerInfo := make([]ProviderInfo, 0, len(providers))
	for name := range providers {
		providerInfo = append(providerInfo, ProviderInfo{
			Name: name,
			URL:  fmt.Sprintf("/api/auth/oauth/%s", name),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers": providerInfo,
	})
}

// StartOAuth starts the OAuth flow for a provider
func (h *OAuthHandler) StartOAuth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerName := vars["provider"]

	provider, err := h.providerFactory.GetProvider(providerName)
	if err != nil {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	// Generate a random state
	state := uuid.New().String()

	// Store the state in the session
	if err := h.sessionManager.SetOAuthState(w, r, state); err != nil {
		http.Error(w, "Failed to set state", http.StatusInternalServerError)
		return
	}

	// Redirect to provider's auth URL
	redirectURL := provider.GetAuthURL(state)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// OAuthCallback handles the OAuth callback
func (h *OAuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	providerName := vars["provider"]

	// Get the state and code
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		http.Error(w, "Missing state or code", http.StatusBadRequest)
		return
	}

	// Verify the state
	valid, err := h.sessionManager.VerifyOAuthState(r, state)
	if err != nil || !valid {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Get the provider
	provider, err := h.providerFactory.GetProvider(providerName)
	if err != nil {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	// Exchange the code for a token
	token, err := provider.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	// Store the token in the session
	if err := h.sessionManager.StoreOAuthToken(w, r, token); err != nil {
		http.Error(w, "Failed to store token", http.StatusInternalServerError)
		return
	}

	// Get user info from the provider
	userInfo, err := provider.GetUserInfo(ctx, token)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Find user by OAuth ID or email
	user, err := h.userManager.FindByOAuth(ctx, providerName, userInfo.ID)
	if err != nil {
		// User not found by OAuth ID, try email
		user, err = h.userManager.FindByEmail(ctx, userInfo.Email)
		if err != nil {
			// User not found, create a new one
			user = &schemas.User{
				ID:           uuid.New(),
				Email:        userInfo.Email,
				FirstName:    userInfo.FirstName,
				LastName:     userInfo.LastName,
				OAuthID:      userInfo.ID,
				OAuthType:    providerName,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				TokenExpiry:  token.Expiry,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			if err := h.userManager.Create(ctx, user); err != nil {
				http.Error(w, "Failed to create user", http.StatusInternalServerError)
				return
			}
		} else {
			// User found by email, update OAuth info
			user.OAuthID = userInfo.ID
			user.OAuthType = providerName
			user.AccessToken = token.AccessToken
			user.RefreshToken = token.RefreshToken
			user.TokenExpiry = token.Expiry
			user.UpdatedAt = time.Now()

			if err := h.userManager.Update(ctx, user); err != nil {
				http.Error(w, "Failed to update user", http.StatusInternalServerError)
				return
			}
		}
	} else {
		// User found by OAuth ID, update token
		user.AccessToken = token.AccessToken
		user.RefreshToken = token.RefreshToken
		user.TokenExpiry = token.Expiry
		user.UpdatedAt = time.Now()

		if err := h.userManager.Update(ctx, user); err != nil {
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}
	}

	// Login the user
	if err := h.sessionManager.Login(ctx, w, r, user); err != nil {
		http.Error(w, "Failed to login", http.StatusInternalServerError)
		return
	}

	// Generate JWT tokens
	tokenPair, err := h.tokenManager.GenerateTokens(user)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Determine redirect URL based on client type
	clientType := r.URL.Query().Get("client_type")
	if clientType == "api" {
		// Return tokens as JSON for API clients
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenPair)
	} else {
		// Redirect to dashboard for web clients
		http.Redirect(w, r, h.baseURL+"/dashboard", http.StatusTemporaryRedirect)
	}
}

// RefreshToken refreshes an access token
func (h *OAuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var requestBody struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the refresh token
	claims, err := h.tokenManager.ValidateToken(requestBody.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Get the user ID from the token
	userIDStr, ok := claims["sub"].(string)
	if !ok {
		http.Error(w, "Invalid token subject", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Get the user
	user, err := h.userManager.FindByID(ctx, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Generate new tokens
	tokenPair, err := h.tokenManager.RefreshTokens(requestBody.RefreshToken, user)
	if err != nil {
		http.Error(w, "Failed to refresh tokens", http.StatusUnauthorized)
		return
	}

	// Return the new tokens
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenPair)
}

// Logout logs out the user
func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.sessionManager.Logout(w, r); err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Logged out successfully",
	})
}
