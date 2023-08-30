package server

import (
	"context"
	"net/http"

	"github.com/b0shka/backend/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.Config, handler http.Handler) *Server {
	// Create a new instance of http.Server and assign it to httpServer field of Server struct.
	httpServer := &http.Server{
		Addr:           ":" + cfg.HTTP.Port,
		Handler:        handler,
		ReadTimeout:    cfg.HTTP.ReadTimeout,
		WriteTimeout:   cfg.HTTP.WriteTimeout,
		MaxHeaderBytes: cfg.HTTP.MaxHeaderMegabytes << 20,
	}

	// Create a new instance of Server struct and assign the httpServer to its httpServer field.
	server := &Server{
		httpServer: httpServer,
	}

	return server
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
