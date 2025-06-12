package mcpgenerator

import (
	"context"
	"fmt"
	"github.com/centralmind/gateway/mcp"
	"github.com/centralmind/gateway/model"
	"github.com/centralmind/gateway/xcontext"
)

func (s *MCPServer) Tools() []model.Endpoint {
	return s.tools
}

func (s *MCPServer) SetTools(tools []model.Endpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var names []string
	for _, t := range tools {
		names = append(names, t.MCPMethod)
	}
	s.server.DeleteTools(names...)
	for _, endpoint := range tools {
		var opts []mcp.ToolOption
		if endpoint.Description != "" {
			opts = append(opts, mcp.WithDescription(endpoint.Description))
		} else if endpoint.Summary != "" {
			opts = append(opts, mcp.WithDescription(endpoint.Summary))
		}
		for _, col := range endpoint.Params {
			if col.Required {
				opts = append(opts, ArgumentOption(col, mcp.Required()))
			} else {
				opts = append(opts, ArgumentOption(col))
			}
		}

		if endpoint.MCPMethod == "search" {
			// Add OutputSchema options specifically for the 'search' tool
			opts = append(opts, mcp.WithOutputSchemaType("object")) // Top-level output is an object
			opts = append(opts, mcp.WithOutputArrayProperty(
				"results", // Name of the array property
				// Schema for the items within the "results" array
				map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":    map[string]interface{}{"type": "string", "description": "ID of the resource."},
						"title": map[string]interface{}{"type": "string", "description": "Title or headline of the resource."},
						"text":  map[string]interface{}{"type": "string", "description": "Text snippet or summary from the resource."},
						"url":   map[string]interface{}{"type": []string{"string", "null"}, "description": "URL of the resource. Optional but needed for citations to work."},
					},
					"required": []string{"id", "title", "text"}, // Properties required within each item object
				},
				// Options for the "results" property itself
				mcp.Required(), // Marks the "results" array as a required property of the output schema
				mcp.Description("A list of matching resources."), // Description for the "results" array property
			))
		}

		s.server.AddTool(mcp.NewTool(
			endpoint.MCPMethod,
			opts...,
		), s.endpoint(endpoint))
	}
	s.tools = tools
}

func (s *MCPServer) endpoint(endpoint model.Endpoint) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arg := request.Params.Arguments
		for _, param := range endpoint.Params {
			if _, ok := arg[param.Name]; !ok {
				arg[param.Name] = nil
			}
		}
		resData, err := s.connector.Query(ctx, endpoint, request.Params.Arguments)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Unable to query: %s", err),
					},
				},
				IsError: true,
			}, nil
		}
		var res []map[string]interface{}
	MAIN:
		for _, row := range resData {
			for _, interceptor := range s.interceptors {
				r, skip := interceptor.Process(row, xcontext.Headers(ctx))
				if skip {
					continue MAIN
				}
				row = r
			}
			res = append(res, row)
		}
		var content []mcp.Content
		content = append(content, mcp.TextContent{
			Type: "text",
			Text: fmt.Sprintf("Found a %v row-(s) in %s.", len(res), endpoint.Group),
		})
		for _, row := range res {
			content = append(content, mcp.TextContent{
				Type: "text",
				Text: jsonify(row),
			})
		}

		return &mcp.CallToolResult{
			Content: content,
		}, nil
	}
}

func ArgumentOption(col model.EndpointParams, opts ...mcp.PropertyOption) mcp.ToolOption {
	opts = append(opts, mcp.Title(fmt.Sprintf("Column %s", col.Name)))
	opts = append(opts, func(m map[string]interface{}) {
		m["default"] = col.Default
	})

	switch col.Type {
	case "integer", "double", "float", "number":
		return mcp.WithNumber(col.Name, opts...)

	case "string":
		return mcp.WithString(col.Name, opts...)

	case "bool", "boolean":
		return mcp.WithBoolean(col.Name, opts...)

	default:
		return mcp.WithString(col.Name)
	}
}
