package http

import (
	"fmt"
	"net/http"
	"time"
)

type httpServer struct {
	router *Router
	port   int
}

func (s *httpServer) Run() {
	var srv *http.Server

	srv = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           s.router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	srv.ListenAndServe()
}
