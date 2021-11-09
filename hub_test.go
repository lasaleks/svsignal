package main

import (
	"context"
	"encoding/json"
	"myutils/rabbitmq"
	"reflect"
	"sync"
	"testing"
)

func TestHubRun(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

	var cfg Config

	hub := newHub()
	hub.CH_SAVE_VALUE = make(chan ValueSignal, 1)
	hub.debug_level = cfg.CONFIG.DEBUG_LEVEL
	ctx_hub, cancel_hub := context.WithCancel(ctx)
	wg.Add(1)
	go hub.run(&wg, ctx_hub)

	val1 := ValueSignal{
		system_key: "IE",
		signal_key: "beacon_1234_rx",
		Value:      100,
		UTime:      1636388515,
		Offline:    0,
		TypeSave:   2,
	}
	jData, _ := json.Marshal(val1)

	hub.CH_MSG_AMPQ <- rabbitmq.MessageAmpq{
		Exchange:     "svsignal",
		Routing_key:  "svsignal.IE.beacon_1234_rx",
		Content_type: "text/plain",
		Data:         jData,
	}

	val2 := ValueSignal{
		system_key: "IE",
		signal_key: "beacon_1234_U",
		Value:      1010,
		UTime:      1636388615,
		Offline:    0,
		TypeSave:   1,
	}
	jData, _ = json.Marshal(val2)

	hub.CH_MSG_AMPQ <- rabbitmq.MessageAmpq{
		Exchange:     "svsignal",
		Routing_key:  "svsignal.IE.beacon_1234_U",
		Content_type: "text/plain",
		Data:         jData,
	}

	val3 := ValueSignal{
		system_key: "IE",
		signal_key: "beacon_1234_rx",
		Value:      2010,
		UTime:      1636388715,
		Offline:    1,
		TypeSave:   2,
	}
	jData, _ = json.Marshal(val3)

	hub.CH_MSG_AMPQ <- rabbitmq.MessageAmpq{
		Exchange:     "svsignal",
		Routing_key:  "svsignal.IE.beacon_1234_rx",
		Content_type: "text/plain",
		Data:         jData,
	}

	value := <-hub.CH_SAVE_VALUE
	if !reflect.DeepEqual(val1, value) {
		t.Errorf("not equal %v!=%v", val1, value)
	}
	value = <-hub.CH_SAVE_VALUE
	if !reflect.DeepEqual(val2, value) {
		t.Errorf("not equal %v!=%v", val2, value)
	}

	value = <-hub.CH_SAVE_VALUE
	if !reflect.DeepEqual(val3, value) {
		t.Errorf("not equal %v!=%v", val3, value)
	}

	close(hub.CH_SAVE_VALUE)
	cancel_hub()
	wg.Wait()
}
