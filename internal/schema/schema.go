package schema

// RawItem represents the raw item fetched from a source
type RawItem struct {
	ID          string `json:"id"`
	SourceID    string `json:"source_id"`
	SourceType  string `json:"source_type"` // rss, api, html
	Title       string `json:"title"`
	URL         string `json:"url"`
	Author      string `json:"author,omitempty"`
	PublishedAt string `json:"published_at"`
	CollectedAt string `json:"collected_at"`
	ContentRaw  string `json:"content_raw"`
	Lang        string `json:"lang"`
}

// Item represents the normalized item
type Item struct {
	ID              string   `json:"id"`
	RawID           string   `json:"raw_id"`
	CanonicalURL    string   `json:"canonical_url"`
	Domain          string   `json:"domain"`
	Title           string   `json:"title"`
	Summary1Line    string   `json:"summary_1line"`
	ContentText     string   `json:"content_text"`
	PublishedAt     string   `json:"published_at"`
	Lang            string   `json:"lang"`
	Tags            []string `json:"tags"`
	SourceWeight    float64  `json:"source_weight"`
	FreshnessScore  float64  `json:"freshness_score"`
	HeatScore       float64  `json:"heat_score"`
	OriginalityScore float64 `json:"originality_score"`
	FinalScore      float64  `json:"final_score"`
	HashURL         string   `json:"hash_url"`
	HashTitle       string   `json:"hash_title"`
	HashContent     string   `json:"hash_content"`
	Status          string   `json:"status"` // ready, reject, pending
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

// DailyDigest represents the daily digest
type DailyDigest struct {
	Date       string   `json:"date"`
	Edition    string   `json:"edition"` // nightly
	Headline   string   `json:"headline"`
	Lead       string   `json:"lead"`
	TopTopicIDs []string `json:"top_topic_ids"`
	TopItemIDs  []string `json:"top_item_ids"`
	Stats      DigestStats `json:"stats"`
}

// DigestStats contains statistics for the digest
type DigestStats struct {
	RawItems       int `json:"raw_items"`
	NormalizedItems int `json:"normalized_items"`
	PublishedItems int `json:"published_items"`
	Topics         int `json:"topics"`
	Sources        int `json:"sources"`
}

// State represents the system state
type State struct {
	Date          string            `json:"date"`
	Collect       map[string]string `json:"collect"` // batch -> status
	Digest        string            `json:"digest"`
	Deploy        string            `json:"deploy"`
	LastUpdatedAt string            `json:"last_updated_at"`
}
