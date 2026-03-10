package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/whoisyurii/gita-cli/internal/gita"
)

// ── Enums ─────────────────────────────────────────────────────────────────────

type panel int

const (
	panelChapters panel = iota
	panelVerses
	panelContent
)

type searchState int

const (
	searchOff searchState = iota
	searchTyping
	searchActive
)

// ── Fixed column widths (outer, includes border chars) ────────────────────────

const (
	chapW = 80 // chapter list panel outer width — wide enough for longest title
	vrsW  = 8  // verse list panel outer width
	// content panel gets: termWidth - chapW - vrsW
)

// ── Model ─────────────────────────────────────────────────────────────────────

type Model struct {
	g          *gita.Gita
	splash     SplashModel
	showSplash bool

	width, height int

	focus         panel
	chapterCursor int
	verseCursor   int

	vp   viewport.Model
	vpOK bool

	search      searchState
	searchQuery string
	searchHits  []gita.SearchResult
	searchIdx   int

	status string
}

func New(g *gita.Gita) Model {
	return Model{g: g, splash: SplashModel{}, showSplash: true}
}

// ── Geometry helpers ──────────────────────────────────────────────────────────

func (m Model) cntW() int {
	w := m.width - chapW - vrsW
	if w < 20 {
		w = 20
	}
	return w
}

// panelInnerW returns the drawable width inside a bordered panel of outer width ow.
// NormalBorder adds 1 char on each side = 2 total.
func panelInnerW(ow int) int {
	if ow > 2 {
		return ow - 2
	}
	return 1
}

// panelInnerH: border top + bottom = 2; title row = 1.
func (m Model) panelInnerH() int {
	// total height - 1 title row - 2 border rows - 1 status bar = height-4
	h := m.height - 4
	if h < 1 {
		h = 1
	}
	return h
}

func (m Model) syncVP() Model {
	vpW := panelInnerW(m.cntW())
	vpH := m.panelInnerH()
	if !m.vpOK {
		m.vp = viewport.New(vpW, vpH)
		m.vpOK = true
	} else {
		m.vp.Width = vpW
		m.vp.Height = vpH
	}
	m.vp.SetContent(m.buildContent())
	return m
}

// ── tea.Model ─────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd { return m.splash.Init() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ── Window resize — always handled ──
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = ws.Width, ws.Height
		m.splash.width, m.splash.height = ws.Width, ws.Height
		if !m.showSplash {
			m = m.syncVP()
		}
		return m, nil
	}

	// ── Splash phase ──
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

	// ── Deferred splash-done message ──
	if _, ok := msg.(doneSplashMsg); ok {
		m.showSplash = false
		m = m.syncVP()
		return m, nil
	}

	// ── Search input mode ──
	if m.search == searchTyping {
		if km, ok := msg.(tea.KeyMsg); ok {
			return m.handleSearchInput(km)
		}
	}

	// ── Normal key handling ──
	if km, ok := msg.(tea.KeyMsg); ok {
		return m.handleKey(km)
	}

	// ── Viewport scrolling (content panel focused) ──
	if m.focus == panelContent && m.vpOK {
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
	return m.draw()
}

// ── draw: compose the full screen ────────────────────────────────────────────

func (m Model) draw() string {
	// Title row height = 1, border = 2, status = 1 → panel outer height
	panelOuterH := m.height - 2 // -1 title -1 status
	if panelOuterH < 3 {
		panelOuterH = 3
	}

	chapPanel := m.drawChapPanel(panelOuterH)
	vrsPanel := m.drawVrsPanel(panelOuterH)
	cntPanel := m.drawCntPanel(panelOuterH)

	body := lipgloss.JoinHorizontal(lipgloss.Top, chapPanel, vrsPanel, cntPanel)
	return lipgloss.JoinVertical(lipgloss.Left, body, m.drawStatus())
}

// ── Chapter panel ─────────────────────────────────────────────────────────────

