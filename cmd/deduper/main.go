package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ai-daily-brief/ai-daily-brief/internal/dedupe"
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
	deduper := dedupe.New()

	// Read normalized items
	itemPath := filepath.Join("data/items", fmt.Sprintf("%s.ndjson", *date))
	itemFile, err := os.Open(itemPath)
	if err != nil {
		return fmt.Errorf("failed to open items file: %w", err)
	}
	defer itemFile.Close()

	var items []*schema.Item
	scanner := bufio.NewScanner(itemFile)
	lineNum := 0

	if *verbose {
		fmt.Printf("Reading items from %s for deduplication\n", itemPath)
	}

	for scanner.Scan() {
		lineNum++
		var item schema.Item
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			if *verbose {
				fmt.Fprintf(os.Stderr, "  Warning: failed to parse line %d: %v\n", lineNum, err)
			}
			continue
		}
		items = append(items, &item)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read items file: %w", err)
	}

	if *verbose {
		fmt.Printf("Read %d items, running deduplication\n", len(items))
	}

	// Deduplicate
	dedupedItems := deduper.Dedup(items)

	if *verbose {
		fmt.Printf("Deduplication: %d items -> %d items (removed %d duplicates)\n",
			len(items), len(dedupedItems), len(items)-len(dedupedItems))
	}

	if *dryRun {
		fmt.Printf("Dry run: would write %d items to data/items/%s_deduped.ndjson\n", len(dedupedItems), *date)
		return nil
	}

	// Write deduplicated items (overwrite original for next stage)
	if err := writeItems(*date, dedupedItems); err != nil {
		return fmt.Errorf("failed to write deduped items: %w", err)
	}

	fmt.Printf("Deduper: %d items -> %d unique items for %s\n",
		len(items), len(dedupedItems), *date)
	return nil
}

// writeItems writes items to a NDJSON file
func writeItems(date string, items []*schema.Item) error {
	dir := "data/items"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, fmt.Sprintf("%s.ndjson", date))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, item := range items {
		if err := encoder.Encode(item); err != nil {
			return err
		}
	}

	return nil
}
