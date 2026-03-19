package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ai-daily-brief/ai-daily-brief/internal/fetch"
	"github.com/ai-daily-brief/ai-daily-brief/internal/schema"
	"github.com/ai-daily-brief/ai-daily-brief/internal/source"
	"github.com/ai-daily-brief/ai-daily-brief/internal/state"
)

var (
	date    = flag.String("date", "", "Date in YYYY-MM-DD format")
	batch   = flag.String("batch", "", "Batch identifier (09, 13, 17, 21)")
	srcFlag = flag.String("source", "", "Specific source ID to collect")
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

	if *batch == "" {
		fmt.Fprintln(os.Stderr, "Error: --batch is required")
		flag.Usage()
		os.Exit(1)
	}

	// Check state for idempotency
	if !*force {
		st, err := state.Load(*date)
		if err == nil && st.Collect[*batch] == "success" {
			fmt.Println("Collection already completed for", *date, "batch", *batch)
			os.Exit(0)
		}
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// Load sources
	sources, err := source.LoadDefault()
	if err != nil {
		return fmt.Errorf("failed to load sources: %w", err)
	}

	// Filter sources if --source flag is set
	sourcesToCollect := sources.Enabled()
	if *srcFlag != "" {
		var filtered []source.Source
		for _, s := range sourcesToCollect {
			if s.ID == *srcFlag {
				filtered = append(filtered, s)
				break
			}
		}
		sourcesToCollect = filtered
	}

	if len(sourcesToCollect) == 0 {
		return fmt.Errorf("no sources to collect")
	}

	if *verbose {
		fmt.Printf("Collecting from %d sources for %s batch %s\n", len(sourcesToCollect), *date, *batch)
	}

	// Collect items from all sources
	var allItems []*schema.RawItem
	for _, src := range sourcesToCollect {
		if *verbose {
			fmt.Printf("  Fetching from %s (%s)...\n", src.Name, src.Type)
		}

		fetcher := fetch.Factory(&src)
		if fetcher == nil {
			if *verbose {
				fmt.Printf("  Unknown source type: %s\n", src.Type)
			}
			continue
		}

		items, err := fetcher.Fetch(ctx, &src)
		if err != nil {
			if *verbose {
				fmt.Printf("  Error fetching %s: %v\n", src.Name, err)
			}
			continue
		}

		if *verbose {
			fmt.Printf("  Collected %d items from %s\n", len(items), src.Name)
		}

		allItems = append(allItems, items...)
	}

	if *dryRun {
		fmt.Printf("Dry run: would write %d items to data/raw/%s.ndjson\n", len(allItems), *date)
		return nil
	}

	// Write items to NDJSON file
	if err := writeRawItems(*date, allItems); err != nil {
		return fmt.Errorf("failed to write raw items: %w", err)
	}

	// Update state
	st, err := state.Load(*date)
	if err != nil {
		st = state.New(*date)
	}
	st.Collect[*batch] = "success"
	if err := state.Save(*date, st); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Printf("Collector: collected %d items for %s batch %s\n", len(allItems), *date, *batch)
	return nil
}

// writeRawItems writes raw items to a NDJSON file
func writeRawItems(date string, items []*schema.RawItem) error {
	// Ensure directory exists
	dir := "data/raw"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, fmt.Sprintf("%s.ndjson", date))
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

// ensureState creates state file if it doesn't exist
func ensureState(date string) (*state.State, error) {
	st, err := state.Load(date)
	if err != nil {
		if os.IsNotExist(err) {
			st = state.New(date)
		} else {
			return nil, err
		}
	}
	return st, nil
}
