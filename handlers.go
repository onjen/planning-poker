package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/tmaxmax/go-sse"
)

func (app *application) mainHandler(w http.ResponseWriter, r *http.Request) {
	data := templateData{
		IsModerator:     app.isModerator(r),
		IsAuthenticated: app.isAuthenticated(r),
		Users:           app.users,
		Poll:            app.poll,
	}

	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
		"./ui/html/partials/controls.tmpl.html",
		"./ui/html/partials/poll.tmpl.html",
		"./ui/html/pages/home.tmpl.html",
	}

	ts, err := template.ParseFiles(files...)
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

func (app *application) triggerHandler(w http.ResponseWriter, r *http.Request) {
	newvoteE, err := sse.NewType("newvote")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	event := &sse.Message{
		Type: newvoteE,
	}
	err = app.sseServer.Publish(event)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
}

func (app *application) statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, time.Now().Format(time.DateTime))
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
	role := RoleUser
	if len(app.users) == 0 {
		role = RoleModerator
	}
	u := User{
		Name: username,
		Role: role,
		ID:   len(app.users),
	}
	app.users = append(app.users, &u)
	app.sessionManager.Put(r.Context(), "authenticatedUserID", u.ID)

	// notify others to reload user list
	pingU, err := sse.NewType("newuser")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	event := &sse.Message{
		Type: pingU,
	}
	err = app.sseServer.Publish(event)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"joined":{"target":"body"}}`)

	data := templateData{IsAuthenticated: true}
	app.renderPartial(w, "controls", data, "./ui/html/partials/controls.tmpl.html")
}

func (app *application) pollHandler(w http.ResponseWriter, r *http.Request) {
	data := templateData{
		IsModerator: app.isModerator(r),
		Poll:        app.poll,
	}
	app.renderPartial(w, "poll", data, "./ui/html/partials/poll.tmpl.html")
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
	app.poll = poll

	// notify everyone to reload the poll
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
	data := templateData{
		Users: app.users,
	}
	app.renderPartial(w, "users", data, "./ui/html/partials/users.tmpl.html")
}

func (app *application) renderPartial(
	w http.ResponseWriter,
	name string,
	data templateData,
	files ...string,
) {
	ts, err := template.ParseFiles(files...)
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
