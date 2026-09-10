// Package htmxtemplates owns the dashboard templates and embeds them into the
// application binary so the server does not depend on its working directory.
package htmxtemplates

import (
	"embed"
	"html/template"
)

//go:embed dashboard.html fragments/*.html
var files embed.FS

// Parse returns a new parsed template set. Callers may safely execute the
// returned set concurrently after parsing completes.
func Parse() (*template.Template, error) {
	return template.New("dashboard").ParseFS(files, "dashboard.html", "fragments/*.html")
}
