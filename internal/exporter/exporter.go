package exporter

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/gophernment/gophern/internal/parser"
	"github.com/gophernment/gophern/web"
)

// ChromeAvailableForTest exposes chromeAvailable to external test packages.
func ChromeAvailableForTest() bool { return chromeAvailable() }

// Export compiles the markdown presentation file into a single self-contained
// PDF, with every slide (including any asset/ images it references) captured
// as a full-resolution image via a locally installed headless Chrome.
func Export(markdownPath, outputPath string) error {
	pres, err := parser.ParseMarkdownFile(markdownPath)
	if err != nil {
		return err
	}
	if len(pres.Slides) == 0 {
		return fmt.Errorf("export: %s has no slides", markdownPath)
	}
	if !chromeAvailable() {
		return fmt.Errorf("export: no local Chrome/Chromium install found; PDF export requires one to be installed")
	}

	tmpHTMLPath, err := renderTempHTML(markdownPath, pres)
	if err != nil {
		return err
	}
	defer os.Remove(tmpHTMLPath)

	captureWidth := pres.SlideWidthPx * 2
	captureHeight := pres.SlideHeightPx * 2

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	images, err := captureSlides(ctx, tmpHTMLPath, len(pres.Slides), captureWidth, captureHeight)
	if err != nil {
		return err
	}

	pdfBytes, err := buildPDF(images, captureWidth, captureHeight)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, pdfBytes, 0o644)
}

// renderTempHTML renders the deck via the existing export.html template into
// a temp file placed next to markdownPath, so the file's relative asset/
// references resolve when Chrome loads it over file://.
func renderTempHTML(markdownPath string, pres *parser.Presentation) (string, error) {
	cssBytes, err := web.Assets.ReadFile("static/css/styles.css")
	if err != nil {
		return "", err
	}
	jsBytes, err := web.Assets.ReadFile("static/js/app.js")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("export.html").Funcs(template.FuncMap{
		"safe":           func(content string) template.HTML { return template.HTML(content) },
		"sansFontFamily": func(custom string) template.CSS { return template.CSS(custom + ", " + parser.DefaultSansFallback) },
		"monoFontFamily": func(custom string) template.CSS { return template.CSS(custom + ", " + parser.DefaultMonoFallback) },
		"cssValue":       func(v string) template.CSS { return template.CSS(v) },
	}).ParseFS(web.Assets, "templates/export.html", "templates/_slide.html")
	if err != nil {
		return "", err
	}

	type ExportData struct {
		Title         string
		Fonts         parser.FontsConfig
		Slides        []parser.Slide
		SlideWidthPx  int
		SlideHeightPx int
		CSS           template.CSS
		JS            template.JS
	}

	data := ExportData{
		Title:         pres.Title,
		Fonts:         pres.Fonts,
		Slides:        pres.Slides,
		SlideWidthPx:  pres.SlideWidthPx,
		SlideHeightPx: pres.SlideHeightPx,
		CSS:           template.CSS(cssBytes),
		JS:            template.JS(jsBytes),
	}

	dir := filepath.Dir(markdownPath)
	tmpFile, err := os.CreateTemp(dir, ".gophern-export-*.html")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if err := tmpl.Execute(tmpFile, data); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	absPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	return absPath, nil
}
