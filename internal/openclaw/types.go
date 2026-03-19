package openclaw

import "time"

// Client handles Openclaw ACP CLI (acpx) integration
type Client struct {
	agent    string
	endpoint string
	token    string
	execPath string
}

// NewClient creates a new Openclaw client
func NewClient(agent, endpoint, token, execPath string) *Client {
	if execPath == "" {
		execPath = "acpx"
	}
	return &Client{
		agent:    agent,
		endpoint: endpoint,
		token:    token,
		execPath: execPath,
	}
}

// Task types supported by Openclaw
const (
	TaskClusterTopics   = "cluster_topics"
	TaskSummarizeItems  = "summarize_items"
	TaskComposeDigest  = "compose_digest"
	TaskQADigest       = "qa_digest"
)

// BaseResponse is the common response structure from Openclaw
type BaseResponse struct {
	Version     string    `json:"version"`
	GeneratedAt time.Time `json:"generated_at"`
	ModelInfo   ModelInfo `json:"model_info"`
}

// ModelInfo contains information about the model used
type ModelInfo struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// ClusterTopicsInput is the input for cluster_topics task
type ClusterTopicsInput struct {
	Date       string         `json:"date"`
	Items      []ClusterItem  `json:"items"`
	MaxClusters int           `json:"max_clusters"`
}

// ClusterItem is a simplified item for clustering
type ClusterItem struct {
	ItemID   string `json:"item_id"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Source   string `json:"source"`
	Domain   string `json:"domain"`
	PubTime  string `json:"published_at"`
}

// ClusterTopicsResponse is the response from cluster_topics task
type ClusterTopicsResponse struct {
	BaseResponse
	Clusters []TopicCluster `json:"clusters"`
}

// TopicCluster represents a cluster of related items
type TopicCluster struct {
	TopicID         string   `json:"topic_id"`
	Name            string   `json:"name"`
	Summary         string   `json:"summary"`
	WhyItMatters    string   `json:"why_it_matters"`
	Keywords        []string `json:"keywords"`
	ImportanceScore float64  `json:"importance_score"`
	ItemIDs         []string `json:"item_ids"`
}

// SummarizeItemsInput is the input for summarize_items task
type SummarizeItemsInput struct {
	Items []SummarizeItem `json:"items"`
}

// SummarizeItem is an item to summarize
type SummarizeItem struct {
	ItemID      string `json:"item_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Source      string `json:"source"`
	PublishedAt string `json:"published_at"`
}

// SummarizeItemsResponse is the response from summarize_items task
type SummarizeItemsResponse struct {
	BaseResponse
	Summaries []ItemSummary `json:"summaries"`
}

// ItemSummary is a one-line summary for an item
type ItemSummary struct {
	ItemID       string `json:"item_id"`
	Summary1Line string `json:"summary_1line"`
}

// ComposeDigestInput is the input for compose_digest task
type ComposeDigestInput struct {
	Date      string            `json:"date"`
	Clusters  []TopicCluster    `json:"clusters"`
	TopItems  []ComposeItem     `json:"top_items"`
	Stats     DigestStats       `json:"stats"`
}

// ComposeItem is a top item for digest composition
type ComposeItem struct {
	ItemID    string `json:"item_id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Source    string `json:"source"`
	Domain    string `json:"domain"`
	URL       string `json:"url"`
	PubTime   string `json:"published_at"`
}

// DigestStats contains digest statistics
type DigestStats struct {
	RawItems        int `json:"raw_items"`
	NormalizedItems int `json:"normalized_items"`
	PublishedItems  int `json:"published_items"`
	Topics          int `json:"topics"`
	Sources         int `json:"sources"`
}

// ComposeDigestResponse is the response from compose_digest task
type ComposeDigestResponse struct {
	BaseResponse
	Headline    string   `json:"headline"`
	Lead        string   `json:"lead"`
	TopTopicIDs []string `json:"top_topic_ids"`
	TopItemIDs  []string `json:"top_item_ids"`
}

// QADigestInput is the input for qa_digest task
type QADigestInput struct {
	Digest  ComposeDigestResponse `json:"digest"`
	Items   []QAItem              `json:"items"`
	Clusters []TopicCluster       `json:"clusters"`
}

// QAItem is an item for QA checking
type QAItem struct {
	ItemID string `json:"item_id"`
	Title  string `json:"title"`
	Summary string `json:"summary"`
	Source string `json:"source"`
	URL    string `json:"url"`
}

// QADigestResponse is the response from qa_digest task
type QADigestResponse struct {
	BaseResponse
	QAResult QAResult `json:"qa_result"`
}

// QAResult contains the QA check results
type QAResult struct {
	Passed   bool     `json:"passed"`
	Issues   []string `json:"issues"`
	Warnings []string `json:"warnings"`
}