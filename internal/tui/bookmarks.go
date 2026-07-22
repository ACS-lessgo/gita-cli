package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) drawBookmarks() string { return "" }

func (m Model) handleBookmarksKey(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}
