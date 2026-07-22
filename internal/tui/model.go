package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ACS-lessgo/gita-cli/internal/gita"
)

// ── Screens & overlays ────────────────────────────────────────────────────

type screen int

const (
	screenHome screen = iota
	screenReading
	screenSearch
	screenBookmarks
)

type overlay int

const (
	overlayNone overlay = iota
	overlayPalette
	overlayHelp
)

// paletteResult is a single jump-palette suggestion (chapter or verse hit).
// Declared here because it's part of Model's state shape; palette.go owns
// the logic that produces and consumes it.
type paletteResult struct {
	kind       string // "ch" or "v"
	chapterNum int
	verseNum   int // 0 for a chapter-kind result
	label      string
}

// ── Model ───────────────────────────────────────────────────────────────

type Model struct {
	g          *gita.Gita
	splash     SplashModel
	showSplash bool

	width, height int

	screen  screen
	overlay overlay

	state persistedState

	// reading
	chapterCursor int
	verseCursor   int
	sidebarOpen   bool
	vp            viewport.Model
	vpOK          bool

	// search
	searchQuery  string
	searchHits   []gita.SearchResult
	searchCursor int

	// bookmarks
	bookmarkCursor int

	// palette
	paletteQuery   string
	paletteResults []paletteResult
	paletteCursor  int

	status string
}

func New(g *gita.Gita) Model {
	st, _ := loadState() // missing/corrupt file: start from zero-value state
	m := Model{
		g:           g,
		splash:      SplashModel{},
		showSplash:  true,
		state:       st,
		sidebarOpen: true,
		screen:      screenHome,
	}
	if st.LastChapter != 0 {
		m.chapterCursor = m.chapterIndex(st.LastChapter)
		m.verseCursor = verseIndex(&m.g.Chapters[m.chapterCursor], st.LastVerse)
	}
	return m
}

// ── Geometry helpers ───────────────────────────────────────────────────────

// panelInnerW returns the drawable width inside a bordered box of outer
// width ow (NormalBorder adds 1 char each side).
func panelInnerW(ow int) int {
	if ow > 2 {
		return ow - 2
	}
	return 1
}

// panelInnerH returns the drawable height inside a bordered box of outer
// height oh.
func panelInnerH(oh int) int {
	if oh > 2 {
		return oh - 2
	}
	return 1
}

// ── tea.Model ──────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd { return m.splash.Init() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = ws.Width, ws.Height
		m.splash.width, m.splash.height = ws.Width, ws.Height
		if !m.showSplash {
			m = m.syncVP()
		}
		return m, nil
	}

	if m.showSplash {
		var cmd tea.Cmd
		m.splash, cmd = m.splash.Update(msg)
		if m.splash.done {
			m.showSplash = false
			m = m.syncVP()
			return m, nil
		}
		return m, cmd
	}

	if _, ok := msg.(doneSplashMsg); ok {
		m.showSplash = false
		m = m.syncVP()
		return m, nil
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		return m.handleKey(km)
	}

	if m.screen == screenReading && m.overlay == overlayNone && m.vpOK {
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	if m.showSplash {
		return m.splash.View()
	}

	switch m.overlay {
	case overlayPalette:
		return lipgloss.JoinVertical(lipgloss.Left, m.drawPalette(), m.drawStatus())
	case overlayHelp:
		return lipgloss.JoinVertical(lipgloss.Left, m.drawHelp(), m.drawStatus())
	}

	var body string
	switch m.screen {
	case screenHome:
		body = m.drawHome()
	case screenReading:
		body = m.drawReading()
	case screenSearch:
		body = m.drawSearch()
	case screenBookmarks:
		body = m.drawBookmarks()
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, m.drawStatus())
}

// ── Key dispatch ───────────────────────────────────────────────────────────

func (m Model) handleKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.overlay == overlayPalette {
		return m.handlePaletteKey(km)
	}
	if m.overlay == overlayHelp {
		switch km.String() {
		case "esc", "?":
			m.overlay = overlayNone
		}
		return m, nil
	}
	if m.screen == screenSearch {
		return m.handleSearchKey(km)
	}

	switch km.String() {
	case "ctrl+c", "q":
		m = m.persistState()
		return m, tea.Quit
	case "1":
		m.screen = screenHome
		m.status = ""
		return m, nil
	case "2":
		return m.enterReading(), nil
	case "3", "/":
		return m.enterSearch(), nil
	case "4", "b":
		m.screen = screenBookmarks
		m.bookmarkCursor = 0
		m.status = ""
		return m, nil
	case "p":
		m.overlay = overlayPalette
		m.paletteQuery = ""
		m.paletteResults = nil
		m.paletteCursor = 0
		return m, nil
	case "?":
		m.overlay = overlayHelp
		return m, nil
	case "esc":
		return m.enterReading(), nil
	}

	switch m.screen {
	case screenHome:
		return m.handleHomeKey(km)
	case screenReading:
		return m.handleReadingKey(km)
	case screenBookmarks:
		return m.handleBookmarksKey(km)
	}
	return m, nil
}

func (m Model) enterReading() Model {
	m.screen = screenReading
	m.overlay = overlayNone
	m.status = ""
	return m.syncVP()
}

// ── Persistence ────────────────────────────────────────────────────────────

// persistState snapshots the current reading position into m.state and
// writes it to disk. Save failures are surfaced as a status message, never
// a crash.
func (m Model) persistState() Model {
	if ch := m.currentChapter(); ch != nil {
		m.state.LastChapter = ch.Chapter
		if v := m.currentVerse(); v != nil {
			m.state.LastVerse = v.Verse
		}
	}
	if err := saveState(m.state); err != nil {
		m.status = "Warning: could not save state"
	}
	return m
}

