package main

import (
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/onjen/planning-poker/ui"
)

type templateData struct {
	IsModerator     bool
	IsAuthenticated bool
	Users           []User
	Poll            string
	PointValues     []int
	Vote            int
	CurrentUserID   int
	Revealed        bool
	CanVote         bool
	VotedCount      int
}

// Partials are cached on their own too, since HTMX swaps them in individually.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		patterns := []string{
			"html/base.tmpl.html",
			"html/partials/*.tmpl.html",
			page,
		}

		ts, err := template.ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[filepath.Base(page)] = ts
	}

	partials, err := fs.Glob(ui.Files, "html/partials/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, partial := range partials {
		ts, err := template.ParseFS(ui.Files, partial)
		if err != nil {
			return nil, err
		}

		cache[filepath.Base(partial)] = ts
	}

	return cache, nil
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
