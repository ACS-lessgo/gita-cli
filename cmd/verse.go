package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/ACS-lessgo/gita-cli/internal/gita"
)

var verseCmd = &cobra.Command{
	Use:   "verse <chapter> <verse>",
	Short: "Retrieve a specific verse",
	Long: `Retrieve and display a specific verse from the Bhagavad Gita.

Example:
  gita verse 2 47
  gita verse 18 66`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		chapterNum, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid chapter number: %q", args[0])
		}
		verseNum, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid verse number: %q", args[1])
		}
		g, err := gita.Load()
		if err != nil {
			return fmt.Errorf("loading data: %w", err)
		}
		ref, err := gita.GetVerse(g, chapterNum, verseNum)
		if err != nil {
			printError(err.Error())
			return nil
		}
		printVerse(ref)
		return nil
	},
}
