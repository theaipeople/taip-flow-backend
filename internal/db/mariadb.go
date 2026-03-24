package db

import (
	"log"
	"taip-flow-backend/internal/config"
	"taip-flow-backend/internal/models"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Initialize(cfg *config.Config) (*gorm.DB, error) {
	var database *gorm.DB
	var err error

	// Retry connection loop ensuring Docker MariaDB natively starts sequentially
	for i := 0; i < 10; i++ {
		database, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			break
		}
		log.Printf("Waiting for MariaDB... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	// Safely auto-migrate strictly scaling schemas dynamically gracefully
	if err := database.AutoMigrate(&models.Workflow{}, &models.AvailableNode{}, &models.Agent{}); err != nil {
		log.Printf("Warning: Failed to auto-migrate schemas: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, err
	}

	// Optimize connection pooling for extreme scalability
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("MariaDB connected successfully and connection pool optimized.")

	return database, nil
}
