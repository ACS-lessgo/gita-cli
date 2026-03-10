package gita

// Verse represents a single verse of the Bhagavad Gita.
type Verse struct {
	Verse int    `json:"verse"`
	Text  string `json:"text"`
}

// Chapter represents a chapter containing multiple verses.
type Chapter struct {
	Chapter int     `json:"chapter"`
	Title   string  `json:"title"`
	Verses  []Verse `json:"verses"`
}

// Gita is the root data structure containing all chapters.
type Gita struct {
	Chapters []Chapter `json:"chapters"`
}

// VerseRef is a lightweight reference to a specific verse location.
type VerseRef struct {
	ChapterNum int
	VerseNum   int
	ChapterTitle string
	Text       string
}
