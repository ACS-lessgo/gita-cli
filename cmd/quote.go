package cmd

import (
	"fmt"
	"math/rand"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/whoisyurii/gita-cli/internal/gita"
)

var quoteCmd = &cobra.Command{
	Use:   "quote",
	Short: "Display an inspiring daily quote",
	Long: `Display a beautifully formatted motivational verse from the Bhagavad Gita.

Example:
  gita quote`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := gita.Load()
		if err != nil {
			return fmt.Errorf("loading data: %w", err)
		}
		all := gita.AllVerses(g)
		if len(all) == 0 {
			printError("no verses found in dataset")
			return nil
		}
		ref := all[rand.Intn(len(all))]
		printQuote(&ref)
		return nil
	},
}

func printQuote(ref *gita.VerseRef) {
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(gold).
		Padding(1, 3).
		Width(66)

	quoteTextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F5E6CA")).
		Italic(true)

	sourceStyle := lipgloss.NewStyle().
		Foreground(lotus).
		Align(lipgloss.Right)

	header := titleStyle.Render("🕉  Bhagavad Gita") + "  " + dimStyle.Render("Daily Quote")
	quoteText := quoteTextStyle.Render(`"` + wordWrap(ref.Text, 58) + `"`)
	source := sourceStyle.Render(
		fmt.Sprintf("— Chapter %d, Verse %d · %s", ref.ChapterNum, ref.VerseNum, ref.ChapterTitle),
	)

	fmt.Println()
	fmt.Println("  " + header)
	fmt.Println()
	fmt.Println("  " + boxStyle.Render(quoteText+"\n\n"+source))
	fmt.Println()
}
