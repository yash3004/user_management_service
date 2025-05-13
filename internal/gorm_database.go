package internal

import (
	"database/sql"

	"github.com/yash3004/user_management_service/cmd"
	"github.com/yash3004/user_management_service/internal/schemas"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// Global variable to store the GORM DB instance
var gormDBInstance *gorm.DB

func CreateMySqlConnection(cfg cmd.Config) (*sql.DB, error) {
	dsn := cfg.DB.CreateDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		klog.Fatalf("Failed to connect to the database: %v", err)
		return nil, err
	}
	
	// Store the GORM DB instance for later use
	gormDBInstance = db
	
	// Auto migrate schemas
	db.AutoMigrate(&schemas.Role{})
	db.AutoMigrate(&schemas.Policy{})
	db.AutoMigrate(&schemas.Project{})
	db.AutoMigrate(&schemas.User{}) // Make sure to migrate User schema as well

	return db.DB()
}

// GetGormDB returns the GORM DB instance
func GetGormDB(cfg cmd.Config) (*gorm.DB, error) {
	if gormDBInstance != nil {
		return gormDBInstance, nil
	}
	
	// If the instance doesn't exist, create a new connection
	dsn := cfg.DB.CreateDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		klog.Errorf("Failed to connect to the database: %v", err)
		return nil, err
	}
	
	gormDBInstance = db
	return db, nil
}
