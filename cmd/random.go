package cmd

import (
	"fmt"
	"math/rand"

	"github.com/spf13/cobra"
	"github.com/whoisyurii/gita-cli/internal/gita"
)

var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "Display a random verse",
	Args:  cobra.NoArgs,
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
		printVerse(&ref)
		return nil
	},
}
