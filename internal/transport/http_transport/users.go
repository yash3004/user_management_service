package http_transport

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
	"github.com/yash3004/user_management_service/auth/oauth"
	"k8s.io/klog/v2"

	kithttp "github.com/go-kit/kit/transport/http"
)

func AddUserRoutes(r *mux.Router, ep *endpoints.UsersEndpoint) {

	// GET - List all users
	r.Methods("GET").Path("/{id}").Handler(kithttp.NewServer(
		ep.GetUser,
		decodeGetUserRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	// POST - Create new user
	r.Methods("POST").Path("").Handler(kithttp.NewServer(
		ep.CreateUser,
		decodeCreateUserRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("PUT").Path("/{id}").Handler(kithttp.NewServer(
		ep.UpdateUser,
		decodeUpdateUserRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("DELETE").Path("/{id}").Handler(kithttp.NewServer(
		ep.DeleteUser,
		decodeDeleteUserRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("POST").Path("/{id}/change-password").Handler(kithttp.NewServer(
		ep.ChangePassword,
		decodeChangePasswordRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	// Create OAuth subrouter
	oauthRouter := r.PathPrefix("/oauth").Subrouter()

	// Add OAuth routes for supported providers
	oauthRouter.Methods("GET").Path("/login/google").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleOAuthLogin(w, r, "google")
	})

	oauthRouter.Methods("GET").Path("/login/facebook").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleOAuthLogin(w, r, "facebook")
	})

	oauthRouter.Methods("GET").Path("/login/github").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleOAuthLogin(w, r, "github")
	})

	oauthRouter.Methods("GET").Path("/login/microsoft").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleOAuthLogin(w, r, "microsoft")
	})

	// Add OAuth callback routes
	oauthRouter.Methods("GET").Path("/callback/google").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleOAuthCallback(w, r, "google", ep)
	})

	oauthRouter.Methods("GET").Path("/callback/facebook").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleOAuthCallback(w, r, "facebook", ep)
	})

	oauthRouter.Methods("GET").Path("/callback/github").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleOAuthCallback(w, r, "github", ep)
	})

	oauthRouter.Methods("GET").Path("/callback/microsoft").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleOAuthCallback(w, r, "microsoft", ep)
	})
}

func decodeGetUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	return endpoints.GetUserRequest{ID: id}, nil
}

func decodeCreateUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	projectId, err := GetProjectIDFromRequest(r)
	if err != nil {
		klog.Errorf("Error getting project ID from request: %v", err)
		return nil, err
	}
	var req endpoints.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorf("Error decoding request body: %v", err)
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	req.ProjectID = projectId
	return req, nil
}

func decodeUpdateUserRequest(ctx_ context.Context, r *http.Request) (interface{}, error) {
	projectId, err := GetProjectIDFromRequest(r)
	if err != nil {
		klog.Errorf("Error getting project ID from request: %v", err)
		return nil, err
	}
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}

	var req endpoints.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	req.ID = id
	req.ProjectId = projectId

	return req, nil
}

func decodeDeleteUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	projectId, err := GetProjectIDFromRequest(r)
	if err != nil {
		klog.Errorf("Error getting project ID from request: %v", err)
		return nil, err
	}
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	return endpoints.DeleteUserRequest{ID: id, ProjectId: projectId}, nil
}

type errorer interface {
	error() error
}

// ErrBadRouting is returned when the route cannot be determined from the URL
var ErrBadRouting = errors.New("inconsistent mapping between route and handler")

func GetProjectIDFromRequest(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["projectId"]
	if !ok {
		return "", ErrBadRouting
	}
	return projectID, nil
}

func decodeChangePasswordRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}

	var req endpoints.ChangePasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	req.ID = id

	return req, nil
}

// OAuth handler functions
var oauthProviders *oauth.ProviderFactory

// InitOAuthProviders initializes the OAuth providers
func InitOAuthProviders(configs map[string]oauth.ProviderConfig) {
	oauthProviders = oauth.NewProviderFactory(configs)
}

// handleOAuthLogin handles the OAuth login request
func handleOAuthLogin(w http.ResponseWriter, r *http.Request, providerName string) {
	if oauthProviders == nil {
		http.Error(w, "OAuth providers not initialized", http.StatusInternalServerError)
		return
	}

	provider, err := oauthProviders.GetProvider(providerName)
	if err != nil {
		http.Error(w, "Provider not found: "+providerName, http.StatusBadRequest)
		return
	}

	// Generate a random state
	state := generateSecureState()

	// Store the state in a cookie for verification during callback
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Get the authorization URL
	authURL := provider.GetAuthURL(state)

	// Redirect the user to the authorization URL
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// handleOAuthCallback handles the OAuth callback request
func handleOAuthCallback(w http.ResponseWriter, r *http.Request, providerName string, ep *endpoints.UsersEndpoint) {
	if oauthProviders == nil {
		http.Error(w, "OAuth providers not initialized", http.StatusInternalServerError)
		return
	}

	// Get the provider
	provider, err := oauthProviders.GetProvider(providerName)
	if err != nil {
		http.Error(w, "Provider not found: "+providerName, http.StatusBadRequest)
		return
	}

	// Get the code from the request
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	// Get the state from the request
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "Missing state parameter", http.StatusBadRequest)
		return
	}

	// Verify the state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != state {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Exchange the code for a token
	token, err := provider.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the user info
	userInfo, err := provider.GetUserInfo(r.Context(), token)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get project ID from the request
	projectID, err := GetProjectIDFromRequest(r)
	if err != nil {
		http.Error(w, "Failed to get project ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse project ID
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		http.Error(w, "Invalid project ID format", http.StatusBadRequest)
		return
	}

	// Use a default role ID for OAuth users (this should be configurable)
	roleUUID, err := uuid.Parse("00000000-0000-0000-0000-000000000000") // Replace with a valid default role ID
	if err != nil {
		http.Error(w, "Invalid role ID format", http.StatusInternalServerError)
		return
	}

	// Create or update the user
	user, err := ep.UserManager.CreateOrUpdateOAuthUser(r.Context(), userInfo, projectUUID, roleUUID)
	if err != nil {
		http.Error(w, "Failed to create or update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate a token for the user
	userUUID := user.ID
	userID, err := uuid.Parse(userUUID)
	if err != nil {
		klog.Errorf("Error parsing user ID: %v", err)
		panic(err)
	}

	gentoken, expiresAt, err := ep.UserManager.GenerateToken(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the user and token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      gentoken,
		"expires_at": expiresAt,
		"user":       user,
	})
}

// generateSecureState generates a secure random state string
func generateSecureState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
