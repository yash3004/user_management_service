package http_transport

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"

	kithttp "github.com/go-kit/kit/transport/http"
)

func AddUserRoutes(r *mux.Router, ep *endpoints.UsersEndpoint) {
	r.Methods("GET").Path("").Handler(kithttp.NewServer(
		httpcontr
	))
}
