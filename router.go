package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/", app.mainHandler)
	router.HandlerFunc(http.MethodGet, "/events", app.eventsHandler)

	return router
}
