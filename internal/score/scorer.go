package score

import (
	"strings"
	"time"

	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
	"gopkg.in/yaml.v3"
	"os"
)

// Scorer calculates scores for items
type Scorer struct {
	hourInDay     int
	sourceWeights map[string]float64
}

// WeightsConfig represents the source weights configuration
type WeightsConfig struct {
	SourceWeights map[string]float64 `yaml:"source_weights"`
}

// New creates a new Scorer with the current hour
func New() *Scorer {
	return &Scorer{
		hourInDay:     time.Now().UTC().Hour(),
		sourceWeights: defaultWeights(),
	}
}

// LoadWeights loads source weights from a YAML file
func (s *Scorer) LoadWeights(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg WeightsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	if cfg.SourceWeights != nil {
		s.sourceWeights = cfg.SourceWeights
	}

	return nil
}

// defaultWeights returns the default category weights
func defaultWeights() map[string]float64 {
	return map[string]float64{
		"official":  1.0,
		"research":  0.85,
		"product":   0.75,
		"community": 0.7,
		"code":      0.65,
		"forum":     0.6,
	}
}

// Score calculates all scores for an item
func (s *Scorer) Score(item *schema.Item, category string) {
	// source_weight: directly from source config
	item.SourceWeight = s.getSourceWeight(category)

	// freshness_score: based on publication time
	item.FreshnessScore = s.calcFreshness(item.PublishedAt)

	// heat_score: based on content length and engagement signals
	item.HeatScore = s.calcHeat(item)

	// originality_score: content length and独特性
	item.OriginalityScore = s.calcOriginality(item)

	// final_score: weighted combination
	item.FinalScore = item.SourceWeight*0.30 +
		item.FreshnessScore*0.25 +
		item.HeatScore*0.20 +
		item.OriginalityScore*0.15
}

// getSourceWeight returns weight for a category
func (s *Scorer) getSourceWeight(category string) float64 {
	if w, ok := s.sourceWeights[category]; ok {
		return w
	}
	return 0.5 // default
}

// calcFreshness calculates freshness score (0-1, higher = fresher)
func (s *Scorer) calcFreshness(publishedAt string) float64 {
	if publishedAt == "" {
		return 0.3 // unknown publication time
	}

	published, err := time.Parse(time.RFC3339, publishedAt)
	if err != nil {
		return 0.3
	}

	hoursOld := time.Since(published).Hours()
	if hoursOld < 0 {
		hoursOld = 0 // future dates treated as now
	}

	switch {
	case hoursOld <= 6:
		return 1.0
	case hoursOld <= 12:
		return 0.8
	case hoursOld <= 24:
		return 0.6
	case hoursOld <= 48:
		return 0.4
	default:
		return 0.2
	}
}

// calcHeat calculates heat score based on content and domain signals
func (s *Scorer) calcHeat(item *schema.Item) float64 {
	score := 0.5

	// Content length bonus (longer content often means more substance)
	if item.ContentText != "" {
		wordCount := len(strings.Fields(item.ContentText))
		switch {
		case wordCount > 500:
			score += 0.15
		case wordCount > 200:
			score += 0.1
		case wordCount > 100:
			score += 0.05
		}
	}

	// Domain authority signals (domains that typically have high-quality content)
	highAuthorityDomains := map[string]bool{
		"arxiv.org":      true,
		"github.com":     true,
		"openai.com":     true,
		"anthropic.com":  true,
		"deepmind.com":   true,
		"microsoft.com":  true,
		"google.com":     true,
		"nature.com":     true,
		"science.org":    true,
	}
	if highAuthorityDomains[item.Domain] {
		score += 0.15
	}

	// Morning bonus (items published early in the day may be more relevant)
	// Items published between 6am-10am UTC get a slight boost
	if item.PublishedAt != "" {
		published, err := time.Parse(time.RFC3339, item.PublishedAt)
		if err == nil {
			hour := published.Hour()
			if hour >= 6 && hour <= 10 {
				score += 0.1
			}
		}
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calcOriginality calculates originality score based on content characteristics
func (s *Scorer) calcOriginality(item *schema.Item) float64 {
	score := 0.5

	// Content length (very short content is often not original)
	if item.ContentText != "" {
		wordCount := len(strings.Fields(item.ContentText))
		switch {
		case wordCount >= 50 && wordCount <= 300:
			score += 0.2 // sweet spot
		case wordCount > 300 && wordCount <= 1000:
			score += 0.15
		case wordCount > 1000:
			score += 0.1
		case wordCount < 20:
			score -= 0.2 // too short
		}
	}

	// Check for repetitive content (low uniqueness)
	if item.ContentText != "" {
		uniqueRatio := uniqueWordRatio(item.ContentText)
		score += uniqueRatio * 0.3
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// uniqueWordRatio calculates the ratio of unique words to total words
func uniqueWordRatio(content string) float64 {
	words := strings.Fields(strings.ToLower(content))
	if len(words) == 0 {
		return 0
	}

	unique := make(map[string]bool)
	for _, w := range words {
		if len(w) > 2 { // skip short words
			unique[w] = true
		}
	}

	return float64(len(unique)) / float64(len(words))
}
