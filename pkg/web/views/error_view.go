package views

import "net/http"

type ErrorView struct {
	renderer *Renderer
	template string
}

func NewErrorView(renderer *Renderer) *ErrorView {
	return &ErrorView{
		renderer: renderer,
		template: "error",
	}
}

func (v *ErrorView) ShowError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	v.renderer.Render(w, v.template, Values{
		"Error": err.Error(),
	})
}
