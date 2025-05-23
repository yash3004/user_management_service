package allManager

import (
	"github.com/yash3004/user_management_service/policies"
	projectusers "github.com/yash3004/user_management_service/project_users"
	"github.com/yash3004/user_management_service/projects"
	"github.com/yash3004/user_management_service/roles"
	"github.com/yash3004/user_management_service/users"
	"gorm.io/gorm"
)

// Managers holds all the service managers
type Managers struct {
	UserManager        users.UserManager
	ProjectManager     projects.ProjectManager
	RoleManager        roles.RoleManager
	PolicyManager      policies.PolicyManager
	ProjectUserManager projectusers.ProjectUserManager
	DB                 *gorm.DB
}

// NewManagers creates a new instance of all managers
func NewManagers(db *gorm.DB) *Managers {
	return &Managers{
		UserManager:        users.NewManager(db),
		ProjectManager:     projects.NewManager(db),
		RoleManager:        roles.NewManager(db),
		PolicyManager:      policies.NewManager(db),
		ProjectUserManager: projectusers.NewManager(db),
		DB:                 db,
	}
}
