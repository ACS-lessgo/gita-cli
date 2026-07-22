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
