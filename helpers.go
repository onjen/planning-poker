package main

import "net/http"

func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessionManager.Exists(r.Context(), "authenticatedUserID")
}

func (app *application) isModerator(r *http.Request) bool {
	if app.sessionManager.Exists(r.Context(), "authenticatedUserID") {
		uid := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
		return app.users[uid].Role == RoleModerator
	}
	return false
}
