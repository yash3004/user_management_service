package http_transport

import (
	"context"
	"encoding/json"
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
	"gorm.io/gorm"
)

// AddAuthRoutes adds authentication routes to the router
func AddAuthRoutes(r *mux.Router, db *gorm.DB) {
	authEndpoint := &endpoints.AuthEndpoint{DB: db}

	r.Methods("POST").Path("/login").Handler(kithttp.NewServer(
		authEndpoint.Login,
		decodeLoginRequest,
		encodeResponse,
		defaultServerOptions()...,
	))
}

// decodeLoginRequest decodes the login request from HTTP request
func decodeLoginRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request endpoints.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}