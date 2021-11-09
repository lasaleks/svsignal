package main

import (
	"context"
	"log"
	"net/http"
	"sync"
)

type HttpSrv struct {
	Addr   string
	server *http.Server
	hub    *Hub
}

func (h *HttpSrv) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	h.server = &http.Server{Addr: h.Addr, Handler: nil}

	http.Handle("/api/requestdata/", &RequestData{})
	http.Handle("/api/listsignal/", &GetListSignal{h.hub.CH_REQUEST_HTTP})

	log.Printf("Starting Http Server at %s\n", h.Addr)
	err := h.server.ListenAndServe()
	if err == http.ErrServerClosed { // graceful shutdown
		log.Println("commencing Http server shutdown...")
		log.Println("Http server was shut down.")
	} else if err != nil {
		log.Printf("Http server error: %v\n", err)
	}
}

func (h *HttpSrv) Close() {
	err := h.server.Shutdown(context.Background())
	// can't do much here except for logging any errors
	if err != nil {
		log.Printf("error Http during shutdown: %v\n", err)
	}
}
