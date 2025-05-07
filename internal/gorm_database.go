package internal

import (
	"database/sql"

	"github.com/yash3004/user_management_service/cmd"
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
	return db.DB()

}
