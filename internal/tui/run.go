package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/whoisyurii/gita-cli/internal/gita"
)

// Run loads data and starts the interactive TUI.
func Run() error {
	g, err := gita.Load()
	if err != nil {
		return fmt.Errorf("loading gita data: %w", err)
	}
	m := New(g)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error running TUI:", err)
		return err
	}
	return nil
}
