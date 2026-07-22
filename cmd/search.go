package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ACS-lessgo/gita-cli/internal/gita"
)

var maxSearchResults int

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search verses by keyword",
	Long: `Search all verses of the Bhagavad Gita for a given keyword or phrase.
The search is case-insensitive.

Example:
  gita search "duty"
  gita search "soul" --limit 5`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword := strings.Join(args, " ")
		g, err := gita.Load()
		if err != nil {
			return fmt.Errorf("loading data: %w", err)
		}
		results := gita.Search(g, keyword, maxSearchResults)
		printSearchResults(keyword, results)
		return nil
	},
}

func init() {
	searchCmd.Flags().IntVarP(&maxSearchResults, "limit", "l", 0, "Maximum number of results (0 = all)")
}

func printSearchResults(keyword string, results []gita.SearchResult) {
	fmt.Println()
	fmt.Println(dividerStyle.Render(divider))
	fmt.Printf(" %s  %s\n",
		titleStyle.Render("Search Results"),
		dimStyle.Render(fmt.Sprintf(`keyword: "%s"`, keyword)),
	)
	if len(results) == 0 {
		fmt.Printf(" %s\n", dimStyle.Render("No verses found matching your query."))
		fmt.Println(dividerStyle.Render(divider))
		fmt.Println()
		return
	}
	fmt.Printf(" %s\n", successStyle.Render(fmt.Sprintf("%d verse(s) found", len(results))))
	fmt.Println(dividerStyle.Render(divider))
	for _, r := range results {
		fmt.Println()
		fmt.Printf(" %s  %s\n",
			metaStyle.Render(fmt.Sprintf("Chapter %d • Verse %d", r.Ref.ChapterNum, r.Ref.VerseNum)),
			chapterTitleStyle.Render("— "+r.Ref.ChapterTitle),
		)
		fmt.Println(verseStyle.Render(wordWrap(r.Ref.Text, 68)))
		fmt.Println(dimStyle.Render(" " + strings.Repeat("·", 50)))
	}
	fmt.Println()
	fmt.Println(dividerStyle.Render(divider))
	fmt.Println()
}
