package main

import (
	"context"
	"database/sql/driver"
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

func TestCreate(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

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
		AddRow(1, "1", "location", "asdf").
		AddRow(2, "1", "desc", "asdfqwer")

	mock.ExpectQuery("^SELECT (.+) FROM svsignal_tag$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "group_id", "group_key", "signal_key", "name", "type_save", "period", "delta"}).
		AddRow(1, 1, "IE", "1234.rx", "rx", 1, 60, float32(10000)).
		AddRow(2, 1, "IE", "1235.rx", "rx", 1, 60, float32(10000))

	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_group").WithArgs("TestSys", "").WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_signal").WithArgs(3, "test1", "Check1", 1, 60, float32(10000)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_ivalue").WithArgs(1, 10, 1636278215, 0).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_signal").WithArgs(3, "test2", "Check2", 1, 60, float32(10000)).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_ivalue").WithArgs(2, 10, 1636278215, 0).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	savesignal := newSVS(Config{})
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	savesignal.CH_SET_SIGNAL <- SetSignal{group_key: "TestSys", signal_key: "test1", TypeSave: 1, Name: "Check1", Period: 60, Delta: 10000}
	for len(savesignal.CH_SET_SIGNAL) > 0 {
		time.Sleep(time.Millisecond * 1)
	}
	savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "TestSys", signal_key: "test1", Value: 10, UTime: 1636278215, Offline: 0}
	for len(savesignal.CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Millisecond * 1)
	}
	savesignal.CH_SET_SIGNAL <- SetSignal{group_key: "TestSys", signal_key: "test2", TypeSave: 1, Name: "Check2", Period: 60, Delta: 10000}
	for len(savesignal.CH_SET_SIGNAL) > 0 {
		time.Sleep(time.Millisecond * 1)
	}
	savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "TestSys", signal_key: "test2", Value: 10, UTime: 1636278215, Offline: 0}
	for len(savesignal.CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Microsecond * 10)
	}

	close(savesignal.CH_SAVE_VALUE)
	cancel_db()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestSave1(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "group_key", "name"}).
		AddRow(1, "Group", "Test")

	mock.ExpectQuery("^SELECT (.+) FROM svsignal_group$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "signal_id", "tag", "value"})
	mock.ExpectQuery("^SELECT (.+) FROM svsignal_tag$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "group_id", "g.group_key", "signal_key", "name", "type_save", "period", "delta"}).
		AddRow(2, 1, "Group", "Test", "-", 1, 60, 10000)
	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)

	for _, v := range []struct {
		signal_id int64
		value     int64
		utime     int64
		offline   int
	}{
		{signal_id: 2, value: 10, utime: 1636278215, offline: 0},
		{signal_id: 2, value: 20, utime: 1636278218, offline: 0},
		{signal_id: 2, value: 20, utime: 1636278221, offline: 1},
		{signal_id: 2, value: 20, utime: 1636278224, offline: 0},
	} {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO svsignal_ivalue").WithArgs(v.signal_id, v.value, v.utime, v.offline).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
	}

	savesignal := newSVS(Config{})
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	for _, value := range []struct {
		value   float64
		utime   int64
		offline int
	}{
		{value: 10, utime: 1636278215, offline: 0},
		{value: 10, utime: 1636278216, offline: 0},
		{value: 10, utime: 1636278217, offline: 0},
		{value: 20, utime: 1636278218, offline: 0},
		{value: 20, utime: 1636278219, offline: 0},
		{value: 20, utime: 1636278220, offline: 0},
		{value: 20, utime: 1636278221, offline: 1},
		{value: 20, utime: 1636278222, offline: 1},
		{value: 20, utime: 1636278223, offline: 1},
		{value: 20, utime: 1636278224, offline: 0},
		{value: 20, utime: 1636278225, offline: 0},
		{value: 20, utime: 1636278226, offline: 0},
		{value: 20, utime: 1636278326, offline: 0},
		{value: 20, utime: 1636278426, offline: 0},
		{value: 20, utime: 1636278526, offline: 0},
	} {
		savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "Group", signal_key: "Test", Value: value.value, UTime: value.utime, Offline: int64(value.offline)}
	}

	for len(savesignal.CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Microsecond * 10)
	}

	close(savesignal.CH_SAVE_VALUE)
	cancel_db()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestSave2(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"group_id", "group_key", "name"}).
		AddRow(1, "Group", "Test")

	mock.ExpectQuery("^SELECT (.+) FROM svsignal_group$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "signal_id", "tag", "value"})
	mock.ExpectQuery("^SELECT (.+) FROM svsignal_tag$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "group_id", "group_key", "signal_key", "name", "type_save", "period", "delta"}).
		AddRow(1, 1, "Group", "Test", "-", 2, 60, 10000)
	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)

	for _, v := range []struct {
		system_id int64
		value     float64
		utime     int64
		offline   int
	}{
		{system_id: 1, value: 15, utime: 1636278265, offline: 0},
		{system_id: 1, value: 25, utime: 1636278325, offline: 0},
	} {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO svsignal_fvalue").WithArgs(v.system_id, v.value, v.utime, v.offline).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
	}

	savesignal := newSVS(Config{})
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	for _, value := range []struct {
		value   float64
		utime   int64
		offline int
	}{
		{value: 10, utime: 1636278240, offline: 0},
		{value: 10, utime: 1636278250, offline: 0},
		{value: 10, utime: 1636278260, offline: 0},
		{value: 20, utime: 1636278270, offline: 0},
		{value: 20, utime: 1636278280, offline: 0},
		{value: 20, utime: 1636278290, offline: 0},
		{value: 20, utime: 1636278300, offline: 0},
		{value: 20, utime: 1636278310, offline: 0},
		{value: 20, utime: 1636278320, offline: 0},
		{value: 30, utime: 1636278330, offline: 0},
		{value: 30, utime: 1636278340, offline: 0},
		{value: 30, utime: 1636278350, offline: 0},
		{value: 30, utime: 1636278360, offline: 0},
		{value: 30, utime: 1636278370, offline: 0},
		{value: 30, utime: 1636278380, offline: 0},
		{value: 30, utime: 1636278390, offline: 0},
		{value: 30, utime: 1636278400, offline: 0},
	} {
		savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "Group", signal_key: "Test", Value: value.value, UTime: value.utime, Offline: int64(value.offline)}
	}

	for len(savesignal.CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Microsecond * 10)
	}

	close(savesignal.CH_SAVE_VALUE)
	cancel_db()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestSave3(t *testing.T) {
	t.Skip()
	var wg sync.WaitGroup
	ctx := context.Background()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "group_key", "name"}).
		AddRow(1, "Group", "Test")

	mock.ExpectQuery("^SELECT (.+) FROM svsignal_group$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "signal_id", "tag", "value"})
	mock.ExpectQuery("^SELECT (.+) FROM svsignal_tag$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "group_id", "group_key", "signal_key", "name", "type_save", "period", "delta"}).
		AddRow(1, 1, "Group", "Test", "-", 2, 60, 10000)
	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)

	for _, v := range []struct {
		system_id int64
		max       float64
		min       float64
		mean      float64
		median    float64
		value     float64
		utime     int64
		offline   int
	}{
		{system_id: 1, max: 100, min: 0, mean: 50, median: 20, utime: 1636278265, offline: 0},
	} {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO svsignal_mvalue").WithArgs(v.system_id, v.value, v.utime, v.offline).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
	}

	savesignal := newSVS(Config{})
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	for _, value := range []struct {
		value   float64
		utime   int64
		offline int
	}{
		{value: 10, utime: 1636278240, offline: 0},
		{value: 10, utime: 1636278250, offline: 0},
		{value: 10, utime: 1636278260, offline: 0},
		{value: 20, utime: 1636278270, offline: 0},
		{value: 20, utime: 1636278280, offline: 0},
		{value: 20, utime: 1636278290, offline: 0},
		{value: 20, utime: 1636278300, offline: 0},
		{value: 20, utime: 1636278310, offline: 0},
		{value: 20, utime: 1636278320, offline: 0},
		{value: 30, utime: 1636278330, offline: 0},
		{value: 30, utime: 1636278340, offline: 0},
		{value: 30, utime: 1636278350, offline: 0},
		{value: 30, utime: 1636278360, offline: 0},
		{value: 30, utime: 1636278370, offline: 0},
		{value: 30, utime: 1636278380, offline: 0},
		{value: 30, utime: 1636278390, offline: 0},
		{value: 30, utime: 1636278400, offline: 0},
	} {
		savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "Group", signal_key: "Test", Value: value.value, UTime: value.utime, Offline: int64(value.offline)}
	}

	for len(savesignal.CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Microsecond * 10)
	}

	close(savesignal.CH_SAVE_VALUE)
	cancel_db()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

