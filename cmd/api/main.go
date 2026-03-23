package main

import (
	"log"
	"taip-flow-backend/internal/config"
	"taip-flow-backend/internal/db"
	"taip-flow-backend/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	database, err := db.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	router := server.SetupRouter(database)

	log.Printf("Starting server on port %s...", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
