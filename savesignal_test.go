package main

import (
	"context"
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
	mock.ExpectExec("INSERT INTO svsignal_signal").WithArgs(3, "test1", "", 1, 60, float32(10000)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_ivalue").WithArgs(1, 10, 1636278215, 0).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_signal").WithArgs(3, "test2", "", 1, 60, float32(10000)).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO svsignal_ivalue").WithArgs(2, 10, 1636278215, 0).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	savesignal := newSVS()
	savesignal.db = db
	ctx_db, cancel_db := context.WithCancel(ctx)
	wg.Add(1)
	go savesignal.run(&wg, ctx_db)

	savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "TestSys", signal_key: "test1", Value: 10, UTime: 1636278215, Offline: 0, TypeSave: 1}
	savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "TestSys", signal_key: "test2", Value: 10, UTime: 1636278215, Offline: 0, TypeSave: 1}

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

	savesignal := newSVS()
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
		savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "Group", signal_key: "Test", Value: value.value, UTime: value.utime, Offline: int64(value.offline), TypeSave: 1}
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

	savesignal := newSVS()
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
		savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "Group", signal_key: "Test", Value: value.value, UTime: value.utime, Offline: int64(value.offline), TypeSave: 2}
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

	savesignal := newSVS()
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
		savesignal.CH_SAVE_VALUE <- ValueSignal{group_key: "Group", signal_key: "Test", Value: value.value, UTime: value.utime, Offline: int64(value.offline), TypeSave: 3}
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
