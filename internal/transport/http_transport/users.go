package http_transport

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
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

func GetRoleIdFromRequest(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	roleID, ok := vars["roleId"]
	if !ok {
		return "", ErrBadRouting
	}
	return roleID, nil
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

// generateSecureState generates a secure random state string
func generateSecureState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
