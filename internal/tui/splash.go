package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Splash frames ─────────────────────────────────────────────────────────────

// Three brightness levels of the same ASCII art — cycled to create a
// "materialising" effect.
var splashFrames = [3]string{
	// frame 0 — full blocks
	` ██████╗ ██╗████████╗ █████╗
 ██╔════╝ ██║╚══██╔══╝██╔══██╗
 ██║  ███╗██║   ██║   ███████║
 ██║   ██║██║   ██║   ██╔══██║
 ╚██████╔╝██║   ██║   ██║  ██║
  ╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝`,

	// frame 1 — medium shade
	` ▓▓▓▓▓▓▓ ▓▓ ▓▓▓▓▓▓▓▓ ▓▓▓▓▓
 ▓▓      ▓▓    ▓▓   ▓▓   ▓▓
 ▓▓  ▓▓▓ ▓▓    ▓▓   ▓▓▓▓▓▓▓
 ▓▓   ▓▓ ▓▓    ▓▓   ▓▓   ▓▓
 ▓▓▓▓▓▓▓ ▓▓    ▓▓   ▓▓   ▓▓
  ▓▓▓▓▓  ▓▓    ▓▓   ▓▓   ▓▓`,

	// frame 2 — light shade
	` ░░░░░░░ ░░ ░░░░░░░░ ░░░░░
 ░░      ░░    ░░   ░░   ░░
 ░░  ░░░ ░░    ░░   ░░░░░░░
 ░░   ░░ ░░    ░░   ░░   ░░
 ░░░░░░░ ░░    ░░   ░░   ░░
  ░░░░░  ░░    ░░   ░░   ░░`,
}

// Colours paired to frames: bright → mid → dim
var splashFG = [3]lipgloss.Color{
	lipgloss.Color("#FFFFFF"),
	lipgloss.Color("#888888"),
	lipgloss.Color("#444444"),
}

// ── Messages ──────────────────────────────────────────────────────────────────

type tickMsg struct{}
type doneSplashMsg struct{}

func doTick() tea.Cmd {
	// 200 ms per tick — visibly slower animation
	return tea.Tick(200*time.Millisecond, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

// ── SplashModel ───────────────────────────────────────────────────────────────

type SplashModel struct {
	width, height int
	tick          int // total ticks elapsed
	done          bool
}

func (s SplashModel) Init() tea.Cmd { return doTick() }

func (s SplashModel) Update(msg tea.Msg) (SplashModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width, s.height = msg.Width, msg.Height
	case tickMsg:
		s.tick++
		// Auto-advance after ~5 s (25 ticks × 200 ms)
		if s.tick >= 25 {
			s.done = true
			return s, func() tea.Msg { return doneSplashMsg{} }
		}
		return s, doTick()
	case tea.KeyMsg:
		s.done = true
		return s, func() tea.Msg { return doneSplashMsg{} }
	}
	return s, nil
}

func (s SplashModel) View() string {
	if s.width == 0 {
		return ""
	}
	w, h := s.width, s.height

	// Pick frame: oscillate 0→1→2→1→0 … so it breathes
	seq := []int{0, 0, 1, 1, 2, 2, 1, 1, 0, 0}
	fi := seq[s.tick%len(seq)]

	artStyle := lipgloss.NewStyle().
		Foreground(splashFG[fi]).
		Background(cBg).
		Bold(true)

	// Pulsing sub-text colour
	subSeq := []lipgloss.Color{
		"#333333", "#555555", "#777777", "#999999",
		"#BBBBBB", "#DDDDDD", "#FFFFFF",
		"#DDDDDD", "#BBBBBB", "#999999",
		"#777777", "#555555",
	}
	subCol := subSeq[s.tick%len(subSeq)]
	subStyle := lipgloss.NewStyle().Foreground(subCol).Background(cBg).Italic(true)

	// Blinking prompt
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Background(cBg)
	if (s.tick/2)%2 == 0 {
		promptStyle = promptStyle.Foreground(lipgloss.Color("#AAAAAA"))
	}

	// Build lines
	var rows []string
	rows = append(rows, "")
	// rows = append(rows, center("ॐ", w, lipgloss.NewStyle().Foreground(subCol).Background(cBg).Bold(true)))
	rows = append(rows, "")

	for _, line := range strings.Split(splashFrames[fi], "\n") {
		rows = append(rows, center(line, w, artStyle))
	}

	rows = append(rows, "")
	rows = append(rows, center("─────────────────────────────────────", w, styleSplashDim))
	rows = append(rows, center("  Bhagavad Gita  Terminal Reader  ", w, subStyle))
	rows = append(rows, center("─────────────────────────────────────", w, styleSplashDim))
	rows = append(rows, "")
	rows = append(rows, center(`"You have the right to perform your actions,`, w, styleSplashDim))
	rows = append(rows, center(` but you are not entitled to the fruits."`, w, styleSplashDim))
	rows = append(rows, "")
	rows = append(rows, center("— Bhagavad Gita 2.47", w,
		lipgloss.NewStyle().Foreground(cDimmer).Background(cBg)))
	rows = append(rows, "")
	rows = append(rows, "")
	rows = append(rows, center("Press any key to continue", w, promptStyle))

	content := strings.Join(rows, "\n")

	// Vertically centre
	nLines := strings.Count(content, "\n") + 1
	topPad := (h - nLines) / 2
	if topPad < 0 {
		topPad = 0
	}

	emptyLine := strings.Repeat(" ", w)
	pad := strings.Repeat(emptyLine+"\n", topPad)
	return styleSplashBg.Width(w).Height(h).Render(pad + content)
}

// center returns s horizontally centred within width w, rendered with style.
func center(s string, w int, style lipgloss.Style) string {
	vis := lipgloss.Width(s)
	pad := (w - vis) / 2
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) + style.Render(s)
}
