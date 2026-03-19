package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ai-daily-brief/ai-daily-brief/internal/openclaw"
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
	itemMap := make(map[string]*schema.Item)
	for _, item := range items {
		if item.Status == "ready" {
			readyItems = append(readyItems, item)
			seenSources[item.SourceID] = true
			itemMap[item.ID] = item
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

	// Build stats
	stats := schema.DigestStats{
		RawItems:        countRawItems(*date),
		NormalizedItems: len(items),
		PublishedItems:  len(readyItems),
		Topics:          0,
		Sources:         len(seenSources),
	}

	// Create Openclaw client
	oc := createOpenclawClient()
	ctx := context.Background()

	// Step 1: Summarize items via Openclaw
	var clusters []openclaw.TopicCluster
	var composeResp *openclaw.ComposeDigestResponse

	if len(topItems) > 0 {
		if *verbose {
			fmt.Println("Calling Openclaw: summarize_items...")
		}
		summarizeInput := buildSummarizeInput(topItems)
		summarizeResp, err := oc.SummarizeItems(ctx, summarizeInput.Items)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: SummarizeItems failed: %v (continuing without summaries)\n", err)
		} else {
			// Apply summaries to items
			for _, s := range summarizeResp.Summaries {
				if item, ok := itemMap[s.ItemID]; ok {
					item.Summary1Line = s.Summary1Line
				}
			}
			if *verbose {
				fmt.Printf("  Generated %d summaries\n", len(summarizeResp.Summaries))
			}
		}

		// Step 2: Cluster topics via Openclaw
		if *verbose {
			fmt.Println("Calling Openclaw: cluster_topics...")
		}
		clusterInput := buildClusterInput(*date, topItems)
		clusterResp, err := oc.ClusterTopics(ctx, *date, clusterInput.Items, 10)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: ClusterTopics failed: %v (continuing without clustering)\n", err)
		} else {
			clusters = clusterResp.Clusters
			stats.Topics = len(clusters)
			if *verbose {
				fmt.Printf("  Generated %d topic clusters\n", len(clusters))
			}
		}

		// Step 3: Compose digest via Openclaw
		if *verbose {
			fmt.Println("Calling Openclaw: compose_digest...")
		}
		composeInput := buildComposeInput(*date, clusters, topItems, stats)
		composeResp, err = oc.ComposeDigest(ctx, *date, clusters, composeInput.TopItems, composeInput.Stats)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: ComposeDigest failed: %v (using fallback headline)\n", err)
			composeResp = nil
		} else if *verbose {
			fmt.Printf("  Headline: %s\n", composeResp.Headline)
		}
	}

	// Build digest
	digest := buildDigest(*date, topItems, clusters, composeResp, stats)

	if *verbose {
		fmt.Printf("Digest: %d items, %d sources, %d topics\n",
			len(readyItems), len(seenSources), len(clusters))
	}

	if *dryRun {
		fmt.Printf("Dry run: would write digest to data/digests/%s.json\n", *date)
		return nil
	}

	// Write updated items with summaries
	if err := writeItems(*date, items); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write items: %v\n", err)
	}

	// Write digest
	if err := writeDigest(*date, digest); err != nil {
		return fmt.Errorf("failed to write digest: %w", err)
	}

	// Write topics
	if err := writeTopics(*date, clusters); err != nil {
		return fmt.Errorf("failed to write topics: %w", err)
	}

	// Step 4: QA digest via Openclaw (best effort)
	if len(topItems) > 0 && composeResp != nil {
		if *verbose {
			fmt.Println("Calling Openclaw: qa_digest...")
		}
		qaInput := openclaw.QADigestInput{
			Digest:   *composeResp,
			Items:    buildQAItems(topItems),
			Clusters: clusters,
		}
		qaResp, err := oc.QADigest(ctx, *composeResp, qaInput.Items, clusters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: QADigest failed: %v\n", err)
		} else if !qaResp.QAResult.Passed {
			fmt.Fprintf(os.Stderr, "Warning: QA issues detected: %v\n", qaResp.QAResult.Issues)
			if len(qaResp.QAResult.Warnings) > 0 && *verbose {
				fmt.Fprintf(os.Stderr, "  Warnings: %v\n", qaResp.QAResult.Warnings)
			}
		}
	}

	fmt.Printf("Digestor: generated digest for %s (%d items, %d sources, %d topics)\n",
		*date, len(readyItems), len(seenSources), len(clusters))
	return nil
}

func createOpenclawClient() *openclaw.Clientv2 {
	execPath := getEnv("OPENCLAW_EXEC", "acpx")
	agent := getEnv("OPENCLAW_AGENT", "")
	endpoint := getEnv("OPENCLAW_ENDPOINT", "")
	token := getEnv("OPENCLAW_TOKEN", "")

	exec := openclaw.NewDefaultExecutor(execPath, agent, endpoint, token)
	return openclaw.NewClientv2(exec)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
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

	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return err
		}
	}

	return nil
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

