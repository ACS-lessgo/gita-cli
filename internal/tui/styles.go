package tui

import "github.com/charmbracelet/lipgloss"

// ── Palette ───────────────────────────────────────────────────────────────────

var (
	cWhite   = lipgloss.Color("#FFFFFF")
	cBright  = lipgloss.Color("#DDDDDD")
	cMid     = lipgloss.Color("#888888")
	cDim     = lipgloss.Color("#444444")
	cDimmer  = lipgloss.Color("#2A2A2A")
	cBlack   = lipgloss.Color("#000000")
	cBg      = lipgloss.Color("#0A0A0A")
	cBgPanel = lipgloss.Color("#0F0F0F")
	cBorder  = lipgloss.Color("#303030")
	cHot     = lipgloss.Color("#FFFFFF")
)

// ── Row styles ────────────────────────────────────────────────────────────────

func chapRowStyle(selected bool) lipgloss.Style {
	if selected {
		return lipgloss.NewStyle().
			Background(cWhite).
			Foreground(cBlack).
			Bold(true)
	}
	return lipgloss.NewStyle().
		Background(cBgPanel).
		Foreground(cBright)
}

func vrsRowStyle(selected bool) lipgloss.Style {
	if selected {
		return lipgloss.NewStyle().
			Background(cWhite).
			Foreground(cBlack).
			Bold(true)
	}
	return lipgloss.NewStyle().
		Background(cBgPanel).
		Foreground(cBright)
}

// ── Panel border ──────────────────────────────────────────────────────────────

func borderCol(active bool) lipgloss.Color {
	if active {
		return cHot
	}
	return cBorder
}

// ── Header strip above each panel ────────────────────────────────────────────

func panelHeader(text string, active bool) lipgloss.Style {
	fg := cMid
	if active {
		fg = cWhite
	}
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(cBg).
		Bold(active).
		PaddingLeft(1)
}

// ── Content area ──────────────────────────────────────────────────────────────

var (
	styleChapHead  = lipgloss.NewStyle().Foreground(cWhite).Bold(true)
	styleVerseNum  = lipgloss.NewStyle().Foreground(cWhite).Bold(true).Underline(true)
	styleVerseBody = lipgloss.NewStyle().Foreground(cBright).Italic(true)
	styleSep       = lipgloss.NewStyle().Foreground(cDimmer)
	styleSearchHL  = lipgloss.NewStyle().Foreground(cBlack).Background(cWhite).Bold(true)
)

// ── Status bar ────────────────────────────────────────────────────────────────

var (
	styleSB     = lipgloss.NewStyle().Background(cBg).Foreground(cMid)
	styleSBKey  = lipgloss.NewStyle().Background(cBg).Foreground(cWhite).Bold(true)
	styleSBSep  = lipgloss.NewStyle().Background(cBg).Foreground(cDim)
	styleSBInfo = lipgloss.NewStyle().Background(cBg).Foreground(cMid)
	styleSBSrch = lipgloss.NewStyle().Background(cBg).Foreground(cWhite).Bold(true)
)

// ── Splash ────────────────────────────────────────────────────────────────────

var (
	styleSplashBg  = lipgloss.NewStyle().Background(cBg)
	styleSplashDim = lipgloss.NewStyle().Background(cBg).Foreground(cDimmer)
)

// ── Column title styles (inside panel box) ────────────────────────────────────

func colTitleStyle(active bool) lipgloss.Style {
	fg := cMid
	if active {
		fg = cWhite
	}
	return lipgloss.NewStyle().
		Background(cBgPanel).
		Foreground(fg).
		Bold(active)
}

var colSepStyle = lipgloss.NewStyle().
	Background(cBgPanel).
	Foreground(cDimmer)

// ── Reading screen (borderless) ───────────────────────────────────────────

var (
	styleDivider   = lipgloss.NewStyle().Foreground(cBorder).Background(cBg)
	styleMarkBadge = lipgloss.NewStyle().Foreground(cWhite).Bold(true)
	styleCiteRule  = lipgloss.NewStyle().Foreground(cDimmer)
	styleCiteText  = lipgloss.NewStyle().Foreground(cMid)

	styleSliderTrack = lipgloss.NewStyle().Foreground(cDim)
	styleSliderMark  = lipgloss.NewStyle().Foreground(cHot).Bold(true)
)
