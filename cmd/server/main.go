package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	allManager "github.com/yash3004/user_management_service"
	"github.com/yash3004/user_management_service/auth/oauth"
	cmd "github.com/yash3004/user_management_service/cmd"
	"github.com/yash3004/user_management_service/internal"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
	"github.com/yash3004/user_management_service/internal/transport/http_transport"
	"k8s.io/klog/v2"
)

type endpointManagers struct {
	ProjectManager     *endpoints.ProjectsEndpoint
	RoleManager        *endpoints.RolesEndpoint
	PolicyManager      *endpoints.PoliciesEndpoint
	UserManager        *endpoints.UsersEndpoint
	ProjectUserManager *endpoints.ProjectUsersEndpoint
}

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

	managers := allManager.NewManagers(gormDB)

	// Create endpoint managers
	endpointMgrs := createEndpointManagers(managers)

	// Create HTTP handler without authentication
	handler := httpHandler(endpointMgrs)

	// Start the server
	port := cfg.Bind.HTTP

	srv := &http.Server{
		Handler:      handler,
		Addr:         ":" + fmt.Sprint(port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	klog.Infof("Starting server on port %s", port)
	log.Fatal(srv.ListenAndServe())
}

func createEndpointManagers(managers *allManager.Managers) *endpointManagers {

	return &endpointManagers{
		ProjectManager:     endpoints.NewProjectsEndpoint(managers.ProjectManager),
		RoleManager:        endpoints.NewRolesEndpoint(managers.RoleManager),
		PolicyManager:      endpoints.NewPoliciesEndpoint(managers.PolicyManager),
		UserManager:        endpoints.NewUsersEndpoint(managers.UserManager),
		ProjectUserManager: endpoints.NewProjectUsersEndpoint(managers.ProjectUserManager),
	}
}

func httpHandler(ep *endpointManagers) http.Handler {
	r := mux.NewRouter()

	apiRouter := r.PathPrefix("/api").Subrouter()

	projectRouter := apiRouter.PathPrefix("/projects").Subrouter()
	http_transport.AddProjectRoutes(projectRouter, ep.ProjectManager)

	rolesRouter := apiRouter.PathPrefix("/roles").Subrouter()
	http_transport.AddRoleRoutes(rolesRouter, ep.RoleManager)

	policiesRouter := apiRouter.PathPrefix("/policies").Subrouter()
	http_transport.AddPolicyRoutes(policiesRouter, ep.PolicyManager)

	// Global user routes
	userRouter := apiRouter.PathPrefix("/users").Subrouter()
	http_transport.AddUserRoutes(userRouter, ep.UserManager)

	// Project-specific user routes
	projectUserRouter := apiRouter.PathPrefix("/projects/{project_id}/users").Subrouter()
	http_transport.AddProjectUserRoutes(projectUserRouter, ep.ProjectUserManager)

	// Initialize OAuth providers
	initOAuthProviders()

	return r
}

// initOAuthProviders initializes the OAuth providers
func initOAuthProviders() {
	// Get configurations from config file
	cfg := cmd.GetConfigurations()

	// Create OAuth provider configurations
	oauthConfigs := map[string]oauth.ProviderConfig{
		"google": {
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			RedirectURL:  cfg.OAuth.Google.RedirectURL,
			Scopes:       cfg.OAuth.Google.Scopes,
		},
		"facebook": {
			ClientID:     cfg.OAuth.Facebook.ClientID,
			ClientSecret: cfg.OAuth.Facebook.ClientSecret,
			RedirectURL:  cfg.OAuth.Facebook.RedirectURL,
			Scopes:       cfg.OAuth.Facebook.Scopes,
		},
		"github": {
			ClientID:     cfg.OAuth.GitHub.ClientID,
			ClientSecret: cfg.OAuth.GitHub.ClientSecret,
			RedirectURL:  cfg.OAuth.GitHub.RedirectURL,
			Scopes:       cfg.OAuth.GitHub.Scopes,
		},
		"microsoft": {
			ClientID:     cfg.OAuth.Microsoft.ClientID,
			ClientSecret: cfg.OAuth.Microsoft.ClientSecret,
			RedirectURL:  cfg.OAuth.Microsoft.RedirectURL,
			Scopes:       cfg.OAuth.Microsoft.Scopes,
		},
	}

	// Initialize OAuth providers
	http_transport.InitOAuthProviders(oauthConfigs)
}
