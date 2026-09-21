package main

import (
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/tmaxmax/go-sse"
)

type config struct {
	port int
	env  string
}

type Role int

const (
	RoleModerator Role = iota
	RoleUser
)

func (r Role) String() string {
	switch r {
	case RoleModerator:
		return "Moderator"
	case RoleUser:
		return "User"
	default:
		return fmt.Sprintf("Role(%d)", int(r))
	}
}

var pointValues = []int{0, 1, 2, 3, 5, 8}

type User struct {
	Name string
	Role Role
	ID   int
	Vote int
}

const noVote = -1

type application struct {
	config         config
	logger         *slog.Logger
	sseServer      *sse.Server
	sessionManager *scs.SessionManager
	templateCache  map[string]*template.Template

	mu       sync.RWMutex
	users    []*User
	poll     string
	revealed bool
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "Server port")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	sseServer := &sse.Server{}
	defer sseServer.Shutdown(nil)

	// Background loop to keep TCP streams alive
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		pingE, err := sse.NewType("ping")
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}

		for range ticker.C {
			ping := &sse.Message{
				Type: pingE,
			}
			_ = sseServer.Publish(ping)
		}
	}()

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		config:         cfg,
		logger:         logger,
		sseServer:      sseServer,
		sessionManager: sessionManager,
		templateCache:  templateCache,
	}

	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", cfg.port),
		Handler:     app.routes(),
		IdleTimeout: time.Minute,
		ReadTimeout: 5 * time.Second,
		ErrorLog:    slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)
	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
