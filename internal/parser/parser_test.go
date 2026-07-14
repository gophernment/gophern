package parser

import (
	"os"
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
