package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/tmaxmax/go-sse"
)

type config struct {
	port int
	env  string
}

type application struct {
	config    config
	logger    *slog.Logger
	sseServer *sse.Server
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "Server port")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	s := &sse.Server{}
	defer s.Shutdown(nil)

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
			_ = s.Publish(ping)
		}
	}()

	app := &application{
		config:    cfg,
		logger:    logger,
		sseServer: s,
	}

	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", cfg.port),
		Handler:     app.routes(),
		IdleTimeout: time.Minute,
		ReadTimeout: 5 * time.Second,
		ErrorLog:    slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)
	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
