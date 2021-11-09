package html

import (
	"html/template"
	"log"
	"net/http"

	"github.com/vshn/odootools/templates"
)

// Renderer is able to render HTML templates. Templates will be compiled the first
// time they are requested, and cached thereafter.
type Renderer struct {
	cache map[string]*template.Template
}

// Values is an arbitrary tree of data to be passed to template rendering.
type Values map[string]interface{}

// NewRenderer returns a new "Renderer" struct
func NewRenderer() *Renderer {
	return &Renderer{
		cache: make(map[string]*template.Template),
	}
}

// Render renders the requested template with the given data into w.
// "template" is suffixed with ".html", and then rendered together with "layout.html".
func (v *Renderer) Render(w http.ResponseWriter, template string, data Values) {
	tpl, err := v.getTemplate(template)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/html")
	if err := tpl.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
