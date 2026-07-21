---
title: "Gophern Usage Guide"
author: "Gophernment"
theme: "slate"
aspectRatio: "16:9"
fonts:
  sans: "Poppins, 'Noto Sans Thai'"
showControls: true
showSlideNumber: true
layout: "cover"
background: "linear-gradient(135deg, #0f172a 0%, #1e293b 100%)"
color: "#ffffff"
---

# Gophern Usage Guide 🐹
### A complete, runnable tour of every feature

This deck is a real `.md` file — run it with `gophern serve USAGE.md`
and navigate with **→**, **Space**, or **Page Down**. A fullscreen
button sits in the bottom-right corner; the prev/next buttons and slide
number are off by default (turned on for this deck — see slide 18).

<!--
This is USAGE.md: a hands-on manual. Every syntax shown on a slide is the
exact syntax that produced that very slide — open the raw file side by
side with the rendered output to see how each piece works.
-->

---
layout: "default"
background: "#0f172a"
color: "#f8fafc"
---

# 1. Installing & Running

### Install
```bash
go install github.com/gophernment/gophern@latest
# or clone the repo and `go build .`
```

### Two commands, that's it
```bash
gophern serve [-port 8080] USAGE.md   # live server + hot reload
gophern export [-o output.html] USAGE.md  # single offline HTML file
```

`serve` starts a local HTTP server with a `/presenter` console and
auto-reloads whenever you save the `.md` file. `export` bundles the CSS,
JS, and every slide into one self-contained HTML file you can email or
host anywhere — no server required to view it.

<!--
Mention that `export` never needs network access at view time — everything
is inlined into the single output file, including the syntax-highlighted
code blocks.
-->

---
layout: "default"
background: "#1e293b"
color: "#f8fafc"
---

# 2. Slide Delimiter

Slides are separated by a line containing exactly three hyphens, `---`,
on its own line.

```markdown
# Slide One
Content here.

---

# Slide Two
More content.
```

