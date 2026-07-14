package parser

import (
	"bytes"
	"os"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v3"
)

// Presentation represents the parsed presentation document.
type Presentation struct {
	Title       string `yaml:"title"`
	Author      string `yaml:"author"`
	Theme       string `yaml:"theme"`
	AspectRatio string `yaml:"aspectRatio"`
	Slides      []Slide
}

// Slide represents a single slide in the presentation.
type Slide struct {
	Index        int
	RawMarkdown  string
	HTMLContent  string
	Layout       string `yaml:"layout"`
	Background   string `yaml:"background"`
	Color        string `yaml:"color"`
	SpeakerNotes string
}

// ParseMarkdownFile parses a Markdown file into a Presentation struct.
func ParseMarkdownFile(path string) (*Presentation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	blocks := splitBySeparator(content)

	pres := &Presentation{
		Title:       "Presentation",
		AspectRatio: "16:9",
		Theme:       "default",
	}

	var slide0Layout, slide0Background, slide0Color string
	var remainingBlocks []string
	if len(blocks) > 0 {
		// Check if first block is empty (meaning the file started with ---)
		if strings.TrimSpace(blocks[0]) == "" && len(blocks) > 1 {
			if _, ok := parseYAMLMap(blocks[1]); ok {
				// Parse blocks[1] as global frontmatter
				var globalConfig Presentation
				if err := yaml.Unmarshal([]byte(blocks[1]), &globalConfig); err == nil {
					if globalConfig.Title != "" {
						pres.Title = globalConfig.Title
					}
					if globalConfig.Author != "" {
						pres.Author = globalConfig.Author
					}
					if globalConfig.Theme != "" {
						pres.Theme = globalConfig.Theme
					}
					if globalConfig.AspectRatio != "" {
						pres.AspectRatio = globalConfig.AspectRatio
					}
				}

				var slide0Config Slide
				if err := yaml.Unmarshal([]byte(blocks[1]), &slide0Config); err == nil {
					slide0Layout = slide0Config.Layout
					slide0Background = slide0Config.Background
					slide0Color = slide0Config.Color
				}

				remainingBlocks = blocks[2:]
			} else {
				remainingBlocks = blocks[1:]
			}
		} else {
			remainingBlocks = blocks
		}
	}


	slideIdx := 0
	for i := 0; i < len(remainingBlocks); {
		block := remainingBlocks[i]

		var slide Slide
		slide.Index = slideIdx
		if slideIdx == 0 {
			slide.Layout = slide0Layout
			slide.Background = slide0Background
			slide.Color = slide0Color
		}

		// Check if the current block is a YAML frontmatter block
		if _, ok := parseYAMLMap(block); ok {
			// Parse block as frontmatter
			_ = yaml.Unmarshal([]byte(block), &slide)

			if i < len(remainingBlocks)-1 {
				contentBlock := remainingBlocks[i+1]
				notes, cleanContent := extractSpeakerNotes(contentBlock)
				slide.RawMarkdown = cleanContent
				slide.SpeakerNotes = notes
				i += 2
			} else {
				slide.RawMarkdown = ""
				slide.SpeakerNotes = ""
				i++
			}
		} else {
			notes, cleanContent := extractSpeakerNotes(block)
			slide.RawMarkdown = cleanContent
			slide.SpeakerNotes = notes
			i++
		}

		htmlContent, err := RenderMarkdownToHTML(slide.RawMarkdown)
		if err != nil {
			return nil, err
		}
		slide.HTMLContent = htmlContent

		pres.Slides = append(pres.Slides, slide)
		slideIdx++
	}

	return pres, nil
}

// splitBySeparator splits the markdown content by lines that are exactly "---".
func splitBySeparator(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")

	var blocks []string
	var currentBlock []string
	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inCodeBlock = !inCodeBlock
		}

		if trimmed == "---" && !inCodeBlock {
			blocks = append(blocks, strings.Join(currentBlock, "\n"))
			currentBlock = []string{}
		} else {
			currentBlock = append(currentBlock, line)
		}
	}
	blocks = append(blocks, strings.Join(currentBlock, "\n"))
	return blocks
}

var coreFrontmatterKeys = map[string]bool{
	"title":       true,
	"author":      true,
	"theme":       true,
	"aspectRatio": true,
	"layout":      true,
	"background":  true,
	"color":       true,
	"class":       true,
	"transition":  true,
	"disabled":    true,
	"clicks":      true,
	"preload":     true,
	"src":         true,
	"name":        true,
	"route":       true,
	"drawings":    true,
}

