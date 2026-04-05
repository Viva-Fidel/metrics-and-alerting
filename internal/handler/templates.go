package handler

import (
	"embed"
	"html/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

var metricsPageTmpl = template.Must(template.ParseFS(templatesFS, "templates/metrics.tmpl"))
