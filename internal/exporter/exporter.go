package exporter

import (
	"html/template"
	"os"

	"github.com/gophernment/gophern/internal/parser"
	"github.com/gophernment/gophern/web"
)

// Export compiles the markdown presentation file into a single self-contained HTML file.
func Export(markdownPath, outputPath string) error {
	pres, err := parser.ParseMarkdownFile(markdownPath)
	if err != nil {
		return err
	}

	cssBytes, err := web.Assets.ReadFile("static/css/styles.css")
	if err != nil {
		return err
	}

	jsBytes, err := web.Assets.ReadFile("static/js/app.js")
	if err != nil {
		return err
	}

	tmpl, err := template.New("export.html").Funcs(template.FuncMap{
		"safe": func(content string) template.HTML { return template.HTML(content) },
	}).ParseFS(web.Assets, "templates/export.html")
	if err != nil {
		return err
	}

	type ExportData struct {
		Title  string
		Slides []parser.Slide
		CSS    template.CSS
		JS     template.JS
	}

	data := ExportData{
		Title:  pres.Title,
		Slides: pres.Slides,
		CSS:    template.CSS(cssBytes),
		JS:     template.JS(jsBytes),
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}
