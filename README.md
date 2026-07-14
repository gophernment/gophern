# Gophern 🐹

**Gophern** is a professional, local Markdown presentation engine built with Go and `htmx`. It compiles standard Markdown files into sleek, interactive online slideshows featuring a synchronized presenter console, real-time Server-Sent Events (SSE) state synchronization, and an offline single-file exporter.

---

## Features

- **Blazing Fast Markdown Rendering**: Uses Goldmark for GFM-compliant markdown compiling.
- **Server-Side Code Highlighting**: Syntax highlights code blocks using Chroma, compiling them into inline CSS styles for zero external dependency rendering.
- **Aspect Ratio Lock**: Strictly maintains a professional 16:9 layout viewport, auto-scaling to fit the browser window.
- **Presenter Dashboard**: Features a clock, elapsed timer, current/next slide previews, and real-time speaker notes display.
- **SSE Real-Time Sync**: Synchronizes slide navigation in real-time between the main viewer and presenter console.
- **Self-Contained Export**: Bundles slides, CSS styles, and JS behavior into a single, portable HTML file that can be opened offline anywhere.

---

## Installation

Ensure you have [Go](https://go.dev/) (version 1.22 or higher) installed. Clone the repository and build the binary:

```bash
git clone https://github.com/gophernment/gophern.git
cd gophern
go build -o gophern main.go
```

---

## Usage

Gophern provides two main subcommands: `serve` and `export`.

### 1. Run Live Presentation Server (`serve`)
Start the local HTTP server to display slides and enable presenter synchronization:

```bash
./gophern serve [-port 8080] example.md
```

- Open the **Presentation View** at `http://localhost:8080/`.
- Open the **Presenter Console** at `http://localhost:8080/presenter`.

Navigating slides on either window (using arrow keys, space, or buttons) will automatically sync the other window instantly via Server-Sent Events.

### 2. Export Standalone Slide Deck (`export`)
Export the presentation into a single self-contained HTML file for distribution or offline use:

```bash
./gophern export [-o output.html] example.md
```

The output file has all layout CSS and navigation JS bundled inline. Double-click it to run it directly from your file system (`file://` protocol) without needing a server.

---

## Presentation Syntax

### Slide Delimiter
Slides are separated by three hyphens (`---`) on a blank line:

```markdown
# Slide One
Content goes here.

---

# Slide Two
Another slide.
```

### Global & Local Frontmatter
You can define slide properties using YAML frontmatter blocks at the beginning of the slides.
- **Global settings** should be placed on the first slide (Slide 0).
- **Local settings** override themes, layouts, or backgrounds for individual slides.

```markdown
---
title: "Project Pitch"
theme: "slate"
layout: "cover"
background: "linear-gradient(135deg, #0f172a 0%, #1e293b 100%)"
color: "#ffffff"
---
# Main Cover Slide
```

### Supported Layouts
- `default`: Normal vertical flex layout.
- `cover`: Centered cover page layout with gradient background support.
- `two-cols`: Dual-column layout (useful for side-by-side text/images or text/code blocks).

### Speaker Notes
Add speaker notes inside an HTML comment (`<!-- ... -->`) placed at the very bottom of the slide block:

```markdown
# Slide Title
Slide content here.

<!-- 
Remember to mention the architecture schema here!
-->
```

---

## License

This project is licensed under the [MIT License](LICENSE).
