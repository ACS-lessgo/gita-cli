# gita-cli

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
