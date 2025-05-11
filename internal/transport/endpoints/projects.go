package endpoints

import(
	allManager "github.com/yash3004/user_management_service"
)

type ProjectsEndpoint struct {
	ProjectsManager allManager.ProjectManager
}

