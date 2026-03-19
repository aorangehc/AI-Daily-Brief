package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	date    = flag.String("date", "", "Date in YYYY-MM-DD format")
	dryRun  = flag.Bool("dry-run", false, "Dry run mode")
	force   = flag.Bool("force", false, "Force execution even if already done")
	verbose = flag.Bool("verbose", false, "Verbose output")
)

func main() {
	flag.Parse()

	if *date == "" {
		fmt.Fprintln(os.Stderr, "Error: --date is required")
		flag.Usage()
		os.Exit(1)
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if *verbose {
		fmt.Printf("Rendering site data for %s\n", *date)
	}

	// Ensure output directories exist
	if err := os.MkdirAll("site/public/data/digests", 0755); err != nil {
		return fmt.Errorf("failed to create digests dir: %w", err)
	}
	if err := os.MkdirAll("site/public/data/topics", 0755); err != nil {
		return fmt.Errorf("failed to create topics dir: %w", err)
	}
	if err := os.MkdirAll("site/public/data/items", 0755); err != nil {
		return fmt.Errorf("failed to create items dir: %w", err)
	}
	if err := os.MkdirAll("site/public/data/indexes", 0755); err != nil {
		return fmt.Errorf("failed to create indexes dir: %w", err)
	}

	// Copy digest and topics to site/public/data
	if err := copyFile(filepath.Join("data/digests", fmt.Sprintf("%s.json", *date)),
		filepath.Join("site/public/data/digests", fmt.Sprintf("%s.json", *date))); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to copy digest: %w", err)
		}
	}

	if err := copyFile(filepath.Join("data/topics", fmt.Sprintf("%s.json", *date)),
		filepath.Join("site/public/data/topics", fmt.Sprintf("%s.json", *date))); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to copy topics: %w", err)
		}
	}

	// Copy items if exists
	if err := copyFile(filepath.Join("data/items", fmt.Sprintf("%s.ndjson", *date)),
		filepath.Join("site/public/data/items", fmt.Sprintf("%s.ndjson", *date))); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to copy items: %w", err)
		}
	}

	// Load today's digest for generating indexes
	digest, err := loadDigest(*date)
	if err != nil {
		return fmt.Errorf("failed to load digest: %w", err)
	}

	// Generate search index
	if err := generateSearchIndex(*date); err != nil {
		return fmt.Errorf("failed to generate search index: %w", err)
	}

	// Generate archive index
	if err := generateArchiveIndex(*date, digest); err != nil {
		return fmt.Errorf("failed to generate archive index: %w", err)
	}

	// Generate latest index
	if err := generateLatestIndex(*date, digest); err != nil {
		return fmt.Errorf("failed to generate latest index: %w", err)
	}

	// Generate sources index
	if err := generateSourcesIndex(*date); err != nil {
		return fmt.Errorf("failed to generate sources index: %w", err)
	}

	if *verbose {
		fmt.Printf("Renderer: generated site data for %s\n", *date)
	}

	fmt.Printf("Renderer: rendered site data for %s\n", *date)
	return nil
}

func loadDigest(date string) (*Digest, error) {
	path := filepath.Join("data/digests", fmt.Sprintf("%s.json", date))
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var digest Digest
	if err := json.NewDecoder(file).Decode(&digest); err != nil {
		return nil, err
	}
	return &digest, nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 32*1024)
	for {
		n, rErr := srcFile.Read(buf)
		if n > 0 {
			if _, wErr := dstFile.Write(buf[:n]); wErr != nil {
				return wErr
			}
		}
		if rErr != nil {
			break
		}
	}
	return nil
}

// Digest represents the daily digest structure
type Digest struct {
	Date        string       `json:"date"`
	Edition     string       `json:"edition"`
	Headline    string       `json:"headline"`
	Lead        string       `json:"lead"`
	TopTopicIDs []string     `json:"top_topic_ids"`
	TopItemIDs  []string     `json:"top_item_ids"`
	Stats       DigestStats  `json:"stats"`
}

// DigestStats contains statistics for the digest
type DigestStats struct {
	RawItems        int `json:"raw_items"`
	NormalizedItems int `json:"normalized_items"`
	PublishedItems  int `json:"published_items"`
	Topics          int `json:"topics"`
	Sources         int `json:"sources"`
}

// Item represents a normalized item
type Item struct {
	ID           string   `json:"id"`
	SourceID     string   `json:"source_id"`
	CanonicalURL string   `json:"canonical_url"`
	Domain       string   `json:"domain"`
	Title        string   `json:"title"`
	Summary1Line string   `json:"summary_1line"`
	ContentText  string   `json:"content_text"`
	PublishedAt  string   `json:"published_at"`
	Lang         string   `json:"lang"`
	Tags         []string `json:"tags"`
	FinalScore   float64  `json:"final_score"`
	Status       string   `json:"status"`
}

// TopicCluster represents a topic cluster
type TopicCluster struct {
	TopicID         string   `json:"topic_id"`
	Name            string   `json:"name"`
	Summary         string   `json:"summary"`
	WhyItMatters    string   `json:"why_it_matters"`
	Keywords        []string `json:"keywords"`
	ImportanceScore float64  `json:"importance_score"`
	ItemIDs         []string `json:"item_ids"`
}

