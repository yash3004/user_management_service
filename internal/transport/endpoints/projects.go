package endpoints

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/projects"
	"k8s.io/klog/v2"
)

// Project represents a project in the response
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UniqueID    string    `json:"unique_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProjectRequest represents the create project request
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UniqueID    string `json:"unique_id"`
}

// CreateProjectResponse represents the create project response
type CreateProjectResponse struct {
	Project Project `json:"project"`
}

// GetProjectRequest represents the get project request
type GetProjectRequest struct {
	ID string `json:"id"`
}

// GetProjectResponse represents the get project response
type GetProjectResponse struct {
	Project Project `json:"project"`
}

// ListProjectsRequest represents the list projects request
type ListProjectsRequest struct {
	// Add pagination parameters if needed
}

// ListProjectsResponse represents the list projects response
type ListProjectsResponse struct {
	Projects []Project `json:"projects"`
}

// UpdateProjectRequest represents the update project request
type UpdateProjectRequest struct {
	ID          string `json:"-"` // From URL path
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateProjectResponse represents the update project response
type UpdateProjectResponse struct {
	Project Project `json:"project"`
}

// DeleteProjectRequest represents the delete project request
type DeleteProjectRequest struct {
	ID string `json:"id"`
}

// DeleteProjectResponse represents the delete project response
type DeleteProjectResponse struct {
	Success bool `json:"success"`
}

// ProjectsEndpoint handles project-related endpoints
type ProjectsEndpoint struct {
	ProjectManager projects.ProjectManager
}

// NewProjectsEndpoint creates a new projects endpoint
func NewProjectsEndpoint(manager projects.ProjectManager) *ProjectsEndpoint {
	return &ProjectsEndpoint{
		ProjectManager: manager,
	}
}

// CreateProject creates a new project
func (e *ProjectsEndpoint) CreateProject(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(CreateProjectRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Delegate to the project manager
	project, err := e.ProjectManager.CreateProject(ctx, req.Name, req.Description, req.UniqueID)
	if err != nil {
		return nil, err
	}

	return CreateProjectResponse{
		Project: Project{
			ID:          project.ID.String(),
			Name:        project.Name,
			Description: project.Description,
			UniqueID:    project.UniqueID,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		},
	}, nil
}

// GetProject gets a project by ID
func (e *ProjectsEndpoint) GetProject(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(GetProjectRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	projectID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid project ID format")
	}

	// Delegate to the project manager
	project, err := e.ProjectManager.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return GetProjectResponse{
		Project: Project{
			ID:          project.ID.String(),
			Name:        project.Name,
			Description: project.Description,
			UniqueID:    project.UniqueID,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		},
	}, nil
}

// ListProjects lists all projects
func (e *ProjectsEndpoint) ListProjects(ctx context.Context, request interface{}) (interface{}, error) {
	// Delegate to the project manager
	projectsList, err := e.ProjectManager.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	projects := make([]Project, len(projectsList))
	for i, p := range projectsList {
		projects[i] = Project{
			ID:          p.ID.String(),
			Name:        p.Name,
			Description: p.Description,
			UniqueID:    p.UniqueID,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	return ListProjectsResponse{
		Projects: projects,
	}, nil
}

// UpdateProject updates a project
func (e *ProjectsEndpoint) UpdateProject(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(UpdateProjectRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	projectID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid project ID format")
	}

	// Delegate to the project manager
	project, err := e.ProjectManager.UpdateProject(ctx, projectID, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	return UpdateProjectResponse{
		Project: Project{
			ID:          project.ID.String(),
			Name:        project.Name,
			Description: project.Description,
			UniqueID:    project.UniqueID,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		},
	}, nil
}

// DeleteProject deletes a project
func (e *ProjectsEndpoint) DeleteProject(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(DeleteProjectRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	projectID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid project ID format")
	}

	// Delegate to the project manager
	err = e.ProjectManager.DeleteProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return DeleteProjectResponse{
		Success: true,
	}, nil
}