package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ai-daily-brief/ai-daily-brief/internal/normalize"
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
	normalizer := normalize.New()

	// Read raw items
	rawPath := filepath.Join("data/raw", fmt.Sprintf("%s.ndjson", *date))
	rawFile, err := os.Open(rawPath)
	if err != nil {
		return fmt.Errorf("failed to open raw file: %w", err)
	}
	defer rawFile.Close()

	var items []*schema.Item
	scanner := bufio.NewScanner(rawFile)
	lineNum := 0

	if *verbose {
		fmt.Printf("Normalizing raw items from %s\n", rawPath)
	}

	for scanner.Scan() {
		lineNum++
		var raw schema.RawItem
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			if *verbose {
				fmt.Fprintf(os.Stderr, "  Warning: failed to parse line %d: %v\n", lineNum, err)
			}
			continue
		}

		item := normalizer.Normalize(&raw)
		items = append(items, item)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read raw file: %w", err)
	}

	if *verbose {
		fmt.Printf("Normalized %d items\n", len(items))
	}

	if *dryRun {
		fmt.Printf("Dry run: would write %d items to data/items/%s.ndjson\n", len(items), *date)
		return nil
	}

	// Write normalized items
	if err := writeItems(*date, items); err != nil {
		return fmt.Errorf("failed to write items: %w", err)
	}

	fmt.Printf("Normalizer: normalized %d items for %s\n", len(items), *date)
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