func generateSearchIndex(date string) error {
	items, err := loadItems(date)
	if err != nil {
		return err
	}

	topics, err := loadTopics(date)
	if err != nil {
		return err
	}

	// Build topic ID to keywords map
	topicKeywords := make(map[string][]string)
	topicNames := make(map[string]string)
	for _, t := range topics {
		topicKeywords[t.TopicID] = t.Keywords
		topicNames[t.TopicID] = t.Name
	}

	// Build item ID to topic IDs map
	itemTopics := make(map[string][]string)
	for _, t := range topics {
		for _, itemID := range t.ItemIDs {
			itemTopics[itemID] = append(itemTopics[itemID], t.TopicID)
		}
	}

	// Build search index
	searchItems := make([]SearchIndexEntry, 0, len(items))
	for _, item := range items {
		if item.Status != "ready" {
			continue
		}
		entry := SearchIndexEntry{
			ID:      item.ID,
			Title:   item.Title,
			Summary: item.Summary1Line,
			URL:     item.CanonicalURL,
			Date:    date,
			Domain:  item.Domain,
			Tags:    item.Tags,
		}
		if topics, ok := itemTopics[item.ID]; ok {
			entry.Topics = topics
			// Also add topic keywords as tags for better search
			for _, tid := range topics {
				entry.Tags = append(entry.Tags, topicKeywords[tid]...)
				if name := topicNames[tid]; name != "" {
					entry.Tags = append(entry.Tags, name)
				}
			}
		}
		// Deduplicate tags
		entry.Tags = deduplicateString(entry.Tags)
		searchItems = append(searchItems, entry)
	}

	return writeJSON(filepath.Join("site/public/data/indexes", "search-index.json"), searchItems)
}

func loadItems(date string) ([]Item, error) {
	path := filepath.Join("data/items", fmt.Sprintf("%s.ndjson", date))
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items []Item
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var item Item
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			continue
		}
		items = append(items, item)
	}
	return items, scanner.Err()
}

func loadTopics(date string) ([]TopicCluster, error) {
	path := filepath.Join("data/topics", fmt.Sprintf("%s.json", date))
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Try to parse as array first
	var topics []TopicCluster
	if err := json.NewDecoder(file).Decode(&topics); err != nil {
		// Try parsing as wrapped object
		file.Seek(0, 0)
		var wrapped struct {
			Clusters []TopicCluster `json:"clusters"`
		}
		if err := json.NewDecoder(file).Decode(&wrapped); err != nil {
			return nil, err
		}
		topics = wrapped.Clusters
	}
	return topics, nil
}

func generateArchiveIndex(date string, digest *Digest) error {
	archivePath := filepath.Join("site/public/data/indexes", "archive.json")

	// Load existing archive
	var archives []ArchiveEntry
	if data, err := os.ReadFile(archivePath); err == nil {
		json.Unmarshal(data, &archives)
	}

	// Add/update today's entry
	newEntry := ArchiveEntry{
		Date:     date,
		Headline: digest.Headline,
	}

	// Find and update if exists
	found := false
	for i, entry := range archives {
		if entry.Date == date {
			archives[i] = newEntry
			found = true
			break
		}
	}
	if !found {
		archives = append(archives, newEntry)
	}

	// Sort by date descending
	sort.Slice(archives, func(i, j int) bool {
		return archives[i].Date > archives[j].Date
	})

	// Keep last 90 days
	if len(archives) > 90 {
		archives = archives[:90]
	}

	return writeJSON(archivePath, archives)
}

func generateLatestIndex(date string, digest *Digest) error {
	latest := LatestEntry{
		Date:     date,
		Headline: digest.Headline,
	}
	return writeJSON(filepath.Join("site/public/data/indexes", "latest.json"), latest)
}

func generateSourcesIndex(date string) error {
	items, err := loadItems(date)
	if err != nil {
		return err
	}

	// Count items per source
	sourceCounts := make(map[string]int)
	for _, item := range items {
		if item.Status == "ready" {
			sourceCounts[item.SourceID]++
		}
	}

	// Convert to array
	var sources []SourceEntry
	for sourceID, count := range sourceCounts {
		sources = append(sources, SourceEntry{
			SourceID: sourceID,
			Count:    count,
		})
	}

	// Sort by count descending
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Count > sources[j].Count
	})

	return writeJSON(filepath.Join("site/public/data/indexes", "sources.json"), sources)
}

func writeJSON(path string, v interface{}) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// SearchIndexEntry represents an entry in the search index
type SearchIndexEntry struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	URL     string   `json:"url"`
	Date    string   `json:"date"`
	Domain  string   `json:"domain"`
	Tags    []string `json:"tags,omitempty"`
	Topics  []string `json:"topics,omitempty"`
}

// ArchiveEntry represents an entry in the archive index
type ArchiveEntry struct {
	Date     string `json:"date"`
	Headline string `json:"headline"`
}

// LatestEntry represents the latest digest info
type LatestEntry struct {
	Date     string `json:"date"`
	Headline string `json:"headline"`
}

// SourceEntry represents a source with its item count
type SourceEntry struct {
	SourceID string `json:"source_id"`
	Count    int    `json:"count"`
}

func deduplicateString(ss []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(ss))
	for _, s := range ss {
		s = strings.ToLower(strings.TrimSpace(s))
		if s != "" && !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}