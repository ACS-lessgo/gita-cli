package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/whoisyurii/gita-cli/internal/gita"
)

var chapterCmd = &cobra.Command{
	Use:   "chapter <number>",
	Short: "Display all verses in a chapter",
	Long: `Display all available verses for a given chapter of the Bhagavad Gita.

Example:
  gita chapter 2`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chapterNum, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid chapter number: %q", args[0])
		}
		g, err := gita.Load()
		if err != nil {
			return fmt.Errorf("loading data: %w", err)
		}
		ch, err := gita.GetChapter(g, chapterNum)
		if err != nil {
			printError(err.Error())
			return nil
		}
		printChapter(ch)
		return nil
	},
}

func printChapter(ch *gita.Chapter) {
	fmt.Println()
	fmt.Println(dividerStyle.Render(divider))
	fmt.Printf(" %s\n", titleStyle.Render(fmt.Sprintf("Chapter %d: %s", ch.Chapter, ch.Title)))
	fmt.Printf(" %s\n", dimStyle.Render(fmt.Sprintf("%d verses in dataset", len(ch.Verses))))
	fmt.Println(dividerStyle.Render(divider))
	for _, v := range ch.Verses {
		fmt.Println()
		fmt.Printf(" %s\n", metaStyle.Render(fmt.Sprintf("Verse %d", v.Verse)))
		fmt.Println(verseStyle.Render(wordWrap(v.Text, 68)))
		fmt.Println(dimStyle.Render(" " + strings.Repeat("·", 50)))
	}
	fmt.Println()
	fmt.Println(dividerStyle.Render(divider))
	fmt.Println()
}
