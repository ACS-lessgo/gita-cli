package gita

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed data/gita.json
var gitaFS embed.FS

var (
	cache     *Gita
	cacheOnce sync.Once
	cacheErr  error
)

// Load returns the Bhagavad Gita data, using a cached version after first load.
func Load() (*Gita, error) {
	cacheOnce.Do(func() {
		data, err := gitaFS.ReadFile("data/gita.json")
		if err != nil {
			cacheErr = fmt.Errorf("failed to read embedded gita data: %w", err)
			return
		}
		var g Gita
		if err := json.Unmarshal(data, &g); err != nil {
			cacheErr = fmt.Errorf("failed to parse gita data: %w", err)
			return
		}
		cache = &g
	})
	return cache, cacheErr
}

// GetChapter returns the chapter with the given number, or an error if not found.
func GetChapter(g *Gita, chapterNum int) (*Chapter, error) {
	for i := range g.Chapters {
		if g.Chapters[i].Chapter == chapterNum {
			return &g.Chapters[i], nil
		}
	}
	return nil, fmt.Errorf("chapter %d not found (available chapters: 1–18)", chapterNum)
}

// GetVerse returns a specific verse by chapter and verse number.
func GetVerse(g *Gita, chapterNum, verseNum int) (*VerseRef, error) {
	ch, err := GetChapter(g, chapterNum)
	if err != nil {
		return nil, err
	}
	for _, v := range ch.Verses {
		if v.Verse == verseNum {
			return &VerseRef{
				ChapterNum:   chapterNum,
				VerseNum:     verseNum,
				ChapterTitle: ch.Title,
				Text:         v.Text,
			}, nil
		}
	}
	return nil, fmt.Errorf("verse %d not found in chapter %d", verseNum, chapterNum)
}

// AllVerses returns every verse in the Gita as a flat slice of VerseRef.
func AllVerses(g *Gita) []VerseRef {
	var refs []VerseRef
	for _, ch := range g.Chapters {
		for _, v := range ch.Verses {
			refs = append(refs, VerseRef{
				ChapterNum:   ch.Chapter,
				VerseNum:     v.Verse,
				ChapterTitle: ch.Title,
				Text:         v.Text,
			})
		}
	}
	return refs
}
