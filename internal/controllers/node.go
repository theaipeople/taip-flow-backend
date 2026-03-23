package controllers

import (
	"net/http"
	"taip-flow-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NodeController struct {
	DB *gorm.DB
}

func (ctrl *NodeController) GetNodes(c *gin.Context) {
	var nodes []models.AvailableNode
	if err := ctrl.DB.Find(&nodes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch nodes"})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

func (ctrl *NodeController) CreateNode(c *gin.Context) {
	var node models.AvailableNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.DB.Create(&node).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create node"})
		return
	}
	c.JSON(http.StatusCreated, node)
}

func (ctrl *NodeController) UpdateNode(c *gin.Context) {
	id := c.Param("id")
	var node models.AvailableNode
	if err := ctrl.DB.First(&node, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.DB.Model(&node).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update node"})
		return
	}
	c.JSON(http.StatusOK, node)
}

func (ctrl *NodeController) DeleteNode(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.DB.Delete(&models.AvailableNode{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete node"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
