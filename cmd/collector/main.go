package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ai-daily-brief/ai-daily-brief/internal/state"
)

var (
	date       = flag.String("date", "", "Date in YYYY-MM-DD format")
	batch      = flag.String("batch", "", "Batch identifier (09, 13, 17, 21)")
	source     = flag.String("source", "", "Specific source ID to collect")
	dryRun     = flag.Bool("dry-run", false, "Dry run mode")
	force      = flag.Bool("force", false, "Force execution even if already done")
	verbose    = flag.Bool("verbose", false, "Verbose output")
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
	if *verbose {
		fmt.Printf("Collecting from sources for %s batch %s\n", *date, *batch)
	}

	// TODO: Implement actual collection logic
	// 1. Load sources from sources/sources.yaml
	// 2. Filter by --source if specified
	// 3. Fetch from each source based on type (rss/api/html)
	// 4. Write raw items to data/raw/YYYY-MM-DD.ndjson
	// 5. Update state file

	fmt.Printf("Collector: collected for %s batch %s\n", *date, *batch)
	return nil
}
