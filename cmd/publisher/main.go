package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	date    = flag.String("date", "", "Date in YYYY-MM-DD format")
	dryRun  = flag.Bool("dry-run", false, "Dry run mode")
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
		fmt.Printf("Publishing %s\n", *date)
	}

	// TODO: Implement publishing
	// 1. Git add and commit data changes
	// 2. Create git tag for the date
	// 3. Trigger render workflow
	// 4. Update state file

	fmt.Printf("Publisher: published %s\n", *date)
	return nil
}
