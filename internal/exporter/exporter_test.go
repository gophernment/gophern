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
	if !strings.Contains(html, "window.goToSlide") {
		t.Error("expected html to contain embedded navigation JavaScript code")
	}
	if !strings.Contains(html, "package") || !strings.Contains(html, "<pre") {
		t.Error("expected html to contain highlighted code block")
	}
}
