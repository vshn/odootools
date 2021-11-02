package templates

import "embed"

//go:embed *.html
var TemplateFS embed.FS

//go:embed favicon.png
//go:embed robots.txt
var PublicFS embed.FS
