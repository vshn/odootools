package controller

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/vshn/odootools/templates"
)

// Renderer is able to render HTML templates. Templates will be compiled the first
// time they are requested, and cached thereafter.
type Renderer struct {
	cache map[string]*template.Template
}

// NewRenderer returns a new "Renderer" struct
func NewRenderer() *Renderer {
	return &Renderer{
		cache: map[string]*template.Template{},
	}
}

// Render renders the requested template with the given data into w.
// "template" is suffixed with ".html", and then rendered together with "layout.html".
func (v *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tpl, err := v.getTemplate(name)
	if err != nil {
		return err
	}
	return tpl.Execute(w, data)
}

func (v *Renderer) getTemplate(name string) (*template.Template, error) {
	if v.cache[name] == nil {
		t, err := template.ParseFS(templates.TemplateFS, "layout.html", name+".html", "nav.html")
		if err != nil {
			return nil, err
		}
		v.cache[name] = t
	}

	return v.cache[name], nil
}
