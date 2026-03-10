package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/whoisyurii/gita-cli/internal/gita"
	"github.com/whoisyurii/gita-cli/internal/tui"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	gold    = lipgloss.Color("#D4AF37")
	saffron = lipgloss.Color("#FF9933")
	lotus   = lipgloss.Color("#C77DFF")
	dimGray = lipgloss.Color("#888888")

	titleStyle = lipgloss.NewStyle().
			Foreground(saffron).
			Bold(true)

	verseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F0E6D3")).
			Italic(true).
			PaddingLeft(2).
			PaddingRight(2)

	dividerStyle = lipgloss.NewStyle().
			Foreground(gold)

	metaStyle = lipgloss.NewStyle().
			Foreground(lotus).
			Bold(true)

	chapterTitleStyle = lipgloss.NewStyle().
				Foreground(saffron).
				Italic(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	highlightStyle = lipgloss.NewStyle().
			Foreground(gold).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#69FF94"))
)

const divider = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

// ── Helpers ───────────────────────────────────────────────────────────────────

func printVerse(ref *gita.VerseRef) {
	fmt.Println()
	fmt.Println(dividerStyle.Render(divider))
	fmt.Printf(" %s  %s\n",
		metaStyle.Render(fmt.Sprintf("Chapter %d • Verse %d", ref.ChapterNum, ref.VerseNum)),
		chapterTitleStyle.Render("— "+ref.ChapterTitle),
	)
	fmt.Println(dividerStyle.Render(divider))
	fmt.Println()
	fmt.Println(verseStyle.Render(wordWrap(ref.Text, 70)))
	fmt.Println()
	fmt.Println(dividerStyle.Render(divider))
	fmt.Println()
}

func printError(msg string) {
	fmt.Fprintln(os.Stderr, errorStyle.Render("✗ Error: "+msg))
}

func wordWrap(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}
	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) > width {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}

// ── Root command ──────────────────────────────────────────────────────────────

var rootCmd = &cobra.Command{
	Use:   "gita",
	Short: "📖 Bhagavad Gita CLI — wisdom from the battlefield",
	Long: lipgloss.NewStyle().Foreground(saffron).Render(`
  ██████╗ ██╗████████╗ █████╗
 ██╔════╝ ██║╚══██╔══╝██╔══██╗
 ██║  ███╗██║   ██║   ███████║
 ██║   ██║██║   ██║   ██╔══██║
 ╚██████╔╝██║   ██║   ██║  ██║
  ╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝  CLI`) + `

` + dimStyle.Render("Access the timeless wisdom of the Bhagavad Gita from your terminal.") + `

` + titleStyle.Render("COMMANDS") + `
  ` + highlightStyle.Render("gita") + `                          Launch interactive TUI browser
  ` + highlightStyle.Render("gita verse <chapter> <verse>") + `   Retrieve a specific verse
  ` + highlightStyle.Render("gita chapter <number>") + `          Display all verses in a chapter
  ` + highlightStyle.Render("gita random") + `                    Show a random verse
  ` + highlightStyle.Render("gita search <keyword>") + `          Search verses by keyword
  ` + highlightStyle.Render("gita quote") + `                     Print an inspiring daily quote

` + dimStyle.Render("Run 'gita [command] --help' for more information about a command."),
	// No args → launch TUI
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printError(err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(verseCmd)
	rootCmd.AddCommand(chapterCmd)
	rootCmd.AddCommand(randomCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(quoteCmd)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
