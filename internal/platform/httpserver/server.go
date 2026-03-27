package httpserver

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func New(addr string, readTimeout, writeTimeout time.Duration, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
	}
}

func (s *Server) Start() error {
	go func() {
		_ = s.httpServer.ListenAndServe()
	}()
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
