package internal

import (
	"database/sql"

	"github.com/yash3004/user_management_service/cmd"
	"github.com/yash3004/user_management_service/internal/schemas"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

func CreateMySqlConnection(cfg cmd.Config) (*sql.DB, error) {
	dsn := cfg.DB.CreateDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		klog.Fatalf("Failed to connect to the database: %v", err)
		return nil, err
	}
	db.AutoMigrate(&schemas.Role{})
	//for policies
	db.AutoMigrate(&schemas.Policy{})
	// for projects
	db.AutoMigrate(&schemas.Project{})

	return db.DB()

}
