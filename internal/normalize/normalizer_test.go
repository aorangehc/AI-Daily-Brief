package normalize

import (
	"testing"

	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://Example.COM/blog", "https://example.com/blog"},
		{"https://openai.com/blog#section", "https://openai.com/blog"},
		{"https://openai.com/blog?utm_source=twitter", "https://openai.com/blog"},
		{"https://openai.com/blog?fbclid=abc", "https://openai.com/blog"},
		{"", ""},
	}

	for _, tt := range tests {
		result := normalizeURL(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeURL(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://openai.com/blog", "openai.com"},
		{"https://blog.google/technology/ai", "blog.google"},
		{"", ""},
	}

	for _, tt := range tests {
		result := extractDomain(tt.input)
		if result != tt.expected {
			t.Errorf("extractDomain(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalize(t *testing.T) {
	n := New()

	raw := &schema.RawItem{
		ID:          "openai_blog_2026-03-19_001",
		SourceID:    "openai_blog",
		Title:       "  OpenAI announces GPT-5  ",
		URL:         "https://openai.com/blog/gpt-5",
		PublishedAt: "Wed, 19 Mar 2026 08:00:00 GMT",
		ContentRaw:  "OpenAI has announced GPT-5.",
		Lang:        "en",
	}

	item := n.Normalize(raw)

	if item.SourceID != "openai_blog" {
		t.Errorf("SourceID = %q, want %q", item.SourceID, "openai_blog")
	}
	if item.Title != "OpenAI announces GPT-5" {
		t.Errorf("Title = %q, want %q", item.Title, "OpenAI announces GPT-5")
	}
	if item.Domain != "openai.com" {
		t.Errorf("Domain = %q, want %q", item.Domain, "openai.com")
	}
	if item.HashURL == "" {
		t.Error("HashURL should not be empty")
	}
	if item.Status != "ready" {
		t.Errorf("Status = %q, want %q", item.Status, "ready")
	}
}

func TestHashFunctions(t *testing.T) {
	url := "https://openai.com/blog"
	url2 := "https://openai.com/blog"
	title := "OpenAI announces GPT-5"

	h1 := hashURL(url)
	h2 := hashURL(url2)

	if h1 != h2 {
		t.Errorf("hashURL should produce same hash for same URL")
	}

	th := hashTitle(title)
	if th == "" {
		t.Error("hashTitle should not be empty")
	}

	// Same title different case
	th2 := hashTitle("openai announces gpt-5")
	if th != th2 {
		t.Errorf("hashTitle should be case insensitive, got %q and %q", th, th2)
	}
}