// Тестирование сохранения Дискретных значений
func TestMultiplyInsertDescreetValues(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "group_key", "name"}).
		AddRow(1, "Group", "Test")

	mock.ExpectQuery("^SELECT (.+) FROM svsignal_group$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "signal_id", "tag", "value"})
	mock.ExpectQuery("^SELECT (.+) FROM svsignal_tag$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "group_id", "g.group_key", "signal_key", "name", "type_save", "period", "delta"}).
		AddRow(2, 1, "Group", "Test", "-", 1, 60, 10000)
	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)

	type bulkInsertValues struct {
		signal_id int64
		value     int64
		utime     int64
		offline   int
	}

	insValues := [][]bulkInsertValues{
		{
			{signal_id: 2, value: 10, utime: 1636278215, offline: 0},
			{signal_id: 2, value: 20, utime: 1636278218, offline: 0},
			{signal_id: 2, value: 20, utime: 1636278221, offline: 1},
			{signal_id: 2, value: 20, utime: 1636278224, offline: 0},
		},
	}

	for _, arrV := range insValues {
		vals := []driver.Value{}
		for _, v := range arrV {
			vals = append(vals, v.signal_id, v.value, v.utime, v.offline)
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO svsignal_ivalue").WithArgs(vals...).WillReturnResult(sqlmock.NewResult(1, int64(len(arrV))))
		mock.ExpectCommit()
	}

	cfg := Config{}
	cfg.CONFIG.DEBUG_LEVEL = 10
	cfg.CONFIG.MaxMultiplyInsert = 10
	cfg.CONFIG.BuffSize = 10
	savesignal := newSVS(cfg)
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	for _, value := range []struct {
		value   float64
		utime   int64
		offline int
	}{
		{value: 10, utime: 1636278215, offline: 0},
		{value: 10, utime: 1636278216, offline: 0},
		{value: 10, utime: 1636278217, offline: 0},
		{value: 20, utime: 1636278218, offline: 0},
		{value: 20, utime: 1636278219, offline: 0},
		{value: 20, utime: 1636278220, offline: 0},
		{value: 20, utime: 1636278221, offline: 1},
		{value: 20, utime: 1636278222, offline: 1},
		{value: 20, utime: 1636278223, offline: 1},
		{value: 20, utime: 1636278224, offline: 0},
		{value: 20, utime: 1636278225, offline: 0},
		{value: 20, utime: 1636278226, offline: 0},
		{value: 20, utime: 1636278326, offline: 0},
		{value: 20, utime: 1636278426, offline: 0},
		{value: 20, utime: 1636278526, offline: 0},
	} {
		savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "Group", signal_key: "Test", Value: value.value, UTime: value.utime, Offline: int64(value.offline)}
	}

	for len(savesignal.CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Microsecond * 10)
	}

	close(savesignal.CH_SAVE_VALUE)
	cancel_db()
	wg.Wait()

	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

