# TUI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace gita-cli's 3-panel TUI browser with a screen-based interface (Home / Reading / Search / Bookmarks + Palette/Help overlays) matching the approved design spec, while preserving the color scheme and splash animation exactly.

**Architecture:** Extend the existing single-`Model` Bubble Tea pattern with a `screen` enum (Home/Reading/Search/Bookmarks) and an `overlay` enum (Palette/Help) instead of introducing per-screen `tea.Model`s. Split `internal/tui` by screen into focused files (`home.go`, `reading.go`, `search.go`, `bookmarks.go`, `palette.go`, `help.go`, `state.go`) so `model.go` stays the state struct + dispatch, not a growing monolith.

**Tech Stack:** Go 1.22, Bubble Tea (`charmbracelet/bubbletea`), Bubbles viewport (`charmbracelet/bubbles/viewport`), Lip Gloss (`charmbracelet/lipgloss`) — all already in `go.mod`, no new dependencies.

## Global Constraints

- Module path: `github.com/ACS-lessgo/gita-cli` (already set).
- `internal/tui/splash.go` is not modified — the splash animation stays exactly as-is.
- `internal/tui/styles.go` is not modified — no new colors; every screen reuses the existing exported vars/functions (`cWhite/cBright/cMid/cDim/cDimmer/cBg/cBgPanel/cBorder/cHot`, `chapRowStyle`, `borderCol`, `colTitleStyle`, `colSepStyle`, `styleChapHead/styleVerseNum/styleVerseBody/styleSep/styleSearchHL`, `styleSB/styleSBKey/styleSBSep/styleSBInfo/styleSBSrch`). Some existing helpers (e.g. `vrsRowStyle`) become unused after this redesign — leave them; Go does not error on unused package-level declarations, and the file is explicitly out of scope.
- `internal/tui/run.go` requires no changes — `New(g *gita.Gita) Model` keeps its existing signature.
- The old 3-panel layout (`panel` enum, `drawChapPanel`/`drawVrsPanel`/`drawCntPanel`, `handleSearchInput`, `jumpToHit`, `navUp`/`navDown`) is fully removed, not kept alongside the new screens.
- Non-interactive subcommands in `cmd/` (`verse`, `chapter`, `random`, `search`, `quote`) are untouched.
- No new dependencies are added to `go.mod`.

---

### Task 1: State persistence (`state.go`)

**Files:**
- Create: `internal/tui/state.go`
- Test: `internal/tui/state_test.go`

**Interfaces:**
- Consumes: nothing (pure, no `Model` dependency).
- Produces: `type persistedState struct { LastChapter int; LastVerse int; Bookmarks []bookmark }`, `type bookmark struct { ChapterNum int; VerseNum int; SavedAt time.Time }`, `func loadState() (persistedState, error)`, `func saveState(st persistedState) error`. Task 2's `Model.persistState()` method and `Model.New()` call these.

- [ ] **Step 1: Write the failing tests**

Create `internal/tui/state_test.go`:

```go
package tui

import (
	"testing"
	"time"
)

func TestSaveLoadStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	want := persistedState{
		LastChapter: 2,
		LastVerse:   47,
		Bookmarks: []bookmark{
			{ChapterNum: 6, VerseNum: 5, SavedAt: time.Now().Truncate(time.Second)},
		},
	}
	if err := saveState(want); err != nil {
		t.Fatalf("saveState: %v", err)
	}

	got, err := loadState()
	if err != nil {
		t.Fatalf("loadState: %v", err)
	}
	if got.LastChapter != want.LastChapter || got.LastVerse != want.LastVerse {
		t.Errorf("LastChapter/LastVerse = %d/%d, want %d/%d",
			got.LastChapter, got.LastVerse, want.LastChapter, want.LastVerse)
	}
	if len(got.Bookmarks) != 1 ||
		got.Bookmarks[0].ChapterNum != 6 ||
		got.Bookmarks[0].VerseNum != 5 {
		t.Errorf("Bookmarks = %+v, want one bookmark at 6.5", got.Bookmarks)
	}
}

func TestLoadStateMissingFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	got, err := loadState()
	if err == nil {
		t.Fatal("expected an error for a missing state file")
	}
	if got.LastChapter != 0 || got.LastVerse != 0 || len(got.Bookmarks) != 0 {
		t.Errorf("got = %+v, want zero value", got)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/tui/... -run TestSaveLoadStateRoundTrip -v`
