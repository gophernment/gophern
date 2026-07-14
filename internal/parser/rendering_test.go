package parser

import (
	"strings"
	"testing"
)

func TestMarkdownToHTML(t *testing.T) {
	markdownInput := "# Hello World\n```go\npackage main\n```"
	html, err := RenderMarkdownToHTML(markdownInput)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(html, "<h1>Hello World</h1>") {
		t.Errorf("Expected HTML to contain heading, got %s", html)
	}
	// Ensure Chroma syntax highlighted structure is present
	if !strings.Contains(html, "<pre") || !strings.Contains(html, "package") {
		t.Errorf("Expected HTML to contain highlighted code block, got %s", html)
	}
	// Ensure inline styles are used rather than classes
	if !strings.Contains(html, "style=\"") {
		t.Errorf("Expected HTML to use inline styles, got %s", html)
	}
}

func TestMarkdownToHTML_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		markdownInput string
		wantContains  []string
	}{
		{
			name:          "unknown language uses fallback lexer",
			markdownInput: "```unknownlang\nsome raw code\n```",
			wantContains:  []string{"<pre", "some raw code"},
		},
		{
			name:          "tilde fenced code block",
			markdownInput: "~~~go\npackage main\n~~~",
			wantContains:  []string{"<pre", "package", "style=\""},
		},
		{
			name:          "empty code block",
			markdownInput: "```go\n```",
			wantContains:  []string{"<pre"},
		},
		{
			name:          "markdown table rendering",
			markdownInput: "| Header 1 | Header 2 |\n|---|---|\n| Cell 1 | Cell 2 |",
			wantContains:  []string{"<table>", "<thead>", "<th>Header 1</th>", "<tbody>", "<td>Cell 1</td>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := RenderMarkdownToHTML(tt.markdownInput)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(html, want) {
					t.Errorf("Expected HTML to contain %q, got %s", want, html)
				}
			}
		})
	}
}
