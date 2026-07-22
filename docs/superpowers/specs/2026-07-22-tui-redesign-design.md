# TUI Redesign — Design Spec

## Context

gita-cli's interactive TUI (`internal/tui`) is currently a 3-panel browser
(chapters | verse numbers | content) with inline search. A design mockup
("Gita Reader.dc.html", imported from claude.ai/design project
`c008593c-5b2c-4375-937b-870f8c300dc0`) proposes a richer flow: a Home
screen with a daily verse and resume-reading, a restructured single-pane
Reading screen with a collapsible chapter sidebar, a dedicated Search
screen, a new Bookmarks feature, a quick jump Palette, and a Help overlay.

Two things must be preserved exactly:
- **Color scheme** — no new hues; reuse the existing monochrome palette in
  `internal/tui/styles.go` (`cWhite/cBright/cMid/cDim/cDimmer/cBg/cBgPanel/
  cBorder/cHot`).
- **Splash screen** — `internal/tui/splash.go` is untouched; its
  "materializing" animation stays exactly as-is.

Only the interactive TUI (`gita` launched with no args) is affected. The
non-interactive subcommands (`verse`, `chapter`, `random`, `search`,
`quote` in `cmd/`) are out of scope and unchanged.

## Architecture

Extend the existing single-`Model` pattern rather than introducing
per-screen `tea.Model`s. A `screen` enum selects one of
Home/Reading/Search/Bookmarks; an `overlay` enum layers Palette/Help on top
of whichever screen is active — the same overlay pattern the current code
already uses for inline search. This avoids duplicating Bubble Tea
message-routing boilerplate across screens that share most of their
underlying data (`*gita.Gita`, cursor state).

To keep files focused as the model grows, split by screen instead of
growing `model.go` indefinitely:

- `model.go` — `Model` struct, `Init`/`Update`/`View`, top-level key
  dispatch (global keys + routing to the active screen's handler)
- `home.go` — Home screen draw + logic
- `reading.go` — Reading screen draw + logic (chapter sidebar + verse pane)
- `search.go` — Search screen draw + logic (replaces old inline search)
- `bookmarks.go` — Bookmarks screen draw + logic
- `palette.go` — Palette overlay draw + logic
- `help.go` — Help overlay draw (static content)
- `state.go` — persistence: load/save `persistedState` to disk
- `splash.go`, `styles.go`, `run.go` — unchanged (styles.go gets no new
  color variables, only reuses existing ones)

The old 3-panel layout (`drawChapPanel`/`drawVrsPanel`/`drawCntPanel`,
`panelChapters`/`panelVerses`/`panelContent`) is fully removed — this is a
full navigation-model replacement, not additive.

## Screens & flow

- **Splash** → unchanged 5s/keypress animation → lands on **Home**.
- **Home**: verse-of-the-day (deterministic, see Daily verse below) +
  "Continue reading" row showing last position, or a neutral first-run
  state if none saved yet. Shortcut hints for `p`/`/`/`b`/`?`.
- **Reading**: collapsible chapter sidebar (`e` toggles visibility) + a
  single verse pane. `h`/`l` = prev/next chapter (verse resets to first in
  new chapter). `j`/`k` = prev/next verse, clamped at chapter boundaries
  (does not roll into the adjacent chapter — matches today's clamp
  behavior). `m` = toggle bookmark on the currently displayed verse. No
  speaker line above the verse text — the data has no speaker field, and
  it's sometimes embedded inline in verse text; showing a fabricated label
  was rejected.
- **Search** (`/` or `3`): query bar at top, results list (left) + preview
  pane (right, showing full text of the selected result). Filters live as
  you type using the existing `gita.Search(g, query, limit)` (limit ~50
  for the list). `↑`/`↓` move selection, `Enter` opens the selected verse
  in Reading, `Esc` returns to Reading without opening anything.
- **Bookmarks** (`b` or `4`): list of saved verses, each showing ref,
  chapter title, relative save time ("2d ago"), and a text snippet.
  `↑`/`↓` select, `Enter` opens in Reading, `x` removes the selected
  bookmark (and re-saves state), `Esc` back to Reading.
- **Palette overlay** (`p`, reachable from any screen): single-line input.
  If the input parses as `N` or `N.M` (chapter or chapter.verse), jump
  directly on `Enter`. Otherwise treat it as free text and fuzzy-match
  against chapter titles and verse text via `gita.Search`, capped at ~8
  results, mixing chapter and verse hits. `↑`/`↓` select, `Enter` goes,
  `Esc` dismisses without navigating.
- **Help overlay** (`?`): static, grouped keybinding reference. Only
  documents keys that actually ship in this design — no placeholder
  bindings invented to match the mockup's exact groups.
- **Global keys** (work from any screen unless a text field has focus):
  `1`/`2`/`3`/`4` jump straight to Home/Reading/Search/Bookmarks. `q` /
  `ctrl+c` saves state and quits. `Esc` dismisses an open overlay; if no
  overlay is open, returns to Reading (matches mockup behavior).

## Data & persistence

New `internal/tui/state.go`:

```go
type persistedState struct {
    LastChapter, LastVerse int
    Bookmarks               []bookmark
}
type bookmark struct {
    ChapterNum, VerseNum int
    SavedAt              time.Time
}
```

- Location: `filepath.Join(os.UserConfigDir(), "gita-cli", "state.json")`.
  Directory created with `0o755`, file with `0o644` on first save.
- Loaded once at startup via `Model.Init`/`New`; a missing file is not an
  error — start from zero-value `persistedState{}`.
- Saved on: `q`/`ctrl+c` quit, and immediately after any bookmark
  add/remove (so a killed process doesn't lose a bookmark). Last-read
  position is updated in memory on every verse navigation but only
  written to disk at save points (quit or bookmark change) — avoids a
  disk write on every keystroke while still persisting reliably on normal
  exit.
- Save/load errors (permissions, disk full, corrupt JSON) never crash the
  TUI: load failure falls back to zero-value state; save failure sets
  `m.status` to a one-line warning and continues.
- Daily verse selection: flatten all verses into a single ordered list
  (chapter order, then verse order — same order `gita.Chapters` already
  provides), hash today's date as `YYYYMMDD` to an `int`, `index :=
  hash % len(allVerses)`. Deterministic per calendar day, same verse for
  everyone on a given day, no state needed.

## Color mapping

No new colors. Existing `styles.go` palette is reused everywhere the
mockup shows an "accent" hue:

- Mockup accent (`--color-accent-*`, used for active/emphasis) → `cWhite`
  / `cHot` (bold/bright text, active borders — same as today's active
  panel border).
- Mockup secondary/muted text → `cMid` / `cDim` / `cDimmer` (same mapping
  as today's inactive panel text).
- Backgrounds stay `cBg` (app background) / `cBgPanel` (panel background)
  exactly as today.

## Error handling

- State load/save: see Persistence above — never fatal.
- Search/palette with zero matches: render an empty results list, not an
  error message.
- No new failure modes beyond the existing `gita.Load()` embedded-data
  read at startup (already handled in `run.go`).

## Testing

- `internal/tui` currently has no tests. This redesign doesn't take on
  full TUI coverage, but `state.go`'s pure logic — load/save round-trip,
  daily-verse index hashing, bookmark add/remove — is straightforward to
  unit test and gets `state_test.go`, consistent with how `internal/gita`
  is tested today.
- Manual verification: launch `gita`, exercise all screens/overlays
  (Home, Reading, Search, Bookmarks, Palette, Help), confirm a bookmark
  and last-read position survive a restart, confirm the splash animation
  is unchanged.
