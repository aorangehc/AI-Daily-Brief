package dedupe

import (
	"testing"

	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
)

func TestSimilarity(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		minScore float64
	}{
		{"OpenAI announces GPT-5", "OpenAI announces GPT-5", 1.0},
		{"OpenAI announces GPT-5", "OpenAI releases GPT-5", 0.4}, // Jaccard on short titles is lower
		{"Machine Learning News Today", "Machine Learning Updates Today", 0.6},
		{"", "", 0},
	}

	for _, tt := range tests {
		score := similarity(tt.a, tt.b)
		if score < tt.minScore {
			t.Errorf("similarity(%q, %q) = %f, want >= %f", tt.a, tt.b, score, tt.minScore)
		}
	}
}

func TestDeduper(t *testing.T) {
	d := New()

	items := []*schema.Item{
		{ID: "1", HashURL: "url1", HashTitle: "title1", HashContent: "content1", Title: "OpenAI announces GPT-5"},
		{ID: "2", HashURL: "url2", HashTitle: "title2", HashContent: "content2", Title: "Claude 4 Released"},
		{ID: "3", HashURL: "url1", HashTitle: "title3", HashContent: "content3", Title: "Another GPT-5 article"}, // duplicate URL
	}

	result := d.Dedup(items)

	if len(result) != 2 {
		t.Errorf("Expected 2 items after dedup, got %d", len(result))
	}

	// Verify the URL duplicate was removed
	urlCount := 0
	for _, item := range result {
		if item.HashURL == "url1" {
			urlCount++
		}
	}
	if urlCount != 1 {
		t.Errorf("Expected 1 item with url1, got %d", urlCount)
	}
}

func TestDeduperTitleSimilarity(t *testing.T) {
	d := New()

	// Use nearly identical titles - Jaccard similarity > 0.8
	items := []*schema.Item{
		{ID: "1", HashURL: "url1", HashTitle: "hash1", HashContent: "content1", Title: "OpenAI announces GPT-5 with improved reasoning capabilities"},
		{ID: "2", HashURL: "url2", HashTitle: "hash2", HashContent: "content2", Title: "OpenAI announces GPT-5 with improved reasoning capabilities today"},
	}

	result := d.Dedup(items)

	// Titles have high similarity (>0.8), should deduplicate
	if len(result) != 1 {
		t.Errorf("Expected 1 item after title similarity dedup, got %d", len(result))
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		minLen   int
	}{
		{"OpenAI announces GPT-5", 3},
		{"The and a are stopwords", 2},
	}

	for _, tt := range tests {
		tokens := tokenize(tt.input)
		for _, tok := range tokens {
			if len(tok) < tt.minLen {
				t.Errorf("token %q too short, want >= %d", tok, tt.minLen)
			}
		}
	}
}
