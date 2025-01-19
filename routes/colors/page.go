package colors

import (
	_ "embed"
	"net/http"
	"sort"
	"strconv"
	"yikes/internal/html"
	mw "yikes/middleware"
	"yikes/services/google"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Page(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := mw.UserFromContext(ctx)

	colors, err := google.Colors(ctx, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	keys := html.MapKeys(colors.Event)
	sort.Slice(keys, func(i int, j int) bool {
		a, _ := strconv.Atoi(keys[i])
		b, _ := strconv.Atoi(keys[j])
		return a < b
	})

	err = Div(
		Map(keys, func(k string) Node {
			c := colors.Event[k]
			return Div(Textf("%s:%s", k, c.Background), Style("background:"+c.Background))
		}),
	).Render(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
