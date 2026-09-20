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
		IsAuthenticated: app.isAuthenticated(r),
		Users:           app.users,
	}

	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
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
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

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
	u := User{
		Name: username,
		Role: RoleModerator,
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
	w.WriteHeader(http.StatusOK)
}

func (app *application) usersHandler(w http.ResponseWriter, r *http.Request) {
	data := templateData{
		Users: app.users,
	}

	files := []string{
		"./ui/html/partials/users.tmpl.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "users", data)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
}
