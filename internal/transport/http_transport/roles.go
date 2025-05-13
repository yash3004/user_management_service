package http_transport

import (
	"context"
	"encoding/json"
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
)

func AddRoleRoutes(r *mux.Router, ep *endpoints.RolesEndpoint) {
	// GET - List all roles
	r.Methods("GET").Path("").Handler(kithttp.NewServer(
		ep.ListRoles,
		decodeListRolesRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("POST").Path("").Handler(kithttp.NewServer(
		ep.CreateRole,
		decodeCreateRoleRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("PUT").Path("/{id}").Handler(kithttp.NewServer(
		ep.UpdateRole,
		decodeUpdateRoleRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("DELETE").Path("/{id}").Handler(kithttp.NewServer(
		ep.DeleteRole,
		decodeDeleteRoleRequest,
		encodeResponse,
		defaultServerOptions()...,
	))
}

func decodeListRolesRequest(ctx_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.ListRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeUpdateRoleRequest(ctx_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}

	var req endpoints.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	req.ID = id

	return req, nil

}

func decodeCreateRoleRequest(ctx_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeDeleteRoleRequest(ctx_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}

	return endpoints.DeleteRoleRequest{
		ID: id,
	}, nil
}
