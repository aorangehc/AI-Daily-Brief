package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	date    = flag.String("date", "", "Date in YYYY-MM-DD format")
	dryRun  = flag.Bool("dry-run", false, "Dry run mode")
	verbose = flag.Bool("verbose", false, "Verbose output")
	force   = flag.Bool("force", false, "Force execution even if already done")
	message = flag.String("message", "", "Custom commit message")
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

	// Check if already published
	if !*force && isPublished(*date) {
		if *verbose {
			fmt.Printf("Already published for %s, skipping\n", *date)
		}
		return nil
	}

	// Build commit message
	commitMsg := *message
	if commitMsg == "" {
		commitMsg = fmt.Sprintf("feat: generate digest for %s", *date)
	}

	if *dryRun {
		fmt.Printf("Dry run: would commit and tag for %s\n", *date)
		fmt.Printf("  Commit message: %s\n", commitMsg)
		return nil
	}

	// Stage data files
	dataDirs := []string{
		"data/digests",
		"data/topics",
		"data/items",
		"site/public/data",
	}

	for _, dir := range dataDirs {
		pattern := filepath.Join(dir, fmt.Sprintf("%s.*", *date))
		if err := gitAdd(pattern); err != nil {
			return fmt.Errorf("git add %s failed: %w", pattern, err)
		}
	}

	// Also stage index files
	indexFiles := []string{
		"site/public/data/indexes/archive.json",
		"site/public/data/indexes/latest.json",
		"site/public/data/indexes/search-index.json",
		"site/public/data/indexes/sources.json",
	}
	for _, f := range indexFiles {
		if _, err := os.Stat(f); err == nil {
			if err := gitAdd(f); err != nil {
				return fmt.Errorf("git add %s failed: %w", f, err)
			}
		}
	}

	// Check if there are staged changes
	staged, err := hasStagedChanges()
	if err != nil {
		return fmt.Errorf("checking staged changes: %w", err)
	}

	if !staged {
		if *verbose {
			fmt.Printf("No changes to commit for %s\n", *date)
		}
	} else {
		// Commit
		if err := gitCommit(commitMsg); err != nil {
			return fmt.Errorf("git commit failed: %w", err)
		}
		if *verbose {
			fmt.Printf("Committed changes for %s\n", *date)
		}
	}

	// Create tag
	tagName := fmt.Sprintf("digest-%s", strings.ReplaceAll(*date, "-", ""))
	if err := gitTag(tagName, commitMsg); err != nil {
		return fmt.Errorf("git tag failed: %w", err)
	}
	if *verbose {
		fmt.Printf("Created tag %s\n", tagName)
	}

	// Push
	if err := gitPush(tagName); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}
	if *verbose {
		fmt.Printf("Pushed tag %s\n", tagName)
	}

	// Update state file
	if err := updateState(*date); err != nil {
		return fmt.Errorf("updating state: %w", err)
	}

	fmt.Printf("Publisher: published %s\n", *date)
	return nil
}

func isPublished(date string) bool {
	statePath := filepath.Join("data/state", fmt.Sprintf("%s.json", date))
	data, err := os.ReadFile(statePath)
	if err != nil {
		return false
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return false
	}

	return state.Deploy == "success"
}

func gitAdd(pattern string) error {
	cmd := exec.Command("git", "add", pattern)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func hasStagedChanges() (bool, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

func gitCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitTag(name, message string) error {
	cmd := exec.Command("git", "tag", "-a", name, "-m", message)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitPush(tagName string) error {
	// Push the tag
	cmd := exec.Command("git", "push", "origin", tagName)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Also push any commits if they exist
	cmd = exec.Command("git", "push", "origin", "HEAD:main")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func updateState(date string) error {
	stateDir := filepath.Join("data/state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return err
	}

	statePath := filepath.Join(stateDir, fmt.Sprintf("%s.json", date))

	var state State
	if data, err := os.ReadFile(statePath); err == nil {
		json.Unmarshal(data, &state)
	}

	state.Date = date
	state.Deploy = "success"
	state.LastUpdatedAt = time.Now().UTC().Format(time.RFC3339)

	file, err := os.OpenFile(statePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(state)
}

// State represents the system state
type State struct {
	Date          string            `json:"date"`
	Collect      map[string]string `json:"collect"`
	Digest       string            `json:"digest"`
	Deploy       string            `json:"deploy"`
	LastUpdatedAt string            `json:"last_updated_at"`
}