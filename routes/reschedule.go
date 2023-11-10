package routes

import (
	_ "embed"
	"net/http"
	mw "yikes/middleware"
	"yikes/services/google"
)

func postReschedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := mw.UserFromContext(ctx)

	err := google.ReschedulePastTasks(ctx, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", "calendarUpdated")

	w.Write([]byte("ok"))
}
