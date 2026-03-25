package controllers

import (
	"net/http"
	"strconv"
	"taip-flow-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NodeController struct {
	DB *gorm.DB
}

func (ctrl *NodeController) GetNodes(c *gin.Context) {
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

	query := ctrl.DB.Model(&models.AvailableNode{})
	if search != "" {
		query = query.Where("name LIKE ? OR category LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var nodes []models.AvailableNode
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&nodes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch nodes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       nodes,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (int(total) + limit - 1) / limit,
	})
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
	var updateData models.AvailableNode
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctrl.DB.Model(&node).Updates(&updateData).Error; err != nil {
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

func (ctrl *NodeController) BulkDeleteNodes(c *gin.Context) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids array is required"})
		return
	}
	if err := ctrl.DB.Delete(&models.AvailableNode{}, "id IN ?", body.IDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete nodes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": len(body.IDs)})
}
