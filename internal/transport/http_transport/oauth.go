package http_transport

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
	"k8s.io/klog/v2"
)

// OAuthState represents the data structure to be encoded in the OAuth state parameter
type OAuthState struct {
	ProjectID string `json:"project_id"`
	Nonce     string `json:"nonce"` // For CSRF protection
}

func AddOAuthRoutes(r *mux.Router, ep *endpoints.OAuthEndpoint) {
	r.Methods("GET").Path("/{projectId}/login/{provider}").Handler(kithttp.NewServer(
		ep.Login,
		decodeOAuthLoginRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("GET").Path("/callback/{provider}").Handler(kithttp.NewServer(
		ep.Callback,
		decodeOAuthCallbackRequest,
		encodeResponse,
		defaultServerOptions()...,
	))
}

// generateNonce creates a random string for CSRF protection
func generateNonce(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// encodeOAuthState encodes project ID and other data into a secure state string
func encodeOAuthState(projectID string) (string, error) {
	// Generate a random nonce for CSRF protection
	nonce, err := generateNonce(16)
	if err != nil {
		return "", err
	}

	// Create state object with project ID and nonce
	stateObj := OAuthState{
		ProjectID: projectID,
		Nonce:     nonce,
	}

	// Serialize to JSON
	stateJSON, err := json.Marshal(stateObj)
	if err != nil {
		return "", err
	}

	// Encode to base64 URL-safe string
	return base64.URLEncoding.EncodeToString(stateJSON), nil
}

// decodeOAuthState decodes the state parameter back into structured data
func decodeOAuthState(state string) (*OAuthState, error) {
	// Decode from base64
	stateJSON, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var stateObj OAuthState
	if err := json.Unmarshal(stateJSON, &stateObj); err != nil {
		return nil, err
	}

	return &stateObj, nil
}

// decodeOAuthLoginRequest decodes the OAuth login request
func decodeOAuthLoginRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	provider, ok := vars["provider"]
	if !ok {
		return nil, ErrBadRouting
	}

	projectID, err := GetProjectIDFromRequest(r)
	if err != nil {
		klog.Errorf("Error getting project ID from request: %v", err)
		return nil, err
	}

	// Encode project ID into state parameter
	state, err := encodeOAuthState(projectID)
	if err != nil {
		klog.Errorf("Error encoding OAuth state: %v", err)
		return nil, err
	}

	return endpoints.OAuthLoginRequest{
		Provider:  provider,
		ProjectID: projectID,
		State:     state,
	}, nil
}

// decodeOAuthCallbackRequest decodes the OAuth callback request
func decodeOAuthCallbackRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	provider, ok := vars["provider"]
	if !ok {
		return nil, ErrBadRouting
	}

	// Get code and state from the request
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, errors.New("missing code parameter")
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		return nil, errors.New("missing state parameter")
	}

	// Decode the state to extract the project ID
	stateObj, err := decodeOAuthState(state)
	if err != nil {
		klog.Errorf("Error decoding OAuth state: %v", err)
		return nil, errors.New("invalid state parameter")
	}

	// Extract the project ID from the state
	projectID := stateObj.ProjectID

	return endpoints.OAuthCallbackRequest{
		Provider:  provider,
		ProjectID: projectID,
		Code:      code,
		State:     state, // Pass original state for verification if needed
	}, nil
}

// In your main code where you handle the OAuth response, ensure your OAuthEndpoint.Login
// uses the state that was generated in decodeOAuthLoginRequest
