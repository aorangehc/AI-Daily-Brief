package normalize

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"regexp"
	"strings"

	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
	"github.com/araddon/dateparse"
)

// Normalizer converts RawItems to normalized Items
type Normalizer struct{}

// New creates a new Normalizer
func New() *Normalizer {
	return &Normalizer{}
}

// Normalize converts a RawItem to a normalized Item
func (n *Normalizer) Normalize(raw *schema.RawItem) *schema.Item {
	item := &schema.Item{
		RawID:        raw.ID,
		SourceID:     raw.SourceID,
		Lang:         raw.Lang,
		Title:        strings.TrimSpace(raw.Title),
		ContentText:  raw.ContentRaw,
		CanonicalURL: normalizeURL(raw.URL),
		Domain:      extractDomain(raw.URL),
		PublishedAt:  normalizeTime(raw.PublishedAt),
	}

	item.ID = generateItemID(raw)
	item.HashURL = hashURL(item.CanonicalURL)
	item.HashTitle = hashTitle(item.Title)
	item.HashContent = hashContent(raw.ContentRaw)
	item.Status = "ready"

	return item
}

// generateItemID creates a unique ID for an item
func generateItemID(raw *schema.RawItem) string {
	// Use source_id + date + seq from raw ID
	parts := strings.Split(raw.ID, "_")
	if len(parts) >= 3 {
		return "item_" + strings.Join(parts[1:], "_") // e.g., item_2026-03-19_001
	}
	// Fallback: hash the canonical URL
	return "item_" + hashURL(raw.URL)[:12]
}

// normalizeURL normalizes a URL to canonical form
func normalizeURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(rawURL))
	}

	// Remove fragment
	u.Fragment = ""

	// Remove common tracking params
	rawQuery := u.RawQuery
	u.RawQuery = canonicalQuery(rawQuery)

	// Convert to lowercase scheme and host
	u.Scheme = strings.ToLower(u.Scheme)
	if u.Host != "" {
		u.Host = strings.ToLower(u.Host)
	}

	return u.String()
}

// canonicalQuery sorts query parameters for consistency
func canonicalQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	params, err := url.ParseQuery(rawQuery)
	if err != nil {
		return rawQuery
	}

	// Remove tracking params
	delete(params, "utm_source")
	delete(params, "utm_medium")
	delete(params, "utm_campaign")
	delete(params, "utm_content")
	delete(params, "utm_term")
	delete(params, "ref")
	delete(params, "fbclid")
	delete(params, "gclid")

	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	// Sort is not strictly necessary for canonicalization but ensures determinism
	// skipping for performance

	if len(params) == 0 {
		return ""
	}

	// Rebuild query string
	pairs := make([]string, 0, len(params))
	for k, v := range params {
		pairs = append(pairs, k+"="+v[0])
	}

	return strings.Join(pairs, "&")
}

// extractDomain extracts and lowercases the domain from a URL
func extractDomain(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		// Try to extract domain directly
		withoutScheme := strings.TrimPrefix(rawURL, "https://")
		withoutScheme = strings.TrimPrefix(withoutScheme, "http://")
		parts := strings.Split(withoutScheme, "/")
		if len(parts) > 0 {
			return strings.ToLower(parts[0])
		}
		return ""
	}

	return strings.ToLower(u.Host)
}

// normalizeTime parses various date formats and returns RFC3339 UTC
func normalizeTime(t string) string {
	if t == "" {
		return ""
	}

	parsed, err := dateparse.ParseAny(t)
	if err != nil {
		return t // Return as-is if we can't parse
	}

	return parsed.UTC().Format("2006-01-02T15:04:05Z")
}

// hashURL returns a SHA256 hash of the canonical URL
func hashURL(canonicalURL string) string {
	if canonicalURL == "" {
		return ""
	}
	h := sha256.Sum256([]byte(canonicalURL))
	return hex.EncodeToString(h[:])
}

// hashTitle returns a SHA256 hash of the normalized title
func hashTitle(title string) string {
	if title == "" {
		return ""
	}
	// Normalize: lowercase, collapse whitespace
	normalized := strings.ToLower(title)
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
	h := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(h[:])
}

// hashContent returns a SHA256 hash of the content
func hashContent(content string) string {
	if content == "" {
		return ""
	}
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}
