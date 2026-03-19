package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
	"github.com/ai-daily-brief/ai-daily-brief/internal/source"
)

// APIFetcher fetches items from REST APIs (e.g., Hacker News)
type APIFetcher struct {
	client *http.Client
}

// HNItem represents a Hacker News item from the API
type HNItem struct {
	ID      int64  `json:"id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	By      string `json:"by"`
	Time    int64  `json:"time"`
	Score   int    `json:"score"`
	Descendants int `json:"descendants"`
}

// Name returns the fetcher type name
func (f *APIFetcher) Name() string {
	return "api"
}

// NewAPIFetcher creates a new API fetcher
func NewAPIFetcher() *APIFetcher {
	return &APIFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Fetch retrieves items from an API source
func (f *APIFetcher) Fetch(ctx context.Context, src *source.Source) ([]*schema.RawItem, error) {
	if src.APIURL == "" {
		return nil, fmt.Errorf("no api_url specified for source %s", src.ID)
	}

	switch src.Parser {
	case "hn_api":
		return f.fetchHackerNews(ctx, src)
	default:
		return nil, fmt.Errorf("unknown api parser: %s", src.Parser)
	}
}

// fetchHackerNews fetches top stories from Hacker News
func (f *APIFetcher) fetchHackerNews(ctx context.Context, src *source.Source) ([]*schema.RawItem, error) {
	baseURL := src.APIURL
	limit := src.RateLimitPerRun
	if limit <= 0 {
		limit = 30
	}

	// Get top story IDs
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/topstories.json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch topstories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HN API returned status %d: %s", resp.StatusCode, string(body))
	}

	var storyIDs []int64
	if err := json.NewDecoder(resp.Body).Decode(&storyIDs); err != nil {
		return nil, fmt.Errorf("failed to decode story IDs: %w", err)
	}

	// Limit number of stories to fetch
	if len(storyIDs) > limit {
		storyIDs = storyIDs[:limit]
	}

	// Fetch story details
	var items []*schema.RawItem
	seq := 0

	for _, id := range storyIDs {
		item, err := f.fetchHNItem(ctx, baseURL, id, src)
		if err != nil {
			// Log error but continue
			continue
		}
		if item != nil {
			seq++
			item.ID = fmt.Sprintf("%s_%s_%03d", src.ID, time.Now().Format("2006-01-02"), seq)
			items = append(items, item)
		}

		// Rate limiting: small delay between requests
		select {
		case <-ctx.Done():
			return items, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}

	return items, nil
}

// fetchHNItem fetches a single Hacker News item
func (f *APIFetcher) fetchHNItem(ctx context.Context, baseURL string, id int64, src *source.Source) (*schema.RawItem, error) {
	url := fmt.Sprintf("%s/item/%d.json", baseURL, id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HN item API returned status %d", resp.StatusCode)
	}

	var hnItem HNItem
	if err := json.NewDecoder(resp.Body).Decode(&hnItem); err != nil {
		return nil, err
	}

	// Skip job postings, comments, etc.
	if hnItem.Type != "story" || hnItem.Title == "" {
		return nil, nil
	}

	return &schema.RawItem{
		SourceID:    src.ID,
		SourceType:  "api",
		Title:       cleanText(hnItem.Title),
		URL:         normalizeURL(hnItem.URL),
		Author:      hnItem.By,
		PublishedAt: time.Unix(hnItem.Time, 0).UTC().Format(time.RFC3339),
		CollectedAt: time.Now().UTC().Format(time.RFC3339),
		Lang:        src.Language,
	}, nil
}
