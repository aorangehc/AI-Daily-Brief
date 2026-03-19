package fetch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
	"github.com/ai-daily-brief/ai-daily-brief/internal/source"
	"github.com/mmcdole/gofeed"
)

// RSSFetcher fetches items from RSS/Atom feeds
type RSSFetcher struct{}

// Name returns the fetcher type name
func (f *RSSFetcher) Name() string {
	return "rss"
}

// Fetch retrieves items from an RSS/Atom feed
func (f *RSSFetcher) Fetch(ctx context.Context, src *source.Source) ([]*schema.RawItem, error) {
	if src.FeedURL == "" {
		return nil, fmt.Errorf("no feed_url specified for source %s", src.ID)
	}

	parser := gofeed.NewParser()
	feed, err := parser.ParseURLWithContext(src.FeedURL, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed %s: %w", src.FeedURL, err)
	}

	var items []*schema.RawItem
	seq := 0

	for _, entry := range feed.Items {
		seq++
		item := &schema.RawItem{
			ID:          fmt.Sprintf("%s_%s_%03d", src.ID, time.Now().Format("2006-01-02"), seq),
			SourceID:    src.ID,
			SourceType:  "rss",
			Title:       cleanText(entry.Title),
			URL:         normalizeURL(entry.Link),
			Author:      cleanText(entry.Author.Name),
			PublishedAt: normalizeTime(entry.Published),
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
			Lang:        src.Language,
		}

		// Build content_raw from description or content
		if entry.Content != "" {
			item.ContentRaw = entry.Content
		} else if entry.Description != "" {
			item.ContentRaw = entry.Description
		}

		// Filter by allow_paths/deny_paths if configured
		if !shouldIncludeURL(item.URL, src.AllowPaths, src.DenyPaths) {
			continue
		}

		items = append(items, item)
	}

	return items, nil
}

// cleanText removes leading/trailing whitespace
func cleanText(s string) string {
	return strings.TrimSpace(s)
}

// normalizeURL ensures URL is not empty and has a scheme
func normalizeURL(link string) string {
	if link == "" {
		return ""
	}
	link = strings.TrimSpace(link)
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		link = "https://" + link
	}
	return link
}

// normalizeTime converts various time formats to RFC3339
func normalizeTime(t string) string {
	if t == "" {
		return time.Now().UTC().Format(time.RFC3339)
	}
	parsed, err := time.Parse(time.RFC1123, t)
	if err != nil {
		parsed, err = time.Parse(time.RFC1123Z, t)
	}
	if err != nil {
		parsed, err = time.Parse("Mon, 02 Jan 2006 15:04:05 MST", t)
	}
	if err != nil {
		layouts := []string{
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04:05",
			"02 Jan 2006 15:04:05 MST",
		}
		for _, layout := range layouts {
			parsed, err = time.Parse(layout, t)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return time.Now().UTC().Format(time.RFC3339)
	}
	return parsed.UTC().Format(time.RFC3339)
}

// shouldIncludeURL checks if URL matches allow_paths and not deny_paths
func shouldIncludeURL(url string, allowPaths, denyPaths []string) bool {
	if len(denyPaths) > 0 {
		for _, deny := range denyPaths {
			if strings.Contains(url, deny) {
				return false
			}
		}
	}

	if len(allowPaths) > 0 {
		for _, allow := range allowPaths {
			if strings.Contains(url, allow) {
				return true
			}
		}
		return false
	}

	return true
}
