package handler

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templatesFS embed.FS

var metricsPageTmpl = template.Must(template.ParseFS(templatesFS, "templates/metrics.html"))
