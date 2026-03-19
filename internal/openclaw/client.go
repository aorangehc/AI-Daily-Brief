package openclaw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Executor defines the interface for executing acpx commands
type Executor interface {
	Execute(ctx context.Context, task string, input interface{}) ([]byte, error)
}

// DefaultExecutor uses exec.Cmd to run acpx
type DefaultExecutor struct {
	execPath string
	agent    string
	endpoint string
	token    string
}

// NewDefaultExecutor creates a new executor for acpx commands
func NewDefaultExecutor(execPath, agent, endpoint, token string) *DefaultExecutor {
	if execPath == "" {
		execPath = "acpx"
	}
	return &DefaultExecutor{
		execPath: execPath,
		agent:    agent,
		endpoint: endpoint,
		token:    token,
	}
}

// Execute runs acpx with the given task and input, returns raw output
func (e *DefaultExecutor) Execute(ctx context.Context, task string, input interface{}) ([]byte, error) {
	// Serialize input to JSON
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Build acpx command args
	args := []string{task}

	// Add optional flags
	if e.agent != "" {
		args = append(args, "--agent", e.agent)
	}
	if e.endpoint != "" {
		args = append(args, "--endpoint", e.endpoint)
	}
	if e.token != "" {
		args = append(args, "--token", e.token)
	}

	// Add input flag (acpx expects --input with JSON string or @filepath)
	args = append(args, "--input", string(inputJSON))

	// For debugging, log the command
	// fmt.Fprintf(os.Stderr, "DEBUG: running %s %v\n", e.execPath, args)

	cmd := exec.CommandContext(ctx, e.execPath, args...)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("acpx %s failed: %w", task, err)
	}

	return output, nil
}

// Client wraps the executor for Openclaw operations
type Clientv2 struct {
	exec Executor
}

// NewClientv2 creates a new Openclaw client with the given executor
func NewClientv2(exec Executor) *Clientv2 {
	return &Clientv2{exec: exec}
}

// ClusterTopics performs topic clustering on items
func (c *Clientv2) ClusterTopics(ctx context.Context, date string, items []ClusterItem, maxClusters int) (*ClusterTopicsResponse, error) {
	input := ClusterTopicsInput{
		Date:        date,
		Items:       items,
		MaxClusters: maxClusters,
	}

	output, err := c.exec.Execute(ctx, TaskClusterTopics, input)
	if err != nil {
		return nil, err
	}

	var resp ClusterTopicsResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse cluster response: %w", err)
	}

	return &resp, nil
}

// SummarizeItems generates one-line summaries for items
func (c *Clientv2) SummarizeItems(ctx context.Context, items []SummarizeItem) (*SummarizeItemsResponse, error) {
	input := SummarizeItemsInput{Items: items}

	output, err := c.exec.Execute(ctx, TaskSummarizeItems, input)
	if err != nil {
		return nil, err
	}

	var resp SummarizeItemsResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse summarize response: %w", err)
	}

	return &resp, nil
}

// ComposeDigest generates the digest headline, lead, and top selections
func (c *Clientv2) ComposeDigest(ctx context.Context, date string, clusters []TopicCluster, topItems []ComposeItem, stats DigestStats) (*ComposeDigestResponse, error) {
	input := ComposeDigestInput{
		Date:     date,
		Clusters: clusters,
		TopItems: topItems,
		Stats:    stats,
	}

	output, err := c.exec.Execute(ctx, TaskComposeDigest, input)
	if err != nil {
		return nil, err
	}

	var resp ComposeDigestResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse compose response: %w", err)
	}

	return &resp, nil
}

// QADigest performs quality assurance on a digest
func (c *Clientv2) QADigest(ctx context.Context, digest ComposeDigestResponse, items []QAItem, clusters []TopicCluster) (*QADigestResponse, error) {
	input := QADigestInput{
		Digest:   digest,
		Items:    items,
		Clusters: clusters,
	}

	output, err := c.exec.Execute(ctx, TaskQADigest, input)
	if err != nil {
		return nil, err
	}

	var resp QADigestResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse QA response: %w", err)
	}

	return &resp, nil
}

// PromptLoader loads prompt templates from files
type PromptLoader struct {
	promptsDir string
}

// NewPromptLoader creates a new prompt loader
func NewPromptLoader(promptsDir string) *PromptLoader {
	if promptsDir == "" {
		promptsDir = "prompts"
	}
	return &PromptLoader{promptsDir: promptsDir}
}

// Load reads a prompt file and returns its contents
func (p *PromptLoader) Load(name string) (string, error) {
	path := filepath.Join(p.promptsDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt %s: %w", name, err)
	}
	return string(data), nil
}

// WriteInput writes input data to a temp file and returns the path
func WriteInput(data interface{}) (string, error) {
	tmpDir := os.Getenv("TMPDIR")
	if tmpDir == "" {
		tmpDir = "/tmp"
	}

	inputJSON, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("openclaw-input-%d.json", time.Now().UnixNano()))
	if err := os.WriteFile(tmpFile, inputJSON, 0600); err != nil {
		return "", err
	}

	return tmpFile, nil
}

// ParseResponse parses JSON output into the given struct
func ParseResponse(output []byte, v interface{}) error {
	// Try to find and extract JSON from output (in case of logging)
	output = bytes.TrimSpace(output)

	// Handle potential markdown code blocks
	if len(output) > 0 && output[0] == '`' {
		// Find first { and last }
		start := bytes.Index(output, []byte("{"))
		end := bytes.LastIndex(output, []byte("}"))
		if start != -1 && end != -1 && end > start {
			output = output[start : end+1]
		}
	}

	return json.Unmarshal(output, v)
}