func (m Model) drawChapPanel(outerH int) string {
	active := m.focus == panelChapters
	iW := panelInnerW(chapW)
	iH := outerH - 2 // subtract border rows

	// Build rows — number + full title on one line, truncate only if needed
	rows := make([]string, len(m.g.Chapters))
	for i, ch := range m.g.Chapters {
		prefix := fmt.Sprintf(" Ch %-2d  ", ch.Chapter)
		available := iW - len([]rune(prefix))
		if available < 1 {
			available = 1
		}
		title := runesTrunc(ch.Title, available)
		label := prefix + title
		rows[i] = chapRowStyle(i == m.chapterCursor).Width(iW).Render(label)
	}

	listH := iH - 2 // title row + separator row inside box
	body := renderList(rows, m.chapterCursor, listH, iW)

	colTitle := colTitleStyle(active).Width(iW).Render(" #     Chapter")
	colSep := colSepStyle.Width(iW).Render(strings.Repeat("─", iW))
	inner := colTitle + "\n" + colSep + "\n" + body

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderCol(active)).
		Width(iW).Height(iH).
		Background(cBgPanel).
		Render(inner)

	return box
}

// ── Verse panel ───────────────────────────────────────────────────────────────

func (m Model) drawVrsPanel(outerH int) string {
	active := m.focus == panelVerses
	iW := panelInnerW(vrsW)
	iH := outerH - 2

	var rows []string
	if ch := m.currentChapter(); ch != nil {
		for i, v := range ch.Verses {
			label := fmt.Sprintf(" %d", v.Verse)
			rows = append(rows, vrsRowStyle(i == m.verseCursor).Width(iW).Render(label))
		}
	}

	listH := iH - 2
	body := renderList(rows, m.verseCursor, listH, iW)

	colTitle := colTitleStyle(active).Width(iW).Render(" v.")
	colSep := colSepStyle.Width(iW).Render(strings.Repeat("─", iW))
	inner := colTitle + "\n" + colSep + "\n" + body

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderCol(active)).
		Width(iW).Height(iH).
		Background(cBgPanel).
		Render(inner)

	return box
}

// ── Content panel ─────────────────────────────────────────────────────────────

func (m Model) drawCntPanel(outerH int) string {
	active := m.focus == panelContent
	cW := m.cntW()
	iW := panelInnerW(cW)
	iH := outerH - 2

	ch := m.currentChapter()
	v := m.currentVerse()

	var crumb string
	if ch != nil && v != nil {
		crumb = fmt.Sprintf(" Chapter %d: %s   Verse %d", ch.Chapter, ch.Title, v.Verse)
	} else {
		crumb = " Select a chapter and verse"
	}

	colTitle := colTitleStyle(active).Width(iW).Render(crumb)
	colSep := colSepStyle.Width(iW).Render(strings.Repeat("─", iW))

	// VP gets 2 fewer rows (title + sep)
	m.vp.Width = iW
	m.vp.Height = iH - 2

	vpContent := colTitle + "\n" + colSep + "\n" + m.vp.View()

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderCol(active)).
		Width(iW).Height(iH).
		Background(cBgPanel).
		Render(vpContent)

	return box
}

// ── Content text builder ──────────────────────────────────────────────────────

func (m Model) buildContent() string {
	ch := m.currentChapter()
	v := m.currentVerse()
	if ch == nil || v == nil {
		return styleSep.Render("No verse selected.")
	}

	wrapW := panelInnerW(m.cntW()) - 2
	if wrapW < 10 {
		wrapW = 10
	}

	var b strings.Builder

	b.WriteString(styleChapHead.Render(fmt.Sprintf("Chapter %d: %s", ch.Chapter, ch.Title)))
	b.WriteString("\n\n")
	b.WriteString(styleVerseNum.Render(fmt.Sprintf("Verse %d", v.Verse)))
	b.WriteString("\n\n")

	txt := v.Text
	if m.search == searchActive && m.searchQuery != "" {
		txt = hlText(txt, m.searchQuery)
	}
	b.WriteString(styleVerseBody.Render(wrap(txt, wrapW)))
	b.WriteString("\n")

	return b.String()
}

