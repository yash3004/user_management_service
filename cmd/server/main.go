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
	OAuthManager       *endpoints.OAuthEndpoint
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
	endpointMgrs := createEndpointManagers(managers, cfg)

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

func createEndpointManagers(managers *allManager.Managers, cfg cmd.Config) *endpointManagers {
	OauthCfg := cfg.OAuth
	// Initialize OAuth providers
	providerConfigs := map[string]oauth.ProviderConfig{
		"google": {
			ClientID:     OauthCfg.Google.ClientID,
			ClientSecret: OauthCfg.Google.ClientSecret,
			RedirectURL:  OauthCfg.Google.RedirectURL,
			Scopes:       OauthCfg.Google.Scopes,
		},
		"facebook": {
			ClientID:     OauthCfg.Facebook.ClientID,
			ClientSecret: OauthCfg.Facebook.ClientSecret,
			RedirectURL:  OauthCfg.Facebook.RedirectURL,
			Scopes:       OauthCfg.Facebook.Scopes,
		},
	}

	providerFactory := oauth.NewProviderFactory(providerConfigs)

	return &endpointManagers{
		ProjectManager:     endpoints.NewProjectsEndpoint(managers.ProjectManager),
		RoleManager:        endpoints.NewRolesEndpoint(managers.RoleManager),
		PolicyManager:      endpoints.NewPoliciesEndpoint(managers.PolicyManager),
		UserManager:        endpoints.NewUsersEndpoint(managers.UserManager),
		ProjectUserManager: endpoints.NewProjectUsersEndpoint(managers.ProjectUserManager),
		OAuthManager:       endpoints.NewOAuthEndpoint(managers.UserManager, providerFactory),
		// Initialize other endpoint managers as needed
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

	projectUserRouter := apiRouter.PathPrefix("/{projectId}/users").Subrouter()
	http_transport.AddProjectUserRoutes(projectUserRouter, ep.ProjectUserManager)

	oauthRouter := apiRouter.PathPrefix("/oauth_user").Subrouter()
	http_transport.AddOAuthRoutes(oauthRouter, ep.OAuthManager)

	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}

		methods, err := route.GetMethods()
		if err != nil {
			return nil
		}

		klog.Infof("\t%v %s\n", methods, path)

		return nil
	})
	if err != nil {
		klog.Errorf("cannot print routes: %v", err)
	}

	return r
}
