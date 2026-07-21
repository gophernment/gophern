---
name: gophern
description: Use when writing, editing, or reviewing a Gophern slide deck (.md presentation file for the gophern serve/export CLI) — creating new slides, adding YAML frontmatter, choosing a layout, adding speaker notes, code blocks, or asset images.
---

# Gophern Slide Decks

## Overview
A Gophern deck is one plain Markdown file (GFM via Goldmark). Slides are split by a lone `---` line. No custom DSL — everything else is standard Markdown plus optional YAML frontmatter blocks.

## Slide separator
`---` must sit alone on its own line, outside any fenced code block (```` ``` ```` or `~~~`). Anything else containing `---` (e.g. inside a code fence) does not split slides.

## Frontmatter
- **Global** (deck-wide): a YAML block that is the very first thing in the file. Recognized keys: `title, author, theme, aspectRatio`. This block IS slide 0's frontmatter — do not add an extra `---` separator after it before slide 1's content; slide 1 also inherits its `layout/background/color`.
- **Local** (per-slide): a YAML block right after a `---` separator, before that slide's content. Recognized keys: `layout, background, color`. Optional — a slide with no local frontmatter at all just uses `default` layout with no explicit background/color.
- A YAML-looking block with none of the recognized keys is treated as ordinary text, not frontmatter — don't invent new keys expecting them to work.

## Layouts
| layout | behavior |
|---|---|
| `default` | normal vertical flex |
| `cover` | centered, made for gradient title/closing slides |
| `two-cols` | CSS grid, 2 columns. Top-level `# h1` spans both columns; every other block after it auto-flows into the grid in document order — pair `## Heading` + its content, then the next `## Heading` + content, to get left/right panes |

## Speaker notes
One HTML comment `<!-- ... -->` as the **last** element in the slide block. Content after it is ignored — never put visible slide content after the notes comment.

## Code blocks
Standard fenced code blocks (any language); highlighted server-side with Chroma, no client JS needed.

## Images / assets
Reference with normal Markdown `![alt](asset/foo.png)`. Assets must live in an `asset/` folder next to the `.md` file — served at `/asset/`.

## Example (one slide, full shape)
```markdown
---
title: "Project Pitch"
author: "Team"
theme: "slate"
aspectRatio: "16:9"
layout: "cover"
background: "linear-gradient(135deg, #0f172a 0%, #1e293b 100%)"
color: "#ffffff"
---

# Project Pitch 🚀
### One-line subtitle

Press **Right Arrow** or **Space** to navigate.

<!--
Speaker note: open with the problem statement.
-->

---
layout: "two-cols"
background: "#1e293b"
color: "#f8fafc"
---

# How It Works

## Left Column
Bullet text here.

## Right Column
```go
fmt.Println("code on the other side")
```
```

## Running it
```bash
gophern serve [-port 8080] deck.md   # live view at /, presenter console at /presenter, SSE hot-reload
gophern export [-o output.pdf] deck.md   # single-file PDF export (each slide rendered as an image)
```

## Common mistakes
- Putting global frontmatter anywhere but the very first block — it's then ignored.
- A `---` inside a fenced code block accidentally splitting the slide (rare, but check fence state if a slide splits unexpectedly).
- Visible content written after the speaker-notes comment — it gets dropped.
- Assets referenced without an `asset/` subfolder next to the `.md` file — they 404.
- No scaffold/init CLI command exists yet — decks are hand-written Markdown from scratch.
