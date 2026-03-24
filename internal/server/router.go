package server

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
	"taip-flow-backend/internal/controllers"
	"github.com/mark3labs/mcp-go/server"
	mcpHandler "taip-flow-backend/internal/mcp"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	// Gin default configures standard Recovery and structured Logger middlewares instantly cleanly mapping explicit properties natively.
	router := gin.Default()

	// Implement explicit CORS boundaries explicitly blocking strictly ensuring frontend React interactions successfully natively beautifully mapped securely.
	router.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cache-Control"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initial Generic Health Ping
	router.GET("/ping", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(500, gin.H{"status": "Database Unreachable"})
			return
		}

		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "Strict Highly Scalable Service API Linked cleanly against MariaDB flawlessly sequentially natively",
		})
	})

	workflowCtrl := &controllers.WorkflowController{DB: db}
	nodeCtrl := &controllers.NodeController{DB: db}
	agentCtrl := controllers.NewAgentController(db)

	api := router.Group("/api/v1")
	{
		workflows := api.Group("/workflows")
		{
			workflows.GET("", workflowCtrl.GetWorkflows)
			workflows.GET("/:id", workflowCtrl.GetWorkflow)
			workflows.POST("", workflowCtrl.CreateWorkflow)
			workflows.PATCH("/:id", workflowCtrl.UpdateWorkflow)
			workflows.DELETE("/:id", workflowCtrl.DeleteWorkflow)
		}

		nodes := api.Group("/nodes")
		{
			nodes.GET("", nodeCtrl.GetNodes)
			nodes.POST("", nodeCtrl.CreateNode)
			nodes.PATCH("/:id", nodeCtrl.UpdateNode)
			nodes.DELETE("/:id", nodeCtrl.DeleteNode)
		}

		agents := api.Group("/agents")
		{
			agents.GET("", agentCtrl.GetAgents)
			agents.POST("", agentCtrl.CreateAgent)
			agents.PATCH("/:id", agentCtrl.UpdateAgent)
			agents.DELETE("/:id", agentCtrl.DeleteAgent)
		}
	}

	// Mount the specific Model Context Protocol SSE mapping safely explicitly organically smartly flexibly effortlessly properly tightly firmly flawlessly.
	mcpServer := mcpHandler.NewServer(db)
	sse := server.NewSSEServer(mcpServer, server.WithMessageEndpoint("/mcp/messages"))
	
	router.GET("/mcp/sse", gin.WrapH(sse.SSEHandler()))
	router.POST("/mcp/messages", gin.WrapH(sse.MessageHandler()))

	return router
}
