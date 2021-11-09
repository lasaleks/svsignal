package main

import (
	"context"
	"encoding/json"
	"log"
	"myutils/rabbitmq"
	"regexp"
	"sync"
)

type Hub struct {
	debug_level        int
	CH_MSG_AMPQ        chan rabbitmq.MessageAmpq
	re_rkey            *regexp.Regexp
	CH_SAVE_VALUE      chan ValueSignal
	CH_REQUEST_HTTP    chan RequestHttp
	CH_REQUEST_HTTP_DB chan RequestHttp
}

func newHub() *Hub {
	//^svsignal.(\w+).(\w+)|([^\n]+)$
	re_rkey, _ := regexp.Compile(`svsignal.(\w+).(\w+)`)
	return &Hub{
		CH_MSG_AMPQ:     make(chan rabbitmq.MessageAmpq, 1),
		re_rkey:         re_rkey,
		CH_REQUEST_HTTP: make(chan RequestHttp, 1),
	}
}

type ValueSignal struct {
	system_key string
	signal_key string
	Value      float64 `json:"value"`
	UTime      int64   `json:"utime"`
	Offline    int64   `json:"offline"`
	TypeSave   int     `json:"typesave"`
}

func (h *Hub) run(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("Hub run Done")
			return
		case msg, ok := <-h.CH_MSG_AMPQ:
			if ok {
				//log.Printf("HUB exchange:%s routing_key:%s content_type:%s len:%d", msg.Exchange, msg.Routing_key, msg.Content_type, len(msg.Data))
				ret_cmd := h.re_rkey.FindStringSubmatch(msg.Routing_key)
				if len(ret_cmd) == 3 {
					sys_key := ret_cmd[1]
					sig_key := ret_cmd[2]
					data := ValueSignal{}
					err := json.Unmarshal(msg.Data, &data)
					if err == nil {
						data.system_key = sys_key
						data.signal_key = sig_key
						h.CH_SAVE_VALUE <- data
					}
				}
			} else {
				return
			}
		case msg := <-h.CH_REQUEST_HTTP:
			switch msg.type_cmd {
			case TYPE_CMD_REQUEST_LIST_SIGNAL:
				h.CH_REQUEST_HTTP_DB <- msg
			}
		}
	}
}
