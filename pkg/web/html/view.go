package html

import (
	"html/template"
	"log"
	"net/http"
	"path"
)

// View is able to render HTML templates. Templates will be comiled the first
// time they are requested, and cached thereafter.
type View struct {
	root string

	cache map[string]*template.Template
}

// NewView returns a new "View" struct
func NewView(root string) *View {
	return &View{
		root:  root,
		cache: make(map[string]*template.Template),
	}
}

// Render renders the requested template with the given data into w.
// "template" is suffixed with ".html", and then rendered together with "layout.html".
func (v *View) Render(w http.ResponseWriter, template string, data interface{}) {
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

func (v *View) getTemplate(name string) (*template.Template, error) {
	if v.cache[name] == nil {
		t, err := template.ParseFiles(
			path.Join(v.root, "layout.html"),
			path.Join(v.root, name+".html"),
		)
		if err != nil {
			return nil, err
		}
		v.cache[name] = t
	}

	return v.cache[name], nil
}
