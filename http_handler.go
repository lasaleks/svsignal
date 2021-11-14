package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const (
	_ = iota
	TYPE_CMD_REQUEST_LIST_SIGNAL
)

type RLS_Tag struct {
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

type RLS_Signal struct {
	Id        int64     `json:"id"`
	SignalKey string    `json:"signalkey"`
	Name      string    `json:"Name"`
	TypeSave  int       `json:"typesave"`
	Period    int       `json:"period"`
	Delta     float32   `json:"delta"`
	Tags      []RLS_Tag `json:"tags"`
}

type RLS_Groups struct {
	Name    string       `json:"name"`
	Signals []RLS_Signal `json:"signals"`
}

type ResponseListSignal struct {
	Groups map[string]*RLS_Groups `json:"groups"`
}

type ReqListSignal struct {
	CH_RR_LIST_SIGNAL chan ResponseListSignal
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
	CH_REQUEST_HTTP chan interface{}
}

func (g *GetListSignal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	var CH_RESPONSE chan ResponseListSignal = make(chan ResponseListSignal, 1)
	g.CH_REQUEST_HTTP <- ReqListSignal{
		CH_RR_LIST_SIGNAL: CH_RESPONSE,
	}
	response := <-CH_RESPONSE
	jData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

/*
type ValueInt struct {
	Id      int64
	Value   int64
	UTime   int64
	OffLine byte
}

type ValueAvg struct {
	Id      int64
	Value   float64
	UTime   int64
	OffLine byte
}
*/

type ValueM struct {
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
	Tags       []RLS_Tag  `json:"tags"`
	TypeSave   int        `json:"typesave"`
	Values     [][4]int64 `json:"values"`
}

type ResponseDataSignalT2 struct {
	GroupKey   string           `json:"groupkey"`
	GroupName  string           `json:"groupname"`
	SignalKey  string           `json:"signalkey"`
	SignalName string           `json:"signalname"`
	Tags       []RLS_Tag        `json:"tags"`
	TypeSave   int              `json:"typesave"`
	Values     [][4]interface{} `json:"values"`
}

type ResponseDataSignalT3 struct {
	GroupKey   string    `json:"groupkey"`
	GroupName  string    `json:"groupname"`
	SignalKey  string    `json:"signalkey"`
	SignalName string    `json:"signalname"`
	Tags       []RLS_Tag `json:"tags"`
	TypeSave   int       `json:"typesave"`
	Values     []ValueM  `json:"values"`
}

type RequestSignalData struct {
	CH_REQUEST_HTTP chan interface{}
}

type ReqSignalData struct {
	begin       int64
	end         int64
	groupkey    string
	signalkey   string
	CH_RESPONSE chan interface{}
}

func (g *RequestSignalData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	get := r.URL.Query()
	fmt.Println(get)
	begin, err := strconv.ParseInt(get["begin"][0], 10, 32)
	if err != nil {
		fmt.Println(err)
	}
	end, err := strconv.ParseInt(get["end"][0], 10, 32)
	if err != nil {
		fmt.Println(err)
	}
	groupkey := get["groupkey"][0]
	signalkey := get["signalkey"][0]
	fmt.Println(begin, end, groupkey, signalkey)

	var CH_RESPONSE chan interface{} = make(chan interface{}, 1)
	g.CH_REQUEST_HTTP <- ReqSignalData{
		begin:       begin,
		end:         end,
		groupkey:    groupkey,
		signalkey:   signalkey,
		CH_RESPONSE: CH_RESPONSE,
	}
	response := <-CH_RESPONSE
	jData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}
