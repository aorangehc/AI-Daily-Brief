package main

import (
	"flag"
	"fmt"
	"os"
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
		fmt.Printf("Normalizing raw items for %s\n", *date)
	}

	// TODO: Implement normalization
	// 1. Read raw items from data/raw/YYYY-MM-DD.ndjson
	// 2. Convert to standardized item format
	// 3. Write to data/items/YYYY-MM-DD.ndjson

	fmt.Printf("Normalizer: normalized items for %s\n", *date)
	return nil
}
