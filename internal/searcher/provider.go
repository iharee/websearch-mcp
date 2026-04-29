package searcher

import "context"

type SearchResult struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

type Provider interface {
	Search(ctx context.Context, query string) ([]SearchResult, error)
}
