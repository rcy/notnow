package events

import (
	_ "embed"
	"html/template"
)

var (
	//go:embed partials.gohtml
	Content  string
	partials = template.Must(template.New("events").Parse(Content))
)
