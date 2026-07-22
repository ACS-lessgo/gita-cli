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