func writeTopics(date string, clusters []openclaw.TopicCluster) error {
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

	// Convert to schema TopicCluster
	schemaClusters := make([]schema.TopicCluster, 0, len(clusters))
	for _, c := range clusters {
		schemaClusters = append(schemaClusters, schema.TopicCluster{
			TopicID:         c.TopicID,
			Name:            c.Name,
			Summary:         c.Summary,
			WhyItMatters:    c.WhyItMatters,
			Keywords:        c.Keywords,
			ImportanceScore: c.ImportanceScore,
			ItemIDs:         c.ItemIDs,
		})
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(schemaClusters)
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

func buildDigest(date string, topItems []*schema.Item, clusters []openclaw.TopicCluster, composeResp *openclaw.ComposeDigestResponse, stats schema.DigestStats) *schema.DailyDigest {
	headline := fmt.Sprintf("AI Daily Brief - %s", formatDisplayDate(date))
	lead := fmt.Sprintf("Digest of %d items from %d sources.", stats.PublishedItems, stats.Sources)
	topTopicIDs := make([]string, 0)
	topItemIDs := make([]string, 0, len(topItems))

	for _, id := range topItems {
		topItemIDs = append(topItemIDs, id.ID)
	}

	if composeResp != nil {
		if composeResp.Headline != "" {
			headline = composeResp.Headline
		}
		if composeResp.Lead != "" {
			lead = composeResp.Lead
		}
		topTopicIDs = composeResp.TopTopicIDs
		topItemIDs = composeResp.TopItemIDs
		if topItemIDs == nil {
			topItemIDs = make([]string, 0, len(topItems))
			for _, id := range topItems {
				topItemIDs = append(topItemIDs, id.ID)
			}
		}
	}

	if topTopicIDs == nil {
		topTopicIDs = make([]string, 0)
		for _, c := range clusters {
			topTopicIDs = append(topTopicIDs, c.TopicID)
		}
	}

	return &schema.DailyDigest{
		Date:        date,
		Edition:     "nightly",
		Headline:    headline,
		Lead:        lead,
		TopTopicIDs: topTopicIDs,
		TopItemIDs:  topItemIDs,
		Stats:       stats,
	}
}

func buildSummarizeInput(items []*schema.Item) openclaw.SummarizeItemsInput {
	summarizeItems := make([]openclaw.SummarizeItem, 0, len(items))
	for _, item := range items {
		summarizeItems = append(summarizeItems, openclaw.SummarizeItem{
			ItemID:      item.ID,
			Title:       item.Title,
			Content:     item.ContentText,
			Source:      item.SourceID,
			PublishedAt: item.PublishedAt,
		})
	}
	return openclaw.SummarizeItemsInput{Items: summarizeItems}
}

func buildClusterInput(date string, items []*schema.Item) openclaw.ClusterTopicsInput {
	clusterItems := make([]openclaw.ClusterItem, 0, len(items))
	for _, item := range items {
		summary := item.Summary1Line
		if summary == "" {
			summary = item.Title
		}
		clusterItems = append(clusterItems, openclaw.ClusterItem{
			ItemID:  item.ID,
			Title:   item.Title,
			Summary: summary,
			Source:  item.SourceID,
			Domain:  item.Domain,
			PubTime: item.PublishedAt,
		})
	}
	return openclaw.ClusterTopicsInput{
		Date:        date,
		Items:       clusterItems,
		MaxClusters: 10,
	}
}

func buildComposeInput(date string, clusters []openclaw.TopicCluster, items []*schema.Item, stats schema.DigestStats) openclaw.ComposeDigestInput {
	topItems := make([]openclaw.ComposeItem, 0, len(items))
	for _, item := range items {
		if len(topItems) >= 10 {
			break
		}
		summary := item.Summary1Line
		if summary == "" {
			summary = item.Title
		}
		topItems = append(topItems, openclaw.ComposeItem{
			ItemID:  item.ID,
			Title:   item.Title,
			Summary: summary,
			Source:  item.SourceID,
			Domain:  item.Domain,
			URL:     item.CanonicalURL,
			PubTime: item.PublishedAt,
		})
	}

	return openclaw.ComposeDigestInput{
		Date:     date,
		Clusters: clusters,
		TopItems: topItems,
		Stats: openclaw.DigestStats{
			RawItems:        stats.RawItems,
			NormalizedItems: stats.NormalizedItems,
			PublishedItems:  stats.PublishedItems,
			Topics:          stats.Topics,
			Sources:         stats.Sources,
		},
	}
}

func buildQAItems(items []*schema.Item) []openclaw.QAItem {
	qaItems := make([]openclaw.QAItem, 0, len(items))
	for _, item := range items {
		if len(qaItems) >= 10 {
			break
		}
		summary := item.Summary1Line
		if summary == "" {
			summary = item.Title
		}
		qaItems = append(qaItems, openclaw.QAItem{
			ItemID:  item.ID,
			Title:   item.Title,
			Summary: summary,
			Source:  item.SourceID,
			URL:     item.CanonicalURL,
		})
	}
	return qaItems
}

func formatDisplayDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("January 2, 2006")
}