A `---` **inside a fenced code block** (between ` ``` ` or `~~~`) is never
treated as a delimiter — it's safe to show shell output or YAML examples
containing `---` inside your slides.

---
layout: "default"
background: "#0f172a"
color: "#f8fafc"
---

# 3. Frontmatter: Global vs. Local

Every slide can start with a YAML frontmatter block (`---` ... `---`)
setting properties for that slide.

- **Global**: put it on **Slide 0** (the very first block in the file).
  It sets defaults — `title`, `author`, `theme`, `aspectRatio`, `fonts` —
  that apply to the whole deck.
- **Local**: any later slide's own frontmatter block overrides `layout`,
  `background`, `color`, `headerFont`, etc. for **that slide only**.

```markdown
---
title: "My Deck"
fonts:
  sans: 'Space Grotesk'
---
# Slide 0 inherits both title and font

---
background: "#1e293b"
---
# Slide 1 only overrides background
```

<!--
Global frontmatter must be the very first block in the file (before any
slide content). If the file has no frontmatter at all, gophern falls back
to sensible defaults: title "Presentation", aspectRatio "16:9".
-->

---
layout: "two-cols"
background: "#0f172a"
color: "#f8fafc"
---

# 4. Per-Slide Config Reference

## Local frontmatter keys
- `layout` — one of `default`, `cover`, `two-cols`, `split-h`, `split-v`, `split-3`, `grid-4`
- `background` — any CSS `background` value (color, gradient, image URL)
- `color` — the slide's base text color
- `headerFont` — overrides just this slide's `<h1>` font

## Split-layout-only keys
- `ratio` — e.g. `"60/40"` for `split-h`/`split-v`, `"30/40/30"` for `split-3`
- `cols` / `rows` — independent ratios, `grid-4` only

Every key above goes in a slide's **own** frontmatter block and affects
**that slide only** — none of them are inherited by later slides.

---
layout: "default"
background: "#1e293b"
color: "#f8fafc"
---

# 5. Layout: `default`

The normal vertical flow layout — headings, paragraphs, lists, and code
blocks stack top to bottom. This is what you get when a slide has no
`layout:` field at all, or `layout: "default"` explicitly.

- Bulleted lists work as expected
- So do **bold**, *italic*, and `inline code`
- Fenced code blocks get server-side syntax highlighting (see slide 9)

```markdown
---
layout: "default"
---
# Any Title
Regular markdown content goes here.
```

---
layout: "cover"
background: "linear-gradient(135deg, #1e3a8a 0%, #0f172a 100%)"
color: "#ffffff"
---

# 6. Layout: `cover`
### Centered, large title — built for opening/closing slides

```markdown
---
layout: "cover"
background: "linear-gradient(135deg, #1e3a8a 0%, #0f172a 100%)"
color: "#ffffff"
---
# Title
### Subtitle
```

<!--
cover centers everything both horizontally and vertically and applies a
larger, gradient-accented heading style — this very slide is the example.
-->

---
layout: "two-cols"
background: "#0f172a"
color: "#f8fafc"
---

# 7. Layout: `two-cols`

## Left-ish content
A CSS-grid two-column layout. Headings and paragraphs auto-flow into the
grid — good for a quick side-by-side without needing `::left::`/`::right::`
markers.

## Right-ish content
For precise control over what goes in which column, prefer the newer
`split-h` layout (slide 11) instead — it lets you address each side
explicitly by name.

```markdown
---
layout: "two-cols"
---
# Title
## Section A
...
## Section B
...
```

---
layout: "default"
background: "#1e293b"
color: "#f8fafc"
---

# 8. Speaker Notes

Add an HTML comment at the very bottom of a slide's content — it never
renders on screen, but shows up in the **presenter console**
(`gophern serve` → open `/presenter`).

```markdown
# Slide Title
Visible content.

<!--
Only you see this, in the presenter view's notes panel.
-->
```

A comment only counts as speaker notes if nothing but whitespace follows
it — an HTML comment used mid-content (or inside a code block) is left
alone and rendered normally.

---
layout: "default"
background: "#0f172a"
color: "#f8fafc"
---

# 9. Code Highlighting

Fenced code blocks are compiled server-side with Chroma — no client-side
JavaScript highlighter, no flash of unstyled code.

````markdown
```go
func main() {
    fmt.Println("gophern")
}
```
````

```go
func main() {
    fmt.Println("gophern")
}
```

---
layout: "two-cols"
background: "#0f172a"
color: "#f8fafc"
---

# 10. Inline HTML & Styling

## Raw HTML passes through
Gophern's markdown renderer allows raw HTML tags directly in your slide
content — for one-off styling that frontmatter fields don't cover.

```markdown
<span style="color:#f472b6; font-weight:700;">
  Custom inline color
</span>

<div style="text-align:right; border:1px solid #333;">
  A hand-styled box
</div>
```

## Use sparingly
<span style="color:#f472b6; font-weight:700;">This text is pink</span> via
the exact snippet above. Prefer `background`/`color`/`headerFont`
frontmatter for whole-slide styling — reach for inline HTML only for
small, one-off cases those fields don't reach.

---
layout: "split-h"
ratio: "55/45"
background: "#1e293b"
color: "#f8fafc"
---

# 11. Split Layouts: `split-h`

::left::
Divide a slide into named regions with `::name::` markers. `split-h`
gives you `::left::` and `::right::`, each rendered independently — code
highlighting, lists, everything works the same as the main slide body.

Control the width split with `ratio: "55/45"` (any two numbers that sum
to whatever you like — they're read as relative weights).

::right::
```markdown
---
layout: "split-h"
ratio: "55/45"
---
::left::
Left content.

::right::
Right content.
```

---
layout: "split-v"
ratio: "40/60"
background: "#0f172a"
color: "#f8fafc"
---

# 12. Split Layouts: `split-v`

::top::
Same idea as `split-h`, but stacked vertically. Regions are `::top::` and
`::bottom::`, and `ratio` controls the height split instead of width.

::bottom::
```markdown
---
layout: "split-v"
ratio: "40/60"
---
::top::
Top content.

::bottom::
Bottom content.
```

---
layout: "split-3"
ratio: "30/40/30"
background: "#1e293b"
color: "#f8fafc"
---

# 13. Split Layouts: `split-3`

::left::
Three columns: `::left::`, `::center::`, `::right::`.

::center::
`ratio` takes three numbers here, one per column — this slide uses
`"30/40/30"` so the center column is a bit wider.

::right::
```markdown
layout: "split-3"
ratio: "30/40/30"
```

---
layout: "grid-4"
cols: "50/50"
rows: "50/50"
background: "#0f172a"
color: "#f8fafc"
---

# 14. Split Layouts: `grid-4`

::tl::
A 2×2 grid: `::tl::`, `::tr::`, `::bl::`, `::br::`.

::tr::
Two independent ratios — `cols` for the column split, `rows` for the row
split — since this layout has two axes at once.

::bl::
```markdown
layout: "grid-4"
cols: "60/40"
rows: "70/30"
```

::br::
Any region can be left out entirely — the grid cell just stays empty.

---
layout: "default"
background: "#1e293b"
color: "#f8fafc"
---

# 15. Custom Fonts: the basics

Set the deck's main fonts once, globally:

```markdown
---
fonts:
  sans: 'Space Grotesk'
  mono: 'JetBrains Mono'
---
```

`gophern serve` and `/presenter` automatically fetch the named font from
Google Fonts — no extra setup. `gophern export` does **not** fetch fonts
(to stay a fully offline, self-contained file); an unavailable font falls
back to the built-in stack (`Inter` / `Fira Code`) instead.

---
headerFont: "Poppins, sans-serif"
background: "#0f172a"
color: "#f8fafc"
---

# 16. Custom Fonts: per-slide heading

This heading uses `headerFont`, set only on this slide's local frontmatter.

Override just **this slide's heading** with `headerFont` in its local
frontmatter — the body text and every other slide keep using the deck's
main font.

```markdown
---
headerFont: "Poppins, sans-serif"
---
# Only this heading uses Poppins
```

---
layout: "default"
background: "#1e293b"
color: "#f8fafc"
---

# 17. Custom Fonts: หลายฟอนต์ / Thai fallback

ข้อความไทยผสมกับ English ในสไลด์เดียวกัน — สไลด์นี้ (และทั้งเดค) ใช้
`fonts.sans: "Poppins, 'Noto Sans Thai'"` ที่ตั้งไว้ใน **global frontmatter**
บนสไลด์ปก (Slide 0) ดังนั้นข้อความไทยตรงนี้จึงแสดงถูกต้องแบบ live จริง

```markdown
---
fonts:
  sans: "Poppins, 'Noto Sans Thai'"
---
```

Any font field accepts a full comma-separated CSS font stack, not just
one name — list a script-specific fallback after your primary font.

<!--
Note: `fonts:` is a Presentation-level (global-only) field — it must live
in Slide 0's frontmatter to take effect, unlike `background`/`color`/
`headerFont` which work per-slide. This deck sets it once on the cover
slide, so every slide (including this one) already renders with the Thai
fallback active.
-->

---
layout: "default"
background: "#0f172a"
color: "#f8fafc"
---

# 18. Show/Hide Navigation Controls

The prev/next buttons and the `1 / N` slide-number indicator in the
**Presentation View** (`serve`, not `/presenter`) are hidden by default.
This deck turns them on via its global frontmatter:

```markdown
---
showControls: true
showSlideNumber: true
---
```

Leave them unset (or `false`) for a chrome-free presentation. The
fullscreen button in the bottom-right is always visible either way.

<!--
Both keys default to false (hidden) and only affect the live serve
Presentation View — the Presenter Console and exported HTML always show
their own controls unconditionally.
-->

---
layout: "cover"
background: "linear-gradient(135deg, #1e3a8a 0%, #0f172a 100%)"
color: "#ffffff"
---

# That's the whole toolkit
### Markdown in, presentation out.

Project Home: [github.com/gophernment/gophern](https://github.com/gophernment/gophern)

Try editing this file and watch `gophern serve` hot-reload instantly.

<!--
Closing slide. Encourage the reader to open USAGE.md next to the rendered
output and experiment — every syntax on every prior slide is copy-pasteable.
-->
