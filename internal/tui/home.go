package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ACS-lessgo/gita-cli/internal/gita"
)

// dailyVerseIndex deterministically picks the same verse for every user on
// a given calendar day.
func dailyVerseIndex(all []gita.VerseRef, day time.Time) int {
	key := day.Year()*10000 + int(day.Month())*100 + day.Day()
	if key < 0 {
		key = -key
	}
	return key % len(all)
}

func (m Model) drawHome() string {
	h := m.height - 1
	if h < 3 {
		h = 3
	}

	contentW := 66
	if contentW > m.width-8 {
		contentW = m.width - 8
	}
	if contentW < 20 {
		contentW = 20
	}

	all := gita.AllVerses(m.g)
	dv := all[dailyVerseIndex(all, time.Now())]

	label := "Begin reading"
	if m.state.LastChapter != 0 {
		label = "Continue reading"
	}
	resumeRef, resumeTitle := "1.1", ""
	if ch := m.currentChapter(); ch != nil {
		if v := m.currentVerse(); v != nil {
			resumeRef = fmt.Sprintf("%d.%d", ch.Chapter, v.Verse)
			resumeTitle = ch.Title
		}
	}

	sep := styleSep.Render(strings.Repeat("─", contentW))

	var b strings.Builder
	b.WriteString(styleSep.Render("B H A G A V A D   G I T A"))
	b.WriteString("\n")
	b.WriteString(sep)
	b.WriteString("\n\n")
	b.WriteString(styleVerseNum.Render("VERSE OF THE DAY"))
	b.WriteString("\n\n")
	b.WriteString(styleVerseBody.Render(wrap(dv.Text, contentW)))
	b.WriteString("\n\n")
	b.WriteString(styleSep.Render(fmt.Sprintf("॥ %d.%d ॥  %s", dv.ChapterNum, dv.VerseNum, dv.ChapterTitle)))
	b.WriteString("\n\n")
	b.WriteString(sep)
	b.WriteString("\n\n")
	b.WriteString(styleSBKey.Render("enter") + styleSB.Render("  "+label+"   ") +
		styleSBInfo.Render(fmt.Sprintf("%s · %s", resumeRef, resumeTitle)))
	b.WriteString("\n\n")
	b.WriteString(styleSBKey.Render("p") + styleSB.Render(" jump anywhere    ") +
		styleSBKey.Render("/") + styleSB.Render(" search    ") +
		styleSBKey.Render("b") + styleSB.Render(" bookmarks    ") +
		styleSBKey.Render("?") + styleSB.Render(" help"))

	box := lipgloss.NewStyle().Width(contentW).Background(cBg).Foreground(cBright).Render(b.String())
	return lipgloss.Place(m.width, h, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceBackground(cBg))
}

func (m Model) handleHomeKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	if km.String() == "enter" {
		m = m.enterReading()
	}
	return m, nil
}
