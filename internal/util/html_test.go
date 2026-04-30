package util

import (
	"testing"
)

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
		{"standard_lib_entity", "&copy; 2024 &reg;", "© 2024 ®"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HTMLToText(tt.in)
			if got != tt.want {
				t.Errorf("HTMLToText(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCollapseWhitespace(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no_extra_whitespace", "hello world", "hello world"},
		{"multiple_spaces", "hello   world", "hello world"},
		{"tabs_and_newlines", "hello\t\n  world", "hello world"},
		{"leading_trailing", "  hello world  ", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CollapseWhitespace(tt.in)
			if got != tt.want {
				t.Errorf("CollapseWhitespace(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