// Тестирование запроса дискретных данных за период, в который попадают данные из буффера на запись
func TestQueryDescreetDataFromCache(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "group_key", "name"}).
		AddRow(1, "IE", "InsiteExpert")

	mock.ExpectQuery("^SELECT (.+) FROM svsignal_group$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "signal_id", "tag", "value"})
	mock.ExpectQuery("^SELECT (.+) FROM svsignal_tag$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "group_id", "g.group_key", "signal_key", "name", "type_save", "period", "delta"}).
		AddRow(2, 1, "IE", "beacon.1234.rx", "rx", 1, 60, 10000)
	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)

	type bulkInsertValues struct {
		signal_id int64
		value     int64
		utime     int64
		offline   int
	}

	insValues := [][]bulkInsertValues{
		{
			{signal_id: 2, value: 10, utime: 1636278200, offline: 0},
			{signal_id: 2, value: 11, utime: 1636278202, offline: 0},
			{signal_id: 2, value: 12, utime: 1636278204, offline: 0},
			{signal_id: 2, value: 13, utime: 1636278206, offline: 0},
			{signal_id: 2, value: 14, utime: 1636278207, offline: 0},
			{signal_id: 2, value: 10, utime: 1636278210, offline: 0},
			{signal_id: 2, value: 12, utime: 1636278212, offline: 0},
			{signal_id: 2, value: 13, utime: 1636278213, offline: 0},
			{signal_id: 2, value: 14, utime: 1636278214, offline: 0},
		},
	}

	for _, arrV := range insValues {
		vals := []driver.Value{}
		for _, v := range arrV {
			vals = append(vals, v.signal_id, v.value, v.utime, v.offline)
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO svsignal_ivalue").WithArgs(vals...).WillReturnResult(sqlmock.NewResult(1, int64(len(arrV))))
		mock.ExpectCommit()
	}

	cfg := Config{}
	cfg.CONFIG.DEBUG_LEVEL = 10
	cfg.CONFIG.MaxMultiplyInsert = 10
	savesignal := newSVS(cfg)
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	for _, value := range []struct {
		value   float64
		utime   int64
		offline int
	}{
		{value: 10, utime: 1636278200, offline: 0},
		{value: 10, utime: 1636278201, offline: 0},
		{value: 11, utime: 1636278202, offline: 0},
		{value: 11, utime: 1636278203, offline: 0},
		{value: 12, utime: 1636278204, offline: 0},
		{value: 12, utime: 1636278205, offline: 0},
		{value: 13, utime: 1636278206, offline: 0},
		{value: 14, utime: 1636278207, offline: 0},
		{value: 14, utime: 1636278208, offline: 0},
		{value: 14, utime: 1636278209, offline: 0},
		{value: 10, utime: 1636278210, offline: 0},
		{value: 10, utime: 1636278211, offline: 0},
		{value: 12, utime: 1636278212, offline: 0},
		{value: 13, utime: 1636278213, offline: 0},
		{value: 14, utime: 1636278214, offline: 0},
		{value: 14, utime: 1636278215, offline: 0},
		{value: 14, utime: 1636278216, offline: 0},
	} {
		savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "IE", signal_key: "beacon.1234.rx", Value: value.value, UTime: value.utime, Offline: int64(value.offline)}
	}

	for len(savesignal.CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Microsecond * 10)
	}

	tt := []struct {
		skip             bool
		caseName         string
		signal_id        int
		begin            int64
		end              int64
		resp_begin_row   *sqlmock.Rows
		resp_query_rows  *sqlmock.Rows
		resp_end_row     *sqlmock.Rows
		response_compare ResponseDataSignalT1
	}{
		{
			skip:            false,
			caseName:        "Проверка запроса данных за период, данные которого находятся полностью в буффере, из БД данные не запрашиваются",
			signal_id:       2,
			begin:           1636278201,
			end:             1636278209,
			resp_begin_row:  nil,
			resp_query_rows: nil,
			response_compare: ResponseDataSignalT1{
				GroupKey:   "IE",
				GroupName:  "InsiteExpert",
				SignalKey:  "beacon.1234.rx",
				SignalName: "rx",
				TypeSave:   1,
				Values: [][4]int64{
					{0, 1636278200, 10, 0},
					{0, 1636278202, 11, 0},
					{0, 1636278204, 12, 0},
					{0, 1636278206, 13, 0},
					{0, 1636278207, 14, 0},
					{0, 1636278210, 10, 0},
				},
				Tags: []RLS_Tag{},
			},
		},
		{
			skip:            false,
			caseName:        "Проверка запроса данных за период, данные которого находятся полностью в буффере, из БД данные не запрашиваются",
			signal_id:       2,
			begin:           1636278201,
			end:             1636278220,
			resp_begin_row:  nil,
			resp_query_rows: nil,
			response_compare: ResponseDataSignalT1{
				GroupKey:   "IE",
				GroupName:  "InsiteExpert",
				SignalKey:  "beacon.1234.rx",
				SignalName: "rx",
				TypeSave:   1,
				Values: [][4]int64{
					{0, 1636278200, 10, 0},
					{0, 1636278202, 11, 0},
					{0, 1636278204, 12, 0},
					{0, 1636278206, 13, 0},
					{0, 1636278207, 14, 0},
					{0, 1636278210, 10, 0},
					{0, 1636278212, 12, 0},
					{0, 1636278213, 13, 0},
					{0, 1636278214, 14, 0},
					{0, 1636278216, 14, 0},
				},
				Tags: []RLS_Tag{},
			},
		},
		{
			skip:            false,
			caseName:        "Запрос за период, данные за весь период в кеше, начало периода равно посл.зн в кеше. Требуется дополнить значением за периодам из БД",
			signal_id:       2,
			begin:           1636278200,
			end:             1636278209,
			resp_begin_row:  sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}).AddRow(6, 1636278198, 9, 0),
			resp_query_rows: nil,
			response_compare: ResponseDataSignalT1{
				GroupKey:   "IE",
				GroupName:  "InsiteExpert",
				SignalKey:  "beacon.1234.rx",
				SignalName: "rx",
				TypeSave:   1,
				Values: [][4]int64{
					{6, 1636278198, 9, 0},
					{0, 1636278200, 10, 0},
					{0, 1636278202, 11, 0},
					{0, 1636278204, 12, 0},
					{0, 1636278206, 13, 0},
					{0, 1636278207, 14, 0},
					{0, 1636278210, 10, 0},
				},
				Tags: []RLS_Tag{},
			},
		},
		{
			skip:            false,
			caseName:        "Запрос за период, часть данных в кеше, begin < первого значения в буффере, end < последнего значения в буфере",
			signal_id:       2,
			begin:           1636278150,
			end:             1636278209,
			resp_begin_row:  sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}).AddRow(2, 1636278188, 7, 0),
			resp_query_rows: sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}).AddRow(3, 1636278150, 9, 0).AddRow(4, 1636278151, 2, 0).AddRow(5, 1636278152, 3, 0).AddRow(6, 1636278153, 6, 0).AddRow(7, 1636278158, 9, 0),
			response_compare: ResponseDataSignalT1{
				GroupKey:   "IE",
				GroupName:  "InsiteExpert",
				SignalKey:  "beacon.1234.rx",
				SignalName: "rx",
				TypeSave:   1,
				Values: [][4]int64{
					{2, 1636278188, 7, 0},
					{3, 1636278150, 9, 0},
					{4, 1636278151, 2, 0},
					{5, 1636278152, 3, 0},
					{6, 1636278153, 6, 0},
					{7, 1636278158, 9, 0},

					{0, 1636278200, 10, 0},
					{0, 1636278202, 11, 0},
					{0, 1636278204, 12, 0},
					{0, 1636278206, 13, 0},
					{0, 1636278207, 14, 0},
					{0, 1636278210, 10, 0},
				},
				Tags: []RLS_Tag{},
			},
		},
		{
			skip:            false,
			caseName:        "Запрос за период, часть данных в кеше, begin < первого значения в буффере, end < первого значения в буфере",
			signal_id:       2,
			begin:           1636278150,
			end:             1636278170,
			resp_begin_row:  sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}).AddRow(2, 1636278188, 7, 0),
			resp_query_rows: sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}).AddRow(3, 1636278150, 9, 0).AddRow(4, 1636278151, 2, 0).AddRow(5, 1636278152, 3, 0).AddRow(6, 1636278153, 6, 0).AddRow(7, 1636278158, 9, 0),
			resp_end_row:    nil,
			response_compare: ResponseDataSignalT1{
				GroupKey:   "IE",
				GroupName:  "InsiteExpert",
				SignalKey:  "beacon.1234.rx",
				SignalName: "rx",
				TypeSave:   1,
				Values: [][4]int64{
					{2, 1636278188, 7, 0},
					{3, 1636278150, 9, 0},
					{4, 1636278151, 2, 0},
					{5, 1636278152, 3, 0},
					{6, 1636278153, 6, 0},
					{7, 1636278158, 9, 0},
					{0, 1636278200, 10, 0},
				},
				Tags: []RLS_Tag{},
			},
		},
		{
			skip:            false,
			caseName:        "Запрос за период, часть данных в кеше, begin < первого значения в буффере, end < первого значения в буфере",
			signal_id:       2,
			begin:           1636278150,
			end:             1636278170,
			resp_begin_row:  sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}).AddRow(2, 1636278188, 7, 0),
			resp_query_rows: sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}).AddRow(3, 1636278150, 9, 0).AddRow(4, 1636278151, 2, 0).AddRow(5, 1636278152, 3, 0).AddRow(6, 1636278153, 6, 0).AddRow(7, 1636278158, 9, 0),
			resp_end_row:    sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}).AddRow(8, 1636278172, 10, 0),
			response_compare: ResponseDataSignalT1{
				GroupKey:   "IE",
				GroupName:  "InsiteExpert",
				SignalKey:  "beacon.1234.rx",
				SignalName: "rx",
				TypeSave:   1,
				Values: [][4]int64{
					{2, 1636278188, 7, 0},
					{3, 1636278150, 9, 0},
					{4, 1636278151, 2, 0},
					{5, 1636278152, 3, 0},
					{6, 1636278153, 6, 0},
					{7, 1636278158, 9, 0},
					{8, 1636278172, 10, 0},
				},
				Tags: []RLS_Tag{},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}
			db_prev := savesignal.db
			defer func() {
				savesignal.db = db_prev
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			savesignal.db = db

			if tc.resp_begin_row != nil {
				mock.ExpectQuery(fmt.Sprintf("SELECT (.+) FROM svsignal_ivalue WHERE signal_id=%d and id=", tc.signal_id)).WillReturnRows(tc.resp_begin_row)
			}
			if tc.resp_query_rows != nil {
				mock.ExpectQuery(fmt.Sprintf("^SELECT (.+) FROM svsignal_ivalue WHERE signal_id=%d and utime >= %d and utime <=%d$", tc.signal_id, tc.begin, tc.end)).WillReturnRows(tc.resp_query_rows)
			}
			if tc.resp_end_row != nil {
				mock.ExpectQuery(fmt.Sprintf("SELECT (.+) FROM svsignal_ivalue WHERE signal_id=%d and id=", tc.signal_id)).WillReturnRows(tc.resp_end_row)
			}

			url := fmt.Sprintf("http://localhost:8080/api/signal/getdata?signalkey=IE.beacon.1234.rx&begin=%d&end=%d", tc.begin, tc.end)
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

			jData, _ := json.Marshal(tc.response_compare)
			if bodyStr != string(jData) {
				t.Errorf("wrong Response: got \n%+v\n, expected \n%+v", bodyStr, string(jData))
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expections: %s", err)
			}
			db.Close()
		})
	}

	close(savesignal.CH_SAVE_VALUE)
	cancel_db()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

