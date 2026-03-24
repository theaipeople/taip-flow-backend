package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"taip-flow-backend/internal/models"
)

// buildTopology constructs a ReactFlow topology from an ordered node list.
// Each node's Links field controls outgoing edges:
//
//	links=1 → linear to next node
//	links=2 → branches to next 2 nodes (with vertical offset)
//	links=0 → terminal, no outgoing edges
func buildTopology(nodes []models.AvailableNode) ([]byte, error) {
	type nodeShape struct {
		ID       string                 `json:"id"`
		Type     string                 `json:"type"`
		Position map[string]int         `json:"position"`
		Data     map[string]interface{} `json:"data"`
	}
	type edgeShape struct {
		ID     string `json:"id"`
		Source string `json:"source"`
		Target string `json:"target"`
		Label  string `json:"label,omitempty"`
	}

	rfNodes := make([]nodeShape, len(nodes))
	var rfEdges []edgeShape
	edgeIdx := 1
	yPositions := make([]int, len(nodes))
	for i := range yPositions {
		yPositions[i] = 100
	}

	for i, n := range nodes {
		nid := fmt.Sprintf("n%d", i+1)

		appearance := map[string]interface{}{
			"shape": "square", "bgColor": "#ffffff",
			"borderColor": "#e5e7eb", "borderWidth": 1,
			"paddingX": 16, "paddingY": 16, "width": 140, "height": 140,
		}
		if len(n.Appearance) > 0 {
			var stored map[string]interface{}
			if json.Unmarshal(n.Appearance, &stored) == nil {
				for k, v := range stored {
					appearance[k] = v
				}
			}
		}

		rfNodes[i] = nodeShape{
			ID:   nid,
			Type: "custom",
			Position: map[string]int{"x": 100 + i*220, "y": yPositions[i]},
			Data: map[string]interface{}{
				"instanceName": n.Name,
				"nodeId":       n.ID,
				"category":     n.Category,
				"priority":     i + 1,
				"maxLinks":     n.Links,
				"appearance":   appearance,
				"linksConfig": map[string]interface{}{
					"type": "smoothstep", "animated": true, "color": "#000000", "width": 2,
				},
				"properties": map[string]interface{}{},
			},
		}

		links := n.Links
		if links <= 0 {
			continue
		}
		if links == 1 && i+1 < len(nodes) {
			rfEdges = append(rfEdges, edgeShape{
				ID:     fmt.Sprintf("e%d", edgeIdx),
				Source: nid,
				Target: fmt.Sprintf("n%d", i+2),
			})
			edgeIdx++
		} else {
			for b := 0; b < links && i+1+b < len(nodes); b++ {
				targetIdx := i + 1 + b
				yPositions[targetIdx] = 100 + b*180
				rfEdges = append(rfEdges, edgeShape{
					ID:     fmt.Sprintf("e%d", edgeIdx),
					Source: nid,
					Target: fmt.Sprintf("n%d", targetIdx+1),
					Label:  fmt.Sprintf("Branch %d", b+1),
				})
				edgeIdx++
			}
		}
	}

	return json.Marshal(map[string]interface{}{"nodes": rfNodes, "edges": rfEdges})
}

