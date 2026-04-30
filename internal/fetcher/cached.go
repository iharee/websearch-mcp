package fetcher

import (
	"context"
	"strings"

	"github.com/iharee/websearch-mcp/internal/cache"
	"github.com/iharee/websearch-mcp/internal/config"
	"github.com/iharee/websearch-mcp/internal/model"
)

const (
	defaultPreviewChars = 900
	summaryPreviewChars = 1200
	titlePreviewChars   = 600
)

// CachedFetcher wraps a Provider with an LRU+TTL cache and content truncation.
type CachedFetcher struct {
	inner        Provider
	cache        *cache.Cache[*model.FetchResult]
	maxEntrySize int
}

// NewCachedFetcher creates a CachedFetcher wrapping the given provider.
func NewCachedFetcher(inner Provider) *CachedFetcher {
	return &CachedFetcher{
		inner: inner,
		cache: cache.New[*model.FetchResult](
			config.CacheMaxEntries(),
			config.CacheTTL(),
		),
		maxEntrySize: config.CacheMaxEntrySize(),
	}
}

// Fetch returns page content, using the cache when available and noCache is false. Content is truncated according to mode.
func (c *CachedFetcher) Fetch(ctx context.Context, url string, mode string, noCache bool) (*model.FetchResult, error) {
	if !noCache {
		if cached, ok := c.cache.Get(url); ok {
			return &model.FetchResult{
				URL:     cached.URL,
				Title:   cached.Title,
				Content: truncateByMode(cached.Content, mode),
			}, nil
		}
	}

	full, err := c.inner.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	if !noCache && c.maxEntrySize > 0 && len(full.Content) <= c.maxEntrySize {
		c.cache.Put(url, full)
	}

	return &model.FetchResult{
		URL:     full.URL,
		Title:   full.Title,
		Content: truncateByMode(full.Content, mode),
	}, nil
}

func truncateByMode(fullText, mode string) string {
	lower := strings.ToLower(mode)
	switch {
	case strings.Contains(lower, "full"):
		return fullText
	case strings.Contains(lower, "title"):
		return previewText(fullText, titlePreviewChars)
	case strings.Contains(lower, "summary") || strings.Contains(lower, "summarize"):
		return previewText(fullText, summaryPreviewChars)
	default:
		return previewText(fullText, defaultPreviewChars)
	}
}

func previewText(s string, maxChars int) string {
	runes := []rune(s)
	if len(runes) <= maxChars {
		return s
	}
	return strings.TrimSpace(string(runes[:maxChars])) + "..."
}
