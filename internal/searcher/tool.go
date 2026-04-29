package searcher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iharee/websearch-mcp-server/internal/mcp"
)

func ToolDefinition() mcp.Tool {
	return mcp.Tool{
		Name:        "search",
		Description: "Search the web and return a list of results with URL, title, and snippet.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.JSONSchema{
				"query": {
					Type:        "string",
					Description: "Search query",
				},
			},
			Required: []string{"query"},
		},
	}
}

func Handler(p Provider) mcp.ToolHandler {
	return func(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResult, error) {
		query, ok := args["query"].(string)
		if !ok || query == "" {
			return &mcp.ToolCallResult{
				Content: []mcp.ContentItem{{Type: "text", Text: "missing required argument: query"}},
				IsError: true,
			}, nil
		}

		results, err := p.Search(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("search failed: %w", err)
		}

		data, err := json.Marshal(results)
		if err != nil {
			return nil, fmt.Errorf("marshal results: %w", err)
		}

		return &mcp.ToolCallResult{
			Content: []mcp.ContentItem{{Type: "text", Text: string(data)}},
		}, nil
	}
}