// parseYAMLMap checks if a block is a valid non-empty YAML map.
func parseYAMLMap(block string) (map[string]interface{}, bool) {
	block = strings.TrimSpace(block)
	if block == "" {
		return nil, false
	}

	var m map[string]interface{}
	err := yaml.Unmarshal([]byte(block), &m)
	if err != nil {
		return nil, false
	}
	if len(m) == 0 {
		return nil, false
	}

	// Validate map keys consist of standard YAML identifiers
	var validate func(interface{}) bool
	validate = func(val interface{}) bool {
		switch v := val.(type) {
		case map[string]interface{}:
			for k, child := range v {
				if k == "" {
					return false
				}
				for _, r := range k {
					if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
						return false
					}
				}
				if !validate(child) {
					return false
				}
			}
		case []interface{}:
			for _, item := range v {
				if !validate(item) {
					return false
				}
			}
		}
		return true
	}

	if !validate(m) {
		return nil, false
	}

	// Ensure the YAML map contains at least one core slide or presentation configuration key.
	// This prevents markdown paragraphs/lists containing colons from being misclassified as frontmatter
	// while allowing frontmatter blocks to contain YAML comments on any line (including the first).
	hasCoreKey := false
	for k := range m {
		if coreFrontmatterKeys[k] {
			hasCoreKey = true
			break
		}
	}
	if !hasCoreKey {
		return nil, false
	}

	// Verify every line of the block is YAML structure to avoid false positives with markdown
	lines := strings.Split(block, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Must be a comment, or start with spaces/tabs (indented YAML), or be a key-value line
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		// Otherwise, it must be a top-level key-value line
		colonIdx := strings.Index(trimmed, ":")
		if colonIdx == -1 || colonIdx == 0 {
			return nil, false
		}
		key := strings.TrimSpace(trimmed[:colonIdx])
		// Key must be valid YAML key (alphanumeric/hyphen/underscore)
		for _, r := range key {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
				return nil, false
			}
		}
	}

	return m, true
}

// extractSpeakerNotes searches for HTML comment blocks (<!-- ... -->) at the bottom of the slide markdown
// and returns the extracted speaker notes and the clean markdown.
func extractSpeakerNotes(markdown string) (string, string) {
	startIdx := strings.LastIndex(markdown, "<!--")
	if startIdx == -1 {
		return "", strings.TrimSpace(markdown)
	}
	endIdx := strings.Index(markdown[startIdx:], "-->")
	if endIdx == -1 {
		return "", strings.TrimSpace(markdown)
	}
	endIdx += startIdx

	// Ensure that only whitespace or newlines are present after the closing --> tag
	afterComment := markdown[endIdx+3:]
	if strings.TrimSpace(afterComment) != "" {
		return "", strings.TrimSpace(markdown)
	}

	notes := markdown[startIdx+4 : endIdx]
	cleanMarkdown := markdown[:startIdx] + markdown[endIdx+3:]
	return strings.TrimSpace(notes), strings.TrimSpace(cleanMarkdown)
}

var markdownRenderer = goldmark.New(
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
		renderer.WithNodeRenderers(
			util.Prioritized(&ChromaRenderer{}, 100),
		),
	),
)

// RenderMarkdownToHTML renders markdown content into HTML with code syntax highlighting.
func RenderMarkdownToHTML(markdownInput string) (string, error) {
	var buf bytes.Buffer
	if err := markdownRenderer.Convert([]byte(markdownInput), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

var chromaFormatter = chromahtml.New(
	chromahtml.WithClasses(false),
)

type ChromaRenderer struct{}

func (r *ChromaRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
}

func (r *ChromaRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.FencedCodeBlock)
	lang := string(n.Language(source))

	var codeBuf bytes.Buffer
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		codeBuf.Write(line.Value(source))
	}

	return r.highlight(w, codeBuf.String(), lang)
}

func (r *ChromaRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.CodeBlock)

	var codeBuf bytes.Buffer
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		codeBuf.Write(line.Value(source))
	}

	return r.highlight(w, codeBuf.String(), "")
}

func (r *ChromaRenderer) highlight(w util.BufWriter, code string, lang string) (ast.WalkStatus, error) {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get("github")
	if style == nil {
		style = styles.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return ast.WalkStop, err
	}

	err = chromaFormatter.Format(w, style, iterator)
	if err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

