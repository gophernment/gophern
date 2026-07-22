package exporter_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gophernment/gophern/internal/exporter"
)

func TestExportHTML_ProducesSelfContainedFile(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "deck.md")

	content := `---
title: Exported Test Deck
fonts:
  sans: 'Space Grotesk'
showControls: true
showSlideNumber: true
---
# Slide 1
Hello world.
---
fragments: true
---
# Slide 2
- Alpha
- Beta
`
	if err := os.WriteFile(mdPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write markdown: %v", err)
	}

	outPath := filepath.Join(dir, "out.html")
	if err := exporter.ExportHTML(mdPath, outPath); err != nil {
		t.Fatalf("ExportHTML failed: %v", err)
	}

	htmlBytes, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output html: %v", err)
	}
	html := string(htmlBytes)

	if !strings.Contains(html, "<title>Exported Test Deck</title>") {
		t.Errorf("expected title in output, got: %s", html)
	}
	if !strings.Contains(html, "Hello world.") {
		t.Errorf("expected slide 1 content in output")
	}
	if !strings.Contains(html, `class="fragment" data-fragment-index="0"`) {
		t.Errorf("expected fragment classing to carry through to the export, got: %s", html)
	}
	// CSS and JS must be inlined (self-contained), not linked externally.
	if strings.Contains(html, `href="/static/css`) || strings.Contains(html, `src="/static/js`) {
		t.Errorf("expected inlined CSS/JS, found external /static/ reference")
	}
	if !strings.Contains(html, "<style>") || !strings.Contains(html, "<script>") {
		t.Errorf("expected inline <style> and <script> blocks in output")
	}
	// showControls/showSlideNumber true must render the nav buttons/number.
	if !strings.Contains(html, `id="btn-prev"`) || !strings.Contains(html, `id="slide-number"`) {
		t.Errorf("expected nav controls and slide number to render when enabled")
	}
	if !strings.Contains(html, `id="btn-fullscreen"`) {
		t.Errorf("expected fullscreen button to always render")
	}
}

func TestExportHTML_ControlsHiddenByDefault(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "deck.md")
	if err := os.WriteFile(mdPath, []byte("# Only slide\n"), 0o644); err != nil {
		t.Fatalf("failed to write markdown: %v", err)
	}

	outPath := filepath.Join(dir, "out.html")
	if err := exporter.ExportHTML(mdPath, outPath); err != nil {
		t.Fatalf("ExportHTML failed: %v", err)
	}

	html, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output html: %v", err)
	}
	if strings.Contains(string(html), `id="btn-prev"`) || strings.Contains(string(html), `id="slide-number"`) {
		t.Errorf("expected nav controls/slide number hidden by default, got: %s", html)
	}
}
