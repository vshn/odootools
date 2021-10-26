package html

import "net/http"

type RequestReportView struct {
	renderer *Renderer
	template string
}

func NewRequestReportView(renderer *Renderer) *RequestReportView {
	return &RequestReportView{
		renderer: renderer,
		template: "report",
	}
}

func (v *RequestReportView) ShowConfigurationForm(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	v.renderer.Render(w, v.template, Values{})
}
