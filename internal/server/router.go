package server

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"taip-flow-backend/internal/auth"
	"taip-flow-backend/internal/config"
	"taip-flow-backend/internal/controllers"
	mcpHandler "taip-flow-backend/internal/mcp"

	"github.com/mark3labs/mcp-go/server"
)

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cache-Control"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/ping", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(500, gin.H{"status": "Database Unreachable"})
			return
		}
		c.JSON(200, gin.H{"message": "pong"})
	})

	// ── Auth (public) ───────────────────────────────────────────────────────
	authSvc := auth.NewService(db, cfg)
	authCtrl := controllers.NewAuthController(authSvc, cfg)

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authCtrl.Register)
		authGroup.POST("/login", authCtrl.Login)
		authGroup.POST("/logout", authCtrl.Logout)
		authGroup.POST("/refresh", authCtrl.Refresh)

		// Protected — needs valid access token
		authGroup.GET("/me", auth.RequireAuth(authSvc), authCtrl.Me)

		// Google OAuth2 stubs — wire up when ready
		authGroup.GET("/google", authCtrl.GoogleLogin)
		authGroup.GET("/google/callback", authCtrl.GoogleCallback)
	}

	// ── API (protected) ─────────────────────────────────────────────────────
	workflowCtrl := &controllers.WorkflowController{DB: db}
	nodeCtrl := &controllers.NodeController{DB: db}
	agentCtrl := controllers.NewAgentController(db)

	api := router.Group("/api/v1", auth.RequireAuth(authSvc))
	{
		workflows := api.Group("/workflows")
		{
			workflows.GET("", workflowCtrl.GetWorkflows)
			workflows.GET("/:id", workflowCtrl.GetWorkflow)
			workflows.POST("", workflowCtrl.CreateWorkflow)
			workflows.POST("/bulk-delete", workflowCtrl.BulkDeleteWorkflows)
			workflows.PATCH("/:id", workflowCtrl.UpdateWorkflow)
			workflows.DELETE("/:id", workflowCtrl.DeleteWorkflow)
		}

		nodes := api.Group("/nodes")
		{
			nodes.GET("", nodeCtrl.GetNodes)
			nodes.POST("", nodeCtrl.CreateNode)
			nodes.POST("/bulk-delete", nodeCtrl.BulkDeleteNodes)
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

	// ── MCP (internal — protected by network, not auth) ─────────────────────
	mcpServer := mcpHandler.NewServer(db)
	sse := server.NewSSEServer(mcpServer, server.WithMessageEndpoint("/mcp/messages"))
	router.GET("/mcp/sse", gin.WrapH(sse.SSEHandler()))
	router.POST("/mcp/messages", gin.WrapH(sse.MessageHandler()))

	return router
}
