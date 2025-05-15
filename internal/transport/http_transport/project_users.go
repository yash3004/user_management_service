package http_transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
	"k8s.io/klog/v2"

	kithttp "github.com/go-kit/kit/transport/http"
)

// AddProjectUserRoutes adds project-specific user routes to the router
func AddProjectUserRoutes(r *mux.Router, ep *endpoints.ProjectUsersEndpoint) {
	// GET - Get a specific user in a project
	r.Methods("GET").Path("/{user_id}").Handler(kithttp.NewServer(
		ep.GetProjectUser,
		decodeGetProjectUserRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	// GET - List all users in a project
	r.Methods("GET").Path("").Handler(kithttp.NewServer(
		ep.ListProjectUsers,
		decodeListProjectUsersRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	// POST - Create a new user in a project
	r.Methods("POST").Path("/{roleId}").Handler(kithttp.NewServer(
		ep.CreateProjectUser,
		decodeCreateProjectUserRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	// PUT - Update a user in a project
	r.Methods("PUT").Path("/{user_id}").Handler(kithttp.NewServer(
		ep.UpdateProjectUser,
		decodeUpdateProjectUserRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	// DELETE - Delete a user from a project
	r.Methods("DELETE").Path("/{user_id}").Handler(kithttp.NewServer(
		ep.DeleteProjectUser,
		decodeDeleteProjectUserRequest,
		encodeResponse,
		defaultServerOptions()...,
	))
}

// decodeGetProjectUserRequest decodes the get project user request
func decodeGetProjectUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, err := GetProjectIDFromRequest(r)
	if err != nil {
		klog.Errorf("Error getting project ID from request: %v", err)
		return nil, err
	}

	userID, ok := vars["user_id"]
	if !ok {
		return nil, ErrBadRouting
	}

	return endpoints.GetProjectUserRequest{
		ProjectID: projectID,
		UserID:    userID,
	}, nil
}

// decodeListProjectUsersRequest decodes the list project users request
func decodeListProjectUsersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	projectID, err := GetProjectIDFromRequest(r)
	if err != nil {
		klog.Errorf("Error getting project ID from request: %v", err)
		return nil, err
	}

	return endpoints.ListProjectUsersRequest{
		ProjectID: projectID,
	}, nil
}

// decodeCreateProjectUserRequest decodes the create project user request
func decodeCreateProjectUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	projectID, err := GetProjectIDFromRequest(r)
	if err != nil {
		klog.Errorf("Error getting project ID from request: %v", err)
		return nil, err
	}

	roleId, err := GetRoleIdFromRequest(r)
	if err != nil {
		klog.Errorf("Error getting role ID from request: %v", err)
		return nil, err
	}

	var req endpoints.CreateProjectUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorf("Error decoding request body: %v", err)
		return nil, err
	}

	req.ProjectID = projectID
	req.RoleID = roleId
	return req, nil
}

// decodeUpdateProjectUserRequest decodes the update project user request
func decodeUpdateProjectUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, err := GetProjectIDFromRequest(r)
	if err != nil {
		klog.Errorf("Error getting project ID from request: %v", err)
		return nil, err
	}

	userID, ok := vars["user_id"]
	if !ok {
		return nil, ErrBadRouting
	}

	var req endpoints.UpdateProjectUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorf("Error decoding request body: %v", err)
		return nil, err
	}

	req.ProjectID = projectID
	req.UserID = userID
	return req, nil
}

// decodeDeleteProjectUserRequest decodes the delete project user request
func decodeDeleteProjectUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, err := GetProjectIDFromRequest(r)
	if err != nil {
		klog.Errorf("Error getting project ID from request: %v", err)
		return nil, err
	}

	userID, ok := vars["user_id"]
	if !ok {
		return nil, ErrBadRouting
	}

	return endpoints.DeleteProjectUserRequest{
		ProjectID: projectID,
		UserID:    userID,
	}, nil
}
