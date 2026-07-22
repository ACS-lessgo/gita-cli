package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) drawPalette() string { return "" }

func (m Model) handlePaletteKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}
