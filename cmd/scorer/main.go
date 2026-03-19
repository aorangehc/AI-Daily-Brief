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
		fmt.Printf("Scoring items for %s\n", *date)
	}

	// TODO: Implement scoring
	// Score = source_weight * 0.30 + freshness_score * 0.25 + heat_score * 0.20 + originality_score * 0.15 + topic_importance_hint * 0.10

	fmt.Printf("Scorer: scored items for %s\n", *date)
	return nil
}
