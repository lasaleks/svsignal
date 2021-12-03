package main

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"sync"
)

type HttpSrv struct {
	Addr     string
	server   *http.Server
	hub      *Hub
	svsignal *SVSignalDB
}

func (h *HttpSrv) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	h.server = &http.Server{Addr: h.Addr, Handler: nil}

	http.Handle("/svs/signal", &TrendView{templates: []string{
		"./templates/datasignal.page.html",
		"./templates/base.layout.html",
	}})

	http.Handle("/svs/signals", &GroupSignalView{templates: []string{
		"./templates/signals.page.html",
		"./templates/base.layout.html",
	}})

	staticHandler := http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static")),
	)
	http.Handle("/static/", staticHandler)
	re_key, _ := regexp.Compile(`^(\w+)\.(.+)$`)
	http.Handle("/svs/api/signal/getdata", &RequestSignalData{CH_REQUEST_HTTP: h.svsignal.CH_REQUEST_HTTP, re_key: re_key})
	http.Handle("/svs/api/getlistsignal", &GetListSignal{h.svsignal.CH_REQUEST_HTTP})
	http.Handle("/svs/api/savevalue", &RequestSaveValue{CH_SAVE_VALUE: h.hub.CH_SAVE_VALUE, re_key: re_key})

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
