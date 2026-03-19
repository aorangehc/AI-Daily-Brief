package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const stateDir = "data/state"

// Load loads the state file for a given date
func Load(date string) (*State, error) {
	path := filepath.Join(stateDir, fmt.Sprintf("%s.json", date))

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}

	return &st, nil
}

// Save saves the state file for a given date
func Save(date string, st *State) error {
	path := filepath.Join(stateDir, fmt.Sprintf("%s.json", date))

	st.LastUpdatedAt = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// New creates a new state for a given date
func New(date string) *State {
	return &State{
		Date:    date,
		Collect: make(map[string]string),
	}
}
