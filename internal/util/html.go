package util

import (
	"html"
	"strings"
)

func HTMLToText(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	inTag := false

	for _, ch := range s {
		switch {
		case ch == '<':
			inTag = true
		case ch == '>':
			inTag = false
		case inTag:
		default:
			buf.WriteRune(ch)
		}
	}

	decoded := html.UnescapeString(buf.String())
	decoded = strings.ReplaceAll(decoded, " ", " ")
	return strings.Join(strings.Fields(decoded), " ")
}

func CollapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
