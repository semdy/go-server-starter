package template

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"sync"
)

//go:embed templates/*.html
var templateFS embed.FS

var (
	engine *Engine
	once   sync.Once
)

// Engine parses and renders HTML email templates.
type Engine struct {
	templates *template.Template
}

// GetEngine returns the singleton template engine (lazy init).
func GetEngine() *Engine {
	once.Do(func() {
		engine = &Engine{
			templates: template.Must(template.ParseFS(templateFS, "templates/*.html")),
		}
	})
	return engine
}

// Render executes the named template with the given data and returns the HTML string.
func (e *Engine) Render(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// MustRender is like Render but panics on error. Use for templates that must exist.
func (e *Engine) MustRender(name string, data any) string {
	html, err := e.Render(name, data)
	if err != nil {
		panic("template render failed: " + err.Error())
	}
	return html
}

// List returns all available template names.
func (e *Engine) List() []string {
	var names []string
	if e.templates != nil {
		for _, t := range e.templates.Templates() {
			names = append(names, t.Name())
		}
	}
	return names
}

// ---- helpers for use in templates ----

// FuncMap returns helper functions for Go templates.
var FuncMap = template.FuncMap{
	"safeHTML": func(s string) template.HTML { return template.HTML(s) },
}

// ParseAndRender parses a template string and renders it. Useful for database-stored templates.
func ParseAndRender(name, tmpl string, data any) (string, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// WalkTemplates walks all embedded template files.
func WalkTemplates(fn func(name string, content []byte) error) error {
	return fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, err := templateFS.ReadFile(path)
		if err != nil {
			return err
		}
		return fn(path, content)
	})
}
