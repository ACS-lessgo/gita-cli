package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// persistedState is saved to disk between runs.
type persistedState struct {
	LastChapter int        `json:"lastChapter"`
	LastVerse   int        `json:"lastVerse"`
	Bookmarks   []bookmark `json:"bookmarks"`
}

type bookmark struct {
	ChapterNum int       `json:"chapterNum"`
	VerseNum   int       `json:"verseNum"`
	SavedAt    time.Time `json:"savedAt"`
}

func statePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gita-cli", "state.json"), nil
}

// loadState reads persisted state from disk. A missing or corrupt file
// returns a zero-value persistedState alongside the error — callers that
// don't care why loading failed can ignore the error and use the result.
func loadState() (persistedState, error) {
	path, err := statePath()
	if err != nil {
		return persistedState{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return persistedState{}, err
	}
	var st persistedState
	if err := json.Unmarshal(data, &st); err != nil {
		return persistedState{}, err
	}
	return st, nil
}

func saveState(st persistedState) error {
	path, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
