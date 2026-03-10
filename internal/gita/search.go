package gita

import (
	"strings"
)

// SearchResult holds a matched verse along with context.
type SearchResult struct {
	Ref     VerseRef
	Keyword string
}

// Search performs a case-insensitive keyword search across all verses.
// Returns up to maxResults results (0 = unlimited).
func Search(g *Gita, keyword string, maxResults int) []SearchResult {
	lower := strings.ToLower(keyword)
	var results []SearchResult

	for _, ref := range AllVerses(g) {
		if strings.Contains(strings.ToLower(ref.Text), lower) {
			results = append(results, SearchResult{Ref: ref, Keyword: keyword})
			if maxResults > 0 && len(results) >= maxResults {
				break
			}
		}
	}
	return results
}
