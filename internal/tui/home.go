package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) drawHome() string { return "" }

func (m Model) handleHomeKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}
