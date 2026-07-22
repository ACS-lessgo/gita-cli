package tui

import (
	"testing"

	"github.com/ACS-lessgo/gita-cli/internal/gita"
)

func testGita() *gita.Gita {
	return &gita.Gita{Chapters: []gita.Chapter{
		{Chapter: 2, Title: "Transcendental Knowledge", Verses: []gita.Verse{
			{Verse: 47, Text: "You have a right to your actions, but never to the fruits of your actions."},
		}},
	}}
}

func TestToggleBookmarkAddAndRemove(t *testing.T) {
	dir := t.TempDir()
	setTestConfigDir(t, dir)

	m := New(testGita())

	m = m.toggleBookmark()
	if len(m.state.Bookmarks) != 1 {
		t.Fatalf("after add: len(Bookmarks) = %d, want 1", len(m.state.Bookmarks))
	}
	if m.state.Bookmarks[0].ChapterNum != 2 || m.state.Bookmarks[0].VerseNum != 47 {
		t.Errorf("bookmark = %+v, want chapter 2 verse 47", m.state.Bookmarks[0])
	}
	if !m.isBookmarked(2, 47) {
		t.Error("isBookmarked(2, 47) = false, want true after adding")
	}

	m = m.toggleBookmark()
	if len(m.state.Bookmarks) != 0 {
		t.Fatalf("after remove: len(Bookmarks) = %d, want 0", len(m.state.Bookmarks))
	}
	if m.isBookmarked(2, 47) {
		t.Error("isBookmarked(2, 47) = true, want false after removing")
	}
}
