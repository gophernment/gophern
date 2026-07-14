package parser

import (
	"os"
	"strings"
	"testing"
)

func TestParseMarkdown(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "slides-*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := `---
title: Spec Test
author: Gopher
theme: slate
---
# Slide 1 Content

---
layout: cover
background: blue
---
# Slide 2 Content
`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	pres, err := ParseMarkdownFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseMarkdownFile failed: %v", err)
	}

	if pres.Title != "Spec Test" {
		t.Errorf("Expected Title 'Spec Test', got '%s'", pres.Title)
	}
	if pres.Author != "Gopher" {
		t.Errorf("Expected Author 'Gopher', got '%s'", pres.Author)
	}
	if pres.Theme != "slate" {
		t.Errorf("Expected Theme 'slate', got '%s'", pres.Theme)
	}
	if len(pres.Slides) != 2 {
		t.Fatalf("Expected 2 slides, got %d", len(pres.Slides))
	}
	if pres.Slides[0].Index != 0 || pres.Slides[1].Index != 1 {
		t.Errorf("Incorrect slide indices")
	}
	if pres.Slides[1].Layout != "cover" {
		t.Errorf("Expected Layout 'cover', got '%s'", pres.Slides[1].Layout)
	}
	if pres.Slides[1].Background != "blue" {
		t.Errorf("Expected Background 'blue', got '%s'", pres.Slides[1].Background)
	}
}

func TestParseMarkdownNoGlobalFrontmatter(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "slides-*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := `# Slide 1 Content

---
layout: cover
background: red
---
# Slide 2 Content
`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	pres, err := ParseMarkdownFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseMarkdownFile failed: %v", err)
	}

	if pres.Title != "Presentation" { // default
		t.Errorf("Expected Title 'Presentation', got '%s'", pres.Title)
	}
	if len(pres.Slides) != 2 {
		t.Fatalf("Expected 2 slides, got %d", len(pres.Slides))
	}
	if pres.Slides[0].RawMarkdown != "# Slide 1 Content" {
		t.Errorf("Expected first slide content '# Slide 1 Content', got %q", pres.Slides[0].RawMarkdown)
	}
	if pres.Slides[1].Layout != "cover" || pres.Slides[1].Background != "red" {
		t.Errorf("Incorrect slide 2 attributes: layout=%s, bg=%s", pres.Slides[1].Layout, pres.Slides[1].Background)
	}
	if pres.Slides[1].RawMarkdown != "# Slide 2 Content" {
		t.Errorf("Expected slide 2 content '# Slide 2 Content', got %q", pres.Slides[1].RawMarkdown)
	}
}

func TestSpeakerNotes(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "slides-*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := `---
title: Notes Test
---
# Slide 1

Some content here.

<!--
These are speaker notes for slide 1.
Multiple lines allowed.
-->

---
# Slide 2
No notes here.
`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	pres, err := ParseMarkdownFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseMarkdownFile failed: %v", err)
	}

	if len(pres.Slides) != 2 {
		t.Fatalf("Expected 2 slides, got %d", len(pres.Slides))
	}

	s1 := pres.Slides[0]
	expectedNotes := "These are speaker notes for slide 1.\nMultiple lines allowed."
	if s1.SpeakerNotes != expectedNotes {
		t.Errorf("Expected speaker notes %q, got %q", expectedNotes, s1.SpeakerNotes)
	}
	expectedMarkdown := "# Slide 1\n\nSome content here."
	if s1.RawMarkdown != expectedMarkdown {
		t.Errorf("Expected clean markdown %q, got %q", expectedMarkdown, s1.RawMarkdown)
	}

	s2 := pres.Slides[1]
	if s2.SpeakerNotes != "" {
		t.Errorf("Expected empty speaker notes for slide 2, got %q", s2.SpeakerNotes)
	}
}

func TestNonExistentFile(t *testing.T) {
	_, err := ParseMarkdownFile("does_not_exist.md")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestParserFixes(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "slides-fixes-*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := `---
title: Global Settings
layout: first-layout
background: first-bg
color: first-color
---
# First Slide

Note: test this colon

---
# Second Slide
` + "```go\n" + `
func main() {
	// Some code
	// ---
}
` + "```\n" + `
---
layout: third-layout
---
# Third Slide
`

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	pres, err := ParseMarkdownFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseMarkdownFile failed: %v", err)
	}

	if len(pres.Slides) != 3 {
		t.Fatalf("Expected 3 slides, got %d", len(pres.Slides))
	}

	// 1. Slide 0 global frontmatter inheritance
	s0 := pres.Slides[0]
	if s0.Layout != "first-layout" {
		t.Errorf("Expected s0 Layout 'first-layout', got '%s'", s0.Layout)
	}
	if s0.Background != "first-bg" {
		t.Errorf("Expected s0 Background 'first-bg', got '%s'", s0.Background)
	}
	if s0.Color != "first-color" {
		t.Errorf("Expected s0 Color 'first-color', got '%s'", s0.Color)
	}

	// 2. Slide content containing colons not parsed as frontmatter
	if !strings.Contains(s0.RawMarkdown, "Note: test this colon") {
		t.Errorf("Expected s0 RawMarkdown to contain 'Note: test this colon', got %q", s0.RawMarkdown)
	}

	// 3. Fenced code block with '---' inside it
	s1 := pres.Slides[1]
	if !strings.Contains(s1.RawMarkdown, "---") {
		t.Errorf("Expected s1 RawMarkdown to preserve '---' inside code block, got %q", s1.RawMarkdown)
	}

	// 4. HTMLContent conversion check
	if !strings.Contains(s0.HTMLContent, "<h1>First Slide</h1>") {
		t.Errorf("Expected s0 HTMLContent to contain '<h1>First Slide</h1>', got %q", s0.HTMLContent)
	}
}

