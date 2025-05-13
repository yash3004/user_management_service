package http_transport

import (
	"context"
	"encoding/json"
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
)

func AddProjectRoutes(r *mux.Router, projects *endpoints.ProjectsEndpoint) {
	r.Methods("POST").Path("/create").Handler(kithttp.NewServer(
		projects.CreateProject,
		decodeCreateProjectRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("GET").Path("/get/{id}").Handler(kithttp.NewServer(
		projects.GetProject,
		decodeGetProjectRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("GET").Path("/list").Handler(kithttp.NewServer(
		projects.ListProjects,
		decodeListProjectsRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("PUT").Path("/update/{id}").Handler(kithttp.NewServer(
		projects.UpdateProject,
		decodeUpdateProjectRequest,
		encodeResponse,
		defaultServerOptions()...,
	))

	r.Methods("DELETE").Path("/delete/{id}").Handler(kithttp.NewServer(
		projects.DeleteProject,
		decodeDeleteProjectRequest,
		encodeResponse,
		defaultServerOptions()...,
	))
}

// Request decoders
func decodeCreateProjectRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request endpoints.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeGetProjectRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	return endpoints.GetProjectRequest{
		ID: vars["id"],
	}, nil
}

func decodeListProjectsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return endpoints.ListProjectsRequest{}, nil
}

func decodeUpdateProjectRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	var request endpoints.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	request.ID = vars["id"]
	return request, nil
}

func decodeDeleteProjectRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	return endpoints.DeleteProjectRequest{
		ID: vars["id"],
	}, nil
}