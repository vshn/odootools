package templates

import "embed"

//go:embed *.html
var TemplateFS embed.FS

//go:embed favicon.png
//go:embed robots.txt
//go:embed bootstrap.min.css
var PublicFS embed.FS
