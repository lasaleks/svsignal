package main

import (
	"fmt"
	"net/http"
)

const (
	_ = iota
	TYPE_CMD_REQUEST_LIST_SIGNAL
)

type RLS_Signal struct {
	Id        int     `json:"id"`
	SignalKey string  `json:"signalkey"`
	Name      string  `json:"Name"`
	TypeSave  int     `json:"typesave"`
	Period    int     `json:"period"`
	Delta     float32 `json:"delta"`
}

type RLS_Groups struct {
	Name     string `json:"name"`
	GroupKey string `json:"groupkey"`
	Signals  []RLS_Signal
}

type ResponseListSignal struct {
	Groups []RLS_Groups
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

	/*jData, err := json.Marshal(result)
	if err != nil {
		// handle error
	}*/
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Привет"))
}

type ValuesT1 struct {
	Id      int64
	Value   int64
	UTime   int64
	OffLine byte
}

type ValuesT2 struct {
	Id      int64
	Value   float64
	UTime   int64
	OffLine byte
}

type ValuesT3 struct {
	Id      int64
	Max     float64
	Min     float64
	Mean    float64
	Median  float64
	UTime   float64
	OffLine byte
}

type ResponseDataSignalT1 struct {
	GroupKey   string     `json:"groupkey"`
	GroupName  string     `json:"groupname"`
	SignalKey  string     `json:"signalkey"`
	SignalName string     `json:"signalname"`
	TypeSave   int        `json:"typesave"`
	Values     [][4]int64 `json:"values"`
}

type ResponseDataSignalT2 struct {
	GroupKey   string           `json:"groupkey"`
	GroupName  string           `json:"groupname"`
	SignalKey  string           `json:"signalkey"`
	SignalName string           `json:"signalname"`
	TypeSave   int              `json:"typesave"`
	Values     [][4]interface{} `json:"values"`
}

type ResponseDataSignalT3 struct {
	GroupKey   string     `json:"groupkey"`
	GroupName  string     `json:"groupname"`
	SignalKey  string     `json:"signalkey"`
	SignalName string     `json:"signalname"`
	TypeSave   int        `json:"typesave"`
	Values     []ValuesT3 `json:"values"`
}

type RequestSignalData struct {
	CH_REQUEST_HTTP chan RequestHttp
}

func (g *RequestSignalData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	/*jData, err := json.Marshal(result)
	if err != nil {
		// handle error
	}*/
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Привет"))
}
