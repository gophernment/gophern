package exporter_test

import (
	"os"
	"strings"
	"testing"

	"github.com/gophernment/gophern/internal/exporter"
)

func TestExport(t *testing.T) {
	// Create temporary markdown file
	tmpMarkdown, err := os.CreateTemp("", "test_deck_*.md")
	if err != nil {
		t.Fatalf("failed to create temp markdown file: %v", err)
	}
	defer os.Remove(tmpMarkdown.Name())

	content := `---
title: Exported Test Deck
author: Gopher
theme: slate
---
# Slide 1
Welcome to export mode.
---
# Slide 2
Here is a code block:
` + "```go\npackage main\n```" + `
`
	if _, err := tmpMarkdown.WriteString(content); err != nil {
		t.Fatalf("failed to write temp markdown file: %v", err)
	}
	tmpMarkdown.Close()

	// Temp output path
	tmpOutput, err := os.CreateTemp("", "exported_*.html")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	tmpOutput.Close()
	defer os.Remove(tmpOutput.Name())

	// Export
	err = exporter.Export(tmpMarkdown.Name(), tmpOutput.Name())
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// Read output html
	htmlBytes, err := os.ReadFile(tmpOutput.Name())
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	html := string(htmlBytes)

	// Assertions
	if !strings.Contains(html, "<title>Exported Test Deck</title>") {
		t.Error("expected html to contain title")
	}
	if !strings.Contains(html, "Welcome to export mode.") {
		t.Error("expected html to contain slide 1 content")
	}
	if !strings.Contains(html, "--slide-width") {
		t.Error("expected html to contain embedded stylesheet CSS styles")
	}
	if !strings.Contains(html, `<body style="--slide-width: 960px; --slide-height: 540px;`) {
		t.Error("expected body style to carry per-deck slide dimensions")
	}
	if !strings.Contains(html, "window.goToSlide") {
		t.Error("expected html to contain embedded navigation JavaScript code")
	}
	if !strings.Contains(html, "package") || !strings.Contains(html, "<pre") {
		t.Error("expected html to contain highlighted code block")
	}
}

func TestExportSplitLayout(t *testing.T) {
	tmpMarkdown, err := os.CreateTemp("", "test_split_deck_*.md")
	if err != nil {
		t.Fatalf("failed to create temp markdown file: %v", err)
	}
	defer os.Remove(tmpMarkdown.Name())

	content := `---
layout: "grid-4"
cols: "60/40"
rows: "70/30"
---
::tl::
Top left

::tr::
Top right

::bl::
Bottom left

::br::
Bottom right
`
	if _, err := tmpMarkdown.WriteString(content); err != nil {
		t.Fatalf("failed to write temp markdown file: %v", err)
	}
	tmpMarkdown.Close()

	tmpOutput, err := os.CreateTemp("", "exported_split_*.html")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	tmpOutput.Close()
	defer os.Remove(tmpOutput.Name())

	if err := exporter.Export(tmpMarkdown.Name(), tmpOutput.Name()); err != nil {
		t.Fatalf("export failed: %v", err)
	}

	htmlBytes, err := os.ReadFile(tmpOutput.Name())
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	html := string(htmlBytes)

	for _, area := range []string{"tl", "tr", "bl", "br"} {
		if !strings.Contains(html, "grid-area: "+area) {
			t.Errorf("expected grid-area: %s in exported html", area)
		}
	}
	if !strings.Contains(html, "grid-template-columns: 60fr 40fr") {
		t.Errorf("expected cols ratio in exported html")
	}
	if !strings.Contains(html, "grid-template-rows: 70fr 30fr") {
		t.Errorf("expected rows ratio in exported html")
	}
}

func TestExportFontFields(t *testing.T) {
	tmpMarkdown, err := os.CreateTemp("", "test_font_deck_*.md")
	if err != nil {
		t.Fatalf("failed to create temp markdown file: %v", err)
	}
	defer os.Remove(tmpMarkdown.Name())

	content := `---
title: Font Export Test
fonts:
  sans: 'Space Grotesk'
  mono: 'JetBrains Mono'
---
# Slide 1

---
headerFont: "Poppins, sans-serif"
---
# Slide 2
`
	if _, err := tmpMarkdown.WriteString(content); err != nil {
		t.Fatalf("failed to write temp markdown file: %v", err)
	}
	tmpMarkdown.Close()

	tmpOutput, err := os.CreateTemp("", "exported_font_*.html")
	if err != nil {
		t.Fatalf("failed to create temp output file: %v", err)
	}
	tmpOutput.Close()
	defer os.Remove(tmpOutput.Name())

	if err := exporter.Export(tmpMarkdown.Name(), tmpOutput.Name()); err != nil {
		t.Fatalf("export failed: %v", err)
	}

	htmlBytes, err := os.ReadFile(tmpOutput.Name())
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	html := string(htmlBytes)

	if !strings.Contains(html, "--font-sans: Space Grotesk, &#39;Inter&#39;, -apple-system, BlinkMacSystemFont, &#34;Segoe UI&#34;, Roboto, Helvetica, Arial, sans-serif;") {
		t.Errorf("expected global sans font override with fallback chain in exported html, got: %s", html)
	}
	if !strings.Contains(html, "--font-mono: JetBrains Mono, &#39;Fira Code&#39;, Consolas, Monaco, &#39;Courier New&#39;, monospace;") {
		t.Errorf("expected global mono font override with fallback chain in exported html, got: %s", html)
	}
	if !strings.Contains(html, "--font-heading: Poppins, sans-serif, &#39;Inter&#39;, -apple-system, BlinkMacSystemFont, &#34;Segoe UI&#34;, Roboto, Helvetica, Arial, sans-serif;") {
		t.Errorf("expected per-slide header font override with fallback chain in exported html, got: %s", html)
	}
	if strings.Contains(html, "fonts.googleapis.com") {
		t.Errorf("expected export output to stay self-contained (no Google Fonts network dependency), got: %s", html)
	}
}
