package score

import (
	"testing"
	"time"

	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
)

func TestCalcFreshness(t *testing.T) {
	s := New()

	tests := []struct {
		publishedAt string
		minExpected float64
	}{
		{time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339), 1.0},  // 2 hours ago
		{time.Now().UTC().Add(-8 * time.Hour).Format(time.RFC3339), 0.8},  // 8 hours ago
		{time.Now().UTC().Add(-18 * time.Hour).Format(time.RFC3339), 0.6}, // 18 hours ago
		{time.Now().UTC().Add(-36 * time.Hour).Format(time.RFC3339), 0.4}, // 36 hours ago
		{time.Now().UTC().Add(-72 * time.Hour).Format(time.RFC3339), 0.2}, // 72 hours ago
		{"", 0.3}, // unknown
	}

	for _, tt := range tests {
		score := s.calcFreshness(tt.publishedAt)
		if score < tt.minExpected {
			t.Errorf("calcFreshness(%q) = %f, want >= %f", tt.publishedAt, score, tt.minExpected)
		}
	}
}

func TestCalcHeat(t *testing.T) {
	s := New()

	item := &schema.Item{
		Domain:      "arxiv.org",
		ContentText: "This is a long content with many words to test the heat calculation properly.",
		PublishedAt: time.Now().UTC().Add(-8 * time.Hour).Format(time.RFC3339),
	}

	score := s.calcHeat(item)
	if score < 0.6 {
		t.Errorf("calcHeat for arxiv.org = %f, want >= 0.6", score)
	}

	// Test with unknown domain
	item.Domain = "example.com"
	score = s.calcHeat(item)
	if score > 0.7 {
		t.Errorf("calcHeat for example.com = %f, want <= 0.7", score)
	}
}

func TestCalcOriginality(t *testing.T) {
	s := New()

	tests := []struct {
		contentLen int
		minScore   float64
	}{
		{100, 0.6},  // normal content
		{500, 0.6},  // long content
		{5, 0.3},    // very short
	}

	for _, tt := range tests {
		words := make([]string, tt.contentLen)
		for i := range words {
			words[i] = "word"
		}
		item := &schema.Item{ContentText: joinWords(words)}
		score := s.calcOriginality(item)
		if score < tt.minScore {
			t.Errorf("calcOriginality with %d words = %f, want >= %f", tt.contentLen, score, tt.minScore)
		}
	}
}

func joinWords(words []string) string {
	result := ""
	for i, w := range words {
		if i > 0 {
			result += " "
		}
		result += w
	}
	return result
}

func TestScore(t *testing.T) {
	s := New()

	item := &schema.Item{
		Domain:      "openai.com",
		ContentText: "OpenAI announces GPT-5 with improved reasoning capabilities.",
		PublishedAt: time.Now().UTC().Add(-4 * time.Hour).Format(time.RFC3339),
	}

	s.Score(item, "official")

	if item.SourceWeight != 1.0 {
		t.Errorf("SourceWeight = %f, want 1.0", item.SourceWeight)
	}
	if item.FreshnessScore != 1.0 {
		t.Errorf("FreshnessScore = %f, want 1.0", item.FreshnessScore)
	}
	if item.FinalScore == 0 {
		t.Error("FinalScore should not be 0")
	}

	// Verify final score formula
	expectedFinal := item.SourceWeight*0.30 + item.FreshnessScore*0.25 + item.HeatScore*0.20 + item.OriginalityScore*0.15
	if item.FinalScore != expectedFinal {
		t.Errorf("FinalScore = %f, want %f", item.FinalScore, expectedFinal)
	}
}
