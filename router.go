package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.Handler(http.MethodGet, "/events", app.sseServer)

	// Add write timeouts here
	router.HandlerFunc(http.MethodGet, "/", app.mainHandler)
	router.HandlerFunc(http.MethodGet, "/trigger", app.triggerHandler)
	router.HandlerFunc(http.MethodGet, "/status", app.statusHandler)

	return router
}