// ── Status bar ─────────────────────────────────────────────────────────────

func (m Model) drawStatus() string {
	var left string
	if m.status != "" {
		left = styleSBInfo.Render(" " + m.status)
	} else {
		left = m.statusHints()
	}
	right := m.statusRight()

	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	gap := m.width - lw - rw
	if gap < 1 {
		gap = 1
	}
	fill := styleSB.Render(strings.Repeat(" ", gap))

	line := left + fill + right
	if lipgloss.Width(line) > m.width {
		line = left + fill
	}
	return styleSB.Render(line)
}

func (m Model) statusHints() string {
	sp := styleSBSep.Render(" │ ")
	hint := func(k, d string) string {
		return styleSBKey.Render(k) + styleSB.Render(" "+d)
	}
	switch m.screen {
	case screenHome:
		return " " + hint("enter", "continue") + sp + hint("p", "jump") + sp +
			hint("/", "search") + sp + hint("b", "bookmarks") + sp +
			hint("?", "help") + sp + hint("q", "quit")
	case screenReading:
		return " " + hint("j k", "verse") + sp + hint("h l", "chapter") + sp +
			hint("e", "sidebar") + sp + hint("m", "mark") + sp +
			hint("p", "jump") + sp + hint("/", "search") + sp + hint("q", "quit")
	case screenSearch:
		return " " + hint("↑ ↓", "results") + sp + hint("enter", "open") + sp +
			hint("esc", "back")
	case screenBookmarks:
		return " " + hint("↑ ↓", "select") + sp + hint("enter", "open") + sp +
			hint("x", "remove") + sp + hint("esc", "back")
	}
	return ""
}

func (m Model) statusRight() string {
	switch m.screen {
	case screenReading:
		if ch := m.currentChapter(); ch != nil {
			if v := m.currentVerse(); v != nil {
				pct := 0
				if m.vpOK {
					pct = int(m.vp.ScrollPercent() * 100)
				}
				return styleSBInfo.Render(fmt.Sprintf(" Ch %d  v%d  %d%% ", ch.Chapter, v.Verse, pct))
			}
		}
	case screenSearch:
		return styleSBInfo.Render(fmt.Sprintf(" %d match(es) ", len(m.searchHits)))
	case screenBookmarks:
		return styleSBInfo.Render(fmt.Sprintf(" %d saved ", len(m.state.Bookmarks)))
	}
	return ""
}

// ── Data helpers ───────────────────────────────────────────────────────────

func (m Model) currentChapter() *gita.Chapter {
	if m.chapterCursor < 0 || m.chapterCursor >= len(m.g.Chapters) {
		return nil
	}
	return &m.g.Chapters[m.chapterCursor]
}

func (m Model) currentVerse() *gita.Verse {
	ch := m.currentChapter()
	if ch == nil || len(ch.Verses) == 0 {
		return nil
	}
	i := m.verseCursor
	if i < 0 || i >= len(ch.Verses) {
		i = 0
	}
	return &ch.Verses[i]
}

func (m Model) chapterIndex(chapterNum int) int {
	for i, ch := range m.g.Chapters {
		if ch.Chapter == chapterNum {
			return i
		}
	}
	return 0
}

func verseIndex(ch *gita.Chapter, verseNum int) int {
	for i, v := range ch.Verses {
		if v.Verse == verseNum {
			return i
		}
	}
	return 0
}

// ── Shared list/text helpers ────────────────────────────────────────────────

// renderList returns a string of exactly `h` lines wide `w`, scrolling to
// keep `cursor` visible and centred.
func renderList(rows []string, cursor, h, w int) string {
	if len(rows) == 0 {
		blank := lipgloss.NewStyle().Width(w).Background(cBgPanel).Render("")
		return strings.Join(repeatStr(blank, h), "\n")
	}

	start := cursor - h/2
	if start < 0 {
		start = 0
	}
	end := start + h
	if end > len(rows) {
		end = len(rows)
		start = end - h
		if start < 0 {
			start = 0
		}
	}
	visible := rows[start:end]

	blank := lipgloss.NewStyle().Width(w).Background(cBgPanel).Render("")
	for len(visible) < h {
		visible = append(visible, blank)
	}

	return strings.Join(visible, "\n")
}

func repeatStr(s string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = s
	}
	return out
}

// wrap word-wraps text to width characters per line.
func wrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}
	var lines []string
	cur := words[0]
	for _, w := range words[1:] {
		if lipgloss.Width(cur)+1+lipgloss.Width(w) > width {
			lines = append(lines, cur)
			cur = w
		} else {
			cur += " " + w
		}
	}
	return strings.Join(append(lines, cur), "\n")
}

// runesTrunc truncates s to at most n visible runes, appending "…" if cut.
func runesTrunc(s string, n int) string {
	if n <= 0 {
		return ""
	}
	rr := []rune(s)
	if len(rr) <= n {
		return s
	}
	if n == 1 {
		return "…"
	}
	return string(rr[:n-1]) + "…"
}

// hlText case-insensitively wraps every occurrence of kw in s with the
// search-highlight style.
func hlText(s, kw string) string {
	if kw == "" {
		return s
	}
	lower := strings.ToLower(s)
	lkw := strings.ToLower(kw)
	var b strings.Builder
	rem, lrem := s, lower
	for {
		i := strings.Index(lrem, lkw)
		if i < 0 {
			b.WriteString(rem)
			break
		}
		b.WriteString(rem[:i])
		b.WriteString(styleSearchHL.Render(rem[i : i+len(kw)]))
		rem = rem[i+len(kw):]
		lrem = lrem[i+len(kw):]
	}
	return b.String()
}
