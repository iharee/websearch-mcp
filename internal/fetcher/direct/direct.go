package direct

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/iharee/websearch-mcp/internal/model"
	"github.com/iharee/websearch-mcp/internal/util"
)

const (
	userAgent      = "websearch-mcp/0.1"
	requestTimeout = 20 * time.Second
	maxRedirects   = 10
)

type Provider struct {
	client *http.Client
}

func NewProvider() *Provider {
	return &Provider{
		client: &http.Client{
			Timeout: requestTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return fmt.Errorf("stopped after %d redirects", maxRedirects)
				}
				return nil
			},
		},
	}
}

func (p *Provider) Fetch(ctx context.Context, rawURL string) (*model.FetchResult, error) {
	requestURL, err := normalizeURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	bodyStr := string(body)
	contentType := resp.Header.Get("Content-Type")

	return &model.FetchResult{
		URL:     resp.Request.URL.String(),
		Title:   extractTitle(bodyStr, contentType),
		Content: fullText(bodyStr, contentType),
	}, nil
}

func fullText(body, contentType string) string {
	return util.CollapseWhitespace(normalizeContent(body, contentType))
}

func normalizeURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	if parsed.Scheme == "http" {
		host := parsed.Hostname()
		if host != "localhost" && host != "127.0.0.1" && host != "::1" {
			parsed.Scheme = "https"
		}
	}
	return parsed.String(), nil
}

func normalizeContent(body, contentType string) string {
	if strings.Contains(contentType, "html") {
		return util.HTMLToText(body)
	}
	return strings.TrimSpace(body)
}

func extractTitle(body, contentType string) string {
	if strings.Contains(contentType, "html") {
		lower := strings.ToLower(body)
		start := strings.Index(lower, "<title>")
		if start == -1 {
			start = strings.Index(lower, "<title ")
		}
		if start != -1 {
			start += strings.Index(body[start:], ">") + 1
			end := strings.Index(lower, "</title>")
			if end != -1 && end > start {
				title := body[start:end]
				return util.CollapseWhitespace(html.UnescapeString(strings.TrimSpace(title)))
			}
		}
	}
	text := normalizeContent(body, contentType)
	if line, _, found := strings.Cut(text, "\n"); found {
		return strings.TrimSpace(line)
	}
	return strings.TrimSpace(text)
}