func NewServer(db *gorm.DB) *server.MCPServer {
	s := server.NewMCPServer("taip-flow-mcp", "1.0.0")

	// ── Tool 1: list_nodes ──────────────────────────────────────────────────────
	// Merged list + search into one tool to keep total tool count low.
	// Smaller models (Llama on Groq) fail with tool_use_failed when >5 tools exist.
	s.AddTool(
		mcp.NewTool("list_nodes",
			mcp.WithDescription("List available workflow nodes. Pass an optional search keyword to filter by name or category. Returns id, name, category, links (max branches) for each node."),
			mcp.WithString("query",
				mcp.Description("Optional keyword to filter nodes, e.g. 'email' or 'condition'. Leave empty to list all."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, _ := req.Params.Arguments.(map[string]interface{})
			query, _ := args["query"].(string)

			var nodes []models.AvailableNode
			var err error
			if query != "" {
				err = db.Where("name LIKE ? OR category LIKE ?", "%"+query+"%", "%"+query+"%").Find(&nodes).Error
			} else {
				err = db.Find(&nodes).Error
			}
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("DB error: %v", err)), nil
			}
			if len(nodes) == 0 {
				if query != "" {
					return mcp.NewToolResultText(fmt.Sprintf("No nodes found matching '%s'. Use create_node to add one.", query)), nil
				}
				return mcp.NewToolResultText("No nodes registered yet. Use create_node to add nodes first."), nil
			}
			result := fmt.Sprintf("Available nodes (%d):\n\n", len(nodes))
			for _, n := range nodes {
				result += fmt.Sprintf("- id:%s | name:%s | category:%s | links:%d\n",
					n.ID, n.Name, n.Category, n.Links)
			}
			return mcp.NewToolResultText(result), nil
		},
	)

	// ── Tool 2: create_node ─────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("create_node",
			mcp.WithDescription("Register a new node type. Use when a needed node does not exist yet."),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Display name, e.g. 'Patient Admission'"),
			),
			mcp.WithString("category",
				mcp.Required(),
				mcp.Description("Group category: communication, condition, data, or trigger"),
			),
			mcp.WithNumber("links",
				mcp.Description("Outgoing branches: 1=linear (default), 2=binary branch, 0=terminal"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, _ := req.Params.Arguments.(map[string]interface{})
			name, _ := args["name"].(string)
			category, _ := args["category"].(string)

			if name == "" || category == "" {
				return mcp.NewToolResultError("'name' and 'category' are required"), nil
			}

			links := 1
			if v, ok := args["links"].(float64); ok {
				links = int(v)
			}

			appearance, _ := json.Marshal(map[string]interface{}{
				"shape": "square", "bgColor": "#ffffff",
				"borderColor": "#e5e7eb", "borderWidth": 1,
				"paddingX": 16, "paddingY": 16, "width": 140, "height": 140,
			})
			node := models.AvailableNode{
				ID:         uuid.New().String(),
				Name:       name,
				Category:   category,
				Links:      links,
				BaseNode:   false,
				Status:     "Active",
				Appearance: datatypes.JSON(appearance),
				Fields:     datatypes.JSON("[]"),
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			if err := db.Create(&node).Error; err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to create node: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf(
				"Node created! id:%s | name:%s | category:%s | links:%d",
				node.ID, node.Name, node.Category, node.Links,
			)), nil
		},
	)

	// ── Tool 3: create_workflow ─────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("create_workflow",
			mcp.WithDescription("Create a workflow from an ordered list of node IDs. Call list_nodes first to get IDs, create_node for any missing ones. The backend builds branching edges from each node's links count automatically."),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Workflow name"),
			),
			mcp.WithString("node_ids",
				mcp.Required(),
				mcp.Description("Comma-separated node IDs in order, e.g. 'uuid1,uuid2,uuid3'"),
			),
			mcp.WithString("status",
				mcp.Description("Draft or Active. Defaults to Draft"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, _ := req.Params.Arguments.(map[string]interface{})
			name, _ := args["name"].(string)
			nodeIDsStr, _ := args["node_ids"].(string)
			status, _ := args["status"].(string)

			if name == "" {
				return mcp.NewToolResultError("'name' is required"), nil
			}
			if nodeIDsStr == "" {
				return mcp.NewToolResultError("ERROR: node_ids is empty. You MUST call list_nodes first to get node IDs, then call create_node for any missing nodes, then retry create_workflow with the actual IDs."), nil
			}
			if status == "" {
				status = "Draft"
			}

			rawIDs := strings.Split(nodeIDsStr, ",")
			resolvedNodes := make([]models.AvailableNode, 0, len(rawIDs))
			for _, raw := range rawIDs {
				id := strings.TrimSpace(raw)
				if id == "" {
					continue
				}
				var node models.AvailableNode
				if err := db.First(&node, "id = ?", id).Error; err != nil {
					return mcp.NewToolResultError(fmt.Sprintf(
						"Node '%s' not found. Call list_nodes or create_node first.", id,
					)), nil
				}
				resolvedNodes = append(resolvedNodes, node)
			}
			if len(resolvedNodes) == 0 {
				return mcp.NewToolResultError("No valid node IDs provided"), nil
			}

			topologyBytes, err := buildTopology(resolvedNodes)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to build topology: %v", err)), nil
			}

			workflow := models.Workflow{
				ID:         uuid.New().String(),
				Name:       name,
				Status:     status,
				NodesCount: len(resolvedNodes),
				Topology:   datatypes.JSON(topologyBytes),
				Categories: datatypes.JSON("[]"),
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			if err := db.Create(&workflow).Error; err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to save workflow: %v", err)), nil
			}

			summary := ""
			for i, n := range resolvedNodes {
				if i > 0 {
					summary += " → "
				}
				if n.Links > 1 {
					summary += fmt.Sprintf("%s[x%d]", n.Name, n.Links)
				} else {
					summary += n.Name
				}
			}
			return mcp.NewToolResultText(fmt.Sprintf(
				"Workflow created! id:%s | name:%s | status:%s | nodes:%d\nTopology: %s",
				workflow.ID, workflow.Name, workflow.Status, workflow.NodesCount, summary,
			)), nil
		},
	)

	// ── Tool 4: update_workflow_status ──────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("update_workflow_status",
			mcp.WithDescription("Change the status of a workflow to Draft or Active"),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Workflow ID"),
			),
			mcp.WithString("status",
				mcp.Required(),
				mcp.Description("New status: Draft or Active"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, _ := req.Params.Arguments.(map[string]interface{})
			id, _ := args["id"].(string)
			status, _ := args["status"].(string)

			if id == "" || status == "" {
				return mcp.NewToolResultError("'id' and 'status' are required"), nil
			}
			var workflow models.Workflow
			if err := db.First(&workflow, "id = ?", id).Error; err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Workflow '%s' not found", id)), nil
			}
			if err := db.Model(&workflow).Updates(map[string]interface{}{
				"status": status, "updated_at": time.Now(),
			}).Error; err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to update: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Workflow '%s' status → '%s'.", workflow.Name, status)), nil
		},
	)

	return s
}
