package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
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
		fmt.Printf("Generating digest for %s\n", *date)
	}

	// Read scored items
	items, err := loadItems(*date)
	if err != nil {
		return fmt.Errorf("failed to load items: %w", err)
	}

	if *verbose {
		fmt.Printf("Loaded %d items\n", len(items))
	}

	// Filter to only "ready" items
	var readyItems []*schema.Item
	seenSources := make(map[string]bool)
	for _, item := range items {
		if item.Status == "ready" {
			readyItems = append(readyItems, item)
			seenSources[item.SourceID] = true
		}
	}

	// Sort by final score descending
	sort.Slice(readyItems, func(i, j int) bool {
		return readyItems[i].FinalScore > readyItems[j].FinalScore
	})

	// Take top items for digest
	topItems := readyItems
	if len(topItems) > 20 {
		topItems = topItems[:20]
	}

	// Collect top item IDs
	topItemIDs := make([]string, 0, len(topItems))
	for _, item := range topItems {
		topItemIDs = append(topItemIDs, item.ID)
	}

	// Build digest
	digest := &schema.DailyDigest{
		Date:        *date,
		Edition:     "nightly",
		Headline:    fmt.Sprintf("AI Daily Brief - %s", formatDisplayDate(*date)),
		Lead:        fmt.Sprintf("Digest generation pending Openclaw integration. %d items processed, %d ready for publication.", len(items), len(readyItems)),
		TopTopicIDs: []string{}, // Topics require Openclaw clustering
		TopItemIDs:  topItemIDs,
		Stats: schema.DigestStats{
			RawItems:        countRawItems(*date),
			NormalizedItems: len(items),
			PublishedItems:  len(readyItems),
			Topics:          0, // Topics require Openclaw clustering
			Sources:         len(seenSources),
		},
	}

	if *verbose {
		fmt.Printf("Digest: %d items, %d sources, top score: %.2f\n",
			len(readyItems), len(seenSources), topItems[0].FinalScore)
	}

	if *dryRun {
		fmt.Printf("Dry run: would write digest to data/digests/%s.json\n", *date)
		return nil
	}

	// Write digest
	if err := writeDigest(*date, digest); err != nil {
		return fmt.Errorf("failed to write digest: %w", err)
	}

	// Write empty topics for compatibility (topics require Openclaw)
	if err := writeEmptyTopics(*date); err != nil {
		return fmt.Errorf("failed to write topics: %w", err)
	}

	fmt.Printf("Digestor: generated digest for %s (%d items, %d sources)\n",
		*date, len(readyItems), len(seenSources))
	return nil
}

func loadItems(date string) ([]*schema.Item, error) {
	path := filepath.Join("data/items", fmt.Sprintf("%s.ndjson", date))
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items []*schema.Item
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		var item schema.Item
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse line %d: %v\n", lineNum, err)
			continue
		}
		items = append(items, &item)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func writeDigest(date string, digest *schema.DailyDigest) error {
	dir := "data/digests"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, fmt.Sprintf("%s.json", date))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(digest)
}

func writeEmptyTopics(date string) error {
	dir := "data/topics"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, fmt.Sprintf("%s.json", date))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write empty topics array for compatibility
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	_, err = file.WriteString("[]\n")
	return err
}

func countRawItems(date string) int {
	path := filepath.Join("data/raw", fmt.Sprintf("%s.ndjson", date))
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}
	return count
}

func formatDisplayDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("January 2, 2006")
}