// Тестирование сохранения усредненных значения
func TestMultiplyInsertValuesAvg(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "group_key", "name"}).
		AddRow(1, "Group", "Test")

	mock.ExpectQuery("^SELECT (.+) FROM svsignal_group$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "signal_id", "tag", "value"})
	mock.ExpectQuery("^SELECT (.+) FROM svsignal_tag$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "group_id", "g.group_key", "signal_key", "name", "type_save", "period", "delta"}).
		AddRow(2, 1, "IE", "beacon.1234.volt", "-", 2, 60, 10000)
	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)

	type bulkInsertValues struct {
		signal_id int64
		value     float64
		utime     int64
		offline   int
	}

	insValues := [][]bulkInsertValues{
		{
			{signal_id: 2, value: 40, utime: 1636278240, offline: 0},
			{signal_id: 2, value: 20, utime: 1636278320, offline: 0},
			{signal_id: 2, value: 50, utime: 1636278360, offline: 0},
			{signal_id: 2, value: 25, utime: 1636278505, offline: 0},
			{signal_id: 2, value: 73, utime: 1636278565, offline: 0},
		},
	}

	for _, arrV := range insValues {
		vals := []driver.Value{}
		for _, v := range arrV {
			vals = append(vals, v.signal_id, v.value, v.utime, v.offline)
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO svsignal_fvalue").WithArgs(vals...).WillReturnResult(sqlmock.NewResult(1, int64(len(arrV))))
		mock.ExpectCommit()
	}

	cfg := Config{}
	cfg.CONFIG.DEBUG_LEVEL = 10
	cfg.CONFIG.MaxMultiplyInsert = 10
	cfg.CONFIG.BuffSize = 10
	savesignal := newSVS(cfg)
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	for _, value := range []struct {
		value   float64
		utime   int64
		offline int
	}{
		//
		{value: 10, utime: 1636278240, offline: 0},
		{value: 20, utime: 1636278240, offline: 0},
		{value: 30, utime: 1636278240, offline: 0},
		{value: 40, utime: 1636278240, offline: 0},
		{value: 20, utime: 1636278320, offline: 0},
		{value: 50, utime: 1636278360, offline: 0},

		{value: 0, utime: 1636278480, offline: 0},
		{value: 10, utime: 1636278490, offline: 0},
		{value: 20, utime: 1636278500, offline: 0},
		{value: 30, utime: 1636278510, offline: 0},
		{value: 40, utime: 1636278520, offline: 0},
		{value: 50, utime: 1636278530, offline: 0},
		{value: 60, utime: 1636278540, offline: 0},

		{value: 70, utime: 1636278550, offline: 0},
		{value: 70, utime: 1636278555, offline: 0},
		{value: 70, utime: 1636278555, offline: 0},
		{value: 70, utime: 1636278560, offline: 0},
		{value: 80, utime: 1636278580, offline: 0},
		{value: 80, utime: 1636278590, offline: 0},
		{value: 80, utime: 1636278600, offline: 0},
		{value: 80, utime: 1636278610, offline: 0},
	} {
		savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "IE", signal_key: "beacon.1234.volt", Value: value.value, UTime: value.utime, Offline: int64(value.offline)}
	}

	for len(savesignal.CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Microsecond * 10)
	}

	close(savesignal.CH_SAVE_VALUE)
	cancel_db()
	wg.Wait()

	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestQueryAvgDataFromCache(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "group_key", "name"}).
		AddRow(1, "IE", "InsiteExpert")

	mock.ExpectQuery("^SELECT (.+) FROM svsignal_group$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "signal_id", "tag", "value"})
	mock.ExpectQuery("^SELECT (.+) FROM svsignal_tag$").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "group_id", "g.group_key", "signal_key", "name", "type_save", "period", "delta"}).
		AddRow(2, 1, "IE", "beacon.1234.volt", "volt", 2, 60, 10000)
	mock.ExpectQuery("^SELECT s.id, s.group_id, g.group_key, s.signal_key, s.name, s.type_save, s.period, s.delta FROM svsignal_signal s inner join svsignal_group g on g.id=s.group_id$").WillReturnRows(rows)

	type bulkInsertValues struct {
		signal_id int64
		value     float64
		utime     int64
		offline   int
	}

	insValues := [][]bulkInsertValues{
		{
			{signal_id: 2, value: 40, utime: 1636278240, offline: 0},
			{signal_id: 2, value: 20, utime: 1636278320, offline: 0},
			{signal_id: 2, value: 50, utime: 1636278360, offline: 0},
			{signal_id: 2, value: 25, utime: 1636278505, offline: 0},
			{signal_id: 2, value: 73, utime: 1636278565, offline: 0},
		},
	}

	for _, arrV := range insValues {
		vals := []driver.Value{}
		for _, v := range arrV {
			vals = append(vals, v.signal_id, v.value, v.utime, v.offline)
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO svsignal_fvalue").WithArgs(vals...).WillReturnResult(sqlmock.NewResult(1, int64(len(arrV))))
		mock.ExpectCommit()
	}

	cfg := Config{}
	cfg.CONFIG.DEBUG_LEVEL = 10
	cfg.CONFIG.MaxMultiplyInsert = 10
	cfg.CONFIG.BuffSize = 10
	savesignal := newSVS(cfg)
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	for _, value := range []struct {
		value   float64
		utime   int64
		offline int
	}{
		//
		{value: 10, utime: 1636278240, offline: 0},
		{value: 20, utime: 1636278240, offline: 0},
		{value: 30, utime: 1636278240, offline: 0},
		{value: 40, utime: 1636278240, offline: 0},
		{value: 20, utime: 1636278320, offline: 0},
		{value: 50, utime: 1636278360, offline: 0},

		{value: 0, utime: 1636278480, offline: 0},
		{value: 10, utime: 1636278490, offline: 0},
		{value: 20, utime: 1636278500, offline: 0},
		{value: 30, utime: 1636278510, offline: 0},
		{value: 40, utime: 1636278520, offline: 0},
		{value: 50, utime: 1636278530, offline: 0},
		{value: 60, utime: 1636278540, offline: 0},

		{value: 70, utime: 1636278550, offline: 0},
		{value: 70, utime: 1636278555, offline: 0},
		{value: 70, utime: 1636278555, offline: 0},
		{value: 70, utime: 1636278560, offline: 0},
		{value: 80, utime: 1636278580, offline: 0},
		{value: 80, utime: 1636278590, offline: 0},
		{value: 80, utime: 1636278600, offline: 0},
		{value: 80, utime: 1636278610, offline: 0},
	} {
		savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "IE", signal_key: "beacon.1234.volt", Value: value.value, UTime: value.utime, Offline: int64(value.offline)}
	}

	for len(savesignal.CH_SAVE_VALUE) > 0 {
		time.Sleep(time.Microsecond * 10)
	}

	tt := []struct {
		skip             bool
		caseName         string
		signal_id        int
		begin            int64
		end              int64
		resp_query_rows  *sqlmock.Rows
		response_compare ResponseDataSignalT2
	}{
		{
			skip:            true,
			caseName:        "Проверка запроса данных за период, данные которого находятся полностью в буффере, из БД данные не запрашиваются",
			signal_id:       2,
			begin:           1636278240,
			end:             1636278565,
			resp_query_rows: nil, // sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}).AddRow(3, 1636278150, 9, 0).AddRow(4, 1636278151, 2, 0).AddRow(5, 1636278152, 3, 0).AddRow(6, 1636278153, 6, 0).AddRow(7, 1636278158, 9, 0),
			response_compare: ResponseDataSignalT2{
				GroupKey:   "IE",
				GroupName:  "InsiteExpert",
				SignalKey:  "beacon.1234.volt",
				SignalName: "volt",
				TypeSave:   2,
				Values: [][4]interface{}{
					{int64(0), int64(1636278240), float64(40), int64(0)},
					{int64(0), int64(1636278320), float64(20), int64(0)},
					{int64(0), int64(1636278360), float64(50), int64(0)},
					{int64(0), int64(1636278505), float64(25), int64(0)},
					{int64(0), int64(1636278565), float64(73), int64(0)},
				},
				Tags: []RLS_Tag{},
			},
		},
		{
			skip:      false,
			caseName:  "Запрос за период, часть данных в кеше. Требуется дополнить из БД",
			signal_id: 2,
			begin:     1636278000,
			end:       1636279000,
			resp_query_rows: sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}).AddRow(
				3, 1636278010, float64(10.1), 0).AddRow(
				4, 1636278020, float64(10.2), 0).AddRow(
				5, 1636278030, float64(10.3), 0).AddRow(
				6, 1636278040, float64(10.4), 0).AddRow(
				7, 1636278050, float64(10.5), 0),
			response_compare: ResponseDataSignalT2{
				GroupKey:   "IE",
				GroupName:  "InsiteExpert",
				SignalKey:  "beacon.1234.volt",
				SignalName: "volt",
				TypeSave:   2,
				Values: [][4]interface{}{
					{int64(3), int64(1636278010), float64(10.1), int64(0)},
					{int64(4), int64(1636278020), float64(10.2), int64(0)},
					{int64(5), int64(1636278030), float64(10.3), int64(0)},
					{int64(6), int64(1636278040), float64(10.4), int64(0)},
					{int64(7), int64(1636278050), float64(10.5), int64(0)},

					{int64(0), int64(1636278240), float64(40), int64(0)},
					{int64(0), int64(1636278320), float64(20), int64(0)},
					{int64(0), int64(1636278360), float64(50), int64(0)},
					{int64(0), int64(1636278505), float64(25), int64(0)},
					{int64(0), int64(1636278565), float64(73), int64(0)},
				},
				Tags: []RLS_Tag{},
			},
		},
		{
			skip:      false,
			caseName:  "Запрос за период, данные за весь период только в БД",
			signal_id: 2,
			begin:     1636278000,
			end:       1636278200,
			resp_query_rows: sqlmock.NewRows([]string{"id", "unixtime", "value", "offline"}).AddRow(
				3, 1636278010, float64(10.1), 0).AddRow(
				4, 1636278020, float64(10.2), 0).AddRow(
				5, 1636278030, float64(10.3), 0).AddRow(
				6, 1636278040, float64(10.4), 0).AddRow(
				7, 1636278050, float64(10.5), 0),
			response_compare: ResponseDataSignalT2{
				GroupKey:   "IE",
				GroupName:  "InsiteExpert",
				SignalKey:  "beacon.1234.volt",
				SignalName: "volt",
				TypeSave:   2,
				Values: [][4]interface{}{
					{int64(3), int64(1636278010), float64(10.1), int64(0)},
					{int64(4), int64(1636278020), float64(10.2), int64(0)},
					{int64(5), int64(1636278030), float64(10.3), int64(0)},
					{int64(6), int64(1636278040), float64(10.4), int64(0)},
					{int64(7), int64(1636278050), float64(10.5), int64(0)},
				},
				Tags: []RLS_Tag{},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}
			db_prev := savesignal.db
			defer func() {
				savesignal.db = db_prev
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			savesignal.db = db

			if tc.resp_query_rows != nil {
				mock.ExpectQuery(fmt.Sprintf("^SELECT (.+) FROM svsignal_fvalue WHERE signal_id=%d and utime >= %d and utime <=%d$", tc.signal_id, tc.begin, tc.end)).WillReturnRows(tc.resp_query_rows)
			}

			url := fmt.Sprintf("http://localhost:8080/api/signal/getdata?signalkey=IE.beacon.1234.volt&begin=%d&end=%d", tc.begin, tc.end)
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

			jData, _ := json.Marshal(tc.response_compare)
			if bodyStr != string(jData) {
				t.Errorf("wrong Response: got \n%+v\n, expected \n%+v", bodyStr, string(jData))
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expections: %s", err)
			}
			db.Close()
		})
	}

	close(savesignal.CH_SAVE_VALUE)
	cancel_db()
	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
