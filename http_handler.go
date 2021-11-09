package main

import (
	"fmt"
	"net/http"
)

const (
	_ = iota
	TYPE_CMD_REQUEST_LIST_SIGNAL
)

type ResponseListSignal struct {
}

type RequestHttp struct {
	type_cmd         int
	CH_RESP_LIST_SIG chan ResponseListSignal
}

type RequestData struct {
}

func (d *RequestData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Привет"))
}

type GetListSignal struct {
	CH_REQUEST_HTTP chan RequestHttp
}

func (g *GetListSignal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	var CH_RESPONSE chan ResponseListSignal = make(chan ResponseListSignal, 1)
	g.CH_REQUEST_HTTP <- RequestHttp{
		type_cmd:         TYPE_CMD_REQUEST_LIST_SIGNAL,
		CH_RESP_LIST_SIG: CH_RESPONSE,
	}
	response := <-CH_RESPONSE
	fmt.Println("RESPONSE", response)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Привет"))
}
