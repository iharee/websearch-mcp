package fetcher

import (
	"context"
	"fmt"
	"strings"

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

		var buf strings.Builder
		fmt.Fprintf(&buf, "Title: %s\n", content.Title)
		fmt.Fprintf(&buf, "URL: %s\n", content.URL)
		if content.Content != "" {
			fmt.Fprintf(&buf, "\n%s", content.Content)
		}

		return &mcp.ToolCallResult{
			Content: []mcp.ContentItem{{Type: "text", Text: buf.String()}},
		}, nil
	}
}