// ── Status bar ────────────────────────────────────────────────────────────────

func (m Model) drawStatus() string {
	// Build a single flat string — no lipgloss Width wrapping on the outer style
	var left string
	switch {
	case m.search == searchTyping:
		left = styleSBSrch.Render(fmt.Sprintf(" / %s▌", m.searchQuery))
	case m.status != "":
		left = styleSBInfo.Render(" " + m.status)
	default:
		sp := styleSBSep.Render(" │ ")
		left = " " +
			styleSBKey.Render("←→") + styleSB.Render(" panels") + sp +
			styleSBKey.Render("↑↓") + styleSB.Render(" nav") + sp +
			styleSBKey.Render("Enter") + styleSB.Render(" select") + sp +
			styleSBKey.Render("/") + styleSB.Render(" search") + sp +
			styleSBKey.Render("g/G") + styleSB.Render(" top/btm") + sp +
			styleSBKey.Render("q") + styleSB.Render(" quit")
	}

	var right string
	if ch := m.currentChapter(); ch != nil {
		if v := m.currentVerse(); v != nil {
			pct := 0
			if m.vpOK {
				pct = int(m.vp.ScrollPercent() * 100)
			}
			right = styleSBInfo.Render(
				fmt.Sprintf(" Ch %d  v%d  %d%% ", ch.Chapter, v.Verse, pct),
			)
		}
	}

	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	gap := m.width - lw - rw
	if gap < 1 {
		gap = 1
	}
	fill := styleSB.Render(strings.Repeat(" ", gap))

	// Clamp total to exactly m.width
	line := left + fill + right
	if lipgloss.Width(line) > m.width {
		line = left + fill
	}
	return styleSB.Render(line)
}

// ── Key handlers ──────────────────────────────────────────────────────────────

func (m Model) handleKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch km.String() {
	case "ctrl+c", "q", "Q":
		return m, tea.Quit

	case "left", "h":
		if m.focus > panelChapters {
			m.focus--
			m.status = ""
			m = m.syncVP()
		}

	case "right", "l", "enter", " ":
		if m.focus < panelContent {
			m.focus++
			m.status = ""
			if m.focus == panelVerses {
				m.verseCursor = 0
			}
			if m.focus == panelContent {
				m.vp.GotoTop()
			}
			m = m.syncVP()
		}

	case "up", "k":
		m = m.navUp()
		m = m.syncVP()

	case "down", "j":
		m = m.navDown()
		m = m.syncVP()

	case "pgup", "u":
		if m.focus == panelContent && m.vpOK {
			m.vp.HalfViewUp()
		}
	case "pgdown", "d":
		if m.focus == panelContent && m.vpOK {
			m.vp.HalfViewDown()
		}

	case "/":
		m.search = searchTyping
		m.searchQuery = ""
		m.status = ""

	case "n":
		if m.search == searchActive && len(m.searchHits) > 0 {
			m.searchIdx = (m.searchIdx + 1) % len(m.searchHits)
			m = m.jumpToHit()
		}
	case "N":
		if m.search == searchActive && len(m.searchHits) > 0 {
			m.searchIdx = (m.searchIdx - 1 + len(m.searchHits)) % len(m.searchHits)
			m = m.jumpToHit()
		}

	case "esc":
		m.search = searchOff
		m.searchQuery = ""
		m.searchHits = nil
		m.status = ""
		m = m.syncVP()

	case "g":
		switch m.focus {
		case panelChapters:
			m.chapterCursor = 0
			m.verseCursor = 0
		case panelVerses:
			m.verseCursor = 0
		case panelContent:
			if m.vpOK {
				m.vp.GotoTop()
			}
		}
		m = m.syncVP()

	case "G":
		switch m.focus {
		case panelChapters:
			m.chapterCursor = len(m.g.Chapters) - 1
			m.verseCursor = 0
		case panelVerses:
			if ch := m.currentChapter(); ch != nil {
				m.verseCursor = len(ch.Verses) - 1
			}
		case panelContent:
			if m.vpOK {
				m.vp.GotoBottom()
			}
		}
		m = m.syncVP()
	}

	return m, nil
}

