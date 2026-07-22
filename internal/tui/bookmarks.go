package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) findBookmark(chapterNum, verseNum int) int {
	for i, b := range m.state.Bookmarks {
		if b.ChapterNum == chapterNum && b.VerseNum == verseNum {
			return i
		}
	}
	return -1
}

func (m Model) isBookmarked(chapterNum, verseNum int) bool {
	return m.findBookmark(chapterNum, verseNum) >= 0
}

// toggleBookmark adds or removes a bookmark for the current verse and
// persists immediately, so a killed process doesn't lose it.
func (m Model) toggleBookmark() Model {
	ch := m.currentChapter()
	v := m.currentVerse()
	if ch == nil || v == nil {
		return m
	}
	if idx := m.findBookmark(ch.Chapter, v.Verse); idx >= 0 {
		m.state.Bookmarks = append(m.state.Bookmarks[:idx], m.state.Bookmarks[idx+1:]...)
		m.status = "Bookmark removed"
	} else {
		bm := bookmark{ChapterNum: ch.Chapter, VerseNum: v.Verse, SavedAt: time.Now()}
		m.state.Bookmarks = append([]bookmark{bm}, m.state.Bookmarks...) // newest first
		m.status = "Bookmarked"
	}
	return m.persistState()
}

func relTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	}
}

func (m Model) handleBookmarksKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch km.String() {
	case "up", "k":
		if m.bookmarkCursor > 0 {
			m.bookmarkCursor--
		}
	case "down", "j":
		if m.bookmarkCursor < len(m.state.Bookmarks)-1 {
			m.bookmarkCursor++
		}
	case "enter":
		if m.bookmarkCursor >= 0 && m.bookmarkCursor < len(m.state.Bookmarks) {
			bm := m.state.Bookmarks[m.bookmarkCursor]
			m.chapterCursor = m.chapterIndex(bm.ChapterNum)
			m.verseCursor = verseIndex(&m.g.Chapters[m.chapterCursor], bm.VerseNum)
			m = m.enterReading()
		}
	case "x":
		if m.bookmarkCursor >= 0 && m.bookmarkCursor < len(m.state.Bookmarks) {
			m.state.Bookmarks = append(
				m.state.Bookmarks[:m.bookmarkCursor],
				m.state.Bookmarks[m.bookmarkCursor+1:]...,
			)
			if m.bookmarkCursor >= len(m.state.Bookmarks) && m.bookmarkCursor > 0 {
				m.bookmarkCursor--
			}
			m = m.persistState()
		}
	}
	return m, nil
}

func (m Model) drawBookmarks() string {
	h := m.height - 1
	if h < 3 {
		h = 3
	}

	contentW := 76
	if contentW > m.width-8 {
		contentW = m.width - 8
	}
	if contentW < 20 {
		contentW = 20
	}

	var b strings.Builder
	b.WriteString(styleVerseNum.Render(fmt.Sprintf("BOOKMARKS   %d saved", len(m.state.Bookmarks))))
	b.WriteString("\n\n")

	if len(m.state.Bookmarks) == 0 {
		b.WriteString(styleSep.Render("No bookmarks yet — press m while reading a verse to save one."))
	} else {
		for i, bm := range m.state.Bookmarks {
			idx := m.chapterIndex(bm.ChapterNum)
			title := m.g.Chapters[idx].Title
			row := fmt.Sprintf(" %d.%-3d  %-28s  %s",
				bm.ChapterNum, bm.VerseNum, runesTrunc(title, 28), relTime(bm.SavedAt))
			b.WriteString(chapRowStyle(i == m.bookmarkCursor).Width(contentW).Render(row))
			b.WriteString("\n")
		}
	}

	box := lipgloss.NewStyle().Width(contentW).Background(cBg).Render(b.String())
	return lipgloss.Place(m.width, h, lipgloss.Center, lipgloss.Top, box,
		lipgloss.WithWhitespaceBackground(cBg))
}
