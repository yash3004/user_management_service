package http_transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
	"k8s.io/klog/v2"

	kithttp "github.com/go-kit/kit/transport/http"
)

func AddUserRoutes(r *mux.Router, ep *endpoints.UsersEndpoint) {

	// GET - List all users
	r.Methods("GET").Path("/{id}").Handler(kithttp.NewServer(
		ep.GetUserByID,
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

	// PUT - Update existing user
	r.Methods("PUT").Path("/{id}").Handler(kithttp.NewServer(
		ep.UpdateUser,
		decodeUpdateUserRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	// DELETE - Delete a user
	r.Methods("DELETE").Path("/{id}").Handler(kithttp.NewServer(
		ep.DeleteUser,
		decodeDeleteUserRequest,
		encodeResponse,
		defaultServerOptions()...,
	))
}

func decodeGetUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	Id, err := uuid.Parse(id)
	if err != nil {
		klog.Errorf("Error parsing UUID: %v", err)
	}
	if !ok {
		return nil, ErrBadRouting
	}
	return endpoints.GetUserRequest{ID: Id}, nil
}

func decodeRegisterUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// Extract projectId from the URL path
	projectID, err := GetProjectIDFromRequest(r)
	if err != nil {
		return nil, err
	}
	
	// Decode the request body
	var req endpoints.RegisterUserRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	
	// Add the projectID to the request
	req.ProjectID = projectID
	
	return req, nil
}

func decodeLoginUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// Extract projectId from the URL path
	projectID, err := GetProjectIDFromRequest(r)
	if err != nil {
		return nil, err
	}
	
	// Decode the request body
	var req endpoints.LoginUserRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	
	// Add the projectID to the request
	req.ProjectID = projectID
	
	return req, nil
}

func decodeCreateUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func decodeUpdateUserRequest(ctx_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}

	var req endpoints.UpdateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	req.ID = id

	return req, nil
}

func decodeDeleteUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	return endpoints.DeleteUserRequest{ID: id}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

// ErrBadRouting is returned when the route cannot be determined from the URL
var ErrBadRouting = errors.New("inconsistent mapping between route and handler")

// GetProjectIDFromRequest extracts the projectId from the URL path
// For routes like /{projectId}/user/register or /{projectId}/user/login
func GetProjectIDFromRequest(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["projectId"]
	if !ok {
		return "", ErrBadRouting
	}
	return projectID, nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case endpoints.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case endpoints.ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func defaultServerOptions() []kithttp.ServerOption {
	return []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}
}
