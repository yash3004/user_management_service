package http_transport

import (
	"context"
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"

	"github.com/yash3004/user_management_service/internal/transport/endpoints"
)

func AddPolicyRoutes(r *mux.Router, ep *endpoints.PoliciesEndpoint) {
	// GET - List all policies
	r.Methods("GET").Path("").Handler(kithttp.NewServer(
		ep.ListPolicies,
		decodeListPoliciesRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	// POST - Create new policy
	r.Methods("POST").Path("").Handler(kithttp.NewServer(
		ep.CreatePolicy,
		decodeCreatePolicyRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("PUT").Path("/{id}").Handler(kithttp.NewServer(
		ep.UpdatePolicy,
		decodeUpdatePolicyRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("DELETE").Path("/{id}").Handler(kithttp.NewServer(
		ep.DeletePolicy,
		decodeDeletePolicyRequest,
		encodeResponse,
		defaultServerOptions()...,
	))
}

func decodeListPoliciesRequest(ctx_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeCreatePolicyRequest(ctx_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeUpdatePolicyRequest(ctx_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeDeletePolicyRequest(ctx_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}
