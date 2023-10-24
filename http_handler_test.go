package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetListHttp(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

	// mock db
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "system_key", "name"}).
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

	// init
	savesignal := newSVS()
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.Run(&wg, ctx_db)

	/*hub := newHub()
	hub.CH_SAVE_VALUE = savesignal.CH_SAVE_VALUE
	hub.CH_REQUEST_HTTP_DB = savesignal.CH_REQUEST_HTTP
	hub.debug_level = 0
	ctx_hub, cancel_hub := context.WithCancel(ctx)
	wg.Add(1)
	go hub.run(&wg, ctx_hub)*/

	//---http GetList
	url := "http://localhost/api/listsignal"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	http_handler := GetListSignal{savesignal.CH_REQUEST_HTTP}
	http_handler.ServeHTTP(w, req)
	StatusCode := 200
	if w.Code != StatusCode {
		t.Errorf("wrong StatusCode: got %d, expected %d", w.Code, StatusCode)
	}

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	bodyStr := string(body)
	list_signal := ResponseListSignal{Groups: map[string]*RLS_Groups{
		"IE": {
			Name: "InsiteExpert",
			Signals: map[string]RLS_Signal{
				"IE.beacon.1234.rx": {
					Name:     "rx",
					TypeSave: 1,
					Period:   60,
					Delta:    10000,
					Tags: []RLS_Tag{
						{Tag: "location", Value: "pk110 1234"},
						{Tag: "desc", Value: "rx/tx 1234"},
					},
				},
				"IE.beacon.1235.rx": {
					Name:     "rx",
					TypeSave: 1,
					Period:   60,
					Delta:    10000,
					Tags: []RLS_Tag{
						{Tag: "location", Value: "pk110 1235"},
						{Tag: "desc", Value: "rx/tx 1235"},
					}},
			},
		},
		"IEBlock": {
			Name: "InsiteExpert BlockCombine",
			// GroupKey: "IEBlock",
			Signals: map[string]RLS_Signal{},
		},
	}}
	jData, _ := json.Marshal(list_signal.Groups)
	/*if err != nil {
		// handle error
	}*/
	w.Header().Set("Content-Type", "application/json")

	cmp_str := string(jData)
	if bodyStr != cmp_str {
		t.Errorf("wrong Response: got \n%+v\nexpected \n%+v\n", bodyStr, cmp_str)
	}
	//-------

	cancel_db()
	// cancel_hub()
	wg.Wait()
}

func TestGetListHttp_Empty(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

	// mock db
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "system_key", "name"}).
		AddRow(1, "IE", "InsiteExpert").
		AddRow(2, "IEBlock", "InsiteExpert BlockCombine")

	mock.ExpectQuery("^SELECT (.+) FROM svsignal_group$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "signal_id", "tag", "value"})
	mock.ExpectQuery("^SELECT (.+) FROM svsignal_tag$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "group_id", "group_key", "signal_key", "name", "type_save", "period", "delta"}).
		AddRow(1, 1, "IE", "beacon.1234.rx", "rx", 1, 60, 10000).
		AddRow(2, 1, "IE", "beacon.1235.rx", "rx", 1, 60, 10000)

	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)
	//-------

	// init
	savesignal := newSVS()
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.Run(&wg, ctx_db)

	/*hub := newHub()
	hub.CH_SAVE_VALUE = savesignal.CH_SAVE_VALUE
	hub.CH_REQUEST_HTTP_DB = savesignal.CH_REQUEST_HTTP
	hub.debug_level = 0
	ctx_hub, cancel_hub := context.WithCancel(ctx)
	wg.Add(1)
	go hub.run(&wg, ctx_hub)*/

	//---http GetList
	url := "http://localhost/api/listsignal"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	http_handler := GetListSignal{savesignal.CH_REQUEST_HTTP}
	http_handler.ServeHTTP(w, req)
	StatusCode := 200
	if w.Code != StatusCode {
		t.Errorf("wrong StatusCode: got %d, expected %d", w.Code, StatusCode)
	}

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	bodyStr := string(body)
	list_signal := ResponseListSignal{Groups: map[string]*RLS_Groups{
		"IE": {
			Name: "InsiteExpert",
			// GroupKey: "IE",
			Signals: map[string]RLS_Signal{
				"IE.beacon.1234.rx": {
					Name:     "rx",
					TypeSave: 1,
					Period:   60,
					Delta:    10000,
					Tags:     []RLS_Tag{},
				},
				"IE.beacon.1235.rx": {
					Name:     "rx",
					TypeSave: 1,
					Period:   60,
					Delta:    10000,
					Tags:     []RLS_Tag{},
				},
			},
		},
		"IEBlock": {
			Name: "InsiteExpert BlockCombine",
			// GroupKey: "IEBlock",
			Signals: map[string]RLS_Signal{},
		},
	}}
	jData, _ := json.Marshal(list_signal.Groups)
	/*if err != nil {
		// handle error
	}*/
	w.Header().Set("Content-Type", "application/json")

	cmp_str := string(jData)
	if bodyStr != cmp_str {
		t.Errorf("wrong Response: got \n%+v\nexpected \n%+v\n", bodyStr, cmp_str)
	}
	//-------

	cancel_db()
	// cancel_hub()
	wg.Wait()
}

