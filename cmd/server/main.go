package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	cmd "github.com/yash3004/user_management_service/cmd"
	"github.com/yash3004/user_management_service/internal"
	"github.com/yash3004/user_management_service/internal/auth"
	"github.com/yash3004/user_management_service/internal/superuser"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
	"github.com/yash3004/user_management_service/internal/transport/http_transport"
	"github.com/yash3004/user_management_service/policies"
	"github.com/yash3004/user_management_service/projects"
	"github.com/yash3004/user_management_service/roles"
	"github.com/yash3004/user_management_service/users"
	"github.com/yash3004/user_management_service"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

func main() {
	//getting the configurations
	cfg := cmd.GetConfigurations()
	fmt.Print(cfg)
	
	//skipping the migration for now
	sqlDB, err := internal.CreateMySqlConnection(cfg)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			klog.Errorf("failed to close db connection: %v", err)
		}
	}()
	
	// Get the GORM DB instance for creating the super user
	gormDB, err := internal.GetGormDB(cfg)
	if err != nil {
		log.Fatalf("failed to get gorm DB: %v", err)
	}
	
	// Initialize super user
	superUserConfig := superuser.DefaultSuperUserConfig()
	if err := superuser.EnsureSuperUser(gormDB, superUserConfig); err != nil {
		klog.Errorf("Failed to ensure super user: %v", err)
	} else {
		klog.Info("Super user initialization completed successfully")
	}

	// Initialize all managers
	managers := main.NewManagers(gormDB)

	// Initialize router
	r := mux.NewRouter()
	
	// Add authentication routes
	authRouter := r.PathPrefix("/auth").Subrouter()
	http_transport.AddAuthRoutes(authRouter, gormDB)
	
	// Create API router with authentication middleware
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(auth.AuthMiddleware(gormDB))
	
	// Create project-specific subrouters
	projectRouter := apiRouter.PathPrefix("/projects").Subrouter()
	projectsEndpoint := endpoints.NewProjectsEndpoint(managers.ProjectManager)
	projectRouter.Use(auth.PolicyMiddleware(gormDB, "projects", "*"))
	http_transport.AddProjectRoutes(projectRouter, projectsEndpoint)
	
	// Add policy middleware for roles routes
	rolesRouter := apiRouter.PathPrefix("/roles").Subrouter()
	rolesRouter.Use(auth.PolicyMiddleware(gormDB, "roles", "*"))
	rolesEndpoint := endpoints.NewRolesEndpoint(managers.RoleManager)
	http_transport.AddRoleRoutes(rolesRouter, rolesEndpoint)
	
	// Add policy middleware for policies routes
	policiesRouter := apiRouter.PathPrefix("/policies").Subrouter()
	policiesRouter.Use(auth.PolicyMiddleware(gormDB, "policies", "*"))
	policiesEndpoint := endpoints.NewPoliciesEndpoint(managers.PolicyManager)
	http_transport.AddPolicyRoutes(policiesRouter, policiesEndpoint)
	
	// Create user subrouter under project
	userRouter := apiRouter.PathPrefix("/users").Subrouter()
	usersEndpoint := endpoints.NewUsersEndpoint(managers.UserManager)
	http_transport.AddUserRoutes(userRouter, usersEndpoint)
	
	// Start the server
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}
	
	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	
	klog.Infof("Starting server on port %s", port)
	log.Fatal(srv.ListenAndServe())
}