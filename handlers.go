package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/tmaxmax/go-sse"

	"github.com/onjen/planning-poker/ui"
)

func (app *application) mainHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	files := []string{
		"html/base.tmpl.html",
		"html/partials/controls.tmpl.html",
		"html/partials/poll.tmpl.html",
		"html/pages/home.tmpl.html",
	}

	ts, err := template.ParseFS(ui.Files, files...)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
}

func (app *application) voteHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := app.userID(r)
	if !ok {
		http.Error(w, "Join before voting", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	points, err := strconv.Atoi(r.FormValue("points"))
	if err != nil || !validPointValue(points) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, "Not a valid story point value")
		return
	}

	app.mu.Lock()
	u := app.findUser(id)
	if u == nil {
		app.mu.Unlock()
		http.Error(w, "Join before voting", http.StatusForbidden)
		return
	}
	if app.poll == "" {
		app.mu.Unlock()
		http.Error(w, "No poll to vote on", http.StatusConflict)
		return
	}
	if app.revealed {
		app.mu.Unlock()
		http.Error(w, "Voting is closed, wait for the next poll", http.StatusConflict)
		return
	}
	u.Vote = points
	if app.allVotedLocked() {
		app.revealed = true
	}
	app.mu.Unlock()

	newvoteE, err := sse.NewType("newvote")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	err = app.sseServer.Publish(&sse.Message{Type: newvoteE})
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	app.controlsHandler(w, r)
}

func (app *application) allVotedLocked() bool {
	if len(app.users) == 0 {
		return false
	}
	for _, u := range app.users {
		if !u.HasVoted() {
			return false
		}
	}
	return true
}

func (app *application) newUserHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, "Username cannot be empty")
		return
	}

	app.mu.Lock()
	role := RoleUser
	if len(app.users) == 0 {
		role = RoleModerator
	}
	u := User{
		Name: username,
		Role: role,
		ID:   len(app.users),
		Vote: noVote,
	}
	app.users = append(app.users, &u)
	app.mu.Unlock()

	app.sessionManager.Put(r.Context(), "authenticatedUserID", u.ID)

	pingU, err := sse.NewType("newuser")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	err = app.sseServer.Publish(&sse.Message{Type: pingU})
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"joined":{"target":"body"}}`)

	app.controlsHandler(w, r)
}

func (app *application) controlsHandler(w http.ResponseWriter, r *http.Request) {
	app.renderPartial(w, "controls", app.newTemplateData(r),
		"html/partials/controls.tmpl.html")
}

func (app *application) pollHandler(w http.ResponseWriter, r *http.Request) {
	app.renderPartial(w, "poll", app.newTemplateData(r),
		"html/partials/poll.tmpl.html")
}

func (app *application) newPollHandler(w http.ResponseWriter, r *http.Request) {
	if !app.isModerator(r) {
		http.Error(w, "Only the moderator can set the poll", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	poll := r.FormValue("poll")
	if poll == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, "Poll cannot be empty")
		return
	}

	app.mu.Lock()
	app.poll = poll
	app.revealed = false
	for _, u := range app.users {
		u.Vote = noVote
	}
	app.mu.Unlock()

	newpollE, err := sse.NewType("newpoll")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	err = app.sseServer.Publish(&sse.Message{Type: newpollE})
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	app.pollHandler(w, r)
}

func (app *application) usersHandler(w http.ResponseWriter, r *http.Request) {
	app.renderPartial(w, "users", app.newTemplateData(r),
		"html/partials/users.tmpl.html")
}

func (app *application) renderPartial(
	w http.ResponseWriter,
	name string,
	data templateData,
	files ...string,
) {
	ts, err := template.ParseFS(ui.Files, files...)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, name, data)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
}
