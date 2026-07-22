package tui

import (
	"testing"
	"time"

	"github.com/ACS-lessgo/gita-cli/internal/gita"
)

func TestDailyVerseIndexDeterministicAndInRange(t *testing.T) {
	all := make([]gita.VerseRef, 700)
	for i := range all {
		all[i] = gita.VerseRef{ChapterNum: 1, VerseNum: i + 1}
	}

	day := time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)
	i1 := dailyVerseIndex(all, day)
	i2 := dailyVerseIndex(all, day)
	if i1 != i2 {
		t.Errorf("dailyVerseIndex not deterministic for the same day: %d != %d", i1, i2)
	}
	if i1 < 0 || i1 >= len(all) {
		t.Errorf("dailyVerseIndex out of range: %d (len %d)", i1, len(all))
	}
}
