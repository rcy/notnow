package events

import (
	_ "embed"
	"html/template"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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
		"str": func(uuid pgtype.UUID) string {
			return StringUUID(uuid)
		},
	}
)

func StringUUID(uuid pgtype.UUID) string {
	value, err := uuid.Value()
	if err != nil {
		return ""
	}
	str, _ := value.(string)
	return str
}
