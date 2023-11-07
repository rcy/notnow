package events

import (
	_ "embed"
	"html/template"
	"time"
)

var (
	//go:embed partials.gohtml
	Content string
	//partials = template.Must(template.New("events").Parse(Content))

	Funcs = template.FuncMap{
		"weekday": func(date string) string {
			tstr, _ := time.Parse(time.DateOnly, date)
			return tstr.Format("Mon")
		},
		"month": func(date string) string {
			tstr, _ := time.Parse(time.DateOnly, date)
			return tstr.Format("Jan")
		},
		"day": func(date string) string {
			tstr, _ := time.Parse(time.DateOnly, date)
			return tstr.Format("2")
		},
	}
)