func TestRequestDataT1Http(t *testing.T) {
	url := "http://localhost:8080/api/signal/getdata?signalkey=IE.beacon.1234.rx&begin=1636507647&end=1636508647"

	data_signal := ResponseDataSignalT1{
		GroupKey:   "IE",
		GroupName:  "InsiteExpert",
		SignalKey:  "beacon.1234.rx",
		SignalName: "rx",
		TypeSave:   1,
		Values:     [][4]int64{},
		Tags: []RLS_Tag{
			{Tag: "location", Value: "pk110 1234"},
			{Tag: "desc", Value: "rx/tx 1234"},
		},
	}

	var begin_utime int64 = 1636507647
	var begin_id int64 = 1
	var i int64
	for i = 0; i < 10; i++ {
		data_signal.Values = append(data_signal.Values, [4]int64{begin_id, begin_utime, i, 0})
		begin_utime += i * 10
		begin_id++
	}
	data_signal.Values = append(data_signal.Values, [4]int64{begin_id, begin_utime, 0, 1})
	begin_utime += 1
	begin_id++
	data_signal.Values = append(data_signal.Values, [4]int64{begin_id, begin_utime, 11, 0})
	begin_utime += 1
	begin_id++

	for i = 0; i < 5; i++ {
		data_signal.Values = append(data_signal.Values, [4]int64{begin_id, begin_utime, i, 0})
		begin_utime += 1
		begin_id++
	}

	var wg sync.WaitGroup
	ctx := context.Background()

	// mock db
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "system_key", "name"}).
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
		AddRow(2, 1, "IE", "beacon.1234.U", "U", 2, 60, 10000)

	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)

	mock.ExpectQuery(fmt.Sprintf("SELECT (.+) FROM svsignal_ivalue WHERE signal_id=%d and id=", 1)).WillReturnRows(sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}))

	rows = sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"})
	for _, value := range data_signal.Values {
		rows.AddRow(value[0], value[1], value[2], value[3])
	}
	mock.ExpectQuery("^SELECT (.+) FROM svsignal_ivalue WHERE signal_id=1 and utime >= 1636507647 and utime <=1636508647$").WillReturnRows(rows)

	mock.ExpectQuery(fmt.Sprintf("SELECT (.+) FROM svsignal_ivalue WHERE signal_id=%d and id=", 1)).WillReturnRows(sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}))
	//-------

	// init
	savesignal := newSVS()
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.Run(&wg, ctx_db)

	/*hub := newHub()
	hub.CH_SAVE_VALUE = savesignal.CH_SAVE_VALUE
	hub.CH_REQUEST_HTTP_DB = savesignal.CH_REQUEST_HTTP
	hub.debug_level = 0
	ctx_hub, cancel_hub := context.WithCancel(ctx)
	wg.Add(1)
	go hub.run(&wg, ctx_hub)*/

	//---http GetList
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	re_rkey, _ := regexp.Compile(`^(\w+)\.(.+)$`)
	http_handler := RequestSignalData{CH_REQUEST_HTTP: savesignal.CH_REQUEST_HTTP, re_key: re_rkey}
	http_handler.ServeHTTP(w, req)
	StatusCode := 200
	if w.Code != StatusCode {
		t.Errorf("wrong StatusCode: got %d, expected %d", w.Code, StatusCode)
	}

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	jData, _ := json.Marshal(data_signal)
	//fmt.Println("response", bodyStr)
	if bodyStr != string(jData) {
		t.Errorf("wrong Response: got %+v, expected %+v", bodyStr, string(jData))
	}
	//-------

	cancel_db()
	//cancel_hub()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestRequestDataT2Http(t *testing.T) {
	var begin int64 = 1636507647
	var end int64 = 1636508647
	url := fmt.Sprintf("http://localhost:8080/api/signal/getdata?signalkey=IE.beacon.1234.U&begin=%d&end=%d", begin, end)

	data_signal := ResponseDataSignalT2{
		GroupKey:   "IE",
		GroupName:  "InsiteExpert",
		SignalKey:  "beacon.1234.U",
		SignalName: "U",
		TypeSave:   2,
		Tags:       []RLS_Tag{},
	}
	var begin_utime int64 = begin
	var begin_id int64 = 1
	var i int64
	var value float64
	for i = 0; i < 10; i++ {
		value += 10.1
		data_signal.Values = append(data_signal.Values, [4]interface{}{begin_id, begin_utime, value, 0})
		begin_utime += i * 10
		begin_id++
	}
	data_signal.Values = append(data_signal.Values, [4]interface{}{begin_id, begin_utime, 0.0, 1})
	begin_utime += 1
	begin_id++
	data_signal.Values = append(data_signal.Values, [4]interface{}{begin_id, begin_utime, 100.2, 0})
	begin_utime += 1
	begin_id++

	for i = 0; i < 5; i++ {
		value += 10.1
		data_signal.Values = append(data_signal.Values, [4]interface{}{begin_id, begin_utime, value, 0})
		begin_utime += 1
		begin_id++
	}

	var wg sync.WaitGroup
	ctx := context.Background()

	// mock db
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "system_key", "name"}).
		AddRow(1, "IE", "InsiteExpert").
		AddRow(2, "IEBlock", "InsiteExpert BlockCombine")

	mock.ExpectQuery("^SELECT (.+) FROM svsignal_group$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "signal_id", "tag", "value"})
	mock.ExpectQuery("^SELECT (.+) FROM svsignal_tag$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "group_id", "group_key", "signal_key", "name", "type_save", "period", "delta"}).
		AddRow(1, 1, "IE", "beacon.1234.rx", "rx", 1, 60, 10000).
		AddRow(2, 1, "IE", "beacon.1234.U", "U", 2, 60, 10000)

	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"})
	for _, value := range data_signal.Values {
		rows.AddRow(value[0], value[1], value[2], value[3])
	}
	mock.ExpectQuery(fmt.Sprintf("^SELECT (.+) FROM svsignal_fvalue WHERE signal_id=2 and utime >= %d and utime <=%d$", begin, end)).WillReturnRows(rows)
	//-------

	// init
	savesignal := newSVS()
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.Run(&wg, ctx_db)

	/*hub := newHub()
	hub.CH_SAVE_VALUE = savesignal.CH_SAVE_VALUE
	hub.CH_REQUEST_HTTP_DB = savesignal.CH_REQUEST_HTTP
	hub.debug_level = 0
	ctx_hub, cancel_hub := context.WithCancel(ctx)
	wg.Add(1)
	go hub.run(&wg, ctx_hub)*/

	//---http GetList
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	re_rkey, _ := regexp.Compile(`^(\w+)\.(.+)$`)
	http_handler := &RequestSignalData{CH_REQUEST_HTTP: savesignal.CH_REQUEST_HTTP, re_key: re_rkey}
	http_handler.ServeHTTP(w, req)
	StatusCode := 200
	if w.Code != StatusCode {
		t.Errorf("wrong StatusCode: got %d, expected %d", w.Code, StatusCode)
	}

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	bodyStr := string(body)

	jData, _ := json.Marshal(data_signal)
	w.Header().Set("Content-Type", "application/json")

	//fmt.Println("response", bodyStr)
	if bodyStr != string(jData) {
		t.Errorf("wrong Response: got \n%+v\n, expected \n%+v", bodyStr, string(jData))
	}
	//-------

	cancel_db()
	//cancel_hub()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

// TODO переписать тесты в один test case
func TestRequestSaveValue(t *testing.T) {
	// mock db
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "system_key", "name"}).
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
	var wg sync.WaitGroup
	ctx := context.Background()

	savesignal := newSVS()
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.Run(&wg, ctx_db)

	key := "IE.beacon.1235.rx"
	value := 10.1
	utime := 1637295512
	offline := 0
	typesave := 1
	url := fmt.Sprintf("http://localhost:8080/api/savevalue?key=%s&value=%f&utime=%d&offline=%d&typesave=%d", key, value, utime, offline, typesave)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_ivalue").WithArgs(2, int(value), utime, offline).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	re_rkey, _ := regexp.Compile(`^(\w+)\.(.+)$`)
	http_handler := RequestSaveValue{CH_SAVE_VALUE: CH_SAVE_VALUE, re_key: re_rkey}
	http_handler.ServeHTTP(w, req)
	//resp := w.Result()

	for len(CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Microsecond * 10)
	}
	close(CH_SAVE_VALUE)

	cancel_db()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestRequestPostSaveValue(t *testing.T) {
	// mock db
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "system_key", "name"}).
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
	var wg sync.WaitGroup
	ctx := context.Background()

	savesignal := newSVS()
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.Run(&wg, ctx_db)

	key := "IE.beacon.1235.rx"
	value := 10.1
	utime := 1637295512
	offline := 0
	url := "http://localhost:8080/api/savevalue"

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_ivalue").WithArgs(2, int(value), utime, offline).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// url := fmt.Sprintf("http://localhost:8080/api/savevalue?key=%s&value=%f&utime=%d&offline=%d&typesave=%d", key, value, utime, offline, typesave)
	svalue := ReqJsonSaveValue{Key: key, Value: value, UTime: int64(utime), OffLine: int64(offline)}
	d, err := json.Marshal(svalue)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", url, bytes.NewBuffer(d))
	w := httptest.NewRecorder()

	re_rkey, _ := regexp.Compile(`^(\w+)\.(.+)$`)
	http_handler := RequestSaveValue{CH_SAVE_VALUE: CH_SAVE_VALUE, re_key: re_rkey}
	http_handler.ServeHTTP(w, req)
	//resp := w.Result()

	for len(CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Microsecond * 10)
	}
	close(CH_SAVE_VALUE)

	cancel_db()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestHttpSetSignal(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"id", "system_key", "name"}).
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
	var wg sync.WaitGroup
	ctx := context.Background()

	savesignal := newSVS()
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.Run(&wg, ctx_db)

	tt := []struct {
		caseName  string
		sig       HTTPRequestSetSignal
		type_save int // 1 - создание, 2 - обновление
		signal_id int
	}{
		{
			caseName:  "Создание",
			type_save: 1,
			signal_id: 3,
			sig: HTTPRequestSetSignal{
				Key:      "IE.beacon_1236_rx",
				TypeSave: 2,
				Period:   60,
				Delta:    10000,
				Name:     "Качество связи",
				Tags: []struct {
					Tag   string `json:"tag"`
					Value string `json:"value"`
				}{
					{Tag: "site", Value: "1236"},
					{Tag: "max_y", Value: "100"},
					{Tag: "min_y", Value: "-10"},
				},
			},
		},
		{
			caseName:  "UPDATE signal",
			type_save: 2,
			signal_id: 3,
			sig: HTTPRequestSetSignal{
				Key:      "IE.beacon_1236_rx",
				TypeSave: 1,
				Period:   0,
				Delta:    0,
				Name:     "-Качество связи-",
				Tags: []struct {
					Tag   string `json:"tag"`
					Value string `json:"value"`
				}{
					{Tag: "site", Value: "1236"},
					{Tag: "max_y", Value: "100"},
					{Tag: "min_y", Value: "-10"},
				},
			},
		},
	}

	url := "http://localhost:8080/api/setsignal"

	var tag_id int64 = 5
	re_rkey, _ := regexp.Compile(`^(\w+)\.(.+)$`)
	http_handler := HTTPSetSignal{CH_SET_SIGNAL: CH_SET_SIGNAL, re_key: re_rkey}
	for _, tc := range tt {
		t.Run(tc.caseName, func(t *testing.T) {

			var signal_key string
			ret_cmd := re_rkey.FindStringSubmatch(tc.sig.Key)
			_ = ret_cmd[1]
			signal_key = ret_cmd[2]
			switch tc.type_save {
			case 1:
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO svsignal_signal").WithArgs(1, signal_key, tc.sig.Name, tc.sig.TypeSave, tc.sig.Period, tc.sig.Delta).WillReturnResult(sqlmock.NewResult(int64(tc.signal_id), 1))
				mock.ExpectCommit()

				for _, tag := range tc.sig.Tags {
					mock.ExpectBegin()
					mock.ExpectExec("INSERT INTO svsignal_tag").WithArgs(tc.signal_id, tag.Tag, tag.Value).WillReturnResult(sqlmock.NewResult(tag_id, 1))
					mock.ExpectCommit()
					tag_id++
				}
			case 2:
				// UPDATE svsignal_signal SET name=?, type_save=?, period=?, delta=? WHERE id=?
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE svsignal_signal").WithArgs(tc.sig.Name, tc.sig.TypeSave, tc.sig.Period, tc.sig.Delta, tc.signal_id).WillReturnResult(sqlmock.NewResult(int64(tc.signal_id), 1))
				mock.ExpectCommit()
			}
			// TODO дополнить проверкой создания удаления svsignal_tags
			d, err := json.Marshal(tc.sig)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("POST", url, bytes.NewBuffer(d))
			w := httptest.NewRecorder()
			http_handler.ServeHTTP(w, req)

			for len(CH_SET_SIGNAL) > 0 {
				time.Sleep(time.Microsecond * 10)
			}
		})
	}

	for len(CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Microsecond * 10)
	}
	close(CH_SAVE_VALUE)

	cancel_db()
	wg.Wait()
}
