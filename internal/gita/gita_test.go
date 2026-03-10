package gita_test

import (
	"testing"

	"github.com/whoisyurii/gita-cli/internal/gita"
)

// testGita returns a minimal in-memory Gita for testing (bypasses embed).
func testGita() *gita.Gita {
	return &gita.Gita{
		Chapters: []gita.Chapter{
			{
				Chapter: 2,
				Title:   "The Yoga of Knowledge",
				Verses: []gita.Verse{
					{Verse: 47, Text: "You have a right to perform your prescribed duty, but you are not entitled to the fruits of action."},
					{Verse: 48, Text: "Be steadfast in yoga, O Arjuna."},
				},
			},
			{
				Chapter: 3,
				Title:   "The Yoga of Action",
				Verses: []gita.Verse{
					{Verse: 8, Text: "Perform your prescribed duty, for doing so is better than not working."},
					{Verse: 19, Text: "Therefore, without being attached to the fruits of activities, one should act as a matter of duty."},
				},
			},
		},
	}
}

// ── GetChapter ────────────────────────────────────────────────────────────────

func TestGetChapter_Found(t *testing.T) {
	g := testGita()
	ch, err := gita.GetChapter(g, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.Chapter != 2 {
		t.Errorf("expected chapter 2, got %d", ch.Chapter)
	}
	if ch.Title != "The Yoga of Knowledge" {
		t.Errorf("unexpected title: %q", ch.Title)
	}
}

func TestGetChapter_NotFound(t *testing.T) {
	g := testGita()
	_, err := gita.GetChapter(g, 99)
	if err == nil {
		t.Fatal("expected error for missing chapter, got nil")
	}
}

// ── GetVerse ──────────────────────────────────────────────────────────────────

func TestGetVerse_Found(t *testing.T) {
	g := testGita()
	ref, err := gita.GetVerse(g, 2, 47)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ChapterNum != 2 || ref.VerseNum != 47 {
		t.Errorf("got chapter=%d verse=%d, want 2/47", ref.ChapterNum, ref.VerseNum)
	}
	if ref.Text == "" {
		t.Error("expected non-empty text")
	}
}

func TestGetVerse_ChapterNotFound(t *testing.T) {
	g := testGita()
	_, err := gita.GetVerse(g, 18, 66)
	if err == nil {
		t.Fatal("expected error for missing chapter, got nil")
	}
}

func TestGetVerse_VerseNotFound(t *testing.T) {
	g := testGita()
	_, err := gita.GetVerse(g, 2, 999)
	if err == nil {
		t.Fatal("expected error for missing verse, got nil")
	}
}

// ── AllVerses ─────────────────────────────────────────────────────────────────

func TestAllVerses_Count(t *testing.T) {
	g := testGita()
	refs := gita.AllVerses(g)
	if len(refs) != 4 {
		t.Errorf("expected 4 verses, got %d", len(refs))
	}
}

func TestAllVerses_Empty(t *testing.T) {
	g := &gita.Gita{}
	refs := gita.AllVerses(g)
	if len(refs) != 0 {
		t.Errorf("expected 0 verses, got %d", len(refs))
	}
}

// ── Search ────────────────────────────────────────────────────────────────────

func TestSearch_Found(t *testing.T) {
	g := testGita()
	results := gita.Search(g, "duty", 0)
	if len(results) == 0 {
		t.Fatal("expected results for keyword 'duty', got none")
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	g := testGita()
	lower := gita.Search(g, "duty", 0)
	upper := gita.Search(g, "DUTY", 0)
	if len(lower) != len(upper) {
		t.Errorf("case sensitivity issue: lower=%d upper=%d", len(lower), len(upper))
	}
}

func TestSearch_NotFound(t *testing.T) {
	g := testGita()
	results := gita.Search(g, "zzznomatch", 0)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_Limit(t *testing.T) {
	g := testGita()
	results := gita.Search(g, "duty", 1)
	if len(results) > 1 {
		t.Errorf("expected max 1 result with limit=1, got %d", len(results))
	}
}

func TestSearch_NoLimit(t *testing.T) {
	g := testGita()
	// "duty" appears in multiple verses
	limited := gita.Search(g, "duty", 1)
	unlimited := gita.Search(g, "duty", 0)
	if len(unlimited) < len(limited) {
		t.Error("unlimited search should return >= results than limited search")
	}
}

// ── VerseRef fields ───────────────────────────────────────────────────────────

func TestGetVerse_ChapterTitle(t *testing.T) {
	g := testGita()
	ref, err := gita.GetVerse(g, 3, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ChapterTitle != "The Yoga of Action" {
		t.Errorf("unexpected chapter title: %q", ref.ChapterTitle)
	}
}
