package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"syscall"
)

type HttpSrv struct {
	Addr       string
	UnixSocket string
	server     *http.Server
	hub        *Hub
	svsignal   *SVSignalDB
	listen     net.Listener
	cfg        *Config
}

func newUnixSocket(path string) (net.Listener, error) {
	if err := syscall.Unlink(path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	/*mask := syscall.Umask(0777)
	defer syscall.Umask(0777)*/
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0777); err != nil {
		log.Printf("Error chmod unixsocket %s;%v", path, err)
	}
	return l, nil
}

func (h *HttpSrv) initUnixSocketServer() error {
	var err error
	if h.Addr != "" {
		log.Printf("listen on %s", h.Addr)
		h.listen, err = net.Listen("tcp", h.Addr)
	} else {
		log.Printf("listen on %s", h.UnixSocket)
		h.listen, err = newUnixSocket(h.UnixSocket)
	}
	if err != nil {
		return err
	}
	return nil
}

func (h *HttpSrv) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	h.server = &http.Server{Addr: h.Addr, Handler: nil}

	http.Handle("/svs/trend/apx/view", &TrendView{
		templates: []string{
			"./templates/trend_apx.page.html",
			"./templates/base.layout.html",
		},
		cfg: h.cfg,
	})

	http.Handle("/svs/trend/apx/view2", &TrendView{templates: []string{
		"./templates/trend_apx2.page.html",
		"./templates/base.layout.html",
	},
		cfg: h.cfg,
	})

	http.Handle("/svs/trend/highchart/view", &TrendView{templates: []string{
		"./templates/trend_highchart.page.html",
		"./templates/base.layout.html",
	},
		cfg: h.cfg,
	})

	http.Handle("/svs/signal/", &TrendView{templates: []string{
		"./templates/datasignal.page.html",
		"./templates/base.layout.html",
	},
		cfg: h.cfg,
	})

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
	http.Handle("/svs/api/setsignal", &HTTPSetSignal{CH_SET_SIGNAL: h.hub.CH_SET_SIGNAL, re_key: re_key})

	log.Printf("Starting Http Server at %s\n", h.Addr)
	err := h.server.Serve(h.listen)
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
