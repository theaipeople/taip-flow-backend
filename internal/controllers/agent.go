package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"taip-flow-backend/internal/models"
)

type AgentController struct {
	DB *gorm.DB
}

func NewAgentController(db *gorm.DB) *AgentController {
	return &AgentController{DB: db}
}

func (c *AgentController) GetAgents(ctx *gin.Context) {
	var agents []models.Agent
	c.DB.Find(&agents)
	// Parse the tools JSON string back to array for the response
	type AgentResponse struct {
		models.Agent
		Tools []string `json:"tools"`
	}
	var result []AgentResponse
	for _, a := range agents {
		var tools []string
		json.Unmarshal([]byte(a.Tools), &tools)
		if tools == nil {
			tools = []string{}
		}
		result = append(result, AgentResponse{Agent: a, Tools: tools})
	}
	if result == nil {
		result = []AgentResponse{}
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *AgentController) CreateAgent(ctx *gin.Context) {
	var input struct {
		Name         string   `json:"name"`
		Category     string   `json:"category"`
		LLM          string   `json:"llm"`
		SystemPrompt string   `json:"systemPrompt"`
		Temperature  float64  `json:"temperature"`
		MaxTokens    int      `json:"maxTokens"`
		Tools        []string `json:"tools"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	toolsJSON, _ := json.Marshal(input.Tools)
	agent := models.Agent{
		ID:           uuid.New().String(),
		Name:         input.Name,
		Category:     input.Category,
		LLM:          input.LLM,
		SystemPrompt: input.SystemPrompt,
		Temperature:  input.Temperature,
		MaxTokens:    input.MaxTokens,
		Tools:        string(toolsJSON),
	}
	c.DB.Create(&agent)
	ctx.JSON(http.StatusCreated, agent)
}

func (c *AgentController) UpdateAgent(ctx *gin.Context) {
	id := ctx.Param("id")
	var agent models.Agent
	if c.DB.Where("id = ?", id).First(&agent).Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}
	var input struct {
		Name         string   `json:"name"`
		Category     string   `json:"category"`
		LLM          string   `json:"llm"`
		SystemPrompt string   `json:"systemPrompt"`
		Temperature  float64  `json:"temperature"`
		MaxTokens    int      `json:"maxTokens"`
		Tools        []string `json:"tools"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	toolsJSON, _ := json.Marshal(input.Tools)
	c.DB.Model(&agent).Updates(map[string]interface{}{
		"name": input.Name, "category": input.Category,
		"llm": input.LLM, "system_prompt": input.SystemPrompt,
		"temperature": input.Temperature, "max_tokens": input.MaxTokens,
		"tools": string(toolsJSON), "updated_at": time.Now(),
	})
	ctx.JSON(http.StatusOK, agent)
}

func (c *AgentController) DeleteAgent(ctx *gin.Context) {
	id := ctx.Param("id")
	c.DB.Where("id = ?", id).Delete(&models.Agent{})
	ctx.JSON(http.StatusOK, gin.H{"message": "Agent deleted"})
}
