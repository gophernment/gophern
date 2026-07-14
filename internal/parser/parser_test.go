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

func TestTask2ReviewIssues(t *testing.T) {
	// 1. A slide frontmatter with comments, multi-line string, and lists parses correctly.
	t.Run("frontmatter comments, multi-line, lists", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "slides-yaml-*.md")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		content := `# Slide 1
---
layout: slide-with-yaml-features
# This is a comment in frontmatter
comments:
  - comment 1
  - comment 2
notes: |
  This is a multi-line
  string
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

		if len(pres.Slides) != 2 {
			t.Fatalf("Expected 2 slides, got %d", len(pres.Slides))
		}

		s0 := pres.Slides[0]
		if s0.RawMarkdown != "# Slide 1" {
			t.Errorf("Expected s0 RawMarkdown '# Slide 1', got %q", s0.RawMarkdown)
		}

		s1 := pres.Slides[1]
		if s1.Layout != "slide-with-yaml-features" {
			t.Errorf("Expected s1 Layout 'slide-with-yaml-features', got %q", s1.Layout)
		}
		if s1.RawMarkdown != "# Slide 2 Content" {
			t.Errorf("Expected s1 RawMarkdown '# Slide 2 Content', got %q", s1.RawMarkdown)
		}
	})

	// 2. A document ending with a frontmatter-only slide (no content and no trailing separator)
	// correctly creates the last slide with that frontmatter and empty content.
	t.Run("ending with frontmatter-only slide", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "slides-end-*.md")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		content := `# Slide 1 Content
---
layout: end-slide
background: black`
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

		s0 := pres.Slides[0]
		if s0.RawMarkdown != "# Slide 1 Content" {
			t.Errorf("Expected s0 RawMarkdown '# Slide 1 Content', got %q", s0.RawMarkdown)
		}

		s1 := pres.Slides[1]
		if s1.Layout != "end-slide" {
			t.Errorf("Expected s1 Layout 'end-slide', got %q", s1.Layout)
		}
		if s1.Background != "black" {
			t.Errorf("Expected s1 Background 'black', got %q", s1.Background)
		}
		if s1.RawMarkdown != "" {
			t.Errorf("Expected s1 RawMarkdown to be empty, got %q", s1.RawMarkdown)
		}
	})

	// 3. Empty/blank slides are not skipped and are indexed correctly.
	t.Run("empty blank slides indexed correctly", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "slides-empty-*.md")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		content := `# Slide 1
---
---
# Slide 3`
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

		s0 := pres.Slides[0]
		if s0.Index != 0 || s0.RawMarkdown != "# Slide 1" {
			t.Errorf("Expected slide 0 with content '# Slide 1', got index=%d, content=%q", s0.Index, s0.RawMarkdown)
		}

		s1 := pres.Slides[1]
		if s1.Index != 1 || s1.RawMarkdown != "" {
			t.Errorf("Expected slide 1 to be blank, got index=%d, content=%q", s1.Index, s1.RawMarkdown)
		}

		s2 := pres.Slides[2]
		if s2.Index != 2 || s2.RawMarkdown != "# Slide 3" {
			t.Errorf("Expected slide 2 with content '# Slide 3', got index=%d, content=%q", s2.Index, s2.RawMarkdown)
		}
	})

	// 4. Fenced code blocks starting with ~~~ containing --- are not split.
	t.Run("tilde fenced code block does not split", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "slides-tilde-*.md")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		content := `# Slide with tildes
~~~go
func hello() {
    // ---
}
~~~`
		if _, err := tmpFile.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		tmpFile.Close()

		pres, err := ParseMarkdownFile(tmpFile.Name())
		if err != nil {
			t.Fatalf("ParseMarkdownFile failed: %v", err)
		}

		if len(pres.Slides) != 1 {
			t.Fatalf("Expected 1 slide, got %d", len(pres.Slides))
		}

		s0 := pres.Slides[0]
		if !strings.Contains(s0.RawMarkdown, "~~~go") || !strings.Contains(s0.RawMarkdown, "---") || !strings.Contains(s0.RawMarkdown, "~~~") {
			t.Errorf("Expected slide 0 RawMarkdown to contain the tilde fenced code block with '---', got %q", s0.RawMarkdown)
		}
	})

	// 5. A markdown file starting with --- but containing no global frontmatter
	// correctly parses Slide 0 without losing it.
	t.Run("starts with --- but no global frontmatter", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "slides-no-global-*.md")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		content := `---
# Slide 0 Content
This has no global frontmatter, but starts with ---
---
# Slide 1 Content
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

		s0 := pres.Slides[0]
		expectedS0Content := "# Slide 0 Content\nThis has no global frontmatter, but starts with ---"
		if s0.RawMarkdown != expectedS0Content {
			t.Errorf("Expected s0 RawMarkdown %q, got %q", expectedS0Content, s0.RawMarkdown)
		}

		s1 := pres.Slides[1]
		expectedS1Content := "# Slide 1 Content"
		if s1.RawMarkdown != expectedS1Content {
			t.Errorf("Expected s1 RawMarkdown %q, got %q", expectedS1Content, s1.RawMarkdown)
		}

		// Title should be default since there was no global frontmatter
		if pres.Title != "Presentation" {
			t.Errorf("Expected Title 'Presentation', got %q", pres.Title)
		}
	})

	// 6. An HTML comment placed inside a fenced code block is not extracted as speaker notes
	// and remains in the code block.
	t.Run("HTML comment inside fenced code block not extracted", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "slides-comment-code-*.md")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		content := `# Slide with Code Block
` + "```html\n" + `
<!-- comment inside code block -->
` + "```\n"
		if _, err := tmpFile.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		tmpFile.Close()

		pres, err := ParseMarkdownFile(tmpFile.Name())
		if err != nil {
			t.Fatalf("ParseMarkdownFile failed: %v", err)
		}

		if len(pres.Slides) != 1 {
			t.Fatalf("Expected 1 slide, got %d", len(pres.Slides))
		}

		s0 := pres.Slides[0]
		if s0.SpeakerNotes != "" {
			t.Errorf("Expected no speaker notes, got %q", s0.SpeakerNotes)
		}

		if !strings.Contains(s0.RawMarkdown, "<!-- comment inside code block -->") {
			t.Errorf("Expected RawMarkdown to contain '<!-- comment inside code block -->', got %q", s0.RawMarkdown)
		}
	})
}
