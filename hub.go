package main

import (
	"context"
	"encoding/json"
	"log"
	"regexp"
	"sync"

	gormq "bitbucket.org/lasaleks/go-rmq"
)

type Hub struct {
	debug_level   int
	CH_MSG_AMPQ   chan gormq.MessageAmpq
	re_rkey       *regexp.Regexp
	CH_SAVE_VALUE chan ValueSignal
	CH_SET_SIGNAL chan SetSignal
}

func newHub() *Hub {
	//^svsignal.(\w+).(\w+)|([^\n]+)$
	re_rkey, _ := regexp.Compile(`^svs\.(\w+)\.(\w+)\.(.+)$`)
	return &Hub{
		CH_MSG_AMPQ: make(chan gormq.MessageAmpq, 1),
		re_rkey:     re_rkey,
		//CH_REQUEST_HTTP: make(chan RequestHttp, 1),
	}
}

type SetSignal struct {
	group_key  string
	signal_key string
	TypeSave   int     `json:"typesave"`
	Period     int     `json:"period"`
	Delta      float32 `json:"delta"`
	Name       string  `json:"name"`
	Tags       []struct {
		Tag   string `json:"tag"`
		Value string `json:"value"`
	} `json:"tags"`
}

type ValueSignal struct {
	group_key  string
	signal_key string
	Value      float64 `json:"value"`
	UTime      int64   `json:"utime"`
	Offline    int64   `json:"offline"`
}

func (h *Hub) run(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	// defer fmt.Println("hub.run wg.Done")
	for {
		select {
		case <-ctx.Done():
			// log.Println("Hub run Done")
			return
		case msg, ok := <-h.CH_MSG_AMPQ:
			if ok {
				if DEBUG_LEVEL >= 8 {
					log.Printf("HUB exchange:%s routing_key:%s content_type:%s len:%d", msg.Exchange, msg.Routing_key, msg.Content_type, len(msg.Data))
				}
				ret_cmd := h.re_rkey.FindStringSubmatch(msg.Routing_key)
				if len(ret_cmd) == 4 {
					type_msg := ret_cmd[1]
					sys_key := ret_cmd[2]
					sig_key := ret_cmd[3]
					switch type_msg {
					case "save":
						data := ValueSignal{}
						err := json.Unmarshal(msg.Data, &data)
						if err == nil {
							data.group_key = sys_key
							data.signal_key = sig_key
							h.CH_SAVE_VALUE <- data
						}
					case "set":
						sig := SetSignal{}
						err := json.Unmarshal(msg.Data, &sig)
						if err == nil {
							sig.group_key = sys_key
							sig.signal_key = sig_key
							h.CH_SET_SIGNAL <- sig
						}
					}
				}
			} else {
				return
			}
			/*case msg := <-h.CH_REQUEST_HTTP:
			switch msg.type_cmd {
			case TYPE_CMD_REQUEST_LIST_SIGNAL:
				h.CH_REQUEST_HTTP_DB <- msg
			}*/
		}
	}
}
