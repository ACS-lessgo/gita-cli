package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const sidebarW = 34 // outer width of the chapter sidebar, border included

func (m Model) readingContentW() int {
	w := m.width
	if m.sidebarOpen {
		w -= sidebarW
	}
	if w < 20 {
		w = 20
	}
	return w
}

func (m Model) readingBodyH() int {
	h := m.height - 1 // 1 row reserved for the status bar
	if h < 3 {
		h = 3
	}
	return h
}

func (m Model) syncVP() Model {
	innerH := panelInnerH(m.readingBodyH()) - 2 // minus header + separator rows
	vpW := panelInnerW(m.readingContentW())
	if !m.vpOK {
		m.vp = viewport.New(vpW, innerH)
		m.vpOK = true
	} else {
		m.vp.Width = vpW
		m.vp.Height = innerH
	}
	m.vp.SetContent(m.buildContent())
	return m
}

func (m Model) drawReading() string {
	bodyH := m.readingBodyH()
	content := m.drawReadingContent(m.readingContentW(), bodyH)
	if !m.sidebarOpen {
		return content
	}
	sidebar := m.drawSidebar(sidebarW, bodyH)
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
}

func (m Model) drawSidebar(outerW, outerH int) string {
	iW := panelInnerW(outerW)
	iH := panelInnerH(outerH)

	rows := make([]string, len(m.g.Chapters))
	for i, ch := range m.g.Chapters {
		active := i == m.chapterCursor
		prefix := fmt.Sprintf(" %02d  ", ch.Chapter)
		count := fmt.Sprintf("%3d", len(ch.Verses))
		available := iW - len([]rune(prefix)) - len([]rune(count)) - 1
		if available < 1 {
			available = 1
		}
		title := runesTrunc(ch.Title, available)
		label := prefix + title
		pad := iW - len([]rune(label)) - len([]rune(count))
		if pad < 1 {
			pad = 1
		}
		rows[i] = chapRowStyle(active).Width(iW).Render(label + strings.Repeat(" ", pad) + count)
	}

	listH := iH - 2 // header + separator rows
	body := renderList(rows, m.chapterCursor, listH, iW)

	header := colTitleStyle(true).Width(iW).Render(" G I T A")
	sep := colSepStyle.Width(iW).Render(strings.Repeat("─", iW))
	inner := header + "\n" + sep + "\n" + body

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderCol(true)).
		Width(iW).Height(iH).
		Background(cBgPanel).
		Render(inner)
}

func (m Model) drawReadingContent(outerW, outerH int) string {
	iW := panelInnerW(outerW)
	iH := panelInnerH(outerH)

	ch := m.currentChapter()
	v := m.currentVerse()

	var crumb string
	if ch != nil && v != nil {
		mark := ""
		if m.isBookmarked(ch.Chapter, v.Verse) {
			mark = "  ◆ marked"
		}
		crumb = fmt.Sprintf(" %d.%d   %s%s", ch.Chapter, v.Verse, ch.Title, mark)
	} else {
		crumb = " Select a chapter"
	}

	header := colTitleStyle(true).Width(iW).Render(crumb)
	sep := colSepStyle.Width(iW).Render(strings.Repeat("─", iW))

	m.vp.Width = iW
	m.vp.Height = iH - 2
	body := header + "\n" + sep + "\n" + m.vp.View()

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderCol(true)).
		Width(iW).Height(iH).
		Background(cBgPanel).
		Render(body)
}

func (m Model) buildContent() string {
	ch := m.currentChapter()
	v := m.currentVerse()
	if ch == nil || v == nil {
		return styleSep.Render("No verse selected.")
	}

	wrapW := panelInnerW(m.readingContentW()) - 2
	if wrapW < 10 {
		wrapW = 10
	}

	var b strings.Builder
	b.WriteString(styleChapHead.Render(fmt.Sprintf("Chapter %d: %s", ch.Chapter, ch.Title)))
	b.WriteString("\n\n")
	b.WriteString(styleVerseNum.Render(fmt.Sprintf("Verse %d", v.Verse)))
	b.WriteString("\n\n")
	b.WriteString(styleVerseBody.Render(wrap(v.Text, wrapW)))
	b.WriteString("\n")
	return b.String()
}

func (m Model) handleReadingKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch km.String() {
	case "h", "left":
		if m.chapterCursor > 0 {
			m.chapterCursor--
			m.verseCursor = 0
			m = m.syncVP()
		}
	case "l", "right":
		if m.chapterCursor < len(m.g.Chapters)-1 {
			m.chapterCursor++
			m.verseCursor = 0
			m = m.syncVP()
		}
	case "j", "down":
		if ch := m.currentChapter(); ch != nil && m.verseCursor < len(ch.Verses)-1 {
			m.verseCursor++
			m = m.syncVP()
		}
	case "k", "up":
		if m.verseCursor > 0 {
			m.verseCursor--
			m = m.syncVP()
		}
	case "e":
		m.sidebarOpen = !m.sidebarOpen
		m = m.syncVP()
	case "m":
		m = m.toggleBookmark()
	case "g":
		m.verseCursor = 0
		m = m.syncVP()
	case "G":
		if ch := m.currentChapter(); ch != nil {
			m.verseCursor = len(ch.Verses) - 1
		}
		m = m.syncVP()
	case "pgup", "u":
		if m.vpOK {
			m.vp.HalfViewUp()
		}
	case "pgdown", "d":
		if m.vpOK {
			m.vp.HalfViewDown()
		}
	}
	return m, nil
}
