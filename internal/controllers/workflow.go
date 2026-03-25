package controllers

import (
	"net/http"
	"strconv"
	"taip-flow-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkflowController struct {
	DB *gorm.DB
}

func (ctrl *WorkflowController) GetWorkflows(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := ctrl.DB.Model(&models.Workflow{})
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var workflows []models.Workflow
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&workflows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workflows"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       workflows,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (int(total) + limit - 1) / limit,
	})
}

func (ctrl *WorkflowController) GetWorkflow(c *gin.Context) {
	id := c.Param("id")
	var workflow models.Workflow
	if err := ctrl.DB.First(&workflow, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}
	c.JSON(http.StatusOK, workflow)
}

func (ctrl *WorkflowController) CreateWorkflow(c *gin.Context) {
	var workflow models.Workflow
	if err := c.ShouldBindJSON(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctrl.DB.Create(&workflow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}
	c.JSON(http.StatusCreated, workflow)
}

func (ctrl *WorkflowController) UpdateWorkflow(c *gin.Context) {
	id := c.Param("id")
	var workflow models.Workflow
	if err := ctrl.DB.First(&workflow, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}
	var updateData models.Workflow
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctrl.DB.Model(&workflow).Updates(&updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workflow"})
		return
	}
	c.JSON(http.StatusOK, workflow)
}

func (ctrl *WorkflowController) DeleteWorkflow(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.DB.Delete(&models.Workflow{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workflow"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (ctrl *WorkflowController) BulkDeleteWorkflows(c *gin.Context) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids array is required"})
		return
	}
	if err := ctrl.DB.Delete(&models.Workflow{}, "id IN ?", body.IDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workflows"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": len(body.IDs)})
}
