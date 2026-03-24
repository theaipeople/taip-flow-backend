package controllers

import (
	"net/http"
	"taip-flow-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkflowController struct {
	DB *gorm.DB
}

func (ctrl *WorkflowController) GetWorkflows(c *gin.Context) {
	var workflows []models.Workflow
	if err := ctrl.DB.Find(&workflows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workflows"})
		return
	}
	c.JSON(http.StatusOK, workflows)
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
