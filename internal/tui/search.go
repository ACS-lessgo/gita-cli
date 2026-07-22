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