func (m Model) handleSearchInput(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch km.String() {
	case "enter":
		if m.searchQuery != "" {
			hits := gita.Search(m.g, m.searchQuery, 0)
			if len(hits) == 0 {
				m.status = fmt.Sprintf(`No results for "%s"`, m.searchQuery)
				m.search = searchOff
			} else {
				m.searchHits = hits
				m.searchIdx = 0
				m.search = searchActive
				m.status = fmt.Sprintf(`"%s"  %d hit(s)   n/N next/prev   Esc clear`,
					m.searchQuery, len(hits))
				m = m.jumpToHit()
			}
		} else {
			m.search = searchOff
		}
	case "esc":
		m.search = searchOff
		m.searchQuery = ""
	case "backspace":
		if len(m.searchQuery) > 0 {
			rr := []rune(m.searchQuery)
			m.searchQuery = string(rr[:len(rr)-1])
		}
	default:
		if len(km.Runes) == 1 {
			m.searchQuery += string(km.Runes)
		}
	}
	m = m.syncVP()
	return m, nil
}

// ── Navigation ────────────────────────────────────────────────────────────────

func (m Model) navUp() Model {
	switch m.focus {
	case panelChapters:
		if m.chapterCursor > 0 {
			m.chapterCursor--
			m.verseCursor = 0
		}
	case panelVerses:
		if m.verseCursor > 0 {
			m.verseCursor--
		}
	case panelContent:
		if m.vpOK {
			m.vp.LineUp(1)
		}
	}
	return m
}

func (m Model) navDown() Model {
	switch m.focus {
	case panelChapters:
		if m.chapterCursor < len(m.g.Chapters)-1 {
			m.chapterCursor++
			m.verseCursor = 0
		}
	case panelVerses:
		if ch := m.currentChapter(); ch != nil {
			if m.verseCursor < len(ch.Verses)-1 {
				m.verseCursor++
			}
		}
	case panelContent:
		if m.vpOK {
			m.vp.LineDown(1)
		}
	}
	return m
}

func (m *Model) jumpToHit() Model {
	if len(m.searchHits) == 0 {
		return *m
	}
	hit := m.searchHits[m.searchIdx].Ref
	for ci, ch := range m.g.Chapters {
		if ch.Chapter != hit.ChapterNum {
			continue
		}
		m.chapterCursor = ci
		for vi, v := range ch.Verses {
			if v.Verse == hit.VerseNum {
				m.verseCursor = vi
				break
			}
		}
		break
	}
	m.focus = panelContent
	if m.vpOK {
		m.vp.GotoTop()
	}
	return m.syncVP()
}

// ── Data helpers ──────────────────────────────────────────────────────────────

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

// ── List rendering helper ─────────────────────────────────────────────────────

// renderList returns a string of exactly `h` lines wide `w`, scrolling to keep
// `cursor` visible and centred.
func renderList(rows []string, cursor, h, w int) string {
	if len(rows) == 0 {
		blank := lipgloss.NewStyle().Width(w).Background(cBgPanel).Render("")
		return strings.Join(repeatStr(blank, h), "\n")
	}

	// Scroll window
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

	// Pad to exactly h lines
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

// ── Text helpers ──────────────────────────────────────────────────────────────

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

// hlText case-insensitively wraps every occurrence of kw in s with searchHL style.
func hlText(s, kw string) string {
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
		lrem = lrem[i+len(lkw):]
	}
	return b.String()
}
