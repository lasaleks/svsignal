package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"text/template"
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
	/*Id        int64     `json:"id"`
	SignalKey string    `json:"signalkey"`*/
	Name     string    `json:"name"`
	TypeSave int       `json:"typesave"`
	Period   int       `json:"period"`
	Delta    float32   `json:"delta"`
	Tags     []RLS_Tag `json:"tags"`
}

type RLS_Groups struct {
	Name    string                `json:"name"`
	Signals map[string]RLS_Signal `json:"signals"`
}

type ResponseListSignal struct {
	Groups map[string]*RLS_Groups
}

type ReqListSignal struct {
	CH_RR_LIST_SIGNAL chan ResponseListSignal
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
	jData, err := json.Marshal(response.Groups)
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
	re_key          *regexp.Regexp
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
	begin, err := strconv.ParseInt(get["begin"][0], 10, 32)
	if err != nil {
		fmt.Println(err)
	}
	end, err := strconv.ParseInt(get["end"][0], 10, 32)
	if err != nil {
		fmt.Println(err)
	}

	ret_cmd := g.re_key.FindStringSubmatch(get["signalkey"][0])
	var groupkey string
	var signalkey string
	if len(ret_cmd) == 3 {
		groupkey = ret_cmd[1]
		signalkey = ret_cmd[2]
	} else {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	var CH_RESPONSE chan interface{} = make(chan interface{}, 1)
	g.CH_REQUEST_HTTP <- ReqSignalData{
		begin:       begin,
		end:         end,
		groupkey:    groupkey,
		signalkey:   signalkey,
		CH_RESPONSE: CH_RESPONSE,
	}
	response := <-CH_RESPONSE
	switch response.(type) {
	case error:
		fmt.Println(response)
		http.Error(w, fmt.Sprintf("%v", response), 500)
		return
	case ResponseDataSignalT1:
		break
	case ResponseDataSignalT2:
		break
	}
	jData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

type RequestSaveValue struct {
	CH_SAVE_VALUE chan ValueSignal
	re_key        *regexp.Regexp
}

type ReqJsonSaveValue struct {
	Key     string  `json:"key"`
	Value   float64 `json:"value"`
	UTime   int64   `json:"utime"`
	OffLine int64   `json:"offline"`
}

func (h *RequestSaveValue) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vsig := ValueSignal{}

	if r.Method == http.MethodPost {
		reqBody, _ := ioutil.ReadAll(r.Body)
		var value ReqJsonSaveValue
		err := json.Unmarshal(reqBody, &value)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		ret_cmd := h.re_key.FindStringSubmatch(value.Key)
		if len(ret_cmd) == 3 {
			vsig.group_key = ret_cmd[1]
			vsig.signal_key = ret_cmd[2]
		} else {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		vsig.Value = value.Value
		vsig.UTime = value.UTime
		vsig.Offline = value.OffLine
	} else if r.Method == http.MethodGet {
		get := r.URL.Query()
		var key string
		if key = get.Get("key"); len(key) == 0 {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		ret_cmd := h.re_key.FindStringSubmatch(key)
		if len(ret_cmd) == 3 {
			vsig.group_key = ret_cmd[1]
			vsig.signal_key = ret_cmd[2]
		} else {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		var err error
		if vsig.Value, err = strconv.ParseFloat(get.Get("value"), 64); err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		if vsig.UTime, err = strconv.ParseInt(get.Get("utime"), 10, 64); err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		if vsig.Offline, err = strconv.ParseInt(get.Get("offline"), 10, 64); err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
	} else {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	h.CH_SAVE_VALUE <- vsig

	w.Write([]byte{})
}

type TrendView struct {
	templates []string
}

/*func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}*/

func (h *TrendView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	/*	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}*/
	get := r.URL.Query()
	files := h.templates

	title := ""
	if len(get["title"]) > 0 {
		title = get["title"][0]
	}

	cols := 1
	if len(get["cols"]) > 0 {
		v, err := strconv.ParseInt(get["cols"][0], 10, 32)
		if err == nil {
			cols = int(v)
		}
	}

	height := "80"
	if len(get["height"]) > 0 {
		height = get["height"][0]
	}

	UseGroupChart := 0
	if len(get["GroupChart"]) > 0 {
		UseGroupChart = 1
	}

	signals := get["signalkey"]

	begin, err := strconv.ParseInt(get["begin"][0], 10, 32)
	if err != nil {
		fmt.Println(err)
	}

	end, err := strconv.ParseInt(get["end"][0], 10, 32)
	if err != nil {
		fmt.Println(err)
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := struct {
		Title         string
		Signals       []string
		Begin         int64
		End           int64
		TimeZone      string
		Cols          int
		Height        string
		UseGroupChart int
	}{
		Title:         title,
		Signals:       signals,
		Begin:         begin,
		End:           end,
		TimeZone:      cfg.CONFIG_SERVER.TIME_ZONE,
		Cols:          cols,
		Height:        height,
		UseGroupChart: UseGroupChart,
	}
	err = ts.Execute(w, data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

type GroupSignalView struct {
	templates []string
}

func (h *GroupSignalView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	/*	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}*/
	//get := r.URL.Query()
	files := h.templates

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = ts.Execute(w, nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

type HTTPSetSignal struct {
	CH_SET_SIGNAL chan SetSignal
	re_key        *regexp.Regexp
}

type HTTPRequestSetSignal struct {
	Key      string  `json:"key"`
	TypeSave int     `json:"typesave"`
	Period   int     `json:"period"`
	Delta    float32 `json:"delta"`
	Name     string  `json:"name"`
	Tags     []struct {
		Tag   string `json:"tag"`
		Value string `json:"value"`
	}
}

func (h *HTTPSetSignal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		reqBody, _ := ioutil.ReadAll(r.Body)
		var sig HTTPRequestSetSignal
		err := json.Unmarshal(reqBody, &sig)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		ret_cmd := h.re_key.FindStringSubmatch(sig.Key)
		set_signal := SetSignal{}
		if len(ret_cmd) == 3 {
			set_signal.group_key = ret_cmd[1]
			set_signal.signal_key = ret_cmd[2]
		} else {
			http.Error(w, "Internal Server Error", 500)
			return
		}

		set_signal.TypeSave = sig.TypeSave
		set_signal.Name = sig.Name
		set_signal.Delta = sig.Delta
		set_signal.Period = sig.Period
		set_signal.Tags = sig.Tags
		h.CH_SET_SIGNAL <- set_signal
	} else {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	w.Write([]byte{})
}

type SRV_STATUS struct {
}

func (g *SRV_STATUS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	SrvStatus.HeapInuse = m.HeapInuse
	SrvStatus.StackInuse = m.StackInuse
	SrvStatus.NumGoroutine = runtime.NumGoroutine()
	jData, err := json.Marshal(&SrvStatus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}
