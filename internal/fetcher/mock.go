package fetcher

import "context"

type MockFetcher struct{}

func (MockFetcher) Fetch(_ context.Context, url string) (*Content, error) {
	return &Content{
		URL:     url,
		Title:   "Mock Page Title",
		Content: "",
	}, nil
}
