package colors

import (
	_ "embed"
	"fmt"
	"net/http"
	"text/template"
	mw "yikes/middleware"
	"yikes/services/google"

	"google.golang.org/api/calendar/v3"
)

var (
	//go:embed page.gohtml
	pageContent  string
	pageTemplate = template.Must(template.New("").Parse(pageContent))
)

func Page(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := mw.UserFromContext(ctx)

	colors, err := google.Colors(ctx, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("%+v\n", colors.Calendar)

	err = pageTemplate.Execute(w, struct {
		Colors map[string]calendar.ColorDefinition
	}{
		Colors: colors.Event,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
