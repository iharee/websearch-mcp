package fetcher

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/iharee/websearch-mcp/internal/config"
	"github.com/iharee/websearch-mcp/internal/fetcher/cdp"
	"github.com/iharee/websearch-mcp/internal/fetcher/direct"
	"github.com/iharee/websearch-mcp/internal/mcp"
)

var (
	directFetcher *CachedFetcher
	cdpFetcher    *CachedFetcher
	fetcherOnce   sync.Once
)

func initFetchers() {
	fetcherOnce.Do(func() {
		directFetcher = NewCachedFetcher(direct.NewProvider())
		cdpFetcher = NewCachedFetcher(cdp.NewProvider())
	})
}

func ToolDefinition() mcp.Tool {
	return mcp.Tool{
		Name:        "fetch_content",
		Description: "Fetch a URL, convert HTML to readable text, and return content based on the prompt intent. Use 'method' parameter to select direct or cdp.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.JSONSchema{
				"url": {
					Type:        "string",
					Description: "URL of the page to fetch",
				},
				"mode": {
					Type:        "string",
					Description: "Content length mode: 'full' (complete), 'summary' (longer preview), 'title' (short preview). Defaults to a 900-char preview.",
				},
				"method": {
					Type:        "string",
					Description: "Fetch method: direct or cdp (case-insensitive). Defaults to FETCH_METHOD env var or direct.",
				},
				"no_cache": {
					Type:        "boolean",
					Description: "If true, bypass the cache and force a fresh fetch. Defaults to false.",
				},
			},
			Required: []string{"url"},
		},
	}
}

func Handler() mcp.ToolHandler {
	return func(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResult, error) {
		url, ok := args["url"].(string)
		if !ok || url == "" {
			return &mcp.ToolCallResult{
				Content: []mcp.ContentItem{{Type: "text", Text: "missing required argument: url"}},
				IsError: true,
			}, nil
		}

		mode := ""
		if m, ok := args["mode"].(string); ok {
			mode = strings.TrimSpace(m)
		}

		noCache := false
		if nc, ok := args["no_cache"].(bool); ok {
			noCache = nc
		}

		initFetchers()
		fetcher := resolveFetcher(args)

		content, err := fetcher.Fetch(ctx, url, mode, noCache)
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

func resolveFetcher(args map[string]interface{}) *CachedFetcher {
	method := ""
	if m, ok := args["method"].(string); ok {
		method = strings.ToLower(strings.TrimSpace(m))
	}
	if method == "" {
		method = config.FetchMethod()
	}

	switch method {
	case "cdp":
		return cdpFetcher
	default:
		return directFetcher
	}
}
