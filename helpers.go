package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	app.logger.Error(err.Error())
	http.Error(w, "InternalServerError", http.StatusInternalServerError)
}

func (app *application) render(
	w http.ResponseWriter,
	status int,
	file string,
	name string,
	data templateData,
) {
	ts, ok := app.templateCache[file]
	if !ok {
		app.serverError(w, fmt.Errorf("the template %s does not exist", file))
		return
	}

	buf := new(bytes.Buffer)
	if err := ts.ExecuteTemplate(buf, name, data); err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessionManager.Exists(r.Context(), "authenticatedUserID")
}

func (app *application) userID(r *http.Request) (int, bool) {
	if !app.isAuthenticated(r) {
		return 0, false
	}
	return app.sessionManager.GetInt(r.Context(), "authenticatedUserID"), true
}

func (app *application) findUser(id int) *User {
	for _, u := range app.users {
		if u.ID == id {
			return u
		}
	}
	return nil
}

func (app *application) isModerator(r *http.Request) bool {
	id, ok := app.userID(r)
	if !ok {
		return false
	}

	app.mu.RLock()
	defer app.mu.RUnlock()

	u := app.findUser(id)
	return u != nil && u.Role == RoleModerator
}

func (u User) HasVoted() bool { return u.Vote != noVote }

func validPointValue(v int) bool {
	for _, p := range pointValues {
		if p == v {
			return true
		}
	}
	return false
}
