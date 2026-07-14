package parser

import (
	"bytes"
	"os"
	"strings"

	"github.com/yuin/goldmark"
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

		var htmlBuf bytes.Buffer
		if err := goldmark.Convert([]byte(slide.RawMarkdown), &htmlBuf); err == nil {
			slide.HTMLContent = htmlBuf.String()
		}

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

// parseYAMLMap checks if a block is a valid non-empty YAML map.
func parseYAMLMap(block string) (map[string]interface{}, bool) {
	block = strings.TrimSpace(block)
	if block == "" {
		return nil, false
	}

	// Reject if the first non-empty line starts with a markdown header, list, blockquote, etc.
	lines := strings.Split(block, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "> ") || strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
			return nil, false
		}
		if strings.HasPrefix(trimmed, "#") {
			i := 0
			for i < len(trimmed) && trimmed[i] == '#' {
				i++
			}
			if i < len(trimmed) && trimmed[i] == ' ' {
				return nil, false
			}
		}
		break // Only check the first non-empty line
	}

	var m map[string]interface{}
	err := yaml.Unmarshal([]byte(block), &m)
	if err != nil {
		return nil, false
	}
	if len(m) == 0 {
		return nil, false
	}

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
				if strings.HasPrefix(k, "-") || strings.HasPrefix(k, "*") || strings.HasPrefix(k, "+") || strings.HasPrefix(k, "#") || strings.HasPrefix(k, ">") {
					return false
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

	notes := markdown[startIdx+4 : endIdx]
	cleanMarkdown := markdown[:startIdx] + markdown[endIdx+3:]
	return strings.TrimSpace(notes), strings.TrimSpace(cleanMarkdown)
}
