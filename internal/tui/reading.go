package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const sidebarW = 34 // outer width of the chapter sidebar (no border)
const maxWrapW = 76 // readable cap on verse body wrap width

func (m Model) readingContentW() int {
	w := m.width
	if m.sidebarOpen {
		w -= sidebarW + 1 // +1 for the divider column
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
	outerW := m.readingContentW()
	vpH := m.readingBodyH() - 2 // minus header + separator rows
	if vpH < 0 {
		vpH = 0
	}
	vpW := outerW

	wrapW := vpW - 4 // small breathing margin either side of the centered text
	if wrapW > maxWrapW {
		wrapW = maxWrapW
	}
	if wrapW < 10 {
		wrapW = 10
	}

	if !m.vpOK {
		m.vp = viewport.New(vpW, vpH)
		m.vpOK = true
	} else {
		m.vp.Width = vpW
		m.vp.Height = vpH
	}
	m.vp.SetContent(placeVerse(m.verseBlock(wrapW), vpW, vpH))
	return m
}

func verticalDivider(h int) string {
	if h < 1 {
		h = 1
	}
	line := styleDivider.Render("│")
	lines := make([]string, h)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func (m Model) drawReading() string {
	bodyH := m.readingBodyH()
	content := m.drawReadingContent(m.readingContentW(), bodyH)
	if !m.sidebarOpen {
		return content
	}
	sidebar := m.drawSidebar(sidebarW, bodyH)
	divider := verticalDivider(bodyH)
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, divider, content)
}

func (m Model) drawSidebar(outerW, outerH int) string {
	iW := outerW
	iH := outerH

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
	if listH < 0 {
		listH = 0
	}
	body := renderList(rows, m.chapterCursor, listH, iW)

	header := colTitleStyle(true).Width(iW).Render(" G I T A")
	sep := colSepStyle.Width(iW).Render(strings.Repeat("─", iW))
	inner := header + "\n" + sep + "\n" + body

	return lipgloss.NewStyle().
		Width(iW).Height(iH).
		Background(cBgPanel).
		Render(inner)
}

// headerRow lays out a left-aligned crumb and a right-aligned badge on one
// colTitleStyle-colored line of width w, gap-filled like drawStatus.
func headerRow(left, right string, w int) string {
	lw, rw := lipgloss.Width(left), lipgloss.Width(right)
	gap := w - lw - rw
	if gap < 0 {
		gap = 0
	}
	line := left + strings.Repeat(" ", gap) + right
	return colTitleStyle(true).Width(w).Render(line)
}

func (m Model) drawReadingContent(outerW, outerH int) string {
	iW := outerW
	iH := outerH

	ch := m.currentChapter()
	v := m.currentVerse()

	var left, rightBadge string
	if ch != nil && v != nil {
		left = fmt.Sprintf(" %d.%d   %s", ch.Chapter, v.Verse, ch.Title)
		if m.isBookmarked(ch.Chapter, v.Verse) {
			rightBadge = styleMarkBadge.Render("◆ marked ")
		}
	} else {
		left = " Select a chapter"
	}

	header := headerRow(left, rightBadge, iW)
	sep := colSepStyle.Width(iW).Render(strings.Repeat("─", iW))

	body := header + "\n" + sep + "\n" + m.vp.View()

	return lipgloss.NewStyle().
		Width(iW).Height(iH).
		Background(cBgPanel).
		Render(body)
}

// verseBlock returns the raw (unplaced) styled verse column: chapter head,
// verse label, wrapped body, and a citation footer.
func (m Model) verseBlock(wrapW int) string {
	ch := m.currentChapter()
	v := m.currentVerse()
	if ch == nil || v == nil {
		return styleSep.Render("No verse selected.")
	}

	var b strings.Builder
	b.WriteString(styleChapHead.Render(fmt.Sprintf("Chapter %d: %s", ch.Chapter, ch.Title)))
	b.WriteString("\n\n")
	b.WriteString(styleVerseNum.Render(fmt.Sprintf("Verse %d", v.Verse)))
	b.WriteString("\n\n")
	b.WriteString(styleVerseBody.Render(wrap(v.Text, wrapW)))
	b.WriteString("\n\n")
	b.WriteString(verseFooter(ch.Chapter, v.Verse))
	return b.String()
}

// verseFooter renders a short decorative rule + citation, centered within a
// fixed footer width (independent of the panel width).
func verseFooter(chapter, verse int) string {
	const footerW = 24
	rule := styleCiteRule.Render(strings.Repeat("─", footerW))
	cite := styleCiteText.Render(fmt.Sprintf("‖ %d.%d ‖", chapter, verse))
	return rule + "\n" + lipgloss.PlaceHorizontal(footerW, lipgloss.Center, cite)
}

// maxTopPad caps how much blank space is inserted above the verse block on
// tall terminals — full vertical centering left too much empty screen.
const maxTopPad = 4

// placeVerse centers raw (a verse block) horizontally always, with a small
// capped top margin (not full vertical centering) when it fits within h;
// taller content falls back to top alignment so viewport scrolling works.
func placeVerse(raw string, w, h int) string {
	centered := lipgloss.PlaceHorizontal(w, lipgloss.Center, raw,
		lipgloss.WithWhitespaceBackground(cBgPanel))

	lines := strings.Count(raw, "\n") + 1
	if h <= 0 || lines > h {
		return centered
	}

	topPad := (h - lines) / 2
	if topPad > maxTopPad {
		topPad = maxTopPad
	}
	if topPad <= 0 {
		return centered
	}

	blank := lipgloss.NewStyle().Width(w).Background(cBgPanel).Render("")
	pad := strings.Repeat(blank+"\n", topPad)
	return pad + centered
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
