# Design Specification: Gophern (Go & htmx Markdown Presentation Engine)

**Status:** Draft (Awaiting User Review)  
**Date:** 2026-07-14  
**Author:** Antigravity (AI Assistant) & Pallat (User)

---

## 1. Overview

Gophern is a command-line tool and local server that compiles Markdown files into professional, interactive online presentations. 

It provides two primary execution modes:
1. **Interactive Local Server (`serve` mode)**: Runs a local web server displaying the presentation, with a fully synchronized **Presenter Mode** (speaker notes, current/next slide, timers) using Server-Sent Events (SSE) and `htmx`.
2. **Self-Contained Export (`export` mode)**: Packages the slides, CSS styles, JavaScript behavior, and syntax-highlighted code blocks into a single, standalone HTML file that can be shared and opened offline.

---

## 2. Architecture & File Structure

The project is implemented in Go as a single binary CLI tool. It leverages Go's built-in `go:embed` to embed frontend templates and static assets, making the binary fully portable.

### Directory Structure

```text
gophern/
├── main.go               # Entrypoint, CLI flag/command parser
├── go.mod                # Go module descriptor
├── docs/
│   └── superpowers/
│       └── specs/
│           └── 2026-07-14-gophern-presentation-design.md  # This design file
├── internal/
│   ├── parser/           # Markdown & YAML Frontmatter parsing logic
│   │   └── parser.go
│   ├── server/           # net/http handlers & SSE broadcaster
│   │   ├── server.go
│   │   └── sse.go
│   └── templates/        # HTML templates parsing & injection
│       └── templates.go
└── web/                  # Embedded frontend templates & static assets
    ├── templates/
    │   ├── presentation.html  # Main presentation screen
    │   ├── presenter.html     # Presenter console screen
    │   └── export.html        # Single-file export layout template
    └── static/
        ├── css/
        │   └── styles.css     # Premium styling, glassmorphism, 16:9 layouts
        └── js/
            └── app.js         # Client navigation, keys, scale calculations
```

### Key Libraries
- **Markdown Parsing**: `github.com/yuin/goldmark` (highly extensible, GFM compliant)
- **Frontmatter Extraction**: `github.com/yuin/goldmark-meta`
- **Code Syntax Highlighting**: `github.com/alecthomas/chroma` (Server-side rendering of themed code blocks)
- **Web Routing**: Go Standard Library `net/http`
- **SSE & Client Interactivity**: `htmx.org` (using the SSE extension)

---

## 3. Slide Parsing & Data Schema

### Delimiter & Frontmatter
- Delimiter: `\n---\n` (three hyphens on a blank line) separates slides.
- **Global Frontmatter**: The first slide block can contain a YAML block containing global presentation metadata:
  ```yaml
  title: "Go & htmx Presentation"
  author: "Gopher"
  theme: "slate"
  aspectRatio: "16:9"
  ```
- **Local (Slide-level) Frontmatter**: Subsequent slide blocks can also start with YAML frontmatter to apply specific layout, classes, or background styling:
  ```yaml
  layout: "cover"
  background: "linear-gradient(135deg, #1e3c72, #2a5298)"
  color: "#ffffff"
  ```

### Data Structures (Go)
```go
type Presentation struct {
	Title       string
	Author      string
	Theme       string
	AspectRatio string
	Slides      []Slide
}

type Slide struct {
	Index        int
	RawMarkdown  string
	HTMLContent  string
	Layout       string            // e.g. "default", "cover", "two-cols"
	Background   string            // CSS background rule
	Color        string            // CSS text color override
	CustomStyle  map[string]string // Miscellaneous key-values
	SpeakerNotes string            // Text after a special speaker notes delimiter (e.g. `??` or `note:`)
}
```

---

## 4. Frontend Design & Aesthetics

To achieve a **professional, premium presentation look**:
1. **16:9 Aspect Ratio Lock**: The slide container maintains a strict 16:9 ratio. Responsive scaling is calculated on window resize using CSS custom properties:
   ```css
   :root {
     --slide-width: 960px;
     --slide-height: 540px;
   }
   /* Scale factor dynamically injected into CSS --scale variable by JS */
   .slide-viewport {
     transform: scale(var(--scale));
     transform-origin: center center;
   }
   ```
2. **Typography**: Minimalist modern Sans-Serif stack (Inter, system-ui) optimized for legibility.
3. **Themes**: Clean Dark mode (deep slate gray background `#0f172a`, subtle card outlines, soft colorful accents) and Light mode.
4. **CSS Transitions**: Slide changes are animated using CSS hardware-accelerated transitions:
   - Active slide: `.slide.active` (opacity: 1, transform: translate(0, 0))
   - Past slides: `.slide.past` (opacity: 0, transform: translate(-100%, 0))
   - Future slides: `.slide.future` (opacity: 0, transform: translate(100%, 0))

---

## 5. Live Server Sync (SSE & htmx)

### State Management
Go server stores the presentation state:
```go
type PresentationState struct {
	CurrentIndex int
	TotalSlides  int
}
```

### Communication Flow (SSE + htmx)
1. Both the **Main View** (`/`) and the **Presenter View** (`/presenter`) connect to `/events` (SSE stream).
2. Keyboard events (left/right arrows) on either view, or button clicks in the presenter view, trigger an `htmx` request:
   - Presenter console triggers `POST /api/slide/next` or `POST /api/slide/prev` via `hx-post`.
3. The Go Server handles the POST request, updates `CurrentIndex` in `PresentationState`, and broadcasts the new state over SSE:
   - Message payload: `{"slide": 3}`
4. The SSE listener on both pages receives the message. A lightweight JS snippet updates the slide classes to activate the slide at index `3`.

---

## 6. Self-Contained Export Mode

The command:
```bash
gophern export <file.md> -o <output.html>
```
Compiles the Markdown file into a single `.html` file.

### Bundling Mechanism
- Reads the main HTML template (`web/templates/export.html`).
- Embeds the main CSS file contents inside a `<style>` block.
- Embeds Chroma's syntax highlighting styles.
- Injects the compiled list of HTML slide divs.
- Embeds the slide navigation JS code inside a `<script>` block.
- In offline mode, the presenter/SSE features are automatically disabled, but standard keyboard slide navigation works flawlessly.

---

## 7. Error Handling & Testing
- **Graceful Parsing Fallbacks**: If YAML frontmatter in a slide is malformed, log a warning but parse the rest as standard Markdown.
- **Robust SSE Reconnection**: Frontend JS will handle SSE connection drops gracefully, attempting auto-reconnect to ensure control is not lost during presentations.
- **Unit Tests**:
  - Markdown parser & slide separator splitting logic.
  - Frontmatter extractor logic.
  - HTML builder integration.
