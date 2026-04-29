package fetcher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iharee/websearch-mcp-server/internal/mcp"
)

func ToolDefinition() mcp.Tool {
	return mcp.Tool{
		Name:        "fetch_content",
		Description: "Fetch the full text content of a web page by URL.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.JSONSchema{
				"url": {
					Type:        "string",
					Description: "URL of the page to fetch",
				},
			},
			Required: []string{"url"},
		},
	}
}

func Handler(f Fetcher) mcp.ToolHandler {
	return func(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResult, error) {
		url, ok := args["url"].(string)
		if !ok || url == "" {
			return &mcp.ToolCallResult{
				Content: []mcp.ContentItem{{Type: "text", Text: "missing required argument: url"}},
				IsError: true,
			}, nil
		}

		content, err := f.Fetch(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("fetch failed: %w", err)
		}

		data, err := json.Marshal(content)
		if err != nil {
			return nil, fmt.Errorf("marshal content: %w", err)
		}

		return &mcp.ToolCallResult{
			Content: []mcp.ContentItem{{Type: "text", Text: string(data)}},
		}, nil
	}
}
