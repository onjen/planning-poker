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

func (app *application) newTemplateData(r *http.Request) templateData {
	data := templateData{
		IsAuthenticated: app.isAuthenticated(r),
		PointValues:     pointValues,
		Vote:            noVote,
		CurrentUserID:   -1,
	}

	app.mu.RLock()
	defer app.mu.RUnlock()

	data.Users = make([]User, len(app.users))
	for i, u := range app.users {
		data.Users[i] = *u
		if u.HasVoted() {
			data.VotedCount++
		}
	}
	data.Poll = app.poll
	data.Revealed = app.revealed

	if id, ok := app.userID(r); ok {
		if u := app.findUser(id); u != nil {
			data.IsModerator = u.Role == RoleModerator
			data.CurrentUserID = u.ID
			data.Vote = u.Vote
		}
	}

	data.CanVote = data.IsAuthenticated && data.Poll != "" && !data.Revealed

	return data
}
