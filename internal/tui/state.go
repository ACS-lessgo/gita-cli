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

// userConfigDir is os.UserConfigDir by default; tests override it directly
// so sandboxing doesn't depend on OS-specific env var behavior (XDG_CONFIG_HOME
// is only honored by os.UserConfigDir on Linux, not macOS or Windows).
var userConfigDir = os.UserConfigDir

func statePath() (string, error) {
	dir, err := userConfigDir()
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
