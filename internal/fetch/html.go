package fetch

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
	"github.com/ai-daily-brief/ai-daily-brief/internal/source"
	"github.com/PuerkitoBio/goquery"
)

// HTMLFetcher fetches and extracts content from HTML pages
type HTMLFetcher struct {
	client *http.Client
}

// Name returns the fetcher type name
func (f *HTMLFetcher) Name() string {
	return "html"
}

// NewHTMLFetcher creates a new HTML fetcher
func NewHTMLFetcher() *HTMLFetcher {
	return &HTMLFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Fetch retrieves items by crawling HTML pages
func (f *HTMLFetcher) Fetch(ctx context.Context, src *source.Source) ([]*schema.RawItem, error) {
	if len(src.AllowPaths) == 0 {
		return nil, fmt.Errorf("no allow_paths specified for HTML source %s", src.ID)
	}

	var items []*schema.RawItem
	seq := 0

	for _, path := range src.AllowPaths {
		pageURL := strings.TrimSuffix(src.BaseURL, "/") + "/" + strings.TrimPrefix(path, "/")

		pageItems, err := f.fetchPage(ctx, pageURL, src)
		if err != nil {
			continue
		}

		for _, item := range pageItems {
			seq++
			item.ID = fmt.Sprintf("%s_%s_%03d", src.ID, time.Now().Format("2006-01-02"), seq)
			items = append(items, item)
		}
	}

	return items, nil
}

// fetchPage fetches a single page and extracts article links
func (f *HTMLFetcher) fetchPage(ctx context.Context, pageURL string, src *source.Source) ([]*schema.RawItem, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "AI-Daily-Brief/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTML fetch returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var items []*schema.RawItem

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		if shouldExcludeLink(href) {
			return
		}

		absoluteURL := makeAbsoluteURL(src.BaseURL, href)
		if absoluteURL == "" {
			return
		}

		title := strings.TrimSpace(s.Text())

		item := &schema.RawItem{
			SourceID:    src.ID,
			SourceType:  "html",
			Title:       title,
			URL:         absoluteURL,
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
			Lang:        src.Language,
		}

		items = append(items, item)
	})

	return items, nil
}

// makeAbsoluteURL converts a potentially relative URL to absolute
func makeAbsoluteURL(baseURL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "mailto:") {
		return ""
	}

	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	base := strings.TrimSuffix(baseURL, "/")
	if strings.HasPrefix(href, "/") {
		return base + href
	}
	return base + "/" + href
}

// shouldExcludeLink returns true for navigation/utility links
func shouldExcludeLink(href string) bool {
	excludedPrefixes := []string{
		"/tag/", "/category/", "/author/",
		"/page/", "/feed/", "/search/",
		"facebook.com", "twitter.com", "github.com",
		"linkedin.com", "instagram.com",
	}
	href = strings.ToLower(href)
	for _, prefix := range excludedPrefixes {
		if strings.Contains(href, prefix) {
			return true
		}
	}
	return false
}
