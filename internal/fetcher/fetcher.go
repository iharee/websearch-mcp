package fetcher

import "context"

type Content struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Fetcher interface {
	Fetch(ctx context.Context, url string) (*Content, error)
}
