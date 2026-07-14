package web_test

import (
	"html/template"
	"testing"

	"github.com/gophernment/gophern/web"
)

func TestAssetsExist(t *testing.T) {
	files := []string{
		"templates/presentation.html",
		"static/css/styles.css",
		"static/js/app.js",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			_, err := web.Assets.ReadFile(file)
			if err != nil {
				t.Fatalf("failed to read embedded file %s: %v", file, err)
			}
		})
	}
}

func TestTemplateCompiles(t *testing.T) {
	htmlData, err := web.Assets.ReadFile("templates/presentation.html")
	if err != nil {
		t.Fatalf("failed to read presentation template: %v", err)
	}

	tmpl, err := template.New("presentation").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) },
	}).Parse(string(htmlData))
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	if tmpl == nil {
		t.Fatal("expected non-nil template")
	}
}
