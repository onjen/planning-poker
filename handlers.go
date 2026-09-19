package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/tmaxmax/go-sse"
)

func (app *application) mainHandler(w http.ResponseWriter, r *http.Request) {
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

	err = ts.ExecuteTemplate(w, "base", nil)
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
