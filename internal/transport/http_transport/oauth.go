package http_transport

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
	"k8s.io/klog/v2"

	kithttp "github.com/go-kit/kit/transport/http"
)

// AddOAuthRoutes adds OAuth routes to the router
func AddOAuthRoutes(r *mux.Router, ep *endpoints.OAuthEndpoint) {
	// OAuth login routes for different providers
	r.Methods("GET").Path("/login/{provider}").Handler(kithttp.NewServer(
		ep.Login,
		decodeOAuthLoginRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	// OAuth callback routes for different providers
	r.Methods("GET").Path("/callback/{provider}").Handler(kithttp.NewServer(
		ep.Callback,
		decodeOAuthCallbackRequest,
		encodeResponse,
		defaultServerOptions()...,
	))
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

	// Generate or get state from the request
	state := r.URL.Query().Get("state")
	if state == "" {
		state = generateRandomState()
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

	projectID, err := GetProjectIDFromRequest(r)
	if err != nil {
		klog.Errorf("Error getting project ID from request: %v", err)
		return nil, err
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

	return endpoints.OAuthCallbackRequest{
		Provider:  provider,
		ProjectID: projectID,
		Code:      code,
		State:     state,
	}, nil
}

func generateRandomState() string {
	return "random-state"
}
