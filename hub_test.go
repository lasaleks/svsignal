package main

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"testing"
	"time"

	gormq "bitbucket.org/lasaleks/go-rmq"
	"github.com/DATA-DOG/go-sqlmock"
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
		group_key:  "IE",
		signal_key: "beacon_1234_rx",
		Value:      100,
		UTime:      1636388515,
		Offline:    0,
	}
	jData, _ := json.Marshal(val1)

	hub.CH_MSG_AMPQ <- gormq.MessageAmpq{
		Exchange:     "svsignal",
		Routing_key:  "svs.save.IE.beacon_1234_rx",
		Content_type: "text/plain",
		Data:         jData,
	}

	val2 := ValueSignal{
		group_key:  "IE",
		signal_key: "beacon_1234_U",
		Value:      1010,
		UTime:      1636388615,
		Offline:    0,
	}
	jData, _ = json.Marshal(val2)

	hub.CH_MSG_AMPQ <- gormq.MessageAmpq{
		Exchange:     "svsignal",
		Routing_key:  "svs.save.IE.beacon_1234_U",
		Content_type: "text/plain",
		Data:         jData,
	}

	val3 := ValueSignal{
		group_key:  "IE",
		signal_key: "beacon_1234_rx",
		Value:      2010,
		UTime:      1636388715,
		Offline:    1,
	}
	jData, _ = json.Marshal(val3)

	hub.CH_MSG_AMPQ <- gormq.MessageAmpq{
		Exchange:     "svsignal",
		Routing_key:  "svs.save.IE.beacon_1234_rx",
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

func TestRmqSetSignal(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

	// mock db
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "group_key", "name"}).
		AddRow(1, "IE", "InsiteExpert").
		AddRow(2, "IEBlock", "InsiteExpert BlockCombine")

	mock.ExpectQuery("^SELECT (.+) FROM svsignal_group$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "signal_id", "tag", "value"}).
		AddRow(1, 1, "location", "pk110 1234").
		AddRow(2, 1, "desc", "rx/tx 1234").
		AddRow(3, 2, "location", "pk110 1235").
		AddRow(4, 2, "desc", "rx/tx 1235")
	mock.ExpectQuery("^SELECT (.+) FROM svsignal_tag$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "group_id", "group_key", "signal_key", "name", "type_save", "period", "delta"}).
		AddRow(1, 1, "IE", "beacon.1234.rx", "rx", 1, 60, 10000).
		AddRow(2, 1, "IE", "beacon.1235.rx", "rx", 1, 60, 10000)

	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)
	//-------

	var cfg Config

	// ch_ack := make(chan interface{}, 1)

	savesignal := newSVS()
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	hub := newHub()
	hub.CH_SAVE_VALUE = savesignal.CH_SAVE_VALUE
	hub.CH_SET_SIGNAL = savesignal.CH_SET_SIGNAL
	hub.debug_level = cfg.CONFIG.DEBUG_LEVEL
	ctx_hub, cancel_hub := context.WithCancel(ctx)
	wg.Add(1)
	go hub.run(&wg, ctx_hub)

	setsig := SetSignal{
		TypeSave: TYPE_FVALUE,
		Period:   60,
		Delta:    5.0,
		Name:     "Качество связи TX/RX",
		Tags: []struct {
			Tag   string `json:"tag"`
			Value string `json:"value"`
		}{
			{Tag: "site", Value: "Выработка в.5-1-2"},
			{Tag: "desc", Value: "Описание"},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_signal").WithArgs(1, "beacon_1236_rx", "Качество связи TX/RX", 2, 60, 5.0).WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_tag").WithArgs(3, "site", "Выработка в.5-1-2").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_tag").WithArgs(3, "desc", "Описание").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	jData, _ := json.Marshal(setsig)
	hub.CH_MSG_AMPQ <- gormq.MessageAmpq{
		Exchange:     "svsignal",
		Routing_key:  "svs.set.IE.beacon_1236_rx",
		Content_type: "text/plain",
		Data:         jData,
	}

	setsig = SetSignal{
		TypeSave: TYPE_FVALUE,
		Period:   70,
		Delta:    6.0,
		Name:     "-Качество связи TX/RX-",
		Tags: []struct {
			Tag   string `json:"tag"`
			Value string `json:"value"`
		}{
			{Tag: "site", Value: "-Выработка в.5-1-2-update"},
			{Tag: "desc", Value: "Описание"},
			{Tag: "newtag", Value: "-test-"},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE svsignal_signal").WithArgs("-Качество связи TX/RX-", 2, 70, 6.0, 3).WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectCommit()

	//UPDATE svsignal_tag SET tag=?, value=? WHERE id=?
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE svsignal_tag").WithArgs("site", "-Выработка в.5-1-2-update", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_tag").WithArgs(3, "newtag", "-test-").WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectCommit()

	jData, _ = json.Marshal(setsig)
	hub.CH_MSG_AMPQ <- gormq.MessageAmpq{
		Exchange:     "svsignal",
		Routing_key:  "svs.set.IE.beacon_1236_rx",
		Content_type: "text/plain",
		Data:         jData,
	}

	for len(hub.CH_MSG_AMPQ) > 0 {
		time.Sleep(time.Microsecond * 10)
	}

	for len(hub.CH_SET_SIGNAL) > 0 {
		time.Sleep(time.Microsecond * 10)
	}

	//time.Sleep(time.Microsecond * 10)
	//close(hub.CH_MSG_AMPQ)
	//close(hub.CH_SET_SIGNAL)
	time.Sleep(time.Microsecond * 1000)
	cancel_hub()
	cancel_db()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
