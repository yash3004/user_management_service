package main

import (
	"fmt"
	"log"

	cmd "github.com/yash3004/user_management_service/cmd"
	"github.com/yash3004/user_management_service/internal"
	"github.com/yash3004/user_management_service/internal/transport/endpoints"
	"k8s.io/klog/v2"
)

type endpointManager struct {
	users    *endpoints.UsersEndpoint
	projects *endpoints.ProjectsEndpoint
	roles    *endpoints.RolesEndpoint
}

func main() {
	//getting the configurations

	cfg := cmd.GetConfigurations()
	fmt.Print(cfg)
	//skipping the migration for now

	db, err := internal.CreateMySqlConnection(cfg)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			klog.Errorf("failed to close db connection: %v", err)
		}
	}()



}
