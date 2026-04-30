package fetcher

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/iharee/websearch-mcp/internal/model"
)

type mockProvider struct {
	content string
	calls   *int32
}

func (m *mockProvider) Fetch(_ context.Context, _ string) (*model.FetchResult, error) {
	atomic.AddInt32(m.calls, 1)
	return &model.FetchResult{
		URL:     "https://example.com",
		Title:   "Example",
		Content: m.content,
	}, nil
}

func TestCachedFetcherHit(t *testing.T) {
	var calls int32
	mock := &mockProvider{
		content: strings.Repeat("hello world ", 100),
		calls:   &calls,
	}

	// Override cache config for testing — use unlimited entries, long TTL.
	c := NewCachedFetcher(mock)

	result1, err := c.Fetch(context.Background(), "https://example.com", "", false)
	if err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Error("first fetch should hit provider")
	}

	result2, err := c.Fetch(context.Background(), "https://example.com", "", false)
	if err != nil {
		t.Fatalf("second fetch failed: %v", err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Error("second fetch should be a cache hit, provider not called again")
	}

	if result2.URL != result1.URL || result2.Title != result1.Title {
		t.Error("cached result should match")
	}
}

func TestCachedFetcherNoCache(t *testing.T) {
	var calls int32
	mock := &mockProvider{
		content: "hello",
		calls:   &calls,
	}

	c := NewCachedFetcher(mock)

	_, _ = c.Fetch(context.Background(), "https://example.com", "", true)
	if atomic.LoadInt32(&calls) != 1 {
		t.Error("first call should hit provider")
	}

	_, _ = c.Fetch(context.Background(), "https://example.com", "", true)
	if atomic.LoadInt32(&calls) != 2 {
		t.Error("no_cache should skip cache, calling provider again")
	}
}

func TestCachedFetcherModeTruncation(t *testing.T) {
	mock := &mockProvider{
		content: strings.Repeat("0123456789", 500), // 5000 chars
		calls:   new(int32),
	}

	c := NewCachedFetcher(mock)

	t.Run("full", func(t *testing.T) {
		r, _ := c.Fetch(context.Background(), "https://x.com", "full", true)
		if len(r.Content) != 5000 {
			t.Errorf("full mode: got %d chars, want 5000", len(r.Content))
		}
	})

	t.Run("title", func(t *testing.T) {
		r, _ := c.Fetch(context.Background(), "https://x.com", "title", true)
		if len([]rune(r.Content)) > titlePreviewChars+3 {
			t.Errorf("title mode: got %d runes, max %d", len([]rune(r.Content)), titlePreviewChars)
		}
	})

	t.Run("summary", func(t *testing.T) {
		r, _ := c.Fetch(context.Background(), "https://x.com", "summary", true)
		if len([]rune(r.Content)) > summaryPreviewChars+3 {
			t.Errorf("summary mode: got %d runes, max %d", len([]rune(r.Content)), summaryPreviewChars)
		}
	})

	t.Run("default", func(t *testing.T) {
		r, _ := c.Fetch(context.Background(), "https://x.com", "", true)
		if len([]rune(r.Content)) > defaultPreviewChars+3 {
			t.Errorf("default mode: got %d runes, max %d", len([]rune(r.Content)), defaultPreviewChars)
		}
	})

	t.Run("short_no_ellipsis", func(t *testing.T) {
		shortMock := &mockProvider{content: "hello", calls: new(int32)}
		c2 := NewCachedFetcher(shortMock)
		r, _ := c2.Fetch(context.Background(), "https://x.com", "title", true)
		if strings.HasSuffix(r.Content, "...") {
			t.Error("short content should not end with ...")
		}
		if r.Content != "hello" {
			t.Errorf("short content unchanged: got %q, want %q", r.Content, "hello")
		}
	})

	t.Run("invalid_mode", func(t *testing.T) {
		r, err := c.Fetch(context.Background(), "https://x.com", "nonsense", true)
		if err == nil {
			t.Error("expected error for invalid mode, got nil")
		}
		if r != nil {
			t.Errorf("expected nil result for invalid mode, got %v", r)
		}
	})
}
