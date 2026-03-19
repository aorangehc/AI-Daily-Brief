package dedupe

import (
	"sort"
	"strings"

	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
)

// Deduper removes duplicate items using three-layer deduplication
type Deduper struct {
	seenURLs    map[string]bool
	seenTitles  map[string]string // hash -> original title
	seenContent map[string]bool

	titleThreshold float64
}

// New creates a new Deduper
func New() *Deduper {
	return &Deduper{
		seenURLs:       make(map[string]bool),
		seenTitles:    make(map[string]string),
		seenContent:   make(map[string]bool),
		titleThreshold: 0.8,
	}
}

// Dedup removes duplicate items
func (d *Deduper) Dedup(items []*schema.Item) []*schema.Item {
	var result []*schema.Item

	for _, item := range items {
		if d.isDuplicate(item) {
			continue
		}

		// Mark as seen
		d.seenURLs[item.HashURL] = true
		d.seenTitles[item.HashTitle] = item.Title
		d.seenContent[item.HashContent] = true

		result = append(result, item)
	}

	return result
}

// isDuplicate checks if an item is a duplicate using three layers
func (d *Deduper) isDuplicate(item *schema.Item) bool {
	// Layer 1: URL exact match (fastest, most reliable)
	if d.seenURLs[item.HashURL] {
		return true
	}

	// Layer 2: Title similarity
	for _, seenTitle := range d.seenTitles {
		if similarity(item.Title, seenTitle) > d.titleThreshold {
			return true
		}
	}

	// Layer 3: Content hash
	if item.HashContent != "" && d.seenContent[item.HashContent] {
		return true
	}

	return false
}

// similarity calculates Jaccard similarity between two strings
func similarity(a, b string) float64 {
	if a == "" || b == "" {
		return 0
	}

	tokensA := tokenize(strings.ToLower(a))
	tokensB := tokenize(strings.ToLower(b))

	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0
	}

	// Jaccard index: |A ∩ B| / |A ∪ B|
	setA := make(map[string]bool)
	setB := make(map[string]bool)

	for _, t := range tokensA {
		setA[t] = true
	}
	for _, t := range tokensB {
		setB[t] = true
	}

	intersection := 0
	union := len(setA)

	for t := range setA {
		if setB[t] {
			intersection++
		}
	}

	for t := range setB {
		if !setA[t] {
			union++
		}
	}

	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// tokenize splits a string into word tokens
func tokenize(s string) []string {
	// Split on whitespace and punctuation
	var tokens []string
	var current strings.Builder

	for _, r := range s {
		if isWordChar(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	// Remove very short tokens and common stop words
	filtered := make([]string, 0, len(tokens))
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "can": true,
		"this": true, "that": true, "these": true, "those": true,
	}

	for _, t := range tokens {
		if len(t) > 2 && !stopWords[t] {
			filtered = append(filtered, t)
		}
	}

	return filtered
}

// isWordChar returns true if rune is a word character (letter, digit)
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// Reset clears all seen items (useful for new date)
func (d *Deduper) Reset() {
	d.seenURLs = make(map[string]bool)
	d.seenTitles = make(map[string]string)
	d.seenContent = make(map[string]bool)
}

// LoadState loads previously seen hashes from deduped items
func (d *Deduper) LoadState(items []*schema.Item) {
	for _, item := range items {
		d.seenURLs[item.HashURL] = true
		d.seenTitles[item.HashTitle] = item.Title
		d.seenContent[item.HashContent] = true
	}
}

// SortByScore sorts items by final score descending
func SortByScore(items []*schema.Item) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].FinalScore > items[j].FinalScore
	})
}
