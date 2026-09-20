package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"

	"github.com/onjen/planning-poker/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.Handler(http.MethodGet, "/static/*filepath", http.FileServerFS(ui.Files))

	router.Handler(http.MethodGet, "/events", app.sseServer)

	//authenticated := alice.New(app.requireAuthentication)

	// Add write timeouts here
	// router.Handler(http.MethodGet, "/", authenticated.ThenFunc(app.mainHandler))
	router.HandlerFunc(http.MethodGet, "/", app.mainHandler)
	router.HandlerFunc(http.MethodGet, "/controls", app.controlsHandler)
	router.HandlerFunc(http.MethodPost, "/vote", app.voteHandler)
	router.HandlerFunc(http.MethodPost, "/join", app.newUserHandler)
	router.HandlerFunc(http.MethodGet, "/users", app.usersHandler)
	router.HandlerFunc(http.MethodGet, "/poll", app.pollHandler)
	router.HandlerFunc(http.MethodPost, "/poll", app.newPollHandler)

	standard := alice.New(app.sessionManager.LoadAndSave, app.logRequest, secureHeaders)

	return standard.Then(router)
}
