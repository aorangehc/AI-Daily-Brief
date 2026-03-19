package source

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Source represents a news source configuration
type Source struct {
	ID              string   `yaml:"id"`
	Name            string   `yaml:"name"`
	Type            string   `yaml:"type"` // rss, api, html
	Enabled         bool     `yaml:"enabled"`
	Category        string   `yaml:"category"`
	BaseURL         string   `yaml:"base_url"`
	FeedURL         string   `yaml:"feed_url,omitempty"`
	APIURL          string   `yaml:"api_url,omitempty"`
	Language        string   `yaml:"language"`
	Weight          float64  `yaml:"weight"`
	Parser          string   `yaml:"parser"`
	RateLimitPerRun int      `yaml:"rate_limit_per_run"`
	AllowPaths      []string `yaml:"allow_paths"`
	DenyPaths       []string `yaml:"deny_paths"`
}

// Sources represents the sources configuration
type Sources struct {
	Sources []Source `yaml:"sources"`
}

// Load loads sources from sources.yaml
func Load(path string) (*Sources, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read sources file: %w", err)
	}

	var srcs Sources
	if err := yaml.Unmarshal(data, &srcs); err != nil {
		return nil, fmt.Errorf("failed to parse sources file: %w", err)
	}

	return &srcs, nil
}

// LoadDefault loads the default sources.yaml
func LoadDefault() (*Sources, error) {
	return Load("sources/sources.yaml")
}

// Enabled returns only enabled sources
func (s *Sources) Enabled() []Source {
	var enabled []Source
	for _, src := range s.Sources {
		if src.Enabled {
			enabled = append(enabled, src)
		}
	}
	return enabled
}
