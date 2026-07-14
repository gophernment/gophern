---
title: "Gophern: Go & htmx Presentation Engine"
author: "Gophernment"
theme: "slate"
aspectRatio: "16:9"
layout: "cover"
background: "linear-gradient(135deg, #0f172a 0%, #1e293b 100%)"
color: "#ffffff"
---

# Gophern 🐹
### Go & htmx Markdown Presentation Engine

Press **Right Arrow**, **Space**, or **Page Down** to navigate.

<!-- 
Welcome to Gophern! 
This presentation serves as both a live demonstration of Gophern's features and its official documentation.
-->

---
layout: "default"
background: "#0f172a"
color: "#f8fafc"
---

# Core Philosophy

Gophern is built to be a simple, lightweight, and modern presentation tool.

- **Developer First**: Write your slides in simple Markdown.
- **Aspect Ratio Lock**: Strictly maintains a professional 16:9 ratio.
- **Ultra Portable**: Serves local presentations via a Go binary, or exports everything into a single offline HTML file.
- **Real-Time Sync**: Synchronizes main view and presenter console using Server-Sent Events (SSE).

<!-- 
Explain why Gophern is a great alternative to heavy JS frameworks like Slidev or Marp when you want lightweight Go-based servers.
-->

---
layout: "two-cols"
background: "#1e293b"
color: "#f8fafc"
---

# How It Works

## Slide Parsing
Slides are separated by `---` lines.
Each slide can contain custom local frontmatter (YAML) to override colors, layouts, and backgrounds.

## Code Highlighting
Fenced code blocks are compiled server-side and styled using Chroma syntax highlighting.

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Gophern!")
}
```

<!-- 
Highlight the use of Goldmark and Chroma on the server side for maximum parsing speed and safety.
-->

---
layout: "default"
background: "#0f172a"
color: "#f8fafc"
---

# Presenter Dashboard

Run `gophern serve` and navigate to `/presenter` to open the presenter console.

- **Split Screen Previews**: Displays scaled-down previews of the current and next slides.
- **Clock & Timer**: Tracks your elapsed presentation time.
- **Speaker Notes**: Automatically extracts and displays HTML comments at the bottom of slides.
- **Sync Control**: Actions on either the presentation view or presenter console sync instantly via SSE.

<!-- 
Explain how HTMX's SSE extension connects to the server-sent event stream to update the preview screens instantly.
-->

---
layout: "default"
background: "#1e293b"
color: "#f8fafc"
---

# CLI Interface

Manage your slide decks directly from the command line.

### Start the Live Server
```bash
gophern serve [-port 8080] example.md
```

### Export for Offline Sharing
```bash
gophern export [-o output.html] example.md
```

<!-- 
Explain the two subcommands: serve starts http.ListenAndServe; export bundles all assets into a single static file.
-->

---
layout: "cover"
background: "linear-gradient(135deg, #1e3a8a 0%, #0f172a 100%)"
color: "#ffffff"
---

# Start Presenting with Gophern!
### Simple, elegant, and blazing fast.

Project Home: [github.com/gophernment/gophern](https://github.com/gophernment/gophern)

<!-- 
End of the presentation deck. Prompt the user for questions or contributions.
-->
