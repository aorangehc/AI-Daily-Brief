package fetch

import (
	"context"

	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
	"github.com/ai-daily-brief/ai-daily-brief/internal/source"
)

// Fetcher defines the interface for fetching items from a source
type Fetcher interface {
	// Fetch retrieves items from a source
	Fetch(ctx context.Context, src *source.Source) ([]*schema.RawItem, error)
	// Name returns the fetcher type name
	Name() string
}

// Factory creates a fetcher based on source type
func Factory(src *source.Source) Fetcher {
	switch src.Type {
	case "rss":
		return &RSSFetcher{}
	case "api":
		return &APIFetcher{}
	case "html":
		return &HTMLFetcher{}
	default:
		return nil
	}
}
