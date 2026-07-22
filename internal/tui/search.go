package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ACS-lessgo/gita-cli/internal/gita"
)

func (m Model) enterSearch() Model {
	m.screen = screenSearch
	m.overlay = overlayNone
	m.status = ""
	return m
}

func (m Model) refreshSearch() Model {
	if m.searchQuery == "" {
		m.searchHits = nil
	} else {
		m.searchHits = gita.Search(m.g, m.searchQuery, 50)
	}
	m.searchCursor = 0
	return m
}

func (m Model) handleSearchKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch km.String() {
	case "esc":
		m = m.enterReading()
	case "enter":
		if m.searchCursor >= 0 && m.searchCursor < len(m.searchHits) {
			hit := m.searchHits[m.searchCursor].Ref
			m.chapterCursor = m.chapterIndex(hit.ChapterNum)
			m.verseCursor = verseIndex(&m.g.Chapters[m.chapterCursor], hit.VerseNum)
			m = m.enterReading()
		}
	case "up":
		if m.searchCursor > 0 {
			m.searchCursor--
		}
	case "down":
		if m.searchCursor < len(m.searchHits)-1 {
			m.searchCursor++
		}
	case "backspace":
		if len(m.searchQuery) > 0 {
			rr := []rune(m.searchQuery)
			m.searchQuery = string(rr[:len(rr)-1])
			m = m.refreshSearch()
		}
	default:
		if len(km.Runes) == 1 {
			m.searchQuery += string(km.Runes)
			m = m.refreshSearch()
		}
	}
	return m, nil
}

func (m Model) drawSearch() string {
	bodyH := m.height - 2 // query bar row + status bar row
	if bodyH < 3 {
		bodyH = 3
	}

	queryBar := m.drawSearchQueryBar()

	listW := m.width * 52 / 100
	if listW < 20 {
		listW = 20
	}
	previewW := m.width - listW

	list := m.drawSearchResults(listW, bodyH)
	preview := m.drawSearchPreview(previewW, bodyH)

	body := lipgloss.JoinHorizontal(lipgloss.Top, list, preview)
	return lipgloss.JoinVertical(lipgloss.Left, queryBar, body)
}

func (m Model) drawSearchQueryBar() string {
	count := fmt.Sprintf("%d matches", len(m.searchHits))
	if m.searchQuery == "" {
		count = "type to search"
	}
	left := styleSBSrch.Render(fmt.Sprintf(" / %s▌", m.searchQuery))
	right := styleSBInfo.Render(count + " ")
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return styleSB.Render(left + strings.Repeat(" ", gap) + right)
}

func (m Model) drawSearchResults(outerW, outerH int) string {
	iW := panelInnerW(outerW)
	iH := panelInnerH(outerH)

	rows := make([]string, len(m.searchHits))
	for i, hit := range m.searchHits {
		active := i == m.searchCursor
		ref := fmt.Sprintf(" %d.%-3d", hit.Ref.ChapterNum, hit.Ref.VerseNum)
		available := iW - len([]rune(ref)) - 1
		if available < 1 {
			available = 1
		}
		snippet := runesTrunc(hit.Ref.Text, available)
		rows[i] = chapRowStyle(active).Width(iW).Render(ref + " " + snippet)
	}

	body := renderList(rows, m.searchCursor, iH, iW)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderCol(true)).
		Width(iW).Height(iH).
		Background(cBgPanel).
		Render(body)
}

func (m Model) drawSearchPreview(outerW, outerH int) string {
	iW := panelInnerW(outerW)
	iH := panelInnerH(outerH)

	var body string
	if m.searchCursor >= 0 && m.searchCursor < len(m.searchHits) {
		hit := m.searchHits[m.searchCursor]
		wrapW := iW - 2
		if wrapW < 10 {
			wrapW = 10
		}
		header := colTitleStyle(true).Width(iW).Render(" " + hit.Ref.ChapterTitle)
		sep := colSepStyle.Width(iW).Render(strings.Repeat("─", iW))
		text := styleVerseBody.Render(wrap(hlText(hit.Ref.Text, m.searchQuery), wrapW))
		body = header + "\n" + sep + "\n\n" + text
	} else {
		body = styleSep.Render(" No results")
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderCol(false)).
		Width(iW).Height(iH).
		Background(cBgPanel).
		Render(body)
}