Expected: FAIL — `undefined: persistedState` (state.go doesn't exist yet)

- [ ] **Step 3: Write the implementation**

Create `internal/tui/state.go`:

```go
package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// persistedState is saved to disk between runs.
type persistedState struct {
	LastChapter int        `json:"lastChapter"`
	LastVerse   int        `json:"lastVerse"`
	Bookmarks   []bookmark `json:"bookmarks"`
}

type bookmark struct {
	ChapterNum int       `json:"chapterNum"`
	VerseNum   int       `json:"verseNum"`
	SavedAt    time.Time `json:"savedAt"`
}

func statePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gita-cli", "state.json"), nil
}

// loadState reads persisted state from disk. A missing or corrupt file
// returns a zero-value persistedState alongside the error — callers that
// don't care why loading failed can ignore the error and use the result.
func loadState() (persistedState, error) {
	path, err := statePath()
	if err != nil {
		return persistedState{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return persistedState{}, err
	}
	var st persistedState
	if err := json.Unmarshal(data, &st); err != nil {
		return persistedState{}, err
	}
	return st, nil
}

func saveState(st persistedState) error {
	path, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/tui/... -run 'TestSaveLoadStateRoundTrip|TestLoadStateMissingFileReturnsError' -v`
Expected: PASS for both tests

- [ ] **Step 5: Commit**

```bash
git add internal/tui/state.go internal/tui/state_test.go
git commit -m "Add TUI state persistence (bookmarks, last-read position)"
```

---

### Task 2: Scaffold Model + screen stubs

**Files:**
- Modify (full rewrite): `internal/tui/model.go`
- Create: `internal/tui/home.go`, `internal/tui/reading.go`, `internal/tui/search.go`, `internal/tui/bookmarks.go`, `internal/tui/palette.go`, `internal/tui/help.go`

**Interfaces:**
- Consumes: `persistedState`, `bookmark`, `loadState`, `saveState` (Task 1); `SplashModel`, `doneSplashMsg` (existing `splash.go`); styles from existing `styles.go`.
- Produces: `Model` struct with fields `g, splash, showSplash, width, height, screen, overlay, state, chapterCursor, verseCursor, sidebarOpen, vp, vpOK, searchQuery, searchHits, searchCursor, bookmarkCursor, paletteQuery, paletteResults, paletteCursor, status`. Methods `New`, `Init`, `Update`, `View`, `handleKey`, `enterReading`, `persistState`, `currentChapter`, `currentVerse`, `chapterIndex`, `drawStatus`. Package functions `panelInnerW`, `panelInnerH`, `verseIndex`, `renderList`, `repeatStr`, `wrap`, `runesTrunc`, `hlText`. Type `paletteResult{kind, chapterNum, verseNum, label}`. Each screen file must define exactly the stub functions listed below so `model.go` compiles against them — later tasks replace stub bodies but keep the same signatures.

- [ ] **Step 1: Replace `internal/tui/model.go` entirely**

```go
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
```

- [ ] **Step 2: Create stub `internal/tui/home.go`**

```go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) drawHome() string { return "" }

func (m Model) handleHomeKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}
```

- [ ] **Step 3: Create stub `internal/tui/reading.go`**

```go
package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) syncVP() Model {
	if !m.vpOK {
		m.vp = viewport.New(1, 1)
		m.vpOK = true
	}
	return m
}

func (m Model) drawReading() string { return "" }

func (m Model) handleReadingKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}
```

- [ ] **Step 4: Create stub `internal/tui/search.go`**

```go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) enterSearch() Model {
	m.screen = screenSearch
	m.overlay = overlayNone
	m.status = ""
	return m
}

func (m Model) drawSearch() string { return "" }

func (m Model) handleSearchKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}
```

- [ ] **Step 5: Create stub `internal/tui/bookmarks.go`**

```go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) drawBookmarks() string { return "" }

func (m Model) handleBookmarksKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}
```

- [ ] **Step 6: Create stub `internal/tui/palette.go`**

```go
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) drawPalette() string { return "" }

func (m Model) handlePaletteKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}
```

- [ ] **Step 7: Create stub `internal/tui/help.go`**

```go
package tui

func (m Model) drawHelp() string { return "" }
```

- [ ] **Step 8: Build and vet**

Run: `go build ./... && go vet ./...`
Expected: no output, exit code 0

- [ ] **Step 9: Commit**

```bash
git add internal/tui/model.go internal/tui/home.go internal/tui/reading.go internal/tui/search.go internal/tui/bookmarks.go internal/tui/palette.go internal/tui/help.go
git commit -m "Scaffold screen-based TUI Model (Home/Reading/Search/Bookmarks + overlays)"
```

---

### Task 3: Home screen

**Files:**
- Modify (replace stub body): `internal/tui/home.go`
- Test: `internal/tui/home_test.go`

**Interfaces:**
- Consumes: `Model.currentChapter/currentVerse` (model.go), `gita.AllVerses`, `gita.VerseRef` (`internal/gita`), `wrap`, `styleSep/styleVerseNum/styleVerseBody/styleSBKey/styleSB/styleSBInfo` (model.go/styles.go), `m.enterReading()`.
- Produces: `func dailyVerseIndex(all []gita.VerseRef, day time.Time) int` (package-level, tested directly). `drawHome`/`handleHomeKey` keep their Task 2 signatures.

- [ ] **Step 1: Write the failing test**

Create `internal/tui/home_test.go`:

```go
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
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/tui/... -run TestDailyVerseIndexDeterministicAndInRange -v`
Expected: FAIL — `undefined: dailyVerseIndex`

- [ ] **Step 3: Replace `internal/tui/home.go`**

```go
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
```

- [ ] **Step 4: Run the test to verify it passes, then build**

Run: `go test ./internal/tui/... -run TestDailyVerseIndexDeterministicAndInRange -v && go build ./... && go vet ./...`
Expected: PASS, then no build/vet output

- [ ] **Step 5: Commit**

```bash
git add internal/tui/home.go internal/tui/home_test.go
git commit -m "Implement Home screen (verse of the day, continue reading)"
```

---

### Task 4: Bookmarks screen

**Files:**
- Modify (replace stub body): `internal/tui/bookmarks.go`
- Test: `internal/tui/bookmarks_test.go`

**Interfaces:**
- Consumes: `Model.currentChapter/currentVerse/chapterIndex/persistState` (model.go), `verseIndex/renderList/runesTrunc/chapRowStyle/borderCol/colTitleStyle/colSepStyle/styleVerseNum/styleSep/cBg/cBgPanel` (model.go/styles.go), `bookmark` type (state.go), `m.enterReading()` (model.go).
- Produces: `func (m Model) findBookmark(chapterNum, verseNum int) int`, `func (m Model) isBookmarked(chapterNum, verseNum int) bool`, `func (m Model) toggleBookmark() Model`, `func relTime(t time.Time) string` — all consumed by Task 5 (`reading.go`). `drawBookmarks`/`handleBookmarksKey` keep their Task 2 signatures.

- [ ] **Step 1: Write the failing test**

Create `internal/tui/bookmarks_test.go`:

```go
package tui

import (
	"testing"

	"github.com/ACS-lessgo/gita-cli/internal/gita"
)

func testGita() *gita.Gita {
	return &gita.Gita{Chapters: []gita.Chapter{
		{Chapter: 2, Title: "Transcendental Knowledge", Verses: []gita.Verse{
			{Verse: 47, Text: "You have a right to your actions, but never to the fruits of your actions."},
		}},
	}}
}

func TestToggleBookmarkAddAndRemove(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	m := New(testGita())

	m = m.toggleBookmark()
	if len(m.state.Bookmarks) != 1 {
		t.Fatalf("after add: len(Bookmarks) = %d, want 1", len(m.state.Bookmarks))
	}
	if m.state.Bookmarks[0].ChapterNum != 2 || m.state.Bookmarks[0].VerseNum != 47 {
		t.Errorf("bookmark = %+v, want chapter 2 verse 47", m.state.Bookmarks[0])
	}
	if !m.isBookmarked(2, 47) {
		t.Error("isBookmarked(2, 47) = false, want true after adding")
	}

	m = m.toggleBookmark()
	if len(m.state.Bookmarks) != 0 {
		t.Fatalf("after remove: len(Bookmarks) = %d, want 0", len(m.state.Bookmarks))
	}
	if m.isBookmarked(2, 47) {
		t.Error("isBookmarked(2, 47) = true, want false after removing")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/tui/... -run TestToggleBookmarkAddAndRemove -v`
Expected: FAIL — `m.toggleBookmark undefined` (bookmarks.go stub doesn't define it yet)

- [ ] **Step 3: Replace `internal/tui/bookmarks.go`**

```go
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
```

- [ ] **Step 4: Run the test to verify it passes, then build**

Run: `go test ./internal/tui/... -run TestToggleBookmarkAddAndRemove -v && go build ./... && go vet ./...`
Expected: PASS, then no build/vet output

- [ ] **Step 5: Commit**

```bash
git add internal/tui/bookmarks.go internal/tui/bookmarks_test.go
git commit -m "Implement Bookmarks screen and bookmark toggle"
```

---

### Task 5: Reading screen

**Files:**
- Modify (replace stub body): `internal/tui/reading.go`

**Interfaces:**
- Consumes: `Model.currentChapter/currentVerse` (model.go), `panelInnerW/panelInnerH/renderList/wrap/runesTrunc` (model.go), `chapRowStyle/borderCol/colTitleStyle/colSepStyle/styleSep/styleChapHead/styleVerseNum/styleVerseBody/cBgPanel` (styles.go), `m.isBookmarked` / `m.toggleBookmark` (Task 4, `bookmarks.go`).
- Produces: `syncVP`/`drawReading`/`handleReadingKey` keep their Task 2 signatures. `readingContentW`/`readingBodyH` are new helper methods used only within this file.

This is the file `model.go`'s `Update` (window resize, splash-done) and `View`/`handleKey` already call into (`syncVP`, `drawReading`, `handleReadingKey`) — no other file changes needed.

- [ ] **Step 1: Replace `internal/tui/reading.go`**

```go
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
```

- [ ] **Step 2: Build and vet**

Run: `go build ./... && go vet ./...`
Expected: no output, exit code 0

- [ ] **Step 3: Manual smoke test**

Run: `go run . `
Expected: splash plays, then Home screen. Press `2` → Reading screen shows the chapter sidebar and a verse. `j`/`k` move verse, `h`/`l` move chapter, `e` hides/shows the sidebar, `m` marks the verse (status bar shows "Bookmarked", `◆ marked` appears in the header). Press `q` to quit.

- [ ] **Step 4: Commit**

```bash
git add internal/tui/reading.go
git commit -m "Implement Reading screen (sidebar + verse pane, replaces 3-panel layout)"
```

---

### Task 6: Search screen

**Files:**
- Modify (replace stub body): `internal/tui/search.go`

**Interfaces:**
- Consumes: `gita.Search`, `gita.SearchResult` (`internal/gita`), `Model.chapterIndex/enterReading` (model.go), `verseIndex/panelInnerW/panelInnerH/renderList/wrap/runesTrunc/hlText` (model.go), `chapRowStyle/borderCol/colTitleStyle/colSepStyle/styleSep/styleVerseBody/styleSBSrch/styleSBInfo/styleSB/cBgPanel` (styles.go).
- Produces: `enterSearch`/`drawSearch`/`handleSearchKey` keep their Task 2 signatures.

- [ ] **Step 1: Replace `internal/tui/search.go`**

```go
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
```

- [ ] **Step 2: Build and vet**

Run: `go build ./... && go vet ./...`
Expected: no output, exit code 0

- [ ] **Step 3: Manual smoke test**

Run: `go run .`
Expected: press `/` from any screen → Search screen. Type a word (e.g. `duty`) → results list fills in live, preview pane on the right shows the selected verse with the term highlighted. `↑`/`↓` move selection. `Enter` opens the selected verse in Reading. `Esc` returns to Reading without opening anything.

- [ ] **Step 4: Commit**

```bash
git add internal/tui/search.go
git commit -m "Implement Search screen (live filter, results + preview pane)"
```

---

### Task 7: Palette overlay

**Files:**
- Modify (replace stub body): `internal/tui/palette.go`

**Interfaces:**
- Consumes: `gita.Search` (`internal/gita`), `Model.chapterIndex/enterReading` (model.go), `verseIndex/runesTrunc` (model.go), `chapRowStyle/borderCol/colSepStyle/styleSep/styleSBSrch/cBg/cBgPanel` (styles.go), `paletteResult` type (model.go).
- Produces: `drawPalette`/`handlePaletteKey` keep their Task 2 signatures. `computePaletteResults`/`parseChapterVerse` are new, used only within this file.

- [ ] **Step 1: Replace `internal/tui/palette.go`**

```go
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
```

- [ ] **Step 2: Build and vet**

Run: `go build ./... && go vet ./...`
Expected: no output, exit code 0

- [ ] **Step 3: Manual smoke test**

Run: `go run .`
Expected: press `p` from any screen → Palette overlay. Type `2.47` and hit `Enter` → jumps straight to Reading at chapter 2 verse 47. Reopen with `p`, type a word (e.g. `yoga`) → mixed chapter/verse suggestions appear, `Enter` on one jumps there, `Esc` dismisses without navigating.

- [ ] **Step 4: Commit**

```bash
git add internal/tui/palette.go
git commit -m "Implement jump Palette overlay (chapter.verse or word jump)"
```

---

### Task 8: Help overlay

**Files:**
- Modify (replace stub body): `internal/tui/help.go`

**Interfaces:**
- Consumes: `borderCol/styleVerseNum/styleSBKey/styleSB/cBgPanel` (styles.go).
- Produces: `drawHelp` keeps its Task 2 signature. Key handling for this overlay already lives in `model.go`'s `handleKey` (Task 2) — no handler function needed here.

- [ ] **Step 1: Replace `internal/tui/help.go`**

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpRow struct{ key, desc string }
type helpGroup struct {
	name string
	rows []helpRow
}

var helpGroups = []helpGroup{
	{"NAVIGATE", []helpRow{
		{"j / k", "next / previous verse"},
		{"h / l", "previous / next chapter"},
		{"g / G", "first / last verse in chapter"},
		{"e", "toggle chapter sidebar"},
	}},
	{"FIND", []helpRow{
		{"p", "jump palette (2.47, or words)"},
		{"/", "search all verses"},
	}},
	{"MANAGE", []helpRow{
		{"m", "bookmark current verse"},
		{"b", "open bookmarks"},
	}},
	{"SESSION", []helpRow{
		{"1 2 3 4", "home / reading / search / bookmarks"},
		{"?", "toggle this help"},
		{"esc", "dismiss / back to reading"},
		{"q", "quit (position saved)"},
	}},
}

func (m Model) drawHelp() string {
	w := m.width * 80 / 100
	if w > 90 {
		w = 90
	}
	if w < 40 {
		w = 40
	}
	if w > m.width-4 {
		w = m.width - 4
	}
	innerW := w - 2

	colW := innerW / 2
	if colW < 20 {
		colW = 20
	}

	var cols [2]strings.Builder
	for i, g := range helpGroups {
		col := &cols[i%2]
		col.WriteString(styleVerseNum.Render(g.name))
		col.WriteString("\n")
		for _, r := range g.rows {
			col.WriteString(styleSBKey.Render(fmt.Sprintf("%-8s", r.key)))
			col.WriteString(styleSB.Render(r.desc))
			col.WriteString("\n")
		}
		col.WriteString("\n")
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(colW).Render(cols[0].String()),
		lipgloss.NewStyle().Width(colW).Render(cols[1].String()),
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderCol(true)).
		Width(innerW).
		Background(cBgPanel).
		Padding(1, 2).
		Render(body)

	h := m.height - 1
	if h < 3 {
		h = 3
	}
	return lipgloss.Place(m.width, h, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceBackground(cBg))
}
```

- [ ] **Step 2: Build and vet**

Run: `go build ./... && go vet ./...`
Expected: no output, exit code 0

- [ ] **Step 3: Manual smoke test**

Run: `go run .`
Expected: press `?` from any screen → Help overlay with grouped keybindings. Press `?` or `Esc` to dismiss.

- [ ] **Step 4: Commit**

```bash
git add internal/tui/help.go
git commit -m "Implement Help overlay"
```

---

### Task 9: Full verification and integration

**Files:** none created — verification only.

**Interfaces:** none — this task exercises the finished package as a whole.

- [ ] **Step 1: Run the full test suite**

Run: `go test ./... -v`
Expected: PASS for all tests across `internal/gita` (unchanged) and `internal/tui` (`TestSaveLoadStateRoundTrip`, `TestLoadStateMissingFileReturnsError`, `TestDailyVerseIndexDeterministicAndInRange`, `TestToggleBookmarkAddAndRemove`)

- [ ] **Step 2: Run build and vet across the whole module**

Run: `go build ./... && go vet ./...`
Expected: no output, exit code 0

- [ ] **Step 3: Full manual walkthrough**

Run: `go run .` and walk through, in order:
1. Splash plays with its "materializing" animation exactly as before (unchanged file) — confirm no visual regression.
2. Lands on **Home**: verse of the day is visible, "Begin reading" shown (first run, no saved state).
3. `Enter` → **Reading**. Sidebar lists all 18 chapters. `j`/`k` move verses, `h`/`l` move chapters, `e` toggles the sidebar.
4. `m` on a verse → status bar shows "Bookmarked", `◆ marked` appears next to the verse ref.
5. `b` → **Bookmarks** shows the verse just marked, with a relative time ("just now"). `Enter` opens it back in Reading. `x` removes it.
6. `/` → **Search**. Type `duty` → live results + preview with highlighted term. `Enter` opens a result in Reading.
7. `p` → **Palette**. Type `2.47`, `Enter` → jumps to chapter 2 verse 47. Reopen, type `yoga`, confirm mixed chapter/verse suggestions.
8. `?` → **Help** overlay with grouped keybindings; `Esc` dismisses.
9. `1` → back to **Home**; resume label now says "Continue reading" with the last-visited ref.
10. `q` to quit, then run `go run .` again — confirm Home's "Continue reading" still points at the same verse, and the bookmark from step 4 (if not removed in step 5) is still in Bookmarks. This proves state persisted to `~/.config/gita-cli/state.json` (or platform equivalent) across process restarts.

- [ ] **Step 4: Final commit**

If any fixes were needed during the walkthrough, commit them:

```bash
git add -A
git status --short   # confirm only intended files are staged
git commit -m "Fix issues found during TUI redesign walkthrough"
```

(Skip this step if the walkthrough required no changes.)
