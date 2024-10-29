package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/femtech-web/baker/internal/models"
	"github.com/femtech-web/baker/ui"
)

type templateData struct {
	CurrentYear     int
	Flash           string
	CSRFToken       string
	IsAuthenticated bool
	Form            any
	User            *models.User
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2026 at 15:34")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	// declare the cache(template.Template map)
	cache := map[string]*template.Template{}

	// get all pages(fs.Glob)
	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	// loop through all pages and:
	for _, page := range pages {
		// 1 get name of page
		name := filepath.Base(page)

		// 2 get the pattern of each template
		pattern := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		// 3 get the template by chaining the New, Functions and parseFS
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, pattern...)
		if err != nil {
			return nil, err
		}

		// 4 if no err, assign a field of the cache to the current template
		cache[name] = ts
	}

	return cache, nil
}
