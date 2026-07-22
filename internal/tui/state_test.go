package tui

import (
	"testing"
	"time"
)

// setTestConfigDir redirects statePath() to dir for the duration of the
// test, bypassing os.UserConfigDir(). t.Setenv("XDG_CONFIG_HOME", ...)
// looks equivalent but only works on Linux — os.UserConfigDir() ignores
// that var entirely on macOS (always ~/Library/Application Support) and
// Windows (always %AppData%), so tests using it were silently unsandboxed
// on those platforms.
func setTestConfigDir(t *testing.T, dir string) {
	t.Helper()
	old := userConfigDir
	userConfigDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { userConfigDir = old })
}

func TestSaveLoadStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	setTestConfigDir(t, dir)

	want := persistedState{
		LastChapter: 2,
		LastVerse:   47,
		Bookmarks: []bookmark{
			{ChapterNum: 6, VerseNum: 5, SavedAt: time.Now().Truncate(time.Second)},
		},
	}
	if err := saveState(want); err != nil {
		t.Fatalf("saveState: %v", err)
	}

	got, err := loadState()
	if err != nil {
		t.Fatalf("loadState: %v", err)
	}
	if got.LastChapter != want.LastChapter || got.LastVerse != want.LastVerse {
		t.Errorf("LastChapter/LastVerse = %d/%d, want %d/%d",
			got.LastChapter, got.LastVerse, want.LastChapter, want.LastVerse)
	}
	if len(got.Bookmarks) != 1 ||
		got.Bookmarks[0].ChapterNum != 6 ||
		got.Bookmarks[0].VerseNum != 5 {
		t.Errorf("Bookmarks = %+v, want one bookmark at 6.5", got.Bookmarks)
	}
}

func TestLoadStateMissingFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	setTestConfigDir(t, dir)

	got, err := loadState()
	if err == nil {
		t.Fatal("expected an error for a missing state file")
	}
	if got.LastChapter != 0 || got.LastVerse != 0 || len(got.Bookmarks) != 0 {
		t.Errorf("got = %+v, want zero value", got)
	}
}
