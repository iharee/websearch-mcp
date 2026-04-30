package direct

import (
	"context"
	"strings"
	"testing"
)

const testURL = "https://iharee.github.io/2026/03/22/mathematical_principles_of_transformer/"

func TestFetch(t *testing.T) {
	p := NewProvider()

	result, err := p.Fetch(context.Background(), testURL)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if result.URL == "" {
		t.Error("URL is empty")
	}
	if result.Title == "" {
		t.Error("Title is empty")
	}
	if result.Content == "" {
		t.Error("Content is empty")
	}

	if strings.Contains(result.Content, "<") && strings.Contains(result.Content, ">") {
		t.Error("Content appears to contain un-stripped HTML tags")
	}

	t.Logf("URL: %s", result.URL)
	t.Logf("Title: %s", result.Title)
	t.Logf("Content length: %d chars", len([]rune(result.Content)))
}

func TestHTMLToText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain_text", "hello world", "hello world"},
		{"simple_tag", "<p>hello</p>", "hello"},
		{"nested_tags", "<div><span>text</span></div>", "text"},
		{"entities", "a&amp;b &lt; c &gt; d &quot;e&quot; &#39;f&#39;", `a&b < c > d "e" 'f'`},
		{"nbsp", "hello&nbsp;world", "hello world"},
		{"whitespace_collapse", "hello   \n\t  world", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := htmlToText(tt.in)
			got = collapseWhitespace(got)
			if got != tt.want {
				t.Errorf("htmlToText(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	t.Run("from_title_tag", func(t *testing.T) {
		html := "<html><head><title>My Page</title></head><body></body></html>"
		got := extractTitle(html, "text/html")
		if got != "My Page" {
			t.Errorf("got %q, want %q", got, "My Page")
		}
	})

	t.Run("from_first_line", func(t *testing.T) {
		text := "First line\nSecond line"
		got := extractTitle(text, "text/plain")
		if got != "First line" {
			t.Errorf("got %q, want %q", got, "First line")
		}
	})
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"https://example.com", "https://example.com", false},
		{"http://example.com", "https://example.com", false},
		{"http://localhost:8080/path", "http://localhost:8080/path", false},
		{"http://127.0.0.1:8080/path", "http://127.0.0.1:8080/path", false},
		{"example.com/path", "https://example.com/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := normalizeURL(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q", tt.in)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("normalizeURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
