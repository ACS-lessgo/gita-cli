# 🕉 gita-cli

> Access the timeless wisdom of the **Bhagavad Gita** directly from your terminal.

A fast, beautifully formatted, open-source CLI tool built in Go — inspired by the simplicity of [christ-cli](https://github.com/whoisyurii/christ-cli).

---

## ✨ Features

| Command | Description |
|---|---|
| `gita verse <chapter> <verse>` | Retrieve a specific verse |
| `gita chapter <number>` | Display all verses in a chapter |
| `gita random` | Show a random verse |
| `gita search <keyword>` | Search verses by keyword |
| `gita quote` | Display an inspiring daily quote |

- 🎨 **Beautiful terminal output** with colors and borders (via `lipgloss`)
- ⚡ **Embedded data** — no external API calls, works offline
- 🔍 **Fast keyword search** across all verses
- 🧪 **Unit tested** core functions
- 🛡️ **Error handling** for invalid chapters/verses

---

## 📦 Installation

### From source (Go 1.22+)

```bash
git clone https://github.com/whoisyurii/gita-cli.git
cd gita-cli
go build -o gita .
```

Then move the binary to your `$PATH`:

```bash
sudo mv gita /usr/local/bin/
```

### Using `go install`

```bash
go install github.com/whoisyurii/gita-cli@latest
```

---

## 🏗️ Build Instructions

**Prerequisites:** Go 1.22 or later

```bash
# Clone the repo
git clone https://github.com/whoisyurii/gita-cli.git
cd gita-cli

# Download dependencies
go mod tidy

# Build
go build -o gita .

# Run
./gita --help
```

---

## 🚀 Usage

### Get a specific verse

```bash
gita verse 2 47
```

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Chapter 2 • Verse 47  — The Yoga of Knowledge
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  You have a right to perform your prescribed duty, but you
  are not entitled to the fruits of action. Never consider
  yourself the cause of the results of your activities, and
  never be attached to not doing your duty.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### View a full chapter

```bash
gita chapter 2
```

### Get a random verse

```bash
gita random
```

### Search verses by keyword

```bash
gita search "duty"
gita search "soul" --limit 5
gita search "lust"
```

### Get an inspiring daily quote

```bash
gita quote
```

```
  🕉  Bhagavad Gita  Daily Quote

  ╭──────────────────────────────────────────────────────────────────╮
  │                                                                  │
  │   "For the soul there is never birth nor death at any time.      │
  │   It has not come into being, does not come into being, and      │
  │   will not come into being. It is unborn, eternal, ever-         │
  │   existing and primeval. It is not slain when the body           │
  │   is slain."                                                     │
  │                                                                  │
  │              — Chapter 2, Verse 20 · The Yoga of Knowledge       │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯
```

---

## 📁 Project Structure

```
gita-cli/
├── cmd/
│   ├── root.go       # Cobra root command + shared styles/helpers
│   ├── verse.go      # gita verse <chapter> <verse>
│   ├── chapter.go    # gita chapter <number>
│   ├── random.go     # gita random
│   ├── search.go     # gita search <keyword>
│   └── quote.go      # gita quote
├── internal/
│   └── gita/
│       ├── verse.go        # Core data structures
│       ├── loader.go       # Embedded data loading + caching
│       ├── search.go       # Keyword search logic
│       ├── data/
│       │   └── gita.json   # Embedded verse dataset
│       └── gita_test.go    # Unit tests
├── data/
│   └── gita.json     # Source JSON dataset (also embedded)
├── main.go
├── go.mod
├── go.sum
└── README.md
```

---

## 🧪 Running Tests

```bash
go test ./...
```

Or with verbose output:

```bash
go test -v ./internal/gita/...
```

---

## 📖 Data Format

Verses are stored in `data/gita.json` and embedded at compile time:

```json
{
  "chapters": [
    {
      "chapter": 2,
      "title": "The Yoga of Knowledge",
      "verses": [
        {
          "verse": 47,
          "text": "You have a right to perform your prescribed duty..."
        }
      ]
    }
  ]
}
```

The dataset currently includes **50+ curated verses** from key chapters. Contributions to expand the dataset are welcome!

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/add-verses`
3. Commit your changes: `git commit -m 'feat: add more verses to dataset'`
4. Push to the branch: `git push origin feature/add-verses`
5. Open a Pull Request

### Ways to contribute

- 📜 **Expand the dataset** — Add more verses to `data/gita.json`
- 🌐 **Translations** — Add support for multiple language translations
- ✨ **New features** — Bookmarks, daily verse notifications, export to PDF
- 🐛 **Bug fixes** — Report or fix issues

---

## 🛠️ Dependencies

| Package | Purpose |
|---|---|
| [`spf13/cobra`](https://github.com/spf13/cobra) | CLI framework |
| [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss) | Terminal styling |

---

## 📜 License

MIT License — see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- Inspired by [christ-cli](https://github.com/whoisyurii/christ-cli)
- Verse content sourced from the public domain translation of the Bhagavad Gita As It Is

---

*"You have a right to perform your prescribed duty, but you are not entitled to the fruits of action."* — Bhagavad Gita 2.47

---

## 🖥️ Interactive TUI

Run `gita` with no arguments to launch the full-screen interactive browser:

```
┌─────────────────────┐ ┌────────────┐ ┌──────────────────────────────────────────────┐
│ 🕉  Chapters        │ │ Verses     │ │ Chapter 2 — The Yoga of Knowledge • Verse 47 │
│─────────────────────│ │────────────│ │──────────────────────────────────────────────│
│ Ch 1   Arjuna's ... │ │  1         │ │                                              │
│ Ch 2   Yoga of K... │ │  2         │ │ Chapter 2: The Yoga of Knowledge             │
│▶Ch 3   Yoga of A... │ │  8         │ │                                              │
│ Ch 4   Yoga of W... │ │▶ 47        │ │ Verse 47                                     │
│ Ch 5   ...          │ │  48        │ │                                              │
│ Ch 6   ...          │ │  55        │ │ You have a right to perform your prescribed  │
│ ...                 │ │  62        │ │ duty, but you are not entitled to the fruits │
│                     │ │  63        │ │ of action...                                 │
└─────────────────────┘ └────────────┘ └──────────────────────────────────────────────┘
 ←→ panels  ↑↓ navigate  Enter select  / search  g/G top/bottom  q quit
```

### TUI Key Bindings

| Key | Action |
|---|---|
| `←` / `→` or `h` / `l` | Switch panels |
| `↑` / `↓` or `k` / `j` | Navigate items |
| `Enter` or `Space` | Move focus right |
| `g` / `G` | Jump to top / bottom |
| `/` | Open search |
| `n` / `N` | Next / previous search result |
| `Esc` | Clear search |
| `q` | Quit |
