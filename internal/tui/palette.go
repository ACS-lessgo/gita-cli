package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ACS-lessgo/gita-cli/internal/gita"
)

const paletteMaxResults = 8

type chapterVerseRef struct {
	chapterNum, verseNum int
}

// parseChapterVerse parses "N" or "N.M" into a chapter (and optionally
// verse) number. Returns ok=false for anything else, including empty input.
func parseChapterVerse(q string) (chapterVerseRef, bool) {
	parts := strings.SplitN(q, ".", 2)
	chapterNum, err := strconv.Atoi(parts[0])
	if err != nil || chapterNum < 1 {
		return chapterVerseRef{}, false
	}
	if len(parts) == 1 {
		return chapterVerseRef{chapterNum: chapterNum}, true
	}
	verseNum, err := strconv.Atoi(parts[1])
	if err != nil || verseNum < 1 {
		return chapterVerseRef{}, false
	}
	return chapterVerseRef{chapterNum: chapterNum, verseNum: verseNum}, true
}

func (m Model) computePaletteResults(query string) []paletteResult {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil
	}

	if ref, ok := parseChapterVerse(q); ok {
		if ref.verseNum > 0 {
			return []paletteResult{{
				kind: "v", chapterNum: ref.chapterNum, verseNum: ref.verseNum,
				label: fmt.Sprintf("Jump to %d.%d", ref.chapterNum, ref.verseNum),
			}}
		}
		return []paletteResult{{
			kind: "ch", chapterNum: ref.chapterNum,
			label: fmt.Sprintf("Jump to chapter %d", ref.chapterNum),
		}}
	}

	var out []paletteResult
	lower := strings.ToLower(q)
	for _, ch := range m.g.Chapters {
		if strings.Contains(strings.ToLower(ch.Title), lower) {
			out = append(out, paletteResult{kind: "ch", chapterNum: ch.Chapter, label: ch.Title})
			if len(out) >= paletteMaxResults {
				return out
			}
		}
	}
	for _, hit := range gita.Search(m.g, q, paletteMaxResults-len(out)) {
		out = append(out, paletteResult{
			kind: "v", chapterNum: hit.Ref.ChapterNum, verseNum: hit.Ref.VerseNum,
			label: runesTrunc(hit.Ref.Text, 60),
		})
	}
	return out
}

func (m Model) handlePaletteKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch km.String() {
	case "esc":
		m.overlay = overlayNone
	case "enter":
		if m.paletteCursor >= 0 && m.paletteCursor < len(m.paletteResults) {
			r := m.paletteResults[m.paletteCursor]
			m.chapterCursor = m.chapterIndex(r.chapterNum)
			if r.kind == "v" {
				m.verseCursor = verseIndex(&m.g.Chapters[m.chapterCursor], r.verseNum)
			} else {
				m.verseCursor = 0
			}
			m = m.enterReading()
		}
	case "up":
		if m.paletteCursor > 0 {
			m.paletteCursor--
		}
	case "down":
		if m.paletteCursor < len(m.paletteResults)-1 {
			m.paletteCursor++
		}
	case "backspace":
		if len(m.paletteQuery) > 0 {
			rr := []rune(m.paletteQuery)
			m.paletteQuery = string(rr[:len(rr)-1])
			m.paletteResults = m.computePaletteResults(m.paletteQuery)
			m.paletteCursor = 0
		}
	default:
		if len(km.Runes) == 1 {
			m.paletteQuery += string(km.Runes)
			m.paletteResults = m.computePaletteResults(m.paletteQuery)
			m.paletteCursor = 0
		}
	}
	return m, nil
}

func (m Model) drawPalette() string {
	w := m.width * 64 / 100
	if w > 80 {
		w = 80
	}
	if w < 30 {
		w = 30
	}
	if w > m.width-4 {
		w = m.width - 4
	}
	innerW := w - 2

	var b strings.Builder
	b.WriteString(styleSBSrch.Render(fmt.Sprintf(" › %s▌", m.paletteQuery)))
	b.WriteString("\n")
	b.WriteString(colSepStyle.Width(innerW).Render(strings.Repeat("─", innerW)))
	b.WriteString("\n")

	if len(m.paletteResults) == 0 {
		b.WriteString(styleSep.Render(" chapter.verse (e.g. 2.47) or type words"))
	} else {
		for i, r := range m.paletteResults {
			row := fmt.Sprintf(" %-2s %s", r.kind, r.label)
			b.WriteString(chapRowStyle(i == m.paletteCursor).Width(innerW).Render(row))
			b.WriteString("\n")
		}
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderCol(true)).
		Width(innerW).
		Background(cBgPanel).
		Render(strings.TrimRight(b.String(), "\n"))

	h := m.height - 1
	if h < 3 {
		h = 3
	}
	return lipgloss.Place(m.width, h, lipgloss.Center, lipgloss.Top, box,
		lipgloss.WithWhitespaceBackground(cBg))